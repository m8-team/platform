module github.com/m8platform/platform

go 1.26.1

require (
	github.com/google/uuid v1.6.0
	go.uber.org/automaxprocs v1.6.0
	go.uber.org/fx v1.24.0
	google.golang.org/grpc v1.80.0
)

require (
	go.uber.org/dig v1.19.0 // indirect
	go.uber.org/multierr v1.10.0 // indirect
	go.uber.org/zap v1.26.0 // indirect
	golang.org/x/mod v0.36.0 // indirect
	golang.org/x/net v0.54.0 // indirect
	golang.org/x/sync v0.20.0 // indirect
	golang.org/x/sys v0.44.0 // indirect
	golang.org/x/text v0.37.0 // indirect
	golang.org/x/tools v0.45.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260120221211-b8f7ae30c516 // indirect
	google.golang.org/protobuf v1.36.11 // indirect
)

replace github.com/m8-team/go-genproto => ./api/generate/go

tool golang.org/x/tools/cmd/stringer
