package v1engine

import (
	"crypto/subtle"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
)

// Handler is an explicitly local-only HTTP adapter, not production IAM.
// No token => deny all. Bind it only to localhost in development.
type Handler struct {
	Store    *Store
	DevToken string
}

func NewHandler(s *Store, devToken string) http.Handler {
	return &Handler{Store: s, DevToken: devToken}
}
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if h.DevToken == "" || subtle.ConstantTimeCompare([]byte(r.Header.Get("Authorization")), []byte("Bearer "+h.DevToken)) != 1 {
		respond(w, 401, map[string]string{"error": "unauthorized"})
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	path := strings.TrimPrefix(r.URL.Path, "/resource-manager/v1/")
	if !strings.HasPrefix(r.URL.Path, "/resource-manager/v1/") {
		http.NotFound(w, r)
		return
	}
	switch {
	case path == "resources" && r.Method == http.MethodPost:
		h.create(w, r)
	case path == "resources" && r.Method == http.MethodGet:
		h.list(w, r)
	case strings.HasPrefix(path, "resources/") && r.Method == http.MethodGet:
		h.get(w, r, strings.TrimPrefix(path, "resources/"))
	case strings.HasPrefix(path, "resources/") && r.Method == http.MethodPatch:
		h.update(w, r, strings.TrimPrefix(path, "resources/"))
	case strings.HasPrefix(path, "resources/") && r.Method == http.MethodPost && strings.HasSuffix(path, ":changeState"):
		h.state(w, r, strings.TrimSuffix(strings.TrimPrefix(path, "resources/"), ":changeState"))
	case strings.HasPrefix(path, "projects/") && strings.HasSuffix(path, ":move") && r.Method == http.MethodPost:
		h.move(w, r, strings.TrimSuffix(strings.TrimPrefix(path, "projects/"), ":move"))
	case path == "group-affiliations" && r.Method == http.MethodPost:
		h.group(w, r)
	case path == "environment-bindings" && r.Method == http.MethodPost:
		h.binding(w, r)
	default:
		http.NotFound(w, r)
	}
}
func decode(r *http.Request, out any) error {
	defer r.Body.Close()
	d := json.NewDecoder(io.LimitReader(r.Body, 1<<20))
	d.DisallowUnknownFields()
	if e := d.Decode(out); e != nil {
		return e
	}
	if e := d.Decode(new(any)); e != io.EOF {
		return ErrInvalid
	}
	return nil
}
func respond(w http.ResponseWriter, code int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(data)
}
func failure(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrInvalid):
		respond(w, 400, map[string]string{"error": err.Error()})
	case errors.Is(err, ErrNotFound):
		respond(w, 404, map[string]string{"error": err.Error()})
	case errors.Is(err, ErrConflict):
		respond(w, 409, map[string]string{"error": err.Error()})
	default:
		respond(w, 500, map[string]string{"error": "internal error"})
	}
}

type createRequest struct {
	Resource       json.RawMessage `json:"resource"`
	IdempotencyKey string          `json:"idempotencyKey"`
}
type metadataInput struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}
type resourceInput struct {
	Metadata       metadataInput     `json:"metadata"`
	TenantID       string            `json:"tenantId"`
	OrganizationID string            `json:"organizationId"`
	Owner          map[string]string `json:"owner"`
	Key            string            `json:"key"`
}

func parseRecord(raw json.RawMessage) (Record, error) {
	var m map[string]json.RawMessage
	if err := json.Unmarshal(raw, &m); err != nil || len(m) != 1 {
		return Record{}, ErrInvalid
	}
	var kind Kind
	var value json.RawMessage
	for name, v := range m {
		value = v
		switch name {
		case "tenant":
			kind = Tenant
		case "group":
			kind = Group
		case "organization":
			kind = Organization
		case "workspace":
			kind = Workspace
		case "project":
			kind = Project
		case "environment":
			kind = Environment
		case "workloadService":
			kind = WorkloadService
		case "descriptor":
			kind = Descriptor
		default:
			return Record{}, ErrInvalid
		}
	}
	var in resourceInput
	if err := json.Unmarshal(value, &in); err != nil {
		return Record{}, ErrInvalid
	}
	r := Record{Kind: kind, Name: in.Metadata.Name, Description: in.Metadata.Description, EnvironmentKey: in.Key}
	switch kind {
	case Group, Organization:
		r.Owner = Owner{Tenant, in.TenantID}
	case Workspace, Environment:
		r.Owner = Owner{Organization, in.OrganizationID}
	case Project, WorkloadService, Descriptor:
		for k, v := range in.Owner {
			switch k {
			case "organizationId":
				r.Owner = Owner{Organization, v}
			case "workspaceId":
				r.Owner = Owner{Workspace, v}
			case "projectId":
				r.Owner = Owner{Project, v}
			default:
				return Record{}, ErrInvalid
			}
		}
	default:
	}
	return r, nil
}
func payload(r Record) any {
	m := map[string]any{"id": r.ID, "name": r.Name, "description": r.Description, "state": "RESOURCE_STATE_" + r.State, "revision": r.Revision, "createTime": r.CreatedAt, "updateTime": r.UpdatedAt}
	v := map[string]any{"metadata": m}
	owner := map[string]string{}
	switch r.Owner.Kind {
	case Tenant:
		owner["tenantId"] = r.Owner.ID
	case Organization:
		owner["organizationId"] = r.Owner.ID
	case Workspace:
		owner["workspaceId"] = r.Owner.ID
	case Project:
		owner["projectId"] = r.Owner.ID
	}
	var variant string
	switch r.Kind {
	case Tenant:
		variant = "tenant"
	case Group:
		variant = "group"
		v["tenantId"] = r.TenantID
	case Organization:
		variant = "organization"
		v["tenantId"] = r.TenantID
	case Workspace:
		variant = "workspace"
		v["organizationId"] = r.OrganizationID
	case Project:
		variant = "project"
		v["owner"] = owner
		v["owningOrganizationId"] = r.OrganizationID
	case Environment:
		variant = "environment"
		v["organizationId"] = r.OrganizationID
		v["key"] = r.EnvironmentKey
		label := strings.ToUpper(r.EnvironmentKey)
		switch label {
		case "PROD":
			label = "PRODUCTION"
		case "DEV":
			label = "DEVELOPMENT"
		case "TEST":
			label = "TESTING"
		}
		switch label {
		case "PRODUCTION", "STAGING", "DEVELOPMENT", "TESTING", "PREVIEW":
			v["kind"] = "ENVIRONMENT_KIND_" + label
		default:
			v["kind"] = "ENVIRONMENT_KIND_CUSTOM"
		}
		v["isolation"] = "ISOLATION_PROFILE_LOGICAL"
	case WorkloadService:
		variant = "workloadService"
		v["owner"] = owner
		v["owningOrganizationId"] = r.OrganizationID
	case Descriptor:
		variant = "descriptor"
		v["owner"] = owner
		v["owningOrganizationId"] = r.OrganizationID
	default:
		return nil
	}
	return map[string]any{variant: v}
}
func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	var req createRequest
	if e := decode(r, &req); e != nil {
		failure(w, ErrInvalid)
		return
	}
	rec, e := parseRecord(req.Resource)
	if e != nil {
		failure(w, e)
		return
	}
	created, e := h.Store.Create(rec, req.IdempotencyKey)
	if e != nil {
		failure(w, e)
		return
	}
	respond(w, 201, payload(created))
}
func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	items, e := h.Store.List(Kind(r.URL.Query().Get("kind")), r.URL.Query().Get("organizationId"))
	if e != nil {
		failure(w, e)
		return
	}
	out := []any{}
	for _, it := range items {
		out = append(out, payload(it))
	}
	respond(w, 200, map[string]any{"resources": out})
}
func splitRef(v string) (Kind, string, bool) {
	p := strings.Split(v, "/")
	if len(p) != 2 || p[0] == "" || p[1] == "" {
		return "", "", false
	}
	return Kind(p[0]), p[1], true
}
func (h *Handler) get(w http.ResponseWriter, r *http.Request, part string) {
	if strings.HasSuffix(part, ":resolveScope") {
		k, id, ok := splitRef(strings.TrimSuffix(part, ":resolveScope"))
		if !ok {
			failure(w, ErrInvalid)
			return
		}
		anc, e := h.Store.Resolve(k, id)
		if e != nil {
			failure(w, e)
			return
		}
		respond(w, 200, map[string]any{"ancestry": anc})
		return
	}
	k, id, ok := splitRef(part)
	if !ok {
		failure(w, ErrInvalid)
		return
	}
	o, e := h.Store.Get(k, id)
	if e != nil {
		failure(w, e)
		return
	}
	respond(w, 200, payload(o))
}
func (h *Handler) update(w http.ResponseWriter, r *http.Request, part string) {
	var req struct {
		Resource         json.RawMessage `json:"resource"`
		ExpectedRevision int64           `json:"expectedRevision"`
	}
	if decode(r, &req) != nil {
		failure(w, ErrInvalid)
		return
	}
	rec, e := parseRecord(req.Resource)
	if e != nil {
		failure(w, e)
		return
	}
	o, e := h.Store.Update(rec.Kind, part, req.ExpectedRevision, rec.Name, rec.Description)
	if e != nil {
		failure(w, e)
		return
	}
	respond(w, 200, payload(o))
}
func (h *Handler) state(w http.ResponseWriter, r *http.Request, id string) {
	var req struct {
		Kind             Kind   `json:"kind"`
		TargetState      string `json:"targetState"`
		ExpectedRevision int64  `json:"expectedRevision"`
	}
	if decode(r, &req) != nil {
		failure(w, ErrInvalid)
		return
	}
	o, e := h.Store.ChangeState(req.Kind, id, req.ExpectedRevision, strings.TrimPrefix(req.TargetState, "RESOURCE_STATE_"))
	if e != nil {
		failure(w, e)
		return
	}
	respond(w, 200, payload(o))
}
func (h *Handler) move(w http.ResponseWriter, r *http.Request, id string) {
	var req struct {
		NewOwner         map[string]string `json:"newOwner"`
		ExpectedRevision int64             `json:"expectedRevision"`
	}
	if decode(r, &req) != nil {
		failure(w, ErrInvalid)
		return
	}
	var owner Owner
	for k, v := range req.NewOwner {
		switch k {
		case "organizationId":
			owner = Owner{Organization, v}
		case "workspaceId":
			owner = Owner{Workspace, v}
		default:
			failure(w, ErrInvalid)
			return
		}
	}
	o, e := h.Store.MoveProject(id, owner, req.ExpectedRevision)
	if e != nil {
		failure(w, e)
		return
	}
	respond(w, 200, payload(o))
}
func (h *Handler) group(w http.ResponseWriter, r *http.Request) {
	var req struct {
		GroupID        string `json:"groupId"`
		OrganizationID string `json:"organizationId"`
	}
	if decode(r, &req) != nil {
		failure(w, ErrInvalid)
		return
	}
	if err := h.Store.SetGroupLink(req.GroupID, req.OrganizationID); err != nil {
		failure(w, err)
		return
	}
	respond(w, 200, map[string]any{"groupId": req.GroupID, "organizationId": req.OrganizationID})
}
func (h *Handler) binding(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Resource         map[string]string `json:"resource"`
		EnvironmentID    string            `json:"environmentId"`
		VersionRef       string            `json:"versionRef"`
		ConfigurationRef string            `json:"configurationRef"`
	}
	if decode(r, &req) != nil {
		failure(w, ErrInvalid)
		return
	}
	k, id := Kind(req.Resource["kind"]), req.Resource["id"]
	b, e := h.Store.SetBinding(Owner{k, id}, req.EnvironmentID, req.VersionRef, req.ConfigurationRef)
	if e != nil {
		failure(w, e)
		return
	}
	respond(w, 200, b)
}
