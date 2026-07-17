---
title: "PADS: глоссарий"
description: "Глоссарий терминов, сокращения и правила ведения."
keywords:
  - "M8 Platform"
  - "PADS"
  - "architecture"
---

# 24. Глоссарий {#pads-glossary}

{% note info "Навигация PADS" %}

[Оглавление PADS](../index.md) | [Предыдущий раздел: 23. Архитектурное управление](23-architecture-governance.md) | [Следующий раздел: Приложение A. Начальная структура репозитория](../appendices/appendix-a-repository-structure.md)

{% endnote %}

## 24.1. Назначение

Глоссарий содержит краткие нормативные определения терминов. Полные значения и владельцы предметных понятий определены главой 5; при расхождении приоритет имеет специализированная глава.

| Термин | Определение |
| --- | --- |
| Acceptance Criterion | Проверяемое условие, доказывающее выполнение требования. |
| Access | Ограниченный контекст, принимающий решения о разрешениях, ролях и отношениях. |
| ACL | Anti-Corruption Layer — слой перевода между моделью M8 и внешней системой. Не путать с access control list. |
| Actor | Пользователь, сервис или automation, инициировавшие действие. |
| ADR | Architecture Decision Record — запись значимого архитектурного решения, альтернатив и последствий. |
| AAL | Authentication Assurance Level — уровень уверенности в подтверждённой идентичности. |
| Aggregate | Граница согласованности и изменения предметного состояния с одним корнем. |
| Annotation | Расширяемые технические метаданные, не определяющие идентичность ресурса. |
| API | Версионируемый контракт вызова capability владельца. |
| API First | Подход, при котором контракт проектируется и проверяется до реализации. |
| Assurance | Степень уверенности, достигнутая аутентификацией. |
| Audit Event | Неизменяемая запись, кто, что, над чем, когда и с каким результатом сделал. |
| Authentication | Процесс подтверждения идентичности Subject до требуемого assurance. |
| Authentication Challenge | Конкретная проверка: OTP, approval, WebAuthn и др. |
| Authentication Transaction | Агрегат, управляющий одним процессом аутентификации. |
| Authorization | Решение, может ли Subject выполнить действие над Resource. |
| Availability | Доля запросов/результатов, доступных в рамках SLO. |
| Backfill | Управляемое заполнение нового поля, таблицы или проекции историческими данными. |
| BFF | Backend for Frontend — слой композиции API для конкретного UI/клиента без присвоения предметного владения. |
| Bounded Context | Граница, внутри которой предметные термины и модели имеют единый смысл и владельца. |
| Bulkhead | Изоляция ресурсов/пулов для ограничения распространения отказа. |
| Capability | Устойчивая способность платформы предоставлять бизнес-результат. |
| Causation ID | Идентификатор непосредственной причины команды, события или действия. |
| CIBA | Client Initiated Backchannel Authentication — backchannel flow запуска аутентификации клиентом. |
| Circuit Breaker | Механизм временной остановки вызовов деградировавшей зависимости. |
| Clean Architecture | Стиль, в котором зависимости направлены к domain/application, а инфраструктура подключается адаптерами. |
| Client | Приложение или интеграция, использующая Authentication/API. |
| Command | Намерение выполнить действие; может быть отклонено. |
| Compensation | Предметное действие, уменьшающее последствия уже подтверждённого шага Saga. |
| Conformist | Отношение Context Map, при котором consumer принимает язык provider без ACL. |
| Consistency Class | Объявленная модель актуальности и транзакционности взаимодействия. |
| Consistency Token | Маркер revision, используемый для требуемой свежести решения Access/хранилища. |
| Constitution Prompt | Верхнеуровневый SPDD-артефакт с неизменяемыми правилами проекта. |
| Consumer | Сервис/процесс, использующий API или событие provider. |
| Context Map | Карта bounded contexts и типов их отношений. |
| Contract Owner | Роль, отвечающая за схему, совместимость и lifecycle API/event. |
| Correlation ID | Идентификатор сквозного бизнес-процесса. |
| Credential | Доказательство, используемое для аутентификации. |
| Data Classification | Класс чувствительности данных и соответствующие controls. |
| Data Owner | Контекст/роль, владеющие смыслом, качеством, retention и удалением данных. |
| Dead Letter Queue | Изолированное хранилище сообщений, которые нельзя обработать стандартным retry. |
| Decision ID | Стабильная ссылка на решение Access или Risk. |
| Desired State | Состояние внешнего/managed ресурса, которого M8 стремится достичь. |
| Domain Event | Свершившийся факт внутри bounded context. |
| Domain Service | Предметная операция, не принадлежащая естественно одной Entity/Value Object. |
| Drift | Расхождение desired и observed state. |
| ETag | Opaque маркер версии представления ресурса для optimistic concurrency. |
| Entity | Предметный объект с устойчивой идентичностью и жизненным циклом. |
| Error Budget | Допустимый объём нарушения SLO за период. |
| Event Envelope | Общая метаоболочка события: ID, type, version, time, producer, correlation. |
| Eventual Consistency | Состояние, при котором копии сходятся после распространения фактов. |
| External Identity | Связь User с внешним issuer+subject. |
| Fail-closed | Отказ в выполнении при невозможности проверить security condition. |
| Fail-open | Продолжение при отказе security control; требует явного исключения/risk acceptance. |
| FieldMask | Контрактный список изменяемых полей частичного update. |
| Fitness Function | Автоматическая проверка архитектурного свойства. |
| Freshness | Возраст/отставание производной копии относительно source of truth. |
| Gateway | Application port для обращения к capability другого контекста. |
| Group | Управляемая совокупность Subjects/Users в Identity. |
| Handoff | Безопасный результат аутентификации, передаваемый следующему участнику flow. |
| Idempotency | Свойство повторного вызова не создавать новый логический эффект. |
| Inbox | Механизм consumer-side дедупликации сообщений. |
| Integration Event | Публичное межконтекстное представление подтверждённого факта. |
| Invariant | Условие, которое должно сохраняться для корректности aggregate. |
| Label | Ограниченные индексируемые key-value метаданные ресурса. |
| Legal Hold | Приостановка удаления данных из-за юридического требования. |
| Long Running Operation | Публичный ресурс наблюдения длительного действия. |
| Membership | Участие Subject/User в organization/workspace/project/group согласно owner model. |
| Metadata | Дополнительные данные контракта; не должны скрывать core domain semantics. |
| mTLS | Mutual TLS — взаимная аутентификация сторон TLS. |
| Observed State | Состояние managed resource, обнаруженное во внешней системе. |
| Open Host Service | Стандартизованный публичный интерфейс контекста для нескольких consumers. |
| Operation | Ресурс со state/progress/result/error длительного действия. |
| Outbox | Атомарная запись события вместе с предметной транзакцией для последующей публикации. |
| Owner Context | Контекст, владеющий инвариантом и нормативным решением. |
| PADS | Platform Architecture & Domain Specification — настоящая спецификация. |
| Permission | Именованное действие над типом ресурса. |
| Policy | Версионируемое правило принятия решений. |
| Principal | Проверенная сторона, от имени которой выполняется security decision. |
| Process Manager | Явный компонент, хранящий состояние межконтекстного процесса. |
| Projection | Локальная производная модель данных другого owner. |
| Protovalidate | Механизм декларативной проверки Protobuf-контрактов. |
| Provider | Контекст или внешняя система, предоставляющие capability/contract. |
| Published Language | Версионируемый язык API/events, публикуемый provider context. |
| RPO | Recovery Point Objective — допустимая потеря данных по времени. |
| RTO | Recovery Time Objective — целевое время восстановления. |
| Reconciliation | Повторяемое сравнение и сближение desired и observed state. |
| Relationship | Типизированная связь Subject и Resource в Access. |
| Requirement | Нормативное проверяемое описание требуемого поведения или свойства. |
| Resource | Адресуемый объект платформы с owner и lifecycle. |
| Resource Manager | Контекст Organization, Workspace, Project и ServiceRegistration. |
| Revision | Монотонная версия предметного состояния. |
| Risk Assessment | Результат оценки динамического риска действия/аутентификации. |
| Risk Decision | Контекст, возвращающий ALLOW/DENY/CHALLENGE/REVIEW. |
| Role | Именованная композиция permissions/relations. |
| Role Binding | Назначение Role субъекту в resource scope. |
| Saga | Последовательность локальных транзакций с coordination/compensation. |
| Secret | Restricted credential/key/token, не допускаемый в коде, логах и prompt. |
| Service Identity | Техническая идентичность workload/service. |
| ServiceRegistration | Представление сервиса, зарегистрированного в Project. |
| Session | Ограниченный во времени security context после аутентификации. |
| Shared Kernel | Минимальный согласованно изменяемый набор межконтекстных контрактных типов. |
| SLI | Service Level Indicator — измеряемый показатель результата. |
| SLO | Service Level Objective — целевое значение SLI. |
| Snapshot | Полное состояние данных на фиксированную позицию/момент. |
| Source of Truth | Авторитетный владелец и хранилище нормативного состояния. |
| SPDD | Structured-Prompt-Driven Development — разработка через версионируемые структурированные промпты. |
| Step-up | Новая проверка для достижения более высокого assurance. |
| Structured Prompt | Формальный SPDD-артефакт с objective, scope, constraints, tests и traceability. |
| Subject | Идентичность, которая аутентифицируется или получает права. |
| Temporal | Workflow engine, используемый через adapter для длительной orchestration. |
| Tenant Scope | Область organization/workspace/project, в пределах которой разрешено действие. |
| Tombstone | Факт удаления/недоступности, распространяемый производным копиям. |
| Trace ID | Идентификатор distributed trace. |
| Ubiquitous Language | Единый предметный язык внутри bounded context и документации. |
| User | Человеческая/учётная identity entity, которой владеет Identity. |
| User Pool | Изолированная область пользователей и identity policies. |
| Value Object | Неизменяемое значение, определяемое содержимым, а не идентичностью. |
| Workflow | Устойчивый процесс orchestration с состоянием, retries, timers и signals. |
| YDB | Базовое распределённое хранилище сервисных данных M8. |
| YDB Topics | Базовый транспорт потоковых событий, если ADR не определяет иной механизм. |
| Zero Trust | Модель, в которой доверие не следует из сети и каждый запрос проверяется явно. |

## 24.2. Сокращения

| Сокращение | Расшифровка |
| --- | --- |
| ACL | Anti-Corruption Layer |
| ADR | Architecture Decision Record |
| AAL | Authentication Assurance Level |
| API | Application Programming Interface |
| BFF | Backend for Frontend |
| CDC | Change Data Capture |
| CIBA | Client Initiated Backchannel Authentication |
| DDD | Domain-Driven Design |
| DLQ | Dead Letter Queue |
| DR | Disaster Recovery |
| IAM | Identity and Access Management |
| LRO | Long Running Operation |
| NFR | Non-Functional Requirement |
| PADS | Platform Architecture & Domain Specification |
| PII | Personally Identifiable Information |
| RPO | Recovery Point Objective |
| RTO | Recovery Time Objective |
| SLI | Service Level Indicator |
| SLO | Service Level Objective |
| SPDD | Structured-Prompt-Driven Development |

## 24.3. Критерии ведения глоссария

- новый нормативный термин добавляется в главу 5 и кратко в главу 24;
- синонимы не используются как разные модели;
- устаревший термин помечается deprecated и содержит replacement;
- названия API/events/code SHOULD соответствовать owner language;
- внешние vendor terms переводятся ACL и не становятся доменными автоматически.

---
