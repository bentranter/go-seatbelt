package seatbelt

import (
	"testing"
)

func Test_parseRoute(t *testing.T) {
	cases := []struct {
		name     string
		verb     string
		path     string
		expected string
	}{
		{
			name:     "root path",
			verb:     "GET",
			path:     "/",
			expected: "root",
		},
		{
			name:     "get single path element",
			verb:     "GET",
			path:     "/orders",
			expected: "orders",
		},
		{
			name:     "get multiple path elemets",
			verb:     "GET",
			path:     "/products/images",
			expected: "product_images",
		},
		{
			name:     "magic new path suffix",
			verb:     "GET",
			path:     "/products/images/new",
			expected: "new_product_image",
		},
		{
			name:     "magic edit path suffix with wildcard",
			verb:     "GET",
			path:     "/notifications/:id/edit",
			expected: "edit_notification",
		},
		{
			name:     "non-magic path suffix",
			verb:     "GET",
			path:     "/notifications/:id/edit",
			expected: "edit_notification",
		},
		{
			name:     "non-magic path suffix",
			verb:     "GET",
			path:     "/profiles/payment",
			expected: "payment_profile",
		},
		{
			name:     "non-magic path suffix, plural path element",
			verb:     "GET",
			path:     "/profiles/payment",
			expected: "payment_profile",
		},
	}

	for _, c := range cases {
		r := parseRoute(c.verb, c.path)
		if r.prefix != c.expected {
			t.Fatalf("expected %s but got %s", c.expected, r.prefix)
		}
	}
}
