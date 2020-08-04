package seatbelt

import (
	"net/http"
	"strings"

	"github.com/gorilla/securecookie"
)

type session struct {
	w    http.ResponseWriter
	r    *http.Request
	s    *securecookie.SecureCookie
	name string
}

// Save saves an arbitrary key-value pair on a session with the given
// cookie name.
func (s *session) Save(key string, value interface{}) {
	values := make(map[string]interface{})

	cookie, err := s.r.Cookie(s.name)
	if err == nil {
		s.s.Decode(s.name, cookie.Value, &values)
	}

	values[key] = value

	encoded, err := s.s.Encode(s.name, values)
	if err != nil {
		return
	}

	http.SetCookie(s.w, &http.Cookie{
		Name:     s.name,
		Value:    encoded,
		Path:     "/",
		Secure:   isTLS(s.r),
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})
}

// Get retrieves the value with the given key from a session.
func (s *session) Get(key string) string {
	cookie, err := s.r.Cookie(s.name)
	if err != nil {
		return ""
	}

	values := make(map[string]interface{})
	if err := s.s.Decode(s.name, cookie.Value, &values); err != nil {
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

// GetSessionInt64 retrieves the value with the given key from a session.
func (s *session) GetSessionInt64(key string) int64 {
	cookie, err := s.r.Cookie(s.name)
	if err != nil {
		return 0
	}

	values := make(map[string]interface{})
	if err := s.s.Decode(s.name, cookie.Value, &values); err != nil {
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

// Delete deletes the key value pair from the session if it exists.
func (s *session) Delete(key string) {
	cookie, err := s.r.Cookie(s.name)
	if err != nil {
		return
	}

	values := make(map[string]interface{})
	if err := s.s.Decode(s.name, cookie.Value, &values); err != nil {
		return
	}

	delete(values, key)

	encoded, err := s.s.Encode(s.name, values)
	if err != nil {
		return
	}

	http.SetCookie(s.w, &http.Cookie{
		Name:     s.name,
		Value:    encoded,
		Path:     "/",
		Secure:   isTLS(s.r),
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})
}

// IsTLS returns true if the request was made over HTTPS.
func (c *Context) IsTLS() bool {
	return isTLS(c.Req)
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
