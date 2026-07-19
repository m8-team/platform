package organization

import (
	"errors"
	"math"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/m8-team/platform/internal/platform/types"
)

var (
	testCreateTime = time.Date(2026, time.July, 19, 10, 0, 0, 123, time.FixedZone("test", 2*60*60))
	testUpdateTime = testCreateTime.Add(time.Hour)
	testPurgeTime  = testUpdateTime.Add(30 * 24 * time.Hour)
)

func TestNewOrganization(t *testing.T) {
	t.Parallel()

	labels := map[string]string{"example.com/team": "platform"}
	aggregate, err := New(CreateParams{
		ID:          MustParseID(testOrganizationID),
		Name:        strings.Repeat("界", MaxNameRunes),
		Description: "control plane",
		Labels:      labels,
		Now:         testCreateTime,
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	labels["example.com/team"] = "mutated"
	if aggregate.ID().String() != testOrganizationID {
		t.Fatalf("ID() = %s, want %s", aggregate.ID(), testOrganizationID)
	}
	if aggregate.State() != StateActive {
		t.Fatalf("State() = %s, want %s", aggregate.State(), StateActive)
	}
	if aggregate.Version() != types.InitialVersion {
		t.Fatalf("Version() = %s, want %s", aggregate.Version(), types.InitialVersion)
	}
	if aggregate.CreateTime().Location() != time.UTC || aggregate.UpdateTime().Location() != time.UTC {
		t.Fatal("New() did not normalize timestamps to UTC")
	}
	if aggregate.CreateTime() != testCreateTime.UTC() || aggregate.UpdateTime() != testCreateTime.UTC() {
		t.Fatal("New() timestamps do not match Now")
	}
	if aggregate.Labels()["example.com/team"] != "platform" {
		t.Fatal("New() retained the caller's labels map")
	}
	if aggregate.DeleteTime() != nil || aggregate.PurgeTime() != nil || aggregate.IsDeleted() {
		t.Fatal("new organization has deletion state")
	}
}

func TestNewOrganizationValidation(t *testing.T) {
	t.Parallel()

	invalidUTF8 := string([]byte{0xff})
	tests := []struct {
		name    string
		mutate  func(*CreateParams)
		wantErr error
	}{
		{name: "empty ID", mutate: func(p *CreateParams) { p.ID = ID{} }, wantErr: ErrEmptyOrganizationID},
		{name: "empty time", mutate: func(p *CreateParams) { p.Now = time.Time{} }, wantErr: ErrEmptyOrganizationTime},
		{name: "invalid name", mutate: func(p *CreateParams) { p.Name = invalidUTF8 }, wantErr: ErrInvalidOrganizationName},
		{name: "long name", mutate: func(p *CreateParams) { p.Name = strings.Repeat("界", MaxNameRunes+1) }, wantErr: ErrOrganizationNameTooLong},
		{name: "invalid description", mutate: func(p *CreateParams) { p.Description = invalidUTF8 }, wantErr: ErrInvalidOrganizationDescription},
		{name: "long description", mutate: func(p *CreateParams) { p.Description = strings.Repeat("界", MaxDescriptionRunes+1) }, wantErr: ErrOrganizationDescriptionTooLong},
		{name: "invalid label key", mutate: func(p *CreateParams) { p.Labels = map[string]string{invalidUTF8: "value"} }, wantErr: ErrInvalidOrganizationLabel},
		{name: "invalid label value", mutate: func(p *CreateParams) { p.Labels = map[string]string{"key": invalidUTF8} }, wantErr: ErrInvalidOrganizationLabel},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			params := validCreateParams()
			tt.mutate(&params)
			if _, err := New(params); !errors.Is(err, tt.wantErr) {
				t.Fatalf("New() error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestRehydrateAndDefensiveCopies(t *testing.T) {
	t.Parallel()

	deleteTime := testUpdateTime
	purgeTime := testPurgeTime
	labels := map[string]string{"team": "platform"}
	snapshot := Snapshot{
		ID:          MustParseID(testOrganizationID),
		State:       StateDeleted,
		Name:        "M8",
		Description: "deleted organization",
		CreateTime:  testCreateTime,
		UpdateTime:  deleteTime,
		DeleteTime:  &deleteTime,
		PurgeTime:   &purgeTime,
		Version:     types.Version(4),
		Labels:      labels,
	}

	aggregate, err := Rehydrate(snapshot)
	if err != nil {
		t.Fatalf("Rehydrate() error = %v", err)
	}
	labels["team"] = "changed"
	deleteTime = deleteTime.Add(time.Hour)
	purgeTime = purgeTime.Add(time.Hour)
	if aggregate.Labels()["team"] != "platform" {
		t.Fatal("Rehydrate() retained labels alias")
	}
	if got := aggregate.DeleteTime(); got == nil || !got.Equal(testUpdateTime.UTC()) {
		t.Fatalf("DeleteTime() = %v, want %v", got, testUpdateTime.UTC())
	}
	if !aggregate.IsDeleted() || aggregate.Version() != types.Version(4) {
		t.Fatal("Rehydrate() did not retain lifecycle state")
	}

	out := aggregate.Snapshot()
	out.Labels["team"] = "snapshot mutation"
	*out.DeleteTime = out.DeleteTime.Add(time.Hour)
	if aggregate.Labels()["team"] != "platform" || !aggregate.DeleteTime().Equal(testUpdateTime.UTC()) {
		t.Fatal("Snapshot() aliases aggregate state")
	}

	clone := aggregate.Clone()
	clone.labels["team"] = "clone mutation"
	*clone.deleteTime = clone.deleteTime.Add(time.Hour)
	if aggregate.Labels()["team"] != "platform" || !aggregate.DeleteTime().Equal(testUpdateTime.UTC()) {
		t.Fatal("Clone() aliases aggregate state")
	}
	if (*Organization)(nil).Clone() != nil {
		t.Fatal("nil Clone() is not nil")
	}
}

func TestRehydrateValidation(t *testing.T) {
	t.Parallel()

	deleteTime := testUpdateTime
	purgeTime := testPurgeTime
	tests := []struct {
		name    string
		mutate  func(*Snapshot)
		wantErr error
	}{
		{name: "empty ID", mutate: func(s *Snapshot) { s.ID = ID{} }, wantErr: ErrEmptyOrganizationID},
		{name: "unspecified state", mutate: func(s *Snapshot) { s.State = StateUnspecified }, wantErr: ErrInvalidOrganizationState},
		{name: "unknown state", mutate: func(s *Snapshot) { s.State = State(99) }, wantErr: ErrInvalidOrganizationState},
		{name: "zero version", mutate: func(s *Snapshot) { s.Version = 0 }, wantErr: ErrInvalidOrganizationVersion},
		{name: "version beyond API range", mutate: func(s *Snapshot) { s.Version = types.Version(math.MaxInt64) + 1 }, wantErr: ErrInvalidOrganizationVersion},
		{name: "zero create time", mutate: func(s *Snapshot) { s.CreateTime = time.Time{} }, wantErr: ErrEmptyOrganizationTime},
		{name: "zero update time", mutate: func(s *Snapshot) { s.UpdateTime = time.Time{} }, wantErr: ErrEmptyOrganizationTime},
		{name: "update before create", mutate: func(s *Snapshot) { s.UpdateTime = s.CreateTime.Add(-time.Second) }, wantErr: ErrInvalidOrganizationTime},
		{name: "unexpected delete time", mutate: func(s *Snapshot) { s.DeleteTime = &deleteTime }, wantErr: ErrUnexpectedDeletionTimes},
		{name: "unexpected purge time", mutate: func(s *Snapshot) { s.PurgeTime = &purgeTime }, wantErr: ErrUnexpectedDeletionTimes},
		{name: "deleted missing delete time", mutate: func(s *Snapshot) { s.State = StateDeleted; s.PurgeTime = &purgeTime }, wantErr: ErrDeleteTimeRequired},
		{name: "deleted missing purge time", mutate: func(s *Snapshot) { s.State = StateDeleted; s.DeleteTime = &deleteTime; s.UpdateTime = deleteTime }, wantErr: ErrPurgeTimeRequired},
		{name: "delete before create", mutate: func(s *Snapshot) {
			before := s.CreateTime.Add(-time.Second)
			s.State = StateDeleted
			s.DeleteTime = &before
			s.PurgeTime = &purgeTime
		}, wantErr: ErrInvalidOrganizationTime},
		{name: "update before delete", mutate: func(s *Snapshot) {
			after := s.UpdateTime.Add(time.Second)
			s.State = StateDeleted
			s.DeleteTime = &after
			s.PurgeTime = &purgeTime
		}, wantErr: ErrInvalidOrganizationTime},
		{name: "purge not after delete", mutate: func(s *Snapshot) {
			s.State = StateDeleted
			s.DeleteTime = &deleteTime
			s.PurgeTime = &deleteTime
			s.UpdateTime = deleteTime
		}, wantErr: ErrInvalidPurgeTime},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			snapshot := validSnapshot()
			tt.mutate(&snapshot)
			if _, err := Rehydrate(snapshot); !errors.Is(err, tt.wantErr) {
				t.Fatalf("Rehydrate() error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestOrganizationUpdate(t *testing.T) {
	t.Parallel()

	name := "renamed"
	description := "updated"
	labels := map[string]string{"team": "runtime"}
	aggregate := mustNewOrganization(t)
	if err := aggregate.Update(UpdateParams{
		Name:            &name,
		Description:     &description,
		Labels:          &labels,
		Now:             testUpdateTime,
		ExpectedVersion: types.InitialVersion,
	}); err != nil {
		t.Fatalf("Update() error = %v", err)
	}
	labels["team"] = "mutated"
	if aggregate.Name() != name || aggregate.Description() != description {
		t.Fatal("Update() did not apply scalar changes")
	}
	if aggregate.Labels()["team"] != "runtime" {
		t.Fatal("Update() retained caller labels alias")
	}
	if aggregate.Version() != 2 || aggregate.UpdateTime() != testUpdateTime.UTC() {
		t.Fatal("Update() did not advance version and update time")
	}

	var cleared map[string]string
	if err := aggregate.Update(UpdateParams{Labels: &cleared, Now: testUpdateTime.Add(time.Minute)}); err != nil {
		t.Fatalf("Update(clear labels) error = %v", err)
	}
	if aggregate.Labels() != nil {
		t.Fatalf("Labels() = %v, want nil", aggregate.Labels())
	}
}

func TestOrganizationUpdateErrorsAreAtomic(t *testing.T) {
	t.Parallel()

	validName := "renamed"
	longName := strings.Repeat("x", MaxNameRunes+1)
	tests := []struct {
		name    string
		prepare func(*testing.T, *Organization)
		params  UpdateParams
		wantErr error
	}{
		{name: "no fields", params: UpdateParams{Now: testUpdateTime}, wantErr: ErrNoOrganizationUpdates},
		{name: "empty time", params: UpdateParams{Name: &validName}, wantErr: ErrEmptyOrganizationTime},
		{name: "time before last update", params: UpdateParams{Name: &validName, Now: testCreateTime.Add(-time.Second)}, wantErr: ErrTimeBeforeLastUpdate},
		{name: "version mismatch", params: UpdateParams{Name: &validName, Now: testUpdateTime, ExpectedVersion: 42}, wantErr: ErrVersionMismatch},
		{name: "invalid new value", params: UpdateParams{Name: &longName, Now: testUpdateTime}, wantErr: ErrOrganizationNameTooLong},
		{name: "deleted", prepare: deleteOrganization, params: UpdateParams{Name: &validName, Now: testUpdateTime.Add(time.Hour)}, wantErr: ErrOrganizationDeleted},
		{name: "deleting", prepare: func(t *testing.T, o *Organization) { o.state = StateDeleting }, params: UpdateParams{Name: &validName, Now: testUpdateTime}, wantErr: ErrInvalidStateTransition},
		{name: "version overflow", prepare: func(t *testing.T, o *Organization) { o.version = types.Version(math.MaxInt64) }, params: UpdateParams{Name: &validName, Now: testUpdateTime}, wantErr: ErrVersionOverflow},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			aggregate := mustNewOrganization(t)
			if tt.prepare != nil {
				tt.prepare(t, aggregate)
			}
			before := aggregate.Snapshot()
			if err := aggregate.Update(tt.params); !errors.Is(err, tt.wantErr) {
				t.Fatalf("Update() error = %v, want %v", err, tt.wantErr)
			}
			if after := aggregate.Snapshot(); !reflect.DeepEqual(after, before) {
				t.Fatalf("failed Update() mutated aggregate\nafter:  %#v\nbefore: %#v", after, before)
			}
		})
	}
}

func TestOrganizationDelete(t *testing.T) {
	t.Parallel()

	aggregate := mustNewOrganization(t)
	if err := aggregate.Delete(DeleteParams{
		Now:             testUpdateTime,
		PurgeTime:       testPurgeTime,
		ExpectedVersion: types.InitialVersion,
	}); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	if !aggregate.IsDeleted() || aggregate.State() != StateDeleted {
		t.Fatalf("State() = %s, want %s", aggregate.State(), StateDeleted)
	}
	if aggregate.Version() != 2 || aggregate.UpdateTime() != testUpdateTime.UTC() {
		t.Fatal("Delete() did not advance version and update time")
	}
	if got := aggregate.DeleteTime(); got == nil || !got.Equal(testUpdateTime.UTC()) {
		t.Fatalf("DeleteTime() = %v", got)
	}
	if got := aggregate.PurgeTime(); got == nil || !got.Equal(testPurgeTime.UTC()) {
		t.Fatalf("PurgeTime() = %v", got)
	}
}

func TestOrganizationDeleteErrorsAreAtomic(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		prepare func(*testing.T, *Organization)
		params  DeleteParams
		wantErr error
	}{
		{name: "empty time", params: DeleteParams{PurgeTime: testPurgeTime}, wantErr: ErrEmptyOrganizationTime},
		{name: "time before update", params: DeleteParams{Now: testCreateTime.Add(-time.Second), PurgeTime: testPurgeTime}, wantErr: ErrTimeBeforeLastUpdate},
		{name: "empty purge time", params: DeleteParams{Now: testUpdateTime}, wantErr: ErrPurgeTimeRequired},
		{name: "purge equals delete", params: DeleteParams{Now: testUpdateTime, PurgeTime: testUpdateTime}, wantErr: ErrInvalidPurgeTime},
		{name: "version mismatch", params: DeleteParams{Now: testUpdateTime, PurgeTime: testPurgeTime, ExpectedVersion: 9}, wantErr: ErrVersionMismatch},
		{name: "already deleted", prepare: deleteOrganization, params: DeleteParams{Now: testUpdateTime.Add(time.Hour), PurgeTime: testPurgeTime}, wantErr: ErrOrganizationAlreadyDeleted},
		{name: "creating", prepare: func(t *testing.T, o *Organization) { o.state = StateCreating }, params: DeleteParams{Now: testUpdateTime, PurgeTime: testPurgeTime}, wantErr: ErrInvalidStateTransition},
		{name: "version overflow", prepare: func(t *testing.T, o *Organization) { o.version = types.Version(math.MaxInt64) }, params: DeleteParams{Now: testUpdateTime, PurgeTime: testPurgeTime}, wantErr: ErrVersionOverflow},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			aggregate := mustNewOrganization(t)
			if tt.prepare != nil {
				tt.prepare(t, aggregate)
			}
			before := aggregate.Snapshot()
			if err := aggregate.Delete(tt.params); !errors.Is(err, tt.wantErr) {
				t.Fatalf("Delete() error = %v, want %v", err, tt.wantErr)
			}
			if after := aggregate.Snapshot(); !reflect.DeepEqual(after, before) {
				t.Fatalf("failed Delete() mutated aggregate\nafter:  %#v\nbefore: %#v", after, before)
			}
		})
	}
}

func TestOrganizationUndelete(t *testing.T) {
	t.Parallel()

	aggregate := mustNewOrganization(t)
	deleteOrganization(t, aggregate)
	deletedVersion := aggregate.Version()
	restoreTime := testUpdateTime.Add(time.Hour)
	if err := aggregate.Undelete(UndeleteParams{Now: restoreTime, ExpectedVersion: deletedVersion}); err != nil {
		t.Fatalf("Undelete() error = %v", err)
	}
	if aggregate.State() != StateActive || aggregate.IsDeleted() {
		t.Fatalf("State() = %s, want %s", aggregate.State(), StateActive)
	}
	if aggregate.DeleteTime() != nil || aggregate.PurgeTime() != nil {
		t.Fatal("Undelete() did not clear deletion timestamps")
	}
	if aggregate.Version() != deletedVersion+1 || aggregate.UpdateTime() != restoreTime.UTC() {
		t.Fatal("Undelete() did not advance version and update time")
	}
}

func TestOrganizationUndeleteErrorsAreAtomic(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		prepare func(*testing.T, *Organization)
		params  UndeleteParams
		wantErr error
	}{
		{name: "not deleted", prepare: func(t *testing.T, o *Organization) {}, params: UndeleteParams{Now: testUpdateTime}, wantErr: ErrOrganizationNotDeleted},
		{name: "empty time", prepare: deleteOrganization, params: UndeleteParams{}, wantErr: ErrEmptyOrganizationTime},
		{name: "time before delete", prepare: deleteOrganization, params: UndeleteParams{Now: testCreateTime}, wantErr: ErrTimeBeforeLastUpdate},
		{name: "at purge deadline", prepare: deleteOrganization, params: UndeleteParams{Now: testPurgeTime}, wantErr: ErrPurgeTimePassed},
		{name: "version mismatch", prepare: deleteOrganization, params: UndeleteParams{Now: testUpdateTime.Add(time.Hour), ExpectedVersion: 99}, wantErr: ErrVersionMismatch},
		{name: "missing purge time", prepare: func(t *testing.T, o *Organization) { deleteOrganization(t, o); o.purgeTime = nil }, params: UndeleteParams{Now: testUpdateTime.Add(time.Hour)}, wantErr: ErrPurgeTimeRequired},
		{name: "version overflow", prepare: func(t *testing.T, o *Organization) {
			deleteOrganization(t, o)
			o.version = types.Version(math.MaxInt64)
		}, params: UndeleteParams{Now: testUpdateTime.Add(time.Hour)}, wantErr: ErrVersionOverflow},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			aggregate := mustNewOrganization(t)
			tt.prepare(t, aggregate)
			before := aggregate.Snapshot()
			if err := aggregate.Undelete(tt.params); !errors.Is(err, tt.wantErr) {
				t.Fatalf("Undelete() error = %v, want %v", err, tt.wantErr)
			}
			if after := aggregate.Snapshot(); !reflect.DeepEqual(after, before) {
				t.Fatalf("failed Undelete() mutated aggregate\nafter:  %#v\nbefore: %#v", after, before)
			}
		})
	}
}

func TestCheckVersion(t *testing.T) {
	t.Parallel()

	aggregate := mustNewOrganization(t)
	tests := []struct {
		name     string
		expected types.Version
		wantErr  error
	}{
		{name: "omitted", expected: 0},
		{name: "match", expected: types.InitialVersion},
		{name: "mismatch", expected: 2, wantErr: ErrVersionMismatch},
		{name: "outside int64 range", expected: types.Version(math.MaxInt64) + 1, wantErr: ErrInvalidOrganizationVersion},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if err := aggregate.CheckVersion(tt.expected); !errors.Is(err, tt.wantErr) {
				t.Fatalf("CheckVersion() error = %v, want %v", err, tt.wantErr)
			}
		})
	}

	var nilAggregate *Organization
	if err := nilAggregate.CheckVersion(0); !errors.Is(err, ErrNilOrganization) {
		t.Fatalf("nil CheckVersion() error = %v, want %v", err, ErrNilOrganization)
	}
}

func validCreateParams() CreateParams {
	return CreateParams{
		ID:          MustParseID(testOrganizationID),
		Name:        "M8",
		Description: "control plane",
		Labels:      map[string]string{"team": "platform"},
		Now:         testCreateTime,
	}
}

func validSnapshot() Snapshot {
	return Snapshot{
		ID:          MustParseID(testOrganizationID),
		State:       StateActive,
		Name:        "M8",
		Description: "control plane",
		CreateTime:  testCreateTime,
		UpdateTime:  testUpdateTime,
		Version:     types.InitialVersion,
		Labels:      map[string]string{"team": "platform"},
	}
}

func mustNewOrganization(t *testing.T) *Organization {
	t.Helper()
	aggregate, err := New(validCreateParams())
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	return aggregate
}

func deleteOrganization(t *testing.T, aggregate *Organization) {
	t.Helper()
	if err := aggregate.Delete(DeleteParams{Now: testUpdateTime, PurgeTime: testPurgeTime}); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
}
