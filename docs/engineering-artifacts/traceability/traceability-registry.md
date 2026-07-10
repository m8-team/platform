---
title: "M8 Platform Traceability Registry"
description: "Реестр трассировки требований, контрактов, ADR, prompts and tests."
keywords:
  - "M8 Platform"
  - "traceability"
---

# M8 Platform Traceability Registry {#traceability-registry}

{% note info "Навигация" %}

[Engineering artifacts](../index.md) | [Requirements Catalog](../../architecture/requirements/index.md) | `traceability.yaml` | `traceability.schema.yaml`

{% endnote %}

_M8-TRACE-000 · Версия 0.1 · 10 июля 2026 года_

| Поле | Значение |
| --- | --- |
| Идентификатор | `M8-TRACE-000` |
| Версия | `0.1` |
| Статус | Базовая machine-readable редакция |
| Владелец | Sergey Gorbachev |
| Нормативная основа | `PADS-000@1.0`, `M8-REQ-000@0.1` |
| Область | Capability → Requirement → Contract → Prompt → Test → Release Evidence |

# 1. Назначение

Реестр обеспечивает двунаправленную трассировку. YAML является машинным источником; Markdown — обзором и правилами.

# 2. Покрытие

| Статус | Количество |
| --- | --- |
| architecture-policy | 51 |
| contract-mapped | 162 |
| pilot-complete | 1 |

# 3. Полнота по контекстам

| Контекст | Всего | API/Event mapped | Architecture policy | Pilot complete |
| --- | --- | --- | --- | --- |
| Platform | 23 | 0 | 23 | 0 |
| Resource Manager | 26 | 22 | 4 | 0 |
| Identity | 28 | 24 | 4 | 0 |
| Authentication | 36 | 29 | 6 | 1 |
| Access | 25 | 22 | 3 | 0 |
| Risk Decision | 20 | 17 | 3 | 0 |
| Provisioning | 25 | 22 | 3 | 0 |
| Audit | 19 | 16 | 3 | 0 |
| Common Operation | 12 | 10 | 2 | 0 |

# 4. Правила CI

- Requirement ID должен существовать в Requirements Catalog.
- Ссылка на API/Event/Data/ADR/Prompt/Test должна разрешаться в соответствующем реестре.
- Требование в `IMPLEMENTING` не может иметь пустые Prompt/Test links.
- Требование в `VERIFIED` должно иметь release evidence.
- Contract breaking change обязан ссылаться на impact analysis и migration ADR.

# 5. Пилот AUTH-FR-017

| Requirement | API | Events | Prompt | Tests |
| --- | --- | --- | --- | --- |
| `AUTH-FR-017` | API-AUTH-001 | EVT-AUTH-001 | `SP-FEATURE-AUTH-017` | `AUTH-FR-017-AC-01`, `AUTH-FR-017-AC-02` |
