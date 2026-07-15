---
title: "M8 Platform Engineering Artifact Set"
description: "Комплект инженерных артефактов M8 Platform."
keywords:
  - "M8 Platform"
  - "engineering artifacts"
---

# M8 Platform Engineering Artifact Set {#engineering-artifacts}

{% note info "Навигация" %}

[PADS](../architecture/pads/index.md) | [Requirements Catalog](../architecture/requirements/index.md) | [Traceability Registry](traceability/traceability-registry.md) | [Pilot AUTH-FR-017](pilot/auth-fr-017/index.md)

{% endnote %}

_Версия 0.1 · 10 июля 2026 года_

Комплект продолжает [PADS-000@1.0](../architecture/pads/index.md) и [M8-REQ-000@0.1](../architecture/requirements/index.md) и доводит архитектурную работу до контрактов, трассировки, SPDD и проверяемого пилота.

## Состав

| Каталог | Назначение |
| --- | --- |
| [foundation](foundation/index.md) | Канонические ссылки на PADS v1.0 и Requirements Catalog v0.1 без повторной публикации текста. |
| [installer](installer/index.md) | Архитектура и первый компилируемый baseline M8 Installer 1.0. |
| [contracts](contracts/index.md) | API, Event, Error catalogs и Protobuf package map. |
| [data](data/index.md) | Data Ownership Registry. |
| [traceability](traceability/index.md) | Registry для 214 требований и schema. |
| [adr](adr/index.md) | ADR baseline и 10 отдельных решений. |
| [spdd](spdd/index.md) | Constitution, семь Context Prompts и шаблоны Feature/Task/Review. |
| [pilot/auth-fr-017](pilot/auth-fr-017/index.md) | Полный пилот от требования до release evidence. |
| [governance-ci](governance-ci/index.md) | Fitness functions, schemas, validator и PR/CODEOWNERS templates. |

## Статистика

- Requirements: **214**.
- API contract candidates: **156**.
- Integration event candidates: **116**.
- Canonical errors: **52**.
- Data ownership entries: **38**.
- Traceability records: **214**.
- ADR: **10**.
- Context Prompts: **7**.
- Pilot Task Prompts: **6**.

## Статусы

Catalog contracts имеют статус `proposed`: названия RPC/event и mapping уже назначены, но поля каждого массового контракта должны пройти отдельный design gate. Пилот `AUTH-FR-017` детализирован до уровня, достаточного для архитектурного review и начала реализации после утверждения открытых продуктовых параметров.

## Проверка

```bash
python governance-ci/validate_artifacts.py
```

## Следующая практическая последовательность

1. Утвердить ADR-0003 и ADR-0010 после benchmark/operational design.
2. Провести contract design для Wave 0: Common, Operations, Audit envelope, Error model.
3. Утвердить Pilot Package `AUTH-FR-017` и реализовать шесть Task Prompts.
4. Подключить artifact validator, Buf и fitness functions в CI.
5. По результатам пилота скорректировать schemas/prompts и перевести выбранные требования в `APPROVED/PLANNED`.
6. Выпускать feature packages волнами: Resource Manager → Identity/Access → Authentication/Risk → Provisioning.
