package seatbelt

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gorilla/csrf"
)

// Renderer is an instance of a template renderer.
type Renderer struct {
	// templates are our HTML templates.
	templates map[string]*template.Template

	// textTemplates are our plaintextL templates.
	textTemplates map[string]*template.Template

	// dir is the root of the templates directory.
	dir string

	// reload, if true, will reload the templates from the filesystem on each
	// request.
	reload bool

	// funcs are the HTML template functions passed to the Renderer instance.
	funcs template.FuncMap
}

// NewRenderer returns a new instance of a renderer.
func NewRenderer(dir string, reload bool, funcs ...template.FuncMap) *Renderer {
	dirPath, err := filepath.Abs(dir)
	if err != nil {
		log.Fatalf("%v failed to determine absolute filepath", err)
	}

	htmlfn := make(template.FuncMap)
	for _, fn := range funcs {
		// If you call NewRenderer as pass `nil` into the optional `funcs`
		// arg, then `fn` will be nil and will overwrite the initialized
		// `htmlfn` map, which causes later calls to set the default template
		// funcs to panic.
		//
		// To get around this, we just need the nil check below.
		if fn != nil {
			htmlfn = fn
		}
	}

	re := &Renderer{dir: dirPath, reload: reload, funcs: htmlfn}

	if err := re.parseTemplates(); err != nil {
		panic(err)
	}

	return re
}

// A templateFile is used during the parsing of our templates to save each
// regular, non-layout template until we can parse them within a layout
// context.
type templateFile struct {
	name    string
	content string
}

// parseTemplates reads and parses the templates from the filesystem.
func (r *Renderer) parseTemplates() error {
	htmlLayouts := template.New("html_layouts")
	textLayouts := template.New("text_layouts")

	htmlTemplateFiles := make([]templateFile, 0)
	textTemplateFiles := make([]templateFile, 0)

	if err := filepath.Walk(r.dir, func(path string, info os.FileInfo, err error) error {
		// Fix same-extension-dirs bug: some dir might be named to:
		// "users.tmpl", "local.html". These dirs should be excluded as they
		// are not valid golang templates, but files under them should be
		// treat as normal. If is a dir, return immediately (dir is not a
		// valid golang template).
		if info == nil || info.IsDir() {
			return nil
		}

		rel, err := filepath.Rel(r.dir, path)
		if err != nil {
			return err
		}

		ext := ""
		if strings.Index(rel, ".") != -1 {
			ext = filepath.Ext(rel)
		}
		if ext != ".txt" && ext != ".html" {
			return fmt.Errorf("templates must end in .html or .txt, got %s", ext)
		}

		buf, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}

		name := (rel[0 : len(rel)-len(ext)])

		// If we're not in the layouts directory, we don't want to parse these
		// regular templates until we have parsed all layout templates. To
		// support this, we'll save these templates unparsed.
		if folder := filepath.Dir(rel); folder != "layouts" {
			if ext == ".html" {
				htmlTemplateFiles = append(htmlTemplateFiles, templateFile{
					name:    name,
					content: string(buf),
				})
			}
			if ext == ".txt" {
				textTemplateFiles = append(textTemplateFiles, templateFile{
					name:    name,
					content: string(buf),
				})
			}
		}

		// Add used supplied HTML functions, and add our defaults.
		funcs := r.funcs
		funcs["csrf"] = func() template.HTML {
			return ""
		}

		if ext == ".html" {
			if _, err := htmlLayouts.New(name).Funcs(funcs).Parse(string(buf)); err != nil {
				return err
			}
		}
		if ext == ".txt" {
			if _, err := textLayouts.New(name).Funcs(funcs).Parse(string(buf)); err != nil {
				return err
			}
		}

		return nil
	}); err != nil {
		return err
	}

	htmlTemplates := make(map[string]*template.Template)
	textTemplates := make(map[string]*template.Template)

	for _, tf := range htmlTemplateFiles {
		t, err := htmlLayouts.Clone()
		if err != nil {
			return err
		}

		if _, err := t.Parse(tf.content); err != nil {
			return err
		}

		htmlTemplates[tf.name] = t
	}

	for _, tf := range textTemplateFiles {
		t, err := textLayouts.Clone()
		if err != nil {
			return err
		}

		if _, err := t.Parse(tf.content); err != nil {
			return err
		}

		textTemplates[tf.name] = t
	}

	r.templates = htmlTemplates
	r.textTemplates = textTemplates

	return nil
}

// RenderOption contains the optional options for rendering templates.
type RenderOption struct {
	// The Layout to use when rendering the template. The default is
	// `application`.
	Layout string

	// Status is the HTTP status code to send when rendering a template. The
	// default is 200.
	Status int
}

// HTML writes an HTML template to a buffer.
//
// The name of the layout does **not** require the "layouts/" prefix, unlike
// other templates.
func (r *Renderer) HTML(w io.Writer, req *http.Request, name string, data interface{}, opts ...RenderOption) error {
	opt := RenderOption{
		Layout: "application",
		Status: 200,
	}
	for _, o := range opts {
		opt = o
	}

	if r.reload {
		if err := r.parseTemplates(); err != nil {
			return err
		}
	}

	buf := &bytes.Buffer{}
	tpl, ok := r.templates[name]
	if !ok {
		return errors.New("the template " + name + " does not exist")
	}

	// Create a new func map that has the CSRF func with the
	// implementation, in addition to the user provided funcs if there are
	// any.
	contextualFuncMap := make(template.FuncMap)
	for fn, impl := range r.funcs {
		contextualFuncMap[fn] = impl
	}
	contextualFuncMap["csrf"] = func() template.HTML {
		if req == nil {
			return ""
		}
		return csrf.TemplateField(req)
	}

	// Provide those funcs.
	tpl.Funcs(contextualFuncMap)

	if err := tpl.ExecuteTemplate(buf, "layouts/"+opt.Layout, data); err != nil {
		return err
	}

	if rw, ok := w.(http.ResponseWriter); ok {
		rw.Header().Set("Content-Type", "text/html")
		rw.WriteHeader(opt.Status)
	}
	_, err := buf.WriteTo(w)
	return err
}

// Text renders the template with the given name to a string. It will render
// templates that end in .txt.
//
// This should be used when rendering a template outside the context of an
// HTTP request, ie, rendering an email template, or a plain text template.
func (r *Renderer) Text(name string, data interface{}, opts ...RenderOption) (string, error) {
	opt := RenderOption{
		Layout: "application",
	}
	for _, o := range opts {
		opt = o
	}

	if r.reload {
		if err := r.parseTemplates(); err != nil {
			return "", err
		}
	}

	buf := &bytes.Buffer{}
	tpl, ok := r.textTemplates[name]
	if !ok {
		return "", errors.New("the template " + name + " does not exist")
	}

	// Like in the HTML method, add user provided funcs.
	contextualFuncMap := make(template.FuncMap)
	for fn, impl := range r.funcs {
		contextualFuncMap[fn] = impl
	}
	tpl.Funcs(contextualFuncMap)

	if err := tpl.ExecuteTemplate(buf, "layouts/"+opt.Layout, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}
