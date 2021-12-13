package seatbelt

import (
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strings"

	"github.com/go-chi/chi"
	"github.com/gorilla/csrf"
	"github.com/gorilla/sessions"
)

// An App contains the data necessary to start and run an application.
//
// An App acts as a router. You must provide your own HTTP server in order to
// start it the application, ie,
//
//	app := seatbelt.New()
//	http.ListenAndServe(":3000", app)
//
// Or,
//
//	app := seatbelt.New()
//	srv := &http.Server{
//		Handler: app,
//	}
//	srv.ListenAndServe()
type App struct {
	store             sessions.Store
	mux               chi.Router
	render            *Renderer
	signingKey        []byte
	skippedCSRFRoutes []string
	middlewares       []MiddlewareFunc
	errorHandler      func(c Context, err error)
}

// MiddlewareFunc is the type alias for Seatbelt middleware.
type MiddlewareFunc func(fn func(ctx Context) error) func(Context) error

// An Option is used to configure a Seatbelt application.
type Option struct {
	TemplateDir       string           // The directory where the templates reside.
	SigningKey        string           // The signing key for the cookie session store.
	Reload            bool             // Whether or not to reload templates on each request.
	SkippedCSRFRoutes []string         // Routes to skip CSRF checks.
	Funcs             template.FuncMap // HTML functions.
}

// setDefaults sets the default values for Seatbelt options.
func (o *Option) setDefaults() {
	if o.TemplateDir == "" {
		o.TemplateDir = "views"
	}
	if o.SigningKey == "" {
		o.SigningKey = "30b22798f5fa4429247dcf8bfd963887cf2a6fadb7eb6c8c1f2e0aa610c69ffd"
	}
}

// New returns a new instance of a Seatbelt application.
func New(opts ...Option) *App {
	var opt Option
	for _, o := range opts {
		opt = o
	}

	opt.setDefaults()

	// Register the encoding for map[string]interface{} with gob such that we
	// can successfully save flash messages in the session.
	gob.Register(map[string]interface{}{})

	signingKey, err := hex.DecodeString(opt.SigningKey)
	if err != nil {
		log.Fatalf("seatbelt: signing key is not a valid hexadecimal string: %+v", err)
	}

	cookieStore := sessions.NewCookieStore([]byte(opt.SigningKey))

	// Set secure defaults for the session cookie store.
	cookieStore.Options.HttpOnly = true
	// TODO (One day, when it's better supported) Change the default back to
	// SameSite="Lax". Right now it appears to cause unexpectd behaviour in the
	// embedded WebKit browser in some iOS apps, especially ones that haven't
	// kept up to date with updates.

	// Force a consistent path for browsers that are sensitive to this during
	// AJAX requests.
	cookieStore.Options.Path = "/"

	// Default to one year for new cookies, since some browsers don't set
	// their cookies with the same defaults.
	cookieStore.Options.MaxAge = 86400 * 365

	// TODO:
	//
	// Here we're assuming the reload value means that we're not in
	// development.
	//
	// This is typically true, but the environment value should be passed down
	// instead.
	cookieStore.Options.Secure = !opt.Reload

	// Initialize the underlying chi mux so that we can setup our default
	// middleware stack.
	mux := chi.NewRouter()
	if len(opt.SkippedCSRFRoutes) > 0 {
		mux.Use(skipCSRF(opt.SkippedCSRFRoutes))
	}
	mux.Use(csrf.Protect(signingKey))

	return &App{
		skippedCSRFRoutes: opt.SkippedCSRFRoutes,
		mux:               mux,
		store:             cookieStore,
		render:            NewRenderer(opt.TemplateDir, opt.Reload, opt.Funcs),
		signingKey:        signingKey,
	}
}

// skipCSRF skips CSRF checks for any request that is an exact match of one of
// the given routes.
func skipCSRF(routes []string) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			for _, route := range routes {
				if r.URL.Path == route {
					r = csrf.UnsafeSkipCheck(r)
				}
			}
			h.ServeHTTP(w, r)
		})
	}
}

// Start is a convenience method for starting the application server with a
// default *http.Server.
//
// Start should not be used in production, as the standard library's default
// HTTP server is not suitable for production use due to a lack of timeouts,
// etc.
//
// Production applications should create their own
// *http.Server, and pass the *seatbelt.App to that *http.Server's `Handler`.
func (a *App) Start(addr string) error {
	return http.ListenAndServe(addr, a)
}

// UseStd registers standard HTTP middleware on the application.
func (a *App) UseStd(middleware ...func(http.Handler) http.Handler) {
	a.mux.Use(middleware...)
}

// Use registers Seatbelt HTTP middleware on the application.
func (a *App) Use(middleware ...MiddlewareFunc) {
	a.middlewares = append(a.middlewares, middleware...)
}

// SetErrorHandler allows you to set a custom error handler that runs when an
// error is returned from an HTTP handler.
func (a *App) SetErrorHandler(fn func(c Context, err error)) {
	a.errorHandler = fn
}

// ErrorHandler is the globally registered error handler.
//
// You can override this function using `SetErrorHandler`.
func (a *App) ErrorHandler(c Context, err error) {
	if a.errorHandler != nil {
		a.errorHandler(c, err)
		return
	}

	fmt.Printf("hit error handler: %#v\n", err)

	switch c.Request().Method {
	case "GET", "HEAD", "OPTIONS":
		c.String(http.StatusInternalServerError, err.Error())
	default:
		from := c.Request().Referer()
		c.Session().Flash("alert", err.Error())
		c.Redirect(from)
	}
}

// serveContext creates and registers a Seatbelt handler for an HTTP request.
func (a *App) serveContext(w http.ResponseWriter, r *http.Request, handle func(c Context) error) {
	c := &context{w: w, r: r, store: a.store, render: a.render}

	// Iterate over the middleware in reverse order, so that the order
	// in which middleware is registered suggests that it is run from
	// the outermost (or leftmost) function to the innermost (or
	// rightmost) function.
	//
	// This means if you register two middlewares like,
	//	app.Use(m1, m2)
	// It will run as:
	//	m1->m2->handler->m2 returned->m1 returned.
	for i := len(a.middlewares) - 1; i >= 0; i-- {
		handle = a.middlewares[i](handle)
	}

	// Add a default template method for accessing all of the flash messages
	// in order to make it easier to render them from any template.
	c.render.funcs["flashes"] = func() map[string]interface{} {
		return c.Session().Flashes()
	}

	if err := handle(c); err != nil {
		a.ErrorHandler(c, err)
	}
}

// handle registers the given handler to handle requests at the given path
// with the given HTTP verb.
func (a *App) handle(verb, path string, handle func(c Context) error) {
	switch verb {
	case "HEAD":
		a.mux.Head(path, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			a.serveContext(w, r, handle)
		}))

	case "OPTIONS":
		a.mux.Options(path, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			a.serveContext(w, r, handle)
		}))

	case "GET":
		a.mux.Get(path, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			a.serveContext(w, r, handle)
		}))

	case "POST":
		a.mux.Post(path, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			a.serveContext(w, r, handle)
		}))

	case "PUT":
		a.mux.Put(path, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			a.serveContext(w, r, handle)
		}))

	case "PATCH":
		a.mux.Patch(path, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			a.serveContext(w, r, handle)
		}))

	case "DELETE":
		a.mux.Delete(path, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			a.serveContext(w, r, handle)
		}))

	default:
		panic("method " + verb + " not allowed")
	}
}

// Head routes HEAD requests to the given path.
func (a *App) Head(path string, handle func(c Context) error) {
	a.handle("HEAD", path, handle)
}

// Options routes OPTIONS requests to the given path.
func (a *App) Options(path string, handle func(c Context) error) {
	a.handle("OPTIONS", path, handle)
}

// Get routes GET requests to the given path.
func (a *App) Get(path string, handle func(c Context) error) {
	a.handle("GET", path, handle)
}

// Post routes POST requests to the given path.
func (a *App) Post(path string, handle func(c Context) error) {
	a.handle("POST", path, handle)
}

// Put routes PUT requests to the given path.
func (a *App) Put(path string, handle func(c Context) error) {
	a.handle("PUT", path, handle)
}

// Patch routes PATCH requests to the given path.
func (a *App) Patch(path string, handle func(c Context) error) {
	a.handle("PATCH", path, handle)
}

// Delete routes DELETE requests to the given path.
func (a *App) Delete(path string, handle func(c Context) error) {
	a.handle("DELETE", path, handle)
}

// FileServer serves the contents of the given directory at the given path.
func (a *App) FileServer(path string, dir string) {
	if strings.ContainsAny(path, "{}*") {
		panic("FileServer does not permit URL parameters.")
	}

	fs := http.StripPrefix(path, http.FileServer(http.Dir(dir)))

	if path != "/" && path[len(path)-1] != '/' {
		a.mux.Get(path, http.RedirectHandler(path+"/", http.StatusMovedPermanently).ServeHTTP)
		path += "/"
	}
	path += "*"

	a.mux.Get(path, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fs.ServeHTTP(w, r)
	}))
}

// ServeHTTP makes the Seatbelt application implement the http.Handler
// interface.
func (a *App) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	a.mux.ServeHTTP(w, r)
}
