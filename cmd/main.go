package main

import (
	"log"

	"github.com/bentranter/go-seatbelt"
)

func handle(c *seatbelt.Context) error {
	c.Flash("message", "hello, world!")
	return c.Render(200, "index", nil)
}

func products(c *seatbelt.Context) error {
	return c.Render(200, "products/show", c.Flashes())
}

func main() {
	app := seatbelt.New(seatbelt.Config{
		Dir: "testdata",
	})

	app.Get("/", handle)
	app.Get("/products", products)

	log.Fatalln(app.Start(":3000"))
}
