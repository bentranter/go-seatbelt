package seatbelt

import (
	"fmt"
	"net/http"

	"github.com/gorilla/sessions"
)

// A Session is a cookie-backed browser session store.
type Session interface {
	// Get returns the value for the given key, if one exists.
	Get(key string) interface{}

	// Put writes a key value pair to the session.
	Put(key string, value interface{})

	// Reset clears and deletes the session.
	Reset()

	// Flash sets a flash value.
	Flash(value interface{})

	// Flashes returns all flash messages.
	Flashes() []interface{}
}

func (c *context) Session() Session {
	return &session{
		r:     c.r,
		w:     c.w,
		name:  "_hussle_session",
		store: c.store,
	}
}

type session struct {
	r     *http.Request
	w     http.ResponseWriter
	name  string
	store sessions.Store
}

// session returns the underlying Gorilla session.
func (s *session) session() *sessions.Session {
	session, err := s.store.Get(s.r, s.name)
	if err != nil {
		fmt.Printf("error getting session from store: %+v\n", err)
		return nil
	}
	return session
}

// Get returns the value for the given key, if one exists.
func (s *session) Get(key string) interface{} {
	v, ok := s.session().Values[key]
	if !ok {
		fmt.Printf("no value exists for key %s\n", key)
		return nil
	}

	return v
}

// Put writes a key value pair to the session.
func (s *session) Put(key string, v interface{}) {
	session := s.session()

	session.Values[key] = v

	if err := session.Save(s.r, s.w); err != nil {
		fmt.Printf("failed to save session: %+v\n", err)
	}
}

// Reset clears and deletes the session.
func (s *session) Reset() {
	session := s.session()
	// Setting the underlying cookie's MaxAge is the "official" way to delete
	// a session, according to the Gorilla docs.
	session.Options.MaxAge = -1
	session.Save(s.r, s.w)
}

// Flash adds a flash message.
func (s *session) Flash(value interface{}) {
	session := s.session()
	session.AddFlash(value)
	session.Save(s.r, s.w)
}

// Flashes returns flash values.
func (s *session) Flashes() []interface{} {
	session := s.session()
	flashes := session.Flashes()
	session.Save(s.r, s.w)
	return flashes
}

// testsession implements a test session object that does not require an HTTP
// request/response cycle to be used. Instead, it uses a map. This should be
// used when writing unit tests.
type testsession struct {
	kv map[string]interface{}
}

func (ts *testsession) Get(key string) interface{} {
	return ts.kv[key]
}

func (ts *testsession) Put(key string, v interface{}) {
	ts.kv[key] = v
}

func (ts *testsession) Reset() {
	ts.kv = make(map[string]interface{})
}

func (ts *testsession) Flash(value interface{}) {
	var flashes []interface{}
	if v, ok := ts.kv["_flash"]; ok {
		flashes = v.([]interface{})
	}
	ts.kv["_flash"] = append(flashes, value)
}

func (ts *testsession) Flashes() []interface{} {
	var flashes []interface{}
	if v, ok := ts.kv["_flash"]; ok {
		delete(ts.kv, "_flash")
		flashes = v.([]interface{})
	}
	return flashes
}
