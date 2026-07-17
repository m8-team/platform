---
title: "Chaos и DR plan"
---

# Chaos и DR plan

[Executable Baseline](../index.md) | [Testing Baseline](index.md)

Сценарии: недоступность Risk Decision, задержка SpiceDB, остановка Kafka broker,
повторная доставка события, падение Temporal worker, read-only YDB,
потеря primary cluster и восстановление backup.

Успех: инварианты не нарушены, mutation не потеряна, duplicate не создаёт
второй эффект, audit chain восстанавливается, SLO breach детектируется.
