package principal

type Principal struct {
	TenantID string
	Type     string
	ID       string
}

func (p Principal) IsZero() bool {
	return p.TenantID == "" && p.Type == "" && p.ID == ""
}
