---
title: "PADS: владение данными"
description: "Владельцы данных, проекции, репликация, удаление и retention."
keywords:
  - "M8 Platform"
  - "PADS"
  - "architecture"
---

# 11. Владение данными {#pads-data-ownership}

{% note info "Навигация PADS" %}

[Оглавление PADS](../index.md) | [Предыдущий раздел: 10. Shared Kernel и общие контракты](../domain/10-shared-kernel.md) | [Следующий раздел: 12. Правила проектирования API](12-api-design.md)

{% endnote %}

## 11.1. Назначение главы

Настоящая глава определяет, какой ограниченный контекст является нормативным владельцем каждого типа данных M8 Platform, где находится источник истины, какие копии разрешены другим сервисам и каким образом обеспечиваются репликация, удаление, перенос, восстановление и проверка происхождения данных.

Владение данными является следствием владения предметными инвариантами. Физическое размещение таблицы, кэша, индекса или аналитической копии само по себе не создаёт владение.

## 11.2. Нормативные понятия

| Понятие | Определение |
| --- | --- |
| Источник истины | Сервис и его хранилище, уполномоченные принимать нормативные решения о текущем состоянии данных. |
| Нормативный владелец | Ограниченный контекст, который определяет смысл, жизненный цикл и инварианты данных. |
| Авторитетный атрибут | Атрибут, изменяемый только владельцем и распространяемый потребителям через контракт. |
| Проекция | Локальная производная копия данных другого контекста, предназначенная для чтения или принятия ограниченного класса решений. |
| Кэш | Восстанавливаемая копия, удаление которой не приводит к потере предметного состояния. |
| Снимок | Зафиксированное состояние данных на определённый момент или позицию журнала. |
| Tombstone | Запись или событие, обозначающее удаление либо прекращение доступности исходного объекта. |
| Линия происхождения | Связь между производными данными, исходным контрактом, событием и версией схемы. |
| Класс свежести | Максимально допустимое отставание производной копии от источника истины. |

## 11.3. Принципы владения данными

| ID | Нормативное правило |
| --- | --- |
| `DATA-001` | Каждый предметный факт MUST иметь ровно одного нормативного владельца. |
| `DATA-002` | Владельцем данных является контекст, владеющий соответствующими инвариантами, а не команда, таблица или технология хранения. |
| `DATA-003` | Только владелец MAY изменять авторитетные атрибуты ресурса. |
| `DATA-004` | Сервис MUST NOT выполнять прямые запросы, соединения таблиц или записи в базу другого сервиса. |
| `DATA-005` | Межконтекстный доступ MUST осуществляться через опубликованный API, событие или управляемую проекцию. |
| `DATA-006` | Любая проекция MUST содержать ссылку на источник, версию схемы и позицию обновления. |
| `DATA-007` | Производная копия MUST NOT становиться источником истины без отдельного решения об изменении границы владения. |
| `DATA-008` | Кэш MUST быть восстанавливаемым из авторитетного источника или журнала событий. |
| `DATA-009` | Репликация MUST распространять подтверждённые факты после фиксации локальной транзакции. |
| `DATA-010` | Удаление данных MUST распространяться не менее надёжно, чем их создание и изменение. |
| `DATA-011` | Исторические ссылки в Audit MUST сохранять идентифицируемость действия без восстановления удалённых секретов или персональных данных. |
| `DATA-012` | Перенос владения данными между контекстами MUST выполняться как управляемая миграция с ADR, планом совместимости и сверкой данных. |
| `DATA-013` | Аналитическое хранилище MUST NOT использоваться для оперативной записи предметного состояния. |
| `DATA-014` | Индекс поиска MUST рассматриваться как проекция, а не как нормативное хранилище. |
| `DATA-015` | Redis MUST использоваться только как кэш, временное состояние, lease, rate limit или вспомогательный механизм; Redis MUST NOT быть единственным источником предметных данных. |
| `DATA-016` | Сырые credentials, refresh token, private key и секреты MUST NOT распространяться через события или проекции. |
| `DATA-017` | Каждая копия персональных или чувствительных данных MUST иметь обоснованную цель, срок хранения и владельца удаления. |
| `DATA-018` | Контракт данных MUST явно указывать семантику отсутствующего, пустого и удалённого значения. |
| `DATA-019` | Схема данных MUST развиваться совместимо с активными потребителями либо через новую major-версию. |
| `DATA-020` | Массовая сверка и восстановление проекций MUST быть предусмотрены для критичных интеграций. |
| `DATA-021` | Сервис MUST различать время предметного события, время записи и время публикации. |
| `DATA-022` | Идентификаторы ресурсов MUST сохранять глобальную однозначность в пределах заявленной области. |
| `DATA-023` | Нормативное состояние MUST изменяться с оптимистической проверкой revision или ETag, если возможны конкурентные записи. |
| `DATA-024` | Любая денормализация MUST иметь зафиксированного владельца обновления и способ исправления рассогласования. |
| `DATA-025` | Резервные копии MUST соответствовать требованиям RPO/RTO владельца данных. |
| `DATA-026` | Экспорт данных MUST фиксировать снимок, фильтр, формат, версию схемы и инициатора. |
| `DATA-027` | Импорт данных MUST проходить валидацию, дедупликацию, авторизацию и аудит. |
| `DATA-028` | Тестовые и небоевые среды MUST NOT получать продуктивные чувствительные данные без маскирования и разрешённого процесса. |
| `DATA-029` | Data lineage MUST быть доступен для аналитических, аудиторских и критичных операционных проекций. |
| `DATA-030` | Structured Prompt MUST указывать владельца каждого изменяемого набора данных и запрещённые источники записи. |

## 11.4. Каноническая матрица владельцев

| Данные или агрегат | Нормативный владелец | Авторитетные атрибуты | Основные потребители | Разрешённое распространение |
| --- | --- | --- | --- | --- |
| Organization | Resource Manager | идентификатор, имя, статус, labels, revision | все контексты | Resource API, `OrganizationCreated/Updated/Deleted` |
| Workspace | Resource Manager | родительская Organization, имя, статус, labels, revision | Access, Provisioning, Audit, UI/BFF | Resource API, Workspace events |
| Project | Resource Manager | Workspace, имя, статус, labels, revision | все продуктовые контексты | Resource API, Project events |
| ServiceRegistration | Resource Manager | Project, service type, lifecycle, metadata | Authentication, Access, Provisioning | API, Service events |
| UserPool | Identity | Project scope, policies, lifecycle | Authentication, Access, Audit | Identity API, UserPool events |
| User | Identity | status, profile, memberships, identity links | Authentication, Access, Audit | Identity API, privacy-filtered events |
| Group | Identity | membership composition, status | Access, UI/BFF | Identity API, Group events |
| ExternalIdentity | Identity | issuer, subject, link state | Authentication | Identity API; события без секретов |
| Client | Authentication | client type, allowed flows, assurance requirements, status | Access, Risk Decision, Audit | Authentication API, Client events |
| AuthenticationTransaction | Authentication | state, challenges, achieved assurance, handoff | Audit, support tools | Authentication API, lifecycle events |
| AuthenticationSession | Authentication | subject, client, assurance, expiry, revocation state | AuthGuard, Audit | защищённый API; минимальные session events |
| AuthorizationModel | Access | schema, model version, status | все сервисы через Access | Access API, model events |
| Role | Access | permissions, scope, lifecycle | UI/BFF, Audit | Access API, Role events |
| RoleBinding | Access | subject, role, resource scope, condition | все сервисы через checks | Access API, binding events |
| AccessRelationship | Access | subject-resource relation | Access evaluation, review tools | Access API, relationship events |
| RiskPolicy | Risk Decision | rules, version, rollout, status | Authentication, Provisioning | Risk API, policy events |
| RiskAssessment | Risk Decision | signals digest, score, decision, reasons | Authentication, Audit | Risk API, decision events |
| ResourceDefinition | Provisioning | schema, driver, lifecycle policy | Resource Manager, UI/BFF | Provisioning API, definition events |
| ManagedResource | Provisioning | desired state, observed state, placement, conditions | Resource Manager, Audit | Provisioning API, resource events |
| Reconciliation | Provisioning | attempt, drift, result, retry state | operations UI, Audit | Provisioning API, events |
| AuditEvent | Audit | actor, action, target, outcome, integrity metadata | compliance, security, export | Audit Query/Export API |
| Operation | сервис, запустивший операцию | state, progress, metadata, result/error | клиент операции, Audit | Common Operation API |
| Trace/Metric/Log | Observability platform, при сохранении предметного владельца источника | telemetry payload, timestamps, resource attributes | SRE, разработчики | OpenTelemetry pipeline |

## 11.5. Владение атрибутами составных представлений

Одно пользовательское представление MAY объединять данные нескольких владельцев. Такая композиция не создаёт нового общего агрегата.

Пример профиля проекта:

| Поле представления | Владелец |
| --- | --- |
| `project.id`, `project.name`, `project.status` | Resource Manager |
| `member_count` | Identity или специальная проекция, если Membership принадлежит Identity |
| `role_bindings_count` | Access |
| `managed_resources_count` | Provisioning |
| `last_security_decision_at` | Risk Decision projection |
| `last_audit_event_at` | Audit |

Композиция выполняется BFF, query service или аналитической проекцией. Она MUST указывать свежесть каждого компонента и MUST NOT принимать запись сразу за несколько владельцев одной транзакцией.

## 11.6. Классы производных копий

| Класс | Назначение | Допустимое отставание | Разрешено для авторизационного решения |
| --- | --- | --- | --- |
| `F0` | синхронно полученное состояние владельца | в пределах вызова | Да |
| `F1` | оперативная проекция | не более 5 секунд или установленного SLO | Только для явно разрешённых low-risk решений |
| `F2` | пользовательская read model | не более 1 минуты | Нет |
| `F3` | отчётная/операционная аналитика | не более 15 минут | Нет |
| `F4` | пакетная аналитика | часы или сутки | Нет |
| `F5` | архивный снимок | фиксированный момент | Нет |

Конкретный контракт MAY задавать более строгие значения. Решения Access, Authentication и Risk MUST использовать `F0`, если PADS или ADR явно не разрешают безопасную деградацию.

## 11.7. Метаданные проекции

Критичная проекция SHOULD хранить:

```yaml
projection_metadata:
  source_context: ResourceManager
  source_contract: m8.platform.resourcemanager.events.v1.ProjectUpdated
  source_resource_id: projects/prj_123
  source_revision: 42
  source_event_id: evt_01J...
  source_occurred_at: 2026-07-10T10:20:30Z
  applied_at: 2026-07-10T10:20:31Z
  schema_version: 1
  projection_version: 7
```

Потребитель MUST уметь определить, является ли проекция полной, отстающей, перестраиваемой или заблокированной ошибкой.

## 11.8. Репликация данных

Разрешённые способы:

1. **Синхронное чтение API** — когда требуется актуальное решение владельца.
2. **Integration Event** — для распространения подтверждённого факта и построения проекций.
3. **Snapshot + Change Stream** — для первичной загрузки большого набора с последующим применением изменений.
4. **Периодическая сверка** — для обнаружения пропущенных событий и расхождений.
5. **Управляемый экспорт** — для аналитики, compliance и миграций.

Запрещены:

- CDC непосредственно из внутренних таблиц сервиса как публичный контракт без владельца схемы;
- совместно используемые таблицы;
- репликация секретов и полных токенов;
- скрытая зависимость от физического порядка колонок или внутренних ключей хранения.

## 11.9. Первичная загрузка и восстановление проекций

Каждая критичная проекция MUST иметь один из путей восстановления:

- полный list/export владельца + позиция журнала;
- replay сохранённых integration events;
- versioned snapshot;
- детерминированная реконструкция из иного авторитетного источника.

Процедура перестроения MUST определять:

- точку отсечения;
- стратегию двойной записи или dual-read во время перестройки;
- проверку полноты;
- переключение версии;
- откат;
- очистку старой версии.

## 11.10. Удаление данных

Удаление различается по семантике:

| Тип | Семантика |
| --- | --- |
| Soft delete | ресурс недоступен обычным операциям, но остаётся восстанавливаемым в ограниченный срок |
| Tombstone | распространяемый факт, что ресурс удалён и проекции должны прекратить его использование |
| Hard delete | физическое удаление после истечения retention и выполнения ограничений |
| Anonymization | необратимое удаление идентифицирующих атрибутов с сохранением статистической или аудиторской структуры |
| Revocation | прекращение действия credentials, session, binding или разрешения без удаления исторического объекта |

Удаление родительского ресурса MUST учитывать дочерние ресурсы, активные операции, retention, legal hold и внешние managed resources. Каскадное физическое удаление между сервисами запрещено; используется управляемый процесс.

## 11.11. Право на удаление и минимизация персональных данных

Identity является координатором удаления персональных данных пользователя, но каждый сервис остаётся владельцем удаления собственных копий. Процесс SHOULD включать:

1. создание privacy operation;
2. определение идентификаторов субъекта;
3. отправку команд владельцам данных;
4. подтверждение удаления или обезличивания;
5. сохранение минимальной audit evidence;
6. закрытие операции после сверки.

Audit MAY сохранить actor pseudonym, event integrity и юридически обязательные сведения, но MUST удалить или маскировать необязательные персональные поля.

## 11.12. Сроки хранения

Каждый тип данных MUST иметь retention policy со следующими полями:

```yaml
retention_policy:
  data_class: authentication_transaction
  owner: m8-authentication
  active_retention: 30d
  archive_retention: 180d
  deletion_mode: hard_delete
  legal_hold_supported: true
  evidence_owner: m8-audit
```

Retention MUST быть согласован с целями продукта, безопасностью, юридическими обязанностями и стоимостью хранения.

## 11.13. Перенос ресурсов и данных

Перемещение Project между Workspace, перенос User Pool, смена placement или миграция между регионами являются предметными процессами, а не прямым обновлением foreign key.

Процесс переноса MUST определять:

- допустимость изменения родителя;
- влияние на resource names и отношения Access;
- обновление локальных проекций;
- блокировку конфликтующих операций;
- сохранение исторических ссылок;
- миграцию managed resources;
- подтверждение целостности;
- rollback или compensation.

Межрегиональный перенос MUST дополнительно учитывать residency, encryption keys, RPO/RTO и доступность внешних провайдеров.

## 11.14. Изменение владельца данных

Изменение owner context требует:

1. ADR и карты влияния;
2. новой published language;
3. backfill в новое хранилище;
4. периода shadow-read;
5. сверки количества, хэшей и инвариантов;
6. переключения writer;
7. переходного события или API;
8. удаления старой записи только после стабилизации;
9. обновления PADS, requirements, SPDD и runbooks.

Dual-write без координатора, reconciliation и ограниченного срока запрещён.

## 11.15. Внешние источники данных

Данные Keycloak, SpiceDB, Kubernetes, облачного провайдера и иных систем MUST проходить через ACL. M8 определяет, какие факты считаются:

- авторитетными во внешней системе;
- желаемыми в M8;
- наблюдаемыми;
- кэшированными;
- подтверждёнными;
- неизвестными.

Provisioning владеет desired state managed resource, но внешний провайдер MAY быть источником истины для observed state. Расхождение фиксируется как drift, а не устраняется скрытой записью в M8.

## 11.16. Аналитические данные

Аналитические витрины MAY объединять факты всех контекстов, но MUST:

- сохранять source IDs и event IDs;
- указывать период и timezone;
- хранить versioned metric definition;
- различать event time и processing time;
- обеспечивать возможность повторного расчёта;
- не использоваться для mutation API;
- иметь отдельные политики доступа и retention.

## 11.17. Классификация данных

| Класс | Примеры | Базовые ограничения |
| --- | --- | --- |
| Public | публичные схемы API, открытая документация | целостность и версионирование |
| Internal | технические метаданные, non-sensitive resource labels | доступ только внутри платформы |
| Confidential | профили пользователей, access relationships, risk reasons | least privilege, encryption, audit |
| Restricted | credentials, token material, secret values, sensitive risk signals | запрещено в событиях и логах, специализированное хранилище |

## 11.18. Контроль качества данных

Владельцы критичных наборов MUST определять:

- completeness;
- uniqueness;
- validity;
- referential validity на уровне контракта;
- freshness;
- consistency;
- reconciliation status.

Нарушения качества MUST формировать метрику, alert и при необходимости AuditEvent.

## 11.19. Трассировка и SPDD

Structured Prompt, изменяющий данные, MUST содержать:

```yaml
 data_ownership:
   owner_context: Identity
   authoritative_entities:
     - User
   writes:
     - identity/users
   reads_external:
     - ResourceManager.ProjectReference
   projections:
     - Access.UserStatusProjection
   forbidden:
     - direct_access_database_write
     - audit_event_mutation
   deletion_behavior:
     - emit UserDeleted tombstone
```

## 11.20. Критерии соответствия главы

Архитектура соответствует главе, если:

1. у каждого нормативного атрибута определён один владелец;
2. отсутствуют межсервисные database joins и записи;
3. проекции имеют источник, свежесть и механизм восстановления;
4. удаление распространяется на производные копии;
5. перенос и смена владельца оформлены как процессы;
6. аналитические и поисковые копии не используются как скрытый writer;
7. чувствительные данные классифицированы и минимизированы;
8. критичные наборы имеют retention, backup и reconciliation;
9. требования и Structured Prompts ссылаются на owner context.

---
