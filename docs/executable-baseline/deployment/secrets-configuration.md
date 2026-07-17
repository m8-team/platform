---
title: "Секреты и конфигурация"
---

# Секреты и конфигурация

[Executable Baseline](../index.md) | [Deployment Baseline](index.md)

- секреты передаются через внешний secret manager и short-lived credentials;
- ConfigMap не содержит ключи, пароли, токены и персональные данные;
- rotation не требует rebuild образа;
- приложения принимают ссылки на секреты, а не сериализованный secret material;
- доступ к provisioning credentials изолирован по driver и project scope;
- все чтения privileged secrets аудируются.
