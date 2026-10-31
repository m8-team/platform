package grpcadapter

import (
	"testing"
	"time"

	pb "github.com/m8-team/go-genproto/m8/platform/resourcemanager/v1"
	"github.com/m8-team/platform/internal/resourcemanager/v1engine"
)

func TestOrganizationMetadataRoundTrip(t *testing.T) {
	input := &pb.ResourceRecord{Resource: &pb.ResourceRecord_Organization{Organization: &pb.Organization{
		Metadata: &pb.ResourceMetadata{
			Name:        "Example",
			Description: "Example organization",
			Labels:      map[string]string{"environment": "production"},
		},
		TenantId: "7f56f4f7-e92d-4b12-9343-b93de0389af8",
	}}}

	record, err := fromRecord(input)
	if err != nil {
		t.Fatalf("fromRecord() error = %v", err)
	}
	if record.Labels["environment"] != "production" {
		t.Fatalf("fromRecord() labels = %v", record.Labels)
	}

	deletedAt := time.Date(2026, time.October, 31, 12, 0, 0, 0, time.UTC)
	record.ID = "5ee1748d-6a24-43f2-bf74-80b5bcf5d3d9"
	record.State = "DELETED"
	record.CreatedAt = deletedAt.Add(-time.Hour)
	record.UpdatedAt = deletedAt
	record.DeletedAt = &deletedAt
	record.Kind = v1engine.Organization

	output := toRecord(record).GetOrganization().GetMetadata()
	if output.GetLabels()["environment"] != "production" {
		t.Fatalf("toRecord() labels = %v", output.GetLabels())
	}
	if output.GetDeleteTime() == nil || !output.GetDeleteTime().AsTime().Equal(deletedAt) {
		t.Fatalf("toRecord() delete_time = %v", output.GetDeleteTime())
	}
}
