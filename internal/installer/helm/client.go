package helm

import "context"

type Release struct {
	Name       string
	Namespace  string
	Chart      string
	Repository string
	Version    string
	Digest     string
	Values     map[string]any
}

type Status struct {
	Name      string `json:"name" yaml:"name"`
	Namespace string `json:"namespace" yaml:"namespace"`
	Revision  int    `json:"revision" yaml:"revision"`
	Phase     string `json:"phase" yaml:"phase"`
}

type Client interface {
	Plan(ctx context.Context, release Release) (ChangeSet, error)
	Apply(ctx context.Context, release Release) error
	Status(ctx context.Context, namespace string, name string) (Status, error)
	Rollback(ctx context.Context, namespace string, name string, revision int) error
}

type ChangeSet struct {
	Create []string `json:"create,omitempty" yaml:"create,omitempty"`
	Update []string `json:"update,omitempty" yaml:"update,omitempty"`
	Delete []string `json:"delete,omitempty" yaml:"delete,omitempty"`
}
