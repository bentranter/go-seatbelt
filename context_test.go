package seatbelt

import (
	"testing"
)

func TestContext(t *testing.T) {
	fn := func(c *Context) error {
		return c.String("hello, world!")
	}

	c := NewTestContext("GET", "/", nil, nil, nil)

	if err := fn(c); err != nil {
		t.Fatalf("%+v sending string response", err)
	}
}
