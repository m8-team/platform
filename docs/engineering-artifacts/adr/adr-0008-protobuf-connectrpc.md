---
title: "ADR-0008: Protobuf и ConnectRPC"
description: "Protobuf и ConnectRPC как основной API transport."
keywords:
  - "M8 Platform"
  - "ADR"
---

# ADR-0008. Protobuf и ConnectRPC как основной API transport {#adr-0008}

{% note info "Навигация" %}

[Engineering artifacts](../index.md) | [ADR index](index.md) | [PADS governance](../../architecture/pads/governance/23-architecture-governance.md)

{% endnote %}

| Поле | Значение |
| --- | --- |
| Статус | `Accepted` |
| Дата | `2026-07-10` |
| Владелец | Sergey Gorbachev |
| Основание | PADS §12 |

## Контекст

M8 Platform требует единого, проверяемого решения в указанной области. Решение должно сохранять bounded-context ownership, безопасность, наблюдаемость и трассировку.

## Решение

Контракты описываются Protobuf, проверяются Buf/Protovalidate и доступны через ConnectRPC; HTTP/JSON является представлением того же контракта.

## Последствия

- Положительные: единообразие реализации, автоматизируемые проверки, снижение скрытых связей.
- Ограничения: команда обязана поддерживать адаптеры, миграции и compatibility tests.
- Риски: технологическая зависимость контролируется ACL, экспортируемыми контрактами и планом замены.

## Проверка

- Контракт и реализация ссылаются на `ADR-0008`.
- CI проверяет применимые архитектурные правила.
- Отклонение требует нового ADR со связью `supersedes` или `amends`.

## Альтернативы

Альтернативы фиксируются при переводе `Proposed` в `Accepted`; отсутствие сравнительного анализа блокирует принятие необратимого решения.
