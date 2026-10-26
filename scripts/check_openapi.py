"""Validate OpenAPI, route coverage and frozen direct-Go wire fixtures.
No source/database calls. Request schemas cover decoding shape; Go still owns
semantic validation, byte budgets, capability lifetimes and effects.
"""
import ast
import argparse
import copy
import json
import re
from pathlib import Path

from jsonschema import Draft202012Validator
from openapi_spec_validator import validate_spec

ROOT = Path(__file__).resolve().parents[1]


def validator(spec, schema):
    # Preserve document-local references without remote resolution/network access.
    return Draft202012Validator({'allOf': [schema], 'components': spec['components']})


def check(spec):
    validate_spec(spec)
    routes = set()
    for file in (ROOT / 'cmd/goo').glob('*http.go'):
        routes.update((m.lower(), p) for m, p in re.findall(r'HandleFunc\("(GET|POST) ([^" ]+)"', file.read_text()))
    routes.add(('post', '/queries/{id}/cancel'))
    documented = {(m, p) for p, item in spec['paths'].items() for m in item if m in ('get', 'post') and 'x-go-source' in item[m]}
    assert routes == documented, f'Go route coverage changed: missing={routes-documented}, stale={documented-routes}'
    # Check the Python-only public API explicitly; proxy-forwarded routes are Go's.
    python_tree = ast.parse((ROOT / 'server.py').read_text())
    literals = {n.value for n in ast.walk(python_tree) if isinstance(n, ast.Constant) and isinstance(n.value, str)}
    for path in ('/package', '/example-questions', '/prebuilts', '/query', '/explain', '/ui-config'):
        assert '/api' + path in literals and path in spec['paths']
    for path, item in spec['paths'].items():
        for method, op in item.items():
            if 'x-go-source' not in op:
                continue
            assert op['x-schema-coverage'] != 'response-envelope-pending', (method, path)
            if method == 'post' and path != '/queries/{id}/cancel':
                assert 'requestBody' in op, path
    for frame in (ROOT / 'testdata/http/tail.sse').read_text().strip().split('\n\n'):
        fields = dict(line.split(': ', 1) for line in frame.splitlines())
        assert fields['id'].isdigit()
        shape = spec['paths']['/tail']['post']['x-sse-events'][fields['event']]
        validator(spec, shape).validate(json.loads(fields['data']))
    responses = requests = mux_refusals = 0
    accepted = []
    for fixture in sorted((ROOT / 'testdata/http').glob('*.json')):
        data = json.loads(fixture.read_text())
        if not isinstance(data, dict) or 'cases' not in data:
            continue
        for case in data['cases']:
            path, method = case['path'], case['method'].lower()
            assert path in spec['paths'], (fixture.name, path)
            if method not in spec['paths'][path]:
                assert case['status'] == 405, (fixture.name, case['name'])
                assert case['headers']['Content-Type'].startswith('text/plain')
                mux_refusals += 1
                continue
            op = spec['paths'][path][method]
            status = str(case['status'])
            assert status in op['responses'], (fixture.name, case['name'], status)
            media = case['headers']['Content-Type'].split(';')[0]
            content = op['responses'][status]['content']
            assert media in content, (fixture.name, case['name'], media)
            payload = json.loads(case['response']) if media == 'application/json' else case['response']
            validator(spec, content[media]['schema']).validate(payload)
            responses += 1
            if case['status'] < 300 or case['status'] == 502:
                if 'requestBody' not in op:
                    continue
                body = case['body'] + 'x' * case.get('fillBytes', 0) + case.get('suffix', '')
                body += ' ' * max(0, case.get('padTo', 0) - len(body.encode()))
                request = json.loads(body)
                shape = op['requestBody']['content']['application/json']['schema']
                validator(spec, shape).validate(request)
                if 'x-go-body-limit-bytes' in op:
                    assert len(body.encode()) <= op['x-go-body-limit-bytes']
                requests += 1
                if isinstance(request, dict):
                    accepted.append((shape, request))
    assert responses == 368 and requests == 131 and mux_refusals == 22, (responses, requests)
    # Rejecting controls: a no-op validator or open input schemas must fail here.
    for shape, request in accepted:
        wrong = copy.deepcopy(request)
        wrong['__unknown_contract_control__'] = True
        assert list(validator(spec, shape).iter_errors(wrong)), 'Unknown-field control accepted'
    assert list(validator(spec, {'$ref':'#/components/schemas/Error'}).iter_errors({'error':17})), 'Error type control accepted'
    for event, shape in spec['paths']['/tail']['post']['x-sse-events'].items():
        value = {'reset': {'connection':'a'*64,'windowMs':60000,'selector':{'environment':'prod'},'continuity':'unknown','replay':False},
                 'batch': {'membershipState':'empty','rows':[],'limited':False,'fromMs':1000,'toMs':61000,'observedAt':'2026-10-22T00:00:00Z'},
                 'gap': {'reason':'source_unavailable'}}[event]
        validator(spec, shape).validate(value)
        wrong = copy.deepcopy(value)
        wrong[{'reset':'replay','batch':'rows','gap':'reason'}[event]] = 123
        assert list(validator(spec, shape).iter_errors(wrong)), f'SSE {event} control accepted'
    counted_shape={'$ref':'#/components/schemas/CountedValuesResult'}
    counted_cases=json.loads((ROOT/'testdata/http/counted-values.json').read_text())['cases']
    counted=next(json.loads(c['response']) for c in counted_cases if c['name']=='success-observed')
    for wrong in ({**counted,'count':2},{**counted,'count':'02'},{**counted,'state':'healthy'},{**counted,'basis':'not-a-content-basis'},{**counted,'groups':[{'g0':17,'count':'2'}]},{**counted,'groups':[{'g0':None,'count':'0'}]}):
        assert list(validator(spec,counted_shape).iter_errors(wrong)), 'Counted values negative control accepted'
    aggregate_shape={'$ref':'#/components/schemas/RecordAggregatesResult'}
    aggregate_cases=json.loads((ROOT/'testdata/http/record-aggregates.json').read_text())['cases']
    aggregate=next(json.loads(c['response']) for c in aggregate_cases if c['name']=='success-rich-series')
    for wrong in ({**aggregate,'count':2},{**aggregate,'state':'healthy'},{**aggregate,'rows':[]},{**aggregate,'resultToken':'a'*64},{**aggregate,'groups':[{'g0':17,'count':'2'}]},{**aggregate,'groups':[{'g0':None,'count':'0'}]}):
        assert list(validator(spec,aggregate_shape).iter_errors(wrong)), 'Record aggregate negative control accepted'
    pool_view_cases=json.loads((ROOT/'testdata/http/pool-view.json').read_text())['cases']
    pool_view=next(json.loads(c['response']) for c in pool_view_cases if c['name']=='validated-count-capacity-wait')
    pool_view_shape={'$ref':'#/components/schemas/PoolView'}
    for change in ({'state':'healthy'},{'state':'observed','reason':'fake'},{'state':'observed','reason':'fake','numerator':'1','denominator':'0'},{'state':'unavailable','reason':'unknown','numerator':'0','denominator':'1'}):
        wrong=copy.deepcopy(pool_view);wrong['utilization']=change
        assert list(validator(spec,pool_view_shape).iter_errors(wrong)), 'Pool-view utilization control accepted'
    assert pool_view['utilization']['state']=='observed' and pool_view['utilization']['numerator']=='1' and pool_view['utilization']['denominator']=='1'
    host_cases=json.loads((ROOT/'testdata/http/host.json').read_text())['cases']
    host_plan=next(json.loads(c['response']) for c in host_cases if c['name']=='host-plan')
    host_plan_shape={'$ref':'#/components/schemas/HostDashboardPlan'}
    for mode in ('manufactured-service','missing-resource','too-many-defaults','healthy'):
        wrong=copy.deepcopy(host_plan)
        if mode=='manufactured-service':wrong['widgets'][0]['question']['definition']['source']['services']=['fake-service']
        if mode=='missing-resource':del wrong['widgets'][0]['question']['definition']['source']['resource']
        if mode=='too-many-defaults':wrong['widgets']=wrong['widgets']*2
        if mode=='healthy':wrong['state']='healthy'
        assert list(validator(spec,host_plan_shape).iter_errors(wrong)), 'Host question contract control accepted'
    runtime_cases=json.loads((ROOT/'testdata/http/runtime-dashboard.json').read_text())['cases']
    runtime_plan=next(json.loads(c['response']) for c in runtime_cases if c['name']=='runtime-preview-omissions')
    assert runtime_plan['runtimeStreamsOmitted']==3
    for field in ('nativeOptionsOmitted','runtimeStreamsOmitted'):
        wrong=copy.deepcopy(runtime_plan);wrong[field]=-1
        assert list(validator(spec,{'$ref':'#/components/schemas/DashboardPlan'}).iter_errors(wrong)), 'Runtime omission control accepted'
    # Startup metadata is configuration only, with closed states and no health flags.
    runtime_shape={'$ref':'#/components/schemas/RuntimeCapabilities'}
    configured={'version':'goo/runtime-capabilities/v1','profile':'generic','preparedHTTP':{'state':'unconfigured'},'declaredOperations':{'state':'configured'}}
    validator(spec,runtime_shape).validate(configured)
    for wrong in ({**configured,'profile':'production'}, {**configured,'preparedHTTP':{'state':'healthy'}}, {**configured,'sourceHealthy':True}):
        assert list(validator(spec,runtime_shape).iter_errors(wrong)), 'Runtime metadata negative control accepted'
    refusal={'$ref':'#/components/schemas/RuntimeCapabilityRefusal'}
    for wrong in ({'error':'unconfigured','code':'capability_unconfigured','capability':'unknown'}, {'error':'unconfigured','capability':'prepared-http'}):
        assert list(validator(spec,refusal).iter_errors(wrong)), 'Capability refusal negative control accepted'
    # A preservation claim must carry a capability; unavailable/empty must not.
    shape=spec['paths']['/tail']['post']['x-sse-events']['batch']
    batch={'rows':[],'limited':False,'fromMs':1000,'toMs':61000,'observedAt':'2026-10-23T00:00:00Z','membershipState':'available'}
    assert list(validator(spec,shape).iter_errors(batch)), 'Unissued membership control accepted'
    batch['membership']={'token':'a'*64,'expiresAt':'2026-10-23T00:02:00Z','basis':'observed-multiplicity'}
    validator(spec,shape).validate(batch)
    for state in ('unavailable','empty'):
        wrong={**batch,'membershipState':state}
        assert list(validator(spec,shape).iter_errors(wrong)), 'False membership control accepted'
    # Origin-aware content must never accept ambiguous or missing manifests.
    shape={'$ref':'#/components/schemas/EvidenceContent'}
    for wrong in ({'origin':'tail-selection','manifestJSON':'{}'},
                  {'origin':'unknown','capture':{},'manifestJSON':'{}'},
                  {'origin':'tail-selection','capture':{},'tail':{},'retainedAt':'2026-10-23T00:00:00Z','manifestJSON':'{}'}):
        assert list(validator(spec,shape).iter_errors(wrong)), 'Origin union control accepted'
    print(f'OpenAPI valid: {len(routes)} Go operations; {responses} fixture responses, {requests} admitted requests, {mux_refusals} mux refusals; negative controls and SSE schemas pass')


def check_observed(spec, path):
    observations = json.loads(Path(path).read_text())
    assert len(observations) >= 12, 'Saved workflow observations missing'
    for case in observations:
        op = spec['paths'][case['path']][case['method'].lower()]
        status = str(case['status'])
        assert status in op['responses'], (case['path'], status)
        validator(spec, op['responses'][status]['content']['application/json']['schema']).validate(case['response'])
        if case.get('request') is not None and 'requestBody' in op:
            validator(spec, op['requestBody']['content']['application/json']['schema']).validate(case['request'])
    contents=[c['response'] for c in observations if c['path']=='/retained-evidence/{id}/content' and c['status']==200]
    capture=next((c for c in contents if c['origin']!='tail-selection'),None)
    tail=next((c for c in contents if c['origin']=='tail-selection'),None)
    if capture and tail:
        shape=spec['paths']['/retained-evidence/{id}/content']['get']['responses']['200']['content']['application/json']['schema']
        ambiguous={**tail,'capture':capture['capture']}
        assert list(validator(spec,shape).iter_errors(ambiguous)), 'Complete mixed-origin control accepted'
        mislabeled={**tail,'origin':'saved-capture'}
        assert list(validator(spec,shape).iter_errors(mislabeled)), 'Relabeled tail control accepted'
    print(f"Observed API workflow: {len(observations)} real handler responses conform (including optional XTDB only when configured)")


def check_stock_observed(spec,path):
    observations=json.loads(Path(path).read_text())
    assert len(observations)==6, 'Expected execute/capture/reopen/delta/monotonic/invalid stock workflow'
    for case in observations:
        op=spec['paths'][case['path']][case['method'].lower()]
        validator(spec,op['responses'][str(case['status'])]['content']['application/json']['schema']).validate(case['response'])
        if case.get('request') is not None and 'requestBody' in op:
            validator(spec,op['requestBody']['content']['application/json']['schema']).validate(case['request'])
    executed=next(c['response'] for c in observations if c['path']=='/measurements' and c['status']==200)
    assert executed['question']['definition']['kind']=='stock' and len(executed['stocks']['series'])==1
    assert all(k not in executed for k in ('gauges','counters','histograms','lifecycle'))
    sample=executed['stocks']['series'][0]['samples'][1]
    shape={'$ref':'#/components/schemas/metrics.StockSample'}
    for wrong in ({**sample,'state':'healthy'},{**sample,'value':'7.5'},{**sample,'state':'unknown'},{k:v for k,v in sample.items() if k!='value'}):
        assert list(validator(spec,shape).iter_errors(wrong)), 'Stock sample negative control accepted'
    assert len([c for c in observations if c['path']=='/measurements' and c['status']==502])==2
    captured=next(c['response']['measurement'] for c in observations if c['path']=='/captures' and c['status']==201)
    reopened=next(c['response']['measurement'] for c in observations if c['path']=='/captures/{id}' and c['status']==200)
    assert captured==reopened, 'Stock capture changed on offline reopen'
    print('Stock API workflow: 6 actual handler responses conform; capture/reopen and rejecting controls pass')


def check_stock_domain_observed(spec,path):
    observations=json.loads(Path(path).read_text())
    assert len(observations)==10, 'Expected domain execute/capture/reopen/refusals/boundary/missing workflow'
    for case in observations:
        op=spec['paths'][case['path']][case['method'].lower()]
        validator(spec,op['responses'][str(case['status'])]['content']['application/json']['schema']).validate(case['response'])
        if case.get('request') is not None and 'requestBody' in op:
            validator(spec,op['requestBody']['content']['application/json']['schema']).validate(case['request'])
    executed=next(c['response'] for c in observations if c['path']=='/measurements' and c['status']==200)
    assert executed['question']['definition']['stockDomain']=='nonnegative-integer'
    for value in (-1,7.5,9007199254740992):
        wrong=copy.deepcopy(executed);wrong['stocks']['series'][0]['samples'][1]['value']=value
        assert list(validator(spec,{'$ref':'#/components/schemas/NativeResult'}).iter_errors(wrong)), 'Stock domain value control accepted'
    assert len([c for c in observations if c['status']==502])==3
    captured=next(c['response']['measurement'] for c in observations if c['path']=='/captures' and c['status']==201)
    reopened=next(c['response']['measurement'] for c in observations if c['path']=='/captures/{id}' and c['status']==200)
    assert captured==reopened
    print('Stock domain API workflow: 10 actual handler responses conform; domain controls and offline reopen pass')


def check_pool_signals_observed(spec,path):
    observations=json.loads(Path(path).read_text())
    assert len(observations)==9
    for case in observations:
        op=spec['paths'][case['path']][case['method'].lower()]
        validator(spec,op['responses'][str(case['status'])]['content']['application/json']['schema']).validate(case['response'])
        if case.get('request') is not None and 'requestBody' in op:
            validator(spec,op['requestBody']['content']['application/json']['schema']).validate(case['request'])
    plan=observations[0]['response'];assert len(plan['capacity'])==1 and len(plan['wait'])==1
    for field in ('capacity','wait'):
        question=plan[field][0]['question'];assert question['fromMs']==plan['parent']['fromMs'] and question['toMs']==plan['parent']['toMs']
    assert plan['capacity'][0]['question']['definition']['stockDomain']=='nonnegative-integer'
    shape={'$ref':'#/components/schemas/PoolSignalsPlan'}
    for wrong in ({**plan,'state':'healthy'},{**plan,'streamsOmitted':-1},{**plan,'capacity':plan['capacity']*3}):
        assert list(validator(spec,shape).iter_errors(wrong)), 'Pool signal contract control accepted'
    captured=[c['response']['measurement'] for c in observations if c['path']=='/captures' and c['status']==201]
    reopened=[c['response']['measurement'] for c in observations if c['path']=='/captures/{id}' and c['status']==200]
    assert captured==reopened and len(captured)==3
    print('Pool signals API workflow: 9 actual handlers conform; independent capacity/wait execution and original-count offline reopen pass')


def check_host_observed(spec,path):
    observed=json.loads(Path(path).read_text());assert len(observed)==8
    for case in observed:
        op=spec['paths'][case['path']][case['method'].lower()]
        validator(spec,op['responses'][str(case['status'])]['content']['application/json']['schema']).validate(case['response'])
        if case.get('request') is not None and 'requestBody' in op:
            validator(spec,op['requestBody']['content']['application/json']['schema']).validate(case['request'])
    inventory=observed[0]['response'];assert len(inventory['hosts'])==1
    host=inventory['hosts'][0];assert 'service.name' not in host['resource'] and 'service.instance.id' not in host['resource']
    plan=observed[1]['response'];assert len(plan['widgets'])==2
    for widget in plan['widgets']:
        definition=widget['question']['definition'];binding=definition['source']
        assert binding['resource']=={'kind':'host','id':'machine-1'} and not binding['services'] and 'namespace' not in binding
        assert binding['exactStream']['resource']==host['resource']
    captured=[c['response']['measurement'] for c in observed if c['status']==201]
    reopened=[c['response']['measurement'] for c in observed if c['path']=='/captures/{id}']
    assert captured==reopened and len(captured)==2
    print('Host API workflow: 8 real handler responses conform; host-only gauge/stock execution and offline exact reopen pass')


if __name__ == '__main__':
    parser = argparse.ArgumentParser()
    parser.add_argument('--observations')
    parser.add_argument('--proxy-observations')
    parser.add_argument('--stock-observations')
    parser.add_argument('--stock-domain-observations')
    parser.add_argument('--pool-signals-observations')
    parser.add_argument('--host-observations')
    args = parser.parse_args()
    spec = json.loads((ROOT / 'api/openapi.json').read_text())
    check(spec)
    if args.observations:
        check_observed(spec, args.observations)
    if args.proxy_observations:
        check_observed(spec, args.proxy_observations)

    if args.stock_observations:
        check_stock_observed(spec,args.stock_observations)

    if args.stock_domain_observations:
        check_stock_domain_observed(spec,args.stock_domain_observations)

    if args.pool_signals_observations:
        check_pool_signals_observed(spec,args.pool_signals_observations)

    if args.host_observations:
        check_host_observed(spec,args.host_observations)
