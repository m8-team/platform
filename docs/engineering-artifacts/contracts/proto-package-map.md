---
title: "Proto Package Map"
description: "Карта protobuf packages, сервисов и Go package namespaces."
keywords:
  - "M8 Platform"
  - "engineering artifacts"
---

# Proto Package Map {#proto-package-map}

{% note info "Навигация" %}

[Engineering artifacts](../index.md) | [Contracts](index.md) | `proto-package-map.yaml`

{% endnote %}

Карта protobuf packages, сервисов и Go package namespaces.

| Context | Package | Service | Go package |
| --- | --- | --- | --- |
| Resource Manager | m8.resource_manager.v1 | ResourceManagerService | github.com/m8-platform/api/gen/go/resource_manager/v1 |
| Identity | m8.identity.v1 | IdentityService | github.com/m8-platform/api/gen/go/identity/v1 |
| Authentication | m8.authentication.v1 | AuthenticationService | github.com/m8-platform/api/gen/go/authentication/v1 |
| Access | m8.access.v1 | AccessService | github.com/m8-platform/api/gen/go/access/v1 |
| Risk Decision | m8.risk_decision.v1 | RiskDecisionService | github.com/m8-platform/api/gen/go/risk_decision/v1 |
| Provisioning | m8.provisioning.v1 | ProvisioningService | github.com/m8-platform/api/gen/go/provisioning/v1 |
| Audit | m8.audit.v1 | AuditService | github.com/m8-platform/api/gen/go/audit/v1 |
| Common Operation | m8.operations.v1 | OperationsService | github.com/m8-platform/api/gen/go/operations/v1 |

Common packages: m8.common.v1, google.longrunning, google.rpc, buf.validate.

## Источник

Машинный источник хранится в файле `proto-package-map.yaml`.
