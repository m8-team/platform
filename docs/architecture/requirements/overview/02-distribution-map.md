---
title: "Requirements Catalog: сводная карта распределения"
description: "Распределение требований по владельцам, типам и сервисам."
keywords:
  - "M8 Platform"
  - "requirements"
---

# 2. Сводная карта распределения {#requirements-distribution-map}

{% note info "Навигация Requirements Catalog" %}

[Оглавление требований](../index.md) | [PADS](../../pads/index.md) | [Engineering artifacts](../../../engineering-artifacts/index.md) | [Предыдущий раздел: 1. Модель требования](01-requirement-model.md) | [Следующий раздел: 3. Платформенные и сквозные требования](../contexts/03-platform-cross-cutting.md)

{% endnote %}

| Владелец | Количество | Основные capability | Сервис |
| --- | ---: | --- | --- |
| Платформенные и сквозные требования | 20 | `CAP-AUD-01`, `CAP-AUD-02`, `CAP-AUTHZ-06`, `CAP-GOV-05`, `CAP-GOV-07`, `CAP-ID-03`… | несколько владельцев (см. детальные требования) |
| Resource Manager | 26 | `CAP-RM-01`, `CAP-RM-02`, `CAP-RM-03`, `CAP-RM-04`, `CAP-RM-05`, `CAP-RM-06`… | `m8-resource-manager` |
| Identity | 28 | `CAP-ID-01`, `CAP-ID-02`, `CAP-ID-03`, `CAP-ID-04`, `CAP-ID-05`, `CAP-ID-06`… | `m8-identity` |
| Authentication | 36 | `CAP-AUTHN-01`, `CAP-AUTHN-02`, `CAP-AUTHN-03`, `CAP-AUTHN-04`, `CAP-AUTHN-05`, `CAP-AUTHN-06`… | `m8-authentication` |
| Access | 25 | `CAP-AUTHZ-01`, `CAP-AUTHZ-02`, `CAP-AUTHZ-03`, `CAP-AUTHZ-04`, `CAP-AUTHZ-05`, `CAP-AUTHZ-06`… | `m8-access` |
| Risk Decision | 20 | `CAP-RISK-01`, `CAP-RISK-02`, `CAP-RISK-03`, `CAP-RISK-04`, `CAP-RISK-05`, `CAP-RISK-06`… | `m8-risk-decision` |
| Provisioning | 25 | `CAP-PROV-01`, `CAP-PROV-02`, `CAP-PROV-03`, `CAP-PROV-04`, `CAP-PROV-05`, `CAP-PROV-07`… | `m8-provisioning` |
| Audit | 19 | `CAP-AUD-01`, `CAP-AUD-02`, `CAP-AUD-03`, `CAP-AUD-04`, `CAP-AUD-05`, `CAP-AUD-06`… | `m8-audit` |
| Common Operation | 12 | `CAP-OPS-01`, `CAP-OPS-02`, `CAP-OPS-03`, `CAP-OPS-04`, `CAP-OPS-05`, `CAP-OPS-06`… | `operation owner service` |
| Архитектурное управление и SPDD | 3 | `CAP-GOV-04`, `CAP-GOV-05`, `CAP-GOV-09` | `SPDD tooling`, `contract owners`, `repository/CI` |

## 2.1. Покрытие по типам

| Контекст | FR | DATA | SEC | NFR | ARC/GOV |
| --- | ---: | ---: | ---: | ---: | ---: |
| Платформенные и сквозные требования | 8 | 0 | 4 | 3 | 5 |
| Resource Manager | 22 | 2 | 1 | 1 | 0 |
| Identity | 24 | 2 | 1 | 1 | 0 |
| Authentication | 30 | 2 | 2 | 2 | 0 |
| Access | 22 | 1 | 1 | 1 | 0 |
| Risk Decision | 17 | 1 | 1 | 1 | 0 |
| Provisioning | 22 | 1 | 1 | 1 | 0 |
| Audit | 16 | 1 | 1 | 1 | 0 |
| Common Operation | 10 | 1 | 0 | 1 | 0 |
| Архитектурное управление и SPDD | 0 | 0 | 0 | 0 | 3 |
