package seatbelt_test

import (
	"io"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/bentranter/go-seatbelt"
)

// teststruct is used for table-testing the Params method.
type teststruct struct {
	PathParam string
	URLParam  string
	JSONParam string
}

func TestContextRequestParams(t *testing.T) {
	cases := []struct {
		name     string
		method   string
		url      string
		isJSON   bool
		body     io.Reader
		params   map[string]string
		testfunc func(t *testing.T, v *teststruct)
	}{
		{
			name:   "Complete POST request with a JSON body, no overwriting",
			method: "POST",
			url:    "/?urlparam=url-param",
			isJSON: true,
			body:   strings.NewReader(`{"jsonparam":"json-param"}`),
			params: map[string]string{"pathparam": "path-param"},
			testfunc: func(t *testing.T, v *teststruct) {
				if v.PathParam != "path-param" {
					t.Fatalf("expected path-param but got %s", v.PathParam)
				}
				if v.URLParam != "url-param" {
					t.Fatalf("expected url-param but got %s", v.URLParam)
				}
				if v.JSONParam != "json-param" {
					t.Fatalf("expected json-param but got %s", v.JSONParam)
				}
			},
		},
		{
			name:   "Complete GET request, no overwriting",
			method: "GET",
			url:    "/?urlparam=url-param",
			body:   nil,
			params: map[string]string{"pathparam": "path-param"},
			testfunc: func(t *testing.T, v *teststruct) {
				if v.PathParam != "path-param" {
					t.Fatalf("expected path-param but got %s", v.PathParam)
				}
				if v.URLParam != "url-param" {
					t.Fatalf("expected url-param but got %s", v.URLParam)
				}
				if v.JSONParam != "" {
					t.Fatalf("expected empty string but got %s", v.JSONParam)
				}
			},
		},
		{
			name:   "Complete GET request, overwriting by path param",
			method: "GET",
			url:    "/?pathparam=not-me",
			body:   nil,
			params: map[string]string{"pathparam": "yes-me"},
			testfunc: func(t *testing.T, v *teststruct) {
				if v.PathParam != "yes-me" {
					t.Fatalf("expected yes-me but got %s", v.PathParam)
				}
				if v.URLParam != "" {
					t.Fatalf("expected empty string but got %s", v.URLParam)
				}
				if v.JSONParam != "" {
					t.Fatalf("expected empty string but got %s", v.JSONParam)
				}
			},
		},
	}

	for _, tc := range cases {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(tc.method, tc.url, tc.body)

		if tc.isJSON {
			r.Header.Set("Content-Type", "application/json")
		}

		c := seatbelt.NewTestContext(w, r, tc.params)

		v := &teststruct{}
		fn := func(c seatbelt.Context) error {
			if err := c.Params(v); err != nil {
				return err
			}
			return c.NoContent()
		}

		// Execute the handler.
		if err := fn(c); err != nil {
			t.Fatalf("%+v executing handler", err)
		}

		// The current test case's validations.
		tc.testfunc(t, v)
	}
}
