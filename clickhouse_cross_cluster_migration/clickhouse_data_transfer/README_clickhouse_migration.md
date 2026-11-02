# Миграция ClickHouse: `zendendb_import` → `zendendb` (Replicated)

**Версия плана:** 2026-11-02 — `Log` преобразуется в `ReplicatedMergeTree`, а `MySQL` сохраняет исходный внешний движок.

Комплект содержит:

- `clickhouse_replicated_migration.py` — генератор SQL по реальному DDL источника. **Ничего не выполняет**.
- `clickhouse_migration_inventory.csv` — список 378 объектов из предоставленного CSV.
- `README_clickhouse_migration.md` — инструкция.

## Матрица миграции

| Источник | Количество | В целевой базе |
|---|---:|---|
| MergeTree | 112 | ReplicatedMergeTree |
| ReplacingMergeTree | 128 | ReplicatedReplacingMergeTree |
| Log | 8 | ReplicatedMergeTree, `ORDER BY tuple()` |
| MySQL | 3 | MySQL, **без копирования данных** |
| View | 127 | View |

Общее количество объектов: **378**.

**Внимание:** MySQL-таблицы по-прежнему читают удалённый MySQL. Реплицируется только определение таблицы (метаданные). У каждого ClickHouse-сервера должен быть доступ к соответствующему MySQL, а также верные учётные данные. SQL-файлы, создаваемые скриптом, **не сохраняют пароли MySQL**: вместо пятого аргумента `MySQL(...)` ставится `__REPLACE_MYSQL_PASSWORD__`. Перед выполнением файла `05_create_mysql_tables.sql` надо подставить реальные пароли **в защищённой среде**. Если используются SQL named collections, проверьте наличие коллекций на каждом сервере.

## Предусловия

1. `zendendb_import` — восстановленная исходная база с исходными данными.
2. `zendendb` существует с `ENGINE = Replicated` и **не содержит объектов**. Генератор откажется работать с непустой целевой базой.
3. Источник и целевой сервер/база доступны для SQL на одном ClickHouse-хосте. Если есть несколько шардов, перенос данных каждого шарда планируется отдельно.
4. Обеспечьте согласованное состояние источника и остановите запись на время cutover.
5. Для генератора необходим доступ к `system.databases`, `system.tables`, `system.parts`; для выполнения DDL/переноса нужны соответствующие права.
6. Перед миграцией проверьте доступное дисковое пространство, Keeper, репликацию и возможность восстановления исходной базы.

## Генерация SQL

```bash
read -r -s CH_PASSWORD
export CH_PASSWORD

python3 clickhouse_replicated_migration.py \
  --url https://<CLICKHOUSE_FQDN>:8443 \
  --user admin \
  --source zendendb_import \
  --target zendendb \
  --out ./migration-sql

unset CH_PASSWORD
```

Для проверки текущего списка объектов относительно загруженного CSV добавьте:

```bash
--csv websql_query_result_2026_11_2_28.csv
```

При необходимости добавьте `--ca-file /path/to/ca.crt`.

Генератор берёт `create_table_query` из `system.tables`, поэтому не пытается восстановить ключи и колонки из инвентаризационного CSV.

## Порядок выполнения

1. `01_create_replicated_tables.sql` — 240 таблиц семейства MergeTree. Ключи, секционирование, TTL, настройки и параметры ReplacingMergeTree переносятся из DDL. Для базы с движком `Replicated` используются значения Keeper/replica из конфигурации кластера; явные пути **не задаются**.
2. `02_attach_partitions.sql` — поочерёдный перенос партиций. `ATTACH PARTITION ID ... FROM ...` выполнить **один раз на одной реплике каждого шарда**; запрос может копировать данные, учитывайте диск и время. При неопределённом результате операции повтор не запускать до сверки.
3. `03_create_log_tables.sql` — 8 таблиц Log → `ReplicatedMergeTree() ORDER BY tuple()`.
4. `04_copy_log_data.sql` — однократное `INSERT SELECT` из соответствующих Log-таблиц (повтор приведёт к дублированию).
5. `05_create_mysql_tables.sql` — 3 таблицы MySQL с исходными колонками, адресами и параметрами внешних источников. Перед выполнением **подставить пароли** (специальный placeholder вместо секретов); никаких `INSERT` для MySQL не делать.
6. `06_create_views.sql` — 127 обычных View. Проверьте порядок зависимостей и корректность ссылок на другие базы.
7. `07_validate.sql` — сравнить объекты, активные части и состояние реплик. Дополнительно проверить `SELECT count()`/контрольные агрегаты по критичным таблицам; MySQL нужно проверить отдельным `SELECT` из ClickHouse.

Не используйте `ON CLUSTER` для DDL в `Replicated`-базе: определения распространяются автоматически. Не запускайте `INSERT`/`ATTACH` для одной и той же партии данных одновременно на всех репликах.

## Контрольные SQL

```sql
SELECT engine, count() AS n
FROM system.tables
WHERE database = 'zendendb'
GROUP BY engine ORDER BY engine;
```

Ожидаемо: `ReplicatedMergeTree` — 120, `ReplicatedReplacingMergeTree` — 128, `MySQL` — 3, `View` — 127, **всего 378**.

```sql
SELECT table, total_replicas, active_replicas, is_readonly, queue_size
FROM system.replicas
WHERE database = 'zendendb'
ORDER BY table;
```

И отдельно:

```sql
SELECT count() FROM zendendb.b_catalog_store_mysql;
```

Эта последняя команда **делает запрос в удалённый MySQL**, не читает локальную копию.

## Предупреждения

- Нельзя бездумно заменять все упоминания имени исходной базы в SQL-представлениях: генератор делает текстовую замену `zendendb_import.` → `zendendb.`, поэтому SQL для View нужно просмотреть вручную, особенно если в запросах встречаются строковые литералы.
- Представления могут зависеть друг от друга. Их DDL надо выполнять в порядке зависимостей.
- `MySQL` не входит в локальные данные ClickHouse и не попадает в его резервные копии как набор строк. Для автономного бэкапа нужны дополнительные действия в MySQL.
- Не удаляйте `zendendb_import` до полного сравнения данных, проверки View и реплик.
- Ключи/пароли нельзя сохранять в общедоступных файлах или в Git; содержимое `05_create_mysql_tables.sql` без подстановки паролей безопаснее для совместной проверки.
