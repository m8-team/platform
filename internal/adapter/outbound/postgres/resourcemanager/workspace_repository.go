package postgres

import (
	"context"
	"sort"
	"strings"

	workspaceentity "github.com/m8platform/platform/internal/entity/resourcemanager/workspace"
	"github.com/m8platform/platform/internal/usecase/resourcemanager/port"
)

type WorkspaceRepository struct {
	Store *Store
}

func (r WorkspaceRepository) GetByID(_ context.Context, id string, includeDeleted bool) (workspaceentity.Entity, error) {
	r.Store.mu.RLock()
	defer r.Store.mu.RUnlock()

	entity, ok := r.Store.workspaces[id]
	if !ok {
		return workspaceentity.Entity{}, workspaceentity.ErrNotFound
	}
	if entity.IsDeleted() && !includeDeleted {
		return workspaceentity.Entity{}, workspaceentity.ErrNotFound
	}
	return cloneWorkspace(entity), nil
}

func (r WorkspaceRepository) Create(_ context.Context, entity workspaceentity.Entity) error {
	r.Store.mu.Lock()
	defer r.Store.mu.Unlock()

	r.Store.workspaces[entity.ID] = cloneWorkspace(entity)
	return nil
}

func (r WorkspaceRepository) Update(_ context.Context, entity workspaceentity.Entity) error {
	r.Store.mu.Lock()
	defer r.Store.mu.Unlock()

	if _, ok := r.Store.workspaces[entity.ID]; !ok {
		return workspaceentity.ErrNotFound
	}
	r.Store.workspaces[entity.ID] = cloneWorkspace(entity)
	return nil
}

func (r WorkspaceRepository) List(_ context.Context, params port.WorkspaceListParams) (port.WorkspacePage, error) {
	r.Store.mu.RLock()
	defer r.Store.mu.RUnlock()

	items := make([]workspaceentity.Entity, 0, len(r.Store.workspaces))
	for _, entity := range r.Store.workspaces {
		if entity.OrganizationID != params.OrganizationID {
			continue
		}
		if entity.IsDeleted() && !params.ShowDeleted {
			continue
		}
		items = append(items, cloneWorkspace(entity))
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
	return port.WorkspacePage{
		Items:         items[start:end],
		NextPageToken: next,
		TotalSize:     int32(len(items)),
	}, nil
}
