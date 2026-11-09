"""Validate actual 388 wire observations with the existing OpenAPI checker."""
import json,re,sys
from pathlib import Path
ROOT=Path(__file__).resolve().parents[2]
sys.path.insert(0,str(ROOT/'scripts'))
from check_openapi import check_observed,validator
root=ROOT;spec=json.loads((root/'api/openapi.json').read_text());reportPath=Path(sys.argv[1]) if len(sys.argv)>1 else root/'artifacts/gauge-observation388/summary.json';report=json.loads(reportPath.read_text());observed=[]
for call in report['calls']:
    if 'response' not in call: continue
    path=call['path'].removeprefix('/api');method=call['method'].lower()
    candidates=[p for p,ops in spec['paths'].items() if method in ops and re.fullmatch(re.sub(r'\{[^}]+\}',r'[^/]+',p),path)]
    if path in candidates: candidates=[path]
    assert len(candidates)==1,(path,method,candidates)
    observed.append({'path':candidates[0],'method':method,'status':call['status'],'request':call.get('body'),'response':call['response']})
out=reportPath.with_name('api-observations.json');out.write_text(json.dumps(observed,indent=2));check_observed(spec,out)
print(json.dumps({'responses':len(observed),'requests':sum(c['request'] is not None for c in observed),'rejectedRequests':sum('rejectedRequest' in c for c in observed),'statuses':{str(s):sum(c['status']==s for c in observed) for s in sorted(set(c['status'] for c in observed))}},sort_keys=True))
