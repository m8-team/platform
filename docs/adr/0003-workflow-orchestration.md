# `ADR-0003`: Система оркестрации длительных процессов

- Идентификатор: `ADR-0003`.
- Дата: `2026-09-19`.
- Статус: `Accepted`.
- Подтверждение согласования: решение владельца проекта от `2026-09-19` — выбрать Temporal как стандартную систему durable orchestration; ссылка на фиксирующий commit будет добавлена после записи решения.

## Контекст

M8 Platform требуется единый механизм оркестрации длительных и отказоустойчивых процессов, которые могут включать несколько сервисов, внешние API, ожидания, повторные попытки, компенсации и ручные или событийные продолжения.

Типичные задачи этого класса:

- provisioning и изменение жизненного цикла ресурсов;
- многошаговые операции между модулями;
- процессы с ожиданием внешнего события или подтверждения;
- saga-процессы с компенсациями;
- длительные операции, которые должны продолжаться после рестарта сервиса;
- фоновые операции с явной историей состояния и наблюдаемым прогрессом.

Система должна:

- сохранять состояние исполнения при сбоях процессов и инфраструктуры;
- поддерживать длительные workflows от секунд до дней, месяцев и более;
- иметь официальный Go SDK;
- поддерживать retries, timeouts, timers, signals/updates и compensation patterns;
- обеспечивать безопасное изменение workflow-кода при наличии уже запущенных исполнений;
- позволять self-hosted эксплуатацию без обязательной привязки к одному cloud provider;
- давать UI/CLI и историю выполнения для диагностики;
- не требовать хранить orchestration state в бизнес-таблицах приложения;
- не заменять broker/event bus и Kubernetes reconciliation там, где они лучше соответствуют задаче.

При подготовке решения использованы установленные skills `temporal-developer` и `temporal-workflow-design-critic`.

## Границы решения

Этот ADR выбирает систему для **durable application workflow orchestration**.

Он не выбирает:

- Kubernetes как платформу развёртывания;
- механизм continuous reconciliation desired/actual state;
- message broker или event streaming platform;
- CI/CD pipeline engine;
- batch/ML pipeline engine;
- scheduler для простых периодических задач, если durable workflow не требуется.

Kubernetes controller/operator остаётся предпочтительным механизмом для постоянного reconciliation Kubernetes/custom-resource состояния. Temporal применяется, когда процесс имеет конечный или управляемый жизненный цикл, бизнес-состояние, длительные ожидания, retries, компенсации или координацию нескольких шагов.

## Рассмотренные варианты

| Вариант | Преимущества | Ограничения и последствия |
| --- | --- | --- |
| [Temporal](https://docs.temporal.io/) | Durable execution с event history и replay; workflows описываются кодом; официальные SDK, включая Go; retries, timers, signals, queries, updates, child workflows и schedules; self-hosted и managed Cloud; MIT license; развитая модель workflow versioning и replay testing. | Workflow code должен быть deterministic; Activities должны быть идемпотентными с учётом at-least-once выполнения; нужно управлять history growth и versioning; self-hosted Temporal требует отдельного persistence backend — Cassandra, PostgreSQL или MySQL для production — и не использует YDB как штатный persistence store. |
| [Cadence](https://cadenceworkflow.io/) | Apache 2.0; durable code-first workflows; Go и Java SDK; self-hosted; зрелая модель, из которой исторически вырос Temporal. | Меньшая экосистема и набор SDK; для нового проекта меньше оснований выбирать Cadence вместо Temporal при схожей модели программирования; переносимость workflow-кода между ними не является гарантией. |
| [Argo Workflows](https://argoproj.github.io/argo-workflows/) | CNCF-проект; Kubernetes-native; удобен для DAG/steps, container jobs, batch, CI/CD, ML/data pipelines; естественная интеграция с Kubernetes. | Основная единица исполнения — Kubernetes workload/container; менее естественен для code-first application workflows, ожиданий пользовательских событий и сложных saga; жёстко привязан к Kubernetes runtime и CRD-модели. |
| [AWS Step Functions](https://docs.aws.amazon.com/step-functions/) | Managed service; встроенные retries, branching, service integrations, визуализация; не требуется эксплуатация control plane. | AWS vendor lock-in; workflow задаётся state-machine DSL; нет self-hosted варианта; pricing зависит от transitions/execution model; Standard Workflow имеет platform quotas и максимальную длительность одного execution в один год. |
| Собственная orchestration/state-machine библиотека | Полный контроль над моделью и минимальный внешний runtime. | Придётся самостоятельно реализовать durable state, timers, retries, locking, recovery, visibility, workflow versioning, replay/compatibility, cancellation и operational tooling; высокий риск скрытых distributed-systems ошибок. |

## Решение и обоснование

Используется **Temporal** как стандартная система durable orchestration M8 Platform.

Причины:

1. **Durable execution как базовая модель.** Workflow state и event history сохраняются Temporal Service, поэтому процесс может продолжаться после рестарта worker или инфраструктурного сбоя.
2. **Code-first orchestration.** Workflow описывается Go-кодом вместо отдельного DSL, что соответствует основному backend-стеку M8.
3. **Поддержка длительных процессов.** Durable timers, signals, updates и Continue-As-New позволяют моделировать ожидания и процессы с длительным жизненным циклом.
4. **Явная модель failure handling.** Activities имеют retries/timeouts/heartbeats, а workflow остаётся детерминированным описанием управления процессом.
5. **Безопасная эволюция.** Temporal предоставляет механизмы workflow versioning и replay, необходимые для изменения кода при существующих длительных executions.
6. **Наблюдаемость исполнения.** Event History, Visibility, Web UI и CLI позволяют диагностировать состояние и причины задержек без создания собственного workflow audit subsystem.
7. **Self-hosted возможность.** Temporal Service можно эксплуатировать самостоятельно; managed Temporal Cloud остаётся альтернативой без изменения application workflow model.
8. **Более подходящая модель, чем Kubernetes DAG.** Для application-level workflows Temporal не требует превращать каждый шаг бизнес-процесса в Kubernetes Job/Pod.
9. **Меньше platform lock-in, чем managed-only orchestration.** Workflow-код не привязан к AWS Step Functions или другой cloud-specific state-machine service.

## Правила использования

Для принятого решения действуют следующие правила:

- Temporal является default orchestration engine для новых long-running и multi-step процессов, когда durability процесса является требованием;
- workflow definitions содержат только deterministic orchestration logic;
- любые сетевые вызовы, I/O, обращения к YDB и другие side effects выполняются через Activities;
- Activities проектируются идемпотентными или используют явный idempotency mechanism;
- retries и timeouts задаются осознанно для конкретных Activities, а не одной глобальной политикой;
- большие payloads не передаются через Workflow History; вместо этого передаются идентификаторы/ссылки на данные;
- для потенциально длинной или высокоактивной истории проектируется Continue-As-New или partitioning strategy;
- изменение workflow-кода должно учитывать replay compatibility и workflow versioning;
- replay tests выполняются перед rollout изменений, способных повлиять на уже запущенные workflows;
- Workflow ID должен отражать бизнес-идентичность процесса и поддерживать требуемую deduplication policy;
- Task Queues используются для routing workers, а не как замена Kafka/YDB Topics или общей business-message queue;
- для обычной фоновой задачи без orchestration предпочтительна более простая модель; Temporal Workflow не создаётся только ради обёртки одного короткого вызова;
- для continuous reconciliation Kubernetes/custom resources используется controller/reconciler pattern; Temporal может запускать или координировать конечную операцию, но не заменяет бесконечный reconciliation loop;
- production deployment Temporal Service, persistence backend, namespaces, security, worker topology, observability и disaster recovery определяются отдельным operational design.

## Self-hosted persistence

Выбор YDB в [ADR-0002](0002-primary-database.md) относится к application system of record и **не означает**, что YDB используется как persistence backend Temporal Service.

Штатно Temporal поддерживает persistence через Cassandra, PostgreSQL и MySQL; SQLite предназначен для development/testing. Поэтому self-hosted Temporal создаёт отдельную инфраструктурную зависимость.

Для первого self-hosted production варианта предлагается рассматривать **PostgreSQL как persistence store Temporal Service**, если отдельное operational решение не выберет Cassandra или MySQL. Это предложение не является частью выбора application database и должно быть подтверждено отдельным ADR или deployment design.

Advanced Visibility должна проектироваться по возможностям выбранной версии Temporal; не следует заранее вводить Elasticsearch, если поддерживаемый SQL Visibility удовлетворяет требованиям.

## Последствия

Положительные последствия:

- исчезает необходимость реализовывать собственные state machines, retry loops и recovery coordination в каждом сервисе;
- долгоживущие процессы получают единый execution model;
- процессы становятся наблюдаемыми через history/visibility;
- отказ worker не означает потерю orchestration state;
- код orchestration остаётся рядом с Go application code;
- появляется единый подход к saga, timers, external signals и compensation;
- self-hosted и managed deployment используют одну application programming model.

Ограничения и риски:

- Temporal становится отдельным critical infrastructure component;
- self-hosted deployment требует отдельной persistence database помимо application YDB;
- нарушение determinism может блокировать replay существующих workflows;
- неправильные retries могут повторить side effects;
- чрезмерные payloads, signal volume или history growth могут ухудшить работу workflow;
- разработчикам необходимо освоить Temporal-specific concepts: Activities, Task Queues, event history, replay, Continue-As-New, versioning и worker deployment;
- нельзя переносить обычные synchronous service methods в Workflows механически;
- Temporal не устраняет необходимость проектировать transactional boundaries и idempotency на стороне application services.

## Критерии принятия

Перед первым production rollout необходимо выполнить минимальный PoC:

1. поднять Temporal development server или test deployment;
2. реализовать Go Workflow минимум из двух Activities;
3. добавить retry/timeout policy и проверить восстановление после падения worker;
4. добавить durable timer и external Signal или Update;
5. проверить idempotency повторно выполняемой Activity;
6. выполнить replay test после совместимого изменения workflow-кода;
7. проверить Web UI/CLI visibility и диагностирование failed Activity;
8. смоделировать compensation/saga для частично выполненного процесса;
9. проверить Continue-As-New или хотя бы оценить history growth для длительного сценария;
10. зафиксировать требования к production worker topology и self-hosted persistence.

## Связи

- [ADR-0002: основная операционная база данных](0002-primary-database.md) — application data хранится в YDB.
- [Обзор архитектуры](../architecture/overview.md).
- Temporal: [Documentation](https://docs.temporal.io/), [Open-source platform](https://temporal.io/), [Server persistence](https://docs.temporal.io/self-hosted-guide/defaults).
- Cadence: [Open-source workflow engine](https://cadenceworkflow.io/docs/concepts/open-source-workflow-engine).
- Argo Workflows: [Documentation](https://argoproj.github.io/argo-workflows/).
- AWS Step Functions: [Workflow types](https://docs.aws.amazon.com/step-functions/latest/dg/choosing-workflow-type.html).
- Связанные ADR: [ADR-0002](0002-primary-database.md).
- Заменяет: нет.
- Заменено: нет.
