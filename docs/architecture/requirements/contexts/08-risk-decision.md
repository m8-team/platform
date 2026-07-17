---
title: "Requirements Catalog: Risk Decision"
description: "Требования Risk Decision."
keywords:
  - "M8 Platform"
  - "requirements"
---

# 8. Risk Decision {#requirements-risk-decision}

{% note info "Навигация Requirements Catalog" %}

[Оглавление требований](../index.md) | [PADS](../../pads/index.md) | [Engineering artifacts](../../../engineering-artifacts/index.md) | [Предыдущий раздел: 7. Access](07-access.md) | [Следующий раздел: 9. Provisioning](09-provisioning.md)

{% endnote %}

Владелец раздела: **Risk Decision**. Требований: **20**.

## Реестр

| ID | Тип | Приоритет | Capability | Название | Статус |
| --- | --- | --- | --- | --- | --- |
| `RISK-FR-001` | functional | Must | `CAP-RISK-01` | Оценить риск Authentication | `ANALYZED` |
| `RISK-FR-002` | functional | Must | `CAP-RISK-03` | Оценить риск привилегированного действия | `ANALYZED` |
| `RISK-FR-003` | functional | Must | `CAP-RISK-04` | Вернуть ALLOW/DENY/CHALLENGE/REVIEW | `ANALYZED` |
| `RISK-FR-004` | functional | Must | `CAP-RISK-04` | Получить Risk Assessment | `ANALYZED` |
| `RISK-FR-005` | functional | Must | `CAP-RISK-04` | Истечь Risk Decision | `ANALYZED` |
| `RISK-FR-010` | functional | Must | `CAP-RISK-05` | Создать Risk Policy | `ANALYZED` |
| `RISK-FR-011` | functional | Must | `CAP-RISK-05` | Опубликовать Risk Policy | `ANALYZED` |
| `RISK-FR-012` | functional | Must | `CAP-RISK-05` | Откатить Risk Policy | `ANALYZED` |
| `RISK-FR-013` | functional | Must | `CAP-RISK-09` | Симулировать Risk Policy | `ANALYZED` |
| `RISK-FR-014` | functional | Must | `CAP-RISK-04` | Объяснить Risk Decision | `ANALYZED` |
| `RISK-FR-020` | functional | Must | `CAP-RISK-06` | Принять device signals | `ANALYZED` |
| `RISK-FR-021` | functional | Must | `CAP-RISK-07` | Проверить velocity | `ANALYZED` |
| `RISK-FR-022` | functional | Must | `CAP-RISK-02` | Принять external risk signal | `ANALYZED` |
| `RISK-FR-030` | functional | Must | `CAP-RISK-10` | Создать Manual Review | `ANALYZED` |
| `RISK-FR-031` | functional | Must | `CAP-RISK-10` | Завершить Manual Review | `ANALYZED` |
| `RISK-FR-040` | functional | Must | `CAP-RISK-11` | Сформировать feedback | `ANALYZED` |
| `RISK-FR-050` | functional | Must | `CAP-RISK-12` | Публиковать факты Risk | `ANALYZED` |
| `RISK-DATA-001` | data | Must | `CAP-RISK-04` | Неизменяемость Risk Assessment | `ANALYZED` |
| `RISK-SEC-001` | security | Must | `CAP-RISK-05` | Защита моделей и reason disclosure | `ANALYZED` |
| `RISK-NFR-001` | non-functional | Must | `CAP-RISK-01` | Задержка онлайн-решения | `ANALYZED` |

## Детальные требования

### RISK-FR-001. Оценить риск Authentication {#risk-fr-001}

| Поле | Значение |
| --- | --- |
| Тип | `functional` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Risk Decision` / `m8-risk-decision` |
| Business capability | `CAP-RISK-01` |
| Согласованность | C0 — локальная строгая согласованность |
| Зависимости | Identity projections, Resource Manager projections, Audit |
| Данные | RiskAssessment, RiskPolicy, RiskSignal |
| Безопасность | sensitive decision access |
| API | — |
| События | — |
| Атрибуты качества | — |
| Основание PADS | §6.12, §7, §9.6, §15 |

**Требование.**

Оценить контекст запуска или продолжения Authentication и вернуть нормативное решение.

**Критерии приёмки.**

- `RISK-FR-001-AC-01` — Результат содержит decision, score/band, policy version и reasons.
- `RISK-FR-001-AC-02` — Отсутствующие обязательные сигналы обрабатываются по policy, а не нулевым значением.

**Трассировка для следующего этапа:**

```yaml
requirement_id: RISK-FR-001
capability: CAP-RISK-01
owner_context: Risk Decision
contracts:
  api: []
  events: []
spdd:
  feature_prompt: null
  task_prompts: []
verification:
  tests: []
  release_evidence: []
```

### RISK-FR-002. Оценить риск привилегированного действия {#risk-fr-002}

| Поле | Значение |
| --- | --- |
| Тип | `functional` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Risk Decision` / `m8-risk-decision` |
| Business capability | `CAP-RISK-03` |
| Согласованность | C0 — локальная строгая согласованность |
| Зависимости | Identity projections, Resource Manager projections, Audit |
| Данные | RiskAssessment, RiskPolicy, RiskSignal |
| Безопасность | sensitive decision access |
| API | — |
| События | — |
| Атрибуты качества | — |
| Основание PADS | §6.12, §7, §9.6, §15 |

**Требование.**

Оценить чувствительное административное или ресурсное действие до commit.

**Критерии приёмки.**

- `RISK-FR-002-AC-01` — Decision привязан к action, actor и target.
- `RISK-FR-002-AC-02` — Истёкшее решение нельзя повторно использовать.

**Трассировка для следующего этапа:**

```yaml
requirement_id: RISK-FR-002
capability: CAP-RISK-03
owner_context: Risk Decision
contracts:
  api: []
  events: []
spdd:
  feature_prompt: null
  task_prompts: []
verification:
  tests: []
  release_evidence: []
```

### RISK-FR-003. Вернуть ALLOW/DENY/CHALLENGE/REVIEW {#risk-fr-003}

| Поле | Значение |
| --- | --- |
| Тип | `functional` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Risk Decision` / `m8-risk-decision` |
| Business capability | `CAP-RISK-04` |
| Согласованность | C0 — локальная строгая согласованность |
| Зависимости | Identity projections, Resource Manager projections, Audit |
| Данные | RiskAssessment, RiskPolicy, RiskSignal |
| Безопасность | sensitive decision access |
| API | — |
| События | — |
| Атрибуты качества | — |
| Основание PADS | §6.12, §7, §9.6, §15 |

**Требование.**

Возвращать один из канонических outcomes с требуемым действием.

**Критерии приёмки.**

- `RISK-FR-003-AC-01` — Outcome имеет machine-readable reason codes.
- `RISK-FR-003-AC-02` — CHALLENGE содержит requested assurance/method constraints.

**Трассировка для следующего этапа:**

```yaml
requirement_id: RISK-FR-003
capability: CAP-RISK-04
owner_context: Risk Decision
contracts:
  api: []
  events: []
spdd:
  feature_prompt: null
  task_prompts: []
verification:
  tests: []
  release_evidence: []
```

### RISK-FR-004. Получить Risk Assessment {#risk-fr-004}

| Поле | Значение |
| --- | --- |
| Тип | `functional` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Risk Decision` / `m8-risk-decision` |
| Business capability | `CAP-RISK-04` |
| Согласованность | C0 — локальная строгая согласованность |
| Зависимости | Identity projections, Resource Manager projections, Audit |
| Данные | RiskAssessment, RiskPolicy, RiskSignal |
| Безопасность | sensitive decision access |
| API | — |
| События | — |
| Атрибуты качества | — |
| Основание PADS | §6.12, §7, §9.6, §15 |

**Требование.**

Получить сохранённое решение по ID в разрешённом scope.

**Критерии приёмки.**

- `RISK-FR-004-AC-01` — Ответ не раскрывает чувствительные detection rules вызывающему без permission.
- `RISK-FR-004-AC-02` — Decision immutable после фиксации; correction создаёт новое assessment.

**Трассировка для следующего этапа:**

```yaml
requirement_id: RISK-FR-004
capability: CAP-RISK-04
owner_context: Risk Decision
contracts:
  api: []
  events: []
spdd:
  feature_prompt: null
  task_prompts: []
verification:
  tests: []
  release_evidence: []
```

### RISK-FR-005. Истечь Risk Decision {#risk-fr-005}

| Поле | Значение |
| --- | --- |
| Тип | `functional` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Risk Decision` / `m8-risk-decision` |
| Business capability | `CAP-RISK-04` |
| Согласованность | C0 — локальная строгая согласованность |
| Зависимости | Identity projections, Resource Manager projections, Audit |
| Данные | RiskAssessment, RiskPolicy, RiskSignal |
| Безопасность | sensitive decision access |
| API | — |
| События | — |
| Атрибуты качества | — |
| Основание PADS | §6.12, §7, §9.6, §15 |

**Требование.**

Определять срок применимости решения и отклонять использование после expiry.

**Критерии приёмки.**

- `RISK-FR-005-AC-01` — TTL зависит от action/policy.
- `RISK-FR-005-AC-02` — Повторная оценка получает новый assessment_id.

**Трассировка для следующего этапа:**

```yaml
requirement_id: RISK-FR-005
capability: CAP-RISK-04
owner_context: Risk Decision
contracts:
  api: []
  events: []
spdd:
  feature_prompt: null
  task_prompts: []
verification:
  tests: []
  release_evidence: []
```

### RISK-FR-010. Создать Risk Policy {#risk-fr-010}

| Поле | Значение |
| --- | --- |
| Тип | `functional` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Risk Decision` / `m8-risk-decision` |
| Business capability | `CAP-RISK-05` |
| Согласованность | C0 — локальная строгая согласованность |
| Зависимости | Identity projections, Resource Manager projections, Audit |
| Данные | RiskAssessment, RiskPolicy, RiskSignal |
| Безопасность | sensitive decision access |
| API | — |
| События | — |
| Атрибуты качества | — |
| Основание PADS | §6.12, §7, §9.6, §15 |

**Требование.**

Создать черновую versioned policy с rules, thresholds и required signals.

**Критерии приёмки.**

- `RISK-FR-010-AC-01` — Draft не влияет на production decisions.
- `RISK-FR-010-AC-02` — Policy проходит schema и semantic validation.

**Трассировка для следующего этапа:**

```yaml
requirement_id: RISK-FR-010
capability: CAP-RISK-05
owner_context: Risk Decision
contracts:
  api: []
  events: []
spdd:
  feature_prompt: null
  task_prompts: []
verification:
  tests: []
  release_evidence: []
```

### RISK-FR-011. Опубликовать Risk Policy {#risk-fr-011}

| Поле | Значение |
| --- | --- |
| Тип | `functional` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Risk Decision` / `m8-risk-decision` |
| Business capability | `CAP-RISK-05` |
| Согласованность | C0 — локальная строгая согласованность |
| Зависимости | Identity projections, Resource Manager projections, Audit |
| Данные | RiskAssessment, RiskPolicy, RiskSignal |
| Безопасность | sensitive decision access |
| API | — |
| События | — |
| Атрибуты качества | — |
| Основание PADS | §6.12, §7, §9.6, §15 |

**Требование.**

Опубликовать одобренную policy version для заданного scope.

**Критерии приёмки.**

- `RISK-FR-011-AC-01` — Publish требует permission и audit.
- `RISK-FR-011-AC-02` — Активная версия выбирается детерминированно.

**Трассировка для следующего этапа:**

```yaml
requirement_id: RISK-FR-011
capability: CAP-RISK-05
owner_context: Risk Decision
contracts:
  api: []
  events: []
spdd:
  feature_prompt: null
  task_prompts: []
verification:
  tests: []
  release_evidence: []
```

### RISK-FR-012. Откатить Risk Policy {#risk-fr-012}

| Поле | Значение |
| --- | --- |
| Тип | `functional` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Risk Decision` / `m8-risk-decision` |
| Business capability | `CAP-RISK-05` |
| Согласованность | C0 — локальная строгая согласованность |
| Зависимости | Identity projections, Resource Manager projections, Audit |
| Данные | RiskAssessment, RiskPolicy, RiskSignal |
| Безопасность | sensitive decision access |
| API | — |
| События | — |
| Атрибуты качества | — |
| Основание PADS | §6.12, §7, §9.6, §15 |

**Требование.**

Вернуться к ранее опубликованной совместимой policy version.

**Критерии приёмки.**

- `RISK-FR-012-AC-01` — Откат не изменяет исторические assessments.
- `RISK-FR-012-AC-02` — Причина и actor фиксируются.

**Трассировка для следующего этапа:**

```yaml
requirement_id: RISK-FR-012
capability: CAP-RISK-05
owner_context: Risk Decision
contracts:
  api: []
  events: []
spdd:
  feature_prompt: null
  task_prompts: []
verification:
  tests: []
  release_evidence: []
```

### RISK-FR-013. Симулировать Risk Policy {#risk-fr-013}

| Поле | Значение |
| --- | --- |
| Тип | `functional` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Risk Decision` / `m8-risk-decision` |
| Business capability | `CAP-RISK-09` |
| Согласованность | C0 — локальная строгая согласованность |
| Зависимости | Identity projections, Resource Manager projections, Audit |
| Данные | RiskAssessment, RiskPolicy, RiskSignal |
| Безопасность | sensitive decision access |
| API | — |
| События | — |
| Атрибуты качества | — |
| Основание PADS | §6.12, §7, §9.6, §15 |

**Требование.**

Запустить draft policy на versioned dataset без влияния на production.

**Критерии приёмки.**

- `RISK-FR-013-AC-01` — Результат сравнивается с baseline.
- `RISK-FR-013-AC-02` — Sensitive dataset access контролируется.

**Трассировка для следующего этапа:**

```yaml
requirement_id: RISK-FR-013
capability: CAP-RISK-09
owner_context: Risk Decision
contracts:
  api: []
  events: []
spdd:
  feature_prompt: null
  task_prompts: []
verification:
  tests: []
  release_evidence: []
```

### RISK-FR-014. Объяснить Risk Decision {#risk-fr-014}

| Поле | Значение |
| --- | --- |
| Тип | `functional` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Risk Decision` / `m8-risk-decision` |
| Business capability | `CAP-RISK-04` |
| Согласованность | C0 — локальная строгая согласованность |
| Зависимости | Identity projections, Resource Manager projections, Audit |
| Данные | RiskAssessment, RiskPolicy, RiskSignal |
| Безопасность | sensitive decision access |
| API | — |
| События | — |
| Атрибуты качества | — |
| Основание PADS | §6.12, §7, §9.6, §15 |

**Требование.**

Вернуть безопасное объяснение outcome и reason codes.

**Критерии приёмки.**

- `RISK-FR-014-AC-01` — Внутренние антифрод-правила могут быть скрыты по disclosure policy.
- `RISK-FR-014-AC-02` — Объяснение воспроизводимо для policy version и input snapshot.

**Трассировка для следующего этапа:**

```yaml
requirement_id: RISK-FR-014
capability: CAP-RISK-04
owner_context: Risk Decision
contracts:
  api: []
  events: []
spdd:
  feature_prompt: null
  task_prompts: []
verification:
  tests: []
  release_evidence: []
```

### RISK-FR-020. Принять device signals {#risk-fr-020}

| Поле | Значение |
| --- | --- |
| Тип | `functional` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Risk Decision` / `m8-risk-decision` |
| Business capability | `CAP-RISK-06` |
| Согласованность | C0 — локальная строгая согласованность |
| Зависимости | Identity projections, Resource Manager projections, Audit |
| Данные | RiskAssessment, RiskPolicy, RiskSignal |
| Безопасность | sensitive decision access |
| API | — |
| События | — |
| Атрибуты качества | — |
| Основание PADS | §6.12, §7, §9.6, §15 |

**Требование.**

Принимать нормализованные сигналы устройства с provenance, freshness и confidence.

**Критерии приёмки.**

- `RISK-FR-020-AC-01` — Неизвестный или неподтверждённый signal не считается достоверным.
- `RISK-FR-020-AC-02` — Device identifier минимизирован и защищён.

**Трассировка для следующего этапа:**

```yaml
requirement_id: RISK-FR-020
capability: CAP-RISK-06
owner_context: Risk Decision
contracts:
  api: []
  events: []
spdd:
  feature_prompt: null
  task_prompts: []
verification:
  tests: []
  release_evidence: []
```

### RISK-FR-021. Проверить velocity {#risk-fr-021}

| Поле | Значение |
| --- | --- |
| Тип | `functional` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Risk Decision` / `m8-risk-decision` |
| Business capability | `CAP-RISK-07` |
| Согласованность | C0 — локальная строгая согласованность |
| Зависимости | Identity projections, Resource Manager projections, Audit |
| Данные | RiskAssessment, RiskPolicy, RiskSignal |
| Безопасность | sensitive decision access |
| API | — |
| События | — |
| Атрибуты качества | — |
| Основание PADS | §6.12, §7, §9.6, §15 |

**Требование.**

Оценить частоту действий по actor, subject, client, device, network и resource keys.

**Критерии приёмки.**

- `RISK-FR-021-AC-01` — Окна и thresholds задаются policy.
- `RISK-FR-021-AC-02` — Повтор одного idempotent request не удваивает business count.

**Трассировка для следующего этапа:**

```yaml
requirement_id: RISK-FR-021
capability: CAP-RISK-07
owner_context: Risk Decision
contracts:
  api: []
  events: []
spdd:
  feature_prompt: null
  task_prompts: []
verification:
  tests: []
  release_evidence: []
```

### RISK-FR-022. Принять external risk signal {#risk-fr-022}

| Поле | Значение |
| --- | --- |
| Тип | `functional` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Risk Decision` / `m8-risk-decision` |
| Business capability | `CAP-RISK-02` |
| Согласованность | C0 — локальная строгая согласованность |
| Зависимости | Identity projections, Resource Manager projections, Audit |
| Данные | RiskAssessment, RiskPolicy, RiskSignal |
| Безопасность | sensitive decision access |
| API | — |
| События | — |
| Атрибуты качества | — |
| Основание PADS | §6.12, §7, §9.6, §15 |

**Требование.**

Нормализовать сигнал внешнего источника через ACL и учесть его provenance.

**Критерии приёмки.**

- `RISK-FR-022-AC-01` — Сбой источника не подменяется безопасным значением.
- `RISK-FR-022-AC-02` — Сигнал имеет expiry.

**Трассировка для следующего этапа:**

```yaml
requirement_id: RISK-FR-022
capability: CAP-RISK-02
owner_context: Risk Decision
contracts:
  api: []
  events: []
spdd:
  feature_prompt: null
  task_prompts: []
verification:
  tests: []
  release_evidence: []
```

### RISK-FR-030. Создать Manual Review {#risk-fr-030}

| Поле | Значение |
| --- | --- |
| Тип | `functional` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Risk Decision` / `m8-risk-decision` |
| Business capability | `CAP-RISK-10` |
| Согласованность | C0 — локальная строгая согласованность |
| Зависимости | Identity projections, Resource Manager projections, Audit |
| Данные | RiskAssessment, RiskPolicy, RiskSignal |
| Безопасность | sensitive decision access |
| API | — |
| События | — |
| Атрибуты качества | — |
| Основание PADS | §6.12, §7, §9.6, §15 |

**Требование.**

Создать review case для outcome REVIEW с snapshot входных данных и SLA.

**Критерии приёмки.**

- `RISK-FR-030-AC-01` — Case не содержит запрещённых секретов.
- `RISK-FR-030-AC-02` — Повтор assessment не создаёт дубликат case.

**Трассировка для следующего этапа:**

```yaml
requirement_id: RISK-FR-030
capability: CAP-RISK-10
owner_context: Risk Decision
contracts:
  api: []
  events: []
spdd:
  feature_prompt: null
  task_prompts: []
verification:
  tests: []
  release_evidence: []
```

### RISK-FR-031. Завершить Manual Review {#risk-fr-031}

| Поле | Значение |
| --- | --- |
| Тип | `functional` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Risk Decision` / `m8-risk-decision` |
| Business capability | `CAP-RISK-10` |
| Согласованность | C0 — локальная строгая согласованность |
| Зависимости | Identity projections, Resource Manager projections, Audit |
| Данные | RiskAssessment, RiskPolicy, RiskSignal |
| Безопасность | sensitive decision access |
| API | — |
| События | — |
| Атрибуты качества | — |
| Основание PADS | §6.12, §7, §9.6, §15 |

**Требование.**

Записать решение reviewer и опубликовать итог для process owner.

**Критерии приёмки.**

- `RISK-FR-031-AC-01` — Reviewer action требует permission и audit.
- `RISK-FR-031-AC-02` — Решение после SLA expiry обрабатывается по owner policy.

**Трассировка для следующего этапа:**

```yaml
requirement_id: RISK-FR-031
capability: CAP-RISK-10
owner_context: Risk Decision
contracts:
  api: []
  events: []
spdd:
  feature_prompt: null
  task_prompts: []
verification:
  tests: []
  release_evidence: []
```

### RISK-FR-040. Сформировать feedback {#risk-fr-040}

| Поле | Значение |
| --- | --- |
| Тип | `functional` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Risk Decision` / `m8-risk-decision` |
| Business capability | `CAP-RISK-11` |
| Согласованность | C0 — локальная строгая согласованность |
| Зависимости | Identity projections, Resource Manager projections, Audit |
| Данные | RiskAssessment, RiskPolicy, RiskSignal |
| Безопасность | sensitive decision access |
| API | — |
| События | — |
| Атрибуты качества | — |
| Основание PADS | §6.12, §7, §9.6, §15 |

**Требование.**

Принять подтверждённый outcome/incident feedback для анализа качества policy.

**Критерии приёмки.**

- `RISK-FR-040-AC-01` — Feedback не меняет исторический decision.
- `RISK-FR-040-AC-02` — Источник и confidence фиксируются.

**Трассировка для следующего этапа:**

```yaml
requirement_id: RISK-FR-040
capability: CAP-RISK-11
owner_context: Risk Decision
contracts:
  api: []
  events: []
spdd:
  feature_prompt: null
  task_prompts: []
verification:
  tests: []
  release_evidence: []
```

### RISK-FR-050. Публиковать факты Risk {#risk-fr-050}

| Поле | Значение |
| --- | --- |
| Тип | `functional` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Risk Decision` / `m8-risk-decision` |
| Business capability | `CAP-RISK-12` |
| Согласованность | C0 — локальная строгая согласованность |
| Зависимости | Identity projections, Resource Manager projections, Audit |
| Данные | RiskAssessment, RiskPolicy, RiskSignal |
| Безопасность | sensitive decision access |
| API | — |
| События | — |
| Атрибуты качества | — |
| Основание PADS | §6.12, §7, §9.6, §15 |

**Требование.**

Публиковать минимизированные события о assessments, policy lifecycle и reviews.

**Критерии приёмки.**

- `RISK-FR-050-AC-01` — Чувствительные signal values не публикуются без необходимости.
- `RISK-FR-050-AC-02` — Event связан с action correlation.

**Трассировка для следующего этапа:**

```yaml
requirement_id: RISK-FR-050
capability: CAP-RISK-12
owner_context: Risk Decision
contracts:
  api: []
  events: []
spdd:
  feature_prompt: null
  task_prompts: []
verification:
  tests: []
  release_evidence: []
```

### RISK-DATA-001. Неизменяемость Risk Assessment {#risk-data-001}

| Поле | Значение |
| --- | --- |
| Тип | `data` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Risk Decision` / `m8-risk-decision` |
| Business capability | `CAP-RISK-04` |
| Согласованность | C0 — локальная строгая согласованность |
| Зависимости | — |
| Данные | — |
| Безопасность | — |
| API | — |
| События | — |
| Атрибуты качества | — |
| Основание PADS | §11 |

**Требование.**

Зафиксированный Risk Assessment должен быть неизменяемым snapshot решения, policy version и нормализованных входных ссылок.

**Критерии приёмки.**

- `RISK-DATA-001-AC-01` — Correction создаёт новое assessment с reference на предыдущее.
- `RISK-DATA-001-AC-02` — Retention входных signals соответствует privacy policy.

**Трассировка для следующего этапа:**

```yaml
requirement_id: RISK-DATA-001
capability: CAP-RISK-04
owner_context: Risk Decision
contracts:
  api: []
  events: []
spdd:
  feature_prompt: null
  task_prompts: []
verification:
  tests: []
  release_evidence: []
```

### RISK-SEC-001. Защита моделей и reason disclosure {#risk-sec-001}

| Поле | Значение |
| --- | --- |
| Тип | `security` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Risk Decision` / `m8-risk-decision` |
| Business capability | `CAP-RISK-05` |
| Согласованность | C0 — локальная строгая согласованность |
| Зависимости | — |
| Данные | — |
| Безопасность | — |
| API | — |
| События | — |
| Атрибуты качества | — |
| Основание PADS | §15 |

**Требование.**

Доступ к правилам, thresholds, raw signals и расширенному объяснению должен быть ограничен отдельными permissions.

**Критерии приёмки.**

- `RISK-SEC-001-AC-01` — Обычный caller получает достаточный reason code без раскрытия механизмов обхода.
- `RISK-SEC-001-AC-02` — Административное чтение аудируется.

**Трассировка для следующего этапа:**

```yaml
requirement_id: RISK-SEC-001
capability: CAP-RISK-05
owner_context: Risk Decision
contracts:
  api: []
  events: []
spdd:
  feature_prompt: null
  task_prompts: []
verification:
  tests: []
  release_evidence: []
```

### RISK-NFR-001. Задержка онлайн-решения {#risk-nfr-001}

| Поле | Значение |
| --- | --- |
| Тип | `non-functional` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Risk Decision` / `m8-risk-decision` |
| Business capability | `CAP-RISK-01` |
| Согласованность | C0 — локальная строгая согласованность |
| Зависимости | — |
| Данные | — |
| Безопасность | — |
| API | — |
| События | — |
| Атрибуты качества | p95<=80ms, p99<=200ms |
| Основание PADS | §19 |

**Требование.**

Онлайн Risk Evaluate должен иметь p95 не более 80 мс и p99 не более 200 мс для стандартной policy.

**Критерии приёмки.**

- `RISK-NFR-001-AC-01` — Медленный внешний signal не блокирует beyond deadline; применяется policy missing-signal behavior.
- `RISK-NFR-001-AC-02` — Timeout outcome детерминирован и наблюдаем.

**Трассировка для следующего этапа:**

```yaml
requirement_id: RISK-NFR-001
capability: CAP-RISK-01
owner_context: Risk Decision
contracts:
  api: []
  events: []
spdd:
  feature_prompt: null
  task_prompts: []
verification:
  tests: []
  release_evidence: []
```
