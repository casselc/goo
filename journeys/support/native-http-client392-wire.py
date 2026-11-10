"""Validate recorded 392 API bodies without synthesizing successful data."""
import json
import re
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]
sys.path.insert(0, str(ROOT / 'scripts'))
from check_openapi import check_observed

spec = json.loads((ROOT / 'api/openapi.json').read_text())
report_path = Path(sys.argv[1]) if len(sys.argv) > 1 else ROOT / 'artifacts/native-http-client392/summary.json'
report = json.loads(report_path.read_text())
observed = []
for call in report['calls']:
    if 'response' not in call:
        continue
    path, method = call['path'].removeprefix('/api'), call['method'].lower()
    matches = [p for p, ops in spec['paths'].items() if method in ops and re.fullmatch(re.sub(r'\{[^}]+\}', r'[^/]+', p), path)]
    if path in matches:
        matches = [path]
    assert len(matches) == 1, (path, method, matches)
    observed.append(dict(path=matches[0], method=method, status=call['status'], request=call.get('body'), response=call['response']))
for read in report['reads']:
    observed.append(dict(path='/measurements', method='post', status=read['status'], request=read['request'], response=read['response']))
for read in report['actor']['complete']['requests']:
    observed.append(dict(path=read['path'].removeprefix('/api'), method='post', status=read['status'], request=read['request'], response=read['response']))
assert len(observed) > 20, 'A complete recorded service workflow is required'
out = report_path.with_name('api-observations.json')
out.write_text(json.dumps(observed, indent=2))
check_observed(spec, out)
print(json.dumps({'responses': len(observed), 'requests': sum(case['request'] is not None for case in observed)}, sort_keys=True))
