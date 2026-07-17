---
title: "ADR-0001: bounded contexts"
description: "Границы bounded contexts и сервисная декомпозиция."
keywords:
  - "M8 Platform"
  - "ADR"
---

# ADR-0001. Границы bounded contexts и сервисная декомпозиция {#adr-0001}

{% note info "Навигация" %}

[Engineering artifacts](../index.md) | [ADR index](index.md) | [PADS governance](../../architecture/pads/governance/23-architecture-governance.md)

{% endnote %}

| Поле | Значение |
| --- | --- |
| Статус | `Accepted` |
| Дата | `2026-07-10` |
| Владелец | Sergey Gorbachev |
| Основание | PADS §8–10 |

## Контекст

M8 Platform требует единого, проверяемого решения в указанной области. Решение должно сохранять bounded-context ownership, безопасность, наблюдаемость и трассировку.

## Решение

Разделить M8 на Resource Manager, Identity, Authentication, Access, Risk Decision, Provisioning и Audit; Common Operation является общим контрактом, а не централизованным владельцем всех операций.

## Последствия

- Положительные: единообразие реализации, автоматизируемые проверки, снижение скрытых связей.
- Ограничения: команда обязана поддерживать адаптеры, миграции и compatibility tests.
- Риски: технологическая зависимость контролируется ACL, экспортируемыми контрактами и планом замены.

## Проверка

- Контракт и реализация ссылаются на `ADR-0001`.
- CI проверяет применимые архитектурные правила.
- Отклонение требует нового ADR со связью `supersedes` или `amends`.

## Альтернативы

Альтернативы фиксируются при переводе `Proposed` в `Accepted`; отсутствие сравнительного анализа блокирует принятие необратимого решения.
