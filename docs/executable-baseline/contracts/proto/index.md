---
title: "Protobuf Sources"
---

# Protobuf Sources

[Executable Baseline](../../index.md) | [Contracts](../index.md)

{% note info %}

Proto files are executable contract sources for API generation and compatibility checks.

{% endnote %}

## Raw evidence

| Артефакт | Путь | Назначение |
| --- | --- | --- |
| buf module | `buf.yaml` | buf module configuration |
| buf generation | `buf.gen.yaml` | generation configuration |
| m8 proto packages | `m8/` | module-owned service contracts |
| google stubs | `google/` | local baseline stubs used by the scaffold |
