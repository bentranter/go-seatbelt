package seatbelt

import (
	"net/http"
	"time"

	"github.com/gorilla/securecookie"
)

type flash struct {
	w    http.ResponseWriter
	r    *http.Request
	f    *securecookie.SecureCookie
	name string
}

func (f *flash) Save(key string, value string) {
	values := make(map[string]string)

	cookie, err := f.r.Cookie(f.name)
	if err == nil {
		f.f.Decode(f.name, cookie.Value, &values)
	}

	values[key] = value

	encoded, err := f.f.Encode(f.name, values)
	if err != nil {
		return
	}

	http.SetCookie(f.w, &http.Cookie{
		Name:     f.name,
		Value:    encoded,
		Path:     "/",
		Secure:   isTLS(f.r),
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})
}

func (f *flash) Get(key string) string {
	cookie, err := f.r.Cookie(f.name)
	if err != nil {
		return ""
	}

	values := make(map[string]string)
	if err := f.f.Decode(f.name, cookie.Value, &values); err != nil {
		return ""
	}

	f.expire()

	return values[key]
}

// Flashes returns a map of all flash messages.
func (f *flash) All() map[string]string {
	flashes := make(map[string]string)

	cookie, err := f.r.Cookie(f.name)
	if err != nil {
		return flashes
	}

	if err := f.f.Decode(f.name, cookie.Value, &flashes); err != nil {
		return flashes
	}

	f.expire()

	return flashes
}

func (f *flash) expire() {
	http.SetCookie(f.w, &http.Cookie{
		Name:     f.name,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		Expires:  time.Unix(1, 0),
		Secure:   isTLS(f.r),
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})
}
