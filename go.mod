module github.com/m8platform/platform

go 1.26.1

require github.com/google/uuid v1.6.0

require (
	github.com/dmarkham/enumer v1.6.3 // indirect
	github.com/pascaldekloe/name v1.0.0 // indirect
	golang.org/x/mod v0.36.0 // indirect
	golang.org/x/sync v0.20.0 // indirect
	golang.org/x/tools v0.45.0 // indirect
)

replace github.com/m8-team/go-genproto => ./api/generate/go

tool golang.org/x/tools/cmd/stringer
