#!/usr/bin/env python3
"""Observe the existing collector/ClickHouse path without changing its schema.

Default operation is offline fixture validation. --run explicitly performs
read-only Docker inspection, one uniquely scoped OTLP write, and bounded reads.
This probe emits raw facts; independent retention verdicts belong to verification.
"""
import argparse
import base64
from fractions import Fraction
import hashlib
import json
import os
from pathlib import Path
import secrets
import subprocess
import time
import urllib.parse
import urllib.request

ROOT = Path(__file__).resolve().parents[1]
FIXTURE = ROOT / 'testdata/source/exponential-histogram-374.json'


def pow2(exponent):
    return Fraction(1 << exponent) if exponent >= 0 else Fraction(1, 1 << -exponent)


def index_of(value, scale):
    # 1 is the exact inclusive upper endpoint of index -1 at every scale.
    if value == 1:
        return -1
    if scale > 8:
        raise ValueError('Fixture oracle intentionally bounds nontrivial exponent work')
    for index in range(-64, 65):
        if scale >= 0:
            powered = value ** (1 << scale)
            lower, upper = pow2(index), pow2(index + 1)
        else:
            powered = value
            lower, upper = pow2(index * (1 << -scale)), pow2((index + 1) * (1 << -scale))
        if lower < powered <= upper:
            return index
    raise ValueError('Fixture observation outside independent oracle search')


def validate_fixture(fixture):
    """Derive bucket masses from compressed rational observations, not Goo code."""
    for point in fixture['cases']:
        expected = {'positive': {}, 'negative': {}}
        zero, total = 0, 0
        for observed in point['observations']:
            value, count = Fraction(observed['value']), int(observed['count'])
            if count < 0:
                raise ValueError('Negative observation mass')
            total += count
            if abs(value) <= Fraction(point['zeroThreshold']):
                zero += count
            else:
                side = 'positive' if value > 0 else 'negative'
                index = index_of(abs(value), point['scale'])
                expected[side][index] = expected[side].get(index, 0) + count
        if total != int(point['count']) or zero != int(point['zeroCount']):
            raise ValueError('Count mass mismatch: ' + point['case'])
        for side in ['positive', 'negative']:
            buckets = point[side]
            actual = {buckets['offset'] + i: int(value) for i, value in enumerate(buckets['bucketCounts']) if int(value)}
            if actual != expected[side]:
                raise ValueError('Independent bucket mass mismatch: ' + point['case'] + '/' + side)
        if any(Fraction(observed['value']) < 0 for observed in point['observations']) and 'sum' in point:
            raise ValueError('Do not emit Sum for signed-observation points')


def attributes(values):
    return [{'key': key, 'value': {'stringValue': value}} for key, value in values.items()]


def payload(fixture, token, base):
    points, metrics = [], {}
    for case in fixture['cases']:
        metric = 'probe374.' + token + '.' + case['metric']
        wire = {'startTimeUnixNano': str(base + case['startStep'] * 2000321000),
                'timeUnixNano': str(base + case['timeStep'] * 2000321000),
                'count': case['count'], 'scale': case['scale'], 'zeroCount': case['zeroCount'],
                'zeroThreshold': float(case['zeroThreshold']),
                'positive': case['positive'], 'negative': case['negative'], 'flags': case['flags'],
                'attributes': attributes({'probe.run': token, 'probe.series': case['series']})}
        for field in ['sum', 'min', 'max']:
            if field in case:
                wire[field] = float(case[field])
        group = metrics.setdefault((metric, case['temporality']), [])
        group.append(wire)
        points.append({'case': case['case'], 'metric': metric, 'series': case['series'],
                       'temporality': case['temporality'], 'startNS': wire['startTimeUnixNano'],
                       'timeNS': wire['timeUnixNano'], 'wire': wire, 'observations': case['observations']})
    resource = {'service.name': 'goo-exponential374-' + token, 'service.namespace': 'readiness374',
                'service.instance.id': 'instance-' + token, 'deployment.environment.name': 'probe374-' + token,
                'probe.run': token}
    data = {'resourceMetrics': [{'schemaUrl': 'https://opentelemetry.io/schemas/1.29.0',
        'resource': {'attributes': attributes(resource)}, 'scopeMetrics': [{
        'schemaUrl': 'https://opentelemetry.io/schemas/1.30.0',
        'scope': {'name': 'ordinary.exponential374', 'version': '374.1',
                  'attributes': attributes({'scope.marker': 'readiness374'}), 'droppedAttributesCount': 2},
        'metrics': [{'name': metric, 'unit': fixture['unit'], 'exponentialHistogram': {
                     'aggregationTemporality': temporality, 'dataPoints': values}}
                    for (metric, temporality), values in metrics.items()]}]}]}
    return data, points, resource['service.name']


def docker(*args):
    env = dict(os.environ)
    for key in ['DOCKER_HOST', 'DOCKER_CONTEXT', 'DOCKER_TLS', 'DOCKER_TLS_VERIFY', 'DOCKER_CERT_PATH']:
        env.pop(key, None)
    return subprocess.check_output(['docker', '--host=unix:///var/run/docker.sock', *args], env=env,
                                   timeout=20, stderr=subprocess.PIPE).decode().strip()


def inspect_collector(container):
    selected = json.loads(docker('inspect', '--format',
        '{"imageID":{{json .Image}},"imageReference":{{json .Config.Image}}}', container))
    selected.update({'container': container,
                     'repoDigests': json.loads(docker('image', 'inspect', '--format', '{{json .RepoDigests}}', selected['imageID'])),
                     'version': docker('exec', container, '/otelcol-contrib', '--version'),
                     'configSHA256': hashlib.sha256((ROOT / 'config/collector.yaml').read_bytes()).hexdigest()})
    return selected


def http(url, body, *, clickhouse=False):
    headers = {'Content-Type': 'text/plain' if clickhouse else 'application/json'}
    if clickhouse:
        user = os.environ.get('GOO_TEST_CLICKHOUSE_USER', 'lab')
        password = os.environ.get('GOO_TEST_CLICKHOUSE_PASSWORD', 'lab-local')
        headers['Authorization'] = 'Basic ' + base64.b64encode((user + ':' + password).encode()).decode()
    request = urllib.request.Request(url, data=body, headers=headers, method='POST')
    with urllib.request.urlopen(request, timeout=20) as response:
        raw = response.read()
        return response.status, raw


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--run', action='store_true')
    parser.add_argument('--out', default=str(ROOT / 'artifacts/exponential-histogram374'))
    parser.add_argument('--collector-container', default='grammar-observability-collector-1')
    args = parser.parse_args()
    fixture = json.loads(FIXTURE.read_text())
    validate_fixture(fixture)
    if not args.run:
        print(json.dumps({'status': 'prepared-offline', 'points': len(fixture['cases']),
                          'oracle': 'independent Fraction arithmetic over compressed observations'}))
        return
    out = Path(args.out).resolve()
    out.mkdir(parents=True, exist_ok=True)
    token, base = secrets.token_hex(6), time.time_ns() - 15000000000 + 123457
    data, points, service = payload(fixture, token, base)
    ch_url = os.environ.get('GOO_TEST_CLICKHOUSE_URL', 'http://127.0.0.1:8123')
    report = {'version': 'goo.exponentialHistogramReadiness374/1', 'token': token,
              'startedAt': time.strftime('%Y-%m-%dT%H:%M:%SZ', time.gmtime()),
              'collector': inspect_collector(args.collector_container),
              'sourcePin': {'tag': 'v0.123.0', 'commit': 'bef0a5cfb993608bbcce1eb5859c37295ac8eed2'},
              'input': data, 'points': points, 'requests': {}}
    def query(sql, parameters=None):
        encoded = urllib.parse.urlencode(parameters or {})
        url = ch_url + ('&' if '?' in ch_url else '?') + encoded if encoded else ch_url
        status, raw = http(url, sql.encode(), clickhouse=True)
        if status != 200:
            raise RuntimeError('ClickHouse read failed')
        return [json.loads(line) for line in raw.decode().splitlines() if line]
    try:
        report['schema'] = {'describe': query('DESCRIBE TABLE otel_metrics_exp_histogram FORMAT JSONEachRow'),
                            'create': query('SHOW CREATE TABLE otel_metrics_exp_histogram FORMAT JSONEachRow')}
        sql = "SELECT *,toString(toUnixTimestamp64Nano(StartTimeUnix)) AS startNS,toString(toUnixTimestamp64Nano(TimeUnix)) AS timeNS,toString(Count) AS count,toString(ZeroCount) AS zeroCount,arrayMap(c->toString(c),PositiveBucketCounts) AS positiveBucketCounts,arrayMap(c->toString(c),NegativeBucketCounts) AS negativeBucketCounts FROM otel_metrics_exp_histogram WHERE ServiceName={service:String} ORDER BY MetricName,TimeUnix,Attributes['probe.series'] FORMAT JSONEachRow"
        parameters = {'param_service': service}
        if query(sql, parameters):
            raise RuntimeError('Unique probe service unexpectedly already exists')
        wire_bytes = json.dumps(data, separators=(',', ':')).encode()
        (out / 'otlp.json').write_bytes(wire_bytes)
        report['inputSHA256'] = hashlib.sha256(wire_bytes).hexdigest()
        status, body = http('http://127.0.0.1:4318/v1/metrics', wire_bytes)
        report['requests']['ingest'] = {'status': status, 'body': body.decode()}
        partial = json.loads(body or b'{}').get('partialSuccess', {})
        if status != 200 or int(partial.get('rejectedDataPoints', 0)):
            raise RuntimeError('OTLP receiver did not accept every point')
        deadline = time.monotonic() + 45
        polls = 0
        while True:
            rows = query(sql, parameters)
            polls += 1
            if len(rows) == len(points):
                break
            if time.monotonic() >= deadline:
                raise RuntimeError('Bounded exporter persistence wait expired')
            time.sleep(0.5)
        report['rows'] = rows
        report['requests'].update({'query': sql, 'parameters': parameters, 'polls': polls})
        report['status'] = 'observed-unjudged'
        print(json.dumps({'status': report['status'], 'points': len(points), 'rows': len(rows),
                          'collectorVersion': report['collector']['version'], 'output': str(out)}))
    except Exception as error:
        report['status'], report['error'] = 'probe-failed', str(error)
        raise
    finally:
        report['finishedAt'] = time.strftime('%Y-%m-%dT%H:%M:%SZ', time.gmtime())
        (out / 'report.json').write_text(json.dumps(report, indent=2) + '\n')


if __name__ == '__main__':
    main()
