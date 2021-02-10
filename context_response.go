package seatbelt

import (
	"encoding/json"
	"net/http"
)

// Response returns the http.ResponseWriter for the current Context.
func (c *context) Response() http.ResponseWriter {
	return c.w
}

// String sends a string response with the given status code.
func (c *context) String(code int, s string) error {
	c.w.Header().Set("Content-Type", "text/plain")
	c.w.WriteHeader(code)
	_, err := c.w.Write([]byte(s))
	return err
}

// JSON sends a JSON response with the given status code.
func (c *context) JSON(code int, v interface{}) error {
	c.w.Header().Set("Content-Type", "application/json")
	c.w.WriteHeader(code)

	data, err := json.Marshal(v)
	if err != nil {
		return err
	}

	_, err = c.w.Write(data)
	return err
}

// NoContent sends a 204 No Content HTTP response. It will always return a nil
// error.
func (c *context) NoContent() error {
	c.w.WriteHeader(204)
	return nil
}

// Render renders an HTML template.
func (c *context) Render(name string, data interface{}, opts ...RenderOption) error {
	return c.render.HTML(c.w, c.r, name, data, opts...)
}

// Redirect redirects the to the given url. It will never return an error.
func (c *context) Redirect(url string) error {
	code := http.StatusFound

	if c.r.Method == http.MethodPost ||
		c.r.Method == http.MethodPut ||
		c.r.Method == http.MethodPatch ||
		c.r.Method == http.MethodDelete {
		code = http.StatusSeeOther
	}

	http.Redirect(c.w, c.r, url, code)
	return nil
}
