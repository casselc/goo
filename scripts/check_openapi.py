"""Validate OpenAPI, route coverage and frozen direct-Go wire fixtures.
No source/database calls. Request schemas cover decoding shape; Go still owns
semantic validation, byte budgets, capability lifetimes and effects.
"""
import argparse
import copy
import json
import re
from pathlib import Path

from jsonschema import Draft202012Validator
from openapi_spec_validator import validate_spec
from go_http_routes import route_sources

ROOT = Path(__file__).resolve().parents[1]


def validator(spec, schema):
    # Preserve document-local references without remote resolution/network access.
    return Draft202012Validator({'allOf': [schema], 'components': spec['components']})


def check_capture_reviews(spec):
    """The new review surface is strict metadata/custody, not a telemetry query."""
    routes=[('post','/record-reviews/tasks'),('get','/record-reviews/tasks/{id}'),
            ('post','/record-reviews/decisions'),('get','/record-reviews/decisions/{id}'),
            ('post','/record-reviews/manifests'),('get','/record-reviews/manifests/{id}'),
            ('get','/record-reviews/manifests/{id}/export'),('get','/record-reviews/records/{captureId}/{index}')]
    for method,path in routes:
        op=spec['paths'][path][method]
        assert op['x-go-source']=='cmd/goo/capture_reviews_http.go'
        assert op['x-schema-coverage']=='typed' and 'source-read' not in op['x-effects']
        assert not any(p['in']=='query' for p in op['parameters'])
        expected={'201','400','404','409','413','500'} if method=='post' else {'200','400','404','500'}
        assert set(op['responses'])==expected, (method,path,set(op['responses']))
        for response in op['responses'].values():
            validator(spec,response['headers']['Cache-Control']['schema']).validate('no-store')
            assert response['headers']['Cache-Control']['required']
        if method=='post':assert op['x-go-body-limit-bytes']==131072

    sha='a'*64
    task={'requestID':sha,'version':1,'name':'failure-shape','title':'Failure shape',
          'labels':[{'key':'timeout','title':'Timeout'},{'key':'not-timeout','title':'Not timeout'}]}
    decision={'requestID':sha,'taskID':sha,'target':{'captureId':'b'*64,'index':100},
              'parents':[],'judgment':{'kind':'label','value':'timeout'},'review':'reviewed','state':'active','note':''}
    manifest={'requestID':sha,'taskID':sha,'selected':['c'*64],'knownHeads':['d'*64],'historical':True}
    bodies={}
    for family,body in [('tasks',task),('decisions',decision),('manifests',manifest)]:
        shape=spec['paths']['/record-reviews/'+family]['post']['requestBody']['content']['application/json']['schema']
        check_body=validator(spec,shape);check_body.validate(body);bodies[family]=check_body
        for key in ('requestID','taskID' if family!='tasks' else 'version'):
            missing=copy.deepcopy(body);missing.pop(key)
            for bad in (missing,{**body,key:None}):
                assert not check_body.is_valid(bad), ('missing/null review field accepted',family,key)
        for bad in (None,[],{**body,'requestID':'A'*64},{**body,'requestID':'a'*63},
                    {**body,'record':{'untrusted':True}},{**body,'recordedAt':'2026-11-07T00:00:00Z'}):
            assert not check_body.is_valid(bad), ('untrusted review envelope accepted',family,bad)

    check=bodies['tasks']
    maximum={**task,'labels':[{'key':f'label-{i}','title':f'Label {i}'} for i in range(16)]}
    check.validate(maximum)
    for fields in ({'version':2},{'name':'Uppercase'},{'title':' '},{'title':'a'*121},
                   {'labels':None},{'labels':[]},{'labels':task['labels']*2},
                   {'labels':maximum['labels']+[{'key':'overflow','title':'Overflow'}]},
                   {'labels':[{'key':'timeout','title':'Timeout','automatic':True}]}):
        assert not check.is_valid({**task,**fields}), ('invalid review task admitted',fields)

    check=bodies['decisions']
    for kind in ('uncertain','abstain'):
        check.validate({**decision,'judgment':{'kind':kind}})
        for value in ('',None,'timeout'):
            assert not check.is_valid({**decision,'judgment':{'kind':kind,'value':value}}), 'nonlabel judgment carried a value'
    check.validate({**decision,'parents':None,'note':None})
    check.validate({**decision,'state':'withdrawn','parents':['c'*64]})
    bad_targets=[None,{}, {'captureId':'b'*64}, {'captureId':'b'*64,'index':None},
                 {'captureId':'b'*64,'index':-1},{'captureId':'b'*64,'index':1000},
                 {'captureId':'b'*64,'index':True},{'captureId':'b'*64,'index':'100'},
                 {'captureId':'b'*64,'index':0,'recordDigest':sha}]
    for target in bad_targets:
        assert not check.is_valid({**decision,'target':target}), ('untrusted occurrence admitted',target)
    for fields in ({'judgment':None},{'judgment':{'kind':'label'}},
                   {'judgment':{'kind':'label','value':None}},
                   {'judgment':{'kind':'abstain','value':'timeout'}},
                   {'judgment':{'kind':'automatic','value':'timeout'}},
                   {'review':'confirmed'},{'state':'deleted'},{'state':'withdrawn','parents':[]},
                   {'parents':[sha,sha]},{'parents':[f'{i:064x}' for i in range(9)]},
                   {'note':'a'*2049},{'note':'bad\x00value'}):
        assert not check.is_valid({**decision,**fields}), ('invalid review decision admitted',fields)

    check=bodies['manifests']
    check.validate({**manifest,'knownHeads':None})
    missing_historical=copy.deepcopy(manifest);missing_historical.pop('historical')
    assert not check.is_valid(missing_historical), 'implicit historical export admitted'
    for fields in ({'historical':False},{'historical':None},{'historical':'true'},
                   {'selected':[]},{'selected':None},{'selected':[sha,sha]},
                   {'selected':[f'{i:064x}' for i in range(33)]},
                   {'knownHeads':[f'{i:064x}' for i in range(65)]},{'latest':True},
                   {'items':[{'label':'timeout','record':{}}]}):
        assert not check.is_valid({**manifest,**fields}), ('implicit/unbounded review manifest admitted',fields)

    op=spec['paths']['/record-reviews/records/{captureId}/{index}']['get']
    index=validator(spec,next(p['schema'] for p in op['parameters'] if p['name']=='index'))
    for value in ('0','100','999'):index.validate(value)
    for value in ('00','01','-1','-0','+0','1.0','1000',' 1'):
        assert not index.is_valid(value), ('noncanonical stored index admitted',value)
    export=spec['paths']['/record-reviews/manifests/{id}/export']['get']
    assert export['x-go-response-limit-bytes']==4194304
    media=export['responses']['200']['content']['application/x-ndjson']
    line={'taskID':sha,'decisionID':'b'*64,'target':decision['target'],'recordDigest':'c'*64,'label':'timeout','record':{'integer':18446744073709551615}}
    check_line=validator(spec,media['x-ndjson-item-schema']);check_line.validate(line)
    assert not check_line.is_valid({**line,'label':None})
    assert not check_line.is_valid({**line,'target':{'captureId':'b'*64}})
    assert not validator(spec,{'$ref':'#/components/schemas/CaptureReviewError'}).is_valid({'error':1})
    # These are actual parsed Go handler observations, not the older frozen
    # body/header envelope. Go tests separately check raw NDJSON framing/headers.
    observations=json.loads((ROOT/'testdata/http/capture-reviews.json').read_text())
    accepted=[];seen=set()
    for observation in observations:
        method,path=observation['method'].lower(),observation['path']
        seen.add((method,path));operation=spec['paths'][path][method]
        response=operation['responses'][str(observation['status'])]
        if path.endswith('/export'):
            check_line.validate(observation['response'])
        else:
            validator(spec,response['content']['application/json']['schema']).validate(observation['response'])
        if 'requestBody' in operation:
            shape=operation['requestBody']['content']['application/json']['schema']
            validator(spec,shape).validate(observation['request'])
            accepted.append((shape,observation['request']))
    assert seen==set(routes), 'actual review observations do not cover all eight operations'
    print(f'Captured-record review: {len(observations)} actual responses and {len(accepted)} requests conform; parsed export lines checked, framing/headers remain Go assertions')
    return len(observations),accepted


def check_package_scopes(spec):
    scopes=[{'version':'goo/environment-selector/v1','policy':policy,'scope':{'mode':mode,**({'value':"prod-'雪"} if mode=='named' else {})}} for policy,mode in [('current-only','named'),('effective','named'),('effective','empty'),('effective','unreported')]]
    base={'adapter':'clickhouse.otel/v1','signal':'spans'}
    for wire in ('exploration.PackageSource','exploration.PackageSourceInput'):
        check=validator(spec,{'$ref':'#/components/schemas/'+wire})
        check.validate({**base,'environment':'prod'})
        for invalid in ({'environment':'prod'},{**base,'signal':'metrics','environment':'prod'},{**base,'adapter':'other','environment':'prod'},{**base,'environment':'bad\x00'}):
            assert not check.is_valid(invalid), 'unsupported package record source accepted'
        for selector in scopes:
            selected={**base,'environmentSelector':selector}
            check.validate(selected)
            for invalid in ({**selected,'environment':''},{**selected,'environment':'prod'},{**selected,'environment':None},{**selected,'unknown':True}):
                assert not check.is_valid(invalid), 'ambiguous package scope accepted'
        for invalid in (base,{**base,'environment':''},{**base,'environment':None},{**base,'environmentSelector':None},{**base,'environment':'prod','environmentSelector':None},{**base,'environmentSelector':{**scopes[0],'scope':{'mode':'empty'}}},{**base,'environmentSelector':{**scopes[1],'scope':{'mode':'empty','value':'prod'}}}):
            assert not check.is_valid(invalid), 'invalid package selector union accepted'
    definition=validator(spec,{'$ref':'#/components/schemas/exploration.TelemetryPackageInput'})
    assert not definition.is_valid({'source':None}), 'null package source accepted'
    assert not definition.is_valid({}), 'missing package source accepted'


def check_typed_reported_edges(spec):
    edge={'parent':{'service':'api', 'namespace':{'mode':'missing'}},'child':{'service':'worker', 'namespace':{'mode':'value','value':''}}}
    base={'signal':'spans','fromMs':1000,'toMs':61000,'limit':100,'offset':0,'reportedEdge':edge,'summary':'span-duration','spanStatus':True,'groupBy':[{'kind':'column','key':'name'}]}
    scopes=[{'version':'goo/environment-selector/v1','policy':policy,'scope':{'mode':mode,**({'value':'prod'} if mode=='named' else {})}} for policy,mode in [('current-only','named'),('effective','named'),('effective','empty'),('effective','unreported')]]
    for wire in ('exploration.Query','exploration.QueryInput'):
        check=validator(spec,{'$ref':'#/components/schemas/'+wire})
        check.validate({**base,'environment':'prod'})
        for selector in scopes:
            selected={**base,'environmentSelector':selector,**({'environment':''} if wire=='exploration.Query' else {})}
            check.validate(selected)
            assert not check.is_valid({**selected,'signal':'logs'}), 'reported edges must remain span-only'
            assert not check.is_valid({**selected,'environment':'prod'}), 'typed edge cannot project to an ambiguous scalar scope'


def check_typed_trace_search(spec):
    base={'version':'goo/trace-search/v1','fromMs':1000,'toMs':61000,'limit':10,'conditions':[{'field':{'kind':'column','key':'service'},'op':'eq','value':'api'}]}
    scopes=[{'version':'goo/environment-selector/v1','policy':policy,'scope':{'mode':mode,**({'value':'prod'} if mode=='named' else {})}} for policy,mode in [('current-only','named'),('effective','named'),('effective','empty'),('effective','unreported')]]
    for wire in ('exploration.TraceSearchRequestInput','exploration.TraceSearchRequest'):
        check=validator(spec,{'$ref':'#/components/schemas/'+wire})
        check.validate({**base,'environment':'prod'})
        for selector in scopes:
            request={**base,'environmentSelector':selector}
            check.validate(request)
            check.validate({**request,'environment':''})
            assert not check.is_valid({**request,'environment':'prod'}), 'mixed trace-search scopes accepted'
        for fields in ({},{'environment':''},{'environmentSelector':None},{'environmentSelector':{}},{'environmentSelector':{**scopes[1],'policy':'unknown'}},{'environmentSelector':{**scopes[1],'policy':'current-only','scope':{'mode':'empty'}}}):
            assert not check.is_valid({**base,**fields}), 'invalid trace-search selection admitted'
    witness=spec['components']['schemas']['exploration.TraceSearchWitness']
    assert witness['properties']['fieldValues']['minItems']==2


def check_view_artifacts(spec):
    fixtures=json.loads((ROOT/'testdata/http/view-artifacts.json').read_text())['cases']
    request=json.loads(next(c['body'] for c in fixtures if c['name']=='encode'))
    schema=spec['paths']['/view-artifacts/encode']['post']['requestBody']['content']['application/json']['schema']
    check=validator(spec,schema)
    check.validate(request)
    for change in (
        lambda a:a.update(version='goo.view-artifact/2'),
        lambda a:a.update(namespace='Team/Service'),
        lambda a:a.update(name=''),
        lambda a:a.update(revision='a'*65),
        lambda a:a.update(unknown=True),
        lambda a:a['view'].update(id='a'*64),
        lambda a:a['view'].update(tail={'signal':'logs','baseline':'a'*64}),
    ):
        bad=copy.deepcopy(request);change(bad['artifact'])
        assert not check.is_valid(bad),'invalid portable artifact admitted'
    for bad in ({}, {'artifact':None}, {'artifact':{}}):
        assert not check.is_valid(bad),'missing portable artifact admitted'
    # Both endpoints remain pure and use the existing bounded HTTP decoder.
    for path in ('/view-artifacts/encode','/view-artifacts/decode'):
        op=spec['paths'][path]['post']
        assert op['x-go-body-limit-bytes']==131072
        assert op['x-effects']=='pure-plan-or-supplied-summary'


def check_view_dataset(spec):
    """The current server adapter, including refusals the generic Go reflection cannot express."""
    check_view = validator(spec, {'$ref': '#/components/schemas/scopedview.Dataset'})
    revision = 'a' * 64
    snapshot = {
        'version': 'goo.dataset/1', 'id': revision + ':groups', 'revision': revision,
        'scope': {'id': revision, 'generation': revision, 'signal': 'spans', 'selector': {
            'signal': 'spans', 'environment': 'prod', 'fromMs': 1000, 'toMs': 61000,
            'limit': 100, 'offset': 0, 'groupBy': [{'kind': 'column', 'key': 'service'}]}},
        'shape': 'ranked', 'origin': 'source-read', 'dimensions': ['category'], 'capturedAt': '2026-10-30T00:00:00Z',
        'fields': [{'key': 'category', 'label': 'column:service', 'type': 'string', 'nullable': True},
                   {'key': 'count', 'label': 'Captured records', 'type': 'number', 'unit': 'records'}],
        'rows': [{'key': 'category:' + 'b' * 64, 'values': {'category': None, 'count': '18446744073709551615'}},
                 {'key': 'category:' + 'c' * 64, 'values': {'category': '', 'count': '0'}}],
        'quality': {'extent': 'complete', 'basis': 'captured-query', 'value': 'derived',
                    'sampling': 'unknown', 'freshness': 'fixed'},
        'completeness': {'basis': 'captured-query', 'revision': revision, 'scopeId': revision, 'aspects': ['groups']}}
    check_view.validate(snapshot)
    pattern=copy.deepcopy(snapshot)
    pattern['scope']['signal']='logs'
    pattern['scope']['selector'].update(signal='logs',groupBy=[{'kind':'column','key':'body_pattern_v1'}])
    pattern['fields'][0].update(label='column:body_pattern_v1',format='log-pattern-v1')
    pattern['fields'][0].pop('nullable')
    pattern['rows'][0]['values']['category']='lp1:p:6f7264657220n'
    pattern['rows'][1]['values']['category']='lp1:e'
    check_view.validate(pattern)
    for change in (
        lambda x:x['fields'][0].pop('format'),
        lambda x:x['fields'][0].update(nullable=True),
        lambda x:x['rows'][0]['values'].update(category=None),
        lambda x:x['scope']['selector'].update(signal='spans'),
        lambda x:x['fields'][0].update(format='renderer-extension'),
    ):
        bad=copy.deepcopy(pattern);change(bad)
        assert not check_view.is_valid(bad), 'invalid finite log-pattern dataset shape accepted'
    unknown = copy.deepcopy(snapshot)
    unknown['quality']['extent'] = 'unknown'
    del unknown['completeness']
    check_view.validate(unknown)
    empty = copy.deepcopy(snapshot)
    empty['rows'] = []
    check_view.validate(empty)
    escaped = copy.deepcopy(snapshot)
    escaped['rows'][0]['values']['category'] = '"' * 1024
    check_view.validate(escaped)

    multi = copy.deepcopy(snapshot)
    multi['dimensions'].append('category1')
    multi['fields'].insert(1, {'key': 'category1', 'label': 'column:name', 'type': 'string', 'nullable': True})
    for row in multi['rows']:
        row['values']['category1'] = None
    check_view.validate(multi)
    time_series = copy.deepcopy(snapshot)
    time_series['shape'] = 'time-series'
    time_series['id'] = revision + ':intervals'
    del time_series['dimensions']
    time_series['scope']['selector']['bucketMs'] = 1000
    time_series['interval'] = {'fromMs': '1000', 'toMs': '3500', 'bucketMs': '1000'}
    time_series['fields'] = [
        {'key': 'fromMs', 'label': 'Interval start', 'type': 'timestamp', 'unit': 'ms'},
        {'key': 'toMs', 'label': 'Interval end', 'type': 'timestamp', 'unit': 'ms'},
        {'key': 'count', 'label': 'Captured records', 'type': 'number', 'unit': 'records', 'nullable': True}]
    time_series['rows'] = [
        {'key': 'interval:1000:2000', 'state': 'observed', 'values': {'fromMs': '1000', 'toMs': '2000', 'count': '18446744073709551615'}},
        {'key': 'interval:2000:3000', 'state': 'no_data', 'values': {'fromMs': '2000', 'toMs': '3000', 'count': '0'}},
        {'key': 'interval:3000:3500', 'state': 'no_data', 'values': {'fromMs': '3000', 'toMs': '3500', 'count': '0'}}]
    time_series['completeness']['aspects'] = ['intervals']
    check_view.validate(time_series)
    uncertain = copy.deepcopy(time_series)
    uncertain['quality']['extent'] = 'unknown'
    del uncertain['completeness']
    uncertain['rows'][1]['state'] = 'unknown'
    uncertain['rows'][1]['values']['count'] = None
    check_view.validate(uncertain)
    del uncertain['rows'][1]['values']['count']
    check_view.validate(uncertain)
    time_mutants = []
    for path, value in [
        (('rows', 0, 'values', 'count'), '0'),
        (('rows', 1, 'values', 'count'), '1'),
        (('rows', 0, 'values', 'fromMs'), 1000),
        (('completeness', 'aspects'), ['groups']),
        (('interval', 'bucketMs'), '0'),
        (('interval',), None),
        (('fields', 0, 'type'), 'number'),
    ]:
        mutant = copy.deepcopy(time_series)
        parent = mutant
        for key in path[:-1]:
            parent = parent[key]
        parent[path[-1]] = value
        time_mutants.append(mutant)
    unknown_zero = copy.deepcopy(uncertain)
    unknown_zero['rows'][1]['values']['count'] = '0'
    time_mutants.append(unknown_zero)
    complete_unknown = copy.deepcopy(time_series)
    complete_unknown['rows'][1]['state'] = 'unknown'
    complete_unknown['rows'][1]['values']['count'] = None
    time_mutants.append(complete_unknown)
    for mutant in time_mutants:
        assert not check_view.is_valid(mutant), 'Invalid explicit interval state accepted'

    measure_snapshot = check_view_measures(spec, snapshot, time_series)
    check_view_summary_comparison(spec, measure_snapshot)
    check_view_status(spec, snapshot, time_series)

    missing = object()
    cases = [
        (('completeness',), missing),
        (('quality', 'extent'), 'unknown'),
        (('completeness', 'aspects'), ['records']),
        (('rows', 0, 'values', 'count'), 18446744073709551615),
        (('rows', 0, 'values', 'count'), '01'),
        (('rows', 0, 'values', 'count'), '1' * 21),
        (('scope', 'selector'), {}),
        (('quality', 'sampling'), 'none'),
        (('fields', 0, 'nullable'), False),
        (('fields', 1, 'unit'), 'seconds'),
        (('rows',), None),
        (('rows', 0, 'key'), 'category:unbounded-raw-text'),
        (('rows', 0, 'values', 'category'), 'x' * 1025),
    ]
    for path, value in cases:
        mutant = copy.deepcopy(snapshot)
        parent = mutant
        for key in path[:-1]:
            parent = parent[key]
        if value is missing:
            del parent[path[-1]]
        else:
            parent[path[-1]] = value
        assert not check_view.is_valid(mutant), f'Invalid supplied view dataset accepted: {path}'


def check_native_histogram_comparison(spec, pair_path, evidence_path=None):
    pair = json.loads(Path(pair_path).read_text())
    check_pair = validator(spec, {'$ref': '#/components/schemas/exploration.ViewHistogramComparison'})
    check_pair.validate(pair)
    assert pair['state'] == 'observed'
    a,b=pair['baseline'],pair['selection']
    assert a['histogram']['total']=='3' and b['histogram']['total']=='8'
    assert a['capturedAt']!=b['capturedAt'] and a['scope']['id']!=b['scope']['id']
    assert a['scope']['selector']['kind']=='measurement-window'
    assert a['rows'][0]['values']['from']=='-inf' and a['rows'][-1]['values']['to']=='+inf'
    mutations=[
        (('baseline','scope','selector','kind'),'measurement-time-bucket'),
        (('baseline','histogram','closure'),'left-closed-right-open'),
        (('baseline','histogram','grain'),'records'),
        (('baseline','histogram','unit'),'records'),
        (('baseline','origin'),'source-read'),
        (('baseline','quality','extent'),'complete'),
        (('baseline','sourceBasis'),'bad'),
        (('baseline','scope','signal'),'spans'),
        (('baseline','histogram','support'),{'total':'3','valid':'3','missing':'0','empty':'0','invalid':'0'}),
        (('reason',),'source-reads-disagree'),
    ]
    for path,value in mutations:
        bad=copy.deepcopy(pair);parent=bad
        for key in path[:-1]:parent=parent[key]
        parent[path[-1]]=value
        assert not check_pair.is_valid(bad), f'Invalid native pair admitted: {path}'
    bad=copy.deepcopy(pair);bad.update(state='diagnostic',reason='source-reads-disagree')
    assert not check_pair.is_valid(bad)
    if evidence_path:
        evidence=json.loads(Path(evidence_path).read_text())
        check_evidence=validator(spec, {'$ref':'#/components/schemas/exploration.EvidenceComparison'})
        check_evidence.validate(evidence)
        source=evidence['baseline']['nativeHistogramSource']
        assert len(source['streams'])==1 and not any(key in source for key in ('rows','intervals','buckets'))
        for key,value in [('measurementBasis','bad'),('histogramBasis','bad'),('rows',[])]:
            bad=copy.deepcopy(evidence);bad['baseline']['nativeHistogramSource'][key]=value
            assert not check_evidence.is_valid(bad),f'Invalid compact source admitted: {key}'
    print('Actual Go native histogram pair and compact source context valid; eleven pair and three metadata negatives rejected')


def check_view_status(spec, ranked, intervals):
    """Status cells are exact count quantities and an operand-bound fraction,
    not another numeric-validity support partition or a comparison delta."""
    check_view = validator(spec, {'$ref': '#/components/schemas/scopedview.Dataset'})
    keys = ['status.error', 'status.ok', 'status.unset', 'status.unknown']
    ratio = 'status.error_fraction'
    fields = [{'key': key, 'label': key, 'type': 'number', 'unit': 'records', 'nullable': True,
               'aggregation': {'family': 'span-status', 'kind': 'count', 'method': 'exact'}} for key in keys]
    fields.append({'key': ratio, 'label': 'Recorded error fraction', 'type': 'rational', 'unit': '1', 'nullable': True,
                   'aggregation': {'family': 'ratio', 'kind': 'fraction', 'method': 'exact-rational',
                                   'numerator': 'status.error', 'denominator': 'count'}})
    supplied = copy.deepcopy(ranked)
    supplied['scope']['selector']['spanStatus'] = True
    supplied['fields'].extend(fields)
    supplied['rows'] = [supplied['rows'][0]]
    row = supplied['rows'][0]
    row['values'].update(count='4', **{key: '1' for key in keys}, **{ratio: '1/4'})
    row['cells'] = {key: {'state': 'derived'} for key in keys + [ratio]}
    check_view.validate(supplied)
    summary = copy.deepcopy(supplied)
    summary.update(shape='summary', id=summary['id'].split(':')[0]+':summary')
    summary.pop('dimensions')
    summary['fields'] = summary['fields'][1:]
    summary['rows'][0]['key'] = 'population'
    summary['rows'][0]['values'].pop('category')
    summary['completeness']['aspects'] = ['summary']
    check_view.validate(summary)
    empty = copy.deepcopy(summary)
    empty['rows'][0]['values'].update(count='0', **{key: '0' for key in keys}, **{ratio: None})
    empty['rows'][0]['cells'][ratio] = {'state': 'no_data'}
    check_view.validate(empty)
    timed = copy.deepcopy(intervals)
    timed['scope']['selector']['spanStatus'] = True
    timed['fields'].extend(fields)
    for item in timed['rows']:
        count = item['values']['count']
        item['values'].update({key: count if key == 'status.error' else '0' for key in keys})
        item['values'][ratio] = None if count == '0' else '1'
        item['cells'] = {key: {'state': 'derived'} for key in keys}
        item['cells'][ratio] = {'state': 'no_data' if count == '0' else 'derived'}
    check_view.validate(timed)
    unknown = copy.deepcopy(timed)
    unknown['quality']['extent'] = 'unknown'
    unknown.pop('completeness')
    item = unknown['rows'][0]
    item['state'] = 'unknown'
    item['values'].update(count=None, **{key: None for key in keys+[ratio]})
    item['cells'] = {key: {'state': 'unknown'} for key in keys+[ratio]}
    check_view.validate(unknown)
    for change in [
        lambda x: x['fields'][-1].update(unit='%'),
        lambda x: x['fields'][-1]['aggregation'].update(numerator='status.ok'),
        lambda x: x['fields'][-1]['aggregation'].update(denominator='numeric.mean'),
        lambda x: x['scope']['selector'].pop('spanStatus'),
        lambda x: x['rows'][0]['cells']['status.error'].update(support={'total':'4','valid':'4','missing':'0','empty':'0','invalid':'0'}),
        lambda x: x['rows'][0]['cells'][ratio].update(support={'total':'4','valid':'4','missing':'0','empty':'0','invalid':'0'}),
        lambda x: x['rows'][0]['values'].pop('status.unknown'),
        lambda x: x['fields'].pop(-2),
    ]:
        mutant = copy.deepcopy(supplied)
        change(mutant)
        assert not check_view.is_valid(mutant), 'Invalid status quantity contract accepted'
    zero_fraction = copy.deepcopy(empty)
    zero_fraction['rows'][0]['values'][ratio] = '0'
    zero_fraction['rows'][0]['cells'][ratio] = {'state':'derived'}
    assert not check_view.is_valid(zero_fraction), 'Empty denominator became a healthy zero fraction'
    unknown_zero = copy.deepcopy(unknown)
    unknown_zero['rows'][0]['values']['status.error'] = '0'
    assert not check_view.is_valid(unknown_zero), 'Unknown interval manufactured a status count'


def check_status_observed(spec, path):
    fixtures = json.loads(Path(path).read_text())
    assert set(fixtures) == {'ranked', 'intervals', 'summary', 'empty-ranked', 'empty-summary', 'unknown-intervals'}, 'Expected all six Go status-view DTOs'
    check_view = validator(spec, {'$ref': '#/components/schemas/scopedview.Dataset'})
    for name, snapshot in fixtures.items():
        check_view.validate(snapshot)
    for change in [
        lambda x: x['fields'][-1].update(unit='%'),
        lambda x: x['fields'][-1]['aggregation'].update(denominator='numeric.mean'),
        lambda x: x['rows'][0]['cells']['status.error'].update(support={'total':'4','valid':'4','missing':'0','empty':'0','invalid':'0'}),
    ]:
        mutant = copy.deepcopy(fixtures['ranked'])
        change(mutant)
        assert not check_view.is_valid(mutant), 'Malformed actual-Go status DTO accepted'
    print('Actual Go status views valid: six ranked/interval/summary/empty/unknown DTOs; metadata/support negative controls pass')


def check_native_histogram_views(spec,path):
    fixtures=json.loads(Path(path).read_text())
    assert set(fixtures)=={'observed','empty'}
    check=validator(spec,{'$ref':'#/components/schemas/scopedview.Dataset'})
    for value in fixtures.values(): check.validate(value)
    for mutate in [
        lambda x:x['histogram'].update(closure='left-closed-right-open'),
        lambda x:x['histogram'].update(unit='ns'),
        lambda x:x['histogram'].update(grain='records'),
        lambda x:x['histogram'].update(temporality='unknown'),
        lambda x:x['histogram'].update(binding={'field':{'kind':'attribute','key':'n'},'type':'uint64','unit':'s','parse':'canonical-decimal'}),
        lambda x:x['scope'].update(signal='spans'),
        lambda x:x['scope']['selector'].update(kind='records'),
        lambda x:x['fields'][2].update(unit='records'),
        lambda x:x['rows'][0]['values'].update(**{'from':None}),
        lambda x:x.update(dimensions=['from']),
    ]:
        mutant=copy.deepcopy(fixtures['observed']);mutate(mutant)
        assert not check.is_valid(mutant),'Malformed native histogram metadata accepted'
    print('Actual Go native histogram views valid: observed/empty; ten family/scope/boundary negatives rejected')



def check_native_histogram_time_views(spec,path):
    fixtures=json.loads(Path(path).read_text())
    assert set(fixtures)=={'observed','empty','shortFinal','consumerGroup'}
    check=validator(spec,{'$ref':'#/components/schemas/scopedview.Dataset'})
    for value in fixtures.values():check.validate(value)
    for mutate in [
        lambda x:x['scope']['selector'].update(widthNS='0'),
        lambda x:x['scope']['selector'].update(bucketIndex=200),
        lambda x:x['scope']['selector'].pop('widthNS'),
        lambda x:x['scope']['selector']['question']['definition'].pop('histogramBucketMs'),
        lambda x:x['scope'].update(id='a'*64+':native:0:0'),
        lambda x:x.update(id='a'*64+':native:0:0:bins'),
        lambda x:x.update(sourceBasis='bad'),
        lambda x:x['histogram'].update(grain='records'),
        lambda x:x['histogram'].update(closure='left-closed-right-open'),
        lambda x:x['scope']['selector'].update(kind='measurement-interval'),
        lambda x:x['rows'][0]['values'].update(**{'from':None}),
        lambda x:x['fields'][2].update(unit='records'),
    ]:
        mutant=copy.deepcopy(fixtures['observed']);mutate(mutant)
        assert not check.is_valid(mutant),'Malformed native time-bucket metadata accepted'
    print('Actual Go native time-bucket views valid: observed/empty/short-final/consumer-group; twelve metadata negatives rejected')


def check_histogram_observed(spec, path):
    fixtures=json.loads(Path(path).read_text())
    assert set(fixtures)=={'uint64','uint64-empty','int64','int64-empty','decimal64','decimal64-empty'}
    check=validator(spec,{'$ref':'#/components/schemas/scopedview.Dataset'})
    for snapshot in fixtures.values(): check.validate(snapshot)
    for change in [
        lambda x:x['histogram'].update(closure='right-closed'),
        lambda x:x['histogram'].update(binning='auto'),
        lambda x:x['rows'].pop(),
        lambda x:x['rows'][0]['values'].update(count=None),
        lambda x:x['rows'][0].update(cells={'count':{'state':'derived'}}),
        lambda x:x['fields'][0].update(type='number'),
        lambda x:x.update(partition={'kind':'disjoint-record-groups','measure':'count','total':'1'}),
    ]:
        mutant=copy.deepcopy(fixtures['uint64']);change(mutant)
        assert not check.is_valid(mutant),'Malformed actual-Go histogram DTO accepted'
    print('Actual Go histogram views valid: six integer/decimal/empty DTOs; metadata/bin/support negatives pass')


def check_histogram_comparisons(spec,path):
    fixtures=json.loads(Path(path).read_text())
    assert set(fixtures)=={'observed','no-valid','empty','diagnostic'}
    check=validator(spec,{'$ref':'#/components/schemas/exploration.ViewHistogramComparison'})
    for value in fixtures.values(): check.validate(value)
    for change in [
        lambda x:x.update(state='incompatible'),
        lambda x:x.update(reason='source-reads-disagree'),
        lambda x:x['baseline'].update(origin='source-read'),
        lambda x:x['baseline'].update(completeness={'basis':'captured-query','revision':x['baseline']['revision'],'scopeId':x['baseline']['scope']['id'],'aspects':['bins']}),
        lambda x:x.update(changes={}),
    ]:
        mutant=copy.deepcopy(fixtures['observed']);change(mutant)
        assert not check.is_valid(mutant),'Malformed held histogram comparison accepted'
    diagnostic=copy.deepcopy(fixtures['diagnostic']);diagnostic.pop('reason')
    assert not check.is_valid(diagnostic),'Diagnostic reason missing'
    print('Actual Go histogram comparisons valid: observed/no-valid/empty/diagnostic; custody/state negatives pass')


def check_duration_histograms(spec,path):
    fixtures=json.loads(Path(path).read_text());assert set(fixtures)=={'observed','empty'}
    check=validator(spec,{'$ref':'#/components/schemas/scopedview.Dataset'})
    for value in fixtures.values():check.validate(value)
    for change in [
        lambda x:x['histogram'].pop('kind'),
        lambda x:x['histogram'].update(support={'total':'5','valid':'5','missing':'0','empty':'0','invalid':'0'}),
        lambda x:x['histogram'].update(kind='numeric'),
        lambda x:x['rows'][0]['values'].update({'from':'+inf'}),
        lambda x:x['rows'][-1]['values'].update(to=None),
        lambda x:x['rows'][-1]['values'].update(to='18446744073709551616'),
        lambda x:x['fields'][0].update(type='rational'),
        lambda x:x['rows'].pop(),
    ]:
        mutant=copy.deepcopy(fixtures['observed']);change(mutant)
        assert not check.is_valid(mutant),'Malformed actual-Go duration histogram accepted'
    print('Actual Go duration histograms valid: observed/empty DTOs; tagged-union/open-bound negatives pass')


def check_view_measures(spec, ranked, intervals):
    check_view = validator(spec, {'$ref': '#/components/schemas/scopedview.Dataset'})
    supplied = copy.deepcopy(ranked)
    supplied['rows'] = [supplied['rows'][0]]
    row = supplied['rows'][0]
    row['values']['count'] = '5'
    row['cells'] = {}
    definitions = {
        'numeric.mean': ('7/6', 'USD', 'numeric', 'mean', 'exact-rational', None),
        'numeric.sum': ('7/2', 'USD', 'numeric', 'sum', 'exact', None),
        'numeric.min': ('1/2', 'USD', 'numeric', 'min', 'exact', None),
        'numeric.max': ('2', 'USD', 'numeric', 'max', 'exact', None),
        'duration.mean': ('3', 'ns', 'duration', 'mean', 'exact-rational', None),
        'duration.min': ('1', 'ns', 'duration', 'min', 'exact', None),
        'duration.max': ('5', 'ns', 'duration', 'max', 'exact', None),
        'duration.p50': ('3', 'ns', 'duration', 'quantile', 'quantilesExactHigh', '0.5'),
        'duration.p95': ('5', 'ns', 'duration', 'quantile', 'quantilesExactHigh', '0.95'),
        'duration.p99': ('5', 'ns', 'duration', 'quantile', 'quantilesExactHigh', '0.99')}
    for key, (value, unit, family, kind, method, quantile) in definitions.items():
        agg = {'family': family, 'kind': kind, 'method': method}
        if quantile:
            agg['quantile'] = quantile
        supplied['fields'].append({'key': key, 'label': key, 'type': 'rational', 'unit': unit, 'nullable': True, 'aggregation': agg})
        row['values'][key] = value
        row['cells'][key] = {'state': 'derived', 'support': {
            'total': '5', 'valid': '3' if family == 'numeric' else '5',
            'missing': '1' if family == 'numeric' else '0', 'empty': '0',
            'invalid': '1' if family == 'numeric' else '0'}}
    supplied['scope']['selector']['numeric'] = {'field': {'kind': 'attribute', 'key': 'cost'}, 'type': 'decimal64', 'scale': 2, 'unit': 'USD', 'parse': 'canonical-decimal'}
    supplied['scope']['selector']['summary'] = 'span-duration'
    check_view.validate(supplied)
    unavailable = copy.deepcopy(supplied)
    for key in definitions:
        if key.startswith('numeric.'):
            unavailable['rows'][0]['values'][key] = None
            unavailable['rows'][0]['cells'][key] = {'state': 'no_valid_values', 'support': {'total': '5', 'valid': '0', 'missing': '2', 'empty': '1', 'invalid': '2'}}
    check_view.validate(unavailable)
    no_data = copy.deepcopy(supplied)
    no_data['rows'][0]['values']['count'] = '0'
    for key in definitions:
        no_data['rows'][0]['values'][key] = None
        no_data['rows'][0]['cells'][key] = {'state': 'no_data', 'support': {part: '0' for part in ('total', 'valid', 'missing', 'empty', 'invalid')}}
    check_view.validate(no_data)
    unknown = copy.deepcopy(intervals)
    unknown['quality']['extent'] = 'unknown'
    del unknown['completeness']
    unknown['fields'].extend(copy.deepcopy(supplied['fields'][2:]))
    unknown['rows'] = [unknown['rows'][0]]
    unknown['rows'][0]['state'] = 'unknown'
    unknown['rows'][0]['values']['count'] = None
    unknown['rows'][0]['cells'] = {}
    for key in definitions:
        unknown['rows'][0]['values'][key] = None
        unknown['rows'][0]['cells'][key] = {'state': 'unknown'}
    check_view.validate(unknown)
    mutants = []
    for path, value in [
        (('rows', 0, 'values', 'numeric.mean'), 7/6),
        (('rows', 0, 'values', 'numeric.mean'), None),
        (('rows', 0, 'cells', 'numeric.mean', 'support', 'valid'), '0'),
        (('rows', 0, 'values', 'duration.p95'), '-1'),
        (('rows', 0, 'values', 'duration.p95'), '1/2'),
        (('fields', 10, 'unit'), 'ms'),
        (('fields', 10, 'aggregation', 'method'), 'estimated'),
        (('fields', 2, 'aggregation', 'method'), 'exact'),
        (('rows', 0, 'cells', 'numeric.mean'), None),
    ]:
        mutant = copy.deepcopy(supplied)
        parent = mutant
        for key in path[:-1]:
            parent = parent[key]
        parent[path[-1]] = value
        mutants.append(mutant)
    zero_sum = copy.deepcopy(unavailable)
    zero_sum['rows'][0]['values']['numeric.sum'] = '0'
    mutants.append(zero_sum)
    missing_cell = copy.deepcopy(supplied)
    del missing_cell['rows'][0]['cells']['numeric.mean']
    mutants.append(missing_cell)
    unknown_support = copy.deepcopy(unknown)
    unknown_support['rows'][0]['cells']['numeric.mean']['support'] = {part: '0' for part in ('total', 'valid', 'missing', 'empty', 'invalid')}
    mutants.append(unknown_support)
    for mutant in mutants:
        assert not check_view.is_valid(mutant), 'Invalid exact measure/support/method accepted'
    return supplied


def check_view_summary_comparison(spec, ranked):
    check_snapshot = validator(spec, {'$ref': '#/components/schemas/scopedview.Dataset'})
    check_comparison = validator(spec, {'$ref': '#/components/schemas/exploration.ViewComparison'})
    summary = copy.deepcopy(ranked)
    summary['shape'] = 'summary'
    summary['id'] = summary['scope']['id'] + ':summary'
    del summary['dimensions']
    summary['fields'] = summary['fields'][1:]
    summary['rows'][0]['key'] = 'population'
    del summary['rows'][0]['values']['category']
    summary['completeness']['aspects'] = ['summary']
    check_snapshot.validate(summary)
    held = copy.deepcopy(summary)
    held['origin'] = 'browser-held'
    held['quality']['extent'] = 'unknown'
    held['sourceBasis'] = 'unverified-caller-claim'
    del held['capturedAt']
    del held['completeness']
    check_snapshot.validate(held)
    comparison = {'version': 'goo.comparison/1', 'state': 'observed',
                  'baseline': copy.deepcopy(held), 'selection': copy.deepcopy(held), 'changes': {}}
    comparison['selection']['revision'] = 'd' * 64
    comparison['selection']['scope']['generation'] = 'd' * 64
    comparison['selection']['capturedAt'] = '2026-10-30T01:00:00Z'
    for field in held['fields']:
        change = {'state': 'derived', 'value': '0', 'unit': field['unit']}
        if 'aggregation' in field:
            change['aggregation'] = copy.deepcopy(field['aggregation'])
        comparison['changes'][field['key']] = change
    comparison['changes']['count']['reason'] = 'absolute-count-difference-not-normalized-rate'
    check_comparison.validate(comparison)
    no_data = copy.deepcopy(comparison)
    no_data['state'] = 'no_data'
    no_data['baseline']['rows'][0]['values']['count'] = '0'
    no_data['changes']['count']['value'] = '5'
    for key in no_data['baseline']['rows'][0]['cells']:
        no_data['baseline']['rows'][0]['values'][key] = None
        no_data['baseline']['rows'][0]['cells'][key] = {'state': 'no_data', 'support': {part: '0' for part in ('total', 'valid', 'missing', 'empty', 'invalid')}}
        no_data['changes'][key].update(state='unavailable', value=None, reason='empty-population')
    check_comparison.validate(no_data)
    diagnostic = copy.deepcopy(comparison)
    diagnostic['state'] = 'diagnostic'
    diagnostic['reason'] = 'baseline-source-reads-disagree'
    for change in diagnostic['changes'].values():
        change.update(state='unavailable', value=None, reason='source-reads-disagree')
    check_comparison.validate(diagnostic)
    incompatible = copy.deepcopy(diagnostic)
    incompatible.update(state='incompatible', reason='incompatible-fields')
    check_comparison.validate(incompatible)
    signed = copy.deepcopy(comparison)
    signed['changes']['duration.p95']['value'] = '-3'
    signed['changes']['duration.mean']['value'] = '-11/3'
    # Schema checks signed exact representation; runtime validates subtraction.
    check_comparison.validate(signed)

    source_mutants = []
    for path, value in [(('rows',), []), (('dimensions',), ['category']), (('fields', 1, 'nullable'), False)]:
        mutant = copy.deepcopy(summary)
        parent = mutant
        for key in path[:-1]:
            parent = parent[key]
        parent[path[-1]] = value
        source_mutants.append(mutant)
    missing_clock = copy.deepcopy(summary)
    del missing_clock['capturedAt']
    source_mutants.append(missing_clock)
    forged = copy.deepcopy(held)
    forged['completeness'] = copy.deepcopy(summary['completeness'])
    source_mutants.append(forged)
    forged_complete = copy.deepcopy(held)
    forged_complete['quality']['extent'] = 'complete'
    source_mutants.append(forged_complete)
    for mutant in source_mutants:
        assert not check_snapshot.is_valid(mutant), 'Invalid/forged population-summary claim accepted'

    changes = [
        (('capturedAt',), '2026-10-30T00:00:00Z'),
        (('scope',), {}),
        (('baseline', 'origin'), 'source-read'),
        (('changes', 'duration.p95', 'unit'), 'ms'),
        (('changes', 'duration.p95', 'aggregation', 'method'), 'exact'),
        (('changes', 'numeric.mean', 'value'), 0.5),
    ]
    for path, value in changes:
        mutant = copy.deepcopy(comparison)
        parent = mutant
        for key in path[:-1]:
            parent = parent[key]
        parent[path[-1]] = value
        assert not check_comparison.is_valid(mutant), f'Invalid descriptive comparison accepted: {path}'
    missing_change = copy.deepcopy(comparison)
    del missing_change['changes']['numeric.mean']
    assert not check_comparison.is_valid(missing_change), 'Incomplete declared comparison union accepted'
    bad_diagnostic = copy.deepcopy(diagnostic)
    bad_diagnostic['changes']['count'].update(state='derived', value='0')
    assert not check_comparison.is_valid(bad_diagnostic), 'Diagnostic count difference admitted'
    bad_empty = copy.deepcopy(no_data)
    bad_empty['changes']['count'].update(state='unavailable', value=None, reason='empty-population')
    assert not check_comparison.is_valid(bad_empty), 'Known empty absolute count difference withheld'


def check_log_pattern_display391(spec):
    """Pattern count intent is finite; labels and runtime authority are not stored."""
    declaration = {'version':'goo.view/1','scope':{'kind':'record-query','signal':'logs'},
                   'kind':{'name':'bar','version':1},
                   'display':{'category':'category','value':'count','profile':'log-pattern-count-v1','scale':'population'}}
    for suffix in ('','Input'):
        check = validator(spec, {'$ref':'#/components/schemas/workbench.ViewDeclaration'+suffix})
        check.validate(declaration)
        table=copy.deepcopy(declaration);table['kind']['name']='table';check.validate(table)
        ordinary=copy.deepcopy(declaration);ordinary['display'].pop('profile');check.validate(ordinary)
        for value in ('',None) if suffix else ('',):
            default=copy.deepcopy(ordinary);default['display']['profile']=value;check.validate(default)
        invalid=[]
        for value in ('log-pattern-v1','ordinary',1,False):
            bad=copy.deepcopy(declaration);bad['display']['profile']=value;invalid.append(bad)
        for value in ('numeric.mean','duration.p95'):
            bad=copy.deepcopy(declaration);bad['display'].update(value=value,scale='max');invalid.append(bad)
        bad=copy.deepcopy(declaration);bad['scope']['signal']='spans';invalid.append(bad)
        bad=copy.deepcopy(declaration);bad['display']['categoryLabel']='forged label';invalid.append(bad)
        for bad in invalid:
            assert not check.is_valid(bad), 'invalid log-pattern declaration shape accepted'
    field_check=validator(spec, {'$ref':'#/components/schemas/scopedview.Field'})
    field={'key':'category','label':'Log pattern','type':'string','format':'log-pattern-v1'}
    field_check.validate(field)
    for changes in ({'key':'count'},{'type':'number'},{'format':'arbitrary-renderer'}):
        bad={**field,**changes}
        assert not field_check.is_valid(bad), 'invalid pattern field format accepted'


def check_document_view_card(spec):
    valid = {'kind':'view','reference':'a'*64,'caption':'Pinned definition','text':'Finding only; evidence separate'}
    for suffix in ('','Input'):
        check = validator(spec, {'$ref':'#/components/schemas/workbench.Card'+suffix})
        check.validate(valid)
        for field,value in [('reference','latest'),('role','finding'),('evidence',['b'*64]),('key','c'*64),('review','d'*64),('readBasis','basis'),('text','x'*4001),('caption','x'*161),('step',{'kind':'record'})]:
            bad=copy.deepcopy(valid);bad[field]=value
            assert list(check.iter_errors(bad)), 'view card accepted execution/evidence or invalid reference'
        bad=copy.deepcopy(valid);del bad['reference']
        assert list(check.iter_errors(bad)), 'view card accepted missing definition reference'


def check_timeline_composition(spec):
    """Temporal intent is optional, finite and separate from ranked options."""
    aggregate = {'version':'goo.view/1','scope':{'kind':'record-query','signal':'spans'},
                 'kind':{'name':'bar','version':1},'display':{'category':'category','value':'count'}}
    records = {'version':'goo.view/1','scope':aggregate['scope'],
               'kind':{'name':'record-table','version':1},'display':{'columns':['utc','summary']}}
    timeline = {'version':'goo.view/1','scope':aggregate['scope'],
                'kind':{'name':'time-bar','version':1},'display':{'value':'count','requiresComplete':True}}
    valid = {'version':'goo.composition/1','kind':{'name':'stack','version':1},
             'children':[{'aggregate':aggregate},{'records':records},{'timeline':timeline}]}
    for suffix in ('','Input'):
        check = validator(spec, {'$ref':'#/components/schemas/workbench.ViewComposition'+suffix})
        check.validate(valid)
        two = copy.deepcopy(valid); two['children'].pop(); check.validate(two)
        table = copy.deepcopy(valid); table['children'][2]['timeline']['kind']['name']='table'; check.validate(table)
        for measure in ('duration.mean','duration.p95'):
            duration=copy.deepcopy(valid);duration['children'][2]['timeline']['display']['value']=measure;check.validate(duration)
        invalid = []
        for key,value in [('value','rate'),('value','duration.p50'),('value','duration.max'),('value','numeric.mean'),('scale','population'),('start','startMs')]:
            bad=copy.deepcopy(valid);bad['children'][2]['timeline']['display'][key]=value;invalid.append(bad)
        bad=copy.deepcopy(valid);bad['children'][2]['timeline']=None;invalid.append(bad)
        bad=copy.deepcopy(valid);bad['children'][2]['records']=records;invalid.append(bad)
        bad=copy.deepcopy(valid);bad['children'].append({'timeline':timeline});invalid.append(bad)
        bad=copy.deepcopy(valid)
        for child in bad['children']:child['hidden']=True
        invalid.append(bad)
        for bad in invalid:
            assert list(check.iter_errors(bad)), 'invalid timeline composition accepted'



def check_record_only_compositions(spec):
    """Raw records are visible without manufacturing an aggregate or template."""
    q={'signal':'logs','environment':'prod','fromMs':1000,'toMs':61000,'limit':100,'offset':0}
    record={'version':'goo.view/1','scope':{'kind':'record-query','signal':'logs'},
            'kind':{'name':'record-list','version':1},'display':{'columns':['utc','summary']}}
    composition={'version':'goo.composition/1','kind':{'name':'stack','version':1},'children':[{'records':record}]}
    view={'version':1,'title':'Raw records','queries':[q],'composition':composition}
    for suffix in ('','Input'):
        cc=validator(spec,{'$ref':'#/components/schemas/workbench.ViewComposition'+suffix})
        vc=validator(spec,{'$ref':'#/components/schemas/workbench.ExplorationView'+suffix})
        cc.validate(composition);vc.validate(view)
        spans=copy.deepcopy(view);spans['queries'][0]['signal']='spans';spans['composition']['children'][0]['records']['scope']['signal']='spans';spans['composition']['children'][0]['records']['kind']['name']='record-span-waterfall';vc.validate(spans)
        invalid=[]
        bad=copy.deepcopy(composition);bad['children'][0]['hidden']=True;invalid.append(bad)
        bad=copy.deepcopy(composition);bad['children'].append(copy.deepcopy(bad['children'][0]));invalid.append(bad)
        bad=copy.deepcopy(composition);bad['children'][0]['records']['scope']['signal']='metrics';invalid.append(bad)
        bad=copy.deepcopy(composition);bad['children'].append({'timeline':{'version':'goo.view/1','scope':record['scope'],'kind':{'name':'time-bar','version':1},'display':{'value':'count'}}});invalid.append(bad)
        bad=copy.deepcopy(composition);bad['children'].append({'comparison':{'version':'goo.view/1','scope':record['scope'],'kind':{'name':'captured-count-comparison','version':1},'baselineCapture':'a'*64,'display':{'value':'count'}}});invalid.append(bad)
        for bad in invalid:
            assert not cc.is_valid(bad),'invalid record-only shape admitted'
        unused=copy.deepcopy(view);unused['queries'][0].update(numeric=None,numericHistogram=None,reportedEdge=None,groupTimeSeries=False);vc.validate(unused)
        for field,value in [('numeric',{'field':{'kind':'attribute','key':'work'},'type':'int64','unit':'items','parse':'canonical-decimal'}),('numericHistogram',{'binning':'powers-of-two'}),('reportedEdge',{'parent':{'service':'api'},'child':{'service':'worker'}}),('groupTimeSeries',True)]:
            bad=copy.deepcopy(view);bad['queries'][0][field]=value
            assert not vc.is_valid(bad),'specialized record-only profile admitted'
        grouped=copy.deepcopy(view);grouped['queries'][0]['groupBy']=[{'kind':'column','key':'service'}]
        grouped['template']={'version':'goo.template/1','group':'destination','profile':'ordinary'}
        assert not vc.is_valid(grouped),'record-only category template admitted'
        for key,value in [('layout','records'),('columns',[]),('tail',None),('composition',None)]:
            bad=copy.deepcopy(view);bad[key]=value
            assert not vc.is_valid(bad),'raw view accepted another presentation owner'


def check_declarative_facets(spec):
    q={'signal':'logs','environment':'prod','fromMs':1000,'toMs':61000,'limit':100,'offset':0}
    scope={'kind':'record-query','signal':'logs'}
    records={'version':'goo.view/1','scope':scope,'kind':{'name':'record-list','version':1},'display':{'columns':['utc','summary']}}
    facet={'version':'goo.view/1','scope':scope,'kind':{'name':'categorical-facet','version':1},'field':{'kind':'resource','key':'service.name'},'display':{'basis':'applied','title':'Service'}}
    view={'version':1,'title':'Logs with service facet','queries':[q],'composition':{'version':'goo.composition/1','kind':{'name':'stack','version':1},'children':[{'records':records},{'facet':facet}]}}
    for suffix in ('','Input'):
        vc=validator(spec,{'$ref':'#/components/schemas/workbench.ExplorationView'+suffix})
        fc=validator(spec,{'$ref':'#/components/schemas/workbench.FacetViewDeclaration'+suffix})
        fc.validate(facet);vc.validate(view)
        hidden=copy.deepcopy(view);hidden['composition']['children'][1]['hidden']=True;vc.validate(hidden)
        span=copy.deepcopy(view);span['queries'][0].update(signal='spans',summary='span-duration',spanStatus=True,bucketMs=1000)
        for child in span['composition']['children']:next(v for k,v in child.items() if k!='hidden')['scope']['signal']='spans'
        vc.validate(span)
        templated=copy.deepcopy(view);templated['queries'][0]['groupBy']=[{'kind':'column','key':'service'}]
        templated['composition']['children'].append({'aggregate':{'version':'goo.view/1','scope':scope,'kind':{'name':'bar','version':1},'display':{'category':'category','value':'count'}}})
        templated['template']={'version':'goo.template/1','group':'destination','profile':'ordinary'}
        vc.validate(templated)
        for mutation in (lambda f:f.update(query=q),lambda f:f.update(values=['api']),lambda f:f['display'].update(title=None),lambda f:f['display'].update(basis='without-own-filter'),lambda f:f.update(field={'kind':'column','key':'duration_ns'})):
            bad=copy.deepcopy(facet);mutation(bad)
            assert not fc.is_valid(bad),'invalid/private facet declaration admitted'
        for mutation in (
            lambda v:v['composition']['children'][0].update(hidden=True),
            lambda v:v['composition']['children'].append(copy.deepcopy(v['composition']['children'][1])),
            lambda v:v['composition']['children'][1].update(records=None),
            lambda v:v['composition']['children'][1].update(facet=None),
            lambda v:v['queries'][0].update(numeric={}),
            lambda v:v['queries'][0].update(groupTimeSeries=True),
            lambda v:v.update(template={'version':'goo.template/1','group':'destination','profile':'ordinary'}),
        ):
            bad=copy.deepcopy(view);mutation(bad)
            assert not vc.is_valid(bad),'invalid facet composition admitted'


def check_source_refusals(spec):
    check=validator(spec,{'$ref':'#/components/schemas/SourceReadFailure'})
    reasons={'cancelled':'Source read cancelled.','timeout':'Source read timed out.','budget':'Source read exceeded a limit.','malformed':'Source response rejected.','unavailable':'Source unavailable.'}
    for kind,reason in reasons.items():
        good={'error':reason,'refusal':{'kind':kind,'detail':'Safe source-boundary detail.'}}
        check.validate(good)
        for bad in [
            {'error':reason},
            {'error':'https://private.invalid/?token=fixture','refusal':good['refusal']},
            {'error':reason,'refusal':{'kind':'unknown','detail':'Safe detail.'}},
            {'error':reason,'refusal':{'kind':kind,'detail':'x'*257}},
            {'error':reason,'refusal':{'kind':kind,'detail':'bad\ncontrol'}},
            {'error':reason,'refusal':{**good['refusal'],'retry':'never'}},
        ]:
            assert not check.is_valid(bad),'invalid source refusal admitted'
    for path in ('/explore','/explore/page','/explore/counted-values','/record-aggregates'):
        operation=spec['paths'][path]['post']
        assert operation['responses']['502']['content']['application/json']['schema']['$ref'].endswith('/SourceReadFailure')
        assert operation['responses']['400']['content']['application/json']['schema']['$ref'].endswith('/Error')

def check_fixed_query_count_views(spec):
    scope={'kind':'record-query','signal':'logs'}
    q={'signal':'logs','environment':'prod','fromMs':1000,'toMs':61000,'limit':100,'offset':0,'groupBy':[{'kind':'column','key':'service'}]}
    comparison={'version':'goo.view/1','scope':scope,'kind':{'name':'query-count-comparison','version':1},'baselineQuery':{**q,'where':{'field':{'kind':'attribute','key':'cohort'},'op':'eq','value':'baseline'}},'display':{'value':'count'}}
    view={'version':1,'title':'Fixed query cohorts','queries':[q],'composition':{'version':'goo.composition/1','kind':{'name':'stack','version':1},'children':[
        {'aggregate':{'version':'goo.view/1','scope':scope,'kind':{'name':'bar','version':1},'display':{'category':'category','value':'count'}}},
        {'records':{'version':'goo.view/1','scope':scope,'kind':{'name':'record-table','version':1},'display':{'columns':['utc','summary']}}},
        {'comparison':comparison,'hidden':True}]}}
    for suffix in ('','Input'):
        cc=validator(spec,{'$ref':'#/components/schemas/workbench.CountComparisonViewDeclaration'+suffix})
        vc=validator(spec,{'$ref':'#/components/schemas/workbench.ExplorationView'+suffix})
        ac=validator(spec,{'$ref':'#/components/schemas/workbench.ViewArtifact'+suffix})
        cc.validate(comparison);vc.validate(view)
        mean=copy.deepcopy(comparison)
        mean['scope']['signal']='spans'
        mean['baselineQuery'].update(signal='spans',summary='span-duration')
        mean['display']={'value':'duration.mean'}
        cc.validate(mean)
        for path,value in [('kind','captured-count-comparison'),('view','groups'),('order','change'),('signal','logs'),('summary','status')]:
            bad=copy.deepcopy(mean)
            if path=='kind':
                bad['kind']['name']=value;bad.pop('baselineQuery');bad['baselineCapture']='a'*64
            elif path in ('view','order'):bad['display'][path]=value
            else:bad['baselineQuery'][path]=value
            assert not cc.is_valid(bad),f'mean comparison admitted invalid {path}'
        paired_facet=copy.deepcopy(view)
        paired_facet['composition']['children'].append({'facet':{'version':'goo.view/1','scope':scope,'kind':{'name':'categorical-facet','version':1},'field':{'kind':'attribute','key':'http.request.method'},'display':{'basis':'applied'}}})
        vc.validate(paired_facet)
        hidden_facet=copy.deepcopy(paired_facet);hidden_facet['composition']['children'][-1]['hidden']=True
        vc.validate(hidden_facet)
        ac.validate({'version':'goo.view-artifact/1','namespace':'platform','name':'paired-facet','revision':'v1','view':paired_facet})
        all_roles=copy.deepcopy(paired_facet)
        all_roles['queries'][0]['bucketMs']=1000
        all_roles['composition']['children'][2]['comparison']['baselineQuery']['bucketMs']=1000
        all_roles['composition']['children'].append({'timeline':{'version':'goo.view/1','scope':scope,'kind':{'name':'time-bar','version':1},'display':{'value':'count'}}})
        vc.validate(all_roles)
        duplicate=copy.deepcopy(all_roles);duplicate['composition']['children'].append(copy.deepcopy(duplicate['composition']['children'][-1]))
        assert not vc.is_valid(duplicate),'sixth/duplicate composition role admitted'
        for hidden in (False,True):
            captured_facet=copy.deepcopy(paired_facet)
            child=captured_facet['composition']['children'][2]
            child['hidden']=hidden
            child['comparison'].pop('baselineQuery')
            child['comparison'].update(kind={'name':'captured-count-comparison','version':1},baselineCapture='a'*64)
            captured_facet['composition']['children'][-1]['hidden']=hidden
            assert not vc.is_valid(captured_facet),'facet admitted with captured comparison'
        previous=copy.deepcopy(view);previous['queries'][0]['fromMs']=61000;previous['queries'][0]['toMs']=121000
        vc.validate(previous)
        ac.validate({'version':'goo.view-artifact/1','namespace':'platform','name':'previous-counts','revision':'v1','view':previous})
        invalid_previous=copy.deepcopy(previous);invalid_previous['composition']['children'][2]['comparison']['baselineQuery']['environment']='prod';invalid_previous['composition']['children'][2]['comparison']['baselineQuery']['environmentSelector']={'version':'goo/environment-selector/v1','policy':'effective','scope':{'mode':'named','value':'prod'}}
        assert not vc.is_valid(invalid_previous),'mixed previous-window environment union accepted'

        ac.validate({'version':'goo.view-artifact/1','namespace':'platform','name':'fixed-cohorts','revision':'v1','view':view})
        for mutation in (
            lambda x:x.update(baselineCapture=''), lambda x:x.update(baselineCapture=None),
            lambda x:x.update(baselineQuery=None), lambda x:x.pop('baselineQuery'),
            lambda x:x.update(kind={'name':'captured-count-comparison','version':1}),
            lambda x:x.update(rows=[]),
        ):
            bad=copy.deepcopy(comparison);mutation(bad)
            assert not cc.is_valid(bad),'mixed/invalid fixed-query operand admitted'
        template=copy.deepcopy(view);template['template']={'version':'goo.template/1','group':'destination','profile':'ordinary','baseline':'destination'}
        assert not vc.is_valid(template),'fixed query pair admitted as reusable template'


def check_native_view_config(spec):
    """Native display configuration never borrows record or evidence authority."""
    paths=('/native-views','/native-views/{id}','/native-views/reuse')
    for path in paths:
        method='get' if '{id}' in path else 'post'
        op=spec['paths'][path][method]
        assert op['x-schema-coverage']=='typed' and 'source-read' not in op['x-effects']
        assert 'X-Goo-Query-ID' not in [p['name'] for p in op['parameters']]
    declaration={'version':'goo.view/1','scope':{'kind':'measurement-question','signal':'metrics'},'selection':{'kind':'export-interval'},'kind':{'name':'histogram','version':1},'display':{'category':'bucket','value':'observation-count','requiresComplete':True}}
    question={'definition':{'apiVersion':'goo/measurement-package/v1alpha1','name':'native-duration','title':'Native duration','kind':'histogram','environmentSelector':{'version':'goo/environment-selector/v1','policy':'effective','scope':{'mode':'named','value':'prod'}},'lookbackMinutes':1,'maxGapMs':60000,'grain':'per-stream','source':{'services':['api'],'scope':'sdk','version':'','schemaURL':'','metric':'duration','unit':'s'}},'fromMs':1000,'toMs':61000}
    view={'version':1,'title':'Interval distribution','question':question,'declaration':declaration}
    shape=spec['paths'][paths[0]]['post']['requestBody']['content']['application/json']['schema']
    create=validator(spec,shape);create.validate(view)
    saved=validator(spec,spec['paths'][paths[0]]['post']['responses']['201']['content']['application/json']['schema'])
    unsaved=validator(spec,spec['paths'][paths[2]]['post']['responses']['200']['content']['application/json']['schema'])
    saved.validate({**view,'id':'a'*64});unsaved.validate(view)
    assert not saved.is_valid(view), 'saved native view lacks immutable ID'
    assert not unsaved.is_valid({**view,'id':'a'*64}), 'reuse retained donor identity'
    table=copy.deepcopy(view);table['declaration']['kind']['name']='table';create.validate(table)
    milliseconds=copy.deepcopy(view);milliseconds['question']['definition']['source']['unit']='ms';create.validate(milliseconds)
    reuse=validator(spec,spec['paths'][paths[2]]['post']['requestBody']['content']['application/json']['schema'])
    reuse.validate({'id':'a'*64,'question':question})
    for bad in ({'id':'a'*63,'question':question},{'id':'A'*64,'question':question},{'id':'a'*64,'question':question,'source':view},{'id':'a'*64,'question':question,'result':{}},{'question':question}):
        assert not reuse.is_valid(bad), ('invalid native reuse envelope accepted',bad)
    mutations=[
        lambda x:x.update(id='a'*64),lambda x:x.update(basis='b'*64),
        lambda x:x.update(rows=[]),lambda x:x.update(readAt='2026-11-07T00:00:00Z'),
        lambda x:x['declaration'].update(stream={'name':'donor'}),
        lambda x:x['declaration'].update(intervalIndex=0),
        lambda x:x['declaration']['selection'].update(fromNS='1000000000'),
        lambda x:x['declaration']['selection'].update(kind='time-bucket'),
        lambda x:x['declaration']['scope'].update(kind='record-query'),
        lambda x:x['declaration']['scope'].update(signal='spans'),
        lambda x:x['declaration']['kind'].update(name='bar'),
        lambda x:x['declaration']['display'].update(requiresComplete=False),
        lambda x:x['declaration']['display'].update(value='count'),
        lambda x:x['declaration']['display'].update(count='10'),
        lambda x:x['question'].update(fromMs=None),
        lambda x:x['question'].update(fromNS=None),
        lambda x:x['question'].update(fromNS='',toNS=''),
        lambda x:x['question'].update(fromNS='1000000000'),
        lambda x:x['question']['definition'].update(kind='counter'),
        lambda x:x['question']['definition'].update(grain='consumer-group'),
        lambda x:x['question']['definition'].update(histogramBucketMs=1000),
        lambda x:x['question']['definition'].update(histogramBucketMs=None),
        lambda x:x['question']['definition']['source'].update(unit='bytes'),
        lambda x:x['question']['definition']['source'].update(resource={'kind':'host','id':'one'}),
        lambda x:x['question']['definition']['source'].update(resource=None),
    ]
    for mutate in mutations:
        bad=copy.deepcopy(view);mutate(bad)
        assert not create.is_valid(bad), ('native declaration/evidence arm accepted',bad)
    for field in ('version','title','question','declaration'):
        bad=copy.deepcopy(view);bad.pop(field)
        assert not create.is_valid(bad), ('native configuration missing required field',field)
    gauge=copy.deepcopy(view);gauge['question']['definition']['kind']='gauge';gauge['question']['definition']['source']['unit']='By'
    gauge['declaration'].update(selection={'kind':'retained-observations'},kind={'name':'timeline','version':1},display={'category':'time','value':'recorded-value','requiresComplete':False})
    create.validate(gauge);saved.validate({**gauge,'id':'c'*64});unsaved.validate(gauge)
    for kind in ('timeline','table'):
        value=copy.deepcopy(gauge);value['declaration']['kind']['name']=kind;create.validate(value)
    reuse.validate({'id':'c'*64,'question':gauge['question']})
    empty=copy.deepcopy(gauge);empty['question']['definition']['source']['unit']='';assert not create.is_valid(empty)
    empty['question']['definition']['source'].pop('unit');assert not create.is_valid(empty)
    gauge_mutations=[lambda x:x['declaration']['display'].pop('requiresComplete'),lambda x:x['declaration']['display'].update(requiresComplete=None),lambda x:x['declaration']['display'].update(requiresComplete=True),lambda x:x['declaration']['kind'].update(name='histogram'),lambda x:x['declaration']['selection'].update(kind='export-interval'),lambda x:x['question']['definition'].update(kind='histogram'),lambda x:x['question']['definition'].update(kind='counter'),lambda x:x['question']['definition'].update(histogramBucketMs=None),lambda x:x.update(result={}),lambda x:x['declaration'].update(selectedObservation=0)]
    for mutate in gauge_mutations:
        bad=copy.deepcopy(gauge);mutate(bad)
        assert not create.is_valid(bad), ('gauge declaration/evidence arm accepted',bad)
    counter=copy.deepcopy(view)
    counter['question']['definition'].update(kind='counter')
    counter['question']['definition']['source']['unit']='{item}'
    counter['declaration'].update(selection={'kind':'evaluated-intervals'},kind={'name':'timeline','version':1},display={'category':'time','value':'counter-rate','requiresComplete':False})
    for domain in ('', 'nonnegative-decimal'):
        for kind in ('timeline', 'table'):
            value=copy.deepcopy(counter);value['declaration']['kind']['name']=kind
            value['question']['definition']['counterDomain']=domain
            create.validate(value);saved.validate({**value,'id':'d'*64});unsaved.validate(value)
            reuse.validate({'id':'d'*64,'question':value['question']})
    create.validate(counter)  # Omitted domain is the existing integer evaluator.
    counter_mutations=[
        lambda x:x['question']['definition'].update(counterDomain='nonnegative-integer'),
        lambda x:x['question']['definition'].update(counterDomain=None),
        lambda x:x['question']['definition'].update(counterDomain='signed'),
        lambda x:x['question']['definition'].update(grain='consumer-group'),
        lambda x:x['question']['definition']['source'].update(resource={'kind':'host','id':'h'}),
        lambda x:x['declaration']['display'].update(requiresComplete=True),
        lambda x:x['declaration']['display'].pop('requiresComplete'),
        lambda x:x['declaration']['display'].update(value='recorded-value'),
        lambda x:x['declaration']['kind'].update(name='histogram'),
        lambda x:x['declaration']['selection'].update(kind='retained-observations'),
        lambda x:x['declaration']['selection'].update(kind='export-interval'),
        lambda x:x['declaration'].update(intervalIndex=0),
        lambda x:x.update(rows=[],basis='x'),
    ]
    for mutate in counter_mutations:
        bad=copy.deepcopy(counter);mutate(bad)
        assert not create.is_valid(bad), ('counter display accepted unsupported policy or evidence',bad)
    for base in (gauge,view):
        bad=copy.deepcopy(base);bad['question']['definition']['counterDomain']=''
        assert not create.is_valid(bad), 'counter policy field leaked into another family'
    # Existing ordinary record-template admission remains a separate schema arm.
    assert not validator(spec,{'$ref':'#/components/schemas/ExplorationView'}).is_valid(view)
    print('Native view configuration: closed histogram/gauge/counter arms, explicit false completeness, whole destination question, donor-evidence refusal and source-free route contracts pass')


def check_native_view_discovery(spec):
    op=spec['paths']['/native-views']['get']
    assert op['x-schema-coverage']=='typed' and 'source-read' not in op['x-effects']
    assert set(op['responses'])=={'200','400','500'}
    assert len(op['parameters'])==1 and op['parameters'][0]['name']=='after'
    page=validator(spec,op['responses']['200']['content']['application/json']['schema'])
    item={'id':'a'*64,'title':'Interval table','kind':'table','unit':'s','family':'histogram','selection':'export-interval'}
    complete={'items':[item],'partial':False,'diagnostics':[]}
    partial={'items':[],'partial':True,'diagnostics':[{'id':'b'*64,'reason':'unavailable-artifact'}]}
    page.validate(complete);page.validate(partial)
    gauge={'id':'c'*64,'title':'Gauge samples','kind':'timeline','unit':'By','family':'gauge','selection':'retained-observations'}
    page.validate({**complete,'items':[gauge]})
    for bad in ({**complete,'total':1},{**complete,'items':[{**item,'question':{}}]},{**complete,'items':[{**item,'kind':'bar'}]},{**complete,'items':[{**item,'unit':'By'}]},{**complete,'diagnostics':partial['diagnostics']},{**partial,'diagnostics':[]},{**partial,'nextCursor':''},{**partial,'diagnostics':[{'id':'b'*64,'reason':'open /private/path'}]},{**complete,'items':[item]*21}):
        assert not page.is_valid(bad), ('native discovery schema accepted malformed metadata',bad)
    counter={'id':'d'*64,'title':'Counter rate','kind':'timeline','unit':'{item}','family':'counter','selection':'evaluated-intervals','counterDomain':'nonnegative-integer'}
    for domain in ('nonnegative-integer','nonnegative-decimal'):
        page.validate({**complete,'items':[{**counter,'counterDomain':domain}]})
    for mutate in (lambda x:x.pop('counterDomain'),lambda x:x.update(counterDomain=''),lambda x:x.update(counterDomain=None),lambda x:x.update(counterDomain='signed'),lambda x:x.update(selection='retained-observations'),lambda x:x.update(kind='histogram')):
        bad=copy.deepcopy(counter);mutate(bad)
        assert not page.is_valid({**complete,'items':[bad]}), ('counter metadata lost applicability',bad)
    for base in (item,gauge):
        assert not page.is_valid({**complete,'items':[{**base,'counterDomain':'nonnegative-integer'}]}), 'non-counter catalog acquired counter policy'
    bank=json.loads((ROOT/'testdata/protocol/native-histogram-view-questions-373.json').read_text())
    assert bank['version']=='goo.native-histogram-question-cases/1'
    cases=bank['cases'];assert len(cases)>=57
    question=validator(spec,{'$ref':'#/components/schemas/NativeViewQuestionInput'})
    for case in cases:
        assert isinstance(case['runtimeValid'],bool) and isinstance(case['schemaValid'],bool)
        assert question.is_valid(case['question'])==case['schemaValid'], ('native question finite schema boundary',case['name'])
    print(f'Native discovery: bounded closed metadata; {len(cases)} explicit native question schema outcomes checked (runtime enclosure remains Go/JS semantic admission)')


def check_histogram_templates407(spec):
    """Whole authored histogram carriers are finite; cross-binding equality is runtime-owned."""
    cases=controls=0
    for suffix in ('','Input'):
        view=validator(spec,{'$ref':'#/components/schemas/workbench.ExplorationView'+suffix})
        declaration=validator(spec,{'$ref':'#/components/schemas/workbench.HistogramViewDeclaration'+suffix})
        request=validator(spec,{'$ref':'#/components/schemas/workbench.ViewTemplateExpansionRequestInput'})
        for signal in ('logs','spans'):
            for typ in ('int64','uint64','decimal64'):
                binding={'field':{'kind':'attribute','key':'work.amount'},'type':typ,'unit':'{item}','parse':'canonical-decimal'}
                if typ=='decimal64':binding['scale']=2
                query={'signal':signal,'environment':'prod','fromMs':1000,'toMs':61000,'limit':7,'offset':0,
                       'numeric':binding,'numericHistogram':{'binning':'powers-of-two'}}
                scope={'kind':'record-query','signal':signal}
                records={'version':'goo.view/1','scope':scope,'kind':{'name':'record-list','version':1},'display':{'columns':['utc','summary']}}
                for kind in ('histogram','table'):
                    histogram={'version':'goo.view/1','scope':scope,'kind':{'name':kind,'version':1},
                               'display':{'category':'from','value':'count','title':'Captured amounts','requiresComplete':True}}
                    source={'version':1,'title':'Captured amounts','queries':[query],
                            'template':{'version':'goo.template/1','group':'none','profile':'ordinary-histogram'},
                            'composition':{'version':'goo.composition/1','kind':{'name':'stack','version':1},'children':[{'histogram':histogram},{'records':records}]}}
                    declaration.validate(histogram);view.validate(source)
                    destination=copy.deepcopy(query);destination.update(environment='destination',fromMs=62000,toMs=122000,limit=3,bucketMs=10000)
                    if signal=='spans':destination.update(summary='span-duration',spanStatus=True)
                    request.validate({'source':source,'query':destination})
                    expanded=copy.deepcopy(source);expanded['queries']=[destination]
                    view.validate(expanded)
                    # Bounded display controls never carry a bin observation or calculation.
                    for change in ({'category':'to'},{'value':'numeric.mean'},{'normalization':'probability'},
                                   {'scale':'population'},{'requiresComplete':None},{'query':'*'},
                                   {'bins':[]},{'title':None}):
                        bad=copy.deepcopy(histogram);bad['display'].update(change)
                        assert not declaration.is_valid(bad),('invalid histogram display accepted',change)
                        controls+=1
                    for changes in ({'offset':1},{'groupBy':[{'kind':'column','key':'service'}]},
                                    {'groupTimeSeries':True},{'numeric':None},{'numericHistogram':None},
                                    {'numericHistogram':{'binning':'adaptive'}}):
                        bad=copy.deepcopy(source);bad['queries'][0].update(changes)
                        assert not view.is_valid(bad),('unsupported histogram question accepted',changes)
                        bad_request={'source':source,'query':{**destination,**changes}}
                        assert not request.is_valid(bad_request),('unsupported histogram destination accepted',changes)
                        controls+=2
                    for changes in ({'group':'destination'},{'profile':'ordinary-summary'},{'baseline':'destination'}):
                        bad=copy.deepcopy(source);bad['template'].update(changes)
                        assert not view.is_valid(bad),('foreign histogram binding accepted',changes)
                        controls+=1
                    for added in ({'summary':{'version':'goo.view/1','scope':scope,'kind':{'name':'measure-summary','version':1},'display':{'fields':['count']}}},
                                  {'histogram':histogram}):
                        bad=copy.deepcopy(source);bad['composition']['children'].append(added)
                        assert not view.is_valid(bad),'histogram acquired another role'
                        controls+=1
                    assert not request.is_valid({'source':source,'query':destination,'baselineCapture':'a'*64})
                    controls+=1;cases+=1
    print(f'Histogram templates407: {cases} whole output/Input carriers and {controls} shape/profile controls pass; exact numeric-binding/binning equality remains Go/JS-owned')


def check_record_rates406(spec):
    """Rates retain their own closed count/exposure contract and ordinary scope."""
    field={'key':'record.rate','label':'Captured log records/s','type':'rational','unit':'records/s','nullable':True,
           'aggregation':{'family':'rate','kind':'records-per-second','method':'exact-rational'}}
    check=validator(spec,{'$ref':'#/components/schemas/scopedview.Field'})
    check.validate(field)
    for changes in ({'unit':'1'},{'type':'number'},{'nullable':False},
                    {'aggregation':{'family':'ratio','kind':'fraction','method':'exact-rational','numerator':'status.error','denominator':'count'}}):
        assert not check.is_valid({**field,**changes}), ('invalid rate field accepted',changes)
    cell=validator(spec,{'$ref':'#/components/schemas/exploration.RecordRateViewCell'})
    cell.validate({'state':'derived','rate':{'count':'0','elapsedMs':'500'}})
    cell.validate({'state':'unknown'})
    for bad in ({'state':'derived'},{'state':'derived','rate':{'count':'1','elapsedMs':'0'}},
                {'state':'unknown','rate':{'count':'1','elapsedMs':'500'}},
                {'state':'no_data','rate':{'count':'0','elapsedMs':'500'}},
                {'state':'derived','rate':{'count':'1','elapsedMs':'500','extra':True}}):
        assert not cell.is_valid(bad), ('invalid rate operands accepted',bad)
    query={'signal':'logs','environment':'prod','fromMs':1000,'toMs':61000,'limit':7,'offset':0}
    scope={'kind':'record-query','signal':'logs'}
    summary={'version':'goo.view/1','scope':scope,'kind':{'name':'measure-summary','version':1},'display':{'fields':['count','record.rate']}}
    timeline={'version':'goo.view/1','scope':scope,'kind':{'name':'time-bar','version':1},'display':{'value':'record.rate'}}
    records={'version':'goo.view/1','scope':scope,'kind':{'name':'record-list','version':1},'display':{'columns':['utc','summary']}}
    source={'version':1,'title':'Captured activity','queries':[query],
            'composition':{'version':'goo.composition/1','kind':{'name':'stack','version':1},'children':[{'summary':summary},{'timeline':timeline},{'records':records}]}}
    for suffix in ('','Input'):
        view=validator(spec,{'$ref':'#/components/schemas/workbench.ExplorationView'+suffix})
        view.validate(source)
        for changes in ({'groupBy':[{'kind':'column','key':'service'}]},{'groupTimeSeries':True}):
            bad=copy.deepcopy(source);bad['queries'][0].update(changes)
            assert not view.is_valid(bad), ('grouped record rate accepted',changes)
        bad=copy.deepcopy(source);bad['composition']['children'][1]['timeline']['display']['profile']='grouped-count-v1'
        assert not view.is_valid(bad), 'record rate gained a grouped profile'
        template=copy.deepcopy(source);template['composition']['children'].pop(1)
        template['template']={'version':'goo.template/1','group':'none','profile':'ordinary-summary'}
        view.validate(template)
    print('Record rates406: typed rate operands, ordinary declarations and closed summary template pass; exact count/window arithmetic remains runtime-owned')


def check_summary_template_bindings404(spec):
    """Finite ungrouped summary template arm through real saved/expansion carriers."""
    binding={'version':'goo.template/1','group':'none','profile':'ordinary-summary'}
    scope={'kind':'record-query','signal':'logs'}
    summary={'version':'goo.view/1','scope':scope,'kind':{'name':'measure-summary','version':1},'display':{'fields':['count']}}
    records={'version':'goo.view/1','scope':scope,'kind':{'name':'record-table','version':1},'display':{'columns':['utc','summary']}}
    query={'signal':'logs','environment':'prod','fromMs':1000,'toMs':61000,'limit':7,'offset':0}
    source={'version':1,'title':'Reusable population facts','queries':[query],
            'composition':{'version':'goo.composition/1','kind':{'name':'stack','version':1},'children':[{'summary':summary},{'records':records}]},
            'template':binding}
    target={**query,'environment':'staging','limit':3,'where':{'field':{'kind':'column','key':'service'},'op':'eq','value':'api'}}
    request={'source':source,'query':target}
    for suffix in ('','Input'):
        bound=validator(spec,{'$ref':'#/components/schemas/workbench.ViewTemplateBinding'+suffix})
        view=validator(spec,{'$ref':'#/components/schemas/workbench.ExplorationView'+suffix})
        bound.validate(binding);view.validate(source)
        for changes in ({'group':'destination'},{'profile':'ordinary'},{'baseline':'destination'},{'baseline':None},{'profile':None},{'extra':True}):
            assert not bound.is_valid({**binding,**changes}), ('invalid summary binding admitted',changes)
        bad_views=[]
        bad=copy.deepcopy(source);bad['composition']['children'].pop();bad_views.append(bad)
        bad=copy.deepcopy(source);bad['composition']['children'][0]['summary']['display']['fields']=['numeric.mean'];bad_views.append(bad)
        for changes in ({'offset':1},{'groupBy':[{'kind':'column','key':'service'}]},
                        {'numeric':{'field':{'kind':'attribute','key':'work.amount'},'type':'int64','unit':'{item}','parse':'canonical-decimal'}},
                        {'groupTimeSeries':True}):
            bad=copy.deepcopy(source);bad['queries'][0].update(changes);bad_views.append(bad)
        bad=copy.deepcopy(source);bad['template']={'version':'goo.template/1','group':'destination','profile':'ordinary'};bad_views.append(bad)
        bad=copy.deepcopy(source);bad['composition']['children'].append({'aggregate':{'version':'goo.view/1','scope':scope,'kind':{'name':'bar','version':1},'display':{'category':'category','value':'count'}}});bad_views.append(bad)
        for bad in bad_views:
            assert not view.is_valid(bad), 'unsupported summary template carrier admitted'
    expansion=validator(spec,{'$ref':'#/components/schemas/workbench.ViewTemplateExpansionRequestInput'})
    expansion.validate(request)
    for changes in ({'offset':1},{'groupBy':[{'kind':'column','key':'service'}]}):
        bad=copy.deepcopy(request);bad['query'].update(changes)
        assert not expansion.is_valid(bad), ('invalid summary destination admitted',changes)
    bad={**request,'baselineCapture':'a'*64}
    assert not expansion.is_valid(bad), 'summary template gained comparison baseline'
    print('Summary templates404: finite closed binding and complete saved/expansion carriers pass; scope equality and field availability remain runtime-owned')


def check_current_rpc_template_bindings(spec):
    """Current RPC is an explicit binding, not a legacy alias or method witness."""
    binding = {'version':'goo.template/1','group':'destination','profile':'otel.rpc-server-current'}
    for suffix in ('','Input'):
        bound = validator(spec, {'$ref':'#/components/schemas/workbench.ViewTemplateBinding'+suffix})
        bound.validate(binding)
        for profile in ('otel.rpc-server','otel.messaging-consumer','rpc-server-current'):
            bad=copy.deepcopy(binding);bad['profile']=profile
            assert not bound.is_valid(bad), 'unsupported template profile admitted'
        reference = validator(spec, {'$ref':'#/components/schemas/workbench.ViewTemplateReference'})
        reference.validate({'id':'otel.rpc-server-current-methods','version':1,'title':'Current RPC methods'})
        bad={'id':'otel.rpc-server-methods','version':1,'title':'Legacy alias'}
        assert not reference.is_valid(bad), 'unshipped RPC template alias admitted'
    request = {'template':'otel.rpc-server-current-methods','version':1,'widget':{
        'key':'dimension','title':'Current RPC methods','state':'supported',
        'grain':'server-span-record','family':'rpc-server-current'}}
    request_check = validator(spec, {'$ref':'#/components/schemas/workbench.ViewTemplateExpansionRequestInput'})
    request_check.validate(request)
    bad=copy.deepcopy(request);bad['template']='otel.rpc-server-methods'
    assert not request_check.is_valid(bad), 'unshipped RPC expansion alias admitted'
    # Destination role and predicate relationships are independently checked by Go;
    # the wire profile is not evidence that a returned population conforms.


def check_comparison_template_bindings(spec):
    """Binding and request presence never silently inherit a specimen capture."""
    scope = {'kind':'record-query','signal':'spans'}
    aggregate = {'version':'goo.view/1','scope':scope,'kind':{'name':'bar','version':1},
                 'display':{'category':'category','value':'count'}}
    records = {'version':'goo.view/1','scope':scope,'kind':{'name':'record-table','version':1},
               'display':{'columns':['utc','summary']}}
    comparison = {'version':'goo.view/1','scope':scope,'kind':{'name':'captured-count-comparison','version':1},
                  'baselineCapture':'a'*64,'display':{'value':'count','view':'groups'}}
    source = {'version':1,'title':'Portable comparison','queries':[{'signal':'spans','environment':'prod',
              'fromMs':1000,'toMs':61000,'limit':100,'offset':0,'groupBy':[{'kind':'column','key':'service'}]}],
              'composition':{'version':'goo.composition/1','kind':{'name':'stack','version':1},
                             'children':[{'aggregate':aggregate},{'records':records},{'comparison':comparison,'hidden':True}]},
              'template':{'version':'goo.template/1','group':'destination','profile':'ordinary','baseline':'destination'}}
    plain = copy.deepcopy(source);plain['composition']['children'].pop();plain['template'].pop('baseline')
    request = {'source':source,'query':copy.deepcopy(source['queries'][0]),'baselineCapture':'b'*64}
    for suffix in ('','Input'):
        view_check = validator(spec, {'$ref':'#/components/schemas/workbench.ExplorationView'+suffix})
        req_check = validator(spec, {'$ref':'#/components/schemas/workbench.ViewTemplateExpansionRequestInput'})
        view_check.validate(source);view_check.validate(plain);req_check.validate(request)
        plain_request = {'source':plain,'query':request['query']};req_check.validate(plain_request)
        bad_requests = []
        for value in (None,'','latest','B'*64):
            bad=copy.deepcopy(request);bad['baselineCapture']=value;bad_requests.append(bad)
        bad=copy.deepcopy(request);bad.pop('baselineCapture');bad_requests.append(bad)
        bad=copy.deepcopy(plain_request);bad['baselineCapture']='b'*64;bad_requests.append(bad)
        bad=copy.deepcopy(request);bad['template']='otel.http-server-routes';bad_requests.append(bad)
        for bad in bad_requests:
            assert not req_check.is_valid(bad), 'invalid comparison-template request presence accepted'
        bad_views = []
        for value in (None,'','source'):
            bad=copy.deepcopy(source);bad['template']['baseline']=value;bad_views.append(bad)
        bad=copy.deepcopy(source);bad['template'].pop('baseline');bad_views.append(bad)
        bad=copy.deepcopy(source);bad['template']['Baseline']=bad['template'].pop('baseline');bad_views.append(bad)
        bad=copy.deepcopy(plain);bad['template']['baseline']='destination';bad_views.append(bad)
        for bad in bad_views:
            assert not view_check.is_valid(bad), 'invalid comparison-template binding presence accepted'


def check_record_span_waterfall(spec):
    declaration = {'version':'goo.view/1','scope':{'kind':'record-query','signal':'spans'},
                   'kind':{'name':'record-span-waterfall','version':1},
                   'display':{'columns':['utc','service','summary','context'],
                              'fieldColumns':[{'kind':kind,'key':key} for kind in ('attribute','resource') for key in ('zone','empty','missing','nested')]}}
    for suffix in ('','Input'):
        check = validator(spec, {'$ref':'#/components/schemas/workbench.RecordViewDeclaration'+suffix})
        check.validate(declaration)
        for signal in ('logs','metrics'):
            bad=copy.deepcopy(declaration);bad['scope']['signal']=signal
            assert not check.is_valid(bad), 'non-span waterfall declaration accepted'
        for kind in ('record-table','record-list'):
            valid=copy.deepcopy(declaration);valid['kind']['name']=kind;valid['scope']['signal']='logs'
            check.validate(valid)
        bad=copy.deepcopy(declaration);bad['kind']['version']=2
        assert not check.is_valid(bad), 'unknown waterfall version accepted'
        bad=copy.deepcopy(declaration);bad['display']['fieldColumns'].append({'kind':'attribute','key':'ninth'})
        assert not check.is_valid(bad), 'waterfall field budget changed'


def check_activity_scope_views(spec):
    """Saved rolling configuration retains exact checked policy and scope."""
    wires = [
        {'version':'goo/environment-selector/v1','policy':'current-only','scope':{'mode':'named','value':'prod'}},
        {'version':'goo/environment-selector/v1','policy':'effective','scope':{'mode':'named','value':'prod'}},
        {'version':'goo/environment-selector/v1','policy':'effective','scope':{'mode':'empty'}},
        {'version':'goo/environment-selector/v1','policy':'effective','scope':{'mode':'unreported'}},
    ]
    for suffix in ('','Input'):
        check = validator(spec, {'$ref':'#/components/schemas/workbench.ActivityView'+suffix})
        for wire in wires:
            view={'version':1,'title':'Exact scope','selector':{'signal':'logs','environment':'','environmentSelector':wire},'columns':['utc','summary']}
            check.validate(view)
            bad=copy.deepcopy(view);bad['selector']['environment']='prod'
            assert list(check.iter_errors(bad)), 'ambiguous Activity environment accepted'
            bad=copy.deepcopy(view);bad['selector']['environmentSelector']['policy']='wildcard'
            assert list(check.iter_errors(bad)), 'unknown Activity selection policy accepted'
    parameters={p['name']:p for p in spec['paths']['/activity-views']['get']['parameters']}
    assert parameters['environmentSelector']['schema']['maxLength']==1024
    assert parameters['environment']['schema']['minLength']==1


def check_record_step_selector(spec):
    selector={'version':'goo/environment-selector/v1','policy':'effective','scope':{'mode':'unreported'}}
    scope={'environment':'','environmentSelector':selector,'fromMs':1000,'toMs':2000}
    step={'version':1,'kind':'record','scope':scope,'view':'Records','record':{'signal':'logs','environment':'','environmentSelector':selector,'fromMs':1000,'toMs':2000,'limit':10,'offset':0}}
    for suffix in ('','Input'):
        check=validator(spec,{'$ref':'#/components/schemas/workbench.Step'+suffix})
        check.validate(step)
        ambiguous=copy.deepcopy(step);ambiguous['scope']['environment']='prod'
        assert list(check.iter_errors(ambiguous)), 'ambiguous authored environment accepted'
        invalid=copy.deepcopy(step);invalid['scope']['environmentSelector']['policy']='unknown'
        assert list(check.iter_errors(invalid)), 'unknown authored policy accepted'
        native={'version':1,'kind':'measurement','scope':copy.deepcopy(scope),'view':'Overview','measurement':{'apiVersion':'goo/measurement-package/v1alpha1','name':'typed-gauge','title':'Typed gauge','kind':'gauge','environmentSelector':copy.deepcopy(selector),'lookbackMinutes':1,'maxGapMs':60000,'grain':'per-stream','source':{'services':['api'],'scope':'sdk','version':'','schemaURL':'','metric':'load','unit':'items'}}}
        check.validate(native)
        native['measurement']['kind']='stock'
        assert list(check.iter_errors(native)), 'typed native stock step accepted'



def check_saved_trace_search_steps(spec):
    selection={'version':'goo/environment-selector/v1','policy':'effective','scope':{'mode':'unreported'}}
    scope={'environment':'','environmentSelector':selection,'fromMs':0,'toMs':2000}
    request={'version':'goo/trace-search/v1','environmentSelector':copy.deepcopy(selection),'fromMs':0,'toMs':2000,'conditions':[
        {'field':{'kind':'column','key':'service'},'op':'eq','value':'api'},
        {'field':{'kind':'column','key':'service'},'op':'eq','value':'worker'},
        {'numeric':{'field':{'kind':'attribute','key':'jobs'},'type':'uint64','unit':'1','parse':'canonical-decimal'},'op':'gte','value':'9007199254740993'}], 'limit':7}
    step={'version':1,'kind':'trace-search','scope':scope,'view':'Traces','traceSearch':request}
    for suffix in ('','Input'):
        check=validator(spec,{'$ref':'#/components/schemas/workbench.Step'+suffix})
        check.validate(step)
        changes=[lambda x:x.update(record={}),lambda x:x.update(measurement=None),lambda x:x.update(traceSearch=None),
                 lambda x:x.update(view='Records'),lambda x:x.update(unexpected=True),
                 lambda x:x['scope'].update(fromNS=''),lambda x:x['scope'].update(fromMs=None),
                 lambda x:x['scope'].pop('fromMs'),lambda x:x['traceSearch'].update(fromMs=None),
                 lambda x:x['traceSearch'].pop('fromMs'),lambda x:x['traceSearch'].update(limit=101),
                 lambda x:x['traceSearch'].update(unsupported=True),
                 lambda x:x['traceSearch']['conditions'][0]['field'].update(unsupported=True),
                 lambda x:x['traceSearch']['conditions'][2]['numeric'].update(unsupported=True),
                 lambda x:x['traceSearch']['conditions'][0].update(value=None),
                 lambda x:x['traceSearch']['conditions'][2]['numeric'].update(scale=None)]
        for change in changes:
            bad=copy.deepcopy(step);change(bad)
            assert list(check.iter_errors(bad)), ('saved search invalid wire accepted',suffix,bad)
        card_check=validator(spec,{'$ref':'#/components/schemas/workbench.Card'+suffix})
        card={'kind':'question','step':copy.deepcopy(step)}
        card_check.validate(card)
        card['key']='a'*64;card['evidence']=['b'*64]
        card_check.validate(card)  # Exact question/evidence compatibility is verified by Go against custody.
    assert step['traceSearch']['conditions'][2]['value']=='9007199254740993'
    print('Saved trace-search intent: closed typed arm, exact strings and evidence reference shape pass')

def check_captured_trace_search(spec):
    request={'version':'goo/trace-search/v1','environment':'prod','fromMs':1000,'toMs':2000,
             'conditions':[{'field':{'kind':'column','key':'service'},'op':'eq','value':'api'}], 'limit':7}
    question={'kind':'trace-search','traceSearch':request}
    for suffix in ('','Input'):
        check=validator(spec,{'$ref':'#/components/schemas/exploration.CaptureQuestion'+suffix})
        check.validate(question)
        changes=[lambda x:x.update(traceSearch=None),lambda x:x.pop('traceSearch'),
                 lambda x:x.update(explore=None),lambda x:x.update(measurement={}),
                 lambda x:x.update(kind='explore'),lambda x:x['traceSearch'].update(limit=101),
                 lambda x:x['traceSearch'].update(fromMs=None),lambda x:x['traceSearch'].update(unknown=True),
                 lambda x:x['traceSearch']['conditions'][0].update(value=None)]
        for change in changes:
            bad=copy.deepcopy(question);change(bad)
            assert list(check.iter_errors(bad)), ('captured search malformed question accepted',suffix,bad)
    print('Captured trace search: typed request, exclusive payload and null/budget controls pass')

def check_trace_search_pages403(spec):
    request={'version':'goo/trace-search/v1','environment':'prod','fromMs':1000,'toMs':2000,
             'conditions':[{'field':{'kind':'column','key':'service'},'op':'eq','value':'api'}],'limit':7}
    position={'version':'goo/trace-search-position/v1','requestHash':'a'*64,'lastTraceId':'1'*32}
    for suffix in ('','Input'):
        page=validator(spec,{'$ref':'#/components/schemas/exploration.TraceSearchPageRequestInput'})
        page.validate({'request':request,'position':position})
        for bad in ({'request':request},{'position':position},{'request':request,'position':None},
                    {'request':None,'position':position},{'request':request,'position':{**position,'lastTraceId':'0'*32}},
                    {'request':request,'position':{**position,'requestHash':'a'*63}},
                    {'request':request,'position':{**position,'extra':True}}):
            assert not page.is_valid(bad),('invalid trace page shape accepted',suffix,bad)
        capture=validator(spec,{'$ref':'#/components/schemas/exploration.CaptureQuestion'+suffix})
        capture.validate({'kind':'trace-search','traceSearch':request,'traceSearchPosition':position})
        for bad in ({'kind':'trace-search','traceSearch':request,'traceSearchPosition':None},
                    {'kind':'explore','explore':{'signal':'spans','environment':'prod','fromMs':1000,'toMs':2000,'limit':7},'traceSearchPosition':position}):
            assert not capture.is_valid(bad),('capture page position granted to wrong family',suffix,bad)
    operation=spec['paths']['/trace-search/page']['post']
    assert operation['x-go-body-limit-bytes']==66560
    assert any(p['name']=='X-Goo-Query-ID' for p in operation['parameters'])
    print('Trace pages403: closed position/page shapes, trace-only capture metadata and query admission pass (hash equality remains runtime-owned)')

def check_capture_envelopes387(spec):
    record={'signal':'spans','environment':'prod','fromMs':1000,'toMs':2000,'limit':100,'offset':0}
    grouped={**record,'groupBy':[{'kind':'column','key':'service'}]}
    measurement={'definition':{'apiVersion':'goo/measurement-package/v1alpha1','name':'temperature','title':'Temperature','kind':'gauge','environment':'prod','lookbackMinutes':1,'maxGapMs':60000,'grain':'per-stream','source':{'services':['api'],'scope':'sdk','version':'','schemaURL':'','metric':'temperature','unit':'Cel'}},'fromMs':1000,'toMs':2000}
    families={'explore':('explore',record),'inventory':('inventory',{'environment':'prod','fromMs':1000,'toMs':2000}),'comparison':('comparison',{'baseline':grouped,'selection':grouped}),'measurement':('measurement',measurement),'explanation':('explanation',{'version':'goo/explanation/v1','baseline':record,'selection':record,'fields':[{'kind':'column','key':'name'}]}),'trace-search':('traceSearch',{'version':'goo/trace-search/v1','environment':'prod','fromMs':1000,'toMs':2000,'conditions':[{'field':{'kind':'column','key':'service'},'op':'eq','value':'api'}],'limit':7})}
    page=validator(spec,{'$ref':'#/components/schemas/exploration.CapturePage'})
    reference=validator(spec,{'$ref':'#/components/schemas/CaptureReceiptMetadata'})
    for suffix in ('','Input'):
        question=validator(spec,{'$ref':'#/components/schemas/exploration.CaptureQuestion'+suffix})
        for kind,(arm,payload) in families.items():
            value={'kind':kind,arm:copy.deepcopy(payload)};question.validate(value)
            for bad in ({},None,{'kind':kind},{**value,arm:None},{**value,'kind':'unsupported'},{**value,'Kind':kind}):
                assert not question.is_valid(bad),('invalid capture question admitted',suffix,kind,bad)
            for foreign,_ in families.values():
                if foreign==arm:continue
                for extra in (None,{}):assert not question.is_valid({**value,foreign:extra}),('mixed capture question',suffix,kind,foreign)
            if suffix:continue
            metadata={'receipt':{'version':1,'id':'a'*64,'title':'Stored evidence','capturedAt':'2026-11-09T00:00:00Z','question':value},'definitionRevision':'','readAt':'','summary':{}}
            reference.validate(metadata)
            assert not page.is_valid(metadata), 'metadata acquired full capture capability'
            for foreign,_ in families.values():
                assert not reference.is_valid({**metadata,foreign:None}), 'reference admitted a payload arm'
            assert not reference.is_valid({**metadata,'retainedRecords':0,'recordsLimited':False}), 'reference admitted paging authority'
    assert len(spec['components']['schemas']['exploration.CapturePage']['oneOf'])==6
    print('Capture387: six exclusive questions, foreign/null refusal and separate receipt-only capability shapes pass')

def check_http_client_profile385(spec):
    check=validator(spec,{'$ref':'#/components/schemas/exploration.DashboardRequestInput'})
    request={'profile':'http-client','scope':{'environment':'prod','namespace':{'mode':'value','value':''},'fromMs':1000,'toMs':61000},
             'service':'caller','signals':['spans'],'fields':[{'kind':'attribute','key':'http.request.method'}]}
    check.validate(request)
    for profile in ('','auto','all-spans','http-server','rpc-server-current','rpc-server','db-client','messaging-consumer','messaging-producer',None):
        check.validate({**request,'profile':profile})
    omitted=copy.deepcopy(request);omitted.pop('profile');check.validate(omitted)
    for bad in ('HTTP-client','http-client-native','http_client',True,17,[]):
        assert list(check.iter_errors({**request,'profile':bad})), ('unknown dashboard profile accepted',bad)
    assert 'not attempts' in spec['components']['schemas']['exploration.DashboardRequestInput']['properties']['profile']['description']
    print('Outgoing HTTP385: exact profile enum, nullable generic selection and native/alias refusal shape pass')

def check(spec):
    review_responses,review_accepted=check_capture_reviews(spec)
    check_package_scopes(spec)
    check_typed_reported_edges(spec)
    check_typed_trace_search(spec)
    check_record_step_selector(spec)
    check_saved_trace_search_steps(spec)
    check_captured_trace_search(spec)
    check_trace_search_pages403(spec)
    check_capture_envelopes387(spec)
    check_http_client_profile385(spec)
    validate_spec(spec)
    check_activity_scope_views(spec)
    check_view_artifacts(spec)
    check_view_dataset(spec)
    check_timeline_composition(spec)
    check_record_only_compositions(spec)
    check_declarative_facets(spec)
    check_source_refusals(spec)
    check_fixed_query_count_views(spec)
    check_comparison_template_bindings(spec)
    check_current_rpc_template_bindings(spec)
    check_summary_template_bindings404(spec)
    check_record_rates406(spec)
    check_histogram_templates407(spec)
    check_native_view_config(spec)
    check_native_view_discovery(spec)
    check_record_span_waterfall(spec)
    check_document_view_card(spec)
    check_log_pattern_display391(spec)
    bank = json.loads((ROOT/'testdata/protocol/environment-selectors.json').read_text())
    for path in ('/explore','/tail'):
        body_schema = spec['paths'][path]['post']['requestBody']['content']['application/json']['schema']
        check_body = validator(spec,body_schema)
        base = {'signal':'logs'}
        if path == '/explore':
            base.update(fromMs=100000,toMs=160000,limit=100,offset=0)
        check_body.validate({**base,'environment':'prod'})
        check_body.validate({**base,'environment':'prod','environmentSelector':None})
        for case in bank['cases']:
            check_body.validate({**base,'environmentSelector':case['selector']})
            check_body.validate({**base,'environment':'','environmentSelector':case['selector']})
        good = bank['cases'][0]['selector']
        for invalid in (base,{**base,'environment':''},{**base,'environment':'prod','environmentSelector':good},{**base,'environmentSelector':None},{**base,'environmentSelector':{**good,'version':'unsupported'}}):
            assert not check_body.is_valid(invalid), f'{path}: invalid selector union accepted: {invalid}'

    for path in ('/inventory','/discovery','/dashboards/plan'):
        check_service=validator(spec,spec['paths'][path]['post']['requestBody']['content']['application/json']['schema'])
        def body(scope):
            return {'scope':scope,'service':'api','signals':['spans'],'fields':[],'instruments':[]} if path=='/dashboards/plan' else scope
        base={'fromMs':100000,'toMs':160000}
        check_service.validate(body({**base,'environment':'prod'}))
        for case in bank['cases']:
            check_service.validate(body({**base,'environmentSelector':case['selector']}))
        for service in ('', 'api', "api'雪"):
            selected=body({**base,'environment':'prod','service':service})
            if path=='/inventory':
                check_service.validate(selected)
            else:
                assert not check_service.is_valid(selected), f'{path}: inventory-only service selector accepted'
        if path=='/inventory':
            for service in (None, 3, 'a'*129, 'a\x00b'):
                assert not check_service.is_valid({**base,'environment':'prod','service':service}), 'invalid inventory service accepted'
        for scope in (base,{**base,'environment':None},{**base,'environment':'prod','environmentSelector':None},{**base,'environment':'prod','environmentSelector':good},{**base,'environmentSelector':None}):
            assert not check_service.is_valid(body(scope)), f'{path}: invalid service environment union accepted'

    native_definition={'apiVersion':'goo/measurement-package/v1alpha1','name':'typed-gauge','title':'Typed gauge','kind':'gauge','lookbackMinutes':1,'maxGapMs':60000,'grain':'per-stream','source':{'services':['api'],'scope':'sdk','version':'','schemaURL':'','metric':'load','unit':'items'}}
    native_check=validator(spec,{'$ref':'#/components/schemas/measurements.DefinitionInput'})
    discovery_check=validator(spec,spec['paths']['/metric-discovery']['post']['requestBody']['content']['application/json']['schema'])
    for case in bank['cases']:
        typed={**native_definition,'environmentSelector':case['selector']};native_check.validate(typed)
        for kind in ('counter','histogram'):
            positive={**typed,'kind':kind,'source':{**typed['source'],'unit':'ms' if kind=='histogram' else 'items'}}
            native_check.validate(positive)
        assert not native_check.is_valid({**typed,'kind':'histogram'}), 'typed non-duration histogram accepted'
        discovery_check.validate({'environmentSelector':case['selector'],'service':'api','fromMs':100000,'toMs':160000})
        for invalid in ({**typed,'environment':'prod'},{**typed,'environment':None},{**typed,'environmentSelector':None},{**typed,'kind':'stock'},{**typed,'grain':'consumer-group'},{**typed,'source':{**typed['source'],'resource':{'kind':'host','id':'host'}}}):
            assert not native_check.is_valid(invalid), 'invalid typed native union/profile admitted'

    routes = set(route_sources(ROOT))
    documented = {(m, p) for p, item in spec['paths'].items() for m in item if m in ('get', 'post') and 'x-go-source' in item[m]}
    assert routes == documented, f'Go route coverage changed: missing={routes-documented}, stale={documented-routes}'
    assert not any(path in spec['paths'] for path in ('/package','/query','/explain')), 'Retired demo API reintroduced'
    assert not any('x-python-source' in op for item in spec['paths'].values() for op in item.values())
    for path in ('/example-questions','/prebuilts','/ui-config','/health'):
        assert spec['paths'][path]['get']['x-go-source'] == 'cmd/goo/serve_ui.go'
    assert set(spec['paths']['/queries/{id}/cancel']['post']['responses']) == {'202','400','429'}
    for profile in ('lab','generic'):
        validator(spec, {'$ref':'#/components/schemas/UIConfig'}).validate({'version':'goo/ui-config/v1','profile':profile,'examples':{'state':'configured' if profile=='lab' else 'unconfigured'}})
    validator(spec, {'$ref':'#/components/schemas/Liveness'}).validate({'ok':True})
    assert list(validator(spec, {'$ref':'#/components/schemas/Liveness'}).iter_errors({'ok':False})), 'Liveness mutation accepted'
    validator(spec, {'$ref':'#/components/schemas/PrebuiltRegistry'}).validate(json.loads((ROOT/'definitions/prebuilts.json').read_text()))
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
    responses,requests,mux_refusals=review_responses,len(review_accepted),0
    accepted = list(review_accepted)
    for fixture in sorted((ROOT / 'testdata/http').glob('*.json')):
        data = json.loads(fixture.read_text())
        if not isinstance(data, dict) or 'cases' not in data:
            continue
        for case in data['cases']:
            path, method = case['path'].split('?',1)[0], case['method'].lower()
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
            if path=='/native-views' and method=='get' and case['status']==200:
                items=[item['id'] for item in payload['items']]
                diagnostics=[item['id'] for item in payload['diagnostics']]
                scanned=items+diagnostics
                assert len(scanned)<=20 and len(set(scanned))==len(scanned)
                assert items==sorted(items) and diagnostics==sorted(diagnostics)
                assert payload['partial']==bool(diagnostics)
                if 'nextCursor' in payload:assert len(scanned)==20 and payload['nextCursor']==max(scanned)
                assert case['headers']['Cache-Control']=='no-store'
            if path.startswith('/record-reviews/'):

                assert case['headers']['Cache-Control']=='no-store', (fixture.name,case['name'])
            if media == 'application/x-ndjson':
                assert case['response'].endswith('\n')
                lines=case['response'].splitlines()
                assert 1<=len(lines)<=32 and len(case['response'].encode())<=4194304
                check_line=validator(spec,content[media]['x-ndjson-item-schema'])
                for line in lines:check_line.validate(json.loads(line))
                assert case['headers']['Content-Disposition'].startswith('attachment;')
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
    # Native histogram discovery adds five actual successful GET responses
    # to the prior 471/186/28 fixture baseline.
    assert responses == 489 and requests == 195 and mux_refusals == 28, (responses, requests, mux_refusals)
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
    memory_cases=json.loads((ROOT/'testdata/http/container-memory.json').read_text())['cases']
    memory=next(json.loads(c['response']) for c in memory_cases if c['name']=='aligned-three-signals')
    memory_shape={'$ref':'#/components/schemas/ContainerMemoryView'}
    for change in ({'state':'observed'},{'state':'observed','value':50,'text':'50.0%'},{'state':'unavailable','reason':'missing','value':'0'},{'state':'healthy'}):
        wrong=copy.deepcopy(memory);wrong['percentages']['request']=change
        assert list(validator(spec,memory_shape).iter_errors(wrong)), 'Container memory percentage control accepted'
    for field,value in (('value',50),('value','-1'),('observationState','healthy'),('role','pod-total')):
        wrong=copy.deepcopy(memory);wrong['signals'][0][field]=value
        assert list(validator(spec,memory_shape).iter_errors(wrong)), 'Container memory signal control accepted'

    cpu_cases=json.loads((ROOT/'testdata/http/container-cpu.json').read_text())['cases']
    cpu=next(json.loads(c['response']) for c in cpu_cases if c['name']=='aligned-three-signals')
    cpu_shape={'$ref':'#/components/schemas/ContainerCPUView'}
    for field,value in (('value',0.125),('value','-0.1'),('value','1e-3'),('value','0.0'),('value','1'*513),('role','working-set'),('observationState','healthy')):
        wrong=copy.deepcopy(cpu);wrong['signals'][0][field]=value
        assert list(validator(spec,cpu_shape).iter_errors(wrong)), 'Container CPU signal control accepted'
    for change in ({'state':'observed'},{'state':'observed','value':50,'text':'50.0%'},{'state':'unavailable','reason':'missing','value':'0'},{'state':'healthy'}):
        wrong=copy.deepcopy(cpu);wrong['percentages']['request']=change
        assert list(validator(spec,cpu_shape).iter_errors(wrong)), 'Container CPU percentage control accepted'

    history_cases=json.loads((ROOT/'testdata/http/container-resource-history.json').read_text())['cases']
    history=next(json.loads(c['response']) for c in history_cases if c['name']=='held-memory-and-cpu')
    history_shape={'$ref':'#/components/schemas/ContainerResourceHistory'}
    for field,value in (('value',50),('value','-1'),('state','healthy'),('observationState','fresh-when-observed'),('rawRowIndices',[]),('rawRowIndices',[0,0]),('rawRowIndices',[10000])):
        wrong=copy.deepcopy(history);wrong['series'][0]['points'][0][field]=value
        assert list(validator(spec,history_shape).iter_errors(wrong)), 'Resource history point control accepted'
    for field,value in (('unit','{cpu}'),('role','usage'),('state','healthy')):
        wrong=copy.deepcopy(history);wrong['series'][0][field]=value
        assert list(validator(spec,history_shape).iter_errors(wrong)), 'Resource history series control accepted'

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
    pod_cases=json.loads((ROOT/'testdata/http/pod-discovery.json').read_text())['cases']
    pod=next(json.loads(c['response']) for c in pod_cases if c['name']=='five-families-resource-variants')
    assert pod['sourceGroups']==7 and len(pod['observations'])==4
    assert any(o['resourceSchemaURL'] is None for o in pod['observations']) and any(o['resourceSchemaURL']=='' for o in pod['observations'])
    pod_shape={'$ref':'#/components/schemas/PodDiscovery'}
    for mode in ('numeric-count','actionable-missing-uid','excess-source-groups','unknown-family'):
        wrong=copy.deepcopy(pod)
        if mode=='numeric-count':wrong['observations'][0]['families'][0]['count']=2
        if mode=='actionable-missing-uid':
            entry=next(o for o in wrong['observations'] if o['podUIDState']=='missing');entry['actionable']=True
        if mode=='excess-source-groups':wrong['sourceGroups']=1001
        if mode=='unknown-family':wrong['observations'][0]['families'][0]['family']='kube-api'
        assert list(validator(spec,pod_shape).iter_errors(wrong)), 'Pod resource discovery control accepted'
    pod_native_cases=json.loads((ROOT/'testdata/http/pod-metrics.json').read_text())['cases']
    pod_native=next(json.loads(c['response']) for c in pod_native_cases if c['name']=='pod-metric-plan')
    assert len(pod_native['widgets'])==2
    native_shape={'$ref':'#/components/schemas/PodDashboardPlan'}
    for mode in ('wrong-profile','too-many-defaults','wrong-grain'):
        wrong=copy.deepcopy(pod_native)
        if mode=='wrong-profile':wrong['widgets'][0]['profile']='any-kubernetes'
        if mode=='wrong-grain':wrong['widgets'][0]['grain']='pod-health'
        if mode=='too-many-defaults':wrong['widgets']*=2
        assert list(validator(spec,native_shape).iter_errors(wrong)), 'Pod native planner control accepted'
    pod_question=pod_native['question']
    shape=spec['paths']['/pod-metric-discovery']['post']['requestBody']['content']['application/json']['schema']
    for mode in ('null','missing'):
        wrong=copy.deepcopy(pod_question)
        if mode=='null':wrong['resourceSchemaURL']=None
        else:wrong.pop('resourceSchemaURL')
        assert list(validator(spec,shape).iter_errors(wrong)), 'Unknown pod metric resource schema accepted'
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


def check_cpu_time_observed(spec,path):
    observations=json.loads(Path(path).read_text())
    assert len(observations)==9
    for case in observations:
        op=spec['paths'][case['path']][case['method'].lower()]
        validator(spec,op['responses'][str(case['status'])]['content']['application/json']['schema']).validate(case['response'])
        if case.get('request') is not None and 'requestBody' in op:
            validator(spec,op['requestBody']['content']['application/json']['schema']).validate(case['request'])
    result=next(c['response'] for c in observations if c['path']=='/measurements' and c['status']==200)
    assert result['question']['definition']['counterDomain']=='nonnegative-decimal'
    interval=result['counters']['series'][0]['intervals'][0]
    assert interval['delta']=='0.5' and interval['exactPerSecond']=='1/4' and interval['perSecond']==.25
    assert sum(c['status']==502 for c in observations)==5
    assert any(c['path']=='/captures/{id}' and c['status']==200 for c in observations)
    print('CPU-time observations: nine real handler responses/requests, five source/profile refusals and offline capture reopen conform')


def check_network_observed(spec,path,metric="system.network.io"):
    observations=json.loads(Path(path).read_text())
    assert len(observations)==11
    for case in observations:
        op=spec['paths'][case['path']][case['method'].lower()]
        validator(spec,op['responses'][str(case['status'])]['content']['application/json']['schema']).validate(case['response'])
        if case.get('request') is not None and 'requestBody' in op:
            validator(spec,op['requestBody']['content']['application/json']['schema']).validate(case['request'])
    r=next(c['response'] for c in observations if c['path']=='/measurements' and c['status']==200)
    assert r['question']['definition']['source']['metric']==metric
    interval=r['counters']['series'][0]['intervals'][0]
    assert interval['delta']=='512' and interval['exactPerSecond']=='256' and interval['perSecond']==256
    assert sum(c['status']==502 for c in observations)==7
    print(metric+' observations: eleven real handler responses/requests, seven source/profile refusals and offline reopen conform')


def check_filesystem_observed(spec,path,utilization=False):
    observations=json.loads(Path(path).read_text())
    assert len(observations)==13
    for case in observations:
        op=spec['paths'][case['path']][case['method'].lower()]
        validator(spec,op['responses'][str(case['status'])]['content']['application/json']['schema']).validate(case['response'])
        if case.get('request') is not None and 'requestBody' in op:
            validator(spec,op['requestBody']['content']['application/json']['schema']).validate(case['request'])
    r=next(c['response'] for c in observations if c['path']=='/measurements' and c['status']==200)
    assert r['question']['definition']['source']['metric']==('system.filesystem.utilization' if utilization else 'system.filesystem.usage')
    assert [p['value'] for p in r['gauges' if utilization else 'stocks']['series'][0]['samples']]==([.25,.75] if utilization else [4096,2048])
    assert sum(c['status']==502 for c in observations)==9
    print('Filesystem '+('utilization' if utilization else 'stock')+' observations: thirteen handler responses/requests, nine source/profile refusals and offline reopen conform')


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


def check_state_observed(spec, path, profile, label):
    cases = json.loads(Path(path).read_text())
    assert [(c['method'], c['path'], c['status']) for c in cases] == [
        ('POST', '/measurements', 200), ('POST', '/captures', 201),
        ('GET', '/captures/{id}', 200), ('POST', '/measurements', 400)]
    for case in cases:
        op = spec['paths'][case['path']][case['method'].lower()]
        validator(spec, op['responses'][str(case['status'])]['content']['application/json']['schema']).validate(case['response'])
        if case['method'] == 'POST':
            validator(spec, op['requestBody']['content']['application/json']['schema']).validate(case['request'])
    result = cases[0]['response']
    assert result['question']['definition']['kind'] == 'state'
    assert result['question']['definition']['stateProfile'] == profile
    assert result['states']['series'][0]['samples'][0]['label'] == label
    assert result['state'] == 'observed'
    assert cases[1]['response']['measurement'] == cases[2]['response']['measurement']
    print('State API: 4 actual handler responses conform; captured state reopens unchanged')


def check_pod_state_observed(spec, path):
    check_state_observed(spec, path, 'k8scluster-pod-phase/0.137.0', 'Unknown')


def check_container_state_observed(spec, path):
    check_state_observed(spec, path, 'k8scluster-container-ready/0.137.0', 'Not ready')
    cases = json.loads(Path(path).read_text())
    result = cases[0]['response']
    assert result['question']['definition']['source']['resource']['kind'] == 'kubernetes-container'
    assert result['states']['series'][0]['samples'][0]['code'] == 0


def check_pod_phase_discovery_observed(spec, path):
    cases = json.loads(Path(path).read_text())
    assert [(c['method'], c['path'], c['status']) for c in cases] == [
        ('POST', '/pod-metric-discovery', 200), ('POST', '/pod-dashboard/plan', 200),
        ('POST', '/measurements', 200)]
    for case in cases:
        op = spec['paths'][case['path']]['post']
        validator(spec, op['requestBody']['content']['application/json']['schema']).validate(case['request'])
        validator(spec, op['responses']['200']['content']['application/json']['schema']).validate(case['response'])
    plan = cases[1]['response']
    assert len(plan['widgets']) == 1 and plan['widgets'][0]['question']['definition']['kind'] == 'state'
    shape = spec['components']['schemas']['measurements.PodDashboardPlan']
    for kind in ('gauge', 'stock'):
        bad = copy.deepcopy(plan)
        for key in ('widgets', 'candidates'):
            bad[key][0]['question']['definition']['kind'] = kind
        assert list(validator(spec, shape).iter_errors(bad)), 'Phase profile accepted scalar kernel'
    bad = copy.deepcopy(plan)
    for key in ('widgets', 'candidates'):
        bad[key][0]['profile'] = 'collector-kubeletstats-v0.137.0'
    assert list(validator(spec, shape).iter_errors(bad)), 'Memory profile accepted state kernel'
    print('Pod phase discovery API: 3 actual handler responses conform; 3 profile/kind mutations rejected')


def check_container_discovery_observed(spec, path):
    check_pod_phase_discovery_observed(spec, path)
    cases = json.loads(Path(path).read_text())
    plan = cases[1]['response']
    widget = plan['widgets'][0]
    assert widget['grain'] == 'container-resource-stream'
    definition = widget['question']['definition']
    assert definition['stateProfile'] == 'k8scluster-container-ready/0.137.0'
    assert definition['source']['resource']['kind'] == 'kubernetes-container'
    assert definition['source']['resource']['id'] == plan['question']['resource']['container.id']
    assert cases[2]['response']['states']['series'][0]['samples'][0]['code'] == 0
    shape = spec['components']['schemas']['measurements.PodDashboardPlan']
    for field, value in [('grain', 'pod-resource-stream'), ('authority', 'kubernetes-pod')]:
        bad = copy.deepcopy(plan)
        for key in ('widgets', 'candidates'):
            if field == 'grain':
                bad[key][0]['grain'] = value
            else:
                bad[key][0]['question']['definition']['source']['resource']['kind'] = value
        assert list(validator(spec, shape).iter_errors(bad)), 'Container readiness accepted pod grain/authority'
    print('Container readiness discovery API: exact container authority and zero state conform; 2 authority/grain mutations rejected')


def check_unsupported_metric_discovery(spec, path):
    """Validate actual discovery wire results and focused rejecting controls."""
    report = json.loads(Path(path).read_text())
    operation = spec['paths']['/metric-discovery']['post']
    request = validator(spec, operation['requestBody']['content']['application/json']['schema'])
    response = validator(spec, operation['responses']['200']['content']['application/json']['schema'])
    for name in ('discovery', 'supportedDiscovery'):
        observed = report[name]
        request.validate(observed['request'])
        response.validate(observed['response'])
    actual = report['discovery']['response']
    assert actual['diagnostics'] == [] and actual['records'] == 13
    assert len(actual['instruments']) == 8
    for item in actual['instruments']:
        assert item['kind'] == 'exponential-histogram'
        assert item['supported'] is False and 'definition' not in item
        assert item['capabilityPlan']['eligible'] is False
    mutants = []
    for diagnostic in (None, [{'kind':'exponential-histogram','reason':'healthy'}],
                       [{'kind':'histogram','reason':'optional-table-missing'}]):
        mutant = copy.deepcopy(actual)
        mutant['diagnostics'] = diagnostic
        mutants.append(mutant)
    executable = next(item['definition'] for item in report['supportedDiscovery']['response']['instruments'] if item.get('definition'))
    for field, value in (('supported', True), ('definition', executable)):
        mutant = copy.deepcopy(actual)
        mutant['instruments'][0][field] = value
        mutants.append(mutant)
    mutant = copy.deepcopy(actual)
    mutant['instruments'][0]['capabilityPlan']['eligible'] = True
    mutants.append(mutant)
    for mutant in mutants:
        assert not response.is_valid(mutant), 'Unsupported discovery wire mutant accepted'
    print('Actual unsupported/supported discovery wire results valid; six rejecting controls pass')


def check_histogram_exemplar_observations(spec, path):
    observed = json.loads(Path(path).read_text())
    native = spec['paths']['/measurements']['post']
    response = validator(spec, native['responses']['200']['content']['application/json']['schema'])
    for name in ('parent', 'reactivatedParent'):
        validator(spec, native['requestBody']['content']['application/json']['schema']).validate(observed[name]['request'])
        response.validate(observed[name]['response'])
    records = spec['paths']['/explore']['post']
    validator(spec, records['requestBody']['content']['application/json']['schema']).validate(observed['child']['request'])
    validator(spec, records['responses']['200']['content']['application/json']['schema']).validate(observed['child']['response'])
    actual = observed['parent']['response']
    report = actual['histogramExemplars']
    assert report['valueEncoding'] == 'source-float64' and report['originalValueType'] == 'unknown'
    assert sorted(export['total'] for export in report['exports']) == ['0', '6']
    primary = next(index for index, export in enumerate(report['exports']) if export['total'] == '6')
    controls = []
    for key, value in (('version', 'unknown'), ('originalValueType', 'int64'), ('limited', 'false')):
        mutant = copy.deepcopy(actual)
        mutant['histogramExemplars'][key] = value
        controls.append(mutant)
    for key, value in (('total', ''), ('state', 'unavailable'), ('items', report['exports'][primary]['items'] * 6)):
        mutant = copy.deepcopy(actual)
        mutant['histogramExemplars']['exports'][primary][key] = value
        controls.append(mutant)
    for mutant in controls:
        assert not response.is_valid(mutant), 'Reported exemplar wire mutant accepted'
    print('Actual native exemplar and scoped span wire results valid; six rejecting controls pass')


if __name__ == '__main__':
    parser = argparse.ArgumentParser()
    parser.add_argument('--observations')
    parser.add_argument('--status-view-observations')
    parser.add_argument('--histogram-view-observations')
    parser.add_argument('--histogram-comparison-observations')
    parser.add_argument('--evidence-histogram-observations')
    parser.add_argument('--duration-histogram-observations')
    parser.add_argument('--stock-observations')
    parser.add_argument('--stock-domain-observations')
    parser.add_argument('--pool-signals-observations')
    parser.add_argument('--host-observations')
    parser.add_argument('--cpu-time-observations')
    parser.add_argument('--network-observations')
    parser.add_argument('--disk-observations')
    parser.add_argument('--filesystem-observations')
    parser.add_argument('--filesystem-utilization-observations')
    parser.add_argument('--pod-state-observations')
    parser.add_argument('--container-state-observations')
    parser.add_argument('--pod-phase-discovery-observations')
    parser.add_argument('--container-discovery-observations')
    parser.add_argument('--native-histogram-view-observations')
    parser.add_argument('--unsupported-metric-discovery-observations')
    parser.add_argument('--histogram-exemplar-observations')
    parser.add_argument('--native-histogram-time-view-observations')
    parser.add_argument('--native-histogram-comparison-observations')
    parser.add_argument('--native-histogram-evidence-observations')
    args = parser.parse_args()
    spec = json.loads((ROOT / 'api/openapi.json').read_text())
    check(spec)
    if args.histogram_exemplar_observations:
        check_histogram_exemplar_observations(spec,args.histogram_exemplar_observations)
    if args.unsupported_metric_discovery_observations:
        check_unsupported_metric_discovery(spec,args.unsupported_metric_discovery_observations)
    if args.native_histogram_comparison_observations:
        check_native_histogram_comparison(spec,args.native_histogram_comparison_observations,args.native_histogram_evidence_observations)
    if args.native_histogram_time_view_observations:
        check_native_histogram_time_views(spec,args.native_histogram_time_view_observations)
    if args.native_histogram_view_observations:
        check_native_histogram_views(spec,args.native_histogram_view_observations)
    if args.duration_histogram_observations:
        check_duration_histograms(spec,args.duration_histogram_observations)

    if args.evidence_histogram_observations:
        observed=json.loads(Path(args.evidence_histogram_observations).read_text())
        validator(spec,{'$ref':'#/components/schemas/exploration.EvidenceComparison'}).validate(observed)
        assert observed['histogramComparison']['baseline']['shape']=='histogram'
        assert len(observed['histogramComparison']['baseline']['scope']['selector']['groupBy'])==2
        print('Actual Go two-group EvidenceComparison histogram operands valid')

    if args.histogram_comparison_observations:
        check_histogram_comparisons(spec,args.histogram_comparison_observations)

    if args.histogram_view_observations:
        check_histogram_observed(spec,args.histogram_view_observations)

    if args.status_view_observations:
        check_status_observed(spec, args.status_view_observations)
    if args.filesystem_utilization_observations:
        check_filesystem_observed(spec,args.filesystem_utilization_observations,True)
    if args.filesystem_observations:
        check_filesystem_observed(spec,args.filesystem_observations)
    if args.disk_observations:
        check_network_observed(spec,args.disk_observations,"system.disk.io")
    if args.network_observations:
        check_network_observed(spec,args.network_observations)
    if args.cpu_time_observations:
        check_cpu_time_observed(spec,args.cpu_time_observations)
    if args.observations:
        check_observed(spec, args.observations)

    if args.stock_observations:
        check_stock_observed(spec,args.stock_observations)

    if args.stock_domain_observations:
        check_stock_domain_observed(spec,args.stock_domain_observations)

    if args.pool_signals_observations:
        check_pool_signals_observed(spec,args.pool_signals_observations)

    if args.host_observations:
        check_host_observed(spec,args.host_observations)

    if args.pod_state_observations:
        check_pod_state_observed(spec,args.pod_state_observations)

    if args.pod_phase_discovery_observations:
        check_pod_phase_discovery_observed(spec,args.pod_phase_discovery_observations)

    if args.container_state_observations:
        check_container_state_observed(spec,args.container_state_observations)

    if args.container_discovery_observations:
        check_container_discovery_observed(spec,args.container_discovery_observations)
