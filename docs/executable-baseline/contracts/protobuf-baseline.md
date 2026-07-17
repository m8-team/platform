---
title: "Protobuf/ConnectRPC Baseline"
---

# Protobuf/ConnectRPC Baseline

[Executable Baseline](../index.md) | [Contracts Baseline](index.md)

Все RPC из утверждённого API Catalog представлены в `.proto`.
Контрактный каркас фиксирует имя RPC, request/response type, visibility,
requirement linkage и использование Common Operation.

`ContractRequest` и `ContractResponse` применяются только как baseline.
Перед реализацией Feature Prompt обязан заменить `google.protobuf.Struct`
на типизированные поля без изменения имени RPC и семантики операции.
Изменение существующего поля после публикации подчиняется `buf breaking`.
