package seatbelt

import (
	"errors"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gorilla/securecookie"
	"github.com/mitchellh/mapstructure"
)

type Contexter interface {
	// session methods
	Get(key string) interface{}
	GetInt64(key string) int64
	Set(key string, value interface{})
	Delete(key string)

	// flash methods
	Flash(level string, message string)
	Flashes() map[string]string

	// http convenience
	Request() *http.Request
	Header() http.Header
	Response() *http.Response
	IsTLS() bool
	IP() string
	Params(v interface{}) error
	PathParam(name string) string
	QueryParam(name string) string
	Form(name string) string
	FormFile(name string) (*multipart.FileHeader, error)

	// http response convenience
	Render(status int, name string, data interface{}) error
	Redirect(url string, flash ...string) error

	// i18n
	T(key string) string
}

const (
	sessionCookieName = "_seatbelt_session"
	flashCookieName   = "_seatbelt_flash"
)

// Context contains values present during the lifetime of an HTTP
// request/response cycle.
type Context struct {
	w http.ResponseWriter
	r *http.Request
	p Params
	t map[string]*template.Template
	s *securecookie.SecureCookie
}

// NewTestContext is used to create Contexts for testing. These cannot be used
// in a real application.
func NewTestContext(method, path string, body io.Reader, ps Params, t map[string]*template.Template) *Context {
	return &Context{
		w: httptest.NewRecorder(),
		r: httptest.NewRequest("GET", "/", nil),
		s: securecookie.New(
			securecookie.GenerateRandomKey(64),
			securecookie.GenerateRandomKey(32),
		),
	}
}

// set encodes and saves a value with the given key in the cookie with the
// given name.
//
// This serves as a convenience function for saving both session and flash
// values.
func (c *Context) set(name string, key string, value interface{}) {
	values := make(map[string]interface{})

	cookie, err := c.r.Cookie(name)
	if err == nil {
		c.s.Decode(name, cookie.Value, &values)
	}

	values[key] = value

	encoded, err := c.s.Encode(name, values)
	if err != nil {
		return
	}

	http.SetCookie(c.w, &http.Cookie{
		Name:     name,
		Value:    encoded,
		Path:     "/",
		Secure:   c.IsTLS(),
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})
}

// Set saves a session value on the given key.
func (c *Context) Set(key string, value interface{}) {
	c.set(sessionCookieName, key, value)
}

// get returns the map of key-value pairs in the cookie with the given name.
func (c *Context) get(name string) map[string]interface{} {
	values := make(map[string]interface{})

	cookie, err := c.r.Cookie(name)
	if err != nil {
		return values
	}

	c.s.Decode(name, cookie.Value, &values)
	return values
}

// Get gets the value associated with the given key from the session.
func (c *Context) Get(key string) string {
	values := c.get(sessionCookieName)

	value, ok := values[key]
	if !ok {
		return ""
	}

	str, ok := value.(string)
	if !ok {
		return ""
	}

	return str
}

// GetInt64 gets the value associated with the given key from the session.
func (c *Context) GetInt64(key string) int64 {
	values := c.get(sessionCookieName)

	value, ok := values[key]
	if !ok {
		return 0
	}

	i, ok := value.(int64)
	if !ok {
		return 0
	}

	return i
}

// Delete deletes the key and its associated value from the session.
func (c *Context) Delete(key string) {
	cookie, err := c.r.Cookie(sessionCookieName)
	if err != nil {
		return
	}

	values := make(map[string]interface{})
	if err := c.s.Decode(sessionCookieName, cookie.Value, &values); err != nil {
		return
	}

	delete(values, key)

	encoded, err := c.s.Encode(sessionCookieName, values)
	if err != nil {
		return
	}

	http.SetCookie(c.w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    encoded,
		Path:     "/",
		Secure:   c.IsTLS(),
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})
}

// Flash saves a flash message.
func (c *Context) Flash(level string, message string) {
	c.set(flashCookieName, level, message)
}

// Flashes returns all flash message key value pairs.
//
// Flashes is accessible within a template as `{{ flashes }}`.
func (c *Context) Flashes() map[string]string {
	values := make(map[string]interface{})
	flashes := make(map[string]string)

	cookie, err := c.r.Cookie(flashCookieName)
	if err != nil {
		return flashes
	}

	if err := c.s.Decode(flashCookieName, cookie.Value, &values); err != nil {
		return flashes
	}

	for k, v := range values {
		flashes[k] = v.(string)
	}

	http.SetCookie(c.w, &http.Cookie{
		Name:     flashCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		Expires:  time.Unix(1, 0),
		Secure:   c.IsTLS(),
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})

	return flashes
}

// Request returns the HTTP request.
func (c *Context) Request() *http.Request {
	return c.r
}

// Header returns the HTTP header for this request.
func (c *Context) Header() http.Header {
	return c.r.Header
}

// Response returns the HTTP response writer.
func (c *Context) Response() http.ResponseWriter {
	return c.w
}

// IP returns the IP address of the incoming request.
func (c *Context) IP() string {
	if ip := c.r.Header.Get("X-Forwarded-For"); ip != "" {
		return strings.Split(ip, ", ")[0]
	}
	if ip := c.r.Header.Get("X-Real-IP"); ip != "" {
		return ip
	}
	ra, _, _ := net.SplitHostPort(c.r.RemoteAddr)
	return ra
}

// IsTLS is a helper to check if a request was performed over HTTPS.
func (c *Context) IsTLS() bool {
	if c.r.TLS != nil {
		return true
	}
	if strings.ToLower(c.r.Header.Get("X-Forwarded-Proto")) == "https" {
		return true
	}
	return false
}

// Params mass-assigns query, path, and form parameters to the given struct or
// map.
//
// v must be a pointer to a map or struct.
func (c *Context) Params(v interface{}) error {
	if err := c.r.ParseForm(); err != nil {
		return err
	}

	// mapstructure doesn't like the map[string][]string that the form data is
	// in, so turn it into a map[string]string.
	values := make(map[string]interface{})
	for k, v := range c.r.Form {
		values[k] = strings.Join(v, "")
	}

	// Overwrite any form values with path params.
	for _, ps := range c.p {
		if _, ok := values[ps.Key]; ok {
			values[ps.Key] = ps.Value
		}
	}

	return mapstructure.WeakDecode(values, v)
}

func parseTemplates(dir string) map[string]*template.Template {
	b, err := ioutil.ReadFile(filepath.Join(dir, "layout.html"))
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	templates := make(map[string]*template.Template)

	layout, err := template.New("layout").
		Funcs(template.FuncMap{
			"flashes": func() map[string]string {
				return make(map[string]string)
			},
		}).
		Parse(string(b))
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		// Fix same-extension-dirs bug: some dir might be named to:
		// "users.tmpl", "local.html". These dirs should be excluded as they
		// are not valid golang templates, but files under them should be
		// treat as normal. If is a dir, return immediately (dir is not a
		// valid golang template).
		if info == nil || info.IsDir() {
			return nil
		}

		rel, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}

		ext := ""
		if strings.Index(rel, ".") != -1 {
			ext = filepath.Ext(rel)
		}

		buf, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}

		name := (rel[0 : len(rel)-len(ext)])

		tmpl, err := template.Must(layout.Clone()).Parse(string(buf))
		if err != nil {
			return err
		}

		templates[filepath.ToSlash(name)] = tmpl
		return nil
	}); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	return templates
}

// String writes a string.
func (c *Context) String(s string) error {
	_, err := c.w.Write([]byte(s))
	return err
}

// Render renders the HTML template with the given data, if any.
func (c *Context) Render(status int, name string, data interface{}) error {
	tmpl, ok := c.t[name]
	if !ok {
		return errors.New(`template "` + name + `" is not defined`)
	}

	flashes := c.Flashes()

	// Override previous func map because Go's templates are weird.
	tmpl.Funcs(template.FuncMap{
		"flashes": func() map[string]string {
			return flashes
		},
	})

	c.w.WriteHeader(status)
	return tmpl.ExecuteTemplate(c.w, "layout", data)
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
	}

	if key != "" && value != "" {
		c.Flash(key, value)
	}

	http.Redirect(c.w, c.r, url, http.StatusFound)
	return nil
}
