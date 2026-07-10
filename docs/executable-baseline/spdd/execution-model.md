---
title: "Исполнимая модель SPDD"
---

# Исполнимая модель SPDD

[Executable Baseline](../index.md) | [SPDD Baseline](index.md)

Для каждого из 214 требований сгенерированы Feature и Review Prompt.
Для 171 функциональных требований созданы два Task Prompt:
проектирование контракта/домена и реализация/тестирование.

Порядок:
`Requirement → Feature Prompt → Task Prompt → Code/Test → Review Prompt → Release Evidence`.

Агент не может менять owner context, публичный контракт или Data Ownership без ADR.
