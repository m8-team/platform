# Фильтры API

Пакет `internal/platform/filter` предоставляет общий технический слой для
разбора CEL-фильтров в API списков. Он не зависит от Resource Manager и других
бизнес-модулей. Пакет доступен всем сервисам внутри Go-модуля M8 Platform;
каждый сервис объявляет собственную схему доступных полей.

Основной поток данных:

```text
строка filter
    -> platform/filter
    -> неизменяемый Conjunction
    -> транслятор бизнес-модуля
    -> типизированный ports.<Resource>Filter
    -> repository
```

Пакет отвечает только за синтаксис CEL, типы литералов, нормализацию выражения
и защитные лимиты. Проверка бизнес-значений, например допустимых состояний или
окружений, остаётся в модуле-владельце ресурса.

## Быстрый старт

Объявите только те поля, которые доступны клиенту конкретного метода списка:

```go
import platformfilter "github.com/m8-team/platform/internal/platform/filter"

var organizationFilterParser, organizationFilterParserError =
    platformfilter.NewCELParser(platformfilter.CELParserConfig{
        Variables: []platformfilter.Variable{
            platformfilter.ScalarVariable("state", platformfilter.StringKind),
            platformfilter.ScalarVariable("name", platformfilter.StringKind),
            platformfilter.MapVariable(
                "labels",
                platformfilter.StringKind,
                platformfilter.StringKind,
            ),
        },
    })
```

Ошибку создания парсера необходимо проверить при запуске сервиса. Один
экземпляр парсера можно безопасно использовать одновременно из нескольких
запросов.

```go
if organizationFilterParserError != nil {
    return fmt.Errorf(
        "initialize organization filter parser: %w",
        organizationFilterParserError,
    )
}
```

Разбор фильтра:

```go
conjunction, err := organizationFilterParser.Parse(
    `state in ["ACTIVE", "SUSPENDED"] && labels.environment == "prod"`,
)
if err != nil {
    return fmt.Errorf("invalid organization filter: %w", err)
}

for _, predicate := range conjunction.Predicates() {
    field := predicate.Field()

    fmt.Printf(
        "field=%s operator=%d values=%d\n",
        field.Name(),
        predicate.Operator(),
        len(predicate.Values()),
    )
}
```

Пустая строка или строка только из пробелов возвращает пустой `Conjunction`
без ошибки.

## Пример трансляции в фильтр модуля

Нейтральные предикаты нельзя передавать напрямую в repository. Модуль должен
полностью перевести их в собственную типизированную модель:

```go
type ServiceListFilter struct {
    Environment *string
    LabelsEqual map[string]string
}

func translateServiceFilter(
    conjunction platformfilter.Conjunction,
) (ServiceListFilter, error) {
    result := ServiceListFilter{
        LabelsEqual: make(map[string]string),
    }

    for _, predicate := range conjunction.Predicates() {
        field := predicate.Field()
        values := predicate.Values()

        switch field.Name() {
        case "environment":
            if predicate.Operator() != platformfilter.EqualsOperator ||
                len(values) != 1 {
                return ServiceListFilter{}, fmt.Errorf(
                    "environment supports only equality",
                )
            }
            if result.Environment != nil {
                return ServiceListFilter{}, fmt.Errorf(
                    "duplicate environment predicate",
                )
            }

            value, ok := values[0].AsString()
            if !ok {
                return ServiceListFilter{}, fmt.Errorf(
                    "environment must be a string",
                )
            }
            if value != "prod" && value != "stage" && value != "dev" {
                return ServiceListFilter{}, fmt.Errorf(
                    "unsupported environment %q",
                    value,
                )
            }
            result.Environment = &value

        case "labels":
            if predicate.Operator() != platformfilter.EqualsOperator ||
                len(values) != 1 {
                return ServiceListFilter{}, fmt.Errorf(
                    "labels support only equality",
                )
            }

            keyLiteral, hasKey := field.Key()
            if !hasKey {
                return ServiceListFilter{}, fmt.Errorf("label key is required")
            }
            key, keyOK := keyLiteral.AsString()
            value, valueOK := values[0].AsString()
            if !keyOK || !valueOK || key == "" {
                return ServiceListFilter{}, fmt.Errorf(
                    "label key and value must be strings",
                )
            }
            if _, duplicate := result.LabelsEqual[key]; duplicate {
                return ServiceListFilter{}, fmt.Errorf(
                    "duplicate label predicate %q",
                    key,
                )
            }
            result.LabelsEqual[key] = value

        default:
            return ServiceListFilter{}, fmt.Errorf(
                "unsupported filter field %q",
                field.Name(),
            )
        }
    }

    if len(result.LabelsEqual) == 0 {
        result.LabelsEqual = nil
    }
    return result, nil
}
```

Такой транслятор является границей бизнес-модуля. Именно здесь проверяются:

- разрешённые операторы для каждого поля;
- количество и тип значений;
- доменные enum-значения;
- пустые и повторяющиеся условия;
- канонический порядок значений, необходимый для стабильных page token.

## Примеры схем для разных списков

### Организации

```go
parser, err := platformfilter.NewCELParser(platformfilter.CELParserConfig{
    Variables: []platformfilter.Variable{
        platformfilter.ScalarVariable("state", platformfilter.StringKind),
        platformfilter.ScalarVariable("name", platformfilter.StringKind),
        platformfilter.MapVariable(
            "labels",
            platformfilter.StringKind,
            platformfilter.StringKind,
        ),
    },
})
```

Примеры запросов:

```cel
state == "ACTIVE"
state in ["ACTIVE", "SUSPENDED"]
name == "Production"
labels.environment == "prod"
labels["example.com/team"] == "platform"
state == "ACTIVE" && labels.environment == "prod"
```

Тот же фильтр через REST:

```bash
curl --get 'http://127.0.0.1:8080/resource-manager/v1/organizations' \
  --data-urlencode 'page_size=50' \
  --data-urlencode 'filter=state == "ACTIVE" && labels.environment == "prod"'
```

`curl --data-urlencode` самостоятельно экранирует пробелы, кавычки и
операторы. В Postman те же значения следует добавить на вкладке **Params**:

| Key | Value |
|---|---|
| `page_size` | `50` |
| `filter` | `state == "ACTIVE" && labels.environment == "prod"` |

Пример вызова через gRPC:

```bash
grpcurl -plaintext \
  -d '{"page_size":50,"filter":"state == \"ACTIVE\" && labels.environment == \"prod\""}' \
  127.0.0.1:9090 \
  m8.platform.resourcemanager.v1.OrganizationService/ListOrganizations
```

### Сервисы и их окружения

```go
parser, err := platformfilter.NewCELParser(platformfilter.CELParserConfig{
    Variables: []platformfilter.Variable{
        platformfilter.ScalarVariable("state", platformfilter.StringKind),
        platformfilter.ScalarVariable("name", platformfilter.StringKind),
        platformfilter.ScalarVariable(
            "environment",
            platformfilter.StringKind,
        ),
        platformfilter.MapVariable(
            "labels",
            platformfilter.StringKind,
            platformfilter.StringKind,
        ),
    },
})
```

Примеры запросов:

```cel
environment == "prod"
environment == "dev" && state == "ACTIVE"
name == "payments" && labels.team == "billing"
```

### Список с разными скалярными типами

```go
parser, err := platformfilter.NewCELParser(platformfilter.CELParserConfig{
    Variables: []platformfilter.Variable{
        platformfilter.ScalarVariable("enabled", platformfilter.BoolKind),
        platformfilter.ScalarVariable("attempts", platformfilter.IntKind),
        platformfilter.ScalarVariable("revision", platformfilter.UintKind),
        platformfilter.ScalarVariable("ratio", platformfilter.DoubleKind),
        platformfilter.ScalarVariable("payload", platformfilter.BytesKind),
    },
})
```

Примеры запросов:

```cel
enabled == true
attempts == -1
attempts in [1, 2, 3]
revision == 9u
ratio in [-1.5, 0.5]
payload == b"ready"
```

## Поддерживаемый синтаксис

| Возможность | Пример |
|---|---|
| Равенство | `name == "payments"` |
| Обратное равенство | `"payments" == name` |
| Вхождение в список | `state in ["ACTIVE", "SUSPENDED"]` |
| Доступ к map через точку | `labels.environment == "prod"` |
| Доступ к map через индекс | `labels["example.com/team"] == "platform"` |
| Объединение условий | `state == "ACTIVE" && name == "payments"` |
| Отрицательное число | `attempts == -1` |

Выражение всегда представляет AND-конъюнкцию. `Conjunction.Predicates()`
возвращает условия в порядке исходного выражения и предоставляет защитную
копию. `Predicate.Values()` и байты из `Literal.AsBytes()` также возвращаются
как защитные копии.

## Поддерживаемые типы

| Тип | CEL | Метод чтения |
|---|---|---|
| `StringKind` | `"value"` | `AsString()` |
| `BoolKind` | `true`, `false` | `AsBool()` |
| `IntKind` | `10`, `-1` | `AsInt()` |
| `UintKind` | `10u` | `AsUint()` |
| `DoubleKind` | `1.5`, `-0.5` | `AsDouble()` |
| `BytesKind` | `b"value"` | `AsBytes()` |

Ключи CEL map могут иметь тип `StringKind`, `BoolKind`, `IntKind` или
`UintKind`. Синтаксис через точку доступен только для строковых ключей;
остальные ключи задаются через индекс.

## Настройка лимитов

Значение `0` включает безопасный лимит по умолчанию:

| Параметр | По умолчанию | Назначение |
|---|---:|---|
| `MaxExpressionRunes` | `1024` | максимальная длина выражения |
| `MaxRecursionDepth` | `32` | максимальная глубина синтаксического дерева |
| `MaxPredicates` | `32` | максимальное количество AND-условий |
| `MaxListItems` | `100` | максимальное количество элементов одного `in` |

Лимиты можно уменьшить для конкретного метода API:

```go
parser, err := platformfilter.NewCELParser(platformfilter.CELParserConfig{
    MaxExpressionRunes: 512,
    MaxRecursionDepth:  16,
    MaxPredicates:      10,
    MaxListItems:       20,
    Variables: []platformfilter.Variable{
        platformfilter.ScalarVariable("state", platformfilter.StringKind),
    },
})
```

Отрицательные значения лимитов считаются ошибкой конфигурации.

## Обработка ошибок

Ошибки конфигурации и пользовательского выражения разделены:

```go
parser, err := platformfilter.NewCELParser(config)
if err != nil {
    if errors.Is(err, platformfilter.ErrInvalidConfiguration) {
        // Ошибка схемы метода API. Сервис не должен запускаться.
        return fmt.Errorf("initialize list filter: %w", err)
    }
    return err
}

conjunction, err := parser.Parse(rawFilter)
if err != nil {
    if errors.Is(err, platformfilter.ErrInvalidExpression) {
        // Ошибка клиента. На transport-границе обычно преобразуется в
        // codes.InvalidArgument или HTTP 400.
        return fmt.Errorf("invalid list filter: %w", err)
    }
    return err
}
```

## Что намеренно не поддерживается

Следующие выражения отклоняются:

```cel
name.startsWith("pay")                 // функции
["prod"].exists(value, value == name) // макросы
state == "ACTIVE" || state == "FAILED" // OR
!(state == "ACTIVE")                  // логическое отрицание
name == ("pay" + "ments")             // вычисляемое значение
unknown == "value"                    // необъявленное поле
state in []                            // пустой список
```

Новые операторы следует добавлять только тогда, когда все использующие их
модули могут корректно преобразовать выражение в storage query. Нельзя сначала
применить pagination, а затем фильтровать полученную страницу в памяти: это
нарушает `page_size`, page token и `total_size`.

## Архитектурные правила

- Создавайте отдельную схему парсера для каждого метода списка.
- Создавайте парсер один раз при composition/startup, а не на каждый запрос.
- Не передавайте raw CEL, `Predicate` или строковые имена полей в repository.
- Repository должен получать только типизированный фильтр своего модуля.
- Database adapter должен сопоставлять поля со статическими SQL-фрагментами;
  значения и map-ключи должны передаваться как bind parameters.
- Parent scope, authorization visibility и `show_deleted` добавляются отдельно
  и не могут переопределяться пользовательским CEL-фильтром.
- Повторяющиеся или эквивалентные значения следует канонизировать до
  вычисления hash для page token.
- Для memory и database adapters нужны одинаковые contract tests на фильтрацию
  и pagination.

## Чек-лист нового метода списка

1. Объявить доступные CEL-поля и их типы.
2. Создать и проверить парсер при запуске сервиса.
3. Преобразовать `Conjunction` в типизированный фильтр бизнес-модуля.
4. Проверить операторы, типы, enum-значения и дубликаты.
5. Канонизировать фильтр перед вычислением page token.
6. Реализовать фильтрацию в repository до pagination.
7. Добавить тесты парсера, транслятора и storage adapters.
