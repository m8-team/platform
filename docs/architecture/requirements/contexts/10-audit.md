---
title: "Requirements Catalog: Audit"
description: "Требования Audit."
keywords:
  - "M8 Platform"
  - "requirements"
---

# 10. Audit {#requirements-audit}

{% note info "Навигация Requirements Catalog" %}

[Оглавление требований](../index.md) | [PADS](../../pads/index.md) | [Engineering artifacts](../../../engineering-artifacts/index.md) | [Предыдущий раздел: 9. Provisioning](09-provisioning.md) | [Следующий раздел: 11. Common Operation](11-common-operation.md)

{% endnote %}

Владелец раздела: **Audit**. Требований: **19**.

## Реестр

| ID | Тип | Приоритет | Capability | Название | Статус |
| --- | --- | --- | --- | --- | --- |
| `AUD-FR-001` | functional | Must | `CAP-AUD-01` | Принять AuditEvent | `ANALYZED` |
| `AUD-FR-002` | functional | Must | `CAP-AUD-02` | Проверить происхождение AuditEvent | `ANALYZED` |
| `AUD-FR-003` | functional | Must | `CAP-AUD-10` | Проверить минимизацию данных | `ANALYZED` |
| `AUD-FR-004` | functional | Must | `CAP-AUD-03` | Сохранить AuditEvent неизменяемо | `ANALYZED` |
| `AUD-FR-005` | functional | Must | `CAP-AUD-11` | Подтвердить доставку обязательного события | `ANALYZED` |
| `AUD-FR-010` | functional | Must | `CAP-AUD-05` | Искать AuditEvent | `ANALYZED` |
| `AUD-FR-011` | functional | Must | `CAP-AUD-05` | Получить событие по ID | `ANALYZED` |
| `AUD-FR-012` | functional | Must | `CAP-AUD-06` | Построить цепочку по correlation | `ANALYZED` |
| `AUD-FR-013` | functional | Must | `CAP-AUD-06` | Получить историю ресурса | `ANALYZED` |
| `AUD-FR-020` | functional | Must | `CAP-AUD-07` | Создать ExportJob | `ANALYZED` |
| `AUD-FR-021` | functional | Must | `CAP-AUD-07` | Получить экспорт | `ANALYZED` |
| `AUD-FR-030` | functional | Must | `CAP-AUD-08` | Применить Retention Policy | `ANALYZED` |
| `AUD-FR-031` | functional | Must | `CAP-AUD-08` | Установить Legal Hold | `ANALYZED` |
| `AUD-FR-040` | functional | Must | `CAP-AUD-04` | Проверить целостность хранения | `ANALYZED` |
| `AUD-FR-041` | functional | Must | `CAP-AUD-04` | Предоставить integrity proof | `ANALYZED` |
| `AUD-FR-050` | functional | Must | `CAP-AUD-12` | Аудировать действия над Audit | `ANALYZED` |
| `AUD-DATA-001` | data | Must | `CAP-AUD-03` | Неизменяемость AuditEvent | `ANALYZED` |
| `AUD-SEC-001` | security | Must | `CAP-AUD-09` | Разделение полномочий Audit | `ANALYZED` |
| `AUD-NFR-001` | non-functional | Must | `CAP-AUD-01` | Надёжность приёма AuditEvent | `ANALYZED` |

## Детальные требования

### AUD-FR-001. Принять AuditEvent {#aud-fr-001}

| Поле | Значение |
| --- | --- |
| Тип | `functional` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Audit` / `m8-audit` |
| Business capability | `CAP-AUD-01` |
| Согласованность | C0 — локальная строгая согласованность |
| Зависимости | All producer services, Access, Object/archive storage |
| Данные | AuditEvent |
| Безопасность | audit read permission, tamper resistance |
| API | — |
| События | — |
| Атрибуты качества | — |
| Основание PADS | §6.14, §7, §9.8, §11, §15 |

**Требование.**

Принять нормативный AuditEvent от аутентифицированного источника и подтвердить его фиксацию.

**Критерии приёмки.**

- `AUD-FR-001-AC-01` — Событие валидируется до хранения.
- `AUD-FR-001-AC-02` — Повтор event_id не создаёт вторую запись.

**Трассировка для следующего этапа:**

```yaml
requirement_id: AUD-FR-001
capability: CAP-AUD-01
owner_context: Audit
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

### AUD-FR-002. Проверить происхождение AuditEvent {#aud-fr-002}

| Поле | Значение |
| --- | --- |
| Тип | `functional` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Audit` / `m8-audit` |
| Business capability | `CAP-AUD-02` |
| Согласованность | C0 — локальная строгая согласованность |
| Зависимости | All producer services, Access, Object/archive storage |
| Данные | AuditEvent |
| Безопасность | audit read permission, tamper resistance |
| API | — |
| События | — |
| Атрибуты качества | — |
| Основание PADS | §6.14, §7, §9.8, §11, §15 |

**Требование.**

Проверить producer identity, schema version, event time и обязательные context fields.

**Критерии приёмки.**

- `AUD-FR-002-AC-01` — Непроверенный producer отклоняется.
- `AUD-FR-002-AC-02` — Clock skew обрабатывается и помечается.

**Трассировка для следующего этапа:**

```yaml
requirement_id: AUD-FR-002
capability: CAP-AUD-02
owner_context: Audit
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

### AUD-FR-003. Проверить минимизацию данных {#aud-fr-003}

| Поле | Значение |
| --- | --- |
| Тип | `functional` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Audit` / `m8-audit` |
| Business capability | `CAP-AUD-10` |
| Согласованность | C0 — локальная строгая согласованность |
| Зависимости | All producer services, Access, Object/archive storage |
| Данные | AuditEvent |
| Безопасность | audit read permission, tamper resistance |
| API | — |
| События | — |
| Атрибуты качества | — |
| Основание PADS | §6.14, §7, §9.8, §11, §15 |

**Требование.**

Отклонить или редактировать AuditEvent, содержащий запрещённые секреты и поля.

**Критерии приёмки.**

- `AUD-FR-003-AC-01` — Token/OTP/private key patterns блокируются.
- `AUD-FR-003-AC-02` — Редактирование не меняет смысл outcome.

**Трассировка для следующего этапа:**

```yaml
requirement_id: AUD-FR-003
capability: CAP-AUD-10
owner_context: Audit
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

### AUD-FR-004. Сохранить AuditEvent неизменяемо {#aud-fr-004}

| Поле | Значение |
| --- | --- |
| Тип | `functional` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Audit` / `m8-audit` |
| Business capability | `CAP-AUD-03` |
| Согласованность | C0 — локальная строгая согласованность |
| Зависимости | All producer services, Access, Object/archive storage |
| Данные | AuditEvent |
| Безопасность | audit read permission, tamper resistance |
| API | — |
| События | — |
| Атрибуты качества | — |
| Основание PADS | §6.14, §7, §9.8, §11, §15 |

**Требование.**

Сохранить событие без возможности неаудируемого update или delete.

**Критерии приёмки.**

- `AUD-FR-004-AC-01` — Исправление создаёт отдельную correction record.
- `AUD-FR-004-AC-02` — Операционная учетная запись не имеет прямого update доступа.

**Трассировка для следующего этапа:**

```yaml
requirement_id: AUD-FR-004
capability: CAP-AUD-03
owner_context: Audit
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

### AUD-FR-005. Подтвердить доставку обязательного события {#aud-fr-005}

| Поле | Значение |
| --- | --- |
| Тип | `functional` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Audit` / `m8-audit` |
| Business capability | `CAP-AUD-11` |
| Согласованность | C0 — локальная строгая согласованность |
| Зависимости | All producer services, Access, Object/archive storage |
| Данные | AuditEvent |
| Безопасность | audit read permission, tamper resistance |
| API | — |
| События | — |
| Атрибуты качества | — |
| Основание PADS | §6.14, §7, §9.8, §11, §15 |

**Требование.**

Отслеживать gaps и подтверждение доставки от критических producers.

**Критерии приёмки.**

- `AUD-FR-005-AC-01` — Пропуск sequence/reconciliation window создаёт alert.
- `AUD-FR-005-AC-02` — Producer может безопасно повторить событие.

**Трассировка для следующего этапа:**

```yaml
requirement_id: AUD-FR-005
capability: CAP-AUD-11
owner_context: Audit
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

### AUD-FR-010. Искать AuditEvent {#aud-fr-010}

| Поле | Значение |
| --- | --- |
| Тип | `functional` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Audit` / `m8-audit` |
| Business capability | `CAP-AUD-05` |
| Согласованность | C0 — локальная строгая согласованность |
| Зависимости | All producer services, Access, Object/archive storage |
| Данные | AuditEvent |
| Безопасность | audit read permission, tamper resistance |
| API | — |
| События | — |
| Атрибуты качества | — |
| Основание PADS | §6.14, §7, §9.8, §11, §15 |

**Требование.**

Искать события по времени, actor, subject, target, action, outcome и scope.

**Критерии приёмки.**

- `AUD-FR-010-AC-01` — Поиск ограничен Access permission и privacy rules.
- `AUD-FR-010-AC-02` — Пагинация стабильна для snapshot/query window.

**Трассировка для следующего этапа:**

```yaml
requirement_id: AUD-FR-010
capability: CAP-AUD-05
owner_context: Audit
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

### AUD-FR-011. Получить событие по ID {#aud-fr-011}

| Поле | Значение |
| --- | --- |
| Тип | `functional` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Audit` / `m8-audit` |
| Business capability | `CAP-AUD-05` |
| Согласованность | C0 — локальная строгая согласованность |
| Зависимости | All producer services, Access, Object/archive storage |
| Данные | AuditEvent |
| Безопасность | audit read permission, tamper resistance |
| API | — |
| События | — |
| Атрибуты качества | — |
| Основание PADS | §6.14, §7, §9.8, §11, §15 |

**Требование.**

Получить одно событие и integrity metadata в разрешённой области.

**Критерии приёмки.**

- `AUD-FR-011-AC-01` — Недоступное событие не раскрывается.
- `AUD-FR-011-AC-02` — Ответ содержит schema version и ingestion time.

**Трассировка для следующего этапа:**

```yaml
requirement_id: AUD-FR-011
capability: CAP-AUD-05
owner_context: Audit
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

### AUD-FR-012. Построить цепочку по correlation {#aud-fr-012}

| Поле | Значение |
| --- | --- |
| Тип | `functional` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Audit` / `m8-audit` |
| Business capability | `CAP-AUD-06` |
| Согласованность | C0 — локальная строгая согласованность |
| Зависимости | All producer services, Access, Object/archive storage |
| Данные | AuditEvent |
| Безопасность | audit read permission, tamper resistance |
| API | — |
| События | — |
| Атрибуты качества | — |
| Основание PADS | §6.14, §7, §9.8, §11, §15 |

**Требование.**

Получить причинно связанную цепочку действий по correlation_id/causation_id/operation_id.

**Критерии приёмки.**

- `AUD-FR-012-AC-01` — Разрывы обозначаются явно.
- `AUD-FR-012-AC-02` — Порядок использует event time и ingestion metadata.

**Трассировка для следующего этапа:**

```yaml
requirement_id: AUD-FR-012
capability: CAP-AUD-06
owner_context: Audit
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

### AUD-FR-013. Получить историю ресурса {#aud-fr-013}

| Поле | Значение |
| --- | --- |
| Тип | `functional` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Audit` / `m8-audit` |
| Business capability | `CAP-AUD-06` |
| Согласованность | C0 — локальная строгая согласованность |
| Зависимости | All producer services, Access, Object/archive storage |
| Данные | AuditEvent |
| Безопасность | audit read permission, tamper resistance |
| API | — |
| События | — |
| Атрибуты качества | — |
| Основание PADS | §6.14, §7, §9.8, §11, §15 |

**Требование.**

Получить последовательность значимых действий для ResourceReference.

**Критерии приёмки.**

- `AUD-FR-013-AC-01` — Resource tombstone сохраняет возможность поиска.
- `AUD-FR-013-AC-02` — Личные данные минимизированы по current disclosure policy.

**Трассировка для следующего этапа:**

```yaml
requirement_id: AUD-FR-013
capability: CAP-AUD-06
owner_context: Audit
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

### AUD-FR-020. Создать ExportJob {#aud-fr-020}

| Поле | Значение |
| --- | --- |
| Тип | `functional` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Audit` / `m8-audit` |
| Business capability | `CAP-AUD-07` |
| Согласованность | C0 — локальная строгая согласованность |
| Зависимости | All producer services, Access, Object/archive storage |
| Данные | AuditEvent |
| Безопасность | audit read permission, tamper resistance |
| API | — |
| События | — |
| Атрибуты качества | — |
| Основание PADS | §6.14, §7, §9.8, §11, §15 |

**Требование.**

Создать длительную Operation формирования контролируемой выгрузки.

**Критерии приёмки.**

- `AUD-FR-020-AC-01` — Export scope фиксируется на старте.
- `AUD-FR-020-AC-02` — Массовый экспорт требует специального permission и может требовать Risk Decision.

**Трассировка для следующего этапа:**

```yaml
requirement_id: AUD-FR-020
capability: CAP-AUD-07
owner_context: Audit
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

### AUD-FR-021. Получить экспорт {#aud-fr-021}

| Поле | Значение |
| --- | --- |
| Тип | `functional` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Audit` / `m8-audit` |
| Business capability | `CAP-AUD-07` |
| Согласованность | C0 — локальная строгая согласованность |
| Зависимости | All producer services, Access, Object/archive storage |
| Данные | AuditEvent |
| Безопасность | audit read permission, tamper resistance |
| API | — |
| События | — |
| Атрибуты качества | — |
| Основание PADS | §6.14, §7, §9.8, §11, §15 |

**Требование.**

Предоставить временную защищённую ссылку или поток после завершения ExportJob.

**Критерии приёмки.**

- `AUD-FR-021-AC-01` — Ссылка имеет TTL и привязку к actor/client.
- `AUD-FR-021-AC-02` — Получение экспорта также аудируется.

**Трассировка для следующего этапа:**

```yaml
requirement_id: AUD-FR-021
capability: CAP-AUD-07
owner_context: Audit
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

### AUD-FR-030. Применить Retention Policy {#aud-fr-030}

| Поле | Значение |
| --- | --- |
| Тип | `functional` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Audit` / `m8-audit` |
| Business capability | `CAP-AUD-08` |
| Согласованность | C0 — локальная строгая согласованность |
| Зависимости | All producer services, Access, Object/archive storage |
| Данные | AuditEvent |
| Безопасность | audit read permission, tamper resistance |
| API | — |
| События | — |
| Атрибуты качества | — |
| Основание PADS | §6.14, §7, §9.8, §11, §15 |

**Требование.**

Хранить и уничтожать события по типу, scope, classification и legal requirements.

**Критерии приёмки.**

- `AUD-FR-030-AC-01` — Уничтожение необратимо и само аудируется агрегированным evidence.
- `AUD-FR-030-AC-02` — Policy version фиксируется для каждого decision.

**Трассировка для следующего этапа:**

```yaml
requirement_id: AUD-FR-030
capability: CAP-AUD-08
owner_context: Audit
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

### AUD-FR-031. Установить Legal Hold {#aud-fr-031}

| Поле | Значение |
| --- | --- |
| Тип | `functional` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Audit` / `m8-audit` |
| Business capability | `CAP-AUD-08` |
| Согласованность | C0 — локальная строгая согласованность |
| Зависимости | All producer services, Access, Object/archive storage |
| Данные | AuditEvent |
| Безопасность | audit read permission, tamper resistance |
| API | — |
| События | — |
| Атрибуты качества | — |
| Основание PADS | §6.14, §7, §9.8, §11, §15 |

**Требование.**

Приостановить уничтожение выбранного набора событий по правомерному основанию.

**Критерии приёмки.**

- `AUD-FR-031-AC-01` — Hold имеет owner, reason, scope и expiry/review.
- `AUD-FR-031-AC-02` — Снятие hold требует отдельного permission.

**Трассировка для следующего этапа:**

```yaml
requirement_id: AUD-FR-031
capability: CAP-AUD-08
owner_context: Audit
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

### AUD-FR-040. Проверить целостность хранения {#aud-fr-040}

| Поле | Значение |
| --- | --- |
| Тип | `functional` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Audit` / `m8-audit` |
| Business capability | `CAP-AUD-04` |
| Согласованность | C0 — локальная строгая согласованность |
| Зависимости | All producer services, Access, Object/archive storage |
| Данные | AuditEvent |
| Безопасность | audit read permission, tamper resistance |
| API | — |
| События | — |
| Атрибуты качества | — |
| Основание PADS | §6.14, §7, §9.8, §11, §15 |

**Требование.**

Периодически проверять hash/chain/segment integrity и выявлять пропуски или подмену.

**Критерии приёмки.**

- `AUD-FR-040-AC-01` — Нарушение создаёт security incident.
- `AUD-FR-040-AC-02` — Проверка не изменяет исходные записи.

**Трассировка для следующего этапа:**

```yaml
requirement_id: AUD-FR-040
capability: CAP-AUD-04
owner_context: Audit
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

### AUD-FR-041. Предоставить integrity proof {#aud-fr-041}

| Поле | Значение |
| --- | --- |
| Тип | `functional` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Audit` / `m8-audit` |
| Business capability | `CAP-AUD-04` |
| Согласованность | C0 — локальная строгая согласованность |
| Зависимости | All producer services, Access, Object/archive storage |
| Данные | AuditEvent |
| Безопасность | audit read permission, tamper resistance |
| API | — |
| События | — |
| Атрибуты качества | — |
| Основание PADS | §6.14, §7, §9.8, §11, §15 |

**Требование.**

Сформировать проверяемое доказательство целостности выбранного набора или сегмента.

**Критерии приёмки.**

- `AUD-FR-041-AC-01` — Proof не раскрывает недоступные события.
- `AUD-FR-041-AC-02` — Алгоритм и версия указаны.

**Трассировка для следующего этапа:**

```yaml
requirement_id: AUD-FR-041
capability: CAP-AUD-04
owner_context: Audit
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

### AUD-FR-050. Аудировать действия над Audit {#aud-fr-050}

| Поле | Значение |
| --- | --- |
| Тип | `functional` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Audit` / `m8-audit` |
| Business capability | `CAP-AUD-12` |
| Согласованность | C0 — локальная строгая согласованность |
| Зависимости | All producer services, Access, Object/archive storage |
| Данные | AuditEvent |
| Безопасность | audit read permission, tamper resistance |
| API | — |
| События | — |
| Атрибуты качества | — |
| Основание PADS | §6.14, §7, §9.8, §11, §15 |

**Требование.**

Фиксировать поиск, просмотр чувствительных событий, экспорт, policy и hold changes.

**Критерии приёмки.**

- `AUD-FR-050-AC-01` — Audit-on-audit хранится с теми же гарантиями.
- `AUD-FR-050-AC-02` — Рекурсивная запись не создаёт бесконечный цикл.

**Трассировка для следующего этапа:**

```yaml
requirement_id: AUD-FR-050
capability: CAP-AUD-12
owner_context: Audit
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

### AUD-DATA-001. Неизменяемость AuditEvent {#aud-data-001}

| Поле | Значение |
| --- | --- |
| Тип | `data` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Audit` / `m8-audit` |
| Business capability | `CAP-AUD-03` |
| Согласованность | C0 — локальная строгая согласованность |
| Зависимости | — |
| Данные | — |
| Безопасность | — |
| API | — |
| События | — |
| Атрибуты качества | — |
| Основание PADS | §11 |

**Требование.**

Сохранённое аудиторское событие не должно изменяться; дополнительный контекст оформляется новой связанной записью.

**Критерии приёмки.**

- `AUD-DATA-001-AC-01` — Нет публичного UpdateAuditEvent.
- `AUD-DATA-001-AC-02` — Storage policy предотвращает неаудируемое изменение и раннее удаление.

**Трассировка для следующего этапа:**

```yaml
requirement_id: AUD-DATA-001
capability: CAP-AUD-03
owner_context: Audit
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

### AUD-SEC-001. Разделение полномочий Audit {#aud-sec-001}

| Поле | Значение |
| --- | --- |
| Тип | `security` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Audit` / `m8-audit` |
| Business capability | `CAP-AUD-09` |
| Согласованность | C0 — локальная строгая согласованность |
| Зависимости | — |
| Данные | — |
| Безопасность | — |
| API | — |
| События | — |
| Атрибуты качества | — |
| Основание PADS | §15 |

**Требование.**

Права просмотра, массового экспорта, retention и legal hold должны быть разделены.

**Критерии приёмки.**

- `AUD-SEC-001-AC-01` — Одна базовая роль не получает автоматически все полномочия.
- `AUD-SEC-001-AC-02` — Критические administrative actions требуют step-up по policy.

**Трассировка для следующего этапа:**

```yaml
requirement_id: AUD-SEC-001
capability: CAP-AUD-09
owner_context: Audit
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

### AUD-NFR-001. Надёжность приёма AuditEvent {#aud-nfr-001}

| Поле | Значение |
| --- | --- |
| Тип | `non-functional` |
| Статус | `ANALYZED` |
| Приоритет | `Must` |
| Владелец | `Audit` / `m8-audit` |
| Business capability | `CAP-AUD-01` |
| Согласованность | C0 — локальная строгая согласованность |
| Зависимости | — |
| Данные | — |
| Безопасность | — |
| API | — |
| События | — |
| Атрибуты качества | availability>=99.99%, ack after durable commit |
| Основание PADS | §19 |

**Требование.**

Подтверждённый AuditEvent не должен теряться; monthly availability ingest API не ниже 99,99%.

**Критерии приёмки.**

- `AUD-NFR-001-AC-01` — Ack выдаётся только после durable commit.
- `AUD-NFR-001-AC-02` — Producer retry безопасен по event_id.

**Трассировка для следующего этапа:**

```yaml
requirement_id: AUD-NFR-001
capability: CAP-AUD-01
owner_context: Audit
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
