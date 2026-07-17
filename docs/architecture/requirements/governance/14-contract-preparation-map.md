---
title: "Requirements Catalog: карта подготовки контрактов"
description: "Карта подготовки API, событий, данных и операций."
keywords:
  - "M8 Platform"
  - "requirements"
---

# 14. Карта подготовки контрактов {#requirements-contract-preparation-map}

{% note info "Навигация Requirements Catalog" %}

[Оглавление требований](../index.md) | [PADS](../../pads/index.md) | [Engineering artifacts](../../../engineering-artifacts/index.md) | [Предыдущий раздел: 13. Сквозные сценарии и декомпозиция требований](13-cross-cutting-scenarios.md) | [Следующий раздел: 15. Приоритеты реализации](15-implementation-priorities.md)

{% endnote %}

| Волна | Контракты | Требования-источники | Результат |
| --- | --- | --- | --- |
| 1 | Common, RequestContext, ResourceReference, Error, Operation | `PLT-FR-001..008`, `OPS-*` | Shared Protobuf packages |
| 2 | Resource Manager API/events | `RM-*` | `resource_manager.v1`, `resource_manager.events.v1` |
| 3 | Identity API/events | `ID-*` | `identity.v1`, `identity.events.v1` |
| 4 | Access API/events | `ACC-*` | `access.v1`, `access.events.v1` |
| 5 | Risk Decision API/events | `RISK-*` | `risk.v1`, `risk.events.v1` |
| 6 | Authentication API/events | `AUTH-*` | `authentication.v1`, `authentication.events.v1` |
| 7 | Audit API/event envelope | `AUD-*`, `PLT-FR-004` | `audit.v1`, common audit envelope |
| 8 | Provisioning API/events | `PROV-*` | `provisioning.v1`, `provisioning.events.v1` |
