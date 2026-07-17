---
title: "Requirements Catalog: приложение A — шаблон требования"
description: "Шаблон нового требования."
keywords:
  - "M8 Platform"
  - "requirements"
---

# Приложение A. Шаблон нового требования {#requirements-appendix-a}

{% note info "Навигация Requirements Catalog" %}

[Оглавление требований](../index.md) | [PADS](../../pads/index.md) | [Engineering artifacts](../../../engineering-artifacts/index.md) | [Предыдущий раздел: 17. Требования к машинному реестру](../governance/17-machine-registry.md) | [Следующий раздел: Приложение B. Проверка целостности редакции](appendix-b-integrity-check.md)

{% endnote %}

```yaml
requirement:
  id: <CTX>-FR-000
  title: ...
  type: functional
  status: PROPOSED
  priority: Must|Should|Could
  owner_context: ...
  owner_service: ...
  capability_ids: [...]
  statement: ...
  preconditions: [...]
  business_rules: [...]
  acceptance_criteria:
    - id: <REQ>-AC-01
      given: [...]
      when: [...]
      then: [...]
  data_ownership:
    reads: [...]
    writes: [...]
  security:
    permission: ...
    risk_evaluation: ...
    audit: ...
  consistency: ...
  contracts:
    api: [...]
    events: [...]
  quality_attributes: [...]
  traceability:
    pads: [...]
    adr: [...]
    prompts: [...]
    tests: [...]
```
