---
title: "Requirements Catalog: приоритеты реализации"
description: "Приоритетные волны реализации."
keywords:
  - "M8 Platform"
  - "requirements"
---

# 15. Приоритеты реализации {#requirements-implementation-priorities}

{% note info "Навигация Requirements Catalog" %}

[Оглавление требований](../index.md) | [PADS](../../pads/index.md) | [Engineering artifacts](../../../engineering-artifacts/index.md) | [Предыдущий раздел: 14. Карта подготовки контрактов](14-contract-preparation-map.md) | [Следующий раздел: 16. SPDD backlog](16-spdd-backlog.md)

{% endnote %}

## 15.1. Foundation

- платформенные архитектурные, security и observability requirements;
- Common contracts, Error Model и Operation;
- Audit ingest;
- базовый Resource Manager;
- базовый Access Check.

## 15.2. Identity and Authentication Core

- User Pool, User, External Identity и ResolveSubject;
- Client и AuthenticationTransaction;
- CIBA, OTP, provider callback, handoff;
- Risk Evaluate для authentication;
- session reference и revocation.

## 15.3. Control Plane Expansion

- Roles, RoleBindings, simulation и reviews;
- Provisioning definitions, drivers, desired/observed state и reconciliation;
- Project deletion and cross-context workflows;
- расширенный Audit search/export/integrity.
