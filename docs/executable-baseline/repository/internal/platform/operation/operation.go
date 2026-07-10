package operation

import "time"

type State string

const (
	Pending   State = "PENDING"
	Running   State = "RUNNING"
	Succeeded State = "SUCCEEDED"
	Failed    State = "FAILED"
	Cancelled State = "CANCELLED"
)

type Operation struct {
	ID              string
	Type            string
	State           State
	ProgressPercent int
	Stage           string
	ResultID        string
	ErrorCode       string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}
