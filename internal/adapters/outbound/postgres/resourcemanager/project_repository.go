package postgres

import (
	"context"
	"sort"
	"strings"

	projectentity "github.com/m8platform/platform/internal/entities/resourcemanager/project"
	"github.com/m8platform/platform/internal/usecase/resourcemanager/ports"
)

type ProjectRepository struct {
	Store *Store
}

func (r ProjectRepository) GetByID(_ context.Context, id string, includeDeleted bool) (projectentity.Entity, error) {
	r.Store.mu.RLock()
	defer r.Store.mu.RUnlock()

	entity, ok := r.Store.projects[id]
	if !ok {
		return projectentity.Entity{}, projectentity.ErrNotFound
	}
	if entity.IsDeleted() && !includeDeleted {
		return projectentity.Entity{}, projectentity.ErrNotFound
	}
	return cloneProject(entity), nil
}

func (r ProjectRepository) Create(_ context.Context, entity projectentity.Entity) error {
	r.Store.mu.Lock()
	defer r.Store.mu.Unlock()

	r.Store.projects[entity.ID] = cloneProject(entity)
	return nil
}

func (r ProjectRepository) Update(_ context.Context, entity projectentity.Entity) error {
	r.Store.mu.Lock()
	defer r.Store.mu.Unlock()

	if _, ok := r.Store.projects[entity.ID]; !ok {
		return projectentity.ErrNotFound
	}
	r.Store.projects[entity.ID] = cloneProject(entity)
	return nil
}

func (r ProjectRepository) List(_ context.Context, params ports.ProjectListParams) (ports.ProjectPage, error) {
	r.Store.mu.RLock()
	defer r.Store.mu.RUnlock()

	items := make([]projectentity.Entity, 0, len(r.Store.projects))
	for _, entity := range r.Store.projects {
		if entity.WorkspaceID != params.WorkspaceID {
			continue
		}
		if entity.IsDeleted() && !params.ShowDeleted {
			continue
		}
		items = append(items, cloneProject(entity))
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
	return ports.ProjectPage{
		Items:         items[start:end],
		NextPageToken: next,
		TotalSize:     int32(len(items)),
	}, nil
}
