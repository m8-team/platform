#!/usr/bin/env python3
"""Validate policy Markdown links and unique rule IDs without third-party packages."""
from pathlib import Path
import re
import sys
from urllib.parse import unquote, urlparse

ROOT = Path(__file__).resolve().parents[1]
DOCS = sorted((ROOT / "docs" / "engineering").rglob("*.md"))
AGENTS = [ROOT / name for name in (
    "AGENTS.md", "CLAUDE.md", "internal/AGENTS.md", "proto/AGENTS.md",
    "ui/console/AGENTS.md",
)]
LINK = re.compile(r"(?<!!)\[[^\]\n]+\]\(([^)]+)\)")
RULE = re.compile(r"^###\s+((?:ARCH|SEC|GO|API|DATA|TEST|AI|FE|UX)-\d{3})\b", re.M)


def validate():
    failures = []
    if not DOCS:
        failures.append("docs/engineering is missing")
    for path in AGENTS:
        if not path.is_file():
            failures.append("missing agent instructions: " + str(path.relative_to(ROOT)))
    seen = {}
    for path in DOCS + [path for path in AGENTS if path.is_file()]:
        contents = path.read_text(encoding="utf-8")
        for match in LINK.finditer(contents):
            raw = match.group(1).strip().split(" ", 1)[0].strip("<>")
            if raw.startswith("#") or raw.startswith("//") or urlparse(raw).scheme:
                continue
            target = unquote(raw.split("#", 1)[0].split("?", 1)[0])
            if target and not (path.parent / target).resolve().is_file():
                failures.append(f"{path.relative_to(ROOT)}: broken link {raw}")
        if path in DOCS:
            for rule in RULE.findall(contents):
                if rule in seen:
                    failures.append(f"duplicate normative rule {rule}: {path} and {seen[rule]}")
                seen[rule] = path
    index = ROOT / "docs" / "engineering" / "README.md"
    if index.is_file() and "PROPOSED; not merged" in index.read_text(encoding="utf-8"):
        failures.append("outdated adoption status: PROPOSED; not merged")
    if len(seen) < 70:
        failures.append(f"only {len(seen)} distinct normative rule IDs (expected >= 70)")
    if failures:
        print("\n".join(f"ERROR: {message}" for message in failures), file=sys.stderr)
        return 1
    print(f"Validated {len(DOCS) + len(AGENTS)} policy files and {len(seen)} rule IDs")
    return 0


if __name__ == "__main__":
    sys.exit(validate())
