package postgres

import (
	"context"
	"database/sql"

	"github.com/m8platform/platform/internal/domain/organization"
	"github.com/m8platform/platform/internal/ports"
)

type OrganizationRepository struct {
	DB *sql.DB
}

func NewOrganizationRepository(db *sql.DB) *OrganizationRepository {
	return &OrganizationRepository{DB: db}
}

func (r *OrganizationRepository) GetByID(context.Context, string, bool) (organization.Organization, error) {
	return organization.Organization{}, ports.ErrNotImplemented
}

func (r *OrganizationRepository) Create(context.Context, organization.Organization) error {
	return ports.ErrNotImplemented
}

func (r *OrganizationRepository) Update(context.Context, organization.Organization) error {
	return ports.ErrNotImplemented
}

func (r *OrganizationRepository) List(context.Context, organization.ListParams) (organization.Page, error) {
	return organization.Page{}, ports.ErrNotImplemented
}
