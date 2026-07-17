---
title: "M8 SPDD Constitution"
description: "SPDD constitution for M8 Platform."
keywords:
  - "M8 Platform"
  - "SPDD"
---

# M8 SPDD Constitution {#spdd-constitution}

{% note info "Навигация" %}

[Engineering artifacts](../index.md) | [SPDD index](index.md) | [PADS: SPDD](../../architecture/pads/governance/22-spdd.md)

{% endnote %}

_M8-SPDD-CONSTITUTION · Версия 1.0 · 10 июля 2026 года_

| Поле | Значение |
| --- | --- |
| Идентификатор | `M8-SPDD-CONSTITUTION` |
| Версия | `1.0` |
| Статус | Нормативная редакция |
| Владелец | Sergey Gorbachev |
| Нормативная основа | `PADS-000@1.0`, `M8-REQ-000@0.1` |
| Область | Все Structured Prompts и AI-assisted changes M8 Platform |

# 1. Иерархия источников

```text
Accepted ADR → PADS → Requirements Catalog → Contract Catalogs
→ Context Prompt → Feature Prompt → Task Prompt → Source Code/Test
```

Агент MUST сообщить о конфликте и не имеет права молча выбрать нижестоящий источник.

# 2. Обязательный вход Structured Prompt

Каждый implementation prompt MUST содержать: prompt ID, requirement IDs, acceptance criteria, owner context/service, aggregate/invariants, API/Event/Error/Data IDs, allowed/forbidden dependencies, security, consistency, observability, tests, definition of done и expected output.

# 3. Границы изменений

- Один Task Prompt изменяет один bounded context и одну проверяемую цель.
- Изменение публичного контракта отделяется от реализации и проходит contract design gate.
- Запрещены прямой доступ к чужой БД, импорт external SDK types в domain, синхронная публикация события до commit, скрытые архитектурные решения и silent fallback security.
- Новая зависимость, ownership или breaking change требует ADR/impact analysis.

# 4. Реализация

- Domain layer не импортирует transport/storage/provider packages.
- Application layer оркестрирует use case, но не хранит provider semantics.
- Infrastructure реализует ports и ACL.
- Mutation идемпотентна; aggregate и Outbox атомарны.
- Consumer имеет Inbox/deduplication и revision guard.
- Секреты не попадают в logs/traces/events/audit.

# 5. Проверка

Обязательны unit, integration, contract, acceptance, security и architecture tests в объёме требования. Агент MUST перечислить изменённые файлы, выполненные проверки, принятые решения, остаточные риски и несоответствия спецификации.

# 6. Запрет на выдумывание

Неопределённое поле контракта, бизнес-правило, permission, retention или fail mode не придумывается. Оно фиксируется как `OPEN_QUESTION` либо предлагается отдельным проектным решением, не маскируемым под реализацию.

# 7. Definition of Done

Изменение завершено только при выполнении acceptance criteria, прохождении checks, наличии telemetry/audit, обновлённой traceability и отсутствии запрещённых зависимостей.
