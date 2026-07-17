# Validation Report

**Status:** PASSED

## Artifact validation

```text
OK: requirements=214, rpcs=156, events=116
```

## Protobuf structural validation

```text
OK: proto structural checks
```

## Architecture boundary validation

```text
OK: Go dependency boundaries
```

## Go tests

```text
All packages compiled.
AUTH-FR-017 application tests passed:
- creates transaction, Operation and Outbox event;
- returns the same result for an equivalent idempotent request;
- stops before mutation when Risk Decision denies the operation.
```

## Result

- errors: 0
- duplicate requirement IDs: 0
- duplicate RPC names: 0
- duplicate event types: 0
- invalid YAML/JSON documents: 0
- detected forbidden Go dependencies: 0
