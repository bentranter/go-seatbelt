package main

import (
	"log"

	"github.com/bentranter/go-seatbelt"
)

func handle(c *seatbelt.Context) error {
	return c.Render(200, "index", nil)
}

func products(c *seatbelt.Context) error {
	return c.Render(200, "products/show", nil)
}

func redirector(c *seatbelt.Context) error {
	return c.Redirect("/", "message", "You've been redirected")
}

func main() {
	app := seatbelt.New(seatbelt.Config{
		Dir: "testdata",
	})

	app.Get("/", handle)
	app.Get("/products", products)
	app.Get("/redirect", redirector)

	log.Fatalln(app.Start(":3000"))
}
