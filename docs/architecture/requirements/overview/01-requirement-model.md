---
title: "Requirements Catalog: модель требования"
description: "Обязательные поля требования, Definition of Ready и Definition of Done."
keywords:
  - "M8 Platform"
  - "requirements"
---

# 1. Модель требования {#requirements-model}

{% note info "Навигация Requirements Catalog" %}

[Оглавление требований](../index.md) | [PADS](../../pads/index.md) | [Engineering artifacts](../../../engineering-artifacts/index.md) | [Предыдущий раздел: 0. Управление артефактом](00-artifact-control.md) | [Следующий раздел: 2. Сводная карта распределения](02-distribution-map.md)

{% endnote %}

## 1.1. Обязательные поля

```yaml
requirement:
  id: AUTH-FR-017
  title: Повторная аутентификация после невозможности refresh
  type: functional
  status: ANALYZED
  priority: Must
  owner_context: Authentication
  owner_service: m8-authentication
  capability_ids: [CAP-AUTHN-09]
  statement: ...
  acceptance_criteria: [...]
  data_ownership: ...
  security: ...
  consistency: ...
  contracts: ...
  traceability:
    pads: [PADS-000@1.0#20, PADS-000@1.0#22]
```

## 1.2. Definition of Ready

Требование готово к реализации, когда определены owner, capability, проверяемые критерии, security/data/consistency impacts, contract impact, зависимости, failure mode, quality targets и отсутствуют нерешённые архитектурные вопросы.

## 1.3. Definition of Done

Требование завершено, когда выполнены acceptance criteria, готовы контракты и миграции, пройдены архитектурные и security checks, работают telemetry и audit, выполнен Review Prompt, а release evidence связано с ID требования.
