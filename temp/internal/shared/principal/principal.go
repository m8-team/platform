package principal

type Principal struct {
	TenantID string
	Type     string
	ID       string
}

func (p Principal) IsZero() bool {
	return p.TenantID == "" && p.Type == "" && p.ID == ""
}

func (p Principal) Equals(other Principal) bool {
	return p.Type == other.Type && p.ID == other.ID
}
