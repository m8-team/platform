---
title: "M8 Go Repository Baseline"
---

# M8 Go Repository Baseline

[Executable Baseline](../index.md) | [Repository](index.md)

Monorepo компилируется без внешних зависимостей. Каждый сервис имеет `cmd`,
локальный `internal/domain` и собственную границу. Пакет Authentication содержит
реализованный application use case `AUTH-FR-017` с идемпотентностью,
Risk Decision, Operation и Outbox.

```bash
go test ./...
```

Остальные сервисы представлены компилируемым operational skeleton и расширяются
только через утверждённые Feature/Task Prompts.
