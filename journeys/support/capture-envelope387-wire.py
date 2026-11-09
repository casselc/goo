"""Admit genuine 387 wire traffic; reject explicitly corrupted negative transports."""
import json,re,sys
from pathlib import Path
ROOT=Path(__file__).resolve().parents[2]
sys.path.insert(0,str(ROOT/'scripts'))
from check_openapi import check_observed,validator
report_path=Path(sys.argv[1]) if len(sys.argv)>1 else ROOT/'artifacts/capture-envelope387/summary.json'
report=json.loads(report_path.read_text());spec=json.loads((ROOT/'api/openapi.json').read_text());observations=[];negative=[]
def route(path,method):
    path=path.removeprefix('/api');choices=[p for p,ops in spec['paths'].items() if method in ops and re.fullmatch(re.sub(r'\{[^}]+\}',r'[^/]+',p),path)]
    if path in choices: choices=[path]
    assert len(choices)==1,(path,method,choices)
    return choices[0]
for call in report['calls']:
    if 'response' not in call: continue
    method=call['method'].lower();path=route(call['path'],method)
    item=dict(path=path,method=method,status=call['status'],request=call.get('body'),response=call['response'])
    if call.get('negativeTransport'):
        assert path=='/captures/{id}' and item['status']==200
        schema=spec['paths'][path][method]['responses']['200']['content']['application/json']['schema']
        if call['negativeTransport']['mode']=='wrong-valid-page':
            assert validator(spec,schema).is_valid(item['response']), 'Wrong-page probe must retain a valid envelope'
        else:
            assert not validator(spec,schema).is_valid(item['response']), 'Corrupted transport admitted'
        item['probe']=call['negativeTransport'];negative.append(item)
    else: observations.append(item)
for call in report['setupCalls']:
    if not call['path'].startswith('/api/') or 'response' not in call: continue
    method=call['method'].lower();observations.append(dict(path=route(call['path'],method),method=method,status=call['status'],request=call.get('request'),response=call['response']))
for mutant in report['fileMutants']:
    observations.append(dict(path='/captures/{id}',method='get',status=mutant['status'],request=None,response=mutant['response']))
probe=report['sourceRefusalProbe'];observations.append(dict(path='/explore',method='post',status=probe['status'],request=probe['request'],response=probe['response']))
assert len(negative)==19
assert sum(x['probe']['mode']=='wrong-valid-page' for x in negative)==1
out=report_path.with_name('api-observations.json');out.write_text(json.dumps(observations,indent=2));report_path.with_name('invalid-transport-observations.json').write_text(json.dumps(negative,indent=2));check_observed(spec,out)
print(json.dumps(dict(responses=len(observations),requests=sum(x['request'] is not None for x in observations),invalidTransportResponses=sum(x['probe']['mode']!='wrong-valid-page' for x in negative),wrongRequestedPageResponses=sum(x['probe']['mode']=='wrong-valid-page' for x in negative),statuses={str(status):sum(x['status']==status for x in observations) for status in sorted({x['status'] for x in observations})}),sort_keys=True))
