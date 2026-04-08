package organizationquery

import (
	organizationboundary "github.com/m8platform/platform/internal/usecase/resourcemanager/organization/boundary"
	"github.com/m8platform/platform/internal/usecase/resourcemanager/port"
)

type ListInputValidator interface {
	Validate(organizationboundary.ListOrganizationsInput) error
}

type QueryValidator struct {
	FilterValidator port.FilterValidator
	OrderValidator  port.OrderValidator
}

func (v QueryValidator) Validate(input organizationboundary.ListOrganizationsInput) error {
	if v.FilterValidator != nil {
		if err := v.FilterValidator.Validate(input.Filter); err != nil {
			return err
		}
	}
	if v.OrderValidator != nil {
		if err := v.OrderValidator.Validate(input.OrderBy); err != nil {
			return err
		}
	}
	return nil
}
