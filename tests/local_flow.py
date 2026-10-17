"""Docker-free application integration: real HTTP/gRPC and OTLP, fake Redis.

Not a substitute for scripts/smoke.py against the full Compose stack.
Run from project root: python tests/local_flow.py
"""
import json
import os
from pathlib import Path
import subprocess
import sys
import threading
import time
import urllib.request
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer
from fakeredis import TcpFakeServer
from opentelemetry.proto.collector.logs.v1.logs_service_pb2 import ExportLogsServiceRequest
from opentelemetry.proto.collector.trace.v1.trace_service_pb2 import ExportTraceServiceRequest
from opentelemetry.proto.collector.metrics.v1.metrics_service_pb2 import ExportMetricsServiceRequest

ROOT = Path(__file__).resolve().parents[1]
received = {'logs': [], 'traces': [], 'metrics': []}


class Collector(BaseHTTPRequestHandler):
    def do_POST(self):
        kind = self.path.rsplit('/', 1)[-1]
        cls = {'logs': ExportLogsServiceRequest, 'traces': ExportTraceServiceRequest,
               'metrics': ExportMetricsServiceRequest}[kind]
        msg = cls.FromString(self.rfile.read(int(self.headers['Content-Length'])))
        received[kind].append(msg)
        self.send_response(200); self.send_header('Content-Type', 'application/x-protobuf'); self.end_headers()

    def log_message(self, *_): pass


class QuietRedis(TcpFakeServer):
    def handle_error(self, request, client_address):
        # Fake server reports peer resets while child processes terminate.
        if isinstance(sys.exc_info()[1], (BrokenPipeError, ConnectionResetError)): return
        super().handle_error(request, client_address)


redis = QuietRedis(('127.0.0.1', 0))
collector = ThreadingHTTPServer(('127.0.0.1', 0), Collector)
for server in (redis, collector):
    threading.Thread(target=server.serve_forever, daemon=True).start()
env = {**os.environ, 'REDIS_URL': f'redis://127.0.0.1:{redis.server_address[1]}/0',
       'OTEL_EXPORTER_OTLP_ENDPOINT': f'http://127.0.0.1:{collector.server_address[1]}',
       'GRPC_TARGET': '127.0.0.1:50051'}
processes = []
try:
    for role in ('grpc', 'consumer', 'web'):
        processes.append(subprocess.Popen([sys.executable, 'sample.py', role], cwd=ROOT, env=env,
                                         stdout=subprocess.DEVNULL, stderr=subprocess.PIPE))
    for attempt in range(50):
        try:
            req = urllib.request.Request('http://127.0.0.1:8081/reserve', data=b'{"scenario":"retry"}',
                                         headers={'Content-Type': 'application/json'})
            with urllib.request.urlopen(req, timeout=3) as response: created = json.load(response)
            break
        except Exception as exc:
            if attempt == 49 and hasattr(exc, 'read'): print(exc.read().decode(), file=sys.stderr)
            if attempt == 49: raise
            time.sleep(.2)
    deadline = time.monotonic() + 20
    while time.monotonic() < deadline:
        logs = [r for msg in received['logs'] for resource in msg.resource_logs
                for scope in resource.scope_logs for r in scope.log_records]
        spans = [s for msg in received['traces'] for resource in msg.resource_spans
                 for scope in resource.scope_spans for s in scope.spans]
        if (any(r.body.string_value == 'reservation.terminal' for r in logs)
                and received['metrics'] and any(s.links for s in spans)
                and any(s.kind == 2 and s.name == 'ReservationService.Create' for s in spans)): break
        time.sleep(.25)
    assert any(r.body.string_value == 'reservation.accepted' for r in logs), 'missing acceptance log'
    assert any(r.body.string_value == 'reservation.retry' for r in logs), 'missing retry log'
    assert any(r.body.string_value == 'reservation.terminal' for r in logs), 'missing terminal log'
    assert any(s.kind == 2 and s.name == 'ReservationService.Create' for s in spans), 'missing gRPC server span: ' + repr([(s.name, s.kind) for s in spans])
    assert any(s.links for s in spans), 'missing async span link'
    assert received['metrics'], 'missing OTLP metrics'
    print('PASS: real HTTP -> real gRPC -> fake Redis Streams -> consumer -> protobuf OTLP logs/traces/metrics')
    print('Reservation', created['id'], '|', len(logs), 'logs |', len(spans), 'spans')
finally:
    for process in processes: process.terminate()
    for process in processes:
        try: _, errors = process.communicate(timeout=5)
        except subprocess.TimeoutExpired: process.kill(); _, errors = process.communicate()
        if sys.exc_info()[0] is not None or process.returncode not in (-15, 0): print(errors.decode()[-2000:], file=sys.stderr)
    redis.shutdown(); collector.shutdown()
