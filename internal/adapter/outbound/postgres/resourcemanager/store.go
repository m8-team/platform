package postgres

import (
	"strconv"
	"strings"
	"sync"
	"time"

	organizationentity "github.com/m8platform/platform/internal/entity/resourcemanager/organization"
	"github.com/m8platform/platform/internal/entity/resourcemanager/shared"
)

type Store struct {
	mu            sync.RWMutex
	organizations map[string]organizationentity.Entity
}

func NewStore() *Store {
	return &Store{
		organizations: make(map[string]organizationentity.Entity),
	}
}

func cloneOrganization(entity organizationentity.Entity) organizationentity.Entity {
	entity.Annotations = shared.CloneMetadata(entity.Annotations)
	return entity
}

func pageWindow(pageSize int32, pageToken string, total int) (start int, end int, next string) {
	size := int(pageSize)
	if size <= 0 {
		size = 50
	}
	start = 0
	if pageToken != "" {
		if value, err := strconv.Atoi(pageToken); err == nil && value >= 0 {
			start = value
		}
	}
	if start > total {
		start = total
	}
	end = start + size
	if end > total {
		end = total
	}
	if end < total {
		next = strconv.Itoa(end)
	}
	return start, end, next
}

func parseOrderBy(raw string) (field string, desc bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", false
	}
	parts := strings.Fields(raw)
	field = parts[0]
	if len(parts) > 1 && strings.EqualFold(parts[1], "desc") {
		desc = true
	}
	return field, desc
}

func compareTime(left time.Time, right time.Time) int {
	return left.Compare(right)
}
