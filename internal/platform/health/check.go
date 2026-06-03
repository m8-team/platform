package health

import "time"

type Check struct {
	Name        string
	Target      Target
	Kinds       []CheckKind
	Criticality Criticality
	Timeout     time.Duration
	Interval    time.Duration
	Checker     Checker
}
