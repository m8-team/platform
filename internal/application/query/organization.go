package query

import (
	"context"
	"fmt"

	"github.com/m8platform/platform/internal/domain/organization"
	"github.com/m8platform/platform/internal/ports"
)

type GetOrganization struct {
	ID string
}

type GetOrganizationHandler struct {
	Repository ports.OrganizationRepository
}

func (h GetOrganizationHandler) Handle(ctx context.Context, q GetOrganization) (organization.Organization, error) {
	aggregate, err := h.Repository.GetByID(ctx, q.ID, true)
	if err != nil {
		return organization.Organization{}, fmt.Errorf("get organization: %w", err)
	}
	return aggregate, nil
}

type ListOrganizations struct {
	PageSize    int32
	PageToken   string
	Filter      string
	OrderBy     string
	ShowDeleted bool
}

type ListOrganizationsHandler struct {
	Repository   ports.OrganizationRepository
	FilterParser ports.FilterParser
	OrderParser  ports.OrderParser
}

func (h ListOrganizationsHandler) Handle(ctx context.Context, q ListOrganizations) (organization.Page, error) {
	if h.FilterParser != nil {
		if err := h.FilterParser.Validate(q.Filter); err != nil {
			return organization.Page{}, err
		}
	}
	if h.OrderParser != nil {
		if err := h.OrderParser.Validate(q.OrderBy); err != nil {
			return organization.Page{}, err
		}
	}
	page, err := h.Repository.List(ctx, organization.ListParams{
		PageSize:    q.PageSize,
		PageToken:   q.PageToken,
		Filter:      q.Filter,
		OrderBy:     q.OrderBy,
		ShowDeleted: q.ShowDeleted,
	})
	if err != nil {
		return organization.Page{}, fmt.Errorf("list organizations: %w", err)
	}
	return page, nil
}
