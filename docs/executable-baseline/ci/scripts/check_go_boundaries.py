from pathlib import Path
import re
import sys

root = Path(sys.argv[1]).resolve()
errors = []
services = list((root / "services").glob("m8-*"))
for service in services:
    for path in service.rglob("*.go"):
        text = path.read_text(encoding="utf-8")
        if "/internal/domain" in str(path):
            if re.search(r'"[^"]+/(adapters|infrastructure|transport)/', text):
                errors.append(f"{path}: domain imports outer layer")
        for other in services:
            if other != service and f"/services/{other.name}/internal/" in text:
                errors.append(f"{path}: imports another service internal package")
if errors:
    print("\n".join(errors))
    raise SystemExit(1)
print("OK: Go dependency boundaries")
