package seatbelt

import (
	sctx "context"
	"html/template"
	"net/http"
	"net/http/httptest"

	"github.com/go-chi/chi"
	"github.com/gorilla/sessions"
)

// Context contains values present during the lifetime of an HTTP
// request/response cycle.
type Context interface {
	// Request returns the *http.Request for the current Context.
	Request() *http.Request

	// Response returns the http.ResponseWriter for the current Context.
	Response() http.ResponseWriter

	// Session returns the session object for the current context.
	Session() Session

	// Params mass-assigns query, path, and form parameters to the given struct or
	// map.
	Params(v interface{}) error

	// FormValue returns the form value with the given name.
	FormValue(name string) string

	// QueryParam returns the URL query parameter with the given name.
	QueryParam(name string) string

	// String sends a string response with the given status code.
	String(code int, s string) error

	// JSON sends a JSON response with the given status code.
	JSON(code int, v interface{}) error

	// Render renders an HTML template.
	Render(name string, data interface{}, opts ...RenderOption) error

	// NoContent sends a 204 No Content HTTP response. The returned error will
	// always be nil.
	NoContent() error

	// Redirect redirects the to the given url. The returned error will always
	// be nil.
	Redirect(url string) error
}

// context implements the Context interface.
type context struct {
	w      http.ResponseWriter
	r      *http.Request
	store  sessions.Store
	render *Renderer
}

// A TestContext is used for unit testing Seatbelt handlers.
//
// A TestContext must be created with `NewTestContext` in order to properly
// initialize the underlying context instance.
type TestContext struct {
	// The underlying context to use in the test.
	*context

	// Underlying in-memory session to use in the test.
	session *testsession

	ResponseRecorder *httptest.ResponseRecorder
	Req              *http.Request
}

// NewTestContext created a new instance of a context suitable for unit
// testing.
func NewTestContext(w http.ResponseWriter, r *http.Request, params ...map[string]string) *TestContext {
	// Set the path parameters, if they are present.
	//
	// The only way to do this is to create a new context with a chi route
	// context on it that has its path params populated. Because on a real
	// chi mux, the path params are stored in the route context, this is the
	// only way to mock path params in unit tests.
	rctx := chi.NewRouteContext()
	for _, param := range params {
		for key, val := range param {
			rctx.URLParams.Add(key, val)
		}
	}
	newCtx := sctx.WithValue(r.Context(), chi.RouteCtxKey, rctx)

	// Force the route context onto the request context.
	r = r.Clone(newCtx)

	tc := &TestContext{
		context: &context{
			w: w,
			r: r,
		},
		Req: r,
		session: &testsession{
			kv: make(map[string]interface{}),
		},
	}

	if rr, ok := w.(*httptest.ResponseRecorder); ok {
		tc.ResponseRecorder = rr
	}

	return tc
}

// AddRenderer adds an instance of a template renderer to a test context
// instance.
func (tc *TestContext) AddRenderer(dir string, funcs template.FuncMap) {
	tc.context.render = NewRenderer(dir, false, funcs)
}

// Session returns a mock session instance, to be used for unit testing.
//
// This overrides the underlying context's session storage.
func (tc *TestContext) Session() Session {
	return tc.session
}
