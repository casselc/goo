"""Exercise sample -> OTLP -> ClickHouse -> single Go UI/API runtime.

This checks observed correlated telemetry, not durable business completion or an
SLO. The sample workload is independent of the Goo runtime.
"""
import json
import time
import urllib.request


def request(url, body=None):
    req = urllib.request.Request(url, data=json.dumps(body).encode() if body is not None else None,
                                 headers={'Content-Type': 'application/json'})
    with urllib.request.urlopen(req, timeout=20) as response:
        return json.load(response)


def main():
    deadline = time.monotonic() + 180
    last = None
    while time.monotonic() < deadline:
        try:
            created = request('http://localhost:8081/reserve', {'scenario': 'good'})
            break
        except Exception as exc:
            last = exc
            time.sleep(2)
    else:
        raise SystemExit('Sample service did not become ready: ' + str(last))

    assert request('http://localhost:8088/health') == {'ok': True}
    config = request('http://localhost:8088/api/ui-config')
    assert config['version'] == 'goo/ui-config/v1'
    # Both signals must include the newly created subject, not just old traffic.
    where = {'field': {'kind': 'attribute', 'key': 'app.reservation.id'},
             'op': 'eq', 'value': created['id']}
    for signal in ('logs', 'spans'):
        while time.monotonic() < deadline:
            try:
                now = int(time.time() * 1000)
                data = request('http://localhost:8088/api/explore', {
                    'signal': signal, 'environment': 'local', 'fromMs': now - 300000,
                    'toMs': now, 'where': where, 'limit': 100, 'offset': 0})
                if data['rows']:
                    print(signal, len(data['rows']), 'correlated observed rows')
                    break
            except Exception as exc:
                last = exc
            time.sleep(2)
        else:
            raise SystemExit('No correlated ' + signal + ' telemetry: ' + str(last))
    print('PASS: sample HTTP -> gRPC -> Redis -> OTLP -> ClickHouse -> Go API')


if __name__ == '__main__':
    main()
