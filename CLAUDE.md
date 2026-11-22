# M8 Platform — Claude Code

@AGENTS.md

The imported `AGENTS.md` is the canonical shared agent policy. This file exists for Claude Code compatibility and Claude-specific workflow hints; do not duplicate engineering standards here.

## Claude-specific instructions

- At the start of a task, identify which engineering documents and nested `AGENTS.md` files apply to the changed paths; read them before implementing.
- For multi-module or high-risk work, use a planning workflow and describe expected checks before changes. Do not take irreversible actions without authorization.
- Before finishing, inspect the final diff and give the verification/evidence report required by the shared policy.
- Do not import all engineering files unconditionally into the session context; read only the ones relevant to the task.

If Claude Code is configured to load both `CLAUDE.md` and `AGENTS.md`, the explicit import remains the authoritative common entrypoint.
