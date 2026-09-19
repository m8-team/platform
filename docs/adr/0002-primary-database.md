# `ADR-0002`: Основная операционная база данных

- Идентификатор: `ADR-0002`.
- Дата: `2026-09-19`.
- Статус: `Accepted`.
- Подтверждение согласования: решение владельца проекта от `2026-09-19` — выбрать YDB как основную операционную СУБД; ссылка на фиксирующий commit будет добавлена после записи решения.

## Контекст

M8 Platform требуется основная транзакционная СУБД для состояния backend-модулей: организаций, workspace, проектов, идентичностей и других принадлежащих модулям сущностей.

Выбор должен учитывать:

- строгую консистентность и ACID-транзакции для операционных данных;
- отказоустойчивость и возможность горизонтального масштабирования без прикладного ручного sharding;
- работу с распределённой нагрузкой и ростом объёма данных;
- официальный Go SDK и пригодность для сервисов на Go;
- возможность локальной разработки и self-hosted эксплуатации;
- предсказуемую модель лицензирования;
- совместимость с принятой системой миграций [ADR-0001](0001-database-migrations.md);
- достаточную наблюдаемость и эксплуатационные инструменты;
- явные правила проектирования ключей, транзакционных границ и retries.

Этот ADR выбирает **основную operational system of record**. Он не означает, что выбранная СУБД должна использоваться для всех типов данных. Search, telemetry, object/blob storage, authorization graph, event streaming и аналитические workloads могут использовать специализированные системы, если это будет обосновано отдельными требованиями или ADR.

При подготовке решения использованы установленные skills `ydb-core` и `ydb-table`.

## Рассмотренные варианты

| Вариант | Преимущества | Ограничения и последствия |
| --- | --- | --- |
| [YDB](https://ydb.tech/docs/en/) | Distributed SQL; automatic sharding и rebalancing; strong consistency и ACID distributed transactions; Serializable по умолчанию; официальный Go SDK; self-hosted open-source edition под Apache 2.0; поддержка goose; единое пространство schema objects. | YQL не является PostgreSQL; схема и primary key должны проектироваться с учётом распределения нагрузки; монотонный первый компонент PK создаёт hotspot; отсутствует PostgreSQL-совместимость как цель; distributed transactions дороже локальных; команда должна освоить YDB-specific tooling и semantics. |
| [PostgreSQL](https://www.postgresql.org/docs/) | Очень зрелая экосистема; знакомый SQL; широкий выбор драйверов, инструментов, операторов и managed services; простой старт для небольших нагрузок. | Основной сервер не предоставляет прозрачную native-модель горизонтального масштабирования записи как distributed SQL; HA, failover и scale-out требуют дополнительной topology, replication, extensions или внешних решений; переход к sharding позднее может стать отдельным архитектурным проектом. |
| [CockroachDB](https://www.cockroachlabs.com/docs/) | Distributed SQL; automatic distribution; Serializable transactions; PostgreSQL-compatible wire ecosystem; встроенная отказоустойчивость. | Self-hosted licensing с ветки 24.3 требует лицензионной модели CockroachDB и в ряде сценариев telemetry/license key; это добавляет коммерческие и эксплуатационные ограничения; PostgreSQL compatibility не устраняет distributed-SQL особенности. |
| [YugabyteDB](https://docs.yugabyte.com/) | Distributed SQL; strong ACID transactions; Apache 2.0; YSQL использует PostgreSQL-compatible wire protocol и большой объём PostgreSQL ecosystem; горизонтальное масштабирование. | PostgreSQL compatibility не полная и имеет документированные различия; отдельный distributed runtime повышает эксплуатационную сложность; приложение всё равно должно проектироваться с учётом распределённой архитектуры. |

## Решение и обоснование

Используется **YDB** как основная операционная СУБД M8 Platform.

Причины предложения:

1. **Распределённость является базовой моделью, а не надстройкой.** YDB автоматически распределяет данные по shards и поддерживает их перемещение и rebalancing при изменении нагрузки и состояния кластера.
2. **Сильные transactional guarantees.** По умолчанию транзакции выполняются в Serializable mode; поддерживаются distributed transactions между несколькими shards.
3. **Соответствие Go-стеку.** Для Go существует официальный `github.com/ydb-platform/ydb-go-sdk/v3`; новые интеграции должны использовать Query Service.
4. **Предсказуемая self-hosted лицензия.** Open-source YDB распространяется под Apache 2.0.
5. **Совместимость с ADR-0001.** Принятый `pressly/goose/v3` имеет upstream YDB driver, поэтому выбор YDB не требует менять механизм миграций.
6. **Подходит для control-plane workloads.** Организации, проекты, конфигурация, состояния provisioning и другие operational entities требуют строгой консистентности и могут расти по количеству tenants и ресурсов.
7. **Не требует PostgreSQL compatibility как архитектурной цели.** Проект новый и не имеет legacy PostgreSQL schema или SQL-кода, который нужно сохранять без изменений.

## Правила использования

Для принятого решения действуют следующие правила:

- YDB становится **default system of record** для новых transactional backend-модулей, если отдельный ADR не обосновывает другое хранилище;
- каждый модуль владеет своими таблицами и не получает право читать внутренние таблицы другого модуля только потому, что они находятся в одной СУБД;
- межмодульные взаимодействия должны идти через согласованные contracts/events, а не через shared-table coupling;
- новые Go-компоненты используют официальный `ydb-go-sdk/v3` и Query Service;
- значения запросов передаются параметрами; SQL/YQL не собирается конкатенацией пользовательских данных;
- primary key проектируется под реальные access patterns и распределение нагрузки; монотонное значение не используется как единственный первый компонент распределяющего ключа;
- приложение не должно рассчитывать на PostgreSQL-specific функции или на `SERIAL` / `AUTO_INCREMENT`;
- транзакции проектируются как можно более локальными; distributed transaction применяется только когда атомарность действительно требуется;
- retries выполняются только с учётом idempotency и общего deadline/cancellation budget;
- schema changes выполняются через goose согласно [ADR-0001](0001-database-migrations.md);
- конкретные topology, replication policy, backup/restore, capacity planning и production deployment YDB должны быть определены отдельно до production rollout;
- выбор YDB Tables не означает автоматического выбора YDB Topics или Coordination для messaging/locking — такие решения принимаются по собственным требованиям.

## Последствия

Положительные последствия:

- horizontal scale и high availability учитываются с начала проекта;
- application layer не должен реализовывать собственный sharding;
- строгие транзакционные гарантии доступны между распределёнными данными;
- Go SDK, CLI, local Docker и migration integration образуют единый toolchain;
- open-source self-hosted вариант не зависит от коммерческого license key;
- архитектура не привязывается к PostgreSQL ecosystem ради совместимости с отсутствующим legacy.

Ограничения и риски:

- разработчикам потребуется знание YDB/YQL и особенностей distributed schema design;
- неправильная форма primary key способна создать hotspot и свести преимущества horizontal scaling к минимуму;
- distributed transactions имеют более высокую стоимость и latency, поэтому границы агрегатов и модулей должны проектироваться внимательно;
- часть привычных PostgreSQL конструкций и инструментов неприменима;
- эксплуатация self-hosted distributed database сложнее single-node PostgreSQL;
- до production необходимо подтвердить backup/restore, upgrade, rolling restart, failure recovery, observability и capacity model на выбранной topology;
- если реальные требования покажут, что workload лучше решается PostgreSQL или специализированным storage, изменение должно быть оформлено новым ADR.

## Критерии принятия

Перед первым production rollout необходимо выполнить минимальный технический PoC:

1. поднять локальный или тестовый YDB cluster;
2. применить schema migrations через goose;
3. реализовать небольшой Go repository через `ydb-go-sdk/v3` Query Service;
4. проверить CRUD и транзакцию, затрагивающую более одной таблицы;
5. проверить concurrent optimistic update или другой реальный conflict scenario;
6. проверить поведение клиента при временной недоступности/перезапуске узла в доступной тестовой topology;
7. выполнить базовый `EXPLAIN` ключевых запросов и подтвердить отсутствие очевидного hotspot в выбранной модели ключей;
8. задокументировать ограничения PoC и вопросы, требующие production-level проверки.

PoC подтверждает применимость инструмента, но не заменяет нагрузочное тестирование и production readiness review.

## Связи

- [ADR-0001: система миграций базы данных](0001-database-migrations.md) — выбран `pressly/goose/v3`.
- [Обзор архитектуры](../architecture/overview.md).
- [Границы модулей](../architecture/module-boundaries.md).
- YDB: [Architecture](https://ydb.tech/docs/en/concepts/architecture), [Transactions](https://ydb.tech/docs/en/concepts/transactions), [Go SDK](https://ydb.tech/docs/en/reference/ydb-sdk/), [Open-source downloads and license](https://ydb.tech/docs/en/downloads/ydb-open-source-database).
- PostgreSQL: [High Availability, Load Balancing, and Replication](https://www.postgresql.org/docs/current/high-availability.html).
- CockroachDB: [Licensing FAQs](https://www.cockroachlabs.com/docs/stable/licensing-faqs).
- YugabyteDB: [PostgreSQL compatibility](https://docs.yugabyte.com/stable/develop/postgresql-compatibility/).
- Связанные ADR: [ADR-0001](0001-database-migrations.md).
- Заменяет: нет.
- Заменено: нет.
