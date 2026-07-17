---
title: "Requirements Catalog: SPDD backlog"
description: "Backlog Structured Prompt Driven Development."
keywords:
  - "M8 Platform"
  - "requirements"
---

# 16. SPDD backlog {#requirements-spdd-backlog}

{% note info "Навигация Requirements Catalog" %}

[Оглавление требований](../index.md) | [PADS](../../pads/index.md) | [Engineering artifacts](../../../engineering-artifacts/index.md) | [Предыдущий раздел: 15. Приоритеты реализации](15-implementation-priorities.md) | [Следующий раздел: 17. Требования к машинному реестру](17-machine-registry.md)

{% endnote %}

Для каждого `APPROVED` требования создаётся цепочка:

```text
Requirement
→ Feature Prompt
→ Design Prompt (при изменении контракта/инварианта)
→ Task Prompts
→ Review Prompt
→ Implementation Manifest
```

Первый рекомендуемый пилот:

| Артефакт | ID |
| --- | --- |
| Requirement | `AUTH-FR-017` |
| API contract | `API-AUTH-START-AUTH-V1` |
| Integration event | `EVT-AUTH-STARTED-V1` |
| Feature Prompt | `SPF-AUTH-017-V1` |
| Domain Task Prompt | `SPT-AUTH-017-DOMAIN-V1` |
| Persistence Task Prompt | `SPT-AUTH-017-PERSISTENCE-V1` |
| Review Prompt | `SPR-AUTH-017-V1` |
