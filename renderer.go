package seatbelt

import (
	"errors"
	"fmt"
	"html/template"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

func parseTemplates(dir string) map[string]*template.Template {
	b, err := ioutil.ReadFile(filepath.Join(dir, "layout.html"))
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	templates := make(map[string]*template.Template)

	layout, err := template.New("layout").
		Funcs(template.FuncMap{
			"flashes": func() map[string]string {
				return make(map[string]string)
			},
		}).
		Parse(string(b))
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		// Fix same-extension-dirs bug: some dir might be named to:
		// "users.tmpl", "local.html". These dirs should be excluded as they
		// are not valid golang templates, but files under them should be
		// treat as normal. If is a dir, return immediately (dir is not a
		// valid golang template).
		if info == nil || info.IsDir() {
			return nil
		}

		rel, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}

		ext := ""
		if strings.Index(rel, ".") != -1 {
			ext = filepath.Ext(rel)
		}

		buf, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}

		name := (rel[0 : len(rel)-len(ext)])

		tmpl, err := template.Must(layout.Clone()).Parse(string(buf))
		if err != nil {
			return err
		}

		templates[filepath.ToSlash(name)] = tmpl
		return nil
	}); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	return templates
}

// Render renders the HTML template with the given data, if any.
func (c *Context) Render(status int, name string, data interface{}) error {
	tmpl, ok := c.templates[name]
	if !ok {
		return errors.New(`template "` + name + `" is not defined`)
	}

	flashes := c.Flash.All()

	// Override previous func map because Go's templates are weird.
	tmpl.Funcs(template.FuncMap{
		"flashes": func() map[string]string {
			return flashes
		},
	})

	c.Resp.WriteHeader(status)
	return tmpl.ExecuteTemplate(c.Resp, "layout", data)
}
