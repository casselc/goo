"""HTTP -> real gRPC (JSON codec) -> Redis Streams -> consumer, with OTel."""
import concurrent.futures
import json
import os
import sys
import time
import uuid
import urllib.request
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer
import grpc
import redis
from opentelemetry import propagate
from opentelemetry.trace import SpanKind, Status, StatusCode, Link
from telemetry import setup, emit
from queue_telemetry import WorkerQueue, CAPACITY, ATTRS, wait_ns

METHOD = '/lab.Reservations/Create'
SCENARIOS = ('good', 'good', 'good', 'retry', 'late', 'missing', 'reject')


def dumps(x):
    return json.dumps(x).encode()


def run(role):
    tracer, log, meter, providers = setup('lab-' + role)
    count = meter.create_counter('lab.operations', unit='{operation}')
    duration = meter.create_histogram('lab.operation.duration', unit='s')
    queue = redis.Redis.from_url(os.getenv('REDIS_URL', 'redis://localhost:6379/0'), decode_responses=True, socket_timeout=3, socket_connect_timeout=3)

    def create(item, context):
        parent = propagate.extract(dict(context.invocation_metadata()))
        with tracer.start_as_current_span('ReservationService.Create', context=parent, kind=SpanKind.SERVER) as span:
            started_ns = time.time_ns()
            started = started_ns / 1e9
            item.update(id=str(uuid.uuid4()), accepted_at=started, accepted_at_ns=str(started_ns))
            span.set_attribute('rpc.system', 'grpc')
            span.set_attribute('rpc.service', 'lab.Reservations')
            span.set_attribute('rpc.method', 'Create')
            span.set_attribute('app.reservation.id', item['id'])
            span.set_attributes({'app.release': item['release'], 'app.region': item['region'], 'app.scenario': item['scenario']})
            with tracer.start_as_current_span('ReservationAccepted publish', kind=SpanKind.PRODUCER) as producer:
                # The async link targets this span, so include the same business
                # identity as the acceptance and consumer for subject drilldown.
                producer.set_attributes({'app.reservation.id': item['id'], 'app.tenant.id': 'demo',
                                         'app.release': item['release'], 'app.region': item['region'],
                                         'app.scenario': item['scenario'], **ATTRS, 'messaging.operation.type': 'send'})
                carrier = {}
                propagate.inject(carrier)
                item['carrier'] = carrier
                queue.xadd('reservations', {'payload': json.dumps(item)}, maxlen=20000)
            emit(log, 'reservation.accepted', item, **{'app.accepted_at': started, 'goo.event.time_ns': str(started_ns)})
            count.add(1, {'operation': 'accept', 'release': item['release']})
            duration.record(time.time() - started, {'operation': 'grpc'})
            return {'id': item['id'], 'scenario': item['scenario']}

    if role == 'grpc':
        server = grpc.server(concurrent.futures.ThreadPoolExecutor(max_workers=8))
        server.add_generic_rpc_handlers((grpc.method_handlers_generic_handler('lab.Reservations', {
            'Create': grpc.unary_unary_rpc_method_handler(create, request_deserializer=json.loads, response_serializer=dumps)}),))
        server.add_insecure_port('[::]:50051')
        server.start()
        server.wait_for_termination()

    elif role == 'web':
        from opentelemetry import trace
        tracer = trace.get_tracer('goo.http-sample', '0.1.0', schema_url='https://opentelemetry.io/schemas/1.37.0')
        channel = grpc.insecure_channel(os.getenv('GRPC_TARGET', 'localhost:50051'))
        call = channel.unary_unary(METHOD, request_serializer=dumps, response_deserializer=json.loads)
        class Handler(BaseHTTPRequestHandler):
            def do_GET(self):
                if self.path == '/health':
                    self.send_response(200); self.end_headers(); self.wfile.write(b'web ready'); return
                if self.path == '/demo-failure':
                    with tracer.start_as_current_span('GET /demo-failure', kind=SpanKind.SERVER) as span:
                        span.set_attributes({'http.request.method': 'GET', 'http.route': '/demo-failure', 'http.response.status_code': 503})
                        span.set_status(Status(StatusCode.ERROR, 'controlled sample HTTP failure'))
                        self.send_response(503); self.end_headers(); self.wfile.write(b'controlled sample HTTP failure')
                    return
                if self.path != '/': self.send_error(404); return
                self.send_response(200); self.send_header('Content-Type', 'text/html; charset=utf-8'); self.end_headers()
                self.wfile.write(b'''<!doctype html><html lang="en"><meta charset="utf-8"><title>Reservation sample</title>
                <body style="font:18px system-ui;max-width:700px;margin:60px auto;padding:20px"><h1>Reservation sample</h1>
                <p>HTTP request to a gRPC service, followed by asynchronous processing.</p>
                <select id="scenario"><option>good</option><option>retry</option><option>late</option><option>missing</option><option>reject</option></select>
                <button id="submit">Create reservation</button><pre id="result"></pre><a href="http://localhost:8088">Open investigation UI</a>
                <script>document.getElementById('submit').onclick=async()=>{try{let r=await fetch('/reserve',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({scenario:document.getElementById('scenario').value})});document.getElementById('result').textContent=await r.text()}catch(e){document.getElementById('result').textContent=e.message}};</script></body></html>''')
            def do_POST(self):
                if self.path != '/reserve':
                    self.send_error(404); return
                try:
                    size = int(self.headers.get('Content-Length', '0'))
                    if size > 4096:
                        self.send_error(413); return
                    item = json.loads(self.rfile.read(size) or '{}')
                    scenario = item.get('scenario', 'good')
                    if scenario not in SCENARIOS:
                        raise ValueError('unknown scenario')
                    item = {'scenario': scenario, 'tenant': 'demo',
                            'release': 'v2' if scenario in ('late', 'missing', 'reject') else 'v1',
                            'region': 'west' if scenario in ('late', 'missing') else 'east'}
                    with tracer.start_as_current_span('POST /reserve', kind=SpanKind.SERVER) as http_span:
                        http_span.set_attributes({'http.request.method': 'POST', 'http.route': '/reserve'})
                        with tracer.start_as_current_span('ReservationService.Create', kind=SpanKind.CLIENT):
                            carrier = {}; propagate.inject(carrier)
                            result = call(item, metadata=list(carrier.items()), timeout=5)
                        http_span.set_attribute('http.response.status_code', 202)
                    data = dumps(result)
                    self.send_response(202); self.send_header('Content-Type', 'application/json')
                    self.end_headers(); self.wfile.write(data)
                except (ValueError, grpc.RpcError) as exc:
                    self.send_error(400 if isinstance(exc, ValueError) else 503, str(exc))
            def log_message(self, *_): pass
        ThreadingHTTPServer(('0.0.0.0', 8081), Handler).serve_forever()

    elif role == 'consumer':
        try: queue.xgroup_create('reservations', 'inventory', id='0', mkstream=True)
        except redis.ResponseError as exc:
            if 'BUSYGROUP' not in str(exc): raise

        from opentelemetry import metrics
        worker_meter = metrics.get_meter('goo.worker-queue', '0.1.0', schema_url='goo://demo/queue/v1')
        worker_queue = WorkerQueue(queue)
        worker_queue.register(worker_meter)
        queue_wait = worker_meter.create_histogram('lab.queue.wait', unit='s')

        def process(message_id, item):
            with worker_queue.working():
                process_work(message_id, item)

        def process_work(message_id, item):
            waiting = wait_ns(message_id)
            if waiting is not None:
                queue_wait.record(waiting / 1e9, ATTRS)
            emit(log, 'reservation.processing_started', item,
                 **{'lab.queue.wait_ns': str(waiting) if waiting is not None else '',
                    'lab.queue.clock_violation': waiting is None})
            # New trace with a producer link exercises asynchronous correlation.
            parent = propagate.extract(item['carrier'])
            from opentelemetry import trace, context
            linked = trace.get_current_span(parent).get_span_context()
            for attempt in range(2 if item['scenario'] == 'retry' else 1):
                with tracer.start_as_current_span('inventory process', context=context.Context(),
                        links=[Link(linked)], kind=SpanKind.CONSUMER) as span:
                    span.set_attributes({'app.reservation.id': item['id'], 'app.attempt': attempt + 1,
                                         'app.release': item['release'], 'app.region': item['region'],
                                         'app.scenario': item['scenario'], 'messaging.system': 'redis',
                                         'messaging.operation.type': 'process', **ATTRS})
                    if attempt == 0:
                        span.set_attribute('lab.queue.clock_violation', waiting is None)
                        if waiting is not None:
                            span.set_attribute('lab.queue.wait_ns', str(waiting))
                    count.add(1, {'operation': 'process', 'scenario': item['scenario']})
                    if item['scenario'] == 'retry' and attempt == 0:
                        emit(log, 'reservation.retry', item)
                        span.set_status(Status(StatusCode.ERROR, 'simulated transient error'))
                        time.sleep(.2); continue
                    if item['scenario'] == 'missing':
                        emit(log, 'reservation.stalled', item)
                        break
                    time.sleep(14 if item['scenario'] == 'late' else .3)
                    outcome = 'rejected' if item['scenario'] == 'reject' else 'confirmed'
                    completed_ns = time.time_ns()
                    completed = completed_ns / 1e9
                    emit(log, 'reservation.terminal', item, **{'app.completed_at': completed, 'app.outcome': outcome, 'goo.event.time_ns': str(completed_ns)})
                    duration.record(completed - item['accepted_at'], {'operation': 'completion', 'outcome': outcome})
            # Demo acknowledges missing work deliberately; no production retry guarantee.
            queue.xack('reservations', 'inventory', message_id)

        with concurrent.futures.ThreadPoolExecutor(max_workers=CAPACITY) as pool:
            while True:
                # Bounded in-flight work; no unbounded executor queue.
                results = queue.xreadgroup('inventory', 'worker', {'reservations': '>'}, count=CAPACITY, block=1000)
                futures = [pool.submit(process, mid, json.loads(fields['payload']))
                           for _, messages in results for mid, fields in messages]
                for future in futures: future.result()

    elif role == 'traffic':
        index = 0
        while True:
            try:
                body = dumps({'scenario': SCENARIOS[index % len(SCENARIOS)]})
                request = urllib.request.Request('http://web:8081/reserve', data=body, headers={'Content-Type': 'application/json'})
                with urllib.request.urlopen(request, timeout=8) as response: response.read()
                index += 1
            except Exception as exc: print('Waiting for web:', exc, flush=True)
            time.sleep(2)
    else:
        raise ValueError('unknown role')


if __name__ == '__main__': run(sys.argv[1])
