# M8 Platform Docs

Документация платформы на базе Diplodoc.

Основные разделы документации ведутся вручную в `docs/services` и `docs/architecture`.

{% note info "Архитектура" %}

Нормативная спецификация PADS разбита на логические разделы: [открыть оглавление PADS](architecture/pads/index.md).

{% endnote %}

## Ключевые инженерные разделы

- [Requirements Catalog](architecture/requirements/index.md) — канонический каталог требований M8 Platform.
- [Engineering Artifact Set](engineering-artifacts/index.md) — контракты, traceability, ADR, SPDD, pilot package и governance CI.
- [Executable Architecture Baseline](executable-baseline/index.md) — утвержденная исполнимая базовая линия: approved contracts, DDL, Go scaffold, CI, SPDD, operations и security evidence.

OpenAPI-справочники генерируются отдельно в `docs/_generated/openapi` и подключаются в навигацию как вложенные разделы API.
