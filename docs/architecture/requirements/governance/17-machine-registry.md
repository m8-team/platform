---
title: "Requirements Catalog: машинный реестр"
description: "Требования к машинному представлению каталога."
keywords:
  - "M8 Platform"
  - "requirements"
---

# 17. Требования к машинному реестру {#requirements-machine-registry}

{% note info "Навигация Requirements Catalog" %}

[Оглавление требований](../index.md) | [PADS](../../pads/index.md) | [Engineering artifacts](../../../engineering-artifacts/index.md) | [Предыдущий раздел: 16. SPDD backlog](16-spdd-backlog.md) | [Следующий раздел: Приложение A. Шаблон нового требования](../appendices/appendix-a-requirement-template.md)

{% endnote %}

Следующая редакция должна экспортировать каждое требование в YAML с JSON Schema validation. Минимальные автоматические проверки:

- уникальность ID;
- существование capability и owner context;
- наличие acceptance criteria;
- допустимость статуса и переходов;
- заполнение security/data/consistency impact;
- ссылки на существующие API/Event/Prompt/Test IDs;
- отсутствие requirement в release без verification evidence;
- coverage report по capabilities и bounded contexts.
