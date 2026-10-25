"""Readonly API contract smoke against the rebuilt local lab, using real data.
Run with requirements-api.txt installed; optional base URL is the Python server.
"""
import json
import sys
import time
import urllib.error
import urllib.request
import uuid
from pathlib import Path

from check_openapi import validator

base = sys.argv[1].rstrip('/') if len(sys.argv) > 1 else 'http://127.0.0.1:8088'
spec = json.loads((Path(__file__).resolve().parents[1] / 'api/openapi.json').read_text())
with urllib.request.urlopen(base + '/openapi.json', timeout=15) as response:
    assert json.load(response) == spec, 'Running specification differs from checkout; rebuild query image'
checks = 0


def check(method, path, pattern, payload=None, want=200):
    global checks
    data = json.dumps(payload).encode() if payload is not None else None
    request = urllib.request.Request(base + '/api' + path, data=data, method=method, headers={'Content-Type':'application/json'})
    try:
        response = urllib.request.urlopen(request, timeout=90)
    except urllib.error.HTTPError as error:
        response = error
    with response:
        result = json.load(response)
        assert response.status == want, (path, response.status, result)
        assert response.headers['Cache-Control'] == 'no-store'
        op = spec['paths'][pattern][method.lower()]
        validator(spec, op['responses'][str(want)]['content']['application/json']['schema']).validate(result)
        if payload is not None:
            validator(spec, op['requestBody']['content']['application/json']['schema']).validate(payload)
    checks += 1
    return result


for path in ('/measurement-packages','/operation-packages','/investigation-workflows','/retained-evidence/status'):
    check('GET', path, path)
check('GET', '/operations?unsupported=1', '/operations', want=400)
now = int(time.time()*1000)
check('POST', '/metric-discovery', '/metric-discovery', {'environment':'local','service':'lab-go-checkout','fromMs':now-120000,'toMs':now})
records = check('POST', '/explore', '/explore', {'signal':'spans','environment':'local','fromMs':now-120000,'toMs':now,'limit':10,'offset':0})
assert int(records['count']) >= len(records['rows']), 'Page became population'
trace_question = {'version':'goo/trace-search/v1', 'environment':'api-contract-' + uuid.uuid4().hex,
                  'fromMs':now-120000, 'toMs':now, 'limit':10,
                  'conditions':[{'field':{'kind':'column','key':'service'},'op':'eq','value':'api-contract'}]}
traces = check('POST', '/trace-search', '/trace-search', trace_question)
assert traces['state'] == 'no_data' and traces['matchedTraces'] == '0' and traces['items'] == [] and traces['hasMore'] is False
check('POST', '/trace-search', '/trace-search', {**trace_question, 'version':'unsupported'}, want=400)
print(f'Deployed API contract passes: matching spec, {checks} real handler responses, span count={records["count"]}, page={len(records["rows"])}; trace-search empty/refusal contracts pass; no records published')
