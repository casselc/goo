#!/usr/bin/env python3
"""Offline by default; --run sends uniquely scoped OTLP and reads existing tables.

No schema/configuration changes, resets, application API calls or synthetic
ClickHouse inserts. Raw retention verdicts are judged separately from this probe.
"""
import argparse
import copy
import hashlib
import json
import os
from pathlib import Path
import secrets
import time
import urllib.error
import urllib.parse

from probe_exponential_histogram_374 import attributes, docker, http, inspect_collector

ROOT = Path(__file__).resolve().parents[1]
PIN = 'bef0a5cfb993608bbcce1eb5859c37295ac8eed2'


def fixture(token, base):
    service = 'goo-exemplars377-' + token
    environment, namespace = 'probe377-' + token, 'readiness377-' + token
    trace_id, span_a, span_b = secrets.token_hex(16), secrets.token_hex(8), secrets.token_hex(8)
    resource = {'service.name': service, 'service.namespace': namespace,
                'service.instance.id': 'source-' + token,
                'deployment.environment.name': environment, 'probe.run': token}
    cases = [('fraction', {'asDouble': 0.125}, trace_id, span_a, 700123456),
             ('large-int64', {'asInt': '9007199254740993'}, trace_id, span_b, 100234567),
             ('integer-zero', {'asInt': '0'}, trace_id, span_a, 500345678),
             ('double-zero', {'asDouble': 0.0}, trace_id, span_b, 300456789),
             ('missing-ids', {'asDouble': 0.25}, None, None, 200567891),
             ('trace-only', {'asDouble': 0.5}, trace_id, None, 400678912)]
    exemplars = []
    for ordinal, (name, value, trace, span, offset) in enumerate(cases):
        ex = {'timeUnixNano': str(base + offset), **value,
              'filteredAttributes': [*attributes({'probe.exemplar': name, 'filtered.route': '/work-雪'}),
                                     {'key': 'typed.marker', 'value': {'intValue': str(ordinal)}}]}
        if trace is not None:
            ex['traceId'] = trace
        if span is not None:
            ex['spanId'] = span
        exemplars.append(ex)
    def point(name, start, end, counts, exemplars):
        return {'attributes': attributes({'probe.case': name, 'http.route': '/work-雪'}),
                'startTimeUnixNano': str(start), 'timeUnixNano': str(end),
                'count': str(sum(counts)), 'bucketCounts': list(map(str, counts)),
                'explicitBounds': [0.25, 1.0], 'exemplars': exemplars}
    primary = point('primary', base, base + 1000000000, [4, 1, 1], exemplars)
    empty = point('empty-exemplars', base + 1000000000, base + 2000000000, [1, 0, 0], [])
    decoy = point('identity-decoy', base, base + 1000000000, [0, 1, 0], [{
        'timeUnixNano': str(base + 888111222), 'asDouble': 0.75,
        'traceId': secrets.token_hex(16), 'spanId': secrets.token_hex(8),
        'filteredAttributes': attributes({'probe.exemplar': 'identity-decoy'})}])
    def resource_metrics(res, points, version, resource_schema, scope_schema):
        return {'resource': {'attributes': attributes(res)}, 'schemaUrl': resource_schema,
                'scopeMetrics': [{'scope': {'name': 'ordinary.exemplars377', 'version': version,
                                           'attributes': attributes({'scope.marker': version}),
                                           'droppedAttributesCount': 2}, 'schemaUrl': scope_schema,
                                  'metrics': [{'name': 'probe.request.duration', 'unit': 's',
                                               'histogram': {'aggregationTemporality': 1, 'dataPoints': points}}]}]}
    metric_data = {'resourceMetrics': [resource_metrics(resource, [primary, empty], '377.1',
        'https://opentelemetry.io/schemas/1.29.0', 'https://opentelemetry.io/schemas/1.30.0'),
        resource_metrics({**resource, 'service.instance.id': 'identity-decoy-' + token}, [decoy], '377.2',
        'https://opentelemetry.io/schemas/1.28.0', 'https://opentelemetry.io/schemas/1.31.0')]}
    def trace_resource(res, span, parent, name):
        wire = {'traceId': trace_id, 'spanId': span, 'name': name, 'kind': 2,
                'startTimeUnixNano': str(base), 'endTimeUnixNano': str(base + 1000000000),
                'attributes': attributes({'probe.trace': name}), 'status': {'code': 1}}
        if parent:
            wire['parentSpanId'] = parent
        return {'resource': {'attributes': attributes(res)}, 'schemaUrl': 'https://opentelemetry.io/schemas/1.29.0',
                'scopeSpans': [{'scope': {'name': 'ordinary.exemplars377.trace', 'version': '377.1'},
                               'spans': [wire]}]}
    peer = {**resource, 'service.name': service + '-peer', 'service.instance.id': 'peer-' + token}
    traces = {'resourceSpans': [trace_resource(resource, span_a, '', 'known-root'),
        trace_resource(peer, span_b, span_a, 'known-peer'),
        trace_resource({**peer, 'service.namespace': namespace + '-foreign'}, span_b, span_a, 'namespace-decoy'),
        trace_resource({**peer, 'deployment.environment.name': environment + '-foreign'}, span_b, span_a, 'environment-decoy')]}
    variants = []
    for name, changes, remove in [('malformed-trace-id', {'traceId': 'not-hex'}, []),
                                  ('short-span-id', {'spanId': '01'}, [])]:
        data = copy.deepcopy(metric_data)
        data['resourceMetrics'] = data['resourceMetrics'][:1]
        rm = data['resourceMetrics'][0]
        variant_service = service + '-' + name
        rm['resource']['attributes'] = attributes({**resource, 'service.name': variant_service})
        ex = copy.deepcopy(exemplars[0])
        ex.update(changes)
        for key in remove:
            ex.pop(key, None)
        ex['filteredAttributes'] = attributes({'probe.exemplar': name})
        rm['scopeMetrics'][0]['metrics'][0]['histogram']['dataPoints'] = [point(name, base, base + 1000000000, [1, 0, 0], [ex])]
        variants.append({'case': name, 'service': variant_service, 'input': data})
    return {'service': service, 'environment': environment, 'namespace': namespace,
            'traceID': trace_id, 'spanA': span_a, 'spanB': span_b, 'input': metric_data,
            'traceInput': traces, 'variants': variants}


def validate(plan):
    points = [p for r in plan['input']['resourceMetrics'] for s in r['scopeMetrics']
              for m in s['metrics'] for p in m['histogram']['dataPoints']]
    assert len(points) == 3 and len(points[0]['exemplars']) == 6 and points[1]['exemplars'] == []
    for p in points:
        assert int(p['count']) == sum(map(int, p['bucketCounts'])) and len(p['bucketCounts']) == len(p['explicitBounds']) + 1
    assert len(plan['traceInput']['resourceSpans']) == 4
    assert len({r['resource']['attributes'][0]['value']['stringValue'] for r in plan['traceInput']['resourceSpans']}) == 2


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--run', action='store_true')
    parser.add_argument('--out', default=str(ROOT / 'artifacts/histogram-exemplars377'))
    parser.add_argument('--collector-container', default='grammar-observability-collector-1')
    args = parser.parse_args()
    token, base = secrets.token_hex(6), time.time_ns() - 15000000000 + 123457
    plan = fixture(token, base)
    validate(plan)
    if not args.run:
        print(json.dumps({'status': 'prepared-offline', 'points': 3, 'primaryExemplars': 6,
                          'traceRows': 4, 'negativeRequests': len(plan['variants'])}))
        return
    # Read-only daemon inspection; no existing containers/volumes are modified.
    docker('info', '--format', '{{.ServerVersion}}')
    out = Path(args.out).resolve()
    out.mkdir(parents=True, exist_ok=True)
    ch_url = os.environ.get('GOO_TEST_CLICKHOUSE_URL', 'http://127.0.0.1:8123')
    report = {'version': 'goo.histogramExemplarReadiness377/1', 'token': token,
              'sourcePin': {'tag': 'v0.123.0', 'commit': PIN}, 'plan': plan,
              'collector': inspect_collector(args.collector_container), 'requests': [],
              'startedAt': time.strftime('%Y-%m-%dT%H:%M:%SZ', time.gmtime())}
    def query(sql, params=None):
        target = ch_url + ('&' if '?' in ch_url else '?') + urllib.parse.urlencode(params or {})
        _, raw = http(target, sql.encode(), clickhouse=True)
        return [json.loads(line) for line in raw.decode().splitlines() if line]
    def ingest(kind, data, name):
        wire = json.dumps(data, separators=(',', ':'), ensure_ascii=False).encode()
        (out / (name + '.json')).write_bytes(wire)
        try:
            status, body = http('http://127.0.0.1:4318/v1/' + kind, wire)
        except urllib.error.HTTPError as error:
            status, body = error.code, error.read()
        fact = {'case': name, 'status': status, 'response': body.decode(),
                'inputSHA256': hashlib.sha256(wire).hexdigest()}
        report['requests'].append(fact)
        return fact
    point_sql = "SELECT *,toString(toUnixTimestamp64Nano(StartTimeUnix)) AS startNS,toString(toUnixTimestamp64Nano(TimeUnix)) AS timeNS,toString(Count) AS pointCount,arrayMap(t->toString(toUnixTimestamp64Nano(t)),`Exemplars.TimeUnix`) AS exemplarTimes,arrayMap(v->toString(v),`Exemplars.Value`) AS exemplarValues,arrayMap(v->toString(reinterpretAsUInt64(v)),`Exemplars.Value`) AS exemplarValueBits FROM otel_metrics_histogram WHERE ServiceName={service:String} ORDER BY ScopeVersion,TimeUnix,Attributes['probe.case'] FORMAT JSONEachRow"
    trace_sql = "SELECT TraceId,SpanId,ParentSpanId,ServiceName,SpanName,ResourceAttributes,ScopeName,ScopeVersion,toString(toUnixTimestamp64Nano(Timestamp)) AS timestampNS,toString(Duration) AS durationNS FROM otel_traces WHERE TraceId={trace:String} AND ResourceAttributes['probe.run']={token:String} ORDER BY ServiceName,SpanName FORMAT JSONEachRow"
    try:
        report['schema'] = {'describe': query('DESCRIBE TABLE otel_metrics_histogram FORMAT JSONEachRow'),
                            'create': query('SHOW CREATE TABLE otel_metrics_histogram FORMAT JSONEachRow')}
        if query(point_sql, {'param_service': plan['service']}) or query(trace_sql, {'param_trace': plan['traceID'], 'param_token': token}):
            raise RuntimeError('Unique readiness source unexpectedly already exists')
        for kind, data, name in [('traces', plan['traceInput'], 'known-traces'), ('metrics', plan['input'], 'histogram')]:
            fact = ingest(kind, data, name)
            partial = json.loads(fact['response'] or '{}').get('partialSuccess', {})
            if fact['status'] != 200 or int(partial.get('rejectedDataPoints', partial.get('rejectedSpans', 0))):
                raise RuntimeError('Valid readiness request was not fully accepted')
        for variant in plan['variants']:
            ingest('metrics', variant['input'], variant['case'])
        deadline, polls = time.monotonic() + 45, 0
        while True:
            points = query(point_sql, {'param_service': plan['service']})
            traces = query(trace_sql, {'param_trace': plan['traceID'], 'param_token': token})
            polls += 1
            if len(points) == 3 and len(traces) == 4:
                break
            if time.monotonic() >= deadline:
                raise RuntimeError('Bounded exemplar/trace persistence wait expired')
            time.sleep(0.5)
        report.update({'rows': points, 'traceRows': traces, 'polls': polls})
        report['variantRows'] = {v['case']: query(point_sql, {'param_service': v['service']}) for v in plan['variants']}
        report['queries'] = {'points': point_sql, 'traces': trace_sql}
        report['status'] = 'observed-unjudged'
        print(json.dumps({'status': report['status'], 'points': len(points), 'traceRows': len(traces),
                          'requests': [(r['case'], r['status']) for r in report['requests']], 'output': str(out)}))
    except Exception as error:
        report.update({'status': 'probe-failed', 'error': str(error)})
        raise
    finally:
        report['finishedAt'] = time.strftime('%Y-%m-%dT%H:%M:%SZ', time.gmtime())
        (out / 'report.json').write_text(json.dumps(report, indent=2, ensure_ascii=False) + '\n')


if __name__ == '__main__':
    main()
