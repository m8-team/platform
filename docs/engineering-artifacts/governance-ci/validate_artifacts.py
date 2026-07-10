#!/usr/bin/env python3
from pathlib import Path
import json
import subprocess
import sys

ROOT = Path(__file__).resolve().parents[1]
JS_YAML = ROOT.parent / "node_modules" / "js-yaml"
errors=[]; warnings=[]

def load(rel):
    script = """
const fs = require('fs');
const yaml = require(process.argv[2]);
const doc = yaml.load(fs.readFileSync(process.argv[1], 'utf8'));
process.stdout.write(JSON.stringify(doc ?? null));
"""
    try:
        result = subprocess.run(
            ["node", "-e", script, str(ROOT / rel), str(JS_YAML)],
            check=True,
            capture_output=True,
            text=True,
        )
        return json.loads(result.stdout or "null") or {}
    except Exception as e:
        errors.append(f"{rel}: {e}")
        return {}

api=load('contracts/api-catalog.yaml').get('contracts',[])
evt=load('contracts/event-catalog.yaml').get('events',[])
data=load('data/data-ownership.yaml').get('entities',[])
trace=load('traceability/traceability.yaml').get('records',[])
ids={x['id'] for x in api+evt+data if 'id' in x}
for coll,name in [(api,'API'),(evt,'Event'),(data,'Data'),(trace,'Trace')]:
    vals=[x.get('id') or x.get('requirement_id') for x in coll]
    dup={v for v in vals if vals.count(v)>1}
    if dup: errors.append(f"duplicate {name} IDs: {sorted(dup)}")
for r in trace:
    for field in ('api_contract_ids','event_contract_ids','data_contract_ids'):
        for ref in r.get(field,[]):
            if ref not in ids: errors.append(f"{r['requirement_id']}: unresolved {field} {ref}")
    if r.get('coverage_status')=='pilot-complete' and (not r.get('prompt_ids') or not r.get('test_ids')):
        errors.append(f"{r['requirement_id']}: pilot-complete without prompt/tests")
for p in ROOT.rglob('*.yaml'):
    load(p.relative_to(ROOT))
print(f"errors={len(errors)} warnings={len(warnings)}")
for e in errors: print('ERROR',e)
for w in warnings: print('WARN',w)
sys.exit(1 if errors else 0)
