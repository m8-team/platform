---
title: "YDB DDL"
---

# YDB DDL

[Executable Baseline](../../index.md) | [Data](../index.md)

{% note info %}

YDB DDL is kept as raw executable schema evidence grouped by bounded-context owner.

{% endnote %}

## Raw evidence

| Артефакт | Путь | Назначение |
| --- | --- | --- |
| m8-access | `m8-access/001_init.sql` | Access schema |
| m8-audit | `m8-audit/001_init.sql` | Audit schema |
| m8-authentication | `m8-authentication/001_init.sql` | Authentication schema |
| m8-identity | `m8-identity/001_init.sql` | Identity schema |
| m8-provisioning | `m8-provisioning/001_init.sql` | Provisioning schema |
| m8-resource-manager | `m8-resource-manager/001_init.sql` | Resource Manager schema |
| m8-risk-decision | `m8-risk-decision/001_init.sql` | Risk Decision schema |
| command owner | `command-owner/001_init.sql` | command ownership support |
| event consumer | `event-consumer/001_init.sql` | consumer inbox support |
| event publisher | `event-publisher/001_init.sql` | outbox support |
| operation owner service | `operation-owner-service/001_init.sql` | operation ownership support |
