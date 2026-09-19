# `ADR-<number>`: <Название решения>

| Поле | Значение |
| --- | --- |
| Идентификатор | `ADR-<number>` |
| Дата | `<YYYY-MM-DD: фактическая дата записи или решения>` |
| Статус | <span class="g-label g-label_theme_warning g-label_size_xs"><span class="g-label__text"><span class="g-label__content">Proposed</span></span></span> |
| Подтверждение согласования | — |

Новый ADR всегда начинается со статуса <span class="g-label g-label_theme_warning g-label_size_xs"><span class="g-label__text"><span class="g-label__content">Proposed</span></span></span>. Шаблон не утверждает решение. Используйте [правила ADR](../../adr/README.md); после переноса в `docs/adr/` пересчитайте ссылку.

Для единообразного отображения используйте следующие статусы:

| Статус | Когда использовать |
| --- | --- |
| <span class="g-label g-label_theme_warning g-label_size_xs"><span class="g-label__text"><span class="g-label__content">Proposed</span></span></span> | Решение предложено и ещё не согласовано. |
| <span class="g-label g-label_theme_success g-label_size_xs"><span class="g-label__text"><span class="g-label__content">Accepted</span></span></span> | Решение явно согласовано; в поле «Подтверждение согласования» укажите ссылку на фактическое подтверждение. |
| <span class="g-label g-label_theme_danger g-label_size_xs"><span class="g-label__text"><span class="g-label__content">Rejected</span></span></span> | Предложение рассмотрено и отклонено. |
| <span class="g-label g-label_theme_unknown g-label_size_xs"><span class="g-label__text"><span class="g-label__content">Deprecated</span></span></span> | Решение больше не рекомендуется к использованию, но не обязательно заменено другим ADR. |
| <span class="g-label g-label_theme_utility g-label_size_xs"><span class="g-label__text"><span class="g-label__content">Superseded</span></span></span> | Решение заменено новым ADR; укажите ссылку на заменяющее решение. |

## Контекст

<Проблема, требования, ограничения и причины, по которым нужно значимое архитектурное решение.>

## Рассмотренные варианты

| Вариант | Преимущества | Ограничения и последствия |
| --- | --- | --- |
| <Вариант> | <Аргументы> | <Компромиссы> |

<Перечислите действительно рассмотренные варианты, без вымышленных обсуждений.>

## Решение и обоснование

<Выбранный вариант и основания. Для Proposed обозначьте его как предложение; открытые вопросы укажите явно.>

## Последствия

<Положительные и отрицательные последствия, обязательства, риски; миграция или проверка решения, если применимы.>

## Связи

- Спецификации и обсуждения: `<ссылки>`.
- Связанные ADR — по ситуации: `<ссылки>`.
- Заменяет — по ситуации: `<ссылка на предыдущий ADR>`.
- Заменено — по ситуации: `<ссылка на новый ADR>`.

Существенный пересмотр Accepted оформляется новым ADR с сохранением истории предыдущего.
