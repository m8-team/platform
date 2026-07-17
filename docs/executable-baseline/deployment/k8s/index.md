---
title: "Kubernetes Base"
---

# Kubernetes Base

[Executable Baseline](../../index.md) | [Deployment](../index.md)

{% note info %}

Kubernetes manifests are raw deployment evidence for the executable baseline.

{% endnote %}

## Raw evidence

| Артефакт | Путь | Назначение |
| --- | --- | --- |
| namespace baseline | `base/namespaces.yaml` | namespaces |
| default deny | `base/default-deny.yaml` | network policy |
| service workloads | `base/` | 9 YAML manifests |
