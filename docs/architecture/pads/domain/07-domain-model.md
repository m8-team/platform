---
title: "PADS: модель предметной области"
description: "Агрегаты, сущности, объекты-значения, события и инварианты."
keywords:
  - "M8 Platform"
  - "PADS"
  - "architecture"
---

# 7. Модель предметной области {#pads-domain-model}

{% note info "Навигация PADS" %}

[Оглавление PADS](../index.md) | [Предыдущий раздел: 6. Карта бизнес-возможностей платформы](06-capability-map.md) | [Следующий раздел: 8. Карта контекстов](08-context-map.md)

{% endnote %}

## 7.1. Назначение главы

Настоящая глава определяет каноническую модель предметной области M8 Platform на уровне, общем для всей платформы. Она фиксирует:

- состав основных агрегатов, сущностей и объектов-значений;
- границы атомарного изменения состояния;
- устойчивые идентификаторы предметных объектов;
- основные инварианты;
- жизненные циклы и допустимые переходы состояний;
- правила ссылок между агрегатами и ограниченными контекстами;
- соотношение предметного ресурса, длительной Operation и Workflow;
- правила публикации предметных фактов;
- точки расширения, которые будут детализированы в спецификациях ограниченных контекстов.

Модель предметной области является нормативной основой для:

- требований уровня контекста и сервиса;
- контрактов Protobuf и событий;
- моделей хранения;
- прикладных сценариев;
- интерфейсов Repository;
- архитектурных тестов;
- Structured Prompts;
- тестов инвариантов и жизненных циклов.

Настоящая глава **НЕ ДОЛЖНА** рассматриваться как физическая схема базы данных. Агрегат, сущность и объект-значение описывают предметную семантику, а не таблицы, документы или сообщения транспорта.

## 7.2. Нормативные правила моделирования

### DM-RULE-001. Агрегат является границей атомарности

Одна прикладная транзакция **ДОЛЖНА** изменять состояние не более одного Aggregate. Изменение нескольких Aggregate координируется через:

- доменные или интеграционные события;
- Process Manager;
- Workflow;
- длительную Operation;
- компенсационные действия.

Распределённая транзакция между сервисами **НЕ ДОПУСКАЕТСЯ**.

### DM-RULE-002. Агрегат имеет единственный корень

Каждый Aggregate **ДОЛЖЕН** иметь один Aggregate Root, через который выполняются все изменения состояния. Внутренние сущности не изменяются напрямую из прикладного слоя.

### DM-RULE-003. Агрегаты должны оставаться небольшими

Большая предметная иерархия **НЕ ДОЛЖНА** автоматически становиться одним Aggregate. В частности:

```text
Organization
Workspace
Project
ServiceRegistration
```

являются независимыми Aggregate, связанными типизированными идентификаторами. Создание Workspace не требует загрузки Organization как полного объектного графа; прикладной сценарий проверяет существование и допустимое состояние Organization через контракт владельца либо локальную доверенную проекцию.

### DM-RULE-004. Межагрегатная ссылка хранится как типизированный идентификатор

Aggregate **ДОЛЖЕН** ссылаться на другой Aggregate через устойчивый идентификатор и, при необходимости, типизированную Resource Reference. Он **НЕ ДОЛЖЕН** хранить изменяемую копию чужого Aggregate как часть собственного состояния.

### DM-RULE-005. Межконтекстная ссылка не создаёт совместного владения

Наличие `project_id`, `user_id`, `subject_ref` или иной внешней ссылки не делает потребляющий контекст владельцем указанного объекта. Потребитель может хранить:

- идентификатор;
- тип ресурса;
- минимальные неизменяемые атрибуты;
- версию проекции;
- время последнего подтверждения.

Изменяемые канонические данные остаются у контекста-владельца.

### DM-RULE-006. Объект-значение неизменяем

Value Object **ДОЛЖЕН** быть логически неизменяемым. Изменение значения создаёт новый экземпляр. Равенство Value Object определяется значением, а не техническим адресом или идентификатором строки хранения.

### DM-RULE-007. Инвариант защищается владельцем

Инвариант Aggregate **ДОЛЖЕН** проверяться внутри доменной модели или в доменной политике владельца до фиксации состояния. Проверка только на уровне интерфейса, контроллера или базы данных недостаточна.

### DM-RULE-008. Состояние предметного ресурса отделено от состояния Operation

`Project.state`, `User.status`, `ManagedResource.state` и другие предметные состояния описывают объект. `Operation.state` описывает выполнение конкретной асинхронной мутации. Эти состояния **НЕ ДОЛЖНЫ** объединяться в одно поле или взаимно подменяться.

### DM-RULE-009. Workflow не является агрегатом

Workflow координирует выполнение, ожидания, повторы и компенсации, но не является каноническим владельцем предметного состояния. Источник истины остаётся в Aggregate и его Repository.

### DM-RULE-010. Внешняя система скрывается антикоррупционным слоем

Keycloak, SpiceDB, Temporal, YDB, Kubernetes, облачные API и другие технологии **НЕ ДОЛЖНЫ** определять названия и структуру доменных типов. Преобразование выполняется Adapter или Driver.

### DM-RULE-011. Каждая мутация имеет ожидаемую версию

Если изменение может конфликтовать с конкурентной мутацией, команда **ДОЛЖНА** принимать `expected_version` либо эквивалентное условие. Успешное изменение увеличивает `ResourceVersion` монотонно.

### DM-RULE-012. Удалённый идентификатор не переиспользуется

Идентификатор Aggregate **НЕ ДОЛЖЕН** назначаться новому предметному объекту после удаления предыдущего. Исторические Audit Event, события и ссылки должны сохранять однозначность.

### DM-RULE-013. Domain Event возникает из изменения Aggregate

Domain Event **ДОЛЖЕН** отражать уже принятый предметный факт и формироваться в той же логической транзакции, что и изменение Aggregate. Публикация внешним потребителям выполняется через Outbox после фиксации.

### DM-RULE-014. Query result не является Aggregate

`PermissionCheck`, `AccessExplanation`, `SearchResult`, `RiskPreview` и другие вычисляемые результаты **НЕ ДОЛЖНЫ** моделироваться как Aggregate без собственного жизненного цикла и идентичности.

### DM-RULE-015. Временное техническое состояние не становится предметным автоматически

Повтор вызова, номер попытки, lease обработчика, offset сообщения, идентификатор Workflow Execution и другие инфраструктурные данные могут храниться рядом с предметным объектом, но **НЕ ДОЛЖНЫ** становиться частью публичной предметной модели без отдельного обоснования.

## 7.3. Базовые строительные блоки

| Блок | Нормативное значение | Пример |
| --- | --- | --- |
| Aggregate | Граница атомарного изменения и защиты инвариантов. | `Project`, `User`, `AuthenticationTransaction`. |
| Aggregate Root | Единственная точка изменения Aggregate. | `ManagedResource`. |
| Entity | Объект с собственной идентичностью внутри Aggregate или контекста. | `AuthenticationChallenge`. |
| Value Object | Неизменяемое понятие, определяемое значением. | `AssuranceLevel`, `ResourceReference`. |
| Domain Service | Предметная операция, не принадлежащая естественным образом одной Entity. | Выбор допустимого Authentication Method. |
| Domain Policy | Версионируемое правило или алгоритм принятия предметного решения. | Политика размещения Managed Resource. |
| Repository | Интерфейс получения и сохранения Aggregate. | `ProjectRepository`. |
| Factory | Создаёт допустимый начальный Aggregate и проверяет входные условия. | `AuthenticationTransactionFactory`. |
| Specification | Повторно используемый предметный предикат. | «Project допускает регистрацию Service». |
| Domain Event | Неизменяемый факт внутри контекста. | `ProjectCreated`. |
| Integration Event | Стабильное опубликованное представление факта. | `m8.resourcemanager.ProjectCreated.v1`. |
| Process Manager | Координация длительного межагрегатного процесса. | Управляемое удаление Organization. |
| Projection | Локальная модель чтения, построенная по опубликованным фактам. | Проекция активных Project в Access. |

## 7.4. Общие объекты-значения платформы

Общие объекты-значения определяют форму ссылок и технически нейтральные правила, но не образуют Shared Kernel предметной логики между всеми контекстами.

| ID | Value Object | Состав и правила |
| --- | --- | --- |
| VO-COM-001 | `ResourceId` | Типизированный неизменяемый идентификатор; глобально либо контекстно уникален согласно спецификации типа. |
| VO-COM-002 | `ResourceName` | Проверенное человекочитаемое имя в определённой области уникальности. Не заменяет ID. |
| VO-COM-003 | `ResourceReference` | `resource_type + resource_id`; может включать канонический parent scope для валидации адресации. |
| VO-COM-004 | `ResourceVersion` | Положительное монотонное значение версии Aggregate. |
| VO-COM-005 | `Labels` | Проверенный набор пар ключ—значение; порядок незначим; не определяет права и инварианты. |
| VO-COM-006 | `Description` | Ограниченный по длине безопасный текст без управляющей предметной семантики. |
| VO-COM-007 | `Timestamp` | Время в UTC с явно заданной точностью; хранится независимо от локальной временной зоны. |
| VO-COM-008 | `TimeRange` | Интервал с определёнными правилами включения границ и условием `from < to`. |
| VO-COM-009 | `IdempotencyKey` | Непрозрачный идентификатор клиентского намерения в ограниченной области уникальности. |
| VO-COM-010 | `CorrelationId` | Идентификатор сквозного процесса. Не используется как Resource ID. |
| VO-COM-011 | `CausationId` | Идентификатор непосредственной причины сообщения. |
| VO-COM-012 | `RequestContext` | Безопасный набор `request_id`, `trace_id`, actor/principal, source, time и сетевого контекста. |
| VO-COM-013 | `PageToken` | Непрозрачный подписанный или защищённый курсор продолжения списка. |
| VO-COM-014 | `FieldMask` | Явно допустимый набор полей мутации или представления. |
| VO-COM-015 | `ErrorCode` | Стабильный машинный код предметной или прикладной ошибки. |
| VO-COM-016 | `ExternalReference` | Типизированная ссылка на объект внешней системы: provider, external_id и необязательная версия. |

### 7.4.1. Правила идентификаторов

- ID **ДОЛЖЕН** быть непрозрачным для потребителя.
- Потребитель **НЕ ДОЛЖЕН** извлекать бизнес-смысл, регион, дату или тип из внутренней структуры ID.
- Тип ID **ДОЛЖЕН** быть различим в доменном коде: `ProjectId` не взаимозаменяем с `UserId`.
- Строковое представление MAY включать префикс типа для диагностики, но префикс не является источником авторизации.
- Внешний идентификатор **НЕ ДОЛЖЕН** использоваться как канонический ID M8 без отдельной гарантии неизменяемости и владения.

### 7.4.2. Правила версии Aggregate

```text
expected_version == current_version
        ↓
применить предметную операцию
        ↓
new_version = current_version + 1
        ↓
атомарно сохранить Aggregate + Outbox
```

При несовпадении версии мутация завершается стабильной ошибкой `CONCURRENT_MODIFICATION` или эквивалентным кодом контекста. Автоматический повтор допустим только когда прикладной сценарий способен безопасно повторно вычислить решение на новом состоянии.

## 7.5. Общая карта агрегатов

```text
Resource Manager
├── Organization
├── Workspace
├── Project
└── ServiceRegistration

Identity
├── UserPool
├── User
│   ├── UserProfile
│   ├── ExternalIdentity
│   └── CredentialReference
├── Group
└── Membership

Authentication
├── Client
├── AuthenticationProviderConfiguration
├── AuthenticationTransaction
│   └── AuthenticationChallenge
└── AuthenticationSession

Access
├── AuthorizationModel
├── Role
├── RoleBinding
└── AccessRelationship

Risk Decision
├── DecisionPolicy
└── RiskAssessment
    ├── RiskSignalSnapshot
    └── DecisionResult

Provisioning
├── ResourceDefinition
├── PlacementPolicy
└── ManagedResource
    ├── DesiredState
    ├── ObservedState
    ├── ResourceCondition
    └── ReconciliationRecord

Audit
├── AuditEvent
├── RetentionPolicy
└── AuditExportJob

Common Operation contract
└── Operation
    ├── OperationProgress
    ├── OperationResult
    └── OperationError
```

`Common Operation` определяет общий контракт и поведение, но **НЕ ОБЯЗАН** быть единым централизованным сервисом. Operation хранится у сервиса, владеющего исходной мутацией, если отдельный ADR не устанавливает иное.

## 7.6. Модель Resource Manager

### 7.6.1. Назначение модели

Resource Manager владеет канонической административной иерархией M8 Platform:

```text
Organization → Workspace → Project → ServiceRegistration
```

Каждый узел является отдельным Aggregate. Связь вложенности принадлежит дочернему Aggregate как обязательная типизированная ссылка на непосредственного родителя, а Resource Manager обеспечивает её валидность.

### 7.6.2. Aggregate `Organization`

| Атрибут | Тип | Правило |
| --- | --- | --- |
| `organization_id` | `OrganizationId` | Неизменяемый и не переиспользуемый. |
| `name` | `ResourceName` | Уникальность определяется политикой платформы. |
| `display_name` | строка | Может изменяться без изменения идентичности. |
| `description` | `Description` | Необязательное описание. |
| `status` | `OrganizationStatus` | Управляет допустимостью дочерних операций. |
| `labels` | `Labels` | Не влияют на авторизацию автоматически. |
| `version` | `ResourceVersion` | Изменяется при каждой мутации. |
| `created_at` | `Timestamp` | Неизменяемое время создания. |
| `updated_at` | `Timestamp` | Время последнего изменения. |

Предлагаемый жизненный цикл:

```text
ACTIVE
  → SUSPENDED
  → ACTIVE

ACTIVE | SUSPENDED
  → DELETION_PENDING
  → DELETED
```

`DELETED` является терминальным логическим состоянием. Физическое удаление канонической записи допускается только в соответствии с политикой хранения и при сохранении исторической адресуемости.

Основные команды:

- `CreateOrganization`;
- `UpdateOrganization`;
- `SuspendOrganization`;
- `ResumeOrganization`;
- `DeleteOrganization`;
- `GetOrganization`;
- `ListOrganizations`.

Основные факты:

- `OrganizationCreated`;
- `OrganizationUpdated`;
- `OrganizationSuspended`;
- `OrganizationResumed`;
- `OrganizationDeletionRequested`;
- `OrganizationDeleted`.

### 7.6.3. Aggregate `Workspace`

`Workspace` является самостоятельным Aggregate и хранит неизменяемый `organization_id` непосредственного родителя.

| Атрибут | Правило |
| --- | --- |
| `workspace_id` | Неизменяемый ID. |
| `organization_id` | Обязательная неизменяемая ссылка на существующую Organization. |
| `name` | Уникально в пределах Organization. |
| `status` | Не может быть активным, когда родитель окончательно удалён. |
| `labels`, `version`, timestamps | По общим правилам ресурса. |

Перемещение Workspace в другую Organization обычной командой Update **ЗАПРЕЩЕНО**. Если возможность переноса понадобится, она оформляется отдельным управляемым процессом с новым идентификатором или специальной миграционной семантикой и ADR.

### 7.6.4. Aggregate `Project`

`Project` является основной границей изоляции и адресации прикладных ресурсов.

| Атрибут | Правило |
| --- | --- |
| `project_id` | Неизменяемый ID. |
| `workspace_id` | Обязательная неизменяемая ссылка на Workspace. |
| `name` | Уникально в пределах Workspace. |
| `status` | Определяет возможность регистрации Service и создания зависимых ресурсов. |
| `labels` | Классификация и поиск. |
| `version` | Оптимистический контроль. |

Project **НЕ ДОЛЖЕН** содержать полные коллекции User, Role, ManagedResource или Service в состоянии Aggregate. Такие объекты принадлежат собственным агрегатам и контекстам.

### 7.6.5. Aggregate `ServiceRegistration`

`ServiceRegistration` представляет каноническую регистрацию прикладного или платформенного Service внутри Project.

Минимальная модель:

```text
ServiceRegistration
├── service_id
├── project_id
├── name
├── display_name
├── service_type
├── owner_reference
├── endpoints metadata
├── status
├── labels
└── version
```

`ServiceRegistration` **НЕ ЯВЛЯЕТСЯ**:

- deployment;
- pod;
- процессом ОС;
- Service Account;
- клиентом OAuth/OIDC;
- Managed Resource.

Связи с Authentication Client, Service Account и Managed Resource создаются через отдельные ссылки и процессы, не через объединение агрегатов.

### 7.6.6. Инварианты Resource Manager

| ID | Инвариант |
| --- | --- |
| INV-RM-001 | Workspace **ДОЛЖЕН** принадлежать ровно одной Organization. |
| INV-RM-002 | Project **ДОЛЖЕН** принадлежать ровно одному Workspace. |
| INV-RM-003 | ServiceRegistration **ДОЛЖЕН** принадлежать ровно одному Project. |
| INV-RM-004 | Parent ID **НЕ ДОЛЖЕН** изменяться обычной командой Update. |
| INV-RM-005 | Resource ID **НЕ ДОЛЖЕН** переиспользоваться после удаления. |
| INV-RM-006 | Имя дочернего ресурса **ДОЛЖНО** быть уникально в установленной родительской области. |
| INV-RM-007 | Создание дочернего ресурса **ЗАПРЕЩЕНО**, если родитель не существует либо находится в несовместимом состоянии. |
| INV-RM-008 | Окончательное удаление родителя **ЗАПРЕЩЕНО**, пока существуют активные дочерние ресурсы, если не запущен управляемый каскадный процесс. |
| INV-RM-009 | Resource State **НЕ ДОЛЖЕН** использоваться как состояние выполняющей его Operation. |
| INV-RM-010 | Каждая успешная мутация **ДОЛЖНА** увеличить ResourceVersion. |
| INV-RM-011 | Label **НЕ ДОЛЖЕН** предоставлять Permission сам по себе. |
| INV-RM-012 | Изменение канонической иерархии **ДОЛЖНО** сформировать Audit Event и опубликованный факт. |

## 7.7. Модель Identity

### 7.7.1. Aggregate `UserPool`

`UserPool` определяет изолированную область Identity. Он содержит настройки жизненного цикла и допустимых типов субъектов, но не включает всех User как внутреннюю коллекцию Aggregate.

Минимальная модель:

```text
UserPool
├── user_pool_id
├── project_id or owning_scope
├── name
├── subject_types
├── identity_policy_ref
├── status
├── labels
└── version
```

Конкретная область владения User Pool (`Project`, `Organization` либо иной поддерживаемый scope) должна быть однозначно задана в спецификации Identity. Один User **ДОЛЖЕН** принадлежать ровно одному User Pool.

### 7.7.2. Aggregate `User`

`User` представляет устойчивую локальную Identity. Внутри Aggregate могут находиться:

- `UserProfile` как Value Object;
- `ExternalIdentity` как Entity с локальной идентичностью;
- `CredentialReference` как Entity или Value Object в зависимости от жизненного цикла;
- статус и причины ограничения;
- безопасные атрибуты восстановления и связи.

`User` **НЕ ДОЛЖЕН** хранить:

- пароль в открытом или обратимо расшифровываемом виде;
- секрет WebAuthn private key;
- полный токен внешнего поставщика без специального защищённого хранилища;
- роли и разрешения Access как часть профиля;
- активную Authentication Transaction.

Предлагаемый жизненный цикл:

```text
PENDING
  → ACTIVE

ACTIVE
  ↔ SUSPENDED

PENDING | ACTIVE | SUSPENDED
  → DISABLED
  → DELETED
```

`SUSPENDED` обозначает временное ограничение с возможностью восстановления. `DISABLED` обозначает запрещённое использование до отдельного административного решения. Точная семантика переходов детализируется в спецификации Identity.

### 7.7.3. Entity `ExternalIdentity`

Минимальная идентичность внешней связи:

```text
issuer + subject
```

Дополнительно могут храниться:

- provider reference;
- подтверждённые атрибуты;
- время последней проверки;
- статус связи;
- безопасный fingerprint внешней учётной записи.

Значение `(issuer, subject)` **ДОЛЖНО** быть уникально в установленной глобальной или User Pool области. Одна подтверждённая External Identity не может одновременно принадлежать двум активным User, если только отдельный тип поставщика явно не допускает коллективную идентичность.

### 7.7.4. Aggregate `Group`

`Group` является именованной коллекцией Subject в пределах User Pool. Group не содержит Role и не является Permission.

Членство в Group может реализовываться через отдельный Aggregate `Membership`, если требуется:

- независимый жизненный цикл;
- временная действительность;
- большое количество участников;
- аудит каждой связи;
- идемпотентное добавление и удаление.

### 7.7.5. Aggregate `Membership`

`Membership` представляет одну связь между Subject и Scope либо Group.

```text
Membership
├── membership_id
├── subject_reference
├── target_reference
├── membership_type
├── valid_from
├── valid_until
├── status
├── source
└── version
```

Membership подтверждает принадлежность, но **НЕ ГАРАНТИРУЕТ** Permission. Access может использовать факт Membership как вход в собственную модель отношений.

### 7.7.6. Инварианты Identity

| ID | Инвариант |
| --- | --- |
| INV-ID-001 | User **ДОЛЖЕН** принадлежать ровно одному User Pool. |
| INV-ID-002 | External Identity **ДОЛЖНА** быть уникальна по нормативному ключу поставщика. |
| INV-ID-003 | User Profile **НЕ ДОЛЖЕН** содержать Credential secret. |
| INV-ID-004 | Удаление User **НЕ ДОЛЖНО** разрушать исторические Audit Reference. |
| INV-ID-005 | Membership **НЕ ДОЛЖНО** считаться Permission. |
| INV-ID-006 | Истёкшее Membership **НЕ ДОЛЖНО** публиковаться как активное. |
| INV-ID-007 | Group **ДОЛЖНА** принадлежать одному User Pool. |
| INV-ID-008 | Subject Reference **ДОЛЖНА** содержать тип субъекта и устойчивый ID. |
| INV-ID-009 | Service Account **НЕ ДОЛЖЕН** автоматически считаться Authentication Client или Service. |
| INV-ID-010 | Объединение Identity **ДОЛЖНО** сохранять журнал происхождения и предотвращать неоднозначные активные связи. |
| INV-ID-011 | Изменение статуса User **ДОЛЖНО** публиковать факт для Authentication и Access. |
| INV-ID-012 | Повтор создания связи с тем же идемпотентным намерением **НЕ ДОЛЖЕН** создавать дубликат. |

## 7.8. Модель Authentication

### 7.8.1. Aggregate `Client`

`Client` представляет зарегистрированное приложение, которому разрешено инициировать Authentication.

Минимальная модель:

```text
Client
├── client_id
├── owning_scope
├── name
├── client_type
├── allowed_flows
├── allowed_authentication_methods
├── redirect/handoff policy
├── assurance policy
├── status
└── version
```

Authentication Client не является Service Account. Связь между ними может существовать, но их идентичность, секреты и жизненные циклы различаются.

### 7.8.2. Aggregate `AuthenticationProviderConfiguration`

Настройка поставщика скрывает внешнюю технологию за предметным интерфейсом.

```text
AuthenticationProviderConfiguration
├── provider_id
├── owning_scope
├── supported_methods
├── capabilities
├── secret_reference
├── routing attributes
├── status
└── version
```

Секреты поставщика хранятся через `secret_reference`, а не в открытом состоянии Aggregate.

### 7.8.3. Aggregate `AuthenticationTransaction`

`AuthenticationTransaction` представляет один ограниченный по времени процесс проверки заявленной Identity.

Ключевые поля:

```text
AuthenticationTransaction
├── authentication_id
├── client_id
├── authentication_subject
├── requested_assurance_level
├── achieved_assurance_level
├── selected_method
├── selected_provider_id
├── risk_decision_reference
├── current_challenge_id
├── challenges[]
├── handoff
├── state
├── state_reason
├── expires_at
└── version
```

`AuthenticationChallenge` является Entity внутри AuthenticationTransaction, поскольку его жизненный цикл подчинён одной транзакции и изменение Challenge влияет на инварианты общего достигнутого уровня подтверждения.

### 7.8.4. Состояния Authentication Transaction

```text
CREATED
  ├──→ CHALLENGE_REQUIRED
  │      └──→ CHALLENGE_PENDING
  │               ├──→ CHALLENGE_REQUIRED   (следующий шаг)
  │               └──→ AUTHENTICATED
  └──→ AUTHENTICATED                       (challenge не требуется)

AUTHENTICATED
  → HANDOFF_CREATED
  → COMPLETED

Из нетерминальных состояний:
  → CANCELLED
  → EXPIRED
  → FAILED
```

Правила переходов:

- `AUTHENTICATED` достигается только когда `achieved_assurance_level` удовлетворяет итоговому требованию.
- `HANDOFF_CREATED` допускается только после успешной Authentication.
- `COMPLETED` означает, что согласованный результат передачи сформирован или употреблён согласно типу handoff.
- `CANCELLED`, `EXPIRED`, `FAILED`, `COMPLETED` являются терминальными, если спецификация конкретного восстановления не создаёт новую Authentication Transaction.
- Повторная аутентификация создаёт новую транзакцию и не изменяет завершённую старую.

### 7.8.5. Entity `AuthenticationChallenge`

```text
AuthenticationChallenge
├── challenge_id
├── method
├── provider_id
├── status
├── requested_at
├── expires_at
├── attempts
├── resend_count
├── provider_reference
└── result_metadata
```

Секретное значение OTP, private key и полный ответ поставщика **НЕ ДОЛЖНЫ** храниться в публичном представлении Challenge.

Рекомендуемые состояния Challenge:

```text
CREATED → DISPATCHED → PENDING → VERIFIED
                     ├→ REJECTED
                     ├→ EXPIRED
                     └→ CANCELLED
```

### 7.8.6. Aggregate `AuthenticationSession`

AuthenticationSession представляет ограниченное во времени подтверждённое состояние, пригодное для re-use согласно политике.

```text
ACTIVE → REVOKED
ACTIVE → EXPIRED
```

Session **НЕ ДОЛЖНА** использоваться после истечения, отзыва, нарушения binding или понижения допустимого уровня доверия. Решение использовать Session для конкретного действия остаётся за Authentication и Risk Decision, а не за клиентом единолично.

### 7.8.7. Инварианты Authentication

| ID | Инвариант |
| --- | --- |
| INV-AUTH-001 | AuthenticationTransaction **ДОЛЖНА** принадлежать одному Client. |
| INV-AUTH-002 | `expires_at` **ДОЛЖЕН** быть позже времени создания. |
| INV-AUTH-003 | Терминальная транзакция **НЕ ДОЛЖНА** принимать новые Challenge. |
| INV-AUTH-004 | `achieved_assurance_level` **НЕ ДОЛЖЕН** превышать подтверждённый набором успешных Challenge уровень. |
| INV-AUTH-005 | Handoff **НЕ ДОЛЖЕН** создаваться до состояния `AUTHENTICATED`. |
| INV-AUTH-006 | Повтор с тем же IdempotencyKey в одной области **ДОЛЖЕН** вернуть существующий результат намерения. |
| INV-AUTH-007 | Отключённый Client **НЕ ДОЛЖЕН** начинать новую Authentication. |
| INV-AUTH-008 | Отключённый Provider **НЕ ДОЛЖЕН** выбираться для нового Challenge. |
| INV-AUTH-009 | Provider secret **НЕ ДОЛЖЕН** попадать в Domain Event, Audit Event или API response. |
| INV-AUTH-010 | Успех внешнего Provider **НЕ ДОЛЖЕН** автоматически означать Permission на бизнес-действие. |
| INV-AUTH-011 | Новая step-up Authentication **ДОЛЖНА** иметь отдельную идентичность и ссылку на исходный контекст. |
| INV-AUTH-012 | Отмена **НЕ ДОЛЖНА** обещать откат уже выполненных внешних действий, если это не поддерживается Provider. |
| INV-AUTH-013 | Выбор метода **ДОЛЖЕН** учитывать Client policy, Provider capabilities и Risk Decision. |
| INV-AUTH-014 | Повторная отправка Challenge **ДОЛЖНА** соблюдать ограничения времени и количества. |
| INV-AUTH-015 | Authentication Subject **НЕ ДОЛЖЕН** подменяться после начала транзакции. |

## 7.9. Модель Access

### 7.9.1. Aggregate `AuthorizationModel`

`AuthorizationModel` определяет версионируемую схему типов Subject, Resource, Relation и Permission.

```text
AuthorizationModel
├── model_id
├── owning_scope
├── schema
├── semantic_version
├── status
├── validation_result
└── version
```

Изменение модели, несовместимое с существующими отношениями, требует миграционного плана и не выполняется как обычное обновление текста схемы.

### 7.9.2. Aggregate `Role`

`Role` является именованным набором Permission в определённой области.

```text
Role
├── role_id
├── owning_scope
├── name
├── permissions[]
├── assignable_subject_types[]
├── status
└── version
```

Role **НЕ ДОЛЖНА** содержать список всех назначенных Subject; назначения принадлежат `RoleBinding`.

### 7.9.3. Aggregate `RoleBinding`

`RoleBinding` связывает Subject или набор Subject с Role в заданной Resource scope.

```text
RoleBinding
├── role_binding_id
├── role_id
├── subjects[] or subject_set_ref
├── scope_reference
├── condition_ref
├── valid_from
├── valid_until
├── status
└── version
```

Для больших наборов субъектов следует использовать отдельные отношения или ссылку на Subject Set, а не неограниченную коллекцию внутри Aggregate.

### 7.9.4. Aggregate `AccessRelationship`

`AccessRelationship` представляет одну каноническую relation tuple в предметном языке Access:

```text
resource_reference
relation
subject_reference
optional condition
```

SpiceDB tuple является инфраструктурным представлением этой модели, но не публичным доменным типом.

### 7.9.5. Permission Check как вычисление

`PermissionCheck` не является Aggregate. Это Query, которая принимает:

- Subject Reference;
- Permission;
- Resource Reference;
- контекст решения;
- требуемую консистентность;
- необязательный snapshot/token модели.

и возвращает:

- `ALLOW`, `DENY` либо `CONDITIONAL/UNKNOWN` согласно контракту;
- основание решения;
- использованную версию модели;
- метаданные консистентности;
- безопасное объяснение.

### 7.9.6. Инварианты Access

| ID | Инвариант |
| --- | --- |
| INV-ACC-001 | Permission **ДОЛЖЕН** быть определён активной AuthorizationModel. |
| INV-ACC-002 | Role **НЕ ДОЛЖНА** включать неизвестный Permission. |
| INV-ACC-003 | RoleBinding **ДОЛЖЕН** ссылаться на существующую активную Role. |
| INV-ACC-004 | Scope RoleBinding **ДОЛЖЕН** быть совместим с типом Role и Resource. |
| INV-ACC-005 | Истёкший RoleBinding **НЕ ДОЛЖЕН** давать доступ. |
| INV-ACC-006 | AccessRelationship **ДОЛЖЕН** соответствовать типам и relation активной модели. |
| INV-ACC-007 | Access **НЕ ДОЛЖЕН** изменять Identity, Resource Manager или Authentication Aggregate. |
| INV-ACC-008 | Проверка доступа **НЕ ДОЛЖНА** считать успешную Authentication достаточной для ALLOW. |
| INV-ACC-009 | Удаление relation **ДОЛЖНО** быть идемпотентным. |
| INV-ACC-010 | Объяснение **НЕ ДОЛЖНО** раскрывать запрещённые связи и данные других субъектов. |
| INV-ACC-011 | Изменение полномочий **ДОЛЖНО** создавать Audit Event. |
| INV-ACC-012 | Локальное состояние и состояние SpiceDB **ДОЛЖНЫ** иметь механизм обнаружения и устранения рассинхронизации. |

## 7.10. Модель Risk Decision

### 7.10.1. Aggregate `DecisionPolicy`

`DecisionPolicy` является версионируемым набором правил, определяющим решение для конкретного типа операции.

```text
DecisionPolicy
├── policy_id
├── owning_scope
├── decision_type
├── rule_set
├── required_signal_types
├── output_contract
├── status
├── effective_from
└── version
```

Изменение активной политики создаёт новую версию. Уже завершённая RiskAssessment сохраняет ссылку на фактически использованную версию.

### 7.10.2. Aggregate `RiskAssessment`

`RiskAssessment` представляет одну оценку риска в определённом контексте.

```text
RiskAssessment
├── assessment_id
├── decision_type
├── subject_reference
├── actor_reference
├── resource_reference
├── operation_context
├── policy_id + policy_version
├── signals_snapshot[]
├── score
├── risk_level
├── decision_result
├── explanation_codes[]
├── state
├── expires_at
└── version
```

`RiskSignalSnapshot` хранит нормализованный факт и происхождение, необходимые для воспроизводимости решения. Сырые чувствительные данные должны храниться минимально и согласно политике безопасности.

### 7.10.3. Состояния Risk Assessment

```text
CREATED
  → COLLECTING_SIGNALS
  → EVALUATING
  → DECIDED

Из нетерминальных состояний:
  → FAILED
  → EXPIRED
```

Результат решения рекомендуется выражать как:

```text
ALLOW
DENY
CHALLENGE
REVIEW
```

`CHALLENGE` должен содержать требуемый Assurance Level или допустимые типы дополнительного подтверждения. Risk Decision не создаёт Authentication Challenge напрямую; это ответственность Authentication.

### 7.10.4. Инварианты Risk Decision

| ID | Инвариант |
| --- | --- |
| INV-RISK-001 | Завершённая Assessment **ДОЛЖНА** ссылаться на точную версию DecisionPolicy. |
| INV-RISK-002 | `DECIDED` **ДОЛЖНО** содержать один допустимый Decision Result. |
| INV-RISK-003 | Истёкшее решение **НЕ ДОЛЖНО** использоваться как актуальное без повторной оценки. |
| INV-RISK-004 | Risk Decision **НЕ ДОЛЖЕН** самостоятельно выдавать Permission. |
| INV-RISK-005 | `CHALLENGE` **ДОЛЖЕН** содержать формализованное требование, понятное Authentication. |
| INV-RISK-006 | Signal source и время наблюдения **ДОЛЖНЫ** быть сохранены для объяснимости. |
| INV-RISK-007 | Недоступность необязательного сигнала **НЕ ДОЛЖНА** автоматически трактоваться как его безопасное значение. |
| INV-RISK-008 | Чувствительные сигналы **НЕ ДОЛЖНЫ** попадать в публичное объяснение. |
| INV-RISK-009 | Повтор оценки с тем же assessment intent MAY вернуть существующее решение, пока оно не истекло и контекст не изменился. |
| INV-RISK-010 | Decision Policy **ДОЛЖНА** пройти валидацию до активации. |
| INV-RISK-011 | Изменение Policy **ДОЛЖНО** быть аудировано. |
| INV-RISK-012 | Risk Score **НЕ ДОЛЖЕН** использоваться вне определённой версии модели и шкалы без преобразования. |

## 7.11. Модель Provisioning

### 7.11.1. Aggregate `ResourceDefinition`

`ResourceDefinition` описывает тип управляемого ресурса, его Desired State schema, наблюдаемые свойства, допустимые операции и совместимые Driver.

```text
ResourceDefinition
├── definition_id
├── resource_type
├── schema_version
├── desired_state_schema
├── observed_state_schema
├── lifecycle_capabilities
├── driver_requirements
├── status
└── version
```

Новая несовместимая схема создаёт новую версию определения и требует стратегии миграции ManagedResource.

### 7.11.2. Aggregate `PlacementPolicy`

`PlacementPolicy` определяет правила выбора Provider, региона, кластера, зоны, класса стоимости и ограничений размещения.

Placement является результатом применения политики к конкретному запросу и сохраняется в ManagedResource. Автоматическое изменение Placement после создания допускается только как управляемая миграция.

### 7.11.3. Aggregate `ManagedResource`

`ManagedResource` является главным Aggregate Provisioning.

```text
ManagedResource
├── managed_resource_id
├── project_id
├── definition_id + schema_version
├── desired_state
├── observed_state
├── placement
├── provider_reference
├── external_resource_id
├── conditions[]
├── reconciliation metadata
├── lifecycle_state
├── generation
├── observed_generation
└── version
```

`generation` увеличивается при изменении Desired State. `observed_generation` показывает поколение Desired State, которое подтверждено последним достоверным наблюдением.

### 7.11.4. Жизненный цикл Managed Resource

```text
REQUESTED
  → PROVISIONING
  → READY

READY
  → UPDATING
  → READY

READY | UPDATING | PROVISIONING
  → DEGRADED
  → RECONCILING
  → READY

Любое активное состояние
  → DRIFT_DETECTED
  → RECONCILING

Допустимое активное состояние
  → SUSPENDED
  → RECONCILING | READY

Активное или ошибочное состояние
  → DELETING
  → DELETED

Нетерминальный отказ:
  → FAILED
```

Точная матрица переходов зависит от ResourceDefinition. `FAILED` не всегда терминально: новый Reconciliation или изменение Desired State может восстановить ресурс.

### 7.11.5. Entity `ResourceCondition`

Condition содержит:

```text
condition_type
status: TRUE | FALSE | UNKNOWN
reason_code
message
observed_at
transition_at
observed_generation
```

Condition должна быть ограничена стабильным словарём типа ресурса и не заменяет общий Lifecycle State.

### 7.11.6. Reconciliation

Reconciliation является повторяемым процессом:

```text
Desired State
    +
Observed State
    +
Resource Definition
    +
Placement / Provider capabilities
        ↓
Reconciliation Plan
        ↓
Driver actions
        ↓
New Observed State and Conditions
```

Temporal может исполнять Reconciliation Workflow, но Workflow state не является каноническим состоянием ManagedResource.

### 7.11.7. Инварианты Provisioning

| ID | Инвариант |
| --- | --- |
| INV-PROV-001 | ManagedResource **ДОЛЖЕН** ссылаться на существующую версию ResourceDefinition. |
| INV-PROV-002 | Desired State **ДОЛЖЕН** соответствовать схеме Definition до сохранения. |
| INV-PROV-003 | Изменение Desired State **ДОЛЖНО** увеличить generation. |
| INV-PROV-004 | `observed_generation` **НЕ ДОЛЖЕН** превышать generation. |
| INV-PROV-005 | External Resource ID **ДОЛЖЕН** быть уникален в области Provider и типа. |
| INV-PROV-006 | Driver **НЕ ДОЛЖЕН** становиться владельцем ManagedResource. |
| INV-PROV-007 | Удаление записи ManagedResource **НЕ ДОЛЖНО** предшествовать управляемому Deprovisioning, кроме явно подтверждённого режима orphan handling. |
| INV-PROV-008 | Повтор Driver action **ДОЛЖЕН** быть идемпотентным либо защищён deduplication token. |
| INV-PROV-009 | Drift **ДОЛЖЕН** определяться сравнением нормализованных состояний, а не только неравенством сырых payload. |
| INV-PROV-010 | Секреты Desired State **ДОЛЖНЫ** храниться через Secret Reference. |
| INV-PROV-011 | Изменение Placement существующего ресурса **ДОЛЖНО** выполняться как отдельная миграционная операция. |
| INV-PROV-012 | Состояние `READY` **ДОЛЖНО** соответствовать обязательным Condition типа ресурса. |
| INV-PROV-013 | Resource Request клиента **НЕ ДОЛЖЕН** передаваться Provider без нормализации и валидации. |
| INV-PROV-014 | Orphan Resource **НЕ ДОЛЖЕН** автоматически удаляться без политики и подтверждения владения. |
| INV-PROV-015 | Каждая значимая попытка Reconciliation **ДОЛЖНА** быть наблюдаема и связана с Operation/Workflow correlation. |

## 7.12. Модель Audit

### 7.12.1. Aggregate `AuditEvent`

AuditEvent является неизменяемой адресуемой записью. После принятия Audit её содержимое **НЕ ДОЛЖНО** изменяться. Коррекция выполняется отдельным связанным событием, а не обновлением исходного.

```text
AuditEvent
├── audit_event_id
├── occurred_at
├── recorded_at
├── source
├── actor
├── subject
├── action
├── targets[]
├── outcome
├── change_set
├── audit_context
├── correlation_id
├── causation_id
├── integrity metadata
└── classification
```

`occurred_at` — время действия у источника; `recorded_at` — время принятия Audit. Они не взаимозаменяемы.

### 7.12.2. Aggregate `RetentionPolicy`

RetentionPolicy задаёт:

- класс Audit Event;
- минимальный и максимальный срок хранения;
- допустимость архивирования;
- требования к Integrity Proof;
- основания и процесс удаления;
- режим legal hold.

Активная политика версионируется. Изменение политики не должно неявно сокращать обязательный срок уже записанных событий.

### 7.12.3. Aggregate `AuditExportJob`

AuditExportJob представляет управляемый асинхронный экспорт:

```text
REQUESTED → RUNNING → SUCCEEDED
                    ├→ FAILED
                    └→ CANCELLED
```

Он связывается с общей Operation, но хранит собственные предметные параметры выборки, формат, классификацию, срок доступности результата и подтверждение доступа.

### 7.12.4. Инварианты Audit

| ID | Инвариант |
| --- | --- |
| INV-AUD-001 | Принятый AuditEvent **НЕ ДОЛЖЕН** изменяться. |
| INV-AUD-002 | `audit_event_id` **НЕ ДОЛЖЕН** переиспользоваться. |
| INV-AUD-003 | Audit Actor и Audit Source **ДОЛЖНЫ** быть различимы. |
| INV-AUD-004 | Секреты, токены и Credential **НЕ ДОЛЖНЫ** записываться в открытом виде. |
| INV-AUD-005 | Audit Context **ДОЛЖЕН** содержать достаточную корреляцию для расследования. |
| INV-AUD-006 | Повторная доставка одного события **НЕ ДОЛЖНА** создавать логический дубликат. |
| INV-AUD-007 | Исправление факта **ДОЛЖНО** быть отдельным связанным AuditEvent. |
| INV-AUD-008 | Retention Policy **НЕ ДОЛЖНА** нарушать обязательный legal hold. |
| INV-AUD-009 | Экспорт **ДОЛЖЕН** проверять Permission и классификацию данных. |
| INV-AUD-010 | Integrity metadata **ДОЛЖНА** позволять обнаружить недопустимое изменение или разрыв последовательности в установленной области. |
| INV-AUD-011 | Недоставка обязательного Audit Event **ДОЛЖНА** быть обнаруживаема и иметь управляемую политику восстановления. |
| INV-AUD-012 | Audit **НЕ ДОЛЖЕН** использоваться как единственная операционная телеметрия сервиса. |

## 7.13. Общая модель Operation

### 7.13.1. Aggregate `Operation`

Operation представляет адресуемое состояние одной асинхронной мутации.

```text
Operation
├── operation_id
├── operation_type
├── owner_service
├── owner_resource_reference
├── actor_reference
├── request_reference
├── state
├── stage
├── progress
├── result
├── error
├── cancel_requested
├── create_time
├── update_time
├── expire_time
└── version
```

Operation **ДОЛЖНА** иметь предметного владельца. Создание Project принадлежит Resource Manager; создание ManagedResource принадлежит Provisioning; Audit Export принадлежит Audit.

### 7.13.2. Состояния Operation

```text
PENDING
  → RUNNING
  → SUCCEEDED

PENDING | RUNNING
  → CANCELLING
  → CANCELLED

PENDING | RUNNING | CANCELLING
  → FAILED

Завершённая Operation после срока хранения:
  → EXPIRED
```

`EXPIRED` означает недоступность или истечение ресурса Operation согласно политике хранения, а не отмену результата предметной мутации.

### 7.13.3. Progress и Stage

- `stage` — стабильная именованная фаза, понятная внешнему потребителю;
- `percent` MAY использоваться только когда прогресс действительно измерим;
- `message` не заменяет стабильный stage/code;
- внутреннее имя Activity Temporal **НЕ ДОЛЖНО** публиковаться автоматически как stage;
- прогресс не является обещанием времени завершения.

### 7.13.4. Cancellation

Cancellation Request означает намерение прекратить дальнейшее выполнение. Он **НЕ ГАРАНТИРУЕТ**:

- мгновенную остановку;
- откат всех выполненных шагов;
- отсутствие предметного результата;
- удаление внешнего ресурса.

Контекст-владелец должен определить точки отменяемости и возможные Compensation.

### 7.13.5. Инварианты Operation

| ID | Инвариант |
| --- | --- |
| INV-OPS-001 | Operation **ДОЛЖНА** иметь одного owner service. |
| INV-OPS-002 | `SUCCEEDED` **ДОЛЖНО** содержать result либо типизированную ссылку на результат, если операция его предполагает. |
| INV-OPS-003 | `FAILED` **ДОЛЖНО** содержать стабильный OperationError. |
| INV-OPS-004 | Терминальная Operation **НЕ ДОЛЖНА** возвращаться в `RUNNING`. |
| INV-OPS-005 | Cancellation Request **НЕ ДОЛЖЕН** трактоваться как гарантированный rollback. |
| INV-OPS-006 | Operation State **НЕ ДОЛЖЕН** подменять Resource State. |
| INV-OPS-007 | Повтор исходной команды с тем же IdempotencyKey **ДОЛЖЕН** возвращать ту же Operation или окончательный эквивалентный результат. |
| INV-OPS-008 | Доступ к Operation **ДОЛЖЕН** проверяться по владельцу, actor и target scope. |
| INV-OPS-009 | OperationError **НЕ ДОЛЖЕН** раскрывать секреты и внутренние stack trace. |
| INV-OPS-010 | Expiration Operation **НЕ ДОЛЖНО** удалять созданный предметный ресурс. |
| INV-OPS-011 | Progress update **НЕ ДОЛЖЕН** уменьшать подтверждённый stage без явной предметной модели возврата. |
| INV-OPS-012 | Operation **ДОЛЖНА** быть связана с trace_id и correlation_id. |

## 7.14. Связи между агрегатами

### 7.14.1. Канонические ссылки

| Источник | Ссылка | Цель | Назначение |
| --- | --- | --- | --- |
| Workspace | `organization_id` | Organization | Непосредственный родитель. |
| Project | `workspace_id` | Workspace | Непосредственный родитель. |
| ServiceRegistration | `project_id` | Project | Область регистрации. |
| UserPool | `owning_scope` | Organization/Project | Область Identity. |
| Membership | `subject_reference` | User/Group/ServiceAccount | Участник связи. |
| Membership | `target_reference` | Group/Organization/Workspace/Project | Область членства. |
| AuthenticationTransaction | `client_id` | Client | Инициирующее приложение. |
| AuthenticationTransaction | `authentication_subject` | Identity reference | Заявленный Subject. |
| RoleBinding | `role_id` | Role | Назначаемая роль. |
| RoleBinding | `scope_reference` | Resource Manager resource | Область действия. |
| RiskAssessment | subject/resource references | Identity/Resource | Контекст решения. |
| ManagedResource | `project_id` | Project | Владелец прикладного ресурса. |
| AuditEvent | actor/subject/targets | Любой контекст | Исторические ссылки. |
| Operation | `owner_resource_reference` | Предметный ресурс | Результат или цель мутации. |

### 7.14.2. Ссылочная целостность между контекстами

Физический внешний ключ между базами сервисов **ЗАПРЕЩЁН**. Ссылочная целостность обеспечивается комбинацией:

1. синхронной проверки владельца перед критической мутацией;
2. локальной проекции опубликованных состояний;
3. версии или времени последнего подтверждения;
4. обработки событий удаления и блокировки;
5. периодической сверки;
6. компенсирующих процессов для гонок и итоговой согласованности.

Для каждой межконтекстной ссылки спецификация сервиса **ДОЛЖНА** определить:

- является ли ссылка жёсткой или исторической;
- требуется ли существование цели в момент создания;
- какие состояния цели допустимы;
- что происходит при её блокировке или удалении;
- какой источник проверки используется;
- допустимый период устаревания проекции.

## 7.15. Межконтекстные процессы

### 7.15.1. Создание Project

```text
Actor
  → Resource Manager: CreateProject
  → Access: check project.create on Workspace
  → Resource Manager: validate Workspace state
  → Project Aggregate: create
  → Repository + Outbox: atomic commit
  → Audit: ProjectCreated action
  → Integration Event: ProjectCreated.v1
  → Consumers update projections
```

Access не создаёт Project, а Audit не является источником факта его существования.

### 7.15.2. Начало Authentication с Risk Decision

```text
Client
  → Authentication: StartAuthentication
  → Identity: resolve Authentication Subject
  → Risk Decision: evaluate context
  → AuthenticationTransaction: select method and create Challenge
  → Repository + Outbox
  → Audit
```

Risk Decision возвращает решение, но Challenge создаёт Authentication.

### 7.15.3. Создание Managed Resource

```text
Caller
  → Access: check permission
  → Provisioning: validate Project projection/reference
  → ResourceDefinition: validate Desired State
  → ManagedResource: create generation 1
  → Operation: create
  → Workflow: reconcile via Driver
  → ManagedResource: update Observed State / Conditions
  → Operation: complete
  → Audit + Integration Events
```

Workflow и Driver не владеют ManagedResource.

### 7.15.4. Деактивация User

```text
Identity: UserDisabled
  → Authentication: revoke/deny reusable sessions according to policy
  → Access: invalidate relevant projections or relationships
  → Audit: record action and downstream outcomes
```

Последствия межконтекстны и итогово согласованы. Identity фиксирует исходный факт, но не изменяет чужие базы напрямую.

## 7.16. Доменные события

### 7.16.1. Минимальный состав Domain Event

Внутреннее Domain Event должно содержать:

- тип факта;
- ID Aggregate;
- новую версию Aggregate;
- время возникновения;
- actor/request context в безопасной форме;
- causation и correlation;
- минимальные предметные данные факта.

### 7.16.2. События по агрегатам

| Aggregate | Основные факты |
| --- | --- |
| Organization | Created, Updated, Suspended, Resumed, DeletionRequested, Deleted. |
| Workspace | Created, Updated, Suspended, Resumed, Deleted. |
| Project | Created, Updated, Suspended, Resumed, DeletionRequested, Deleted. |
| ServiceRegistration | Registered, Updated, Disabled, Deregistered. |
| UserPool | Created, Updated, Disabled, Deleted. |
| User | Created, Activated, ProfileUpdated, Suspended, Disabled, Restored, Deleted. |
| Membership | Granted, Updated, Revoked, Expired. |
| Client | Registered, Updated, Disabled, SecretRotated. |
| AuthenticationTransaction | Started, ChallengeCreated, ChallengeVerified, Authenticated, HandoffCreated, Completed, Failed, Cancelled, Expired. |
| AuthenticationSession | Created, Revoked, Expired. |
| Role | Created, PermissionsChanged, Disabled, Deleted. |
| RoleBinding | Created, Updated, Revoked, Expired. |
| AccessRelationship | Written, Deleted. |
| DecisionPolicy | Created, Validated, Activated, Superseded, Disabled. |
| RiskAssessment | Started, SignalsCollected, Decided, Failed, Expired. |
| ResourceDefinition | Registered, VersionAdded, Activated, Deprecated. |
| ManagedResource | Requested, DesiredStateChanged, PlacementSelected, ProvisioningStarted, Observed, Ready, DriftDetected, ReconciliationFailed, DeletionRequested, Deleted. |
| AuditExportJob | Requested, Started, Completed, Failed, Cancelled, Expired. |
| Operation | Created, Started, Progressed, CancellationRequested, Succeeded, Failed, Cancelled, Expired. |

Конкретный Integration Event **ДОЛЖЕН** проектироваться отдельно и не обязан повторять внутренний Domain Event один к одному.

## 7.17. Команды и методы Aggregate

Команда выражает намерение, а метод Aggregate принимает решение на основании текущего состояния.

Пример:

```text
SuspendUser(command)
  1. проверить expected_version;
  2. проверить, что User допускает переход в SUSPENDED;
  3. проверить reason;
  4. изменить status;
  5. увеличить version;
  6. сформировать UserSuspended.
```

Команда **НЕ ДОЛЖНА**:

- непосредственно изменять поля Aggregate из transport handler;
- обходить проверку состояния;
- публиковать сообщение до commit;
- возвращать инфраструктурную модель хранения;
- молча исправлять недопустимое намерение клиента.

## 7.18. Репозитории и единица работы

Для каждого Aggregate Root определяется отдельный Repository интерфейс в доменном или прикладном порту.

Пример:

```go
// Псевдоконтракт; конкретная сигнатура определяется сервисом.
type ProjectRepository interface {
    Get(ctx context.Context, id ProjectID) (*Project, error)
    Save(ctx context.Context, project *Project, expected Version) error
}
```

Нормативные правила:

- Repository возвращает Aggregate, а не строку таблицы;
- Repository не содержит бизнес-решений;
- `Save` проверяет ожидаемую версию;
- сохранение Aggregate и Outbox выполняется атомарно;
- единица работы ограничена одним сервисом и одним хранилищем;
- чтение чужого Aggregate через прямой Repository **ЗАПРЕЩЕНО**;
- пакетные операции должны явно определять частичный успех и атомарность.

## 7.19. Удаление, обезличивание и исторические ссылки

Понятие «удалить» имеет разные предметные значения:

| Режим | Значение |
| --- | --- |
| Deactivate/Disable | Запретить дальнейшее использование, сохранив объект. |
| Soft Delete | Сделать объект недоступным для обычных операций, сохранив каноническую запись. |
| Deprovision | Удалить фактический внешний ресурс. |
| Anonymize | Удалить или заменить персональные атрибуты с сохранением технической идентичности истории. |
| Purge | Физически удалить данные по разрешённой политике хранения. |

Каждый Aggregate **ДОЛЖЕН** определить поддерживаемые режимы. Команда `Delete*` не должна использоваться без точного предметного результата.

Исторические Audit Event и Integration Event не переписываются при удалении объекта. Они могут содержать минимальную типизированную ссылку и обезличенное представление согласно политике.

## 7.20. Время и срок действия

- Все канонические времена хранятся в UTC.
- Временные интервалы должны явно определять включение границ.
- Истечение не должно зависеть только от периодического фонового задания: Query и Command обязаны учитывать текущее время через абстракцию Clock.
- Событие истечения MAY публиковаться асинхронно, но объект не считается действующим только потому, что событие ещё не обработано.
- Доменные тесты используют управляемый Clock.

## 7.21. Идемпотентность предметных операций

Идемпотентность определяется на уровне намерения, а не только HTTP-запроса.

Запись идемпотентности должна связывать:

```text
owner scope
operation/command type
actor or client
idempotency key
normalized request fingerprint
result resource or Operation
status
expiration
```

Если тот же ключ повторно используется с отличающимся значимым payload, сервис **ДОЛЖЕН** вернуть ошибку конфликта идемпотентности, а не выполнить новое намерение.

## 7.22. Модель ошибок предметной области

Доменные ошибки должны быть стабильными и пригодными для сопоставления с транспортным Error Model.

Основные категории:

| Категория | Пример |
| --- | --- |
| Not Found | `PROJECT_NOT_FOUND`. |
| Invalid State | `USER_NOT_ACTIVE`, `OPERATION_ALREADY_TERMINAL`. |
| Invariant Violation | `WORKSPACE_PARENT_REQUIRED`. |
| Conflict | `RESOURCE_NAME_ALREADY_EXISTS`, `CONCURRENT_MODIFICATION`. |
| Policy Denied | `AUTHENTICATION_METHOD_NOT_ALLOWED`. |
| Expired | `CHALLENGE_EXPIRED`, `ROLE_BINDING_EXPIRED`. |
| External Dependency | `PROVIDER_TEMPORARILY_UNAVAILABLE`. |
| Security | `ACCESS_DENIED`, без раскрытия существования защищённого объекта при необходимости. |

Domain Error **НЕ ДОЛЖНА** содержать HTTP status, gRPC code или текст конкретной базы данных. Mapping выполняется на транспортной границе.

## 7.23. Модель чтения

Read Model может денормализовать данные нескольких Aggregate и контекстов, но:

- не становится каноническим владельцем;
- должна указывать источник и свежесть;
- не используется для критической мутации без допустимого уровня консистентности;
- должна корректно обрабатывать удаление и изменение версии;
- может быть восстановлена из канонических данных и событий;
- не должна публиковать чувствительные данные шире исходных политик доступа.

Примеры:

- дерево Organization → Workspace → Project;
- карточка User с активными Membership;
- Access Explorer;
- состояние ManagedResource вместе с Operation;
- временная линия Audit.

## 7.24. Модель консистентности на уровне предметной области

| Сценарий | Требуемая консистентность |
| --- | --- |
| Инварианты одного Aggregate | Строгая локальная транзакционная. |
| Уникальность имени в scope | Строгая в пределах владельца и выбранного индекса/реестра. |
| Публикация Integration Event после мутации | Гарантированная доставка через Outbox, итоговая у потребителей. |
| Реакция Authentication на UserDisabled | Итоговая с определённым максимальным окном и механизмом немедленной проверки для критических сценариев. |
| Синхронизация Access со SpiceDB | Итоговая либо read-after-write через token согласно контракту. |
| Observed State ManagedResource | Итоговая, с явным временем наблюдения и generation. |
| Audit ingestion | Как минимум один раз с дедупликацией; обязательная обнаружимость недоставки. |
| Operation progress | Итоговая и монотонная по подтверждённым стадиям. |

## 7.25. Запрещённые модели

Следующие решения противоречат настоящей главе:

1. единый Aggregate `Organization`, содержащий все Workspace, Project, Service и пользователей;
2. общая таблица `resources`, изменяемая несколькими сервисами без предметного владельца;
3. сохранение SpiceDB tuple как публичной модели Access;
4. использование Keycloak realm/user/session как канонических доменных сущностей M8;
5. хранение состояния Temporal Workflow как единственного состояния ManagedResource;
6. моделирование Permission Check как сохраняемого Aggregate без предметной необходимости;
7. смешение User, Client, Service Account и Service в одном неразличимом типе;
8. изменение parent ID обычным update;
9. прямое каскадное удаление между базами сервисов;
10. объединение Operation State и Resource State;
11. хранение чужого Aggregate как вложенного изменяемого JSON;
12. использование Audit Log как источника актуального состояния;
13. публикация события до фиксации Aggregate;
14. повторное использование удалённого ID;
15. применение raw provider payload как Desired State без нормализации;
16. размещение предметных инвариантов только в SQL constraint или UI;
17. использование Labels как скрытой политики доступа;
18. общий Repository, позволяющий любому модулю записывать любой Aggregate.

## 7.26. Трассировка модели до требований и SPDD

Каждый Structured Prompt, изменяющий предметную модель, **ДОЛЖЕН** ссылаться на:

```yaml
traceability:
  pads:
    chapter: 7
    aggregates:
      - DM-AUTH-AUTHENTICATION-TRANSACTION
    invariants:
      - INV-AUTH-003
      - INV-AUTH-005
    state_transitions:
      - AUTHENTICATED_TO_HANDOFF_CREATED
  requirements:
    - AUTH-FR-017
  contracts:
    - auth.v1.AuthenticationService.StartAuthentication
```

Для устойчивых ссылок на агрегаты вводятся следующие ID:

| ID | Aggregate |
| --- | --- |
| DM-RM-ORG | Organization |
| DM-RM-WS | Workspace |
| DM-RM-PRJ | Project |
| DM-RM-SVC | ServiceRegistration |
| DM-ID-POOL | UserPool |
| DM-ID-USER | User |
| DM-ID-GROUP | Group |
| DM-ID-MEM | Membership |
| DM-AUTH-CLIENT | Client |
| DM-AUTH-PROVIDER | AuthenticationProviderConfiguration |
| DM-AUTH-TX | AuthenticationTransaction |
| DM-AUTH-SESSION | AuthenticationSession |
| DM-ACC-MODEL | AuthorizationModel |
| DM-ACC-ROLE | Role |
| DM-ACC-BINDING | RoleBinding |
| DM-ACC-REL | AccessRelationship |
| DM-RISK-POLICY | DecisionPolicy |
| DM-RISK-ASSESS | RiskAssessment |
| DM-PROV-DEF | ResourceDefinition |
| DM-PROV-PLACEMENT | PlacementPolicy |
| DM-PROV-RESOURCE | ManagedResource |
| DM-AUD-EVENT | AuditEvent |
| DM-AUD-RETENTION | RetentionPolicy |
| DM-AUD-EXPORT | AuditExportJob |
| DM-OPS-OPERATION | Operation |

## 7.27. Проверка соответствия модели предметной области

Изменение соответствует настоящей главе, если:

1. определён контекст-владелец изменяемого понятия;
2. указано, является объект Aggregate, Entity, Value Object, Query Result или Process;
3. граница атомарности ограничена одним Aggregate;
4. межагрегатные и межконтекстные ссылки типизированы;
5. инварианты имеют устойчивые ID и проверяются владельцем;
6. жизненный цикл и терминальные состояния определены;
7. Resource State отделён от Operation State;
8. внешняя технология скрыта Adapter или Driver;
9. конкурентная мутация защищена ResourceVersion;
10. мутация формирует Domain Event и Outbox атомарно;
11. обязательное действие формирует Audit Event;
12. удаление имеет однозначную предметную семантику;
13. идемпотентность определена для повторяемой команды;
14. ошибки не зависят от транспорта и хранилища;
15. Structured Prompt содержит ссылки на Aggregate и Invariant ID;
16. архитектурные тесты предотвращают прямой доступ к чужому Repository;
17. модель чтения не становится скрытым владельцем данных;
18. новый Aggregate не создаётся только ради таблицы, endpoint или фонового задания.

---
