package main

import (
	"html/template"
	"log"
	"strings"

	"github.com/bentranter/go-seatbelt"
)

func handle(c seatbelt.Context) error {
	c.Session().Put("key", "value")
	return c.Render("home/index", nil)
}

func products(c seatbelt.Context) error {
	return c.Render("products/show", nil)
}

type product struct {
	Name  string
	Price int
}

func newProduct(c seatbelt.Context) error {
	p := &product{}

	if err := c.Params(p); err != nil {
		return err
	}

	return c.Render("products/new", p)
}

func redirector(c seatbelt.Context) error {
	return c.Redirect("/")
}

func main() {
	app := seatbelt.New(seatbelt.Option{
		TemplateDir: "testdata",
		Funcs: template.FuncMap{
			"lower": strings.ToLower,
		},
	})

	app.Get("/", handle)
	app.Get("/products", products)
	app.Post("/products", newProduct)
	app.Get("/redirect", redirector)

	log.Fatalln(app.Start(":3000"))
}
