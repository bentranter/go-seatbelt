package seatbelt

import (
	"bytes"
	"net/http"
	"strings"
	"text/tabwriter"

	"github.com/gertd/go-pluralize"
	"github.com/julienschmidt/httprouter"
)

// handle registers the given handler to handle requests at the given path
// with the given verb.
func (a *App) handle(verb, path string, handle func(c *Context) error) {
	r := parseRoute(verb, path)
	a.routes[r.prefix] = r

	a.router.Handle(verb, path, httprouter.Handle(func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		if err := handle(&Context{
			templates: a.templates,
			Session: &session{
				w:    w,
				r:    r,
				s:    a.cookie,
				name: "_seatbelt_session",
			},
			Flash: &flash{
				w:    w,
				r:    r,
				f:    a.cookie,
				name: "_seatbelt_flash",
			},
			Resp:   w,
			Req:    r,
			Params: Params(ps),
		}); err != nil {
			panic(err)
		}
	}))
}

func (a *App) Head(path string, handle func(c *Context) error) {
	a.handle("HEAD", path, handle)
}

func (a *App) Options(path string, handle func(c *Context) error) {
	a.handle("OPTIONS", path, handle)
}

func (a *App) Get(path string, handle func(c *Context) error) {
	a.handle("GET", path, handle)
}

func (a *App) Post(path string, handle func(c *Context) error) {
	a.handle("POST", path, handle)
}

func (a *App) Put(path string, handle func(c *Context) error) {
	a.handle("PUT", path, handle)
}

func (a *App) Patch(path string, handle func(c *Context) error) {
	a.handle("GET", path, handle)
}

func (a *App) Delete(path string, handle func(c *Context) error) {
	a.handle("GET", path, handle)
}

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
