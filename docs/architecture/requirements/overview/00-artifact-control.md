---
title: "Requirements Catalog: управление артефактом"
description: "Назначение, нормативные правила, состояния и статистика редакции."
keywords:
  - "M8 Platform"
  - "requirements"
---

# 0. Управление артефактом {#requirements-artifact-control}

{% note info "Навигация Requirements Catalog" %}

[Оглавление требований](../index.md) | [PADS](../../pads/index.md) | [Engineering artifacts](../../../engineering-artifacts/index.md) | [Следующий раздел: 1. Модель требования](01-requirement-model.md)

{% endnote %}

## 0.1. Назначение

Каталог является нормативным мостом между архитектурной спецификацией и реализацией. Он распределяет требования по владельцам, фиксирует проверяемое поведение и создаёт устойчивые идентификаторы для API, событий, ADR, Structured Prompts, кода, тестов и release evidence.

```text
PADS / Business Capability
        ↓
Requirement
        ↓
Acceptance Criteria
        ↓
API / Event / Data / Workflow design
        ↓
Structured Prompt
        ↓
Code / Tests / Release Evidence
```

## 0.2. Нормативные правила

- ID требования после публикации не переиспользуется.
- У требования ровно один owner context.
- Supporting dependency не становится совладельцем требования.
- `Must` означает обязательность для целевой архитектуры, но не обязательно для первого релиза.
- Перед реализацией требование должно получить статус `APPROVED` и release target.
- Любое изменение владельца, инварианта или публичного обязательства требует impact analysis и при необходимости ADR.

## 0.3. Состояния

```text
PROPOSED → ANALYZED → APPROVED → PLANNED → IMPLEMENTING
→ VERIFIED → RELEASED → DEPRECATED → RETIRED
```

Дополнительные терминальные состояния: `REJECTED`, `SUPERSEDED`.

## 0.4. Статистика редакции

- Всего требований: **214**.
- Функциональных: **171**.
- Требований данных: **11**.
- Требований безопасности: **12**.
- Нефункциональных: **12**.
- Архитектурных/управленческих: **8**.
