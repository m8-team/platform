---
title: "Threat Model: Ciba Authentication"
---

# Threat Model: Ciba Authentication

[Executable Baseline](../../index.md) | [Threat Models](../index.md)

        | Threat | Scenario | Required controls |
        | --- | --- | --- |
        | Push approval phishing | Attacker induces approval | transaction binding, clear user intent, device/risk signals, expiry |
| Callback spoofing | Forged provider callback | mTLS/signature, nonce, audience and replay protection |
| Transaction enumeration | Guess authentication ID | unguessable IDs, permission checks, response minimization |
| OTP fallback abuse | Fallback bypasses stronger method | policy-controlled fallback and risk re-evaluation |

        ## Verification

        Controls проверяются unit/integration/security tests и подтверждаются release evidence.
        Остаточный риск принимается только Security Owner через ADR/security exception.
