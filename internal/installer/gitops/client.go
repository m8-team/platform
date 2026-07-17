package gitops

import "context"

type Application struct {
	Name      string
	Namespace string
	Project   string
	Path      string
	Revision  string
	Wave      int
}

type SyncStatus struct {
	Name     string `json:"name" yaml:"name"`
	Healthy  bool   `json:"healthy" yaml:"healthy"`
	Synced   bool   `json:"synced" yaml:"synced"`
	Message  string `json:"message,omitempty" yaml:"message,omitempty"`
	Revision string `json:"revision,omitempty" yaml:"revision,omitempty"`
}

type Client interface {
	ApplyRoot(ctx context.Context, projectYAML []byte, applicationSetYAML []byte) error
	Sync(ctx context.Context, application string) error
	WaitHealthy(ctx context.Context, application string) (SyncStatus, error)
}
