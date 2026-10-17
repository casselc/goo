"""Smoke-test the running Compose stack, including collector ingestion and queries."""
import json
import time
import urllib.request


def request(url, body=None):
    req = urllib.request.Request(url, data=json.dumps(body).encode() if body is not None else None,
                                 headers={'Content-Type': 'application/json'})
    with urllib.request.urlopen(req, timeout=20) as response: return json.load(response)


deadline = time.monotonic() + 180
last = None
while time.monotonic() < deadline:
    try:
        created = request('http://localhost:8081/reserve', {'scenario': 'good'})
        break
    except Exception as exc:
        last = exc; time.sleep(2)
else: raise SystemExit('Sample service did not become ready: ' + str(last))

while time.monotonic() < deadline:
    try:
        data = request('http://localhost:8088/api/query', {'view': 'completion', 'reservationId': created['id']})
        if data['rows'] and data['rows'][0]['classification'] == 'good': break
    except Exception as exc: last = exc
    time.sleep(2)
else: raise SystemExit('No good completion found: ' + str(last))

for view in ('logs', 'traces', 'latency', 'metrics'):
    data = request('http://localhost:8088/api/query', {'view': view})
    assert data['rows'], f'{view}: expected ingested telemetry'
    print(view, len(data['rows']), 'rows')
print('PASS: HTTP -> gRPC -> Redis -> consumer -> OTLP -> ClickHouse -> query API')
