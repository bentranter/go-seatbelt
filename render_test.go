package seatbelt_test

import (
	"bytes"
	"html/template"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/bentranter/go-seatbelt"
	"github.com/gorilla/csrf"
)

func TestContextRender(t *testing.T) {
	cases := []struct {
		template        string
		expectedName    string
		expectedPartial string
	}{
		{
			template:        "home/index",
			expectedName:    "Home",
			expectedPartial: "HomePartial",
		},
		{
			template:        "account/index",
			expectedName:    "Accounts",
			expectedPartial: "AccountsPartial",
		},
	}

	for _, c := range cases {
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		w := httptest.NewRecorder()

		ctx := seatbelt.NewTestContext(w, r)
		ctx.AddRenderer("testdata", template.FuncMap{
			"lower": strings.ToLower,
		})

		if err := ctx.Render(c.template, nil); err != nil {
			t.Errorf("%+v rendering template", err)
		}

		rendered := w.Body.String()

		if !strings.Contains(rendered, c.expectedName) {
			t.Fatalf("expected %s in %s", c.expectedName, rendered)
		}

		if !strings.Contains(rendered, c.expectedPartial) {
			t.Fatalf("expected partial %s in %s", c.expectedPartial, rendered)
		}
	}
}

func TestRenderCSRF(t *testing.T) {
	t.Parallel()

	app := seatbelt.New(seatbelt.Option{
		TemplateDir: "testdata",
		Funcs: template.FuncMap{
			"lower": strings.ToLower,
		},
	})

	app.Get("/", func(c seatbelt.Context) error {
		return c.Render("home/index", nil)
	})

	srv := httptest.NewServer(csrf.Protect([]byte("sss"))(app))
	defer srv.Close()

	req, err := http.NewRequest(http.MethodGet, srv.URL+"/", nil)
	if err != nil {
		t.Fatalf("%+v creating new http request", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("%+v executing http request", err)
	}

	rendered, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("%+v reading body", err)
	}

	if !strings.Contains(string(rendered), "csrf") {
		t.Fatalf("expected:\n%s\nto contain csrf", rendered)
	}
}

func TestStandaloneRender(t *testing.T) {
	t.Parallel()

	r := seatbelt.NewRenderer("testdata", false, template.FuncMap{
		"lower": strings.ToLower,
	})

	t.Run("html template", func(t *testing.T) {
		const expectedOutput = "<!DOCTYPE html>\n<html lang=\"en\">\n<head>\n  <meta charset=\"UTF-8\">\n  <meta name=\"viewport\" content=\"width=device-width, initial-scale=1.0\">\n  <title>ok</title>\n</head>\n<body>\n  ok\n</body>\n</html>\n"

		buf := &bytes.Buffer{}
		if err := r.HTML(buf, nil, "plaintext/plain", nil); err != nil {
			t.Fatalf("%+v rendering html template", err)
		}
		if output := buf.String(); output != expectedOutput {
			t.Fatalf("got %#v but expected %s", output, expectedOutput)
		}
	})

	t.Run("plaintext template", func(t *testing.T) {
		output, err := r.Text("plaintext/plain", nil)
		if err != nil {
			t.Fatalf("%+v rendering plaintext template", err)
		}
		if output != "ok\n" {
			t.Fatalf("expected ok but got %#v", output)
		}
	})
}

func TestRenderTemplateFuncs(t *testing.T) {
	t.Parallel()

	app := seatbelt.New(seatbelt.Option{
		TemplateDir: "testdata",
		Funcs: template.FuncMap{
			"lower": strings.ToLower,
		},
	})

	app.Get("/", func(c seatbelt.Context) error {
		return c.Render("home/func", nil)
	})

	srv := httptest.NewServer(csrf.Protect([]byte("sss"))(app))
	defer srv.Close()

	req, err := http.NewRequest(http.MethodGet, srv.URL+"/", nil)
	if err != nil {
		t.Fatalf("%+v creating new http request", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("%+v executing http request", err)
	}

	rendered, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("%+v reading body", err)
	}

	if !strings.Contains(string(rendered), "hi") {
		t.Fatalf("expected:\n%s\nto contain hi", rendered)
	}
	if !strings.Contains(string(rendered), "hey") {
		t.Fatalf("expected:\n%s\nto contain hey", rendered)
	}
}
