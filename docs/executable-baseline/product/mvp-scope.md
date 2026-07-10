---
title: "M8 Platform MVP Scope v1.0"
---

# M8 Platform MVP Scope v1.0

[Executable Baseline](../index.md) | [Product Baseline](index.md)

## Решение

MVP состоит из минимального control plane, способного создать ресурсную область,
управлять субъектами, проверить доступ, провести CIBA/step-up аутентификацию,
зафиксировать аудит и представить длительную операцию.

## MVP-1

- Resource Manager: Organization → Workspace → Project, чтение и метки.
- Identity: User Pool, User, External Identity, Group и Membership.
- Access: CheckPermission, Role/Binding, Relationship, публикация модели SpiceDB.
- Authentication: Start/Get/Wait/Cancel, CIBA, OTP, handoff, reauthentication и step-up.
- Audit: ingestion, provenance, minimization, search и integrity verification.
- Common Operation: Get/List/Wait/Cancel, progress, completion и failure.
- Platform: idempotency, audit, tracing, SLO, Outbox/Inbox и contract compatibility.

Количество требований MVP-1: **83**.

## Не входит в MVP-1

- автоматическое provisioning внешней инфраструктуры;
- расширенные risk policies и manual review;
- multi-region active/active;
- полная UI-панель администратора;
- Terraform provider.

Эти функции утверждены, но поставляются последующими волнами.
