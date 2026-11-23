package policyjournal

import (
	"testing"

	"github.com/m8-team/platform/internal/access/app/projection"
	"github.com/m8-team/platform/internal/access/domain"
)

func TestManifestRequiresCompleteIndependentTargets(t *testing.T) {
	tenant, _ := domain.NewTenantID("aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa")
	first := projection.Key{TenantID: tenant, Kind: projection.KindRoleBinding, AggregateID: "11111111-1111-4111-8111-111111111111"}
	second := projection.Key{TenantID: tenant, Kind: projection.KindRoleBinding, AggregateID: "22222222-2222-4222-8222-222222222222"}
	manifest := Mutation{TenantID: tenant, ExpectedHead: 7, ID: "role-update-42", Jobs: []Job{{first, 5}, {second, 9}}}
	if err := manifest.Validate(); err != nil {
		t.Fatal(err)
	}
	a, err := manifest.Digest()
	if err != nil {
		t.Fatal(err)
	}
	manifest.Jobs[0], manifest.Jobs[1] = manifest.Jobs[1], manifest.Jobs[0]
	b, err := manifest.Digest()
	if err != nil || a != b {
		t.Fatalf("nondeterministic digest: %v", err)
	}
	manifest.Jobs = append(manifest.Jobs, manifest.Jobs[0])
	if manifest.Validate() == nil {
		t.Fatal("duplicate binding accepted")
	}
	manifest.Jobs = manifest.Jobs[:2]
	manifest.Jobs[0].RequiredRevision = 0
	if manifest.Validate() == nil {
		t.Fatal("unspecified binding revision accepted")
	}
}
func TestManifestRejectsCrossTenantAndRMKeys(t *testing.T) {
	tenant, _ := domain.NewTenantID("aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa")
	foreign, _ := domain.NewTenantID("bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb")
	key := projection.Key{TenantID: foreign, Kind: projection.KindGrant, AggregateID: "11111111-1111-4111-8111-111111111111"}
	m := Mutation{TenantID: tenant, ID: "grant-revoke-99", Jobs: []Job{{key, 3}}}
	if m.Validate() == nil {
		t.Fatal("cross-tenant mutation accepted")
	}
	key.TenantID = tenant
	key.Kind = projection.KindRMWorkspace
	m.Jobs = []Job{{key, 3}}
	if m.Validate() == nil {
		t.Fatal("RM resource change published as Access policy")
	}
}
