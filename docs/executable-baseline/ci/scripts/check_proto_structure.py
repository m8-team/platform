from pathlib import Path
import re
import sys

root = Path(sys.argv[1])
errors = []
for path in root.rglob("*.proto"):
    text = path.read_text(encoding="utf-8")
    if text.count("{") != text.count("}"):
        errors.append(f"{path}: unbalanced braces")
    if not re.search(r'syntax\s*=\s*"proto3"', text):
        errors.append(f"{path}: missing proto3 syntax")
    rpc_names = re.findall(r"\brpc\s+(\w+)\s*\(", text)
    if len(rpc_names) != len(set(rpc_names)):
        errors.append(f"{path}: duplicate RPC")
if errors:
    print("\n".join(errors))
    raise SystemExit(1)
print("OK: proto structural checks")
