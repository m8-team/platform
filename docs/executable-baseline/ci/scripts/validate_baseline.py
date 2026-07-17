from pathlib import Path
import json
import subprocess
import sys

root = Path(sys.argv[1]).resolve()
docs_root = next((p for p in [root, *root.parents] if (p / "node_modules/js-yaml").exists()), None)
js_yaml = docs_root / "node_modules/js-yaml" if docs_root else None

def yaml_load(path):
    if js_yaml is None:
        raise RuntimeError("js-yaml is not installed; run npm install in docs")
    code = """
const fs = require('fs');
const yaml = require(process.argv[1]);
const file = process.argv[2];
const docs = yaml.loadAll(fs.readFileSync(file, 'utf8'));
process.stdout.write(JSON.stringify(docs.length === 1 ? docs[0] : docs));
"""
    result = subprocess.run(["node", "-e", code, str(js_yaml), str(path)], text=True, capture_output=True)
    if result.returncode != 0:
        raise RuntimeError(result.stderr.strip() or result.stdout.strip())
    return json.loads(result.stdout)

def yaml_load_all(path):
    if js_yaml is None:
        raise RuntimeError("js-yaml is not installed; run npm install in docs")
    code = """
const fs = require('fs');
const yaml = require(process.argv[1]);
const file = process.argv[2];
yaml.loadAll(fs.readFileSync(file, 'utf8'));
"""
    result = subprocess.run(["node", "-e", code, str(js_yaml), str(path)], text=True, capture_output=True)
    if result.returncode != 0:
        raise RuntimeError(result.stderr.strip() or result.stdout.strip())

errors = []

for path in root.rglob("*.yaml"):
    try:
        yaml_load_all(path)
    except Exception as exc:
        errors.append(f"{path}: YAML: {exc}")

for path in root.rglob("*.json"):
    try:
        json.loads(path.read_text(encoding="utf-8"))
    except Exception as exc:
        errors.append(f"{path}: JSON: {exc}")

req = yaml_load(root / "product/approved_requirements.yaml")
ids = [item["id"] for item in req["requirements"]]
if len(ids) != len(set(ids)):
    errors.append("duplicate requirement IDs")
if any(item["approval_status"] != "APPROVED" for item in req["requirements"]):
    errors.append("not all requirements are approved")

api = yaml_load(root / "contracts/api_catalog.approved.yaml")
rpcs = [item["rpc"] for item in api["contracts"]]
if len(rpcs) != len(set(rpcs)):
    errors.append("duplicate RPC names")

events = yaml_load(root / "contracts/event_catalog.approved.yaml")
types = [item["event_type"] for item in events["events"]]
if len(types) != len(set(types)):
    errors.append("duplicate event types")

if errors:
    print("\n".join(errors))
    raise SystemExit(1)
print(f"OK: requirements={len(ids)}, rpcs={len(rpcs)}, events={len(types)}")
