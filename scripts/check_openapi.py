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
    for path in ('/package', '/example-questions', '/prebuilts', '/query', '/explain'):
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
    assert responses == 277 and requests == 93 and mux_refusals == 15, (responses, requests)
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


if __name__ == '__main__':
    parser = argparse.ArgumentParser()
    parser.add_argument('--observations')
    parser.add_argument('--proxy-observations')
    args = parser.parse_args()
    spec = json.loads((ROOT / 'api/openapi.json').read_text())
    check(spec)
    if args.observations:
        check_observed(spec, args.observations)
    if args.proxy_observations:
        check_observed(spec, args.proxy_observations)
