package seatbelt_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bentranter/go-seatbelt"
)

func TestRouter(t *testing.T) {
	app := seatbelt.New()

	fn := func(c seatbelt.Context) error {
		return c.String(200, "ok")
	}

	app.Head("/", fn)
	app.Options("/", fn)
	app.Get("/", fn)
	app.Post("/", fn)
	app.Put("/", fn)
	app.Patch("/", fn)
	app.Delete("/", fn)

	srv := httptest.NewServer(app)
	defer srv.Close()

	idempotentCases := []string{
		"HEAD",
		"OPTIONS",
		"GET",
	}
	protectedCases := []string{
		"POST",
		"PUT",
		"PATCH",
		"DELETE",
	}

	for _, c := range idempotentCases {
		t.Run(c, func(t *testing.T) {
			req, err := http.NewRequest(c, srv.URL+"/", nil)
			if err != nil {
				t.Fatalf("%+v", err)
			}

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Fatalf("%+v for %s", err, c)
			}
			if resp.StatusCode != 200 {
				t.Fatalf("expected 200 but got %d", resp.StatusCode)
			}
		})
	}

	for _, c := range protectedCases {
		t.Run(c, func(t *testing.T) {
			req, err := http.NewRequest(c, srv.URL+"/", nil)
			if err != nil {
				t.Fatalf("%+v", err)
			}

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Fatalf("%+v for %s", err, c)
			}
			if resp.StatusCode != 403 {
				t.Fatalf("expected 403 but got %d", resp.StatusCode)
			}
		})
	}
}
