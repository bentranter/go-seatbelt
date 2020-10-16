package main

import (
	"log"

	"github.com/bentranter/go-seatbelt"
)

func handle(c seatbelt.Context) error {
	return c.Render(200, "index", nil)
}

func products(c seatbelt.Context) error {
	return c.Render(200, "products/show", nil)
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

	return c.Render(201, "products/new", p)
}

func redirector(c seatbelt.Context) error {
	return c.Redirect("/", "message", "You've been redirected")
}

func main() {
	app := seatbelt.New(&seatbelt.Config{
		Dir: "testdata",
	})

	app.Get("/", handle)
	app.Get("/products", products)
	app.Post("/products", newProduct)
	app.Get("/redirect", redirector)

	app.ErrorHandler(func(err error, c seatbelt.Context) {
		if err := c.String(err.Error()); err != nil {
			log.Println("failed to write response", err)
		}
	})

	log.Fatalln(app.Start(":3000"))
}
