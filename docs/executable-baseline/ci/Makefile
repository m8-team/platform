.PHONY: validate test proto-check schema-check architecture-check

validate: schema-check proto-check architecture-check test

test:
	cd ../repository && go test ./...

schema-check:
	python3 scripts/validate_baseline.py ..

proto-check:
	python3 scripts/check_proto_structure.py ../contracts/proto

architecture-check:
	python3 scripts/check_go_boundaries.py ../repository
