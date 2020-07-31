package seatbelt

import (
	"net/http"
	"strings"
	"time"
)

// saveSession saves an arbitrary key-value pair on a session with the given
// cookie name.
func (c *Context) saveSession(cookieName, key string, value interface{}) {
	values := make(map[string]interface{})

	// Atempt to decode the existing session key value pairs before creating
	// new ones.
	cookie, err := c.Req.Cookie(cookieName)
	if err == nil {
		c.sessionCookie.Decode(cookieName, cookie.Value, &values)
	}

	values[key] = value

	encoded, err := c.sessionCookie.Encode(cookieName, values)
	if err != nil {
		return
	}

	http.SetCookie(c.Resp, &http.Cookie{
		Name:     cookieName,
		Value:    encoded,
		Path:     "/",
		Secure:   isTLS(c.Req),
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})
}

// SaveSession saves an arbitrary key-value pair on a session.
func (c *Context) SaveSession(key string, value interface{}) {
	c.saveSession(sessionCookieName, key, value)
}

// Flash sets a flash message with the given key and value.
func (c *Context) Flash(key, value string) {
	c.saveSession(flashCookieName, key, value)
}

// GetSession retrieves the value with the given key from a session.
func (c *Context) GetSession(key string) string {
	cookie, err := c.Req.Cookie(sessionCookieName)
	if err != nil {
		return ""
	}

	values := make(map[string]interface{})
	if err := c.sessionCookie.Decode(sessionCookieName, cookie.Value, &values); err != nil {
		return ""
	}

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

// Flashes returns a map of all flash messages.
func (c *Context) Flashes() map[string]string {
	flashes := make(map[string]string)

	cookie, err := c.Req.Cookie(flashCookieName)
	if err != nil {
		return flashes
	}

	values := make(map[string]interface{})
	if err := c.flashCookie.Decode(flashCookieName, cookie.Value, &values); err != nil {
		return flashes
	}

	for k, v := range values {
		flashes[k] = v.(string)
	}

	http.SetCookie(c.Resp, &http.Cookie{
		Name:     flashCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		Expires:  time.Unix(1, 0),
		Secure:   isTLS(c.Req),
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})

	return flashes
}

// GetSessionInt64 retrieves the value with the given key from a session.
func (c *Context) GetSessionInt64(key string) int64 {
	cookie, err := c.Req.Cookie(sessionCookieName)
	if err != nil {
		return 0
	}

	values := make(map[string]interface{})
	if err := c.sessionCookie.Decode(sessionCookieName, cookie.Value, &values); err != nil {
		return 0
	}

	value, ok := values[key]
	if !ok {
		return 0
	}

	num, ok := value.(int64)
	if !ok {
		return 0
	}

	return num
}

// DeleteSession deletes the key value pair from the session if it exists.
func (c *Context) DeleteSession(key string) {
	cookie, err := c.Req.Cookie(sessionCookieName)
	if err != nil {
		return
	}

	values := make(map[string]interface{})
	if err := c.sessionCookie.Decode(sessionCookieName, cookie.Value, &values); err != nil {
		return
	}

	delete(values, key)

	encoded, err := c.sessionCookie.Encode(sessionCookieName, values)
	if err != nil {
		return
	}

	http.SetCookie(c.Resp, &http.Cookie{
		Name:     sessionCookieName,
		Value:    encoded,
		Path:     "/",
		Secure:   isTLS(c.Req),
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})
}

// isTLS is a helper to check if a request was performed over HTTPS.
func isTLS(r *http.Request) bool {
	if r.TLS != nil {
		return true
	}
	if strings.ToLower(r.Header.Get("X-Forwarded-Proto")) == "https" {
		return true
	}
	return false
}
