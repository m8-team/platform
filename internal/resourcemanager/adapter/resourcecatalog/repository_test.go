package resourcecatalog

import (
	"context"
	"errors"
	"testing"

	"github.com/m8-team/platform/internal/resourcemanager/domain/resource"
	"github.com/m8-team/platform/internal/resourcemanager/v1engine"
)

type stubCatalog struct {
	record v1engine.Record
	err    error
}

func (s stubCatalog) Get(context.Context, v1engine.Kind, string) (v1engine.Record, error) {
	return s.record, s.err
}

func TestRepositoryMapsAuthoritativeRecord(t *testing.T) {
	_, err := New(nil)
	if !errors.Is(err, ErrCatalogRequired) {
		t.Fatalf("nil catalog accepted: %v", err)
	}
	record := v1engine.Record{
		Kind:           v1engine.Workspace,
		ID:             "w1",
		TenantID:       "t1",
		Owner:          v1engine.Owner{Kind: v1engine.Organization, ID: "o1"},
		OrganizationID: "o1",
		State:          "ACTIVE",
		Revision:       3,
	}
	repo, _ := New(stubCatalog{record: record})
	snapshot, err := repo.Load(context.Background(), resource.Workspace, "w1")
	if err != nil || snapshot.OwnerID != "o1" || snapshot.Revision != 3 {
		t.Fatalf("unexpected result %+v %v", snapshot, err)
	}
}
