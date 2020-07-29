package main

import (
	"log"

	"github.com/bentranter/go-seatbelt"
)

func handle(c *seatbelt.Context) error {
	return c.Render(200, "index", nil)
}

func main() {
	app := seatbelt.New(seatbelt.Config{Dir: "testdata"})
	app.Get("/", handle)
	log.Fatalln(app.Start(":3000"))
}
