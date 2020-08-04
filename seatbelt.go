package seatbelt

import (
	"encoding/hex"
	"html/template"
	"net/http"

	"github.com/gorilla/securecookie"
	"github.com/julienschmidt/httprouter"
)

// Params is an alias for httprouter's params.
type Params httprouter.Params

// Context contains values present during the lifetime of an HTTP
// request/response cycle.
type Context struct {
	templates map[string]*template.Template

	Resp    http.ResponseWriter
	Req     *http.Request
	Params  Params
	Session *session
	Flash   *flash
}

// Redirect issues a 302 redirect. If a flash message is provided, the first
// string is the flash key, and the second is the value.
func (c *Context) Redirect(url string, flash ...string) error {
	key := ""
	value := ""

	for i, f := range flash {
		if i%2 == 0 {
			key = f
		} else {
			value = f
		}

		c.Flash.Save(key, value)
	}

	http.Redirect(c.Resp, c.Req, url, http.StatusFound)
	return nil
}

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
	Hash string

	// Block is the block key for creating a secure cookie, used for sessions
	// and flashes.
	Block string
}

// New creates a new instance of an App.
func New(config Config) *App {
	hashKey := config.Hash
	blockKey := config.Block

	if hashKey == "" {
		hashKey = "96f567cab5f00312c562c31156fb7c870e9ac4d560f7bdb7a61e34b2453b9b4155363b313f98c87f8aae9152203a54546aee310cab208e5c09fc6f999414a3d6"
	}
	if blockKey == "" {
		blockKey = "08d611a5f0df41d353c61300d8c28febf864d445126f1ccacfe0fc9db3c00268"
	}

	hash, err := hex.DecodeString(hashKey)
	if err != nil {
		panic(err)
	}
	block, err := hex.DecodeString(blockKey)
	if err != nil {
		panic(err)
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
