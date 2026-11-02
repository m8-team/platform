# Выполнение большого `02_attach_partitions.sql`

Скрипт `run_attach_partitions.py` предназначен **только для** файла `02_attach_partitions.sql`, сгенерированного предыдущим генератором миграции ClickHouse `zendendb_import` → `zendendb`. Использует стандартную библиотеку Python 3.10+ (без `pip install`).

## Перед началом

- Выполнен `01_create_replicated_tables.sql`: в `zendendb` существуют все необходимые **пустые** реплицируемые таблицы.
- На **этом же ClickHouse-хосте** доступны исходные таблицы `zendendb_import`, источник не меняется во время миграции.
- Целевая база `zendendb` имеет `ENGINE = Replicated`. У пользователя есть права на `ALTER` и чтение метаданных/таблиц.
- Запускайте только **на одной реплике каждого шарда**. Не запускайте параллельно на разных репликах, чтобы не дублировать данные. Для нескольких шардов нужен план отдельно по каждому шарду.
- Сохраняйте неизменными SQL-файл и `*.progress.jsonl`. **Журнал — часть процедуры восстановления; не удаляйте его.**

## 1. Пароль и проверка файла (без изменений)

```bash
read -rs CH_PASSWORD
export CH_PASSWORD

python3 run_attach_partitions.py \
  --sql ./migration-sql/02_attach_partitions.sql
```

Команда без `--execute` проверяет синтаксис, количество операций и показывает первые 10.

## 2. Выполнить первые 10 операций

```bash
python3 -u run_attach_partitions.py \
  --sql ./migration-sql/02_attach_partitions.sql \
  --url https://rc1d-7hqiainm2j9d2vhd.mdb.yandexcloud.net:8443 \
  --user analytics \
  --source zendendb_import \
  --target zendendb \
  --ca-file ./certs/RootCA.pem \
  --execute --limit 10
```

Файл `./migration-sql/02_attach_partitions.sql.progress.jsonl` создаётся автоматически и сохраняет этап каждой операции (`STARTED`, `OK`, `UNCERTAIN`) и идентификатор ClickHouse-запроса.

## 3. Выполнить оставшиеся операции (можно повторять)

```bash
python3 -u run_attach_partitions.py \
  --sql ./migration-sql/02_attach_partitions.sql \
  --url https://rc1d-7hqiainm2j9d2vhd.mdb.yandexcloud.net:8443 \
  --user analytics \
  --source zendendb_import \
  --target zendendb \
  --ca-file ./certs/RootCA.pem \
  --execute
```

Успешно завершённые операции пропускаются. При Ctrl+C, таймауте или ошибке **неизвестный результат `ATTACH` не повторяется автоматически**; выполнение останавливается до сверки данных. Дополнительно можно использовать `--sleep 0.2` между запросами и `--timeout 7200` (в секундах).

## 4. Проверить локальный прогресс (без подключения к серверу)

```bash
python3 run_attach_partitions.py \
  --sql ./migration-sql/02_attach_partitions.sql \
  --status
```

## 5. Если операция прервана или завершилась ошибкой

Сначала найдите в журнале `query_id`, имя таблицы и номер партиции. До повторного запуска проверьте на том же хосте в ClickHouse:

```sql
SELECT query_id, query, elapsed
FROM system.processes
WHERE query_id = 'migrate-attach-...';

SELECT database, table, partition_id,
       count() AS parts, sum(rows) AS rows
FROM system.parts
WHERE active AND table = 'ИМЯ_ТАБЛИЦЫ'
  AND partition_id = 'ID_ПАРТИЦИИ'
  AND database IN ('zendendb_import', 'zendendb')
GROUP BY database, table, partition_id
ORDER BY database;
```

Проверьте также `system.query_log` для нужного `query_id` (при включённом журналировании) и убедитесь, что запрос более не выполняется. Наличие строк в целевой партиции **само по себе не гарантирует**, что все исходные данные были перенесены; сверяйте количество строк, части и возможные слияния/TTL.

После ручной проверки можно пометить соответствующую операцию, например с индексом `123`, как выполненную:

```bash
python3 run_attach_partitions.py \
  --sql ./migration-sql/02_attach_partitions.sql \
  --url https://rc1d-7hqiainm2j9d2vhd.mdb.yandexcloud.net:8443 \
  --user analytics --ca-file ./certs/RootCA.pem \
  --resolve done --index 123 --ack I_VERIFIED_CLICKHOUSE
```

Или разрешить **однократный повтор**, только если точно подтверждено, что целевая партиция пуста и исходный запрос прекратил работу:

```bash
python3 run_attach_partitions.py \
  --sql ./migration-sql/02_attach_partitions.sql \
  --url https://rc1d-7hqiainm2j9d2vhd.mdb.yandexcloud.net:8443 \
  --user analytics --ca-file ./certs/RootCA.pem \
  --resolve retry --index 123 --ack I_VERIFIED_CLICKHOUSE
```

После разрешения снова выполните команду из шага 3. `--resolve` требует явного подтверждения и не выполняет `ATTACH` напрямую.

## Ограничения и безопасность

- `ATTACH PARTITION ... FROM` копирует данные; повторное выполнение может добавить дубликаты. Скрипт откажется выполнять операцию, если партиция уже содержит активные части и эта операция отсутствует как успешно завершённая в журнале.
- Скрипт считает успешным `ATTACH` только после получения успешного HTTP-ответа с `wait_end_of_query=1` и записи события в локальный журнал. При потере соединения результат считается неопределённым.
- Этот исполнитель **не** запускает `01`/`03`/`04`/`05`/`06`; особенно не используйте его для `INSERT` из Log или для DDL MySQL. Он принимает только конкретный SQL `ALTER TABLE ... ATTACH PARTITION ID ... FROM ...` в одну строку, как генерирует ваш исходный скрипт.
- Скрипт не выполняет автоматическую сверку полной целостности. После окончания проверьте исходные и целевые строки по таблицам и партициям, состояния реплик и ошибочные запросы. Не удаляйте `zendendb_import` до окончания проверки.
- `--status` и запуск без `--execute` не требуют пароля. Для реальных запросов пароль читается из `CH_PASSWORD` или запрашивается интерактивно; не передавайте его в командной строке.
