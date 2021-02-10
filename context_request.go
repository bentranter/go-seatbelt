package seatbelt

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/go-chi/chi"
	"github.com/mitchellh/mapstructure"
)

// Params mass-assigns query, path, and form parameters to the given struct or
// map.
//
// v must be a pointer to a struct or a map.
//
// The precedence is as follows:
//
//	1. Path params (highest).
//	2. Body params.
//	3. Query params.
//
// For POST, PUT, and PATCH requests, the body will be read. For any other
// request, it will not.
func (c *context) Params(v interface{}) error {
	// TODO: Consider if c.r.ParseMultipartForm is prefereable, as it calls
	// ParseForm anyway and runs only once instead of being idempotent.
	if err := c.r.ParseForm(); err != nil {
		return err
	}

	// mapstructure doesn't like the map[string][]string that the query and
	// form data is in, so we turn it into a map[string]string.
	values := make(map[string]interface{})

	// Parse the body query parameters using the built-in Form map, as calling
	// ParseForm() already does what we want to do.
	for key, val := range c.r.Form {
		values[key] = strings.Join(val, "")
	}

	// Parse the JSON body if the content type and HTTP verb correct.
	if c.r.Method == "POST" || c.r.Method == "PUT" || c.r.Method == "PATCH" {
		if c.r.Header.Get("Content-Type") == "application/json" {
			if err := json.NewDecoder(c.r.Body).Decode(&values); err != nil {
				return err
			}
			defer c.r.Body.Close()
		}
	}

	// Finally, overwrite any values with path params.
	if rctx := chi.RouteContext(c.r.Context()); rctx != nil {
		for i, key := range rctx.URLParams.Keys {
			values[key] = rctx.URLParams.Values[i]
		}
	}

	// The config below is the same as mapstructure's `WeakDecode`, but with
	// the tag name "params" instead of "mapstructure".
	config := &mapstructure.DecoderConfig{
		Metadata:         nil,
		Result:           v,
		WeaklyTypedInput: true,
		TagName:          "params",
	}

	decoder, err := mapstructure.NewDecoder(config)
	if err != nil {
		return err
	}

	return decoder.Decode(values)
}

// Request returns the *http.Request for the current Context.
func (c *context) Request() *http.Request {
	return c.r
}

// FormValue returns the form value with the given name.
func (c *context) FormValue(name string) string {
	return c.r.FormValue(name)
}

// QueryParam returns the URL query parameter with the given name.
func (c *context) QueryParam(name string) string {
	return c.r.URL.Query().Get(name)
}
