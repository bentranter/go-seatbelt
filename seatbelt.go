package seatbelt

import (
	"bytes"
	"html/template"
	"net/http"
	"strings"
	"text/tabwriter"

	"github.com/gertd/go-pluralize"
	"github.com/gorilla/securecookie"
	"github.com/julienschmidt/httprouter"
)

// Params is an alias for httprouter's params.
type Params httprouter.Params

// An App is contains the data necessary to start and run an application.
type App struct {
	// Dependencies for populating a Context on each request.
	templates map[string]*template.Template
	cookie    *securecookie.SecureCookie

	// App specific dependencies.
	middleware []func(http.Handler) http.Handler
	router     *httprouter.Router
	routes     map[string]route
}

// A Config is used to configure an App.
type Config struct {
	// Dir is the directory containing your Go HTML templates.
	Dir string

	// Hash is a the hash for creating a secure cookie, used for sessions and
	// flashes.
	Hash []byte

	// Block is the block key for creating a secure cookie, used for sessions
	// and flashes.
	Block []byte
}

// New creates a new instance of an App.
func New(config Config) *App {
	hash := config.Hash
	block := config.Block

	if hash == nil {
		hash = securecookie.GenerateRandomKey(64)
	}
	if block == nil {
		block = securecookie.GenerateRandomKey(32)
	}

	return &App{
		templates:  parseTemplates(config.Dir),
		cookie:     securecookie.New(hash, block),
		middleware: make([]func(http.Handler) http.Handler, 0),
		router:     httprouter.New(),
		routes:     make(map[string]route),
	}
}

// Start starts the app on the given address.
func (a *App) Start(addr string) error {
	return http.ListenAndServe(addr, a.router)
}

// handle registers the given handler to handle requests at the given path
// with the given verb.
func (a *App) handle(verb, path string, handle func(c *Context) error) {
	r := parseRoute(verb, path)
	a.routes[r.prefix] = r

	a.router.Handle(verb, path, httprouter.Handle(func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		if err := handle(&Context{
			w: w,
			r: r,
			t: a.templates,
			p: Params(ps),
			s: a.cookie,
		}); err != nil {
			panic(err)
		}
	}))
}

// Head routes HEAD requests to the given path.
func (a *App) Head(path string, handle func(c *Context) error) {
	a.handle("HEAD", path, handle)
}

// Options routes OPTIONS requests to the given path.
func (a *App) Options(path string, handle func(c *Context) error) {
	a.handle("OPTIONS", path, handle)
}

// Get routes GET requests to the given path.
func (a *App) Get(path string, handle func(c *Context) error) {
	a.handle("GET", path, handle)
}

// Post routes POST requests to the given path.
func (a *App) Post(path string, handle func(c *Context) error) {
	a.handle("POST", path, handle)
}

// Put routes PUT requests to the given path.
func (a *App) Put(path string, handle func(c *Context) error) {
	a.handle("PUT", path, handle)
}

// Patch routes PATCH requests to the given path.
func (a *App) Patch(path string, handle func(c *Context) error) {
	a.handle("GET", path, handle)
}

// Delete routes DELETE requests to the given path.
func (a *App) Delete(path string, handle func(c *Context) error) {
	a.handle("GET", path, handle)
}

// Routes returns a human readable string containing all routes.
func (a *App) Routes() string {
	buf := &bytes.Buffer{}
	w := tabwriter.NewWriter(buf, 1, 4, 1, ' ', 0)

	w.Write([]byte("Prefix\t"))
	w.Write([]byte("Verb\t"))
	w.Write([]byte("URI Pattern\n"))

	for _, r := range a.routes {
		w.Write([]byte(r.prefix))
		w.Write([]byte("\t"))
		w.Write([]byte(r.verb))
		w.Write([]byte("\t"))
		w.Write([]byte(r.pattern))
		w.Write([]byte("\n"))
	}

	if err := w.Flush(); err != nil {
		panic(err)
	}

	return buf.String()
}

var inflect = pluralize.NewClient()

// A route is a Rails-style definition of a route.
type route struct {
	// The prefix is the Rails-style key of the URL path, ie, `root_path`
	prefix string

	// The HTTP verb for this route.
	verb string

	// The pattern is the URL pattern of the route.
	pattern string
}

// parseRoute generates the Rails-style route prefix from an HTTP verb and a
// URL path.
func parseRoute(verb, pattern string) route {
	if pattern == "/" {
		return route{
			prefix:  "root",
			verb:    verb,
			pattern: pattern,
		}
	}

	pattern = strings.TrimPrefix(pattern, "/")
	pattern = strings.TrimSuffix(pattern, "/")

	paths := strings.Split(pattern, "/")
	words := make([]string, 0)

	last := len(paths) - 1

	for i, p := range paths {
		// If this is the last element, append it.
		if i == last {
			if inflect.IsSingular(p) {
				words = prepend(words, p)
				continue
			}
			words = append(words, p)
			continue
		}

		if p == ":id" {
			continue
		}

		if inflect.IsPlural(p) {
			words = append(words, inflect.Singular(p))
			continue
		}

		words = append(words, p)
	}

	return route{
		prefix:  strings.Join(words, "_"),
		verb:    verb,
		pattern: pattern,
	}
}

// prepend prepends the given string to the given slice.
func prepend(slice []string, s string) []string {
	return append([]string{s}, slice...)
}
