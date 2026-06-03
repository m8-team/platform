package health

type CheckKind string

const (
	CheckKindLiveness  CheckKind = "LIVENESS"
	CheckKindReadiness CheckKind = "READINESS"
	CheckKindStartup   CheckKind = "STARTUP"
	CheckKindDeep      CheckKind = "DEEP"
)
