---
title: "Artifact validator"
description: "Artifact validator script."
keywords:
  - "M8 Platform"
  - "governance CI"
---

# Artifact validator {#artifact-validator}

{% note info "Навигация" %}

[Engineering artifacts](../index.md) | [Governance CI](index.md) | [validation report](validation-report.md) | `validate_artifacts.py`

{% endnote %}

Валидатор проверяет YAML, уникальность ID и разрешимость ссылок API/Event/Data из traceability registry. Для разбора YAML он использует `docs/node_modules/js-yaml`, поэтому перед запуском должен быть выполнен `npm install` в `docs/`.

Команда запуска: python governance-ci/validate_artifacts.py
