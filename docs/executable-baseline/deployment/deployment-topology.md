---
title: "Топология развертывания"
---

# Топология развертывания

[Executable Baseline](../index.md) | [Deployment Baseline](index.md)

## MVP

- один primary Kubernetes cluster и отдельный recovery cluster;
- отдельные namespaces: `m8-system`, `m8-iam`, `m8-control-plane`, `m8-observability`;
- YDB и Kafka эксплуатируются как независимые stateful-платформы;
- Temporal, Keycloak и SpiceDB не встраиваются внутрь доменных сервисов;
- service-to-service доступ использует mTLS и отдельные service identities;
- ingress открыт только к API gateway/BFF; внутренние сервисы не публикуются наружу.

## Эволюция

Active/passive → cell-based multi-region. Переход допускается после измерения
объёма межрегионального состояния, RPO/RTO и задержки permission/authentication paths.
