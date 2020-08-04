package seatbelt

import (
	"net/http/httptest"
	"testing"

	"github.com/gorilla/securecookie"
)

func TestContextFlash(t *testing.T) {
	f := &flash{
		w:    httptest.NewRecorder(),
		r:    httptest.NewRequest("GET", "/", nil),
		name: "_test_flash",
		f: securecookie.New(
			securecookie.GenerateRandomKey(32),
			securecookie.GenerateRandomKey(64),
		),
	}

	t.Run("set and get a flash message", func(t *testing.T) {
		const k = "notice"
		const v = "Test message"

		f.Save(k, v)

		if actual := f.Get(k); actual != v {
			t.Fatalf("expected %s but got %s", v, actual)
		}
	})

	t.Run("get all flash messages", func(t *testing.T) {
		const k = "notice"
		const v = "Test message"

		f.Save(k, v)

		flashes := f.All()

		actual, ok := flashes[k]
		if !ok {
			t.Fatalf("failed to save flash message")
		}
		if actual != v {
			t.Fatalf("expected %s but got %s", v, actual)
		}
	})
}
