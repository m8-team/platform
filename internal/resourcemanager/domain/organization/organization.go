package organization

import (
	"fmt"
	"time"
	"unicode/utf8"

	"github.com/m8-team/platform/internal/platform/types"
)

const (
	ResourceType        = "resourcemanager.organization"
	MaxNameRunes        = 256
	MaxDescriptionRunes = 1024
)

type Organization struct {
	id          ID
	state       State
	name        string
	description string
	createTime  time.Time
	updateTime  time.Time
	deleteTime  *time.Time
	purgeTime   *time.Time
	version     types.Version
	labels      map[string]string
}

type CreateParams struct {
	ID          ID
	Name        string
	Description string
	Labels      map[string]string
	Now         time.Time
}

// Snapshot is the persistence-safe representation of an Organization.
// Its maps and timestamp pointers never alias the aggregate.
type Snapshot struct {
	ID          ID
	State       State
	Name        string
	Description string
	CreateTime  time.Time
	UpdateTime  time.Time
	DeleteTime  *time.Time
	PurgeTime   *time.Time
	Version     types.Version
	Labels      map[string]string
}

type UpdateParams struct {
	Name            *string
	Description     *string
	Labels          *map[string]string
	Now             time.Time
	ExpectedVersion types.Version
}

type DeleteParams struct {
	Now             time.Time
	PurgeTime       time.Time
	ExpectedVersion types.Version
}

type UndeleteParams struct {
	Now             time.Time
	ExpectedVersion types.Version
}

func New(params CreateParams) (*Organization, error) {
	if err := params.ID.Validate(); err != nil {
		return nil, err
	}
	if err := validateMutableFields(params.Name, params.Description, params.Labels); err != nil {
		return nil, err
	}
	if err := validateNonZeroTime(params.Now); err != nil {
		return nil, err
	}

	now := params.Now.UTC()
	return &Organization{
		id:          params.ID,
		state:       StateActive,
		name:        params.Name,
		description: params.Description,
		createTime:  now,
		updateTime:  now,
		version:     types.NewInitialVersion(),
		labels:      cloneLabels(params.Labels),
	}, nil
}

// Rehydrate restores an Organization from a trusted persistence snapshot while
// still enforcing all aggregate invariants.
func Rehydrate(snapshot Snapshot) (*Organization, error) {
	if err := validateSnapshot(snapshot); err != nil {
		return nil, err
	}

	return &Organization{
		id:          snapshot.ID,
		state:       snapshot.State,
		name:        snapshot.Name,
		description: snapshot.Description,
		createTime:  snapshot.CreateTime.UTC(),
		updateTime:  snapshot.UpdateTime.UTC(),
		deleteTime:  cloneTime(snapshot.DeleteTime),
		purgeTime:   cloneTime(snapshot.PurgeTime),
		version:     snapshot.Version,
		labels:      cloneLabels(snapshot.Labels),
	}, nil
}

func (o *Organization) Snapshot() Snapshot {
	if o == nil {
		return Snapshot{}
	}

	return Snapshot{
		ID:          o.id,
		State:       o.state,
		Name:        o.name,
		Description: o.description,
		CreateTime:  o.createTime,
		UpdateTime:  o.updateTime,
		DeleteTime:  cloneTime(o.deleteTime),
		PurgeTime:   cloneTime(o.purgeTime),
		Version:     o.version,
		Labels:      cloneLabels(o.labels),
	}
}

func (o *Organization) Clone() *Organization {
	if o == nil {
		return nil
	}

	return &Organization{
		id:          o.id,
		state:       o.state,
		name:        o.name,
		description: o.description,
		createTime:  o.createTime,
		updateTime:  o.updateTime,
		deleteTime:  cloneTime(o.deleteTime),
		purgeTime:   cloneTime(o.purgeTime),
		version:     o.version,
		labels:      cloneLabels(o.labels),
	}
}

func (o *Organization) ID() ID                    { return o.id }
func (o *Organization) State() State              { return o.state }
func (o *Organization) Name() string              { return o.name }
func (o *Organization) Description() string       { return o.description }
func (o *Organization) CreateTime() time.Time     { return o.createTime }
func (o *Organization) UpdateTime() time.Time     { return o.updateTime }
func (o *Organization) DeleteTime() *time.Time    { return cloneTime(o.deleteTime) }
func (o *Organization) PurgeTime() *time.Time     { return cloneTime(o.purgeTime) }
func (o *Organization) Version() types.Version    { return o.version }
func (o *Organization) Labels() map[string]string { return cloneLabels(o.labels) }
func (o *Organization) IsDeleted() bool           { return o.state == StateDeleted }

// CheckVersion checks an optional optimistic concurrency precondition. Zero
// means that the caller did not supply a precondition.
func (o *Organization) CheckVersion(expected types.Version) error {
	if o == nil {
		return ErrNilOrganization
	}
	if expected.IsZero() {
		return nil
	}
	if err := validateVersion(expected); err != nil {
		return err
	}
	if !o.version.Equal(expected) {
		return fmt.Errorf("%w: expected %s, current %s", ErrVersionMismatch, expected, o.version)
	}

	return nil
}

func (o *Organization) Update(params UpdateParams) error {
	if o == nil {
		return ErrNilOrganization
	}
	if o.state == StateDeleted {
		return ErrOrganizationDeleted
	}
	if !canUpdate(o.state) {
		return fmt.Errorf("%w: cannot update from %s", ErrInvalidStateTransition, o.state)
	}
	if params.Name == nil && params.Description == nil && params.Labels == nil {
		return ErrNoOrganizationUpdates
	}
	if err := validateMutationTime(params.Now, o.updateTime); err != nil {
		return err
	}
	if err := o.CheckVersion(params.ExpectedVersion); err != nil {
		return err
	}

	name := o.name
	if params.Name != nil {
		name = *params.Name
	}
	description := o.description
	if params.Description != nil {
		description = *params.Description
	}
	labels := o.labels
	if params.Labels != nil {
		labels = *params.Labels
	}
	if err := validateMutableFields(name, description, labels); err != nil {
		return err
	}

	nextVersion, err := nextVersion(o.version)
	if err != nil {
		return err
	}

	o.name = name
	o.description = description
	if params.Labels != nil {
		o.labels = cloneLabels(labels)
	}
	o.updateTime = params.Now.UTC()
	o.version = nextVersion
	return nil
}

func (o *Organization) Delete(params DeleteParams) error {
	if o == nil {
		return ErrNilOrganization
	}
	if o.state == StateDeleted {
		return ErrOrganizationAlreadyDeleted
	}
	if !canDelete(o.state) {
		return fmt.Errorf("%w: cannot delete from %s", ErrInvalidStateTransition, o.state)
	}
	if err := validateMutationTime(params.Now, o.updateTime); err != nil {
		return err
	}
	if err := validateNonZeroTime(params.PurgeTime); err != nil {
		return fmt.Errorf("%w: %v", ErrPurgeTimeRequired, err)
	}
	if !params.PurgeTime.After(params.Now) {
		return ErrInvalidPurgeTime
	}
	if err := o.CheckVersion(params.ExpectedVersion); err != nil {
		return err
	}

	nextVersion, err := nextVersion(o.version)
	if err != nil {
		return err
	}
	deleteTime := params.Now.UTC()
	purgeTime := params.PurgeTime.UTC()
	o.state = StateDeleted
	o.deleteTime = &deleteTime
	o.purgeTime = &purgeTime
	o.updateTime = deleteTime
	o.version = nextVersion
	return nil
}

func (o *Organization) Undelete(params UndeleteParams) error {
	if o == nil {
		return ErrNilOrganization
	}
	if o.state != StateDeleted {
		return ErrOrganizationNotDeleted
	}
	if err := validateMutationTime(params.Now, o.updateTime); err != nil {
		return err
	}
	if o.purgeTime == nil {
		return ErrPurgeTimeRequired
	}
	if !params.Now.Before(*o.purgeTime) {
		return ErrPurgeTimePassed
	}
	if err := o.CheckVersion(params.ExpectedVersion); err != nil {
		return err
	}

	nextVersion, err := nextVersion(o.version)
	if err != nil {
		return err
	}
	o.state = StateActive
	o.deleteTime = nil
	o.purgeTime = nil
	o.updateTime = params.Now.UTC()
	o.version = nextVersion
	return nil
}

func validateSnapshot(snapshot Snapshot) error {
	if err := snapshot.ID.Validate(); err != nil {
		return err
	}
	if !snapshot.State.IsValid() {
		return fmt.Errorf("%w: %s", ErrInvalidOrganizationState, snapshot.State)
	}
	if err := validateMutableFields(snapshot.Name, snapshot.Description, snapshot.Labels); err != nil {
		return err
	}
	if err := validateVersion(snapshot.Version); err != nil {
		return err
	}
	if err := validateNonZeroTime(snapshot.CreateTime); err != nil {
		return fmt.Errorf("%w: create_time", err)
	}
	if err := validateNonZeroTime(snapshot.UpdateTime); err != nil {
		return fmt.Errorf("%w: update_time", err)
	}
	if snapshot.UpdateTime.Before(snapshot.CreateTime) {
		return fmt.Errorf("%w: update_time precedes create_time", ErrInvalidOrganizationTime)
	}

	if snapshot.State != StateDeleted {
		if snapshot.DeleteTime != nil || snapshot.PurgeTime != nil {
			return ErrUnexpectedDeletionTimes
		}
		return nil
	}

	if snapshot.DeleteTime == nil || snapshot.DeleteTime.IsZero() {
		return ErrDeleteTimeRequired
	}
	if snapshot.PurgeTime == nil || snapshot.PurgeTime.IsZero() {
		return ErrPurgeTimeRequired
	}
	if snapshot.DeleteTime.Before(snapshot.CreateTime) || snapshot.UpdateTime.Before(*snapshot.DeleteTime) {
		return fmt.Errorf("%w: delete_time is outside the resource lifetime", ErrInvalidOrganizationTime)
	}
	if !snapshot.PurgeTime.After(*snapshot.DeleteTime) {
		return ErrInvalidPurgeTime
	}

	return nil
}

func validateMutableFields(name, description string, labels map[string]string) error {
	if !utf8.ValidString(name) {
		return ErrInvalidOrganizationName
	}
	if utf8.RuneCountInString(name) > MaxNameRunes {
		return ErrOrganizationNameTooLong
	}
	if !utf8.ValidString(description) {
		return ErrInvalidOrganizationDescription
	}
	if utf8.RuneCountInString(description) > MaxDescriptionRunes {
		return ErrOrganizationDescriptionTooLong
	}
	for key, value := range labels {
		if !utf8.ValidString(key) || !utf8.ValidString(value) {
			return fmt.Errorf("%w: key %q", ErrInvalidOrganizationLabel, key)
		}
	}

	return nil
}

func validateNonZeroTime(value time.Time) error {
	if value.IsZero() {
		return ErrEmptyOrganizationTime
	}
	return nil
}

func validateMutationTime(now, lastUpdate time.Time) error {
	if err := validateNonZeroTime(now); err != nil {
		return err
	}
	if now.Before(lastUpdate) {
		return ErrTimeBeforeLastUpdate
	}
	return nil
}

func validateVersion(version types.Version) error {
	if err := version.Validate(); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidOrganizationVersion, err)
	}
	return nil
}

func nextVersion(current types.Version) (types.Version, error) {
	next, err := current.Next()
	if err != nil {
		return 0, fmt.Errorf("%w: %v", ErrVersionOverflow, err)
	}
	return next, nil
}

func canUpdate(state State) bool {
	switch state {
	case StateCreating, StateActive, StateSuspended, StateFailed:
		return true
	default:
		return false
	}
}

func canDelete(state State) bool {
	switch state {
	case StateActive, StateSuspended, StateFailed:
		return true
	default:
		return false
	}
}

func cloneLabels(input map[string]string) map[string]string {
	if input == nil {
		return nil
	}
	result := make(map[string]string, len(input))
	for key, value := range input {
		result[key] = value
	}
	return result
}

func cloneTime(input *time.Time) *time.Time {
	if input == nil {
		return nil
	}
	value := input.UTC()
	return &value
}
