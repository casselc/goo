"""Check recorded 391 API bodies against OpenAPI; never manufacture data."""
import json
import re
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]
sys.path.insert(0, str(ROOT / 'scripts'))
from check_openapi import check_observed

spec = json.loads((ROOT / 'api/openapi.json').read_text())
report_path = Path(sys.argv[1]) if len(sys.argv) > 1 else ROOT / 'artifacts/log-pattern-display391/summary.json'
report = json.loads(report_path.read_text())
observed = []
for call in report['calls']:
    if 'response' not in call:
        continue
    path = call['path'].removeprefix('/api')
    method = call['method'].lower()
    candidates = [p for p, ops in spec['paths'].items() if method in ops and re.fullmatch(re.sub(r'\{[^}]+\}', r'[^/]+', p), path)]
    if path in candidates:
        candidates = [path]
    assert len(candidates) == 1, (path, method, candidates)
    observed.append(dict(path=candidates[0], method=method, status=call['status'], request=call.get('body'), response=call['response']))
for case in report['definitions']:
    if case['name'] == 'destination-setup':
        observed.append(dict(path='/exploration-views', method='post', status=case['status'], request=case['request'], response=case['response']))
assert len(observed) >= 10, 'A complete recorded journey is required'
out = report_path.with_name('api-observations.json')
out.write_text(json.dumps(observed, indent=2))
check_observed(spec, out)
print(json.dumps({'responses': len(observed), 'requests': sum(c['request'] is not None for c in observed)}, sort_keys=True))
