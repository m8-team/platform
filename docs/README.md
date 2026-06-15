# M8 Platform Docs

Документация в `docs/` разделена на два слоя:

- ручной контент в `docs/services`, `docs/index.md`, `docs/toc.yaml`;
- локальные developer guides, например `docs/local-clickstack.md`;
- автогенерируемый OpenAPI-контент в `docs/_generated/openapi`;
- автогенерируемый gRPC reference в `docs/services/*/api-reference/grpc`.

## Команды

```bash
npm install
npm run sync:api
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
npm run sync:api
```

Команда `sync:api` запускает:

- `sync:openapi`, который обновляет REST reference из OpenAPI-файлов;
- `sync:grpc`, который строит protobuf descriptor через `buf build` и генерирует gRPC reference из комментариев `.proto`.

3. Собрать сайт:

```bash
npm run build
```

В публикацию попадают только OpenAPI-файлы с непустым `paths`, поэтому служебные protobuf-схемы без HTTP-методов не появляются в сгенерированном REST reference.
