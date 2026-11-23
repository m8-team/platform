package shared

import (
	"errors"
	"strings"
)

var (
	ErrResourceTypeRequired = errors.New("resource type is required")
	ErrResourceIDRequired   = errors.New("resource id is required")
)

// Resource is a reference only; tenancy, owner and lifecycle come from Resource Manager.
type Resource struct{ resourceType, id string }

func NewResource(resourceType, id string) (Resource, error) {
	r := Resource{resourceType: strings.TrimSpace(resourceType), id: strings.TrimSpace(id)}
	return r, r.Validate()
}
func (r Resource) Validate() error {
	if r.resourceType == "" {
		return ErrResourceTypeRequired
	}
	if r.id == "" {
		return ErrResourceIDRequired
	}
	return nil
}
func (r Resource) Type() string { return r.resourceType }
func (r Resource) ID() string   { return r.id }
