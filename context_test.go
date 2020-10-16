package seatbelt

import (
	"net/http/httptest"
	"testing"
)

func TestContextSmoke(t *testing.T) {
	const expected = "hello, world!"

	fn := func(c Context) error {
		return c.String(expected)
	}

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", nil)
	c := NewTestContext(w, r, nil)

	if err := fn(c); err != nil {
		t.Fatalf("%+v sending string response", err)
	}
	if actual := w.Body.String(); actual != expected {
		t.Fatalf("expected %s but got %s", expected, actual)
	}
}
