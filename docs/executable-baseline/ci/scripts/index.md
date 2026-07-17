---
title: "CI Scripts"
---

# CI Scripts

[Executable Baseline](../../index.md) | [CI](../index.md)

{% note info %}

Scripts validate the executable baseline, proto structure and Go boundaries.

{% endnote %}

## Raw evidence

| Артефакт | Путь | Назначение |
| --- | --- | --- |
| baseline validator | `validate_baseline.py` | YAML/JSON and catalog uniqueness checks |
| proto structure checker | `check_proto_structure.py` | proto syntax and duplicate RPC checks |
| Go boundary checker | `check_go_boundaries.py` | service-internal import guard |
