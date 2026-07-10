---
title: "PADS: архитектура безопасности"
description: "Trust boundaries, authentication, authorization, audit, secrets, threat modeling и Secure SDLC."
keywords:
  - "M8 Platform"
  - "PADS"
  - "architecture"
---

# 15. Архитектура безопасности {#pads-security}

{% note info "Навигация PADS" %}

[Оглавление PADS](../index.md) | [Предыдущий раздел: 14. Модель интеграции и согласованности](14-integration-consistency.md) | [Следующий раздел: Long-Running Operations (LRO)](16-operations.md)

{% endnote %}

## 15.1. Назначение главы

Настоящая глава определяет базовую модель безопасности M8 Platform: границы доверия, идентичности участников, аутентификацию, авторизацию, оценку риска, изоляцию проектов, защиту данных, service-to-service взаимодействия, управление секретами, аудит и требования безопасной разработки.

Безопасность является предметной и архитектурной функцией, а не только инфраструктурным слоем. Нахождение запроса во внутренней сети не означает доверия.

## 15.2. Цели безопасности

| ID | Цель |
| --- | --- |
| `SEC-G-001` | Каждый запрос имеет проверенную идентичность вызывающей стороны. |
| `SEC-G-002` | Каждое чувствительное действие имеет явное решение Access. |
| `SEC-G-003` | Риск действия оценивается независимо от статического права, когда это требуется политикой. |
| `SEC-G-004` | Данные и ресурсы одного Project не раскрываются другому Project без явной связи. |
| `SEC-G-005` | Секреты не попадают в логи, события, audit payload и Structured Prompts. |
| `SEC-G-006` | Все security-sensitive изменения доказуемы через Audit. |
| `SEC-G-007` | Компрометация одного сервиса имеет ограниченный blast radius. |
| `SEC-G-008` | Security controls проверяются автоматически и регулярно. |

## 15.3. Нормативные принципы

| ID | Правило |
| --- | --- |
| `SEC-001` | M8 MUST применять Zero Trust: каждое взаимодействие аутентифицируется и авторизуется независимо от сетевого расположения. |
| `SEC-002` | Actor, Subject, Client и Service Identity MUST различаться. |
| `SEC-003` | Сервис MUST доверять только проверенному security context. |
| `SEC-004` | Authorization MUST выполняться до предметной mutation. |
| `SEC-005` | Risk Decision MUST NOT заменять Access Decision, а Access MUST NOT заменять Authentication. |
| `SEC-006` | High-risk action MUST поддерживать step-up до требуемого assurance level. |
| `SEC-007` | Credentials и private keys MUST храниться в специализированном secret store или identity provider. |
| `SEC-008` | Tokens MUST иметь минимальные scope, audience и lifetime. |
| `SEC-009` | Service-to-service credentials MUST быть ротируемыми и не общими для нескольких независимых сервисов. |
| `SEC-010` | Все authorization decisions SHOULD иметь decision ID и reason code. |
| `SEC-011` | Отказ security dependency для mutation по умолчанию ведёт к fail-closed. |
| `SEC-012` | Security-sensitive endpoints MUST иметь rate limiting и abuse detection. |
| `SEC-013` | Data classification MUST управлять encryption, logging и access policy. |
| `SEC-014` | Все внешние callbacks/webhooks MUST проверять подпись и replay protection. |
| `SEC-015` | Tenant/project scope MUST извлекаться из авторитетного контекста и проверяться с resource scope. |
| `SEC-016` | Audit record MUST создаваться для успешных и отклонённых критичных действий. |
| `SEC-017` | Security policy changes MUST version, approve, audit and support rollback. |
| `SEC-018` | Продуктивные данные MUST быть отделены от test/development environments. |
| `SEC-019` | Dependency и container vulnerabilities MUST иметь управляемый remediation SLA. |
| `SEC-020` | Structured Prompt MUST содержать security requirements и MUST NOT включать реальные секреты. |

## 15.4. Участники безопасности

| Понятие | Смысл |
| --- | --- |
| Actor | тот, кто инициировал действие: пользователь, сервис, администратор, automation |
| Subject | объект идентичности, права или аутентификации которого рассматриваются |
| Client | приложение или интеграция, инициирующая authentication flow/API call |
| Service Identity | техническая идентичность workload |
| Resource | объект, над которым выполняется действие |
| Session | ограниченный во времени security context |
| Credential | доказательство, используемое для аутентификации |
| Assurance Level | достигнутый уровень уверенности в идентичности |

Actor и Subject MAY совпадать, но MUST храниться раздельно. Администратор, изменяющий пользователя, является Actor, а пользователь — Subject/Target.

## 15.5. Границы доверия

Основные trust boundaries:

1. пользовательское устройство ↔ edge/API gateway;
2. внешний client ↔ Authentication/AuthGuard;
3. сервис ↔ сервис;
4. M8 ↔ Keycloak/identity providers;
5. M8 ↔ SpiceDB;
6. M8 ↔ Temporal;
7. Provisioning ↔ Kubernetes/cloud providers;
8. operational plane ↔ audit/observability;
9. production ↔ non-production;
10. organization/project ↔ другой tenant scope.

Для каждой границы MUST быть определены authentication, encryption, allowed protocols, input validation, timeout, logging policy и incident owner.

## 15.6. Модель аутентификации

Authentication владеет:

- AuthenticationTransaction;
- challenge selection/execution;
- achieved assurance;
- session/handoff lifecycle;
- provider adapters;
- re-authentication и step-up.

Identity подтверждает существование и состояние Subject. Keycloak является внешним identity/authentication engine за ACL и не определяет язык M8.

## 15.7. Базовый flow

```text
Client
→ StartAuthentication
→ resolve Subject through Identity
→ evaluate Risk
→ choose Challenge
→ execute provider interaction
→ verify result
→ set achieved assurance
→ create handoff/session result
→ Audit
```

Primary flow MAY использовать CIBA. Refresh выполняется через Keycloak refresh token при наличии. Если refresh недействителен, начинается новая AuthenticationTransaction; предыдущая session не восстанавливается скрыто.

## 15.8. Challenge model

Поддерживаемые классы:

- `OTP`;
- `APPROVAL`;
- `MOBILE_ID`;
- `WEBAUTHN`;
- `OIDC`;
- `SAML`;
- `PASSWORD`, если разрешено policy;
- `RECOVERY`, только через отдельную усиленную policy.

Каждый challenge MUST иметь:

- уникальный ID;
- ограниченный lifetime;
- attempt limit;
- resend/rate policy;
- provider reference;
- state machine;
- achieved assurance contribution;
- audit events.

## 15.9. Assurance level

Уровни assurance MUST быть упорядочены и определены policy, например:

| Уровень | Пример требований |
| --- | --- |
| `AAL0` | идентичность не подтверждена |
| `AAL1` | single-factor low assurance |
| `AAL2` | подтверждённое устройство или второй фактор |
| `AAL3` | phishing-resistant strong authentication |

Точное соответствие methods уровням задаётся versioned policy. Client не может сам объявить достигнутый уровень.

## 15.10. Step-up

Step-up требуется, когда:

- Risk Decision возвращает `CHALLENGE`;
- действие требует AAL выше session AAL;
- session age превышает policy;
- изменился security-sensitive context;
- выполняется privileged action.

Step-up создаёт новую AuthenticationTransaction, связанную с исходным actor/session/action. После успеха token/session context обновляется только согласно policy.

## 15.11. Session и token model

Tokens MUST:

- иметь issuer и audience;
- иметь короткий access lifetime;
- поддерживать revocation/rotation where applicable;
- не содержать избыточные profile data;
- иметь token ID при необходимости расследования;
- быть проверены по signature, expiry, audience, issuer и required claims.

Refresh token является restricted data и MUST NOT попадать в domain events, logs или audit payload.

## 15.12. AuthGuard

AuthGuard/BFF отвечает за:

- извлечение credentials;
- проверку token/session;
- построение trusted RequestContext;
- audience/scope validation;
- propagation service identity;
- вызов Access/Risk при заданной policy;
- защиту от confused deputy;
- correlation и audit context.

AuthGuard MUST NOT владеть User, Role или RiskPolicy.

## 15.13. Авторизация

Access владеет:

- permission vocabulary;
- role и role binding;
- relationship model;
- resource/subject relations;
- check, explain, simulate;
- revision и consistency token решения.

SpiceDB является evaluation engine. Сервисы используют Access API и MUST NOT формировать tuple strings самостоятельно.

## 15.14. Permission model

Permission SHOULD иметь форму:

```text
<resource_type>.<action>
```

Примеры:

```text
project.get
project.update
service.register
identity.user.disable
authentication.policy.update
access.role.bind
provisioning.resource.create
audit.event.export
```

Каждый mutating API MUST объявлять permission. Scope решения MUST соответствовать целевому ResourceReference.

## 15.15. Relationship-based access

Access MAY вычислять права через отношения:

```text
user → member → project
user → admin → organization
service → owner → managed_resource
```

Отношения MUST иметь owner, lifecycle и audit. Удаление subject/resource MUST приводить к revocation/tombstone обработке.

## 15.16. Explain и simulate

Explain API SHOULD быть доступен только авторизованным администраторам и возвращать:

- decision;
- evaluated model revision;
- matched relationship path;
- missing relation/permission;
- decision ID.

Он MUST не раскрывать restricted information другого tenant.

## 15.17. Risk Decision

Risk Decision принимает контекст и возвращает:

- `ALLOW`;
- `DENY`;
- `CHALLENGE`;
- `REVIEW`.

Ответ SHOULD включать:

- decision ID;
- risk level/score, если разрешено;
- reason codes;
- required assurance/challenge class;
- policy version;
- expiry/freshness.

Risk signals MUST быть минимизированы и классифицированы.

## 15.18. Связь Access и Risk

Типовой порядок:

1. Authentication establishes identity/assurance;
2. Access checks static/contextual permission;
3. Risk evaluates dynamic context;
4. Authentication executes step-up if needed;
5. Access MAY re-check after assurance change;
6. use case commits;
7. Audit records decisions.

Порядок MAY отличаться для оптимизации, но fail-open без явной policy запрещён.

## 15.19. Service-to-service security

Service identity SHOULD обеспечиваться workload identity/mTLS или подписанным short-lived token.

Требования:

- уникальная identity per service/environment;
- explicit audience;
- least privilege;
- rotation;
- no static shared passwords;
- service authorization;
- propagation original actor separately;
- prevention of actor impersonation.

## 15.20. Delegation

При вызове сервиса от имени пользователя MUST различаться:

- authenticated service identity;
- delegated actor/subject;
- delegation scope;
- originating client;
- assurance;
- chain depth.

Сервис не может подменять actor без разрешённого delegation contract.

## 15.21. Tenant и Project isolation

Каждый запрос MUST определить scope. Контроль включает:

- resource name scope validation;
- Access relation;
- database partition/key scope;
- cache key scope;
- event partition/ACL;
- log/audit scope;
- export scope;
- metrics labels без утечки sensitive tenant data.

Нельзя полагаться только на фильтр UI.

## 15.22. Защита данных

Данные MUST защищаться:

- TLS in transit;
- encryption at rest;
- key rotation;
- data classification;
- field-level protection where required;
- backup encryption;
- access logging;
- secure deletion.

Restricted fields SHOULD быть tokenized или encrypted отдельно при наличии угрозы массового раскрытия.

## 15.23. Управление ключами

Encryption/signing keys MUST:

- иметь owner и purpose;
- храниться в KMS/HSM или утверждённом secret store;
- иметь rotation schedule;
- поддерживать versioned key IDs;
- не экспортироваться в application logs/config;
- иметь break-glass procedure;
- быть разделены по environment и purpose.

## 15.24. Secret management

Секреты MUST:

- внедряться runtime-механизмом;
- не храниться в Git, image или prompt;
- иметь short access path;
- ротироваться без полной остановки;
- иметь usage audit;
- маскироваться в diagnostics;
- не передаваться через обычные events.

## 15.25. Входные данные и защита интерфейсов

Все внешние входы MUST проходить:

- size limit;
- schema validation;
- canonicalization;
- content/type validation;
- injection protection;
- path/resource name validation;
- decompression limits;
- request timeout;
- rate limit.

Свободные expressions, filters, templates и policies требуют sandbox/allowlist.

## 15.26. Webhook/callback security

Обязательны:

- signature verification;
- key ID;
- timestamp tolerance;
- nonce/event ID deduplication;
- endpoint allowlist when applicable;
- TLS;
- body size limit;
- no trust in source IP alone;
- audit of failed verification.

## 15.27. Audit requirements

Security Audit SHOULD фиксировать:

- actor/service identity;
- subject;
- action;
- resource;
- access decision ID;
- risk decision ID;
- assurance level;
- client;
- network/device digest, если разрешено;
- outcome/error code;
- correlation/trace;
- policy revisions.

Секреты и полные tokens не записываются.

## 15.28. Logging security

Запрещено логировать:

- passwords;
- OTP values;
- refresh/access tokens;
- private keys;
- raw authorization headers;
- full cookies;
- unrestricted personal profiles;
- sensitive risk signals без redaction.

Logging libraries MUST поддерживать structured redaction.

## 15.29. Abuse prevention

Authentication, Access, Audit Export и Provisioning MUST иметь:

- rate limits;
- velocity rules;
- attempt limits;
- anomaly metrics;
- lockout/cooldown policy;
- anti-enumeration responses;
- alerting;
- recovery path.

## 15.30. Threat modeling

Каждая значимая capability MUST иметь threat model с:

- assets;
- actors;
- trust boundaries;
- entry points;
- threats;
- mitigations;
- residual risk;
- verification tests.

Threat model пересматривается при изменении authentication flow, public API, data classification или external integration.

## 15.31. Secure SDLC

Release gate SHOULD включать:

- SAST;
- dependency scanning;
- secret scanning;
- container/IaC scanning;
- API fuzz/property tests;
- authorization tests;
- threat model check;
- SBOM;
- signed artifacts;
- provenance/attestation;
- protected branch/review.

## 15.32. Vulnerability management

Vulnerability MUST иметь severity, affected component, exploitability, owner, remediation deadline и exception. Критичные externally exploitable issues блокируют release согласно policy.

## 15.33. Break-glass

Emergency access MUST:

- быть ограниченным по времени;
- требовать усиленной аутентификации;
- иметь reason/ticket;
- уведомлять security owner;
- полностью audit;
- не использовать общий account;
- автоматически отзываться.

## 15.34. Incident response

Для security incident MUST быть:

- detection signal;
- severity classification;
- containment steps;
- credential/key rotation plan;
- evidence preservation;
- notification procedure;
- recovery verification;
- post-incident review;
- tracked corrective actions.

## 15.35. Security testing

Обязательные классы:

- authentication state machine tests;
- token validation tests;
- access matrix/property tests;
- cross-tenant isolation tests;
- step-up tests;
- replay/CSRF/webhook signature tests;
- rate-limit tests;
- secret leakage tests;
- dependency outage fail-mode tests;
- privilege escalation tests;
- audit completeness tests.

## 15.36. SPDD-требования

Каждый feature/task prompt MUST перечислять:

```yaml
security:
  actor: USER
  subject: request.subject
  client: request.client_id
  permission: identity.user.disable
  resource_scope: projects/{project}/userPools/{pool}/users/{user}
  required_assurance: AAL2
  risk_evaluation: required
  audit: required
  sensitive_fields:
    - reason_internal
  forbidden_outputs:
    - access_token
    - otp_value
```

## 15.37. Критерии соответствия главы

Security design соответствует PADS, если идентичности разделены, trusted context проверен, Access/Risk/Authentication responsibilities не смешаны, tenant isolation доказана тестами, секреты защищены, failure mode определён, audit полный, а security controls включены в CI/CD и SPDD.

---
