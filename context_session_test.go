package seatbelt_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bentranter/go-seatbelt"
)

func TestContextSession(t *testing.T) {
	// A reference to the message stored in the session, so that we can check
	// it later in the tests.
	var message string

	// Handlers to be executed during the test suite.
	var (
		put = func(c seatbelt.Context) error {
			msg := c.QueryParam("msg")
			c.Session().Put("msg", msg)
			return c.NoContent()
		}

		get = func(c seatbelt.Context) error {
			v := c.Session().Get("msg")

			if m, ok := v.(string); ok {
				message = m
			} else {
				message = ""
			}

			return c.NoContent()
		}

		reset = func(c seatbelt.Context) error {
			c.Session().Reset()
			return c.NoContent()
		}
	)

	app := seatbelt.New()

	app.Get("/", get)
	app.Put("/", put)
	app.Delete("/", reset)

	srv := httptest.NewServer(app)
	defer srv.Close()

	// Save a reference to the cookie so we can persist it between requests.
	var cookie *http.Cookie

	const expectedMsg = "ok"

	t.Run("put session", func(t *testing.T) {
		req, err := http.NewRequest("PUT", srv.URL+"/?msg="+expectedMsg, nil)
		if err != nil {
			t.Fatalf("%+v creating request", err)
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("%+v executing request", err)
		}

		cookies := resp.Cookies()
		if len(cookies) == 0 {
			t.Fatal("expected at least one cookie to bet set but got zero")
		}

		for _, c := range cookies {
			if c.Name == "_hussle_session" {
				cookie = c
			}
		}

		if cookie == nil {
			t.Fatalf("cookie must be set before tests can continue")
		}
	})

	t.Run("get session should return the expected message", func(t *testing.T) {
		req, err := http.NewRequest("GET", srv.URL+"/", nil)
		if err != nil {
			t.Fatalf("%+v creating request", err)
		}

		req.AddCookie(cookie)

		if _, err := http.DefaultClient.Do(req); err != nil {
			t.Fatalf("%+v creating request", err)
		}

		if message != expectedMsg {
			t.Fatalf("expected %s but got %s", expectedMsg, message)
		}
	})

	t.Run("reset session", func(t *testing.T) {
		req, err := http.NewRequest("DELETE", srv.URL+"/", nil)
		if err != nil {
			t.Fatalf("%+v creating request", err)
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("%+v executing request", err)
		}

		cookies := resp.Cookies()
		if len(cookies) == 0 {
			t.Fatal("expected at least one cookie to bet set but got zero")
		}

		for _, c := range cookies {
			if c.Name == "_hussle_session" {
				cookie = c
			}
		}
	})

	t.Run("get session should return an empty string", func(t *testing.T) {
		req, err := http.NewRequest("GET", srv.URL+"/", nil)
		if err != nil {
			t.Fatalf("%+v creating request", err)
		}

		req.AddCookie(cookie)

		if _, err := http.DefaultClient.Do(req); err != nil {
			t.Fatalf("%+v creating request", err)
		}

		if message != "" {
			t.Fatalf("expected an empty string but got %s", message)
		}
	})
}
