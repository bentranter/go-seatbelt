package seatbelt

import (
	"errors"
	"net/http"

	"github.com/gorilla/sessions"
	"github.com/rs/zerolog/log"
)

// ErrKeyNotFound occurs when trying to access a value for a key that doesn't
// exist in the session map.
var ErrKeyNotFound = errors.New("value not found for key in session")

// A Session is a cookie-backed browser session store.
type Session interface {
	// Get returns the value for the given key, if one exists.
	Get(key string) interface{}

	// Put writes a key value pair to the session.
	Put(key string, value interface{})

	// Del deletes the value with the given key, if one exists.
	Del(key string)

	// Reset clears and deletes the session.
	Reset()

	// Flash sets a flash message with the given key.
	Flash(key string, value interface{})

	// GetFlash returns the flash message with the given key.
	GetFlash(key string) (interface{}, bool)

	// Flashes returns all flash messages.
	Flashes() map[string]interface{}
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
		log.Error().Err(err).Msg("failed to get session from store, creating new session")
		return sessions.NewSession(s.store, s.name)
	}
	return session
}

// Get returns the value for the given key, if one exists.
func (s *session) Get(key string) interface{} {
	v, ok := s.session().Values[key]
	if !ok {
		log.Warn().Err(ErrKeyNotFound).Str("key", key).Msg("no session value exists for key")
		return nil
	}
	return v
}

// Put writes a key value pair to the session.
func (s *session) Put(key string, v interface{}) {
	session := s.session()

	if session.Values == nil {
		session.Values = make(map[interface{}]interface{})
	}

	session.Values[key] = v

	if err := session.Save(s.r, s.w); err != nil {
		log.Error().Err(err).Msg("failed to save session")
	}
}

// Del deletes a key value pair from the session.
func (s *session) Del(key string) {
	session := s.session()

	delete(session.Values, key)

	if err := session.Save(s.r, s.w); err != nil {
		log.Error().Err(err).Msg("failed to delete from session")
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

// Flash adds a flash message with the given key.
func (s *session) Flash(key string, value interface{}) {
	session := s.session()

	var flashMap map[string]interface{}

	// Check if there is an existing flash map before writing to it, so that
	// we're not overwriting existing flashes within the same context.
	flashes := session.Flashes()
	if len(flashes) > 0 {
		if m, ok := flashes[0].(map[string]interface{}); ok {
			flashMap = m
		} else {
			flashMap = make(map[string]interface{})
		}
	} else {
		flashMap = make(map[string]interface{})
	}

	flashMap[key] = value

	session.AddFlash(flashMap)

	if err := session.Save(s.r, s.w); err != nil {
		log.Error().Err(err).Str("key", key).Interface("value", value).Msg("failed to save flash in session")
	}
}

// GetFlash returns the flash message with the given key, if one exists. If
// one does not, the returned boolean will be false, otherwise it is true.
func (s *session) GetFlash(key string) (interface{}, bool) {
	session := s.session()
	flashes := session.Flashes()

	if len(flashes) < 1 {
		return nil, false
	}

	flashMap, ok := flashes[0].(map[string]interface{})
	if !ok {
		return nil, false
	}

	value, ok := flashMap[key]
	return value, ok
}

// Flashes returns all flash messages values.
func (s *session) Flashes() map[string]interface{} {
	session := s.session()
	flashes := session.Flashes()

	if len(flashes) < 1 {
		return nil
	}

	flashMap, ok := flashes[0].(map[string]interface{})
	if !ok {
		return nil
	}

	if err := session.Save(s.r, s.w); err != nil {
		log.Error().Err(err).Msg("failed to save flashes in session")
	}

	return flashMap
}

// testsession implements a test session object that does not require an HTTP
// request/response cycle to be used. Instead, it uses a map. This should be
// used when writing unit tests.
type testsession struct {
	kv       map[string]interface{}
	flashMap map[string]interface{}
}

func (ts *testsession) Get(key string) interface{} {
	return ts.kv[key]
}

func (ts *testsession) Put(key string, v interface{}) {
	ts.kv[key] = v
}

func (ts *testsession) Del(key string) {
	delete(ts.kv, key)
}

func (ts *testsession) Reset() {
	ts.kv = make(map[string]interface{})
}

func (ts *testsession) Flash(key string, value interface{}) {
	ts.flashMap[key] = value
}

func (ts *testsession) GetFlash(key string) (interface{}, bool) {
	value, ok := ts.flashMap[key]
	return value, ok
}

func (ts *testsession) Flashes() map[string]interface{} {
	return ts.flashMap
}
