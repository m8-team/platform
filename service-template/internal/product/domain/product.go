package domain

import (
	"errors"
	"strings"
	"unicode/utf8"
)

type ID string
type TenantID string
type Product struct {
	ID     ID       `json:"id"`
	Tenant TenantID `json:"tenant"`
	Name   string   `json:"name"`
}

var ErrInvalidName = errors.New("product name must contain 1 to 200 characters without surrounding whitespace")

func New(id ID, tenant TenantID, name string) (Product, error) {
	if id == "" || tenant == "" {
		return Product{}, errors.New("product and tenant identifiers required")
	}
	if !utf8.ValidString(name) || strings.TrimSpace(name) != name || name == "" || utf8.RuneCountInString(name) > 200 {
		return Product{}, ErrInvalidName
	}
	return Product{id, tenant, name}, nil
}
