---
title: "Requirements Catalog: Authentication"
description: "Требования Authentication."
keywords:
  - "M8 Platform"
  - "requirements"
---

# 6. Authentication {#requirements-authentication}

{% note info "Навигация Requirements Catalog" %}

[Оглавление требований](../index.md) | [PADS](../../pads/index.md) | [Engineering artifacts](../../../engineering-artifacts/index.md) | [Предыдущий раздел: 5. Identity](05-identity.md) | [Следующий раздел: 7. Access](07-access.md)

{% endnote %}

Владелец раздела: **Authentication**. Требований: **36**.

## Реестр

| ID | Тип | Приоритет | Capability | Название | Статус |
| --- | --- | --- | --- | --- | --- |
| `AUTH-FR-001` | functional | Must | `CAP-AUTHN-03` | Начать AuthenticationTransaction | `ANALYZED` |
| `AUTH-FR-002` | functional | Must | `CAP-AUTHN-02` | Выбрать Authentication Provider | `ANALYZED` |
| `AUTH-FR-003` | functional | Must | `CAP-AUTHN-04` | Создать Challenge | `ANALYZED` |
| `AUTH-FR-004` | functional | Must | `CAP-AUTHN-11` | Получить состояние Authentication | `ANALYZED` |
| `AUTH-FR-005` | functional | Must | `CAP-AUTHN-11` | Ожидать изменение Authentication | `ANALYZED` |
| `AUTH-FR-006` | functional | Must | `CAP-AUTHN-11` | Отменить Authentication | `ANALYZED` |
| `AUTH-FR-007` | functional | Must | `CAP-AUTHN-04` | Завершить Authentication по provider callback | `ANALYZED` |
| `AUTH-FR-008` | functional | Must | `CAP-AUTHN-11` | Истечь AuthenticationTransaction | `ANALYZED` |
| `AUTH-FR-009` | functional | Must | `CAP-AUTHN-11` | Завершить Authentication отказом | `ANALYZED` |
| `AUTH-FR-010` | functional | Must | `CAP-AUTHN-12` | Создать Handoff | `ANALYZED` |
| `AUTH-FR-011` | functional | Must | `CAP-AUTHN-12` | Погасить Handoff | `ANALYZED` |
| `AUTH-FR-012` | functional | Must | `CAP-AUTHN-13` | Создать Authentication Session reference | `ANALYZED` |
| `AUTH-FR-013` | functional | Must | `CAP-AUTHN-13` | Отозвать Authentication Session | `ANALYZED` |
| `AUTH-FR-014` | functional | Must | `CAP-AUTHN-01` | Проверить активность Client | `ANALYZED` |
| `AUTH-FR-015` | functional | Must | `CAP-AUTHN-01` | Зарегистрировать Client | `ANALYZED` |
| `AUTH-FR-016` | functional | Must | `CAP-AUTHN-01` | Изменить Client policy | `ANALYZED` |
| `AUTH-FR-017` | functional | Must | `CAP-AUTHN-09` | Повторная аутентификация после невозможности refresh | `ANALYZED` |
| `AUTH-FR-018` | functional | Must | `CAP-AUTHN-05` | Запустить CIBA-аутентификацию | `ANALYZED` |
| `AUTH-FR-019` | functional | Must | `CAP-AUTHN-05` | Обработать CIBA approval или denial | `ANALYZED` |
| `AUTH-FR-020` | functional | Must | `CAP-AUTHN-10` | Выполнить step-up | `ANALYZED` |
| `AUTH-FR-021` | functional | Must | `CAP-AUTHN-10` | Выбрать step-up challenge | `ANALYZED` |
| `AUTH-FR-022` | functional | Must | `CAP-AUTHN-06` | Отправить OTP challenge | `ANALYZED` |
| `AUTH-FR-023` | functional | Must | `CAP-AUTHN-06` | Проверить OTP | `ANALYZED` |
| `AUTH-FR-024` | functional | Must | `CAP-AUTHN-04` | Повторить отправку Challenge | `ANALYZED` |
| `AUTH-FR-025` | functional | Must | `CAP-AUTHN-07` | Проверить WebAuthn assertion | `ANALYZED` |
| `AUTH-FR-026` | functional | Must | `CAP-AUTHN-07` | Зарегистрировать WebAuthn credential | `ANALYZED` |
| `AUTH-FR-027` | functional | Must | `CAP-AUTHN-08` | Начать федеративную аутентификацию | `ANALYZED` |
| `AUTH-FR-028` | functional | Must | `CAP-AUTHN-08` | Обработать federated callback | `ANALYZED` |
| `AUTH-FR-029` | functional | Must | `CAP-AUTHN-13` | Отозвать Client access | `ANALYZED` |
| `AUTH-FR-030` | functional | Must | `CAP-AUTHN-14` | Публиковать факты аутентификации | `ANALYZED` |
| `AUTH-DATA-001` | data | Must | `CAP-AUTHN-11` | Источник истины AuthenticationTransaction | `ANALYZED` |
| `AUTH-DATA-002` | data | Must | `CAP-AUTHN-11` | Срок хранения authentication data | `ANALYZED` |
| `AUTH-SEC-001` | security | Must | `CAP-AUTHN-12` | Защита callback и handoff | `ANALYZED` |
| `AUTH-SEC-002` | security | Must | `CAP-AUTHN-04` | Ограничение частоты authentication attempts | `ANALYZED` |
| `AUTH-NFR-001` | non-functional | Must | `CAP-AUTHN-03` | Задержка старта Authentication | `ANALYZED` |
| `AUTH-NFR-002` | non-functional | Must | `CAP-AUTHN-11` | Доступность проверки состояния | `ANALYZED` |

## Детальные требования

### AUTH-FR-001. Начать AuthenticationTransaction {#auth-fr-001}

| Поле | Значение |
| --- | --- |
| Тип | `functional` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Authentication` / `m8-authentication` |
| Business capability | `CAP-AUTHN-03` |
| Согласованность | C0 — локальная строгая согласованность |
| Зависимости | Identity, Risk Decision, Access, Keycloak/provider ACL, Audit |
| Данные | AuthenticationTransaction, AuthenticationChallenge |
| Безопасность | client authentication, risk/assurance policy, secret minimization |
| API | — |
| События | — |
| Атрибуты качества | — |
| Основание PADS | §6.10, §7, §9.4, §15, §16 |

**Требование.**

Создать новую AuthenticationTransaction после проверки Client, Subject, разрешённого метода и начального решения риска.

**Критерии приёмки.**

- `AUTH-FR-001-AC-01` — Возвращаются authentication_id, state, expires_at и Operation при длительном старте.
- `AUTH-FR-001-AC-02` — Транзакция и Outbox фиксируются атомарно.

**Трассировка для следующего этапа:**

```yaml
requirement_id: AUTH-FR-001
capability: CAP-AUTHN-03
owner_context: Authentication
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

### AUTH-FR-002. Выбрать Authentication Provider {#auth-fr-002}

| Поле | Значение |
| --- | --- |
| Тип | `functional` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Authentication` / `m8-authentication` |
| Business capability | `CAP-AUTHN-02` |
| Согласованность | C0 — локальная строгая согласованность |
| Зависимости | Identity, Risk Decision, Access, Keycloak/provider ACL, Audit |
| Данные | AuthenticationTransaction, AuthenticationChallenge |
| Безопасность | client authentication, risk/assurance policy, secret minimization |
| API | — |
| События | — |
| Атрибуты качества | — |
| Основание PADS | §6.10, §7, §9.4, §15, §16 |

**Требование.**

Выбрать допустимого provider на основании Client policy, requested method, Subject и доступности.

**Критерии приёмки.**

- `AUTH-FR-002-AC-01` — Недоступный requested provider возвращает контролируемую ошибку или policy fallback.
- `AUTH-FR-002-AC-02` — Выбор фиксируется без раскрытия внутренней конфигурации.

**Трассировка для следующего этапа:**

```yaml
requirement_id: AUTH-FR-002
capability: CAP-AUTHN-02
owner_context: Authentication
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

### AUTH-FR-003. Создать Challenge {#auth-fr-003}

| Поле | Значение |
| --- | --- |
| Тип | `functional` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Authentication` / `m8-authentication` |
| Business capability | `CAP-AUTHN-04` |
| Согласованность | C0 — локальная строгая согласованность |
| Зависимости | Identity, Risk Decision, Access, Keycloak/provider ACL, Audit |
| Данные | AuthenticationTransaction, AuthenticationChallenge |
| Безопасность | client authentication, risk/assurance policy, secret minimization |
| API | — |
| События | — |
| Атрибуты качества | — |
| Основание PADS | §6.10, §7, §9.4, §15, §16 |

**Требование.**

Создать challenge допустимого типа и связать его с AuthenticationTransaction.

**Критерии приёмки.**

- `AUTH-FR-003-AC-01` — Одновременно активные challenges соответствуют policy.
- `AUTH-FR-003-AC-02` — Challenge secret не хранится в открытом виде.

**Трассировка для следующего этапа:**

```yaml
requirement_id: AUTH-FR-003
capability: CAP-AUTHN-04
owner_context: Authentication
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

### AUTH-FR-004. Получить состояние Authentication {#auth-fr-004}

| Поле | Значение |
| --- | --- |
| Тип | `functional` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Authentication` / `m8-authentication` |
| Business capability | `CAP-AUTHN-11` |
| Согласованность | C0 — локальная строгая согласованность |
| Зависимости | Identity, Risk Decision, Access, Keycloak/provider ACL, Audit |
| Данные | AuthenticationTransaction, AuthenticationChallenge |
| Безопасность | client authentication, risk/assurance policy, secret minimization |
| API | — |
| События | — |
| Атрибуты качества | — |
| Основание PADS | §6.10, §7, §9.4, §15, §16 |

**Требование.**

Получить текущее состояние транзакции и активного challenge без раскрытия секретов.

**Критерии приёмки.**

- `AUTH-FR-004-AC-01` — State отражает committed domain state.
- `AUTH-FR-004-AC-02` — Caller видит только собственную или разрешённую транзакцию.

**Трассировка для следующего этапа:**

```yaml
requirement_id: AUTH-FR-004
capability: CAP-AUTHN-11
owner_context: Authentication
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

### AUTH-FR-005. Ожидать изменение Authentication {#auth-fr-005}

| Поле | Значение |
| --- | --- |
| Тип | `functional` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Authentication` / `m8-authentication` |
| Business capability | `CAP-AUTHN-11` |
| Согласованность | C0 — локальная строгая согласованность |
| Зависимости | Identity, Risk Decision, Access, Keycloak/provider ACL, Audit |
| Данные | AuthenticationTransaction, AuthenticationChallenge |
| Безопасность | client authentication, risk/assurance policy, secret minimization |
| API | — |
| События | — |
| Атрибуты качества | — |
| Основание PADS | §6.10, §7, §9.4, §15, §16 |

**Требование.**

Ожидать изменение состояния без создания новой транзакции.

**Критерии приёмки.**

- `AUTH-FR-005-AC-01` — Timeout возвращает текущее состояние.
- `AUTH-FR-005-AC-02` — Повтор wait безопасен и не изменяет агрегат.

**Трассировка для следующего этапа:**

```yaml
requirement_id: AUTH-FR-005
capability: CAP-AUTHN-11
owner_context: Authentication
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

### AUTH-FR-006. Отменить Authentication {#auth-fr-006}

| Поле | Значение |
| --- | --- |
| Тип | `functional` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Authentication` / `m8-authentication` |
| Business capability | `CAP-AUTHN-11` |
| Согласованность | C0 — локальная строгая согласованность |
| Зависимости | Identity, Risk Decision, Access, Keycloak/provider ACL, Audit |
| Данные | AuthenticationTransaction, AuthenticationChallenge |
| Безопасность | client authentication, risk/assurance policy, secret minimization |
| API | — |
| События | — |
| Атрибуты качества | — |
| Основание PADS | §6.10, §7, §9.4, §15, §16 |

**Требование.**

Отменить незавершённую транзакцию при допустимом состоянии и правах.

**Критерии приёмки.**

- `AUTH-FR-006-AC-01` — COMPLETED транзакция не переходит в CANCELLED.
- `AUTH-FR-006-AC-02` — Provider cancellation выполняется best effort и фиксируется отдельно.

**Трассировка для следующего этапа:**

```yaml
requirement_id: AUTH-FR-006
capability: CAP-AUTHN-11
owner_context: Authentication
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

### AUTH-FR-007. Завершить Authentication по provider callback {#auth-fr-007}

| Поле | Значение |
| --- | --- |
| Тип | `functional` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Authentication` / `m8-authentication` |
| Business capability | `CAP-AUTHN-04` |
| Согласованность | C0 — локальная строгая согласованность |
| Зависимости | Identity, Risk Decision, Access, Keycloak/provider ACL, Audit |
| Данные | AuthenticationTransaction, AuthenticationChallenge |
| Безопасность | client authentication, risk/assurance policy, secret minimization |
| API | — |
| События | — |
| Атрибуты качества | — |
| Основание PADS | §6.10, §7, §9.4, §15, §16 |

**Требование.**

Обработать проверенный callback поставщика и выполнить допустимый переход состояния.

**Критерии приёмки.**

- `AUTH-FR-007-AC-01` — Callback аутентифицируется и дедуплицируется.
- `AUTH-FR-007-AC-02` — Повтор callback не создаёт вторую session/handoff.

**Трассировка для следующего этапа:**

```yaml
requirement_id: AUTH-FR-007
capability: CAP-AUTHN-04
owner_context: Authentication
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

### AUTH-FR-008. Истечь AuthenticationTransaction {#auth-fr-008}

| Поле | Значение |
| --- | --- |
| Тип | `functional` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Authentication` / `m8-authentication` |
| Business capability | `CAP-AUTHN-11` |
| Согласованность | C0 — локальная строгая согласованность |
| Зависимости | Identity, Risk Decision, Access, Keycloak/provider ACL, Audit |
| Данные | AuthenticationTransaction, AuthenticationChallenge |
| Безопасность | client authentication, risk/assurance policy, secret minimization |
| API | — |
| События | — |
| Атрибуты качества | — |
| Основание PADS | §6.10, §7, §9.4, §15, §16 |

**Требование.**

Перевести незавершённую транзакцию в EXPIRED после deadline.

**Критерии приёмки.**

- `AUTH-FR-008-AC-01` — Истечение идемпотентно.
- `AUTH-FR-008-AC-02` — Поздний callback не возобновляет expired transaction.

**Трассировка для следующего этапа:**

```yaml
requirement_id: AUTH-FR-008
capability: CAP-AUTHN-11
owner_context: Authentication
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

### AUTH-FR-009. Завершить Authentication отказом {#auth-fr-009}

| Поле | Значение |
| --- | --- |
| Тип | `functional` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Authentication` / `m8-authentication` |
| Business capability | `CAP-AUTHN-11` |
| Согласованность | C0 — локальная строгая согласованность |
| Зависимости | Identity, Risk Decision, Access, Keycloak/provider ACL, Audit |
| Данные | AuthenticationTransaction, AuthenticationChallenge |
| Безопасность | client authentication, risk/assurance policy, secret minimization |
| API | — |
| События | — |
| Атрибуты качества | — |
| Основание PADS | §6.10, §7, §9.4, §15, §16 |

**Требование.**

Завершить транзакцию FAILED с канонической безопасной причиной.

**Критерии приёмки.**

- `AUTH-FR-009-AC-01` — Внешняя техническая ошибка отделена от user denial.
- `AUTH-FR-009-AC-02` — Ошибка не содержит секретов provider.

**Трассировка для следующего этапа:**

```yaml
requirement_id: AUTH-FR-009
capability: CAP-AUTHN-11
owner_context: Authentication
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

### AUTH-FR-010. Создать Handoff {#auth-fr-010}

| Поле | Значение |
| --- | --- |
| Тип | `functional` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Authentication` / `m8-authentication` |
| Business capability | `CAP-AUTHN-12` |
| Согласованность | C0 — локальная строгая согласованность |
| Зависимости | Identity, Risk Decision, Access, Keycloak/provider ACL, Audit |
| Данные | AuthenticationTransaction, AuthenticationChallenge |
| Безопасность | client authentication, risk/assurance policy, secret minimization |
| API | — |
| События | — |
| Атрибуты качества | — |
| Основание PADS | §6.10, §7, §9.4, §15, §16 |

**Требование.**

После успешной аутентификации создать одноразовый безопасный handoff для следующего протокольного компонента.

**Критерии приёмки.**

- `AUTH-FR-010-AC-01` — Handoff имеет короткий TTL и одноразовое использование.
- `AUTH-FR-010-AC-02` — Raw tokens не возвращаются там, где контракт требует authorization code/session handoff.

**Трассировка для следующего этапа:**

```yaml
requirement_id: AUTH-FR-010
capability: CAP-AUTHN-12
owner_context: Authentication
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

### AUTH-FR-011. Погасить Handoff {#auth-fr-011}

| Поле | Значение |
| --- | --- |
| Тип | `functional` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Authentication` / `m8-authentication` |
| Business capability | `CAP-AUTHN-12` |
| Согласованность | C0 — локальная строгая согласованность |
| Зависимости | Identity, Risk Decision, Access, Keycloak/provider ACL, Audit |
| Данные | AuthenticationTransaction, AuthenticationChallenge |
| Безопасность | client authentication, risk/assurance policy, secret minimization |
| API | — |
| События | — |
| Атрибуты качества | — |
| Основание PADS | §6.10, §7, §9.4, §15, §16 |

**Требование.**

Обменять действующий handoff ровно один раз проверенным Client.

**Критерии приёмки.**

- `AUTH-FR-011-AC-01` — Повтор использования отклоняется.
- `AUTH-FR-011-AC-02` — Client binding проверяется до раскрытия результата.

**Трассировка для следующего этапа:**

```yaml
requirement_id: AUTH-FR-011
capability: CAP-AUTHN-12
owner_context: Authentication
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

### AUTH-FR-012. Создать Authentication Session reference {#auth-fr-012}

| Поле | Значение |
| --- | --- |
| Тип | `functional` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Authentication` / `m8-authentication` |
| Business capability | `CAP-AUTHN-13` |
| Согласованность | C0 — локальная строгая согласованность |
| Зависимости | Identity, Risk Decision, Access, Keycloak/provider ACL, Audit |
| Данные | AuthenticationTransaction, AuthenticationChallenge |
| Безопасность | client authentication, risk/assurance policy, secret minimization |
| API | — |
| События | — |
| Атрибуты качества | — |
| Основание PADS | §6.10, §7, §9.4, §15, §16 |

**Требование.**

Связать успешную аутентификацию с внутренним session reference без копирования полной модели Keycloak.

**Критерии приёмки.**

- `AUTH-FR-012-AC-01` — Session reference не является access token.
- `AUTH-FR-012-AC-02` — Revocation state синхронизируется по определённому контракту.

**Трассировка для следующего этапа:**

```yaml
requirement_id: AUTH-FR-012
capability: CAP-AUTHN-13
owner_context: Authentication
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

### AUTH-FR-013. Отозвать Authentication Session {#auth-fr-013}

| Поле | Значение |
| --- | --- |
| Тип | `functional` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Authentication` / `m8-authentication` |
| Business capability | `CAP-AUTHN-13` |
| Согласованность | C0 — локальная строгая согласованность |
| Зависимости | Identity, Risk Decision, Access, Keycloak/provider ACL, Audit |
| Данные | AuthenticationTransaction, AuthenticationChallenge |
| Безопасность | client authentication, risk/assurance policy, secret minimization |
| API | — |
| События | — |
| Атрибуты качества | — |
| Основание PADS | §6.10, §7, §9.4, §15, §16 |

**Требование.**

Инициировать отзыв сессии и связанных continuation capabilities.

**Критерии приёмки.**

- `AUTH-FR-013-AC-01` — Отзыв идемпотентен.
- `AUTH-FR-013-AC-02` — Завершение внешнего provider session отслеживается Operation или событием.

**Трассировка для следующего этапа:**

```yaml
requirement_id: AUTH-FR-013
capability: CAP-AUTHN-13
owner_context: Authentication
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

### AUTH-FR-014. Проверить активность Client {#auth-fr-014}

| Поле | Значение |
| --- | --- |
| Тип | `functional` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Authentication` / `m8-authentication` |
| Business capability | `CAP-AUTHN-01` |
| Согласованность | C0 — локальная строгая согласованность |
| Зависимости | Identity, Risk Decision, Access, Keycloak/provider ACL, Audit |
| Данные | AuthenticationTransaction, AuthenticationChallenge |
| Безопасность | client authentication, risk/assurance policy, secret minimization |
| API | — |
| События | — |
| Атрибуты качества | — |
| Основание PADS | §6.10, §7, §9.4, §15, §16 |

**Требование.**

Проверять состояние, разрешённые flows, redirect/handoff policy и требования AAL Client.

**Критерии приёмки.**

- `AUTH-FR-014-AC-01` — Disabled Client не может начать flow.
- `AUTH-FR-014-AC-02` — Изменение Client policy не меняет уже завершённую транзакцию.

**Трассировка для следующего этапа:**

```yaml
requirement_id: AUTH-FR-014
capability: CAP-AUTHN-01
owner_context: Authentication
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

### AUTH-FR-015. Зарегистрировать Client {#auth-fr-015}

| Поле | Значение |
| --- | --- |
| Тип | `functional` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Authentication` / `m8-authentication` |
| Business capability | `CAP-AUTHN-01` |
| Согласованность | C0 — локальная строгая согласованность |
| Зависимости | Identity, Risk Decision, Access, Keycloak/provider ACL, Audit |
| Данные | AuthenticationTransaction, AuthenticationChallenge |
| Безопасность | client authentication, risk/assurance policy, secret minimization |
| API | — |
| События | — |
| Атрибуты качества | — |
| Основание PADS | §6.10, §7, §9.4, §15, §16 |

**Требование.**

Создать Client с типом, разрешёнными flows и security policy.

**Критерии приёмки.**

- `AUTH-FR-015-AC-01` — Секрет создаётся или привязывается через безопасный механизм.
- `AUTH-FR-015-AC-02` — Public и confidential client имеют разные обязательные проверки.

**Трассировка для следующего этапа:**

```yaml
requirement_id: AUTH-FR-015
capability: CAP-AUTHN-01
owner_context: Authentication
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

### AUTH-FR-016. Изменить Client policy {#auth-fr-016}

| Поле | Значение |
| --- | --- |
| Тип | `functional` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Authentication` / `m8-authentication` |
| Business capability | `CAP-AUTHN-01` |
| Согласованность | C0 — локальная строгая согласованность |
| Зависимости | Identity, Risk Decision, Access, Keycloak/provider ACL, Audit |
| Данные | AuthenticationTransaction, AuthenticationChallenge |
| Безопасность | client authentication, risk/assurance policy, secret minimization |
| API | — |
| События | — |
| Атрибуты качества | — |
| Основание PADS | §6.10, §7, §9.4, §15, §16 |

**Требование.**

Изменять разрешённые methods, providers, assurance и handoff settings с revision.

**Критерии приёмки.**

- `AUTH-FR-016-AC-01` — Ослабление критической policy требует отдельного permission и audit.
- `AUTH-FR-016-AC-02` — Несовместимые active flows обрабатываются по documented policy.

**Трассировка для следующего этапа:**

```yaml
requirement_id: AUTH-FR-016
capability: CAP-AUTHN-01
owner_context: Authentication
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

### AUTH-FR-017. Повторная аутентификация после невозможности refresh {#auth-fr-017}

| Поле | Значение |
| --- | --- |
| Тип | `functional` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Authentication` / `m8-authentication` |
| Business capability | `CAP-AUTHN-09` |
| Согласованность | C0 — локальная строгая согласованность |
| Зависимости | Identity, Risk Decision, Access, Keycloak/provider ACL, Audit |
| Данные | AuthenticationTransaction, AuthenticationChallenge |
| Безопасность | client authentication, risk/assurance policy, secret minimization |
| API | — |
| События | — |
| Атрибуты качества | — |
| Основание PADS | §6.10, §7, §9.4, §15, §16 |

**Требование.**

Создать новую AuthenticationTransaction, если refresh token отсутствует, просрочен, отозван или не может быть безопасно использован.

**Критерии приёмки.**

- `AUTH-FR-017-AC-01` — Новая транзакция не продолжает failed refresh attempt.
- `AUTH-FR-017-AC-02` — Одинаковый idempotency key возвращает один authentication_id.

**Трассировка для следующего этапа:**

```yaml
requirement_id: AUTH-FR-017
capability: CAP-AUTHN-09
owner_context: Authentication
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

### AUTH-FR-018. Запустить CIBA-аутентификацию {#auth-fr-018}

| Поле | Значение |
| --- | --- |
| Тип | `functional` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Authentication` / `m8-authentication` |
| Business capability | `CAP-AUTHN-05` |
| Согласованность | C0 — локальная строгая согласованность |
| Зависимости | Identity, Risk Decision, Access, Keycloak/provider ACL, Audit |
| Данные | AuthenticationTransaction, AuthenticationChallenge |
| Безопасность | client authentication, risk/assurance policy, secret minimization |
| API | — |
| События | — |
| Атрибуты качества | — |
| Основание PADS | §6.10, §7, §9.4, §15, §16 |

**Требование.**

Запустить backchannel authentication request и отслеживать подтверждение пользователя.

**Критерии приёмки.**

- `AUTH-FR-018-AC-01` — Клиент не получает device-bound secret.
- `AUTH-FR-018-AC-02` — Polling/push поведение соответствует policy и rate limits.

**Трассировка для следующего этапа:**

```yaml
requirement_id: AUTH-FR-018
capability: CAP-AUTHN-05
owner_context: Authentication
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

### AUTH-FR-019. Обработать CIBA approval или denial {#auth-fr-019}

| Поле | Значение |
| --- | --- |
| Тип | `functional` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Authentication` / `m8-authentication` |
| Business capability | `CAP-AUTHN-05` |
| Согласованность | C0 — локальная строгая согласованность |
| Зависимости | Identity, Risk Decision, Access, Keycloak/provider ACL, Audit |
| Данные | AuthenticationTransaction, AuthenticationChallenge |
| Безопасность | client authentication, risk/assurance policy, secret minimization |
| API | — |
| События | — |
| Атрибуты качества | — |
| Основание PADS | §6.10, §7, §9.4, §15, §16 |

**Требование.**

Обработать подтверждение или отказ из доверенного канала.

**Критерии приёмки.**

- `AUTH-FR-019-AC-01` — Approval связан с исходной транзакцией и субъектом.
- `AUTH-FR-019-AC-02` — Повторное подтверждение после completion не меняет результат.

**Трассировка для следующего этапа:**

```yaml
requirement_id: AUTH-FR-019
capability: CAP-AUTHN-05
owner_context: Authentication
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

### AUTH-FR-020. Выполнить step-up {#auth-fr-020}

| Поле | Значение |
| --- | --- |
| Тип | `functional` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Authentication` / `m8-authentication` |
| Business capability | `CAP-AUTHN-10` |
| Согласованность | C0 — локальная строгая согласованность |
| Зависимости | Identity, Risk Decision, Access, Keycloak/provider ACL, Audit |
| Данные | AuthenticationTransaction, AuthenticationChallenge |
| Безопасность | client authentication, risk/assurance policy, secret minimization |
| API | — |
| События | — |
| Атрибуты качества | — |
| Основание PADS | §6.10, §7, §9.4, §15, §16 |

**Требование.**

Повысить achieved assurance до требуемого AAL новой проверкой, сохранив связь с исходным действием.

**Критерии приёмки.**

- `AUTH-FR-020-AC-01` — Step-up не понижает AAL.
- `AUTH-FR-020-AC-02` — Истёкший step-up нельзя применить к новой операции без policy.

**Трассировка для следующего этапа:**

```yaml
requirement_id: AUTH-FR-020
capability: CAP-AUTHN-10
owner_context: Authentication
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

### AUTH-FR-021. Выбрать step-up challenge {#auth-fr-021}

| Поле | Значение |
| --- | --- |
| Тип | `functional` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Authentication` / `m8-authentication` |
| Business capability | `CAP-AUTHN-10` |
| Согласованность | C0 — локальная строгая согласованность |
| Зависимости | Identity, Risk Decision, Access, Keycloak/provider ACL, Audit |
| Данные | AuthenticationTransaction, AuthenticationChallenge |
| Безопасность | client authentication, risk/assurance policy, secret minimization |
| API | — |
| События | — |
| Атрибуты качества | — |
| Основание PADS | §6.10, §7, §9.4, §15, §16 |

**Требование.**

Выбрать метод, достаточный для requested AAL и допустимый для Subject/Client.

**Критерии приёмки.**

- `AUTH-FR-021-AC-01` — Недостаточный метод не считается успешным.
- `AUTH-FR-021-AC-02` — Fallback соответствует policy и Risk Decision.

**Трассировка для следующего этапа:**

```yaml
requirement_id: AUTH-FR-021
capability: CAP-AUTHN-10
owner_context: Authentication
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

### AUTH-FR-022. Отправить OTP challenge {#auth-fr-022}

| Поле | Значение |
| --- | --- |
| Тип | `functional` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Authentication` / `m8-authentication` |
| Business capability | `CAP-AUTHN-06` |
| Согласованность | C0 — локальная строгая согласованность |
| Зависимости | Identity, Risk Decision, Access, Keycloak/provider ACL, Audit |
| Данные | AuthenticationTransaction, AuthenticationChallenge |
| Безопасность | client authentication, risk/assurance policy, secret minimization |
| API | — |
| События | — |
| Атрибуты качества | — |
| Основание PADS | §6.10, §7, §9.4, §15, §16 |

**Требование.**

Отправить одноразовый код на подтверждённый channel с rate limit и TTL.

**Критерии приёмки.**

- `AUTH-FR-022-AC-01` — Ответ не раскрывает полный адрес назначения.
- `AUTH-FR-022-AC-02` — Повторная отправка инвалидирует или сохраняет предыдущий код согласно policy.

**Трассировка для следующего этапа:**

```yaml
requirement_id: AUTH-FR-022
capability: CAP-AUTHN-06
owner_context: Authentication
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

### AUTH-FR-023. Проверить OTP {#auth-fr-023}

| Поле | Значение |
| --- | --- |
| Тип | `functional` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Authentication` / `m8-authentication` |
| Business capability | `CAP-AUTHN-06` |
| Согласованность | C0 — локальная строгая согласованность |
| Зависимости | Identity, Risk Decision, Access, Keycloak/provider ACL, Audit |
| Данные | AuthenticationTransaction, AuthenticationChallenge |
| Безопасность | client authentication, risk/assurance policy, secret minimization |
| API | — |
| События | — |
| Атрибуты качества | — |
| Основание PADS | §6.10, §7, §9.4, §15, §16 |

**Требование.**

Проверить одноразовый код с ограничением попыток и защитой от replay.

**Критерии приёмки.**

- `AUTH-FR-023-AC-01` — Успешный код нельзя использовать повторно.
- `AUTH-FR-023-AC-02` — Превышение попыток блокирует challenge и создаёт risk signal.

**Трассировка для следующего этапа:**

```yaml
requirement_id: AUTH-FR-023
capability: CAP-AUTHN-06
owner_context: Authentication
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

### AUTH-FR-024. Повторить отправку Challenge {#auth-fr-024}

| Поле | Значение |
| --- | --- |
| Тип | `functional` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Authentication` / `m8-authentication` |
| Business capability | `CAP-AUTHN-04` |
| Согласованность | C0 — локальная строгая согласованность |
| Зависимости | Identity, Risk Decision, Access, Keycloak/provider ACL, Audit |
| Данные | AuthenticationTransaction, AuthenticationChallenge |
| Безопасность | client authentication, risk/assurance policy, secret minimization |
| API | — |
| События | — |
| Атрибуты качества | — |
| Основание PADS | §6.10, §7, §9.4, §15, §16 |

**Требование.**

Повторить доставку challenge только после cooldown и в пределах лимита.

**Критерии приёмки.**

- `AUTH-FR-024-AC-01` — Слишком ранний запрос возвращает retry_after.
- `AUTH-FR-024-AC-02` — Повтор не создаёт новую AuthenticationTransaction.

**Трассировка для следующего этапа:**

```yaml
requirement_id: AUTH-FR-024
capability: CAP-AUTHN-04
owner_context: Authentication
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

### AUTH-FR-025. Проверить WebAuthn assertion {#auth-fr-025}

| Поле | Значение |
| --- | --- |
| Тип | `functional` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Authentication` / `m8-authentication` |
| Business capability | `CAP-AUTHN-07` |
| Согласованность | C0 — локальная строгая согласованность |
| Зависимости | Identity, Risk Decision, Access, Keycloak/provider ACL, Audit |
| Данные | AuthenticationTransaction, AuthenticationChallenge |
| Безопасность | client authentication, risk/assurance policy, secret minimization |
| API | — |
| События | — |
| Атрибуты качества | — |
| Основание PADS | §6.10, §7, §9.4, §15, §16 |

**Требование.**

Проверить challenge, origin, rpId, signature, credential state и sign counter.

**Критерии приёмки.**

- `AUTH-FR-025-AC-01` — Replay и неверный origin отклоняются.
- `AUTH-FR-025-AC-02` — Clone suspicion формирует risk signal.

**Трассировка для следующего этапа:**

```yaml
requirement_id: AUTH-FR-025
capability: CAP-AUTHN-07
owner_context: Authentication
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

### AUTH-FR-026. Зарегистрировать WebAuthn credential {#auth-fr-026}

| Поле | Значение |
| --- | --- |
| Тип | `functional` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Authentication` / `m8-authentication` |
| Business capability | `CAP-AUTHN-07` |
| Согласованность | C0 — локальная строгая согласованность |
| Зависимости | Identity, Risk Decision, Access, Keycloak/provider ACL, Audit |
| Данные | AuthenticationTransaction, AuthenticationChallenge |
| Безопасность | client authentication, risk/assurance policy, secret minimization |
| API | — |
| События | — |
| Атрибуты качества | — |
| Основание PADS | §6.10, §7, §9.4, §15, §16 |

**Требование.**

Провести защищённую регистрацию passkey/credential для подтверждённого Subject.

**Критерии приёмки.**

- `AUTH-FR-026-AC-01` — Регистрация требует достаточного AAL.
- `AUTH-FR-026-AC-02` — Credential public data хранится у определённого владельца без private key.

**Трассировка для следующего этапа:**

```yaml
requirement_id: AUTH-FR-026
capability: CAP-AUTHN-07
owner_context: Authentication
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

### AUTH-FR-027. Начать федеративную аутентификацию {#auth-fr-027}

| Поле | Значение |
| --- | --- |
| Тип | `functional` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Authentication` / `m8-authentication` |
| Business capability | `CAP-AUTHN-08` |
| Согласованность | C0 — локальная строгая согласованность |
| Зависимости | Identity, Risk Decision, Access, Keycloak/provider ACL, Audit |
| Данные | AuthenticationTransaction, AuthenticationChallenge |
| Безопасность | client authentication, risk/assurance policy, secret minimization |
| API | — |
| События | — |
| Атрибуты качества | — |
| Основание PADS | §6.10, §7, §9.4, §15, §16 |

**Требование.**

Начать OIDC/SAML flow через provider adapter с защищённым state/nonce.

**Критерии приёмки.**

- `AUTH-FR-027-AC-01` — Callback state и nonce проверяются.
- `AUTH-FR-027-AC-02` — Внешние claims преобразуются через ACL и затем разрешаются Identity.

**Трассировка для следующего этапа:**

```yaml
requirement_id: AUTH-FR-027
capability: CAP-AUTHN-08
owner_context: Authentication
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

### AUTH-FR-028. Обработать federated callback {#auth-fr-028}

| Поле | Значение |
| --- | --- |
| Тип | `functional` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Authentication` / `m8-authentication` |
| Business capability | `CAP-AUTHN-08` |
| Согласованность | C0 — локальная строгая согласованность |
| Зависимости | Identity, Risk Decision, Access, Keycloak/provider ACL, Audit |
| Данные | AuthenticationTransaction, AuthenticationChallenge |
| Безопасность | client authentication, risk/assurance policy, secret minimization |
| API | — |
| События | — |
| Атрибуты качества | — |
| Основание PADS | §6.10, §7, §9.4, §15, §16 |

**Требование.**

Проверить ответ внешнего IdP и завершить challenge или создать контролируемую ошибку linking.

**Критерии приёмки.**

- `AUTH-FR-028-AC-01` — Непроверенный claim не становится Subject.
- `AUTH-FR-028-AC-02` — Конфликт external identity не вызывает автоматический account takeover.

**Трассировка для следующего этапа:**

```yaml
requirement_id: AUTH-FR-028
capability: CAP-AUTHN-08
owner_context: Authentication
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

### AUTH-FR-029. Отозвать Client access {#auth-fr-029}

| Поле | Значение |
| --- | --- |
| Тип | `functional` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Authentication` / `m8-authentication` |
| Business capability | `CAP-AUTHN-13` |
| Согласованность | C0 — локальная строгая согласованность |
| Зависимости | Identity, Risk Decision, Access, Keycloak/provider ACL, Audit |
| Данные | AuthenticationTransaction, AuthenticationChallenge |
| Безопасность | client authentication, risk/assurance policy, secret minimization |
| API | — |
| События | — |
| Атрибуты качества | — |
| Основание PADS | §6.10, §7, §9.4, §15, §16 |

**Требование.**

Отозвать активные session references и handoff capabilities конкретного Client.

**Критерии приёмки.**

- `AUTH-FR-029-AC-01` — Отзыв не затрагивает другие Client без явного scope.
- `AUTH-FR-029-AC-02` — Результат и незавершённые внешние действия видны в Operation.

**Трассировка для следующего этапа:**

```yaml
requirement_id: AUTH-FR-029
capability: CAP-AUTHN-13
owner_context: Authentication
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

### AUTH-FR-030. Публиковать факты аутентификации {#auth-fr-030}

| Поле | Значение |
| --- | --- |
| Тип | `functional` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Authentication` / `m8-authentication` |
| Business capability | `CAP-AUTHN-14` |
| Согласованность | C0 — локальная строгая согласованность |
| Зависимости | Identity, Risk Decision, Access, Keycloak/provider ACL, Audit |
| Данные | AuthenticationTransaction, AuthenticationChallenge |
| Безопасность | client authentication, risk/assurance policy, secret minimization |
| API | — |
| События | — |
| Атрибуты качества | — |
| Основание PADS | §6.10, §7, §9.4, §15, §16 |

**Требование.**

Публиковать минимизированные события о старте, challenge, success, failure, expiry и revocation.

**Критерии приёмки.**

- `AUTH-FR-030-AC-01` — События не содержат token, OTP или raw credential.
- `AUTH-FR-030-AC-02` — Каждое событие имеет transaction revision и correlation.

**Трассировка для следующего этапа:**

```yaml
requirement_id: AUTH-FR-030
capability: CAP-AUTHN-14
owner_context: Authentication
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

### AUTH-DATA-001. Источник истины AuthenticationTransaction {#auth-data-001}

| Поле | Значение |
| --- | --- |
| Тип | `data` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Authentication` / `m8-authentication` |
| Business capability | `CAP-AUTHN-11` |
| Согласованность | C0 — локальная строгая согласованность |
| Зависимости | — |
| Данные | — |
| Безопасность | — |
| API | — |
| События | — |
| Атрибуты качества | — |
| Основание PADS | §11 |

**Требование.**

Authentication должен владеть состоянием транзакции, challenge, requested/achieved AAL и handoff lifecycle независимо от provider state.

**Критерии приёмки.**

- `AUTH-DATA-001-AC-01` — Provider data хранится как external reference и normalized state.
- `AUTH-DATA-001-AC-02` — Callback не может пропустить проверку допустимого domain transition.

**Трассировка для следующего этапа:**

```yaml
requirement_id: AUTH-DATA-001
capability: CAP-AUTHN-11
owner_context: Authentication
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

### AUTH-DATA-002. Срок хранения authentication data {#auth-data-002}

| Поле | Значение |
| --- | --- |
| Тип | `data` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Authentication` / `m8-authentication` |
| Business capability | `CAP-AUTHN-11` |
| Согласованность | C0 — локальная строгая согласованность |
| Зависимости | — |
| Данные | — |
| Безопасность | — |
| API | — |
| События | — |
| Атрибуты качества | — |
| Основание PADS | §11, §15 |

**Требование.**

Транзакции и challenge metadata должны храниться ограниченный срок, достаточный для безопасности, расследования и дедупликации.

**Критерии приёмки.**

- `AUTH-DATA-002-AC-01` — Expired secrets удаляются раньше metadata.
- `AUTH-DATA-002-AC-02` — Retention различает operational record и Audit evidence.

**Трассировка для следующего этапа:**

```yaml
requirement_id: AUTH-DATA-002
capability: CAP-AUTHN-11
owner_context: Authentication
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

### AUTH-SEC-001. Защита callback и handoff {#auth-sec-001}

| Поле | Значение |
| --- | --- |
| Тип | `security` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Authentication` / `m8-authentication` |
| Business capability | `CAP-AUTHN-12` |
| Согласованность | C0 — локальная строгая согласованность |
| Зависимости | — |
| Данные | — |
| Безопасность | — |
| API | — |
| События | — |
| Атрибуты качества | — |
| Основание PADS | §15 |

**Требование.**

Provider callback и handoff redemption должны иметь криптографическую привязку, защиту от replay и проверку Client.

**Критерии приёмки.**

- `AUTH-SEC-001-AC-01` — Повторный callback или handoff не создаёт новый успех.
- `AUTH-SEC-001-AC-02` — Неверный binding отклоняется и формирует security audit.

**Трассировка для следующего этапа:**

```yaml
requirement_id: AUTH-SEC-001
capability: CAP-AUTHN-12
owner_context: Authentication
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

### AUTH-SEC-002. Ограничение частоты authentication attempts {#auth-sec-002}

| Поле | Значение |
| --- | --- |
| Тип | `security` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Authentication` / `m8-authentication` |
| Business capability | `CAP-AUTHN-04` |
| Согласованность | C0 — локальная строгая согласованность |
| Зависимости | — |
| Данные | — |
| Безопасность | — |
| API | — |
| События | — |
| Атрибуты качества | — |
| Основание PADS | §15 |

**Требование.**

Start, resend и verify должны иметь rate/velocity controls по Client, Subject, device и network signals.

**Критерии приёмки.**

- `AUTH-SEC-002-AC-01` — Превышение лимита возвращает retry metadata без раскрытия существования Subject.
- `AUTH-SEC-002-AC-02` — Risk Decision получает агрегированные сигналы.

**Трассировка для следующего этапа:**

```yaml
requirement_id: AUTH-SEC-002
capability: CAP-AUTHN-04
owner_context: Authentication
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

### AUTH-NFR-001. Задержка старта Authentication {#auth-nfr-001}

| Поле | Значение |
| --- | --- |
| Тип | `non-functional` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Authentication` / `m8-authentication` |
| Business capability | `CAP-AUTHN-03` |
| Согласованность | C0 — локальная строгая согласованность |
| Зависимости | — |
| Данные | — |
| Безопасность | — |
| API | — |
| События | — |
| Атрибуты качества | p95<=300ms |
| Основание PADS | §19 |

**Требование.**

StartAuthentication должен иметь p95 не более 300 мс без учёта ожидания пользовательского challenge.

**Критерии приёмки.**

- `AUTH-NFR-001-AC-01` — Измерение включает Identity и Risk calls в nominal mode.
- `AUTH-NFR-001-AC-02` — Provider initiation, требующий длительного времени, возвращает Operation.

**Трассировка для следующего этапа:**

```yaml
requirement_id: AUTH-NFR-001
capability: CAP-AUTHN-03
owner_context: Authentication
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

### AUTH-NFR-002. Доступность проверки состояния {#auth-nfr-002}

| Поле | Значение |
| --- | --- |
| Тип | `non-functional` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Authentication` / `m8-authentication` |
| Business capability | `CAP-AUTHN-11` |
| Согласованность | C0 — локальная строгая согласованность |
| Зависимости | — |
| Данные | — |
| Безопасность | — |
| API | — |
| События | — |
| Атрибуты качества | availability>=99.99%, p95<=100ms |
| Основание PADS | §19 |

**Требование.**

GetAuthentication должен иметь доступность не ниже 99,99% и p95 не более 100 мс.

**Критерии приёмки.**

- `AUTH-NFR-002-AC-01` — Read path не зависит от доступности внешнего provider для committed state.
- `AUTH-NFR-002-AC-02` — Degraded provider state обозначается отдельно.

**Трассировка для следующего этапа:**

```yaml
requirement_id: AUTH-NFR-002
capability: CAP-AUTHN-11
owner_context: Authentication
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
