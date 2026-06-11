# Access

`Access` отвечает за проверку доступа и административные контракты авторизации
в M8 Platform.

Сервис предоставляет публичный AuthZEN-compatible API для Policy Decision Point
и внутренний M8 API для управления action definitions, resource types, role
bindings и симуляцией решений.

## Зона ответственности

- Access decisions.
- Action registry.
- Resource type authorization metadata.
- Role bindings.
- Access simulation.

## Не отвечает за

- Пользователей и identities.
- Иерархию организаций, рабочих пространств и проектов.
- Расчет risk score.
- Хранение audit events.

## API

- REST reference: `Справочник API -> REST`.
- gRPC reference: `Справочник API -> gRPC`.
