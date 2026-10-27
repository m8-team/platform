package orgunit

import (
    "errors"
    "fmt"
    "math"
    "strings"
    "time"
    "unicode/utf8"
)

var (
    ErrInvalidChangeSet = errors.New("invalid org unit change set")
    ErrConflictingChanges = errors.New("conflicting mutations of the same org unit")
    ErrUnitAlreadyExists = errors.New("org unit already exists")
    ErrUnitHasActiveChildren = errors.New("org unit contains active children")
    ErrSourceAuthority = errors.New("org unit is owned by another source")
    ErrDeferredEffectiveTime = errors.New("scheduled or backdated changes require temporal reconciler")
)

type ChangeKind string
const (
    ChangeCreate ChangeKind = "CREATE"
    ChangeMove ChangeKind = "MOVE"
    ChangeRename ChangeKind = "RENAME"
    ChangeClose ChangeKind = "CLOSE"
)

type Change struct {
    Kind ChangeKind
    UnitID string
    ParentID string
    DisplayName string
    AuthoritativeSourceID string
}

// ChangeSet is a proposed revision-guarded mutation of one organization.
// Authorization and cross-module impact checks belong to the application.
type ChangeSet struct {
    ID string
    OrganizationID string
    ExpectedRevision uint64
    EffectiveAt time.Time
    RecordedAt time.Time
    SourceID string
    Changes []Change
}

// Apply checks every final parent relationship before publishing a snapshot.
// Only immediate effective-time changes are supported in this first write slice.
func Apply(current Structure, plan ChangeSet) (Structure, error) {
    if plan.ID == "" || plan.OrganizationID == "" || plan.OrganizationID != current.OrganizationID ||
        plan.ExpectedRevision == 0 || len(plan.Changes) == 0 || len(plan.Changes) > 1000 ||
        plan.EffectiveAt.IsZero() || plan.RecordedAt.IsZero() {
        return Structure{}, ErrInvalidChangeSet
    }
    if plan.ExpectedRevision != current.Revision { return Structure{}, ErrRevisionConflict }
    if !plan.EffectiveAt.Equal(plan.RecordedAt) { return Structure{}, ErrDeferredEffectiveTime }
    if current.Revision == math.MaxUint64 { return Structure{}, ErrRevisionOverflow }
    units := append([]Unit(nil), current.Units...)
    index := make(map[string]int,len(units))
    for i,u := range units { index[u.ID]=i }
    seen := make(map[string]struct{},len(plan.Changes))
    changed := false
    for _,m := range plan.Changes {
        if m.UnitID == "" { return Structure{}, ErrInvalidChangeSet }
        if _,exists:=seen[m.UnitID];exists {return Structure{},ErrConflictingChanges}
        seen[m.UnitID]=struct{}{}
        idx,exists:=index[m.UnitID]
        if m.Kind==ChangeCreate {
            if exists {return Structure{},ErrUnitAlreadyExists}
            if !validName(m.DisplayName) {return Structure{},ErrInvalidUnit}
            if m.AuthoritativeSourceID!="" && m.AuthoritativeSourceID!=plan.SourceID {return Structure{},ErrSourceAuthority}
            source:=m.AuthoritativeSourceID
            if source=="" {source=plan.SourceID}
            u:=Unit{ID:m.UnitID,OrganizationID:plan.OrganizationID,ParentID:m.ParentID,State:StateActive,DisplayName:m.DisplayName,AuthoritativeSourceID:source}
            units=append(units,u);index[u.ID]=len(units)-1;changed=true
            continue
        }
        if !exists {return Structure{},fmt.Errorf("%w: %s",ErrUnitNotFound,m.UnitID)}
        u:=&units[idx]
        if u.State!=StateActive{return Structure{},ErrUnitInactive}
        if u.AuthoritativeSourceID!=plan.SourceID {return Structure{},ErrSourceAuthority}
        switch m.Kind {
        case ChangeMove:
            if u.ParentID!=m.ParentID {u.ParentID=m.ParentID;changed=true}
        case ChangeRename:
            if !validName(m.DisplayName) {return Structure{},ErrInvalidUnit}
            if u.DisplayName!=m.DisplayName {u.DisplayName=m.DisplayName;changed=true}
        case ChangeClose:
            u.State=StateClosed;changed=true
        default:
            return Structure{},ErrInvalidChangeSet
        }
    }
    for _,u:=range units {
        if u.ParentID=="" {continue}
        parentIndex,ok:=index[u.ParentID]
        if !ok {return Structure{},ErrMissingParent}
        parent:=units[parentIndex]
        if parent.OrganizationID!=u.OrganizationID {return Structure{},ErrCrossOrganization}
        if u.State==StateActive && parent.State!=StateActive {return Structure{},ErrUnitHasActiveChildren}
    }
    revision:=current.Revision
    if changed {revision++}
    return Build(current.OrganizationID,revision,units)
}
func validName(v string) bool {
    v=strings.TrimSpace(v)
    return v!="" && utf8.ValidString(v) && utf8.RuneCountInString(v)<=256
}
