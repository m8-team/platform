package requestmetadata

import "strings"

// Source identifies where request metadata originated.
type Source string

const (
	SourceUnknown Source = "unknown"
	SourceHTTP    Source = "http"
	SourceGRPC    Source = "grpc"
	SourceCLI     Source = "cli"
	SourceWorker  Source = "worker"
)

func normalizeSource(source Source) Source {
	normalized := Source(strings.ToLower(strings.TrimSpace(string(source))))
	if normalized == "" {
		return SourceUnknown
	}
	return normalized
}

func isValidSource(source Source) bool {
	switch source {
	case SourceUnknown, SourceHTTP, SourceGRPC, SourceCLI, SourceWorker:
		return true
	default:
		return false
	}
}
