---
title: "SDK и CLI Specification"
---

# SDK и CLI Specification

[Executable Baseline](../index.md) | [SDK and CLI Baseline](index.md)

## SDK

- Go и TypeScript генерируются из Protobuf;
- клиенты автоматически передают request/correlation IDs;
- mutation требует idempotency key;
- retry разрешён только для retryable codes;
- pagination реализуется iterator-обёрткой;
- Operation имеет `Wait`, `Cancel` и typed result helpers.

## CLI

Команды: `m8 org`, `m8 workspace`, `m8 project`, `m8 identity`,
`m8 auth`, `m8 access`, `m8 risk`, `m8 resource`, `m8 operation`, `m8 audit`.

CLI не сохраняет long-lived access tokens в открытом виде.
