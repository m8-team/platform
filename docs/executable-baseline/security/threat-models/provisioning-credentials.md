---
title: "Threat Model: Provisioning Credentials"
---

# Threat Model: Provisioning Credentials

[Executable Baseline](../../index.md) | [Threat Models](../index.md)

        | Threat | Scenario | Required controls |
        | --- | --- | --- |
        | Credential exfiltration | Driver secret leaked | secret references, short-lived credentials and isolation |
| Confused deputy | Driver acts in wrong project | scope binding and explicit placement policy |
| Malicious desired state | Payload causes unsafe resource | schema validation, policy gate and allowlisted drivers |

        ## Verification

        Controls проверяются unit/integration/security tests и подтверждаются release evidence.
        Остаточный риск принимается только Security Owner через ADR/security exception.
