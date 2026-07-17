---
title: "Disaster Recovery Baseline"
---

# Disaster Recovery Baseline

[Executable Baseline](../index.md) | [Deployment Baseline](index.md)

| Компонент | RPO MVP | RTO MVP | Метод |
| --- | ---: | ---: | --- |
| YDB | 15 мин | 4 ч | snapshot + transaction log |
| Kafka | 15 мин | 4 ч | replicated topics + archive |
| Temporal | 15 мин | 4 ч | DB backup + worker redeploy |
| Keycloak | 15 мин | 4 ч | DB backup + realm export |
| SpiceDB | 15 мин | 4 ч | datastore backup + model registry |
| Audit | 0–15 мин | 4 ч | replicated append store + integrity validation |

DR exercise проводится ежеквартально. Успех подтверждается release evidence.
