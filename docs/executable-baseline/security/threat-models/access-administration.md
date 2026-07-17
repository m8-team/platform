---
title: "Threat Model: Access Administration"
---

# Threat Model: Access Administration

[Executable Baseline](../../index.md) | [Threat Models](../index.md)

        | Threat | Scenario | Required controls |
        | --- | --- | --- |
        | Privilege escalation | Actor grants own admin role | separate grant permission, policy simulation and audit |
| Stale authorization | Revoked binding remains cached | revision tokens, bounded cache TTL and invalidation events |
| Model poisoning | Unauthorized model publication | four-eyes approval and signed model version |

        ## Verification

        Controls проверяются unit/integration/security tests и подтверждаются release evidence.
        Остаточный риск принимается только Security Owner через ADR/security exception.
