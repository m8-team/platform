# M8 Platform Docs

Документация в `docs/` разделена на два слоя:

- ручной контент в `docs/services`, `docs/index.md`, `docs/toc.yaml`;
- автогенерируемый OpenAPI-контент в `docs/_generated/openapi`.

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

## Обновление OpenAPI Reference

1. Перегенерировать protobuf-артефакты и OpenAPI:

```bash
buf generate
```

2. Обновить сервисные разделы Diplodoc:

```bash
npm run sync:openapi
```

Команда `sync:openapi` изменяет только `docs/_generated/openapi`.

3. Собрать сайт:

```bash
npm run build
```

В публикацию попадают только OpenAPI-файлы с непустым `paths`, поэтому служебные protobuf-схемы без HTTP-методов не появляются в сгенерированном REST reference.
