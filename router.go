package seatbelt

import (
	"encoding/hex"
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
	store      sessions.Store
	mux        chi.Router
	render     *Renderer
	signingKey []byte
}

// An Option is used to configure a Seatbelt application.
type Option struct {
	TemplateDir string           // The directory where the templates reside.
	SigningKey  string           // The signing key for the cookie session store.
	Reload      bool             // Whether or not to reload templates on each request.
	Funcs       template.FuncMap // HTML functions.
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

	signingKey, err := hex.DecodeString(opt.SigningKey)
	if err != nil {
		log.Fatalf("seatbelt: signing key is not a valid hexadecimal string: %+v", err)
	}

	cookieStore := sessions.NewCookieStore([]byte(opt.SigningKey))

	// Set secure defaults for the session cookie store.
	cookieStore.Options.HttpOnly = true
	cookieStore.Options.SameSite = http.SameSiteStrictMode
	cookieStore.Options.Secure = true

	// Initialize the underlying chi mux so that we can setup our default
	// middleware stack.
	mux := chi.NewRouter()
	mux.Use(csrf.Protect(signingKey))

	return &App{
		mux:        chi.NewRouter(),
		store:      cookieStore,
		render:     NewRenderer(opt.TemplateDir, opt.Reload, opt.Funcs),
		signingKey: signingKey,
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

// Use registers standard HTTP middleware on the application.
func (a *App) Use(middleware ...func(http.Handler) http.Handler) {
	a.mux.Use(middleware...)
}

// handle registers the given handler to handle requests at the given path
// with the given HTTP verb.
func (a *App) handle(verb, path string, handle func(c Context) error) {
	switch verb {
	case "HEAD":
		a.mux.Head(path, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			c := &context{w: w, r: r, store: a.store, render: a.render}
			if err := handle(c); err != nil {
				log.Printf("seatbelt: error in HTTP handler: %+v", err)
				c.w.Write([]byte(err.Error()))
			}
		}))

	case "OPTIONS":
		a.mux.Options(path, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			c := &context{w: w, r: r, store: a.store, render: a.render}
			if err := handle(c); err != nil {
				log.Printf("seatbelt: error in HTTP handler: %+v", err)
				c.w.Write([]byte(err.Error()))
			}
		}))

	case "GET":
		a.mux.Get(path, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			c := &context{w: w, r: r, store: a.store, render: a.render}
			if err := handle(c); err != nil {
				log.Printf("seatbelt: error in HTTP handler: %+v", err)
				c.w.Write([]byte(err.Error()))
			}
		}))

	case "POST":
		a.mux.Post(path, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			c := &context{w: w, r: r, store: a.store, render: a.render}
			if err := handle(c); err != nil {
				log.Printf("seatbelt: error in HTTP handler: %+v", err)
				c.w.Write([]byte(err.Error()))
			}
		}))

	case "PUT":
		a.mux.Put(path, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			c := &context{w: w, r: r, store: a.store, render: a.render}
			if err := handle(c); err != nil {
				log.Printf("seatbelt: error in HTTP handler: %+v", err)
				c.w.Write([]byte(err.Error()))
			}
		}))

	case "PATCH":
		a.mux.Patch(path, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			c := &context{w: w, r: r, store: a.store, render: a.render}
			if err := handle(c); err != nil {
				log.Printf("seatbelt: error in HTTP handler: %+v", err)
				c.w.Write([]byte(err.Error()))
			}
		}))

	case "DELETE":
		a.mux.Delete(path, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			c := &context{w: w, r: r, store: a.store, render: a.render}
			if err := handle(c); err != nil {
				log.Printf("seatbelt: error in HTTP handler: %+v", err)
				c.w.Write([]byte(err.Error()))
			}
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
		a.mux.Get(path, http.RedirectHandler(path+"/", 301).ServeHTTP)
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
