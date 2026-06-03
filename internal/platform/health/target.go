package health

type TargetKind string

const (
	TargetApplication TargetKind = "APPLICATION"
	TargetModule      TargetKind = "MODULE"
	TargetService     TargetKind = "SERVICE"
	TargetDependency  TargetKind = "DEPENDENCY"
)

type Criticality string

const (
	CriticalityRequired Criticality = "REQUIRED"
	CriticalityOptional Criticality = "OPTIONAL"
)

type Target struct {
	Kind   TargetKind `json:"kind"`
	Name   string     `json:"name"`
	Module string     `json:"module,omitempty"`
}
