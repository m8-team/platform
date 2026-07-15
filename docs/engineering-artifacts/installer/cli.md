---
title: "m8ctl CLI Specification"
---

# CLI Specification

## Commands

| Command | MVP Status | Purpose |
| --- | --- | --- |
| `m8ctl preflight` | implemented | Validate config and cluster prerequisites. |
| `m8ctl plan` | implemented | Build deterministic installation plan. |
| `m8ctl bootstrap` | stub | Apply bootstrap layer directly. |
| `m8ctl install` | stub | Execute saved plan and hand off to Argo CD. |
| `m8ctl status` | stub | Show installation and operation status. |
| `m8ctl doctor` | stub | Functional diagnostics and optional diagnostic bundle. |
| `m8ctl upgrade plan` | stub | Plan version upgrade. |
| `m8ctl upgrade execute` | stub | Execute approved upgrade. |
| `m8ctl rollback` | stub | Roll back within safe boundary. |
| `m8ctl backup create/status/verify` | stub | Backup operations. |
| `m8ctl restore plan/execute` | stub | Restore operations. |
| `m8ctl bundle export/verify/import` | stub | Air-gapped bundle lifecycle. |
| `m8ctl uninstall` | stub | Stateless uninstall by default. |

## Common Flags

| Flag | Meaning |
| --- | --- |
| `--output table|json|yaml` | Human or machine-readable output. |
| `--quiet` | Future: suppress non-result output. |
| `--verbose` | Future: include details. |
| `--debug` | Future: structured debug logs without secrets. |
| `--non-interactive` | Future: CI mode, no prompts or ANSI. |
| `--yes` | Future: accept non-destructive prompts. |
| `--dry-run` | Future: plan-only executor mode. |

## `preflight`

```bash
m8ctl preflight -f installation.yaml --output table
m8ctl preflight -f installation.yaml --output json
m8ctl preflight -f installation.yaml --skip-cluster
```

Additional flags:

- `--kubeconfig`
- `--context`
- `--skip-cluster`

Exit codes:

- `0`: no failed checks;
- `2`: one or more failed checks;
- `3`: usage error;
- `1`: unexpected execution error.

## `plan`

```bash
m8ctl plan -f installation.yaml --output table
m8ctl plan -f installation.yaml --output json
m8ctl plan -f installation.yaml --output plan.yaml
```

Additional flags:

- `--catalog catalog/releases`
- `--allow-unsigned-release` for local development only.

Plan output includes:

- config digest;
- release catalog digest;
- sync waves;
- readiness gates;
- rollback boundaries;
- irreversible actions;
- risks.

## Exit Codes

| Code | Meaning |
| ---: | --- |
| 0 | Success |
| 1 | Execution error |
| 2 | Check failed |
| 3 | Usage error |
| 4 | Command defined but not implemented in MVP |

## Interactive Behavior

Interactive commands will show progress events and require confirmation for destructive actions. CI mode must be deterministic, non-ANSI and prompt-free. Destructive data deletion will require both installation name and a destructive action token.

## Security Rules

- Secrets are never accepted through command-line flags.
- Secret values must come from stdin or external references.
- Temporary files use restrictive permissions and are removed after use.
- Plan files are written with `0600` permissions.
- Output must not contain passwords, tokens, private keys or authorization headers.

