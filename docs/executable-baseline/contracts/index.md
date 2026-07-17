---
title: "Contracts Baseline"
---

# Contracts Baseline

[Executable Baseline](../index.md)

{% note info %}

Contracts baseline фиксирует имена RPC, события, ошибки, AsyncAPI и Protobuf layout. Human-readable каталоги требований и контрактов живут в разделе Engineering Artifacts.

{% endnote %}

## Документы

| Документ | Назначение |
| --- | --- |
| [Protobuf Baseline](protobuf-baseline.md) | правила protobuf-first baseline |

## Raw evidence

| Артефакт | Путь | Назначение |
| --- | --- | --- |
| API catalog approved | `api_catalog.approved.yaml` | 156 утвержденных RPC |
| Event catalog approved | `event_catalog.approved.yaml` | 116 event contracts |
| Error catalog approved | `error_catalog.approved.yaml` | 52 canonical errors |
| AsyncAPI | `events/asyncapi.yaml` | event channel baseline |
| Event schemas | `events/schemas/` | 117 JSON Schema files |
| Protobuf sources | `proto/` | 15 proto files |
