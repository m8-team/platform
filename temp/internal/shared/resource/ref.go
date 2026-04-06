package resource

type Ref struct {
	TenantID string
	Type     string
	ID       string
}

func (r Ref) IsZero() bool {
	return r.TenantID == "" && r.Type == "" && r.ID == ""
}

func (r Ref) Equals(other Ref) bool {
	return r.Type == other.Type && r.ID == other.ID
}
