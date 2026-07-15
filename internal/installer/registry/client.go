package registry

import (
	"context"
	"io"
)

type Artifact struct {
	Repository string
	Reference  string
	Digest     string
	MediaType  string
}

type Client interface {
	Resolve(ctx context.Context, repository string, reference string) (Artifact, error)
	Pull(ctx context.Context, artifact Artifact) (io.ReadCloser, error)
	Push(ctx context.Context, repository string, content io.Reader, mediaType string) (Artifact, error)
	VerifyDigest(ctx context.Context, artifact Artifact) error
}
