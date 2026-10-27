package orgstructure

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"sort"
	"strings"
	"time"

	"github.com/m8-team/platform/internal/resourcemanager/domain/orgunit"
)

var (
	ErrReadNotConfigured = errors.New("org structure reader not configured")
	ErrInvalidPageToken  = errors.New("invalid or expired org structure page token")
	ErrInvalidList       = errors.New("invalid org structure list request")
	ErrUnitNotVisible    = errors.New("org unit not found or not visible")
)

// HierarchyReader must load adjacency and closure from the SAME immutable
// snapshot ID. YDB Store implements this contract without mixing revisions.
type HierarchyReader interface {
	Load(ctx context.Context, tenantID, organizationID string) (orgunit.Structure, error)
}

// Deny by default. Organizational permission does not imply that every unit
// is visible. Always filter individual units before applying pagination.
type HierarchyReadGate interface {
	AuthorizeOrganization(ctx context.Context, actor Actor, organizationID string) error
	CanReadUnit(ctx context.Context, actor Actor, organizationID, unitID string) (bool, error)
}
type QueryService struct {
	reader   HierarchyReader
	gate     HierarchyReadGate
	tokenKey []byte
}

func NewQueryService(reader HierarchyReader, gate HierarchyReadGate, tokenKey []byte) (*QueryService, error) {
	if reader == nil || gate == nil || len(tokenKey) < 32 {
		return nil, ErrReadNotConfigured
	}
	return &QueryService{reader: reader, gate: gate, tokenKey: append([]byte(nil), tokenKey...)}, nil
}

type UnitResult struct {
	Unit              orgunit.Unit
	StructureRevision uint64
}

func (s *QueryService) Get(ctx context.Context, actor Actor, org, unitID string) (UnitResult, error) {
	if e := s.check(ctx, actor, org); e != nil {
		return UnitResult{}, e
	}
	if unitID == "" {
		return UnitResult{}, orgunit.ErrInvalidUnit
	}
	v, e := s.reader.Load(ctx, actor.TenantID, org)
	if e != nil {
		return UnitResult{}, e
	}
	for _, u := range v.Units {
		if u.ID == unitID {
			ok, e := s.gate.CanReadUnit(ctx, actor, org, unitID)
			if e != nil {
				return UnitResult{}, e
			}
			if !ok {
				return UnitResult{}, ErrUnitNotVisible
			}
			return UnitResult{Unit: u, StructureRevision: v.Revision}, nil
		}
	}
	return UnitResult{}, ErrUnitNotVisible
}

type ListUnitsRequest struct {
	OrganizationID   string
	AncestorID       string
	IncludeSelf      bool
	State            orgunit.State
	ExpectedRevision uint64
	PageSize         int
	PageToken        string
}
type ListUnitsResponse struct {
	Units             []orgunit.Unit
	StructureRevision uint64
	NextPageToken     string
}

const maxUnitPageSize = 200
const defaultUnitPageSize = 50
const cursorTTL = 15 * time.Minute

type unitCursor struct {
	TenantID       string        `json:"t"`
	PrincipalID    string        `json:"p"`
	OrganizationID string        `json:"o"`
	AncestorID     string        `json:"a"`
	IncludeSelf    bool          `json:"s"`
	State          orgunit.State `json:"f"`
	Revision       uint64        `json:"r"`
	LastUnitID     string        `json:"l"`
	ExpiresAt      int64         `json:"e"`
}

// List applies authorization filters BEFORE keyset pagination. The cursor
// binds caller, tenant, organization, filter and active structure revision.
func (s *QueryService) List(ctx context.Context, actor Actor, req ListUnitsRequest) (ListUnitsResponse, error) {
	if e := s.check(ctx, actor, req.OrganizationID); e != nil {
		return ListUnitsResponse{}, e
	}
	if req.PageSize == 0 {
		req.PageSize = defaultUnitPageSize
	}
	if req.PageSize < 1 || req.PageSize > maxUnitPageSize || !validListState(req.State) {
		return ListUnitsResponse{}, ErrInvalidList
	}
	var cursor unitCursor
	if req.PageToken != "" {
		var e error
		cursor, e = s.decode(req.PageToken)
		if e != nil {
			return ListUnitsResponse{}, e
		}
		if cursor.TenantID != actor.TenantID || cursor.PrincipalID != actor.PrincipalID || cursor.OrganizationID != req.OrganizationID || cursor.AncestorID != req.AncestorID || cursor.IncludeSelf != req.IncludeSelf || cursor.State != req.State || cursor.LastUnitID == "" {
			return ListUnitsResponse{}, ErrInvalidPageToken
		}
		if req.ExpectedRevision != 0 && req.ExpectedRevision != cursor.Revision {
			return ListUnitsResponse{}, orgunit.ErrRevisionConflict
		}
	}
	v, e := s.reader.Load(ctx, actor.TenantID, req.OrganizationID)
	if e != nil {
		return ListUnitsResponse{}, e
	}
	if req.ExpectedRevision != 0 && req.ExpectedRevision != v.Revision {
		return ListUnitsResponse{}, orgunit.ErrRevisionConflict
	}
	if req.PageToken != "" && cursor.Revision != v.Revision {
		return ListUnitsResponse{}, orgunit.ErrRevisionConflict
	}
	if req.AncestorID != "" {
		exists := false
		for _, u := range v.Units {
			if u.ID == req.AncestorID {
				exists = true
				break
			}
		}
		if !exists {
			return ListUnitsResponse{}, ErrUnitNotVisible
		}
		ok, e := s.gate.CanReadUnit(ctx, actor, req.OrganizationID, req.AncestorID)
		if e != nil {
			return ListUnitsResponse{}, e
		}
		if !ok {
			return ListUnitsResponse{}, ErrUnitNotVisible
		}
	}
	descendants := map[string]bool{}
	if req.AncestorID != "" {
		for _, r := range v.Closure {
			if r.AncestorID == req.AncestorID && (req.IncludeSelf || r.Depth != 0) {
				descendants[r.DescendantID] = true
			}
		}
	}
	visible := make([]orgunit.Unit, 0, len(v.Units))
	for _, u := range v.Units {
		if req.AncestorID != "" && !descendants[u.ID] {
			continue
		}
		if req.State != "" && u.State != req.State {
			continue
		}
		ok, e := s.gate.CanReadUnit(ctx, actor, req.OrganizationID, u.ID)
		if e != nil {
			return ListUnitsResponse{}, e
		}
		if !ok || (req.PageToken != "" && u.ID <= cursor.LastUnitID) {
			continue
		}
		visible = append(visible, u)
	}
	sort.Slice(visible, func(i, j int) bool { return visible[i].ID < visible[j].ID })
	out := ListUnitsResponse{Units: visible, StructureRevision: v.Revision}
	if len(visible) > req.PageSize {
		out.Units = append([]orgunit.Unit(nil), visible[:req.PageSize]...)
		next := unitCursor{TenantID: actor.TenantID, PrincipalID: actor.PrincipalID, OrganizationID: req.OrganizationID, AncestorID: req.AncestorID, IncludeSelf: req.IncludeSelf, State: req.State, Revision: v.Revision, LastUnitID: out.Units[len(out.Units)-1].ID, ExpiresAt: time.Now().Add(cursorTTL).Unix()}
		out.NextPageToken, e = s.encode(next)
		if e != nil {
			return ListUnitsResponse{}, e
		}
	}
	return out, nil
}

type AncestorsResponse struct {
	Ancestors         []orgunit.ClosureRow
	StructureRevision uint64
}

func (s *QueryService) Ancestors(ctx context.Context, actor Actor, org, unitID string, rev uint64) (AncestorsResponse, error) {
	if e := s.check(ctx, actor, org); e != nil {
		return AncestorsResponse{}, e
	}
	if unitID == "" {
		return AncestorsResponse{}, orgunit.ErrInvalidUnit
	}
	v, e := s.reader.Load(ctx, actor.TenantID, org)
	if e != nil {
		return AncestorsResponse{}, e
	}
	if rev != 0 && rev != v.Revision {
		return AncestorsResponse{}, orgunit.ErrRevisionConflict
	}
	exists := false
	for _, u := range v.Units {
		if u.ID == unitID {
			exists = true
			break
		}
	}
	if !exists {
		return AncestorsResponse{}, ErrUnitNotVisible
	}
	ok, e := s.gate.CanReadUnit(ctx, actor, org, unitID)
	if e != nil {
		return AncestorsResponse{}, e
	}
	if !ok {
		return AncestorsResponse{}, ErrUnitNotVisible
	}
	out := AncestorsResponse{StructureRevision: v.Revision, Ancestors: []orgunit.ClosureRow{}}
	for _, r := range v.Closure {
		if r.DescendantID == unitID {
			ok, e := s.gate.CanReadUnit(ctx, actor, org, r.AncestorID)
			if e != nil {
				return AncestorsResponse{}, e
			}
			if ok {
				out.Ancestors = append(out.Ancestors, r)
			}
		}
	}
	sort.Slice(out.Ancestors, func(i, j int) bool { return out.Ancestors[i].Depth < out.Ancestors[j].Depth })
	return out, nil
}

type DescendantsResponse struct {
	Descendants       []orgunit.ClosureRow
	StructureRevision uint64
	NextPageToken     string
}

func (s *QueryService) Descendants(ctx context.Context, actor Actor, req ListUnitsRequest) (DescendantsResponse, error) {
	if req.AncestorID == "" {
		return DescendantsResponse{}, ErrInvalidList
	}
	list, e := s.List(ctx, actor, req)
	if e != nil {
		return DescendantsResponse{}, e
	}
	v, e := s.reader.Load(ctx, actor.TenantID, req.OrganizationID)
	if e != nil {
		return DescendantsResponse{}, e
	}
	if v.Revision != list.StructureRevision {
		return DescendantsResponse{}, orgunit.ErrRevisionConflict
	}
	depth := map[string]uint32{}
	for _, r := range v.Closure {
		if r.AncestorID == req.AncestorID {
			depth[r.DescendantID] = r.Depth
		}
	}
	out := DescendantsResponse{StructureRevision: v.Revision, NextPageToken: list.NextPageToken, Descendants: make([]orgunit.ClosureRow, 0, len(list.Units))}
	for _, u := range list.Units {
		out.Descendants = append(out.Descendants, orgunit.ClosureRow{OrganizationID: req.OrganizationID, AncestorID: req.AncestorID, DescendantID: u.ID, Depth: depth[u.ID]})
	}
	return out, nil
}
func validListState(s orgunit.State) bool {
	return s == "" || s == orgunit.StatePlanned || s == orgunit.StateActive || s == orgunit.StateClosed
}
func (s *QueryService) check(ctx context.Context, actor Actor, org string) error {
	if s == nil || s.reader == nil || s.gate == nil || len(s.tokenKey) < 32 {
		return ErrReadNotConfigured
	}
	if ctx == nil {
		return ErrInvalidList
	}
	if e := ctx.Err(); e != nil {
		return e
	}
	if actor.TenantID == "" || actor.PrincipalID == "" || org == "" {
		return ErrInvalidList
	}
	return s.gate.AuthorizeOrganization(ctx, actor, org)
}
func (s *QueryService) encode(c unitCursor) (string, error) {
	b, e := json.Marshal(c)
	if e != nil {
		return "", e
	}
	mac := hmac.New(sha256.New, s.tokenKey)
	mac.Write(b)
	return base64.RawURLEncoding.EncodeToString(b) + "." + base64.RawURLEncoding.EncodeToString(mac.Sum(nil)), nil
}
func (s *QueryService) decode(token string) (unitCursor, error) {
	var c unitCursor
	if len(token) > 4096 {
		return c, ErrInvalidPageToken
	}
	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		return c, ErrInvalidPageToken
	}
	b, e := base64.RawURLEncoding.DecodeString(parts[0])
	if e != nil {
		return c, ErrInvalidPageToken
	}
	sig, e := base64.RawURLEncoding.DecodeString(parts[1])
	if e != nil {
		return c, ErrInvalidPageToken
	}
	mac := hmac.New(sha256.New, s.tokenKey)
	mac.Write(b)
	if !hmac.Equal(sig, mac.Sum(nil)) {
		return c, ErrInvalidPageToken
	}
	if e = json.Unmarshal(b, &c); e != nil {
		return c, ErrInvalidPageToken
	}
	if c.Revision == 0 || c.ExpiresAt <= time.Now().Unix() {
		return c, ErrInvalidPageToken
	}
	return c, nil
}
