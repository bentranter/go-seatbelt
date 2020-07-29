package seatbelt

import (
	"html/template"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

// Params is an alias for httprouter's params.
type Params httprouter.Params

// Context contains values present during the lifetime of an HTTP
// request/response cycle.
type Context struct {
	templates map[string]*template.Template

	Resp   http.ResponseWriter
	Req    *http.Request
	Params Params
}

// An App is contains the data necessary to start and run an application.
type App struct {
	// Dependencies for populating a Context on each request.
	templates map[string]*template.Template

	// App specific dependencies.
	middleware []func(http.Handler) http.Handler
	router     *httprouter.Router
	routes     map[string]route
}

// A Config is used to configure an App.
type Config struct {
	// Dir is the directory containing your Go HTML templates.
	Dir string

	// Test defines whether or not to start the app in test mode.
	Test bool
}

// New creates a new instance of an App.
func New(config Config) *App {
	return &App{
		templates:  parseTemplates(config.Dir),
		middleware: make([]func(http.Handler) http.Handler, 0),
		router:     httprouter.New(),
		routes:     make(map[string]route),
	}
}

// Start starts the app on the given address.
func (a *App) Start(addr string) error {
	return http.ListenAndServe(addr, a.router)
}
