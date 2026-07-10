---
title: "PADS: приложение C — план последующих артефактов"
description: "Следующие обязательные артефакты и рекомендуемый пилот."
keywords:
  - "M8 Platform"
  - "PADS"
  - "architecture"
---

# Приложение C. План последующих артефактов {#pads-appendix-c}

{% note info "Навигация PADS" %}

[Оглавление PADS](../index.md) | [Предыдущий раздел: Приложение B. Минимальное определение готовности](appendix-b-definition-of-done.md)

{% endnote %}

Настоящее приложение не заменяет главы 20–23, а задаёт рекомендуемый порядок практического применения PADS.

## C.1. Обязательные следующие артефакты

1. **Requirements Catalog** — полный реестр `CAP-*`, `PLT-*`, `RM-*`, `ID-*`, `AUTH-*`, `ACC-*`, `RISK-*`, `PROV-*`, `AUD-*`, `OPS-*` в YAML/Markdown.
2. **Traceability Registry** — machine-readable граф связей capabilities, requirements, invariants, contracts, prompts и tests.
3. **ADR Baseline** — начальный набор решений о YDB, YDB Topics/Kafka, Temporal, Keycloak CIBA, SpiceDB, Common Operation, API transport и multi-region assumptions.
4. **API Catalog** — Protobuf packages, methods, permissions, errors, idempotency и LRO semantics.
5. **Event Catalog** — event types, schemas, partition keys, retention, producers, consumers и rebuild paths.
6. **Data Ownership Registry** — authoritative entities/attributes, classification, retention, projections и deletion obligations.
7. **Context Prompts** — семь SPDD Context Prompts и общий Constitution Prompt.
8. **Pilot Feature Package** — одно требование, полностью проведённое от requirement до code/test/release evidence.

## C.2. Рекомендуемый пилот

В качестве пилота рекомендуется `AUTH-FR-017`:

```text
Refresh token cannot be used
→ start a new CIBA AuthenticationTransaction
→ evaluate Risk
→ create Operation + Outbox atomically
→ publish AuthenticationStarted
→ verify idempotency, security, audit and observability
```

Пилот затрагивает достаточно механизмов для проверки PADS/SPDD, но остаётся ограниченным контекстом Authentication.

## C.3. Критерий перехода к массовой реализации

Массовая декомпозиция сервисов начинается после того, как пилот доказал:

- пригодность requirement schema;
- полноту Context/Feature/Task/Review Prompt;
- автоматическую traceability validation;
- корректность API/Event registries;
- архитектурные fitness functions;
- идемпотентность и Outbox/Inbox patterns;
- security/observability gates;
- возможность независимого review generated changes.

---
