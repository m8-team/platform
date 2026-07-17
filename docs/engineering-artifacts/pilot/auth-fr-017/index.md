---
title: "Pilot Feature Package: AUTH-FR-017"
description: "Pilot package index."
keywords:
  - "M8 Platform"
  - "AUTH-FR-017"
---

# Pilot Feature Package: AUTH-FR-017 {#pilot-auth-fr-017}

{% note info "Навигация" %}

[Engineering artifacts](../../index.md) | [Pilot index](index.md) | [Requirements: AUTH-FR-017](../../../architecture/requirements/contexts/06-authentication.md#auth-fr-017) | [Traceability](../../traceability/traceability-registry.md) | [SPDD](../../spdd/index.md)

{% endnote %}

_M8-PILOT-AUTH-017 · Версия 0.1 · 10 июля 2026 года_

| Поле | Значение |
| --- | --- |
| Идентификатор | `M8-PILOT-AUTH-017` |
| Версия | `0.1` |
| Статус | Готов к архитектурному и продуктовому review |
| Владелец | Sergey Gorbachev |
| Нормативная основа | `PADS-000@1.0`, `M8-REQ-000@0.1` |
| Область | Повторная CIBA-аутентификация после невозможности refresh |

# Состав пакета

1. Уточнённое требование и invariants.
2. Use Case и sequence.
3. API contract `StartAuthentication`.
4. Integration events.
5. Error mapping.
6. Feature Prompt.
7. Шесть Task Prompts.
8. Review Prompt.
9. Acceptance test specification.
10. Implementation Manifest и Release Evidence template.

# Главный принцип

M8 Authentication не принимает и не анализирует raw refresh token. Клиент/authorization component сообщает проверенный reason `REFRESH_UNAVAILABLE`; создаётся новая независимая AuthenticationTransaction с CIBA/step-up policy, новым risk decision и новым lifecycle.
