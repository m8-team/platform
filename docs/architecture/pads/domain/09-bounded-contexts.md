---
title: "PADS: спецификации ограниченных контекстов"
description: "Нормативные спецификации контекстов Resource Manager, Identity, Authentication, Access, Risk Decision, Provisioning, Audit и Common Operation."
keywords:
  - "M8 Platform"
  - "PADS"
  - "architecture"
---

# 9. Спецификации ограниченных контекстов {#pads-bounded-contexts}

{% note info "Навигация PADS" %}

[Оглавление PADS](../index.md) | [Предыдущий раздел: 8. Карта контекстов](08-context-map.md) | [Следующий раздел: 10. Shared Kernel и общие контракты](10-shared-kernel.md)

{% endnote %}

## 9.1. Назначение главы

Настоящая глава определяет нормативные спецификации ограниченных контекстов M8 Platform. Для каждого контекста фиксируются:

- предметное назначение и граница ответственности;
- бизнес-возможности, реализуемые контекстом;
- принадлежащие контексту агрегаты, сущности и объекты-значения;
- инварианты и жизненные циклы;
- команды, запросы, API и события;
- входящие и исходящие зависимости;
- модель согласованности и владения данными;
- требования безопасности, аудита и наблюдаемости;
- поведение при отказах;
- запрещённые обязанности и зависимости;
- пространство идентификаторов требований;
- состав Context Prompt для SPDD.

Спецификация ограниченного контекста описывает **предметную границу**, а не структуру репозитория, процесса или Kubernetes Deployment. Один контекст в базовой архитектуре соответствует одному логическому сервису-владельцу, однако решение о физическом развертывании **МОЖЕТ** меняться без изменения предметной границы, если сохраняются владение моделью, данными и контрактами.

## 9.2. Нормативный шаблон контекста

Каждый ограниченный контекст **ДОЛЖЕН** иметь паспорт следующего вида:

| Поле | Назначение |
| --- | --- |
| `context_id` | Устойчивый идентификатор `CTX-*`. |
| `service_id` | Логический сервис-владелец. |
| `classification` | Core, Supporting или Generic Subdomain. |
| `mission` | Одно предложение, определяющее предметный результат контекста. |
| `owns` | Канонические предметные факты и агрегаты. |
| `does_not_own` | Явно исключённые обязанности. |
| `published_languages` | Контракты, доступные другим контекстам. |
| `upstream_dependencies` | Поставщики фактов и решений. |
| `downstream_consumers` | Потребители опубликованных контрактов. |
| `consistency_boundary` | Локальная транзакционная граница. |
| `requirement_namespace` | Префикс требований контекста. |

Для каждого контекста также **ДОЛЖНЫ** быть определены:

1. владелец каждой команды изменения состояния;
2. владелец каждого публичного запроса;
3. источник истины каждого публикуемого факта;
4. способ подтверждения полномочий;
5. обязательные аудиторские события;
6. политика идемпотентности;
7. способ публикации интеграционных событий;
8. правила обработки недоступности зависимостей;
9. метрики здоровья и предметные метрики;
10. минимальный набор контрактных и архитектурных тестов.

## 9.3. Общие правила для всех контекстов

### 9.3.1. Изоляция предметной модели

Внутренняя предметная модель контекста **НЕ ДОЛЖНА** импортировать:

- protobuf-сообщения;
- ConnectRPC-обработчики;
- типы YDB, Redis, Temporal, Keycloak или SpiceDB;
- модели хранения другого контекста;
- транспортные коды ошибок;
- внешние SDK поставщиков инфраструктуры.

Преобразование внешних контрактов во внутренние модели выполняется в Adapter или Anti-Corruption Layer.

### 9.3.2. Транзакционная граница

Локальная транзакция **ДОЛЖНА** ограничиваться хранилищем одного сервиса. Изменение агрегата и запись Outbox-сообщения выполняются атомарно. Изменение состояния двух контекстов одной распределённой транзакцией запрещено.

### 9.3.3. Публичные контракты

Публичный контракт контекста **ДОЛЖЕН** описывать предметный язык контекста и **НЕ ДОЛЖЕН** раскрывать:

- структуру таблиц;
- внутренние версии агрегатов, если они не нужны для concurrency control;
- идентификаторы workflow-задач;
- названия очередей и топиков как часть бизнес-контракта;
- специфичные типы внешнего поставщика;
- секреты, credential material и внутренние risk signals.

### 9.3.4. Идемпотентность

Все внешние команды, способные повторно поступить из-за retries, **ДОЛЖНЫ** поддерживать идемпотентность. Контекст хранит связь между областью идемпотентности, ключом, отпечатком команды и результатом обработки.

Повтор команды с тем же ключом и отличающимся значимым содержимым **ДОЛЖЕН** завершаться конфликтом идемпотентности.

### 9.3.5. События

Контекст публикует интеграционное событие только после фиксации соответствующего предметного факта. Событие **ДОЛЖНО** содержать `event_id`, `event_type`, `event_version`, `occurred_at`, `producer`, `subject`, `correlation_id`, `causation_id` и версию источника, когда она применима.

### 9.3.6. Безопасность

Каждый публичный вызов **ДОЛЖЕН**:

- аутентифицировать вызывающую сторону на доверенной границе;
- определить Actor, Subject и Client, когда эти роли применимы;
- выполнить авторизационную проверку через M8 Access или локально разрешённую bootstrap-политику;
- передать Project или другой Resource Scope явно;
- сформировать аудит решения и изменения;
- не включать чувствительные данные в логи и ошибки.

### 9.3.7. Наблюдаемость

Каждый контекст **ДОЛЖЕН** публиковать:

- технические RED-метрики API;
- метрики внешних зависимостей;
- метрики Outbox и Inbox lag;
- метрики конфликтов версий и идемпотентности;
- предметные метрики жизненных циклов;
- трассировки с сохранением correlation и causation;
- структурированные логи без секретов и персональных данных сверх необходимого минимума.

---

## 9.4. Resource Manager — управление ресурсной иерархией

### 9.4.1. Паспорт контекста

| Поле | Значение |
| --- | --- |
| Context ID | `CTX-RM` |
| Service ID | `m8-resource-manager` |
| Классификация | Core Domain |
| Пространство требований | `RM-*` |
| Миссия | Предоставлять каноническую, версионируемую и безопасную модель административной иерархии M8 Platform. |
| Владеет | Organization, Workspace, Project, ServiceRegistration, их состояниями, родительскими связями, именами, метками и версиями. |
| Не владеет | Пользователями, членством, ролями, аутентификацией, risk decisions, инфраструктурными экземплярами и аудит-хранилищем. |
| Consistency boundary | Один агрегат ресурса и его Outbox-записи в одной локальной транзакции. |
| Published Language | `PL-RESOURCE`, `PL-RESOURCE-EVENTS`. |

### 9.4.2. Назначение и ответственность

Resource Manager является единственным источником истины для структуры:

```text
Organization
└── Workspace
    └── Project
        └── ServiceRegistration
```

Контекст **ДОЛЖЕН**:

- создавать и изменять ресурсы иерархии;
- обеспечивать уникальность имён в определённой области;
- проверять допустимость родительской связи;
- управлять состояниями жизненного цикла;
- обеспечивать optimistic concurrency control;
- публиковать факты об изменении ресурсов;
- предоставлять разрешение типизированных `ResourceReference`;
- поддерживать управляемое удаление и, при необходимости, перемещение ресурса;
- обеспечивать пагинацию и фильтрацию списков;
- сохранять стабильность идентификатора при изменении отображаемого имени.

### 9.4.3. Исключённая ответственность

Resource Manager **НЕ ДОЛЖЕН**:

- хранить полный профиль пользователя;
- определять, является ли пользователь участником Project;
- хранить RoleBinding или AccessRelationship;
- выполнять challenge аутентификации;
- создавать Kafka topic, Kubernetes namespace или другой внешний ресурс напрямую;
- интерпретировать внутренние состояния Provisioning;
- принимать risk decision;
- быть владельцем общей аудиторской истории.

Требование, одновременно изменяющее ресурсную иерархию и членство, **ДОЛЖНО** быть разделено на команды Resource Manager и Identity либо оформлено как сквозной workflow с одним владельцем процесса.

### 9.4.4. Реализуемые бизнес-возможности

| Capability ID | Возможность | Роль контекста |
| --- | --- | --- |
| `CAP-RM-ORG` | Управление Organization | Владелец полного жизненного цикла. |
| `CAP-RM-WS` | Управление Workspace | Владелец полного жизненного цикла. |
| `CAP-RM-PRJ` | Управление Project | Владелец полного жизненного цикла. |
| `CAP-RM-SVC` | Регистрация Service | Владелец канонической регистрации. |
| `CAP-RM-HIER` | Навигация и проверка иерархии | Владелец связей parent/child. |
| `CAP-RM-LIFE` | Управляемое удаление и перемещение | Владелец предметного решения и Operation. |

### 9.4.5. Предметная модель

| Aggregate ID | Корень агрегата | Основные данные | Ключевые инварианты |
| --- | --- | --- | --- |
| `DM-RM-ORG` | Organization | ID, name, display name, labels, state, version | Организация является корнем; имя уникально в платформенной области; удаление управляемое. |
| `DM-RM-WS` | Workspace | ID, organization reference, name, labels, state, version | Workspace принадлежит одной Organization; parent immutable без Move operation. |
| `DM-RM-PRJ` | Project | ID, workspace reference, name, labels, state, version | Project принадлежит одному Workspace; Project является основной isolation scope. |
| `DM-RM-SVC` | ServiceRegistration | ID, project reference, service type, state, labels, version | Service зарегистрирован в одном Project; service type валиден по каталогу. |

Resource Manager **НЕ ДОЛЖЕН** загружать всю иерархию как один агрегат. Проверка существования родителя выполняется до создания дочернего агрегата, а межагрегатная согласованность поддерживается локальными индексами, событиями и контролируемыми lifecycle operations.

### 9.4.6. Команды

| Command ID | Команда | Результат |
| --- | --- | --- |
| `RM-CMD-001` | `CreateOrganization` | Organization или Operation, если требуется внешняя инициализация. |
| `RM-CMD-002` | `UpdateOrganization` | Обновлённая Organization с новой версией. |
| `RM-CMD-003` | `DeleteOrganization` | Operation управляемого удаления. |
| `RM-CMD-004` | `CreateWorkspace` | Workspace. |
| `RM-CMD-005` | `UpdateWorkspace` | Workspace с новой версией. |
| `RM-CMD-006` | `DeleteWorkspace` | Operation. |
| `RM-CMD-007` | `CreateProject` | Project. |
| `RM-CMD-008` | `UpdateProject` | Project с новой версией. |
| `RM-CMD-009` | `DeleteProject` | Operation. |
| `RM-CMD-010` | `RegisterService` | ServiceRegistration. |
| `RM-CMD-011` | `UpdateServiceRegistration` | ServiceRegistration с новой версией. |
| `RM-CMD-012` | `UnregisterService` | Operation или завершённый результат по политике удаления. |
| `RM-CMD-013` | `MoveResource` | Operation с проверкой допустимости нового родителя. |
| `RM-CMD-014` | `SetResourceLabels` | Новая версия ресурса. |
| `RM-CMD-015` | `RestoreResource` | Восстановленный ресурс, если policy допускает восстановление. |

Команда изменения **ДОЛЖНА** принимать ожидаемую версию или ETag для операций, где возможна конкуренция изменений.

### 9.4.7. Запросы

| Query ID | Запрос | Гарантия |
| --- | --- | --- |
| `RM-QRY-001` | `GetOrganization` | Актуальное состояние из источника истины. |
| `RM-QRY-002` | `ListOrganizations` | Стабильная пагинация. |
| `RM-QRY-003` | `GetWorkspace` | Актуальное состояние Workspace. |
| `RM-QRY-004` | `ListWorkspaces` | Фильтрация по Organization и состоянию. |
| `RM-QRY-005` | `GetProject` | Актуальное состояние Project. |
| `RM-QRY-006` | `ListProjects` | Фильтрация по Workspace/Organization и меткам. |
| `RM-QRY-007` | `GetServiceRegistration` | Каноническая регистрация сервиса. |
| `RM-QRY-008` | `ListServiceRegistrations` | Сервисы Project с пагинацией. |
| `RM-QRY-009` | `ResolveResourceReference` | Тип, состояние и ancestry ресурса. |
| `RM-QRY-010` | `GetResourceAncestors` | Канонический путь до корня. |

Списочные запросы **НЕ ДОЛЖНЫ** использовать нестабильную offset pagination для больших наборов. Page token связывается с порядком сортировки и фильтрами.

### 9.4.8. Публикуемые события

| Event ID | Событие | Минимальные данные |
| --- | --- | --- |
| `RM-EVT-001` | `OrganizationCreated` | resource, name, state, version. |
| `RM-EVT-002` | `OrganizationUpdated` | resource, changed fields, version. |
| `RM-EVT-003` | `WorkspaceCreated` | resource, organization reference, version. |
| `RM-EVT-004` | `ProjectCreated` | resource, workspace reference, version. |
| `RM-EVT-005` | `ServiceRegistered` | resource, project reference, service type, version. |
| `RM-EVT-006` | `ResourceStateChanged` | resource, previous state, current state, reason, version. |
| `RM-EVT-007` | `ResourceMoved` | resource, previous parent, current parent, version. |
| `RM-EVT-008` | `ResourceDeleted` | resource, deletion mode, tombstone version. |
| `RM-EVT-009` | `ResourceLabelsChanged` | resource, resulting labels, version. |

События Resource Manager являются источником локальных resource projections в Access, Identity, Authentication, Provisioning и Audit.

### 9.4.9. Потребляемые контракты

| Поставщик | Контракт | Назначение |
| --- | --- | --- |
| Access | `CheckPermission` | Проверка административных действий над ресурсом. |
| Provisioning | Lifecycle result events | Информирование о завершении внешней очистки при каскадном удалении. |
| Audit | Delivery acknowledgement, при необходимости | Эксплуатационный контроль надёжной доставки, не бизнес-зависимость. |

Resource Manager **НЕ ДОЛЖЕН** синхронно вызывать Access из обработчика события Access, чтобы не формировать цикл.

### 9.4.10. Согласованность и удаление

Создание и простое изменение одного ресурса выполняются синхронно. Удаление ресурса с дочерними объектами, активными memberships, access bindings или managed resources выполняется как Operation.

Перед финализацией удаления Resource Manager **ДОЛЖЕН**:

1. перевести ресурс в состояние, запрещающее новые зависимые объекты;
2. опубликовать факт начала управляемого удаления;
3. дождаться завершения обязательных cleanup steps через workflow;
4. сохранить tombstone, достаточный для исторических ссылок;
5. опубликовать `ResourceDeleted` только после фиксации конечного состояния.

### 9.4.11. Безопасность и аудит

Обязательные проверки включают:

- permission на создание дочернего ресурса у родителя;
- permission на изменение и удаление конкретного ресурса;
- step-up для особо чувствительных операций по политике Risk Decision;
- запрет использования удалённого или suspended Project как активной scope;
- защиту от confused deputy через явные Actor, Client и Resource Scope.

Каждая мутация **ДОЛЖНА** формировать AuditIntent с actor, target, command, result, changed fields и correlation metadata.

### 9.4.12. Наблюдаемость и SLO

Минимальные предметные метрики:

- количество активных Organization, Workspace, Project и ServiceRegistration;
- частота создания и удаления ресурсов;
- количество конфликтов версий;
- длительность lifecycle operations;
- число ресурсов в переходных и ошибочных состояниях;
- lag публикации resource events;
- число orphan projections, выявленных reconciliation-проверкой.

Целевые значения latency и availability задаются в Quality Attributes, но `Get*` и обычные команды изменения **ДОЛЖНЫ** проектироваться как online control-plane операции.

### 9.4.13. Отказы и деградация

| Отказ | Поведение |
| --- | --- |
| Access недоступен | Новая чувствительная мутация отклоняется `UNAVAILABLE`; чтение может продолжаться по локальной policy только для явно разрешённых системных сценариев. |
| Event transport недоступен | Команда фиксирует агрегат и Outbox; публикация возобновляется асинхронно. |
| Provisioning cleanup недоступен | Delete Operation остаётся pending/blocked, ресурс не считается удалённым. |
| Конфликт версии | Команда отклоняется как concurrency conflict без автоматического перезаписывания. |
| Повтор команды | Возвращается ранее сохранённый результат при совпадении отпечатка. |

### 9.4.14. Запрещённые зависимости

Resource Manager **НЕ ДОЛЖЕН**:

- импортировать SDK Keycloak или SpiceDB в domain/application;
- читать таблицы Identity или Provisioning;
- делать Access владельцем ancestry;
- хранить Membership как часть Project;
- использовать Audit как синхронную транзакционную зависимость;
- завершать удаление до подтверждения обязательных cleanup steps.

### 9.4.15. SPDD Context Prompt

Context Prompt `CTX-RM` **ДОЛЖЕН** включать:

- иерархию ресурсов и все `DM-RM-*`;
- инварианты `INV-RM-*`;
- команды `RM-CMD-*`, запросы `RM-QRY-*` и события `RM-EVT-*`;
- запрет владения Membership, RoleBinding и ManagedResource;
- правила optimistic locking, Outbox и lifecycle Operation;
- обязательные Access и Audit hooks;
- архитектурные тесты, запрещающие внешние SDK в domain/application.

---

## 9.5. Identity — управление идентичностями

### 9.5.1. Паспорт контекста

| Поле | Значение |
| --- | --- |
| Context ID | `CTX-ID` |
| Service ID | `m8-identity` |
| Классификация | Core Domain |
| Пространство требований | `ID-*` |
| Миссия | Управлять устойчивой идентичностью субъектов независимо от механизмов аутентификации и принятия решений доступа. |
| Владеет | UserPool, User, Group, Membership, ExternalIdentity, профильными атрибутами и состоянием жизненного цикла идентичности. |
| Не владеет | Credential secrets, AuthenticationChallenge, токенами, RoleBinding, AccessRelationship, risk decisions. |
| Consistency boundary | Один identity aggregate и его Outbox-записи. |
| Published Language | `PL-IDENTITY`, `PL-IDENTITY-EVENTS`. |

### 9.5.2. Назначение и ответственность

Identity **ДОЛЖЕН**:

- создавать и изолировать User Pool;
- создавать и сопровождать User;
- связывать пользователя с внешними идентичностями;
- управлять Group и Membership;
- разрешать SubjectReference по поддерживаемым идентификаторам;
- обеспечивать уникальность issuer/subject и других нормализованных ключей;
- управлять блокировкой, деактивацией, восстановлением и обезличиванием;
- публиковать факты, необходимые Authentication и Access;
- отделять business identity от authentication credential.

### 9.5.3. Исключённая ответственность

Identity **НЕ ДОЛЖЕН**:

- проверять пароль, OTP, passkey или подтверждение CIBA;
- выпускать access/refresh token;
- принимать решение `ALLOW` или `DENY` для доступа к ресурсу;
- хранить SpiceDB relationships как источник истины;
- владеть Organization, Workspace или Project;
- раскрывать Authentication внутренние персональные данные, не необходимые для subject resolution.

### 9.5.4. Реализуемые бизнес-возможности

| Capability ID | Возможность |
| --- | --- |
| `CAP-ID-POOL` | Управление User Pool и его политиками идентичности. |
| `CAP-ID-USER` | Жизненный цикл User. |
| `CAP-ID-EXT` | Связывание и отвязывание ExternalIdentity. |
| `CAP-ID-GRP` | Управление Group и Group Membership. |
| `CAP-ID-MEM` | Membership субъекта в Organization, Workspace или Project. |
| `CAP-ID-RESOLVE` | Разрешение SubjectReference. |
| `CAP-ID-PRIV` | Обезличивание, экспорт и ограничение профильных данных. |

### 9.5.5. Предметная модель

| Aggregate ID | Корень | Назначение |
| --- | --- | --- |
| `DM-ID-POOL` | UserPool | Изоляция пространства пользователей, namespaces и identity policies. |
| `DM-ID-USER` | User | Устойчивая идентичность, профиль, состояние и external identities. |
| `DM-ID-GROUP` | Group | Именованная группа субъектов внутри User Pool или допустимой scope. |
| `DM-ID-MEMBERSHIP` | Membership | Участие Subject в Resource Scope с типом и состоянием членства. |

CredentialReference **МОЖЕТ** храниться как непрозрачная ссылка на внешний credential provider, но секретный материал **НЕ ДОЛЖЕН** становиться частью агрегата User.

### 9.5.6. Команды

| Command ID | Команда |
| --- | --- |
| `ID-CMD-001` | `CreateUserPool` |
| `ID-CMD-002` | `UpdateUserPool` |
| `ID-CMD-003` | `DeleteUserPool` |
| `ID-CMD-004` | `CreateUser` |
| `ID-CMD-005` | `UpdateUserProfile` |
| `ID-CMD-006` | `DisableUser` |
| `ID-CMD-007` | `EnableUser` |
| `ID-CMD-008` | `AnonymizeUser` |
| `ID-CMD-009` | `LinkExternalIdentity` |
| `ID-CMD-010` | `UnlinkExternalIdentity` |
| `ID-CMD-011` | `CreateGroup` |
| `ID-CMD-012` | `AddGroupMember` |
| `ID-CMD-013` | `RemoveGroupMember` |
| `ID-CMD-014` | `AssignMembership` |
| `ID-CMD-015` | `ChangeMembershipState` |
| `ID-CMD-016` | `RevokeMembership` |
| `ID-CMD-017` | `MergeUserIdentities` через управляемую Operation |

### 9.5.7. Запросы

| Query ID | Запрос | Назначение |
| --- | --- | --- |
| `ID-QRY-001` | `GetUserPool` | Получение политики и состояния User Pool. |
| `ID-QRY-002` | `GetUser` | Получение User по каноническому ID. |
| `ID-QRY-003` | `SearchUsers` | Ограниченный поиск по нормализованным разрешённым атрибутам. |
| `ID-QRY-004` | `ResolveSubject` | Разрешение email, phone, username или external subject в SubjectReference. |
| `ID-QRY-005` | `ListGroups` | Список групп с пагинацией. |
| `ID-QRY-006` | `ListGroupMembers` | Состав группы. |
| `ID-QRY-007` | `ListMemberships` | Membership по субъекту или Resource Scope. |
| `ID-QRY-008` | `GetIdentityStatus` | Минимальный статус для Authentication и Access. |

`ResolveSubject` **ДОЛЖЕН** возвращать минимальный результат и не раскрывать, существует ли пользователь, внешнему недоверенному клиенту без соответствующего permission.

### 9.5.8. Публикуемые события

| Event ID | Событие |
| --- | --- |
| `ID-EVT-001` | `UserPoolCreated` |
| `ID-EVT-002` | `UserCreated` |
| `ID-EVT-003` | `UserProfileChanged` |
| `ID-EVT-004` | `UserStateChanged` |
| `ID-EVT-005` | `UserAnonymized` |
| `ID-EVT-006` | `ExternalIdentityLinked` |
| `ID-EVT-007` | `ExternalIdentityUnlinked` |
| `ID-EVT-008` | `GroupCreated` |
| `ID-EVT-009` | `GroupMembershipChanged` |
| `ID-EVT-010` | `MembershipAssigned` |
| `ID-EVT-011` | `MembershipStateChanged` |
| `ID-EVT-012` | `MembershipRevoked` |
| `ID-EVT-013` | `UserIdentitiesMerged` |

События, доступные широкому кругу потребителей, **ДОЛЖНЫ** содержать только минимально необходимый набор персональных данных. Для большинства интеграций достаточно SubjectReference, статуса и source version.

### 9.5.9. Потребляемые контракты

| Поставщик | Контракт | Назначение |
| --- | --- | --- |
| Resource Manager | Resource events / ResolveResourceReference | Проверка существования и состояния scope для Membership. |
| Access | CheckPermission | Авторизация административных действий. |
| Authentication provider adapter | Credential/external account lifecycle callbacks | Синхронизация непрозрачных ссылок, если предусмотрено ADR. |

Identity **НЕ ДОЛЖЕН** принимать профильные изменения из Authentication как канонические без отдельной подтверждённой команды или trusted provisioning contract.

### 9.5.10. Уникальность и нормализация

Нормализация email, phone и username **ДОЛЖНА** быть версионируемой и привязана к User Pool policy. Уникальность определяется явно:

- `issuer + external_subject` — глобально либо в заданной provider scope;
- нормализованный username — внутри User Pool;
- email/phone — по политике User Pool, а не неявно глобально;
- Membership — уникально по Subject, Resource Scope и membership type.

Изменение политики уникальности, влияющее на существующие данные, требует migration plan и ADR.

### 9.5.11. Жизненный цикл и приватность

Состояния User включают как минимум `ACTIVE`, `DISABLED`, `LOCKED`, `ANONYMIZED`, `DELETED_TOMBSTONE`. Физическое удаление **НЕ ДОЛЖНО** разрушать исторические ссылки Audit.

Обезличивание должно:

1. удалить или заменить профильные атрибуты;
2. отвязать идентификаторы по policy;
3. сохранить непрозрачный стабильный subject ID;
4. опубликовать `UserAnonymized`;
5. не позволять восстановить удалённые данные из обычных операционных логов и событий.

### 9.5.12. Безопасность и аудит

Особо чувствительные операции:

- link/unlink external identity;
- merge identities;
- anonymize user;
- восстановление disabled user;
- изменение Membership с административными последствиями.

Для них **СЛЕДУЕТ** применять step-up и risk evaluation по policy. Все изменения ExternalIdentity и Membership **ДОЛЖНЫ** включать previous/current state в AuditChangeSet без раскрытия секретов.

### 9.5.13. Наблюдаемость

Минимальные метрики:

- active/disabled/anonymized users;
- число User Pool;
- частота subject resolution и доля ambiguous/not found;
- конфликты уникальности;
- external identity link failures;
- lag identity events;
- длительность identity merge/anonymization operations;
- количество membership inconsistencies с Resource Manager projections.

### 9.5.14. Отказы и деградация

| Отказ | Поведение |
| --- | --- |
| Resource Manager недоступен | Новое Membership не создаётся без подтверждённой локальной проекции допустимого ресурса; чувствительная операция fail closed. |
| External identity provider недоступен | Link operation остаётся pending или завершается retryable error; локальная связь не объявляется активной преждевременно. |
| Event transport недоступен | Состояние и Outbox фиксируются атомарно. |
| Дубликат external subject | Команда завершается conflict; автоматический merge запрещён. |
| Неоднозначный subject resolution | Возвращается предметная ошибка ambiguity, а не произвольный User. |

### 9.5.15. Запрещённые зависимости

Identity **НЕ ДОЛЖЕН**:

- хранить пароли, OTP secrets и private keys;
- считать успешный login доказательством права доступа;
- изменять Resource Manager hierarchy;
- напрямую записывать SpiceDB;
- раскрывать полный профиль в identity events по умолчанию;
- объединять пользователей без явного workflow, аудита и criteria.

### 9.5.16. SPDD Context Prompt

Context Prompt `CTX-ID` **ДОЛЖЕН** включать:

- `UserPool`, `User`, `Group`, `Membership`, `ExternalIdentity`;
- правила нормализации и уникальности;
- privacy constraints и минимизацию данных;
- запрет владения credentials, tokens и access decisions;
- события `ID-EVT-*` и правила PII redaction;
- обязательные test cases для duplicate identity, ambiguous resolution, anonymization и idempotency.

---

## 9.6. Authentication — выполнение аутентификации

### 9.6.1. Паспорт контекста

| Поле | Значение |
| --- | --- |
| Context ID | `CTX-AUTHN` |
| Service ID | `m8-authentication` |
| Классификация | Core Domain |
| Пространство требований | `AUTH-*` |
| Миссия | Управлять проверяемым процессом подтверждения идентичности и уровня уверенности, не присваивая субъекту полномочия. |
| Владеет | Client, AuthenticationTransaction, AuthenticationChallenge, AuthenticationSession, Handoff, requested/achieved assurance. |
| Не владеет | Каноническим профилем User, ролями, permission graph, risk policy и audit storage. |
| Consistency boundary | Одна AuthenticationTransaction/Session и Outbox в локальной транзакции; внешние шаги оркестрируются отдельно. |
| Published Language | `PL-AUTHENTICATION`, `PL-AUTHENTICATION-EVENTS`. |

### 9.6.2. Назначение и ответственность

Authentication **ДОЛЖЕН**:

- запускать транзакцию аутентификации;
- разрешать Subject через Identity;
- запрашивать Risk Decision;
- выбирать или предлагать допустимые challenge methods;
- управлять challenge lifecycle;
- интегрироваться с Keycloak и другими providers через ACL;
- фиксировать requested и achieved assurance level;
- создавать безопасный Handoff;
- поддерживать cancel, expire, resend, retry и re-authentication;
- связывать AuthenticationSession с Client и Subject;
- публиковать минимальные факты об исходе аутентификации.

Успешная Authentication **НЕ ОЗНАЧАЕТ** разрешение выполнить конкретное действие. Авторизация остаётся ответственностью Access.

### 9.6.3. Исключённая ответственность

Authentication **НЕ ДОЛЖЕН**:

- изменять профиль User;
- назначать Role или Membership;
- самостоятельно определять business permission;
- хранить password/OTP/passkey secret вне provider-specific secure storage;
- копировать Keycloak session model в предметную модель;
- возвращать клиенту внутренние risk signals;
- считать refresh token единственным средством восстановления контекста при невозможности его использовать.

### 9.6.4. Реализуемые бизнес-возможности

| Capability ID | Возможность |
| --- | --- |
| `CAP-AUTH-START` | Запуск AuthenticationTransaction. |
| `CAP-AUTH-CHALLENGE` | Выбор, выполнение, resend и fallback challenge. |
| `CAP-AUTH-CIBA` | Backchannel authentication через CIBA. |
| `CAP-AUTH-WEBAUTHN` | Passkey/WebAuthn challenge через adapter. |
| `CAP-AUTH-FED` | OIDC/SAML federation через ACL. |
| `CAP-AUTH-STEPUP` | Повышение assurance level. |
| `CAP-AUTH-REAUTH` | Новая аутентификация при невозможности refresh. |
| `CAP-AUTH-HANDOFF` | Безопасная передача результата. |
| `CAP-AUTH-SESSION` | Управление AuthenticationSession и её состоянием. |

### 9.6.5. Предметная модель

| Aggregate ID | Корень | Назначение |
| --- | --- | --- |
| `DM-AUTH-CLIENT` | Client | Политика допустимых flows, redirect/handoff, assurance и provider configuration reference. |
| `DM-AUTH-TX` | AuthenticationTransaction | Полный предметный жизненный цикл одной попытки аутентификации. |
| `DM-AUTH-SESSION` | AuthenticationSession | Результат подтверждённой идентичности и assurance с ограниченным сроком. |
| `DM-AUTH-PROVIDER` | AuthenticationProviderRegistration | Доменная регистрация provider без утечки vendor types. |

`AuthenticationChallenge` является сущностью внутри AuthenticationTransaction, если его жизненный цикл не требует независимой адресации. Provider callback связывается с challenge через непрозрачный correlation reference.

### 9.6.6. Состояния AuthenticationTransaction

```text
CREATED
  → SUBJECT_RESOLVED
  → DECISION_PENDING
  → CHALLENGE_REQUIRED
  → CHALLENGE_PENDING
  → AUTHENTICATED
  → HANDOFF_CREATED
  → COMPLETED

Конечные альтернативы:
FAILED | DENIED | CANCELLED | EXPIRED
```

Переходы **ДОЛЖНЫ** быть явными и проверяться агрегатом. Callback от provider **НЕ ДОЛЖЕН** напрямую устанавливать `COMPLETED`, минуя проверку текущего состояния, assurance и anti-replay.

### 9.6.7. Команды

| Command ID | Команда |
| --- | --- |
| `AUTH-CMD-001` | `RegisterClient` |
| `AUTH-CMD-002` | `UpdateClient` |
| `AUTH-CMD-003` | `StartAuthentication` |
| `AUTH-CMD-004` | `SelectChallenge` |
| `AUTH-CMD-005` | `ResendChallenge` |
| `AUTH-CMD-006` | `CompleteChallenge` |
| `AUTH-CMD-007` | `HandleProviderCallback` |
| `AUTH-CMD-008` | `CancelAuthentication` |
| `AUTH-CMD-009` | `ExpireAuthentication` |
| `AUTH-CMD-010` | `CreateAuthenticationHandoff` |
| `AUTH-CMD-011` | `ExchangeHandoff` |
| `AUTH-CMD-012` | `StartStepUpAuthentication` |
| `AUTH-CMD-013` | `StartReauthentication` |
| `AUTH-CMD-014` | `RevokeAuthenticationSession` |

`ResendChallenge` **ДОЛЖЕН** учитывать cooldown, attempt limit и Risk Decision. Повторный callback обрабатывается идемпотентно.

### 9.6.8. Запросы

| Query ID | Запрос |
| --- | --- |
| `AUTH-QRY-001` | `GetClient` |
| `AUTH-QRY-002` | `GetAuthentication` |
| `AUTH-QRY-003` | `GetCurrentChallenge` |
| `AUTH-QRY-004` | `GetAuthenticationSession` |
| `AUTH-QRY-005` | `ListAvailableAuthenticationMethods` |
| `AUTH-QRY-006` | `GetAuthenticationOperation` |

Ответы **НЕ ДОЛЖНЫ** раскрывать provider secret, raw token, internal risk score или диагностические данные, позволяющие обходить challenge.

### 9.6.9. Публикуемые события

| Event ID | Событие |
| --- | --- |
| `AUTH-EVT-001` | `AuthenticationStarted` |
| `AUTH-EVT-002` | `AuthenticationMethodSelected` |
| `AUTH-EVT-003` | `ChallengeRequired` |
| `AUTH-EVT-004` | `ChallengeDispatched` |
| `AUTH-EVT-005` | `ChallengeCompleted` |
| `AUTH-EVT-006` | `AuthenticationCompleted` |
| `AUTH-EVT-007` | `AuthenticationDenied` |
| `AUTH-EVT-008` | `AuthenticationFailed` |
| `AUTH-EVT-009` | `AuthenticationCancelled` |
| `AUTH-EVT-010` | `AuthenticationExpired` |
| `AUTH-EVT-011` | `AuthenticationSessionCreated` |
| `AUTH-EVT-012` | `AuthenticationSessionRevoked` |
| `AUTH-EVT-013` | `StepUpCompleted` |

Событие `AuthenticationCompleted` содержит SubjectReference, ClientReference, achieved assurance, authentication methods и timestamps, но **НЕ ДОЛЖНО** содержать credential material или полный token set.

### 9.6.10. Потребляемые контракты

| Поставщик | Контракт | Назначение |
| --- | --- | --- |
| Identity | `ResolveSubject`, `GetIdentityStatus` | Разрешение и проверка состояния Subject. |
| Risk Decision | `EvaluateAuthenticationRisk` | ALLOW, DENY, CHALLENGE или REVIEW; требуемый assurance. |
| Access | `CheckPermission` | Проверка права Client/Actor запускать административные или delegated flows. |
| Resource Manager | Resource projection | Проверка активности Project и ServiceRegistration. |
| Keycloak / provider | CIBA, OIDC, SAML, WebAuthn, OTP adapters | Исполнение технического механизма challenge. |

### 9.6.11. Provider ACL

ACL для Keycloak **ДОЛЖЕН**:

- преобразовывать Keycloak session/challenge state в доменные состояния;
- скрывать provider-specific error codes за M8 error taxonomy;
- не допускать передачи Keycloak model в domain/application;
- обеспечивать anti-replay и callback validation;
- хранить provider identifiers только как непрозрачные references;
- иметь contract tests против поддерживаемой версии provider.

Замена Keycloak **НЕ ДОЛЖНА** требовать изменения публичного Authentication API, кроме явно provider-specific administrative capabilities.

### 9.6.12. Assurance и step-up

Authentication различает:

- `requested_assurance_level` — уровень, требуемый Client или downstream policy;
- `required_assurance_level` — итоговое требование Risk Decision;
- `achieved_assurance_level` — фактически доказанный уровень;
- `authentication_methods` — применённые методы.

Транзакция не может завершиться успешно, если achieved assurance ниже required assurance. Step-up создаёт новую связанную AuthenticationTransaction или отдельную подоперацию по принятой модели, но **НЕ ДОЛЖЕН** молча изменять исторический результат предыдущей транзакции.

### 9.6.13. Re-authentication и refresh failure

Если refresh token отсутствует, истёк, отозван или provider отклонил его, Authentication **ДОЛЖЕН** запускать новую независимую аутентификацию. Старая транзакция не переоткрывается, а прежняя session не считается восстановленной автоматически.

### 9.6.14. Безопасность и аудит

Обязательны:

- rate limit и velocity checks;
- anti-replay для callback и handoff;
- одноразовость Handoff;
- binding Handoff к Client и ожидаемому redirect/channel;
- защита от subject enumeration;
- ограниченный TTL всех временных артефактов;
- masking email/phone в пользовательских ответах;
- audit без OTP, token, assertion и biometric data.

### 9.6.15. Наблюдаемость

Минимальные метрики:

- authentication starts/completions/failures по method и client;
- conversion по состояниям;
- challenge dispatch/completion latency;
- resend и fallback rate;
- provider callback errors;
- risk decision latency/outcomes;
- achieved assurance distribution;
- expired/cancelled transactions;
- handoff replay attempts;
- reauthentication rate после refresh failure.

### 9.6.16. Отказы и деградация

| Отказ | Поведение |
| --- | --- |
| Identity недоступен | Новая транзакция не проходит subject resolution; retryable failure без создания ложной session. |
| Risk Decision недоступен | По умолчанию fail closed; явно разрешённая low-risk policy может быть оформлена ADR. |
| Provider недоступен | Transaction остаётся в retryable/pending состоянии либо предлагает разрешённый fallback. |
| Access недоступен | Административные операции fail closed. |
| Callback повторён | Идемпотентный no-op или возврат сохранённого результата. |
| Outbox transport недоступен | Завершённое локальное состояние сохраняется, событие публикуется после восстановления. |

### 9.6.17. Запрещённые зависимости

Authentication **НЕ ДОЛЖЕН**:

- считать Identity базой credentials;
- читать Keycloak DB;
- выдавать permission decision;
- записывать RoleBinding;
- хранить raw refresh/access tokens в логах или Audit;
- принимать achieved assurance только со слов Client;
- повторно использовать AuthenticationTransaction после конечного состояния.

### 9.6.18. SPDD Context Prompt

Context Prompt `CTX-AUTHN` **ДОЛЖЕН** включать:

- state machine AuthenticationTransaction;
- определения Subject, Actor, Client и Session;
- requested/required/achieved assurance;
- разрешённые методы и fallback rules;
- зависимости Identity, Risk Decision, Access и provider ACL;
- anti-replay, idempotency, TTL и rate-limit constraints;
- запрет утечки provider models и secret material;
- обязательные unit, contract, integration и security test cases.

---

## 9.7. Access — управление полномочиями

### 9.7.1. Паспорт контекста

| Поле | Значение |
| --- | --- |
| Context ID | `CTX-ACC` |
| Service ID | `m8-access` |
| Классификация | Core Domain |
| Пространство требований | `ACC-*` |
| Миссия | Предоставлять единый язык полномочий и проверяемое решение о возможности Subject выполнить Action над Resource. |
| Владеет | AuthorizationModel, Permission, Role, RoleBinding, AccessRelationship, CheckResult и Explanation. |
| Не владеет | Доказательством аутентификации, профилем User, ресурсной иерархией, risk policy и audit storage. |
| Consistency boundary | Изменение одного authorization aggregate и transactional dispatch; graph storage обновляется через согласованный adapter path. |
| Published Language | `PL-ACCESS`, `PL-ACCESS-EVENTS`. |

### 9.7.2. Назначение и ответственность

Access **ДОЛЖЕН**:

- определять канонические Permission и Resource Type;
- управлять Role и RoleBinding;
- управлять отношениями Subject–Resource;
- выполнять `CheckPermission`;
- объяснять результат в безопасной форме;
- поддерживать batch check и simulation;
- синхронизировать модель с SpiceDB через adapter;
- обрабатывать resource и identity projections;
- обеспечивать ревизию и отзыв полномочий.

### 9.7.3. Исключённая ответственность

Access **НЕ ДОЛЖЕН**:

- считать наличие аутентификации достаточным основанием доступа;
- владеть Organization/Workspace/Project;
- изменять User или Membership напрямую;
- исполнять challenge step-up;
- определять риск операции вместо Risk Decision;
- возвращать неограниченное внутреннее объяснение внешнему пользователю.

### 9.7.4. Реализуемые бизнес-возможности

| Capability ID | Возможность |
| --- | --- |
| `CAP-ACC-MODEL` | Управление AuthorizationModel. |
| `CAP-ACC-PERM` | Каталог Permission. |
| `CAP-ACC-ROLE` | Управление Role. |
| `CAP-ACC-BIND` | RoleBinding и прямые отношения. |
| `CAP-ACC-CHECK` | Permission check и batch check. |
| `CAP-ACC-EXPLAIN` | Безопасное объяснение решения. |
| `CAP-ACC-SIM` | Симуляция и what-if analysis. |
| `CAP-ACC-REVIEW` | Access review и отзыв избыточных полномочий. |

### 9.7.5. Предметная модель

| Aggregate ID | Корень | Назначение |
| --- | --- | --- |
| `DM-ACC-MODEL` | AuthorizationModel | Версионируемая схема типов ресурсов, отношений и permissions. |
| `DM-ACC-ROLE` | Role | Именованный набор permissions и применимая scope. |
| `DM-ACC-BINDING` | RoleBinding | Назначение Role субъекту или группе в Resource Scope. |
| `DM-ACC-REL` | AccessRelationship | Типизированное прямое отношение Subject–Resource. |

Результат `CheckPermission` является decision record, но не обязательно долгоживущим агрегатом. Для чувствительных операций **МОЖЕТ** сохраняться decision evidence, достаточное для Audit и расследования.

### 9.7.6. Команды

| Command ID | Команда |
| --- | --- |
| `ACC-CMD-001` | `PublishAuthorizationModel` |
| `ACC-CMD-002` | `CreateRole` |
| `ACC-CMD-003` | `UpdateRole` |
| `ACC-CMD-004` | `DeleteRole` |
| `ACC-CMD-005` | `BindRole` |
| `ACC-CMD-006` | `RevokeRoleBinding` |
| `ACC-CMD-007` | `WriteRelationship` |
| `ACC-CMD-008` | `DeleteRelationship` |
| `ACC-CMD-009` | `StartAccessReview` |
| `ACC-CMD-010` | `ApplyAccessReviewDecision` |

### 9.7.7. Запросы и решения

| Query ID | Запрос |
| --- | --- |
| `ACC-QRY-001` | `CheckPermission` |
| `ACC-QRY-002` | `BatchCheckPermissions` |
| `ACC-QRY-003` | `ExplainPermission` |
| `ACC-QRY-004` | `ListRoles` |
| `ACC-QRY-005` | `ListRoleBindings` |
| `ACC-QRY-006` | `ReadRelationships` |
| `ACC-QRY-007` | `LookupResources` |
| `ACC-QRY-008` | `LookupSubjects` |
| `ACC-QRY-009` | `SimulatePermission` |

`CheckPermission` **ДОЛЖЕН** возвращать как минимум `ALLOW` или `DENY`, evaluated model version, decision timestamp и безопасный reason code. `UNKNOWN` не должен неявно трактоваться как `ALLOW`.

### 9.7.8. Публикуемые события

| Event ID | Событие |
| --- | --- |
| `ACC-EVT-001` | `AuthorizationModelPublished` |
| `ACC-EVT-002` | `RoleCreated` |
| `ACC-EVT-003` | `RoleChanged` |
| `ACC-EVT-004` | `RoleDeleted` |
| `ACC-EVT-005` | `RoleBindingCreated` |
| `ACC-EVT-006` | `RoleBindingRevoked` |
| `ACC-EVT-007` | `AccessRelationshipWritten` |
| `ACC-EVT-008` | `AccessRelationshipDeleted` |
| `ACC-EVT-009` | `AccessReviewCompleted` |

### 9.7.9. Потребляемые события и проекции

| Источник | Факт | Использование |
| --- | --- | --- |
| Resource Manager | Resource created/state changed/deleted | Resource type, ancestry и active status projection. |
| Identity | User/Group/Membership state changed | Subject и group projection, отзыв полномочий disabled subject. |
| Risk Decision | Policy metadata, если требуется | Не заменяет базовый permission check; используется для orchestration совместно с вызывающим контекстом. |

Access **ДОЛЖЕН** уметь пересобрать projections из Published Language без прямого чтения чужих баз.

### 9.7.10. SpiceDB Adapter

SpiceDB является реализационной технологией, а не предметным владельцем. Adapter **ДОЛЖЕН**:

- переводить M8 ResourceReference и SubjectReference в SpiceDB object/subject;
- скрывать zed token и vendor errors;
- обеспечивать schema compatibility checks;
- поддерживать consistency token там, где это нужно для read-your-writes;
- не допускать прямого использования SpiceDB SDK вне adapter/infrastructure;
- иметь reconciliation между каноническими Access writes и graph storage.

### 9.7.11. Согласованность решений

Для administrative control-plane операций после изменения RoleBinding **СЛЕДУЕТ** обеспечивать read-your-writes с помощью revision/consistency token. Для массовых lookup допускается bounded staleness, если это явно обозначено в контракте.

Кэш положительных решений **ДОЛЖЕН** иметь короткий TTL и учитывать model/relation version. Кэш отрицательных решений **МОЖЕТ** использоваться осторожно, но не должен препятствовать быстрому предоставлению нового доступа.

### 9.7.12. Безопасность и аудит

Изменение AuthorizationModel, системных Role и privileged bindings требует усиленного permission и, по policy, step-up. Audit **ДОЛЖЕН** фиксировать:

- кто изменил полномочия;
- кому предоставлено или отозвано право;
- Resource Scope;
- Role/Relationship;
- предыдущую и новую версию;
- reason/ticket, если policy требует обоснование.

`ExplainPermission` внешнему пользователю **НЕ ДОЛЖЕН** раскрывать полный graph path, внутренние группы или сведения о других субъектах.

### 9.7.13. Наблюдаемость

Минимальные метрики:

- check rate и p50/p95/p99 latency;
- allow/deny/error ratio;
- SpiceDB dependency latency и error rate;
- relation write lag;
- model publication failures;
- stale projection count;
- reconciliation mismatch;
- privileged binding count;
- access review findings и revoke lead time.

### 9.7.14. Отказы и деградация

| Отказ | Поведение |
| --- | --- |
| SpiceDB недоступен | Security-sensitive checks fail closed; публичное API возвращает retryable unavailable. |
| Projection отстаёт | Решение помечается соответствующей revision; удалённые/disabled subjects должны отзываться приоритетно. |
| Resource unknown | DENY или NOT_FOUND по контракту без раскрытия лишней информации. |
| Identity event дублирован | Inbox/idempotent projection update. |
| Model incompatible | Новая версия не публикуется; предыдущая active version сохраняется. |

### 9.7.15. Запрещённые зависимости

Access **НЕ ДОЛЖЕН**:

- читать Resource Manager/Identity DB;
- использовать email как канонический subject key;
- возвращать ALLOW при ошибке graph backend;
- смешивать risk score с базовым permission graph;
- позволять сервисам писать отношения напрямую в SpiceDB в обход M8 Access;
- сохранять provider-specific zed token в публичной доменной модели.

### 9.7.16. SPDD Context Prompt

Context Prompt `CTX-ACC` **ДОЛЖЕН** включать:

- M8 authorization language и model versioning;
- Role, RoleBinding, AccessRelationship и Permission;
- SpiceDB только как adapter;
- fail-closed semantics;
- projection/reconciliation rules;
- безопасное explanation;
- обязательные тесты на stale data, revoke, read-your-writes, model compatibility и forbidden direct SDK usage.

---

## 9.8. Risk Decision — оценка риска и принятие решения

### 9.8.1. Паспорт контекста

| Поле | Значение |
| --- | --- |
| Context ID | `CTX-RISK` |
| Service ID | `m8-risk-decision` |
| Классификация | Core Domain |
| Пространство требований | `RISK-*` |
| Миссия | Преобразовывать проверяемый контекст и сигналы риска в объяснимое решение о допустимом следующем действии. |
| Владеет | RiskAssessment, RiskSignalDefinition, DecisionPolicy, PolicyVersion, Decision, DecisionExplanation. |
| Не владеет | Исполнением challenge, базовым permission graph, профилем User, lifecycle ManagedResource. |
| Consistency boundary | Оценка и decision evidence; изменение одной policy/version и Outbox. |
| Published Language | `PL-RISK`, `PL-RISK-EVENTS`. |

### 9.8.2. Назначение и ответственность

Risk Decision **ДОЛЖЕН**:

- принимать нормализованный DecisionContext;
- собирать или получать разрешённые RiskSignal;
- применять активную версию DecisionPolicy;
- возвращать `ALLOW`, `DENY`, `CHALLENGE` или `REVIEW`;
- указывать required assurance/challenge constraints;
- предоставлять безопасное объяснение;
- поддерживать simulation и policy testing;
- фиксировать decision evidence и версии входов;
- публиковать факты решений и изменений policy.

### 9.8.3. Исключённая ответственность

Risk Decision **НЕ ДОЛЖЕН**:

- выполнять OTP/CIBA/WebAuthn;
- самостоятельно предоставлять permission;
- изменять AuthenticationTransaction;
- создавать или удалять ManagedResource;
- раскрывать внутренние anti-fraud rules и sensitive signals внешнему клиенту;
- использовать неописанные персональные признаки без законного основания и governance.

### 9.8.4. Реализуемые бизнес-возможности

| Capability ID | Возможность |
| --- | --- |
| `CAP-RISK-ASSESS` | Online risk assessment. |
| `CAP-RISK-POLICY` | Версионирование и публикация policy. |
| `CAP-RISK-SIGNAL` | Каталог и нормализация signals. |
| `CAP-RISK-VELOCITY` | Velocity и frequency checks. |
| `CAP-RISK-DEVICE` | Device и session intelligence. |
| `CAP-RISK-STEPUP` | Определение required assurance/challenge. |
| `CAP-RISK-REVIEW` | Маршрутизация в manual review. |
| `CAP-RISK-SIM` | Simulation, replay и тестирование policy. |

### 9.8.5. Предметная модель

| Aggregate ID | Корень | Назначение |
| --- | --- | --- |
| `DM-RISK-POLICY` | DecisionPolicy | Логическая policy и набор версионируемых правил. |
| `DM-RISK-ASSESS` | RiskAssessment | Контекст оценки, набор signal references и итоговый Decision. |
| `DM-RISK-SIGNAL` | RiskSignalDefinition | Тип, источник, срок актуальности и правила использования сигнала. |
| `DM-RISK-REVIEW` | ReviewCase | Ручная проверка, если outcome `REVIEW` требует собственного lifecycle. |

### 9.8.6. Решение

Канонические outcomes:

| Outcome | Значение |
| --- | --- |
| `ALLOW` | Риск не требует дополнительного действия; не заменяет permission check. |
| `DENY` | Операция не должна продолжаться. |
| `CHALLENGE` | Требуется дополнительное подтверждение с указанным assurance/constraints. |
| `REVIEW` | Требуется ручное или отложенное решение. |

Decision **ДОЛЖЕН** содержать policy version, decision ID, expires at или validity window, reason codes и evidence reference. Внешний reason code отделяется от внутреннего explanation.

### 9.8.7. Команды

| Command ID | Команда |
| --- | --- |
| `RISK-CMD-001` | `CreateDecisionPolicy` |
| `RISK-CMD-002` | `CreatePolicyVersion` |
| `RISK-CMD-003` | `PublishPolicyVersion` |
| `RISK-CMD-004` | `RetirePolicyVersion` |
| `RISK-CMD-005` | `RegisterRiskSignalDefinition` |
| `RISK-CMD-006` | `ObserveRiskSignal` |
| `RISK-CMD-007` | `CreateReviewCase` |
| `RISK-CMD-008` | `ResolveReviewCase` |

Online evaluate операции являются decision requests, а не командами изменения внешнего агрегата.

### 9.8.8. Запросы решений

| Query/Decision ID | Операция |
| --- | --- |
| `RISK-DEC-001` | `EvaluateAuthenticationRisk` |
| `RISK-DEC-002` | `EvaluateAccessRisk` |
| `RISK-DEC-003` | `EvaluateProvisioningRisk` |
| `RISK-DEC-004` | `EvaluateAdministrativeActionRisk` |
| `RISK-QRY-001` | `GetRiskAssessment` |
| `RISK-QRY-002` | `SimulateDecision` |
| `RISK-QRY-003` | `ExplainDecision` для уполномоченного оператора |
| `RISK-QRY-004` | `ListPolicyVersions` |

### 9.8.9. Публикуемые события

| Event ID | Событие |
| --- | --- |
| `RISK-EVT-001` | `RiskAssessmentCreated` |
| `RISK-EVT-002` | `RiskDecisionMade` |
| `RISK-EVT-003` | `PolicyVersionPublished` |
| `RISK-EVT-004` | `PolicyVersionRetired` |
| `RISK-EVT-005` | `RiskSignalObserved` |
| `RISK-EVT-006` | `ReviewCaseCreated` |
| `RISK-EVT-007` | `ReviewCaseResolved` |

События **НЕ ДОЛЖНЫ** раскрывать raw device fingerprint, secret rule thresholds или персональные признаки сверх утверждённого data contract.

### 9.8.10. Потребляемые контракты

Risk Decision принимает DecisionContext от Authentication, Access, Provisioning или Resource Manager. Контекст включает только необходимые references и нормализованные признаки.

Внешние signal providers подключаются через ACL. Каждый signal **ДОЛЖЕН** иметь:

- источник;
- время наблюдения;
- freshness/TTL;
- confidence;
- purpose limitation;
- классификацию чувствительности;
- правила fallback при недоступности.

### 9.8.11. Policy lifecycle

Новая policy version проходит стадии:

```text
DRAFT → VALIDATED → SHADOW → ACTIVE → RETIRED
```

Переход в `ACTIVE` требует:

- статической валидации;
- набора тестовых cases;
- simulation/replay;
- оценки влияния;
- approval по governance;
- возможности rollback на предыдущую версию.

Изменение active policy in-place запрещено.

### 9.8.12. Объяснимость и повторяемость

Для каждого Decision **ДОЛЖНЫ** сохраняться:

- policy version;
- нормализованный input hash;
- использованные signal versions/freshness;
- outcome;
- внутренние reason codes;
- внешний safe reason code;
- timing и dependency status.

Это должно позволять воспроизвести решение в controlled environment в пределах retention policy.

### 9.8.13. Безопасность и аудит

Policy authoring и publication являются privileged actions. Требуются separation of duties и, для production policy, approval workflow. Логи и Audit **НЕ ДОЛЖНЫ** содержать raw secrets или полный device fingerprint.

### 9.8.14. Наблюдаемость

Минимальные метрики:

- decisions по outcome/use case/policy version;
- decision latency;
- signal provider latency/error/freshness;
- challenge/deny/review rate;
- policy shadow divergence;
- fallback usage;
- review queue age;
- replay/simulation mismatch;
- доля решений с incomplete evidence.

### 9.8.15. Отказы и деградация

| Отказ | Поведение |
| --- | --- |
| Critical signal недоступен | Используется явно заданный policy outcome, обычно DENY/CHALLENGE; не неявный ALLOW. |
| Non-critical signal недоступен | Decision помечается degraded, применяется документированный fallback. |
| Policy store недоступен | Используется последняя локально подтверждённая active version, если её integrity подтверждена. |
| Velocity store недоступен | Для защищаемых сценариев fail closed или CHALLENGE по policy. |
| Simulation failure | Policy version не публикуется. |

### 9.8.16. Запрещённые зависимости

Risk Decision **НЕ ДОЛЖЕН**:

- напрямую изменять Authentication/Provisioning state;
- считать `ALLOW` разрешением Access;
- использовать mutable active policy;
- принимать raw external provider model как доменную модель;
- скрывать факт degraded decision от вызывающего контекста;
- публиковать sensitive evidence в общую event stream.

### 9.8.17. SPDD Context Prompt

Context Prompt `CTX-RISK` **ДОЛЖЕН** включать:

- outcomes и их семантику;
- policy lifecycle и immutable versioning;
- signal freshness/confidence/purpose;
- safe/internal explanation split;
- deterministic evaluation requirements;
- degraded/fail-closed rules;
- обязательные simulation, replay, shadow и security tests.

---

## 9.9. Provisioning — управление жизненным циклом управляемых ресурсов

### 9.9.1. Паспорт контекста

| Поле | Значение |
| --- | --- |
| Context ID | `CTX-PROV` |
| Service ID | `m8-provisioning` |
| Классификация | Core Domain |
| Пространство требований | `PROV-*` |
| Миссия | Преобразовывать декларативное желаемое состояние в проверяемое наблюдаемое состояние управляемого ресурса. |
| Владеет | ResourceDefinition, ResourceRequest, ManagedResource, Placement, Reconciliation, DriverRegistration, desired/observed state. |
| Не владеет | Project hierarchy, User identity, permission graph, внешними cloud objects как канонической бизнес-моделью. |
| Consistency boundary | Один ManagedResource/ResourceRequest и Outbox; длительные действия выполняются workflow. |
| Published Language | `PL-PROVISIONING`, `PL-PROVISIONING-EVENTS`. |

### 9.9.2. Назначение и ответственность

Provisioning **ДОЛЖЕН**:

- принимать декларативную ResourceRequest;
- валидировать тип и schema desired state;
- выбирать Placement по policy;
- создавать ManagedResource;
- запускать Temporal workflow;
- вызывать Driver;
- хранить desired и observed state;
- выполнять reconciliation и drift detection;
- поддерживать update, retry, suspend и delete;
- публиковать состояние и результат provisioning;
- предоставлять Operation progress.

### 9.9.3. Исключённая ответственность

Provisioning **НЕ ДОЛЖЕН**:

- быть владельцем Project;
- назначать права пользователю;
- включать vendor SDK в domain/application;
- считать внешнюю cloud API единственным источником desired state;
- изменять requested specification без явного policy/defaulting;
- удалять внешний ресурс без проверяемого ownership marker.

### 9.9.4. Реализуемые бизнес-возможности

| Capability ID | Возможность |
| --- | --- |
| `CAP-PROV-TYPE` | Каталог ResourceDefinition и schema. |
| `CAP-PROV-REQ` | Приём ResourceRequest. |
| `CAP-PROV-PLACE` | Placement и capacity policy. |
| `CAP-PROV-LIFE` | Create, update, suspend, resume, delete. |
| `CAP-PROV-RECON` | Reconciliation desired/observed. |
| `CAP-PROV-DRIFT` | Drift detection и correction policy. |
| `CAP-PROV-DRV` | Driver registration и compatibility. |
| `CAP-PROV-OPS` | Operation progress и recovery. |

### 9.9.5. Предметная модель

| Aggregate ID | Корень | Назначение |
| --- | --- | --- |
| `DM-PROV-DEF` | ResourceDefinition | Тип ресурса, schema, defaults, lifecycle capabilities и driver requirements. |
| `DM-PROV-REQ` | ResourceRequest | Запрос на создание/изменение с caller intent и idempotency. |
| `DM-PROV-MR` | ManagedResource | Desired state, observed state summary, placement, state и external references. |
| `DM-PROV-DRV` | DriverRegistration | Поддерживаемые types/versions/capabilities. |
| `DM-PROV-REC` | Reconciliation | Один цикл согласования и его результат. |

ExternalReference является непрозрачной ссылкой, а не заменой ManagedResource ID.

### 9.9.6. Состояния ManagedResource

```text
REQUESTED → PROVISIONING → READY
                    ↘ DEGRADED
READY → UPDATING → READY
READY/DEGRADED → DELETING → DELETED

Дополнительные:
SUSPENDED | ERROR | ORPHANED
```

`READY` означает соответствие обязательной части desired state, а не только успешный ответ первого API вызова.

### 9.9.7. Команды

| Command ID | Команда |
| --- | --- |
| `PROV-CMD-001` | `RegisterResourceDefinition` |
| `PROV-CMD-002` | `PublishResourceDefinitionVersion` |
| `PROV-CMD-003` | `CreateManagedResource` |
| `PROV-CMD-004` | `UpdateDesiredState` |
| `PROV-CMD-005` | `SuspendManagedResource` |
| `PROV-CMD-006` | `ResumeManagedResource` |
| `PROV-CMD-007` | `DeleteManagedResource` |
| `PROV-CMD-008` | `ReconcileManagedResource` |
| `PROV-CMD-009` | `RetryProvisioning` |
| `PROV-CMD-010` | `RegisterDriver` |
| `PROV-CMD-011` | `AdoptExternalResource` через отдельную policy и Operation |

### 9.9.8. Запросы

| Query ID | Запрос |
| --- | --- |
| `PROV-QRY-001` | `GetResourceDefinition` |
| `PROV-QRY-002` | `ListResourceDefinitions` |
| `PROV-QRY-003` | `GetManagedResource` |
| `PROV-QRY-004` | `ListManagedResources` |
| `PROV-QRY-005` | `GetObservedState` |
| `PROV-QRY-006` | `GetReconciliationHistory` |
| `PROV-QRY-007` | `GetProvisioningOperation` |
| `PROV-QRY-008` | `PreviewPlacement` |

Observed state может быть stale; ответ **ДОЛЖЕН** содержать `observed_at`, `reconciliation_id` и freshness metadata.

### 9.9.9. Публикуемые события

| Event ID | Событие |
| --- | --- |
| `PROV-EVT-001` | `ManagedResourceRequested` |
| `PROV-EVT-002` | `ProvisioningStarted` |
| `PROV-EVT-003` | `ManagedResourceReady` |
| `PROV-EVT-004` | `ManagedResourceStateChanged` |
| `PROV-EVT-005` | `DesiredStateChanged` |
| `PROV-EVT-006` | `ObservedStateChanged` |
| `PROV-EVT-007` | `DriftDetected` |
| `PROV-EVT-008` | `ReconciliationCompleted` |
| `PROV-EVT-009` | `ProvisioningFailed` |
| `PROV-EVT-010` | `ManagedResourceDeleted` |
| `PROV-EVT-011` | `ExternalResourceOrphaned` |

### 9.9.10. Потребляемые контракты

| Поставщик | Контракт | Назначение |
| --- | --- | --- |
| Resource Manager | Project/Service state | Scope и ownership контекст. |
| Access | CheckPermission | Право create/update/delete managed resource. |
| Risk Decision | EvaluateProvisioningRisk | Approval, step-up, deny или review для чувствительных ресурсов. |
| Temporal | Workflow execution | Надёжная оркестрация, retries и compensation. |
| Driver | Apply/Observe/Delete | Работа с внешней системой через нормализованный driver contract. |

### 9.9.11. Driver contract

Driver **ДОЛЖЕН** реализовывать, в зависимости от capability:

- `Validate`;
- `Plan`;
- `Apply`;
- `Observe`;
- `Delete`;
- `CheckHealth`;
- `GetCapabilities`.

Каждый вызов Driver **ДОЛЖЕН** быть идемпотентным по operation/reconciliation key. Driver не принимает business permission decisions и не хранит M8 credentials в открытом виде.

### 9.9.12. Reconciliation

Reconciliation сравнивает desired и observed state, формирует Plan и применяет допустимые действия. Он **ДОЛЖЕН**:

- учитывать generation desired state;
- не применять результат устаревшей generation;
- сохранять before/after summary;
- различать transient и terminal errors;
- ограничивать retries;
- поддерживать backoff и circuit breaker;
- обнаруживать drift;
- не исправлять drift автоматически, если policy требует review.

### 9.9.13. Temporal workflow

Temporal используется для оркестрации, но workflow state **НЕ ЯВЛЯЕТСЯ** каноническим предметным состоянием. ManagedResource и Operation остаются читаемыми из сервиса даже при временной недоступности Temporal.

Workflow ID, retry policy и activity types являются инфраструктурными деталями и не должны становиться частью публичного API.

### 9.9.14. Безопасность и аудит

Обязательны:

- permission и risk decision до создания/изменения sensitive resource;
- secure secret references вместо secret values;
- ownership tags/markers во внешней системе;
- защита delete от чужого external resource;
- audit desired state changes и driver actions без секретов;
- separation of duties для production/high-impact resources.

### 9.9.15. Наблюдаемость

Минимальные метрики:

- resources по type/state/placement;
- provisioning success/failure latency;
- reconciliation duration и backlog;
- driver errors/throttling;
- drift count и drift age;
- orphaned resource count;
- operation retry count;
- desired-observed generation lag;
- deletion stuck duration;
- external API quota usage.

### 9.9.16. Отказы и восстановление

| Отказ | Поведение |
| --- | --- |
| Driver timeout | Activity retry по policy; состояние не объявляется READY. |
| Temporal недоступен | Новые workflows не стартуют; локальные commands могут фиксироваться как pending только при гарантированном dispatch. |
| External API partial success | Observe определяет фактическое состояние; workflow продолжает reconciliation. |
| Resource Manager scope deleted | Запускается governed cleanup; новые changes запрещаются. |
| Risk Decision недоступен | Sensitive operation fail closed. |
| Service restart | Workflow и reconciliation восстанавливаются по durable state. |

### 9.9.17. Запрещённые зависимости

Provisioning **НЕ ДОЛЖЕН**:

- размещать vendor resource object в domain;
- завершать Operation только на основании отправки API request;
- автоматически усыновлять внешний ресурс без ownership verification;
- хранить plaintext secrets;
- считать Temporal history единственным источником состояния;
- удалять resource hierarchy в Resource Manager.

### 9.9.18. SPDD Context Prompt

Context Prompt `CTX-PROV` **ДОЛЖЕН** включать:

- desired/observed state и generation rules;
- ManagedResource state machine;
- ResourceDefinition/Driver separation;
- Temporal как infrastructure orchestration;
- idempotent driver calls, retries, compensation и recovery;
- secret handling и ownership verification;
- обязательные failure-injection и reconciliation tests.

---

## 9.10. Audit — неизменяемая история значимых действий

### 9.10.1. Паспорт контекста

| Поле | Значение |
| --- | --- |
| Context ID | `CTX-AUD` |
| Service ID | `m8-audit` |
| Классификация | Supporting Domain |
| Пространство требований | `AUD-*` |
| Миссия | Сохранять проверяемую, неизменяемую и доступную по правилам историю значимых действий и решений платформы. |
| Владеет | AuditEvent, AuditActor, AuditTarget, AuditContext, AuditChangeSet, RetentionPolicy, ExportJob, IntegrityEvidence. |
| Не владеет | Исходным business state, permission/risk decision и операционными логами сервисов. |
| Consistency boundary | Append одной записи/пакета и integrity metadata; export/retention выполняются отдельными operations. |
| Published Language | `PL-AUDIT`. |

### 9.10.2. Назначение и ответственность

Audit **ДОЛЖЕН**:

- принимать нормализованные audit records;
- валидировать обязательные поля;
- обеспечивать append-only semantics;
- дедуплицировать записи по event ID;
- сохранять actor, target, action, outcome и changes;
- поддерживать поиск с авторизацией;
- обеспечивать retention, legal hold и export;
- формировать integrity evidence;
- сохранять связь с correlation/causation/operation;
- отделять audit trail от debug logs и domain event stream.

### 9.10.3. Исключённая ответственность

Audit **НЕ ДОЛЖЕН**:

- блокировать локальную бизнес-транзакцию синхронным удалённым вызовом;
- становиться источником истины business resource;
- повторно исполнять business command;
- хранить secrets, raw tokens, passwords, OTP или private keys;
- предоставлять неограниченный поиск без permission и scope filtering;
- позволять обычное update/delete AuditEvent.

### 9.10.4. Реализуемые бизнес-возможности

| Capability ID | Возможность |
| --- | --- |
| `CAP-AUD-INGEST` | Надёжный приём AuditEvent. |
| `CAP-AUD-STORE` | Append-only хранение. |
| `CAP-AUD-SEARCH` | Авторизованный поиск и timeline. |
| `CAP-AUD-INTEGRITY` | Проверка целостности. |
| `CAP-AUD-RET` | Retention и legal hold. |
| `CAP-AUD-EXP` | Экспорт и подтверждение набора данных. |
| `CAP-AUD-PRIV` | Redaction/pseudonymization по policy без разрушения evidence. |

### 9.10.5. Предметная модель

| Aggregate ID | Корень | Назначение |
| --- | --- | --- |
| `DM-AUD-EVENT` | AuditEvent | Неизменяемая запись значимого действия или решения. |
| `DM-AUD-RET` | RetentionPolicy | Сроки, legal hold и правила удаления архивных сегментов. |
| `DM-AUD-EXP` | ExportJob | Управляемый экспорт с scope, filters и integrity manifest. |
| `DM-AUD-HOLD` | LegalHold | Запрет удаления соответствующего набора записей. |

### 9.10.6. Минимальная схема AuditEvent

AuditEvent **ДОЛЖЕН** содержать:

- `audit_event_id`;
- `occurred_at` и `recorded_at`;
- producer service и producer event ID;
- ActorReference и, при отличии, SubjectReference;
- ClientReference;
- action;
- TargetReference;
- Resource Scope;
- outcome и error/reason code;
- AuditChangeSet;
- request, trace, correlation, causation и operation IDs;
- source version;
- data classification;
- integrity metadata.

### 9.10.7. Команды и запросы

| ID | Операция |
| --- | --- |
| `AUD-CMD-001` | `AppendAuditEvent` |
| `AUD-CMD-002` | `AppendAuditBatch` |
| `AUD-CMD-003` | `CreateRetentionPolicy` |
| `AUD-CMD-004` | `CreateLegalHold` |
| `AUD-CMD-005` | `ReleaseLegalHold` |
| `AUD-CMD-006` | `CreateExportJob` |
| `AUD-QRY-001` | `GetAuditEvent` |
| `AUD-QRY-002` | `SearchAuditEvents` |
| `AUD-QRY-003` | `GetActorTimeline` |
| `AUD-QRY-004` | `GetResourceTimeline` |
| `AUD-QRY-005` | `VerifyIntegrity` |
| `AUD-QRY-006` | `GetExportJob` |

### 9.10.8. Ingestion model

Бизнес-сервис **ДОЛЖЕН** записывать AuditIntent локально атомарно с изменением агрегата или обеспечивать эквивалентную надёжность через единый transactional outbox. Audit получает запись асинхронно и дедуплицирует по producer + producer event ID.

Для событий отказа, не сопровождаемых commit, сервис **ДОЛЖЕН** использовать надёжный security/audit channel, соответствующий критичности события. Обычный удалённый fire-and-forget вызов недостаточен.

### 9.10.9. Неизменяемость и целостность

После записи AuditEvent запрещены update и delete на уровне обычного API. Исправление ошибок выполняется новой корректирующей записью, ссылающейся на исходную.

Integrity **МОЖЕТ** обеспечиваться комбинацией:

- append-only storage policy;
- hash chaining по сегментам;
- подписанными manifests;
- immutable object storage для архивов;
- ограничением административного доступа;
- периодической verification job.

Конкретный механизм фиксируется ADR.

### 9.10.10. Поиск и доступ

SearchAuditEvents **ДОЛЖЕН**:

- применять permission и Resource Scope;
- иметь ограничение диапазона времени;
- использовать cursor pagination;
- поддерживать поля actor, target, action, outcome, operation и correlation;
- маскировать персональные и чувствительные поля по роли;
- записывать аудит самого доступа к особо чувствительной истории.

### 9.10.11. Retention и privacy

RetentionPolicy определяет срок по типу события, юрисдикции, tenant/scope и legal hold. Удаление по retention выполняется сегментно и проверяемо.

Когда требуется обезличивание Subject, Audit сохраняет непрозрачный stable reference и удаляет или псевдонимизирует display data в соответствии с policy, не разрушая доказательную связность.

### 9.10.12. Публикуемые события

| Event ID | Событие |
| --- | --- |
| `AUD-EVT-001` | `AuditEventRecorded` — только для ограниченных служебных потребителей. |
| `AUD-EVT-002` | `AuditExportCompleted` |
| `AUD-EVT-003` | `RetentionPolicyChanged` |
| `AUD-EVT-004` | `IntegrityViolationDetected` |
| `AUD-EVT-005` | `LegalHoldChanged` |

Audit **НЕ ДОЛЖЕН** републиковать полный поток AuditEvent в неограниченную общую шину.

### 9.10.13. Наблюдаемость

Минимальные метрики:

- ingest rate и lag по producer;
- duplicate rate;
- invalid/rejected event count;
- storage growth;
- search latency и scanned volume;
- export duration/size;
- retention backlog;
- integrity verification status;
- missing producer sequence/gap, где поддерживается;
- доля audit intents старше допустимого delivery window.

### 9.10.14. Отказы и деградация

| Отказ | Поведение |
| --- | --- |
| Audit API/store недоступен | Producer сохраняет AuditIntent локально и повторяет доставку; бизнес-commit не откатывается удалённо. |
| Event invalid | Запись помещается в quarantine с причиной; producer получает диагностический сигнал. |
| Duplicate | Идемпотентный success. |
| Integrity violation | Немедленный alert, блокировка затронутого export и incident workflow. |
| Search backend деградирован | Ingestion продолжает работать при разделённой архитектуре write/read. |

### 9.10.15. Запрещённые зависимости

Audit **НЕ ДОЛЖЕН**:

- читать business DB для восстановления missing fields в момент ingest;
- изменять source aggregate;
- хранить raw secret material;
- позволять hard delete в обход retention/legal hold;
- использовать mutable display name как единственный actor/target key;
- становиться общей системой application logs.

### 9.10.16. SPDD Context Prompt

Context Prompt `CTX-AUD` **ДОЛЖЕН** включать:

- append-only semantics;
- обязательную схему AuditEvent;
- producer-side reliable intent/outbox;
- deduplication и integrity;
- permission-aware search;
- retention/legal hold/privacy constraints;
- запрет secret material и business-state mutation;
- тесты на duplicate, delayed delivery, integrity failure, redaction и scoped search.

---

## 9.11. Common Operation — общий контракт длительных операций

### 9.11.1. Статус

Common Operation является общим контрактом и общим языком, но **НЕ ОБЯЗАТЕЛЬНО** отдельным ограниченным контекстом или централизованным сервисом.

Владелец команды, породившей длительную работу, остаётся владельцем Operation. Например:

- удаление Project — Resource Manager;
- merge identities — Identity;
- длительная authentication orchestration — Authentication;
- access review — Access;
- manual risk review — Risk Decision;
- provisioning resource — Provisioning;
- audit export — Audit.

### 9.11.2. Общая модель

Operation **ДОЛЖНА** содержать:

- уникальное имя/ID;
- service owner;
- action string;
- target resource reference;
- create/update/end timestamps;
- status;
- progress stage, percent и message code;
- metadata;
- result или error;
- cancellability;
- correlation/trace references;
- caller/actor scope, необходимую для проверки доступа.

### 9.11.3. Канонические состояния

```text
PENDING → RUNNING → SUCCEEDED
                  ↘ FAILED
                  ↘ CANCELLED
                  ↘ TIMED_OUT
```

`CANCEL_REQUESTED` **МОЖЕТ** использоваться как промежуточное состояние. Cancel не гарантирует мгновенную остановку и **НЕ ДОЛЖЕН** объявляться успешным до подтверждённого безопасного состояния.

### 9.11.4. Общие операции API

| Operation ID | Операция |
| --- | --- |
| `OPS-QRY-001` | `GetOperation` |
| `OPS-QRY-002` | `ListOperations` |
| `OPS-QRY-003` | `WaitOperation` |
| `OPS-CMD-001` | `CancelOperation` |
| `OPS-CMD-002` | `DeleteOperationMetadata` — только если разрешено policy и не удаляет business/audit evidence. |

### 9.11.5. Правила владения

- Operation хранится у сервиса-владельца команды.
- Общий protobuf package содержит только стабильный контракт.
- Service-specific metadata/result используют `Any` или типизированный wrapper только по принятому контракту.
- Operation **НЕ ДОЛЖНА** быть единственным источником состояния предметного ресурса.
- Удаление Operation metadata **НЕ ДОЛЖНО** удалять результат, Audit или предметный ресурс.
- Progress является наблюдаемым представлением и может быть приблизительным, но stage должен быть предметно значимым.

---

## 9.12. Сводная матрица контекстов

| Контекст | Главный предметный вопрос | Основной агрегат | Ключевое решение/результат | Внешняя технология за ACL |
| --- | --- | --- | --- | --- |
| Resource Manager | Где находится ресурс в административной иерархии и каково его состояние? | Project/Workspace/Organization | Канонический ResourceReference и lifecycle | YDB/Temporal как infrastructure |
| Identity | Кто является субъектом и в каком состоянии его идентичность? | User | SubjectReference и identity status | External IdP directory при наличии |
| Authentication | Доказал ли субъект идентичность с требуемым assurance? | AuthenticationTransaction | Authentication result/session/handoff | Keycloak, WebAuthn, OTP providers |
| Access | Может ли субъект выполнить action над resource? | AuthorizationModel/RoleBinding | ALLOW/DENY | SpiceDB |
| Risk Decision | Требует ли контекст deny, challenge или review? | RiskAssessment/DecisionPolicy | ALLOW/DENY/CHALLENGE/REVIEW | Signal providers, rule engine |
| Provisioning | Соответствует ли внешний ресурс desired state? | ManagedResource | READY/DEGRADED/ERROR и Operation | Temporal, Kubernetes, cloud APIs |
| Audit | Что произошло, кто инициировал и каков результат? | AuditEvent | Неизменяемая evidence record | Storage/search/archive technologies |

## 9.13. Владение сквозными сценариями

Сквозной сценарий **ДОЛЖЕН** иметь одного process owner, даже если в нём участвуют несколько контекстов.

| Сценарий | Владелец процесса | Участники |
| --- | --- | --- |
| Создание Project | Resource Manager | Access, Audit, при необходимости Provisioning. |
| Назначение пользователя в Project | Identity | Resource Manager, Access, Audit. |
| Login/CIBA | Authentication | Identity, Risk Decision, Access, provider, Audit. |
| Step-up перед privileged action | Вызывающий контекст до запуска; Authentication владеет challenge | Risk Decision, Authentication, Access, Audit. |
| Изменение RoleBinding | Access | Identity, Resource Manager, Risk Decision при policy, Audit. |
| Создание ManagedResource | Provisioning | Resource Manager, Access, Risk Decision, Driver, Audit. |
| Удаление Project с ресурсами | Resource Manager | Identity, Access, Provisioning, Audit. |
| Экспорт Audit | Audit | Access, object storage adapter. |

Process owner отвечает за:

- состояние workflow;
- Operation;
- таймауты и retries;
- compensation;
- пользовательский результат;
- сквозную correlation;
- completion criteria.

Участник процесса остаётся владельцем своих агрегатов и **НЕ ПЕРЕДАЁТ** process owner право прямой записи в своё хранилище.

## 9.14. Правила физической декомпозиции сервисов

В базовой архитектуре каждый `CTX-*` реализуется отдельным логическим сервисом. Разделение контекста на несколько deployment units допускается, если:

1. сохраняется единый предметный владелец;
2. внутренние компоненты не объявляются самостоятельными контекстами без анализа;
3. публичный Published Language остаётся стабильным;
4. данные не становятся совместно изменяемыми несколькими независимыми владельцами;
5. внутренняя сеть не превращается в замену явным application contracts;
6. операционная сложность оправдана масштабированием, безопасностью или независимостью жизненного цикла.

Объединение двух контекстов в один deployment **МОЖЕТ** быть допустимо на ранней стадии, но кодовые модули, данные и зависимости **ДОЛЖНЫ** сохранять логические границы. Общая база с прямыми cross-context joins запрещена даже при совместном развертывании.

## 9.15. Критерии выделения нового контекста

Новый bounded context рассматривается, если одновременно проявляются несколько признаков:

- самостоятельный предметный язык;
- собственные инварианты и lifecycle;
- отдельный источник истины;
- конфликтующее значение существующих терминов;
- независимый ритм изменений;
- отдельные security/compliance boundaries;
- необходимость автономного масштабирования;
- собственная команда-владелец и capability;
- сложность существующего контекста стала препятствием для изменений.

Само наличие новой таблицы, очереди, UI-страницы, адаптера или background job **НЕ ЯВЛЯЕТСЯ** достаточным основанием.

## 9.16. Правила изменения границ

Изменение владения агрегатом или capability между контекстами требует:

1. ADR с причиной и альтернативами;
2. обновления Capability Map;
3. обновления Domain Model и Context Map;
4. плана переноса source of truth;
5. стратегии dual-read/dual-write, если она временно необходима;
6. версии Published Language;
7. migration и rollback plan;
8. обновления требований и traceability;
9. обновления Context/Feature/Review Prompts;
10. архитектурных и контрактных тестов после миграции.

Постоянный dual ownership одного предметного факта запрещён.

## 9.17. Пространства идентификаторов требований

| Контекст | Функциональные требования | Нефункциональные требования | Инварианты | Сценарии |
| --- | --- | --- | --- | --- |
| Resource Manager | `RM-FR-*` | `RM-NFR-*` | `INV-RM-*` | `UC-RM-*` |
| Identity | `ID-FR-*` | `ID-NFR-*` | `INV-ID-*` | `UC-ID-*` |
| Authentication | `AUTH-FR-*` | `AUTH-NFR-*` | `INV-AUTH-*` | `UC-AUTH-*` |
| Access | `ACC-FR-*` | `ACC-NFR-*` | `INV-ACC-*` | `UC-ACC-*` |
| Risk Decision | `RISK-FR-*` | `RISK-NFR-*` | `INV-RISK-*` | `UC-RISK-*` |
| Provisioning | `PROV-FR-*` | `PROV-NFR-*` | `INV-PROV-*` | `UC-PROV-*` |
| Audit | `AUD-FR-*` | `AUD-NFR-*` | `INV-AUD-*` | `UC-AUD-*` |
| Common Operation | `OPS-FR-*` | `OPS-NFR-*` | `INV-OPS-*` | `UC-OPS-*` |

Каждое сервисное требование **ДОЛЖНО** ссылаться как минимум на:

- Capability ID;
- Context ID;
- Aggregate или decision owner;
- Context Map relationship ID при интеграции;
- acceptance criteria;
- contract или явно указанный internal behavior;
- Structured Prompt ID после перехода к реализации.

## 9.18. Минимальный набор проверок контекста

Для каждого контекста pipeline **ДОЛЖЕН** включать:

1. проверку запрещённых импортов;
2. unit tests агрегатных инвариантов;
3. application tests команд и идемпотентности;
4. repository tests optimistic locking;
5. Outbox/Inbox integration tests;
6. protobuf breaking-change checks;
7. provider/adapter contract tests;
8. authorization negative tests;
9. audit completeness tests;
10. telemetry attribute tests для критичных сценариев;
11. failure-injection tests внешних зависимостей;
12. requirement coverage check;
13. SPDD prompt schema validation;
14. Review Prompt compliance report.

## 9.19. Критерии соответствия главы

Архитектурное или функциональное изменение соответствует спецификациям ограниченных контекстов, если:

1. определён один `CTX-*` владелец изменяемого предметного факта;
2. capability и requirement namespace соответствуют владельцу;
3. агрегатная и транзакционная граница не пересекает контексты;
4. исключённые обязанности не перенесены в сервис скрыто;
5. команда, запрос и событие имеют устойчивый идентификатор;
6. публичный контракт использует Published Language;
7. внешняя технология скрыта ACL/Adapter/Driver;
8. определены authorization, risk и audit requirements;
9. идемпотентность и optimistic locking заданы там, где применимо;
10. длительная работа представлена Operation владельца команды;
11. определено поведение при недоступности каждой синхронной зависимости;
12. события публикуются через надёжный transactional mechanism;
13. consumer projections являются rebuildable и version-aware;
14. logs/events не раскрывают secrets и лишние персональные данные;
15. наблюдаемость включает технические и предметные метрики;
16. тесты проверяют инварианты, контракты, отказы и безопасность;
17. requirement traceability до SPDD и acceptance evidence определена;
18. изменение границы оформлено ADR, если затронуто владение.

---
