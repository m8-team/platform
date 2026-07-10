---
title: "Requirements Catalog: сквозные сценарии"
description: "Сквозные сценарии и правила декомпозиции требований."
keywords:
  - "M8 Platform"
  - "requirements"
---

# 13. Сквозные сценарии и декомпозиция требований {#requirements-cross-cutting-scenarios}

{% note info "Навигация Requirements Catalog" %}

[Оглавление требований](../index.md) | [PADS](../../pads/index.md) | [Engineering artifacts](../../../engineering-artifacts/index.md) | [Предыдущий раздел: 12. Архитектурное управление и SPDD](12-architecture-governance-spdd.md) | [Следующий раздел: 14. Карта подготовки контрактов](14-contract-preparation-map.md)

{% endnote %}

## 13.1. Создание Project

```text
RM-FR-020 Create Project
  → PLT-FR-005 Access check
  → PLT-FR-006 Risk evaluation при чувствительной политике
  → PLT-ARC-002 aggregate + Outbox
  → RM-FR-050 ProjectCreated
  → AUD-FR-001 Audit ingest
  → consumers update local projections
```

## 13.2. Step-up для привилегированной операции

```text
ACC-FR-001 confirms permission
  → RISK-FR-002 returns CHALLENGE(required AAL)
  → AUTH-FR-020 performs step-up
  → RISK-FR-002 re-evaluates bound action
  → owner context commits action
  → PLT-FR-004 writes audit evidence
```

## 13.3. Удаление Project

```text
RM-FR-022 owns process and Operation
  → PROV-FR-012 deprovisions managed resources
  → ACC-FR-015/021 revokes bindings and relationships
  → ID-FR-050 applies scoped privacy policy where required
  → AUD-FR-030 preserves or removes evidence by retention/legal hold
  → Resource Manager finalizes tombstone
```
