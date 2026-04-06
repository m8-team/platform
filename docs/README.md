# M8 Platform Docs

Документация в `docs/` собирается на базе Diplodoc и автоматически подтягивает OpenAPI-спецификации из `../api/generate/openapi`.

## Команды

```bash
npm install
npm run sync:openapi
npm run build
```

Для локальной разработки в watch-режиме:

```bash
npm run dev
```

## Обновление API Reference

1. Перегенерировать protobuf-артефакты и OpenAPI:

```bash
buf generate
```

2. Обновить сервисные разделы Diplodoc:

```bash
npm run sync:openapi
```

3. Собрать сайт:

```bash
npm run build
```

В публикацию попадают только OpenAPI-файлы с непустым `paths`, поэтому служебные protobuf-схемы без HTTP-методов в API Reference не появляются.
