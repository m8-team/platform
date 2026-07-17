---
title: "Threat Model: Audit Integrity"
---

# Threat Model: Audit Integrity

[Executable Baseline](../../index.md) | [Threat Models](../index.md)

        | Threat | Scenario | Required controls |
        | --- | --- | --- |
        | Audit deletion | Privileged actor removes evidence | append-only storage, separate duty and integrity chain |
| Sensitive data leakage | Secrets enter audit payload | minimization validation and schema allowlist |
| False provenance | Producer identity forged | service identity, signed channel and provenance validation |

        ## Verification

        Controls проверяются unit/integration/security tests и подтверждаются release evidence.
        Остаточный риск принимается только Security Owner через ADR/security exception.
