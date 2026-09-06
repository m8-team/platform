package memory

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/m8-team/platform/internal/resourcemanager/app/ports"
	"github.com/m8-team/platform/internal/resourcemanager/domain/organization"
)

func BenchmarkOrganizationListPage(b *testing.B) {
	repository := NewOrganizationRepository()
	for i := range 1000 {
		value, err := organization.New(organization.CreateParams{ID: organization.NewID(), Name: fmt.Sprintf("org-%04d", i), Now: time.Now(), Labels: map[string]string{"env": "prod"}})
		if err != nil {
			b.Fatal(err)
		}
		if err := repository.Create(context.Background(), value); err != nil {
			b.Fatal(err)
		}
	}
	b.ReportAllocs()
	for b.Loop() {
		if _, err := repository.List(context.Background(), ports.ListOrganizationsOptions{PageSize: 10}); err != nil {
			b.Fatal(err)
		}
	}
}
