package seatbelt

import (
	"encoding/hex"
	"net/http"
	"strings"

	"github.com/gorilla/securecookie"
)

// A session stores key value pairs on an HTTP cookie.
type session struct {
	sc   *securecookie.SecureCookie
	name string
}

// NewSession creates a new instance of a session store.
//
// TODO Obviously don't hardcode secret session keys!
func newSession() *session {
	const hashKey = "96f567cab5f00312c562c31156fb7c870e9ac4d560f7bdb7a61e34b2453b9b4155363b313f98c87f8aae9152203a54546aee310cab208e5c09fc6f999414a3d6"
	const blockKey = "08d611a5f0df41d353c61300d8c28febf864d445126f1ccacfe0fc9db3c00268"

	hash, err := hex.DecodeString(hashKey)
	if err != nil {
		panic(err)
	}
	block, err := hex.DecodeString(blockKey)
	if err != nil {
		panic(err)
	}

	return &session{
		sc:   securecookie.New(hash, block),
		name: "_megandoodle_session",
	}
}

// Save saves an arbitrary key-value pair on a session.
func (s *session) Save(w http.ResponseWriter, r *http.Request, key string, value interface{}) {
	values := make(map[string]interface{})

	// Atempt to decode the existing session key value pairs before creating
	// new ones.
	cookie, err := r.Cookie(s.name)
	if err == nil {
		s.sc.Decode(s.name, cookie.Value, &values)
	}

	values[key] = value

	encoded, err := s.sc.Encode(s.name, values)
	if err != nil {
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     s.name,
		Value:    encoded,
		Path:     "/",
		Secure:   IsTLS(r),
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})
}

// Get retrieves the value with the given key from a session.
func (s *session) Get(r *http.Request, key string) string {
	cookie, err := r.Cookie(s.name)
	if err != nil {
		return ""
	}

	values := make(map[string]interface{})
	if err := s.sc.Decode(s.name, cookie.Value, &values); err != nil {
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

// GetInt64 retrieves the value with the given key from a session.
func (s *session) GetInt64(r *http.Request, key string) int64 {
	cookie, err := r.Cookie(s.name)
	if err != nil {
		return 0
	}

	values := make(map[string]interface{})
	if err := s.sc.Decode(s.name, cookie.Value, &values); err != nil {
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
func (s *session) Delete(w http.ResponseWriter, r *http.Request, key string) {
	cookie, err := r.Cookie(s.name)
	if err != nil {
		return
	}

	values := make(map[string]interface{})
	if err := s.sc.Decode(s.name, cookie.Value, &values); err != nil {
		return
	}

	delete(values, key)

	encoded, err := s.sc.Encode(s.name, values)
	if err != nil {
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     s.name,
		Value:    encoded,
		Path:     "/",
		Secure:   isTLS(r),
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
