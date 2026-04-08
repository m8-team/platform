package postgres

import (
	"context"
	"sort"
	"strings"

	organizationentity "github.com/m8platform/platform/internal/entity/resourcemanager/organization"
	"github.com/m8platform/platform/internal/usecase/resourcemanager/port"
)

type OrganizationRepository struct {
	Store *Store
}

func (r OrganizationRepository) GetByID(_ context.Context, id string, includeDeleted bool) (organizationentity.Entity, error) {
	r.Store.mu.RLock()
	defer r.Store.mu.RUnlock()

	entity, ok := r.Store.organizations[id]
	if !ok {
		return organizationentity.Entity{}, organizationentity.ErrNotFound
	}
	if entity.IsDeleted() && !includeDeleted {
		return organizationentity.Entity{}, organizationentity.ErrNotFound
	}
	return cloneOrganization(entity), nil
}

func (r OrganizationRepository) Create(_ context.Context, entity organizationentity.Entity) error {
	r.Store.mu.Lock()
	defer r.Store.mu.Unlock()

	r.Store.organizations[entity.ID] = cloneOrganization(entity)
	return nil
}

func (r OrganizationRepository) Update(_ context.Context, entity organizationentity.Entity) error {
	r.Store.mu.Lock()
	defer r.Store.mu.Unlock()

	if _, ok := r.Store.organizations[entity.ID]; !ok {
		return organizationentity.ErrNotFound
	}
	r.Store.organizations[entity.ID] = cloneOrganization(entity)
	return nil
}

func (r OrganizationRepository) SoftDelete(ctx context.Context, entity organizationentity.Entity) error {
	return r.Update(ctx, entity)
}

func (r OrganizationRepository) Undelete(ctx context.Context, entity organizationentity.Entity) error {
	return r.Update(ctx, entity)
}

func (r OrganizationRepository) List(_ context.Context, params port.OrganizationListParams) (port.OrganizationPage, error) {
	r.Store.mu.RLock()
	defer r.Store.mu.RUnlock()

	items := make([]organizationentity.Entity, 0, len(r.Store.organizations))
	for _, entity := range r.Store.organizations {
		if entity.IsDeleted() && !params.ShowDeleted {
			continue
		}
		items = append(items, cloneOrganization(entity))
	}
	field, desc := parseOrderBy(params.OrderBy)
	sort.SliceStable(items, func(i int, j int) bool {
		compare := 0
		switch field {
		case "name":
			compare = strings.Compare(items[i].Name, items[j].Name)
		case "update_time":
			compare = compareTime(items[i].UpdateTime, items[j].UpdateTime)
		case "id":
			compare = strings.Compare(items[i].ID, items[j].ID)
		default:
			compare = compareTime(items[i].CreateTime, items[j].CreateTime)
		}
		if compare == 0 {
			compare = strings.Compare(items[i].ID, items[j].ID)
		}
		if desc {
			return compare > 0
		}
		return compare < 0
	})

	start, end, next := pageWindow(params.PageSize, params.PageToken, len(items))
	return port.OrganizationPage{
		Items:         items[start:end],
		NextPageToken: next,
		TotalSize:     int32(len(items)),
	}, nil
}
