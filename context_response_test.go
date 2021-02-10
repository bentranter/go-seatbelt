package seatbelt_test

import (
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/bentranter/go-seatbelt"
)

func TestContextResponse(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", nil)

	c := seatbelt.NewTestContext(w, r, nil)

	fn := func(c seatbelt.Context) error {
		if !reflect.DeepEqual(w, c.Response()) {
			t.Fatalf("expected %+v and %+v to be equal", r, w)
		}

		return c.String(200, "ok")
	}

	if err := fn(c); err != nil {
		t.Fatalf("%+v executing handler", err)
	}
}

func TestContextString(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", nil)

	c := seatbelt.NewTestContext(w, r, nil)

	fn := func(c seatbelt.Context) error {
		return c.String(200, "ok")
	}

	if err := fn(c); err != nil {
		t.Fatalf("%+v executing handler", err)
	}

	if status := c.ResponseRecorder.Code; status != 200 {
		t.Fatalf("expected HTTP 200 but got %d", status)
	}
	if body := c.ResponseRecorder.Body.String(); body != "ok" {
		t.Fatalf("expected body to be ok but got %s", body)
	}
}

func TestContextNoContent(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", nil)

	c := seatbelt.NewTestContext(w, r, nil)

	fn := func(c seatbelt.Context) error {
		return c.NoContent()
	}

	if err := fn(c); err != nil {
		t.Fatalf("%+v executing handler", err)
	}

	if status := c.ResponseRecorder.Code; status != 204 {
		t.Fatalf("expected HTTP 204 but got %d", status)
	}
}
