package health

import "time"

type Result struct {
	Name        string            `json:"name"`
	Status      Status            `json:"status"`
	Message     string            `json:"message,omitempty"`
	Error       string            `json:"error,omitempty"`
	Latency     time.Duration     `json:"latency"`
	CheckedAt   time.Time         `json:"checked_at"`
	Target      Target            `json:"target"`
	Criticality Criticality       `json:"criticality"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}
