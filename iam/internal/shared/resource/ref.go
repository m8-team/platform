package resource

type Ref struct {
	TenantID string
	Type     string
	ID       string
}

func (r Ref) IsZero() bool {
	return r.TenantID == "" && r.Type == "" && r.ID == ""
}
