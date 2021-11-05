package main

import (
	"html/template"
	"log"
	"strings"

	"github.com/bentranter/go-seatbelt"
)

func handle(c seatbelt.Context) error {
	return c.Render("home/index", nil)
}

type product struct {
	Name  string
	Price int
}

func newProduct(c seatbelt.Context) error {
	return c.Render("products/new", nil)
}

func createProduct(c seatbelt.Context) error {
	p := &product{}

	if err := c.Params(p); err != nil {
		return err
	}

	c.Session().Flash("notice", "Successfully added product "+p.Name)
	return c.Redirect("/")
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
	app.Get("/products/new", newProduct)
	app.Post("/products", createProduct)
	app.Get("/redirect", redirector)

	log.Fatalln(app.Start(":3000"))
}
