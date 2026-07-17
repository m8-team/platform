---
title: "Threat Model: Step Up"
---

# Threat Model: Step Up

[Executable Baseline](../../index.md) | [Threat Models](../index.md)

        | Threat | Scenario | Required controls |
        | --- | --- | --- |
        | AAL downgrade | Client requests lower assurance | server computes required AAL and rejects downgrade |
| Session confusion | Step-up applied to another session | bind transaction to subject, client and session reference |
| Replay | Old challenge reused | one-time challenge, nonce and expiry |

        ## Verification

        Controls проверяются unit/integration/security tests и подтверждаются release evidence.
        Остаточный риск принимается только Security Owner через ADR/security exception.
