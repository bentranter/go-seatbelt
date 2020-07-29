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
		// TODO: funcs go here and only here
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
func (c *Context) Render(status int, template string, data interface{}) error {
	tmpl, ok := c.templates[template]
	if !ok {
		return errors.New(`template "` + template + `" is not defined`)
	}

	c.Resp.WriteHeader(status)
	return tmpl.ExecuteTemplate(c.Resp, "layout", data)
}
