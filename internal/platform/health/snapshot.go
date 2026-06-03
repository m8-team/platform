package health

import "time"

type Snapshot struct {
	Status    Status    `json:"status"`
	Kind      CheckKind `json:"kind"`
	CheckedAt time.Time `json:"checked_at"`
	Results   []Result  `json:"results"`
}
