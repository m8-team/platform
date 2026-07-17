---
title: "M8 Platform Data Ownership Registry"
description: "Реестр владения данными M8 Platform."
keywords:
  - "M8 Platform"
  - "data ownership"
---

# M8 Platform Data Ownership Registry {#data-ownership-registry}

{% note info "Навигация" %}

[Engineering artifacts](../index.md) | [PADS: владение данными](../../architecture/pads/platform/11-data-ownership.md) | `data-ownership.yaml`

{% endnote %}

_M8-DATA-000 · Версия 0.1 · 10 июля 2026 года_

| Поле | Значение |
| --- | --- |
| Идентификатор | `M8-DATA-000` |
| Версия | `0.1` |
| Статус | Базовая нормативная редакция |
| Владелец | Sergey Gorbachev |
| Нормативная основа | `PADS-000@1.0`, `M8-REQ-000@0.1` |
| Область | Источники истины, классификация, retention, deletion и projections |

# 1. Правила владения

- У каждой сущности ровно один authoritative owner.
- Проекция не может принимать предметные решения за владельца.
- Репликация персональных/секретных данных требует purpose, freshness, retention и deletion propagation.
- Прямой доступ к таблицам другого контекста запрещён.
- Исторические ссылки сохраняются как минимальные typed references/tombstones.

# 2. Реестр

| ID | Сущность | Владелец | Store | Класс | Retention | Удаление | Проекции |
| --- | --- | --- | --- | --- | --- | --- | --- |
| `DATA-RM-001` | Organization | m8-resource-manager | YDB | internal | active + 7y tombstone | archive then tombstone | Access, Audit, Provisioning |
| `DATA-RM-002` | Workspace | m8-resource-manager | YDB | internal | active + 7y tombstone | archive then tombstone | Access, Audit |
| `DATA-RM-003` | Project | m8-resource-manager | YDB | internal | active + 7y tombstone | managed deletion workflow | Access, Audit, Provisioning |
| `DATA-RM-004` | ServiceRegistration | m8-resource-manager | YDB | internal | active + 3y | soft delete | Access, Audit |
| `DATA-ID-001` | UserPool | m8-identity | YDB | internal | active + 7y tombstone | suspend/delete by workflow | Authentication, Access, Audit |
| `DATA-ID-002` | User | m8-identity | YDB | personal | purpose + legal retention | anonymize/delete by privacy workflow | Authentication, Access, Audit |
| `DATA-ID-003` | ProfileAttribute | m8-identity | YDB | personal/sensitive | attribute policy | delete or anonymize | none by default |
| `DATA-ID-004` | ExternalIdentity | m8-identity | YDB | personal | active + audit reference | unlink/tombstone | Authentication |
| `DATA-ID-005` | Group | m8-identity | YDB | internal | active + 3y | soft delete | Access |
| `DATA-ID-006` | Membership | m8-identity | YDB | internal | active + 3y | revoke/tombstone | Access, Audit |
| `DATA-AUTH-001` | Client | m8-authentication | YDB | confidential | active + 3y | disable/tombstone | Access, Audit |
| `DATA-AUTH-002` | AuthenticationTransaction | m8-authentication | YDB | sensitive | 90d default | expire then purge/minimize | Audit, Risk projection |
| `DATA-AUTH-003` | AuthenticationChallenge | m8-authentication | YDB/Redis | secret-adjacent | TTL + 30d metadata | secret purge; metadata retain | none |
| `DATA-AUTH-004` | AuthenticationHandoff | m8-authentication | Redis/YDB | secret | minutes | one-time consume then purge | none |
| `DATA-AUTH-005` | AuthenticationSessionReference | m8-authentication | YDB | sensitive | session + 90d | revoke then purge | Audit |
| `DATA-AUTH-006` | WebAuthnCredentialReference | m8-authentication | YDB | sensitive | until revoke + policy | revoke/tombstone | Identity reference only |
| `DATA-ACC-001` | PermissionDefinition | m8-access | YDB | internal | version lifetime + 7y | retire, never reuse name | all services read cache |
| `DATA-ACC-002` | Role | m8-access | YDB | internal | active + 7y | soft delete | Audit |
| `DATA-ACC-003` | RoleBinding | m8-access | YDB/SpiceDB | sensitive authorization | active + 7y | revoke/tombstone | Audit, service caches |
| `DATA-ACC-004` | AccessRelationship | m8-access | YDB/SpiceDB | sensitive authorization | active + 7y | delete/tombstone | service caches |
| `DATA-ACC-005` | AuthorizationModel | m8-access | YDB/SpiceDB | internal | all published versions | retire version | all services |
| `DATA-RISK-001` | RiskAssessment | m8-risk-decision | YDB | sensitive | 180d default | expire then minimize | Audit |
| `DATA-RISK-002` | RiskPolicy | m8-risk-decision | YDB | confidential | all published versions | retire version | runtime cache |
| `DATA-RISK-003` | RiskSignal | m8-risk-decision | YDB/stream | sensitive | 30-180d by type | purge/minimize | none |
| `DATA-RISK-004` | ManualReview | m8-risk-decision | YDB | sensitive | 7y | close then retain | Audit |
| `DATA-PROV-001` | ResourceDefinition | m8-provisioning | YDB | internal | all published versions | retire version | drivers/runtime cache |
| `DATA-PROV-002` | ManagedResource | m8-provisioning | YDB | internal/confidential | lifetime + 7y | managed deletion/tombstone | Resource Manager, Audit |
| `DATA-PROV-003` | DesiredState | m8-provisioning | YDB | confidential | current + revision history | redact secret refs | drivers |
| `DATA-PROV-004` | ObservedState | m8-provisioning | YDB | confidential | rolling 90d + snapshots | purge after retention | monitoring projection |
| `DATA-PROV-005` | Placement | m8-provisioning | YDB | internal | lifetime + 3y | release/tombstone | Audit |
| `DATA-PROV-006` | Driver | m8-provisioning | YDB | confidential | active + 3y | disable/tombstone | runtime |
| `DATA-AUD-001` | AuditEvent | m8-audit | immutable store | sensitive audit | policy 1-10y | retention/legal hold only | search index |
| `DATA-AUD-002` | AuditExport | m8-audit | object storage | sensitive export | short TTL 1-30d | secure delete | none |
| `DATA-AUD-003` | IntegrityProof | m8-audit | immutable store | internal | same as covered events | retention policy | none |
| `DATA-OPS-001` | Operation | operation owner service | YDB | internal | 90d-7y by operation type | delete record only after policy | read API, Audit |
| `DATA-COMMON-001` | OutboxRecord | event publisher | YDB | internal | until published + replay window | purge after evidence | broker |
| `DATA-COMMON-002` | InboxRecord | event consumer | YDB | internal | dedup window | purge after window | none |
| `DATA-COMMON-003` | IdempotencyRecord | command owner | YDB/Redis | internal | documented idempotency window | purge after window | none |

# 3. Перенос данных

Перенос ownership выполняется через ADR, dual-read/dual-write только на ограниченное время, backfill с revision guards, сверку контрольных сумм, переключение источника истины и удаление устаревшей копии.
