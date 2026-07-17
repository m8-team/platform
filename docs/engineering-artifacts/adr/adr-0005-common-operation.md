---
title: "ADR-0005: Common Operation"
description: "Common Operation на основе google.longrunning."
keywords:
  - "M8 Platform"
  - "ADR"
---

# ADR-0005. Common Operation на основе google.longrunning {#adr-0005}

{% note info "Навигация" %}

[Engineering artifacts](../index.md) | [ADR index](index.md) | [PADS governance](../../architecture/pads/governance/23-architecture-governance.md)

{% endnote %}

| Поле | Значение |
| --- | --- |
| Статус | `Accepted` |
| Дата | `2026-07-10` |
| Владелец | Sergey Gorbachev |
| Основание | PADS §10, §16 |

## Контекст

M8 Platform требует единого, проверяемого решения в указанной области. Решение должно сохранять bounded-context ownership, безопасность, наблюдаемость и трассировку.

## Решение

Возвращать google.longrunning.Operation для длительных команд; Operation хранится и авторизуется сервисом-владельцем предметной операции.

## Последствия

- Положительные: единообразие реализации, автоматизируемые проверки, снижение скрытых связей.
- Ограничения: команда обязана поддерживать адаптеры, миграции и compatibility tests.
- Риски: технологическая зависимость контролируется ACL, экспортируемыми контрактами и планом замены.

## Проверка

- Контракт и реализация ссылаются на `ADR-0005`.
- CI проверяет применимые архитектурные правила.
- Отклонение требует нового ADR со связью `supersedes` или `amends`.

## Альтернативы

Альтернативы фиксируются при переводе `Proposed` в `Accepted`; отсутствие сравнительного анализа блокирует принятие необратимого решения.
