---
title: "PADS: приложение A — структура репозитория"
description: "Начальная структура репозитория M8 Platform."
keywords:
  - "M8 Platform"
  - "PADS"
  - "architecture"
---

# Приложение A. Начальная структура репозитория {#pads-appendix-a}

{% note info "Навигация PADS" %}

[Оглавление PADS](../index.md) | [Предыдущий раздел: 24. Глоссарий](../governance/24-glossary.md) | [Следующий раздел: Приложение B. Минимальное определение готовности](appendix-b-definition-of-done.md)

{% endnote %}

```text
/cmd
  /resourcemanager-api
  /identity-api
  /authentication-api
  /access-api
  /risk-decision-api
  /provisioning-api
  /audit-api

/internal
  /modules
    /resourcemanager
      /domain
      /application
      /adapter
      /infrastructure
    /identity
    /authentication
    /access
    /riskdecision
    /provisioning
    /audit
  /platform
    /config
    /logger
    /metrics
    /tracing
    /module

/api
  /proto
    /m8/platform/common
    /m8/platform/resourcemanager
    /m8/platform/identity
    /m8/platform/authentication
    /m8/platform/access
    /m8/platform/riskdecision
    /m8/platform/provisioning
    /m8/platform/audit

/docs
  /01-domain
  /02-architecture
  /03-requirements
  /04-contracts
  /05-decisions
  /07-spdd
  /08-validation
```
