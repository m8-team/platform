.PHONY: test race vet proto local integration run
test:
	go test ./...
	cd ../platform-sdk && go test ./...
race:
	go test -race ./...
	cd ../platform-sdk && go test -race ./...
vet:
	go vet ./...
	cd ../platform-sdk && go vet ./...
proto:
	buf lint
	buf generate
local:
	docker compose -f deploy/compose.yaml up -d --wait postgres kafka
	docker compose -f deploy/compose.yaml run --rm topics
integration:
	@test -n "$(TEST_DATABASE_URL)" || (echo 'Set TEST_DATABASE_URL to the isolated migrated fixture'; exit 1)
	@test -n "$(TEST_KAFKA_BROKERS)" || (echo 'Set TEST_KAFKA_BROKERS'; exit 1)
	go test -race -count=1 ./internal/product/adapter/postgres
	cd ../platform-sdk && go test -race -count=1 ./kafka
run:
	go run ./cmd/api
