"""Current browser proxy wire behavior, using actual HTTP on both sides."""
import http.client
import json
import os
import re
from urllib.parse import urlsplit
from pathlib import Path
import socket
import threading
import unittest
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer
from unittest.mock import patch

from server import Handler


class Upstream(BaseHTTPRequestHandler):
    calls = []
    mode = 'error'
    ended = threading.Event()

    def handle_request(self):
        body = self.rfile.read(int(self.headers.get('Content-Length', 0)))
        type(self).calls.append((self.command, self.path, dict(self.headers), body))
        if self.path == '/tail' and self.mode == 'stream':
            self.send_response(200)
            self.send_header('Content-Type', 'text/event-stream')
            self.end_headers()
            self.wfile.write(b'id: 1\nevent: reset\ndata: {"windowMs":60000,"selector":{"environment":"prod"},"continuity":"unknown","replay":false}\n\n')
            self.wfile.flush()
            self.connection.settimeout(4)
            try:
                if self.connection.recv(1) == b'':
                    type(self).ended.set()
            except (ConnectionResetError, BrokenPipeError):
                type(self).ended.set()
            except socket.timeout:
                pass
            return
        code = 200 if self.mode == 'cancel' else 429 if self.mode == 'busy' else 400
        payload = b'{"state":"cancellation-requested"}' if self.mode == 'cancel' else b'{"error":"backend refusal"}'
        if self.mode == 'malformed':
            payload = b'plain backend failure'
        self.send_response(code)
        self.send_header('Content-Type', 'text/plain' if self.mode in ('text', 'malformed') else 'application/json')
        self.send_header('Retry-After', '7')
        self.end_headers()
        self.wfile.write(payload)

    do_POST = do_GET = handle_request

    def log_message(self, *_):
        pass


class ProxyContracts(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        cls.observations = []
        cls.upstream = ThreadingHTTPServer(('127.0.0.1', 0), Upstream)
        cls.proxy = ThreadingHTTPServer(('127.0.0.1', 0), Handler)
        for server in (cls.upstream, cls.proxy):
            threading.Thread(target=server.serve_forever, daemon=True).start()
        cls.config = patch.dict(os.environ, {'PACKAGES_URL': f'http://127.0.0.1:{cls.upstream.server_port}'})
        cls.config.start()

    @classmethod
    def tearDownClass(cls):
        output = os.getenv('GOO_PROXY_OBSERVATIONS')
        if output:
            Path(output).parent.mkdir(parents=True, exist_ok=True)
            Path(output).write_text(json.dumps(cls.observations, indent=2))
        cls.config.stop()
        for server in (cls.proxy, cls.upstream):
            server.shutdown()
            server.server_close()

    def setUp(self):
        Upstream.calls = []
        Upstream.mode = 'error'
        Upstream.ended.clear()

    def request(self, method, path, body=None, headers=None):
        conn = http.client.HTTPConnection('127.0.0.1', self.proxy.server_port, timeout=5)
        try:
            try:
                conn.request(method, path, body=body, headers=headers or {})
            except BrokenPipeError:
                # The size guard may reject from Content-Length and close while
                # a large body is still being sent. Its HTTP refusal is evidence.
                pass
            response = conn.getresponse()
            data = response.read()
            spec = json.loads(Path('api/openapi.json').read_text())
            route = urlsplit(path).path.removeprefix('/api')
            pattern = next(p for p in spec['paths'] if re.fullmatch(re.sub(r'\{[^}]+\}', '[^/]+', p), route))
            entry = {'method': method, 'path': pattern, 'status': response.status, 'response': json.loads(data)}
            if body:
                entry['request'] = json.loads(body)
            type(self).observations.append(entry)
            return response.status, dict(response.getheaders()), data
        finally:
            conn.close()

    def test_exact_and_over_browser_limits(self):
        spec = json.loads(Path('api/openapi.json').read_text())
        for path in ('/dashboards/plan', '/metric-discovery', '/annotations', '/deployments', '/compare', '/measurements', '/trace-search'):
            limit = spec['paths'][path]['post']['x-browser-body-limit-bytes']
            with self.subTest(path=path):
                before = len(Upstream.calls)
                status, _, body = self.request('POST', '/api' + path, b'{}' + b' ' * (limit-2))
                self.assertEqual(status, 400)
                self.assertEqual(json.loads(body)['error'], 'backend refusal')
                self.assertEqual(len(Upstream.calls), before+1)
                status, headers, body = self.request('POST', '/api' + path, b'{}' + b' ' * (limit-1))
                self.assertEqual(status, 400)
                self.assertEqual(json.loads(body)['error'], 'JSON request exceeds size limit')
                self.assertEqual(headers['Cache-Control'], 'no-store')
                self.assertEqual(len(Upstream.calls), before+1)
        status, _, _ = self.request('POST', '/api/explore', b'')
        self.assertEqual(status, 400)

    def test_query_capability_forwarded_but_retry_header_not_forwarded(self):
        Upstream.mode = 'cancel'
        query_id = '01234567-1234-1234-1234-012345678901'
        status, _, body = self.request('POST', '/api/queries/'+query_id+'/cancel', b'{}', {'X-Goo-Query-ID': query_id})
        self.assertEqual(status, 200)
        self.assertEqual(json.loads(body), {'state':'cancellation-requested'})
        self.assertEqual(Upstream.calls[-1][1], '/queries/'+query_id+'/cancel')
        self.assertEqual(dict((k.lower(),v) for k,v in Upstream.calls[-1][2].items())['x-goo-query-id'], query_id)
        Upstream.mode = 'busy'
        status, headers, body = self.request('POST', '/api/measurements', b'{}', {'X-Goo-Query-ID':query_id})
        self.assertEqual(status, 429)
        self.assertNotIn('Retry-After', headers)
        self.assertEqual(json.loads(body)['error'], 'backend refusal')

    def test_counted_values_proxy_route_limits_and_cancellation_header(self):
        query_id = '01234567-1234-1234-1234-012345678901'
        payload = b'{"signal":"logs","environment":"prod","fromMs":1000,"toMs":61000,"limit":10,"groupBy":[{"kind":"column","key":"severity"}]}'
        status, headers, body = self.request('POST', '/api/explore/counted-values', payload, {'X-Goo-Query-ID':query_id})
        self.assertEqual(status, 400)
        self.assertEqual(json.loads(body), {'error':'backend refusal'})
        self.assertEqual(Upstream.calls[-1][1], '/explore/counted-values')
        self.assertEqual(Upstream.calls[-1][3], payload)
        self.assertEqual({k.lower():v for k,v in Upstream.calls[-1][2].items()}['x-goo-query-id'], query_id)
        accepted = payload + b' ' * (16384-len(payload))
        self.assertEqual(self.request('POST','/api/explore/counted-values',accepted)[0],400)
        self.assertEqual(Upstream.calls[-1][3],accepted)
        count = len(Upstream.calls)
        self.assertEqual(self.request('POST','/api/explore/counted-values',accepted+b' ')[0],400)
        self.assertEqual(len(Upstream.calls),count)

    def test_record_aggregates_proxy_route_limits_and_cancellation_header(self):
        query_id = '01234567-1234-1234-1234-012345678901'
        payload = b'{"signal":"spans","environment":"prod","fromMs":1000,"toMs":61000,"limit":100,"groupBy":[{"kind":"column","key":"service"}],"summary":"span-duration","spanStatus":true,"bucketMs":10000}'
        status, headers, body = self.request('POST', '/api/record-aggregates', payload, {'X-Goo-Query-ID':query_id})
        self.assertEqual(status, 400)
        self.assertEqual(json.loads(body), {'error':'backend refusal'})
        self.assertEqual(Upstream.calls[-1][1], '/record-aggregates')
        self.assertEqual(Upstream.calls[-1][3], payload)
        self.assertEqual({k.lower():v for k,v in Upstream.calls[-1][2].items()}['x-goo-query-id'], query_id)
        accepted = payload + b' ' * (16384-len(payload))
        self.assertEqual(self.request('POST','/api/record-aggregates',accepted)[0],400)
        self.assertEqual(Upstream.calls[-1][3],accepted)
        count = len(Upstream.calls)
        self.assertEqual(self.request('POST','/api/record-aggregates',accepted+b' ')[0],400)
        self.assertEqual(len(Upstream.calls),count)

    def test_pool_signals_proxy_exact_64k_limit(self):
        payload = b'{}' + b' ' * (65536-2)
        status, _, body = self.request('POST','/api/pool-signals/plan',payload)
        self.assertEqual(status,400)
        self.assertEqual(json.loads(body),{'error':'backend refusal'})
        self.assertEqual(Upstream.calls[-1][1],'/pool-signals/plan')
        self.assertEqual(Upstream.calls[-1][3],payload)
        before=len(Upstream.calls)
        status, _, body = self.request('POST','/api/pool-signals/plan',payload+b' ')
        self.assertEqual(status,400)
        self.assertEqual(len(Upstream.calls),before)

    def test_pool_view_proxy_exact_2mib_limit(self):
        payload = b'{}' + b' ' * (2097152-2)
        status, _, body = self.request('POST','/api/pool-signals/view',payload)
        self.assertEqual(status,400)
        self.assertEqual(json.loads(body),{'error':'backend refusal'})
        self.assertEqual(Upstream.calls[-1][1],'/pool-signals/view')
        self.assertEqual(Upstream.calls[-1][3],payload)
        before=len(Upstream.calls)
        status, _, body = self.request('POST','/api/pool-signals/view',payload+b' ')
        self.assertEqual(status,400)
        self.assertEqual(len(Upstream.calls),before)

    def test_container_memory_view_proxy_exact_8mib_limit(self):
        payload = b'{}' + b' ' * (8388608-2)
        status, _, body = self.request('POST','/api/container-memory/view',payload)
        self.assertEqual(status,400)
        self.assertEqual(json.loads(body),{'error':'backend refusal'})
        self.assertEqual(Upstream.calls[-1][1],'/container-memory/view')
        self.assertEqual(Upstream.calls[-1][3],payload)
        before=len(Upstream.calls)
        status, _, body = self.request('POST','/api/container-memory/view',payload+b' ')
        self.assertEqual(status,400)
        self.assertEqual(len(Upstream.calls),before)

    def test_container_cpu_view_proxy_exact_8mib_limit(self):
        payload = b'{}' + b' ' * (8388608-2)
        status, _, body = self.request('POST','/api/container-cpu/view',payload)
        self.assertEqual(status,400)
        self.assertEqual(json.loads(body),{'error':'backend refusal'})
        self.assertEqual(Upstream.calls[-1][1],'/container-cpu/view')
        self.assertEqual(Upstream.calls[-1][3],payload)
        before=len(Upstream.calls)
        status, _, body = self.request('POST','/api/container-cpu/view',payload+b' ')
        self.assertEqual(status,400)
        self.assertEqual(len(Upstream.calls),before)

    def test_pool_view_asset_is_served_without_upstream(self):
        conn=http.client.HTTPConnection('127.0.0.1',self.proxy.server_port,timeout=5)
        try:
            conn.request('GET','/pool-view.js')
            response=conn.getresponse()
            self.assertEqual(response.status,200)
            self.assertEqual(response.read(),Path('ui/pool-view.js').read_bytes())
            self.assertEqual(Upstream.calls,[])
        finally:
            conn.close()

    def test_host_routes_transport_and_assets(self):
        for path,limit in (('/api/host-discovery',16384),('/api/host-dashboard/plan',65536)):
            payload=b'{}'+b' '*(limit-2)
            status,_,body=self.request('POST',path,payload)
            self.assertEqual(status,400)
            self.assertEqual(json.loads(body),{'error':'backend refusal'})
            self.assertEqual(Upstream.calls[-1][1],path.removeprefix('/api'))
            self.assertEqual(Upstream.calls[-1][3],payload)
            before=len(Upstream.calls)
            self.assertEqual(self.request('POST',path,payload+b' ')[0],400)
            self.assertEqual(len(Upstream.calls),before)
        for asset in ('host-browser.js','host-dashboard.js','host.css'):
            conn=http.client.HTTPConnection('127.0.0.1',self.proxy.server_port,timeout=5)
            try:
                conn.request('GET','/'+asset);response=conn.getresponse()
                self.assertEqual(response.status,200)
                self.assertEqual(response.read(),Path('ui/'+asset).read_bytes())
            finally:
                conn.close()

    def test_runtime_observations_asset_is_served_without_upstream(self):
        conn = http.client.HTTPConnection('127.0.0.1', self.proxy.server_port, timeout=5)
        try:
            conn.request('GET', '/runtime-observations.js')
            response = conn.getresponse()
            self.assertEqual(response.status, 200)
            self.assertEqual(response.read(), Path('ui/runtime-observations.js').read_bytes())
            self.assertEqual(Upstream.calls, [])
        finally:
            conn.close()

    def test_pod_discovery_limits_query_id_and_assets(self):
        query_id = '01234567-1234-1234-1234-012345678901'
        payload = b'{}' + b' ' * (16384-2)
        status, _, body = self.request('POST', '/api/pod-discovery', payload, {'X-Goo-Query-ID': query_id})
        self.assertEqual(status, 400)
        self.assertEqual(json.loads(body), {'error': 'backend refusal'})
        self.assertEqual(Upstream.calls[-1][1], '/pod-discovery')
        self.assertEqual(Upstream.calls[-1][3], payload)
        self.assertEqual({k.lower(): v for k, v in Upstream.calls[-1][2].items()}['x-goo-query-id'], query_id)
        before = len(Upstream.calls)
        self.assertEqual(self.request('POST', '/api/pod-discovery', payload+b' ')[0], 400)
        self.assertEqual(len(Upstream.calls), before)
        for asset in ('container-memory.js', 'container-memory-view.js', 'container-cpu.js', 'container-cpu-view.js', 'kubernetes-context.js', 'kubernetes-browser.js', 'resource-scope.js', 'resource-tabs.js', 'kubernetes.css'):
            conn = http.client.HTTPConnection('127.0.0.1', self.proxy.server_port, timeout=5)
            try:
                conn.request('GET', '/'+asset)
                response = conn.getresponse()
                self.assertEqual(response.status, 200)
                self.assertEqual(response.read(), Path('ui/'+asset).read_bytes())
            finally:
                conn.close()

    def test_pod_native_limits_query_id_and_assets(self):
        query_id = '01234567-1234-1234-1234-012345678901'
        payload = b'{}' + b' ' * (65536-2)
        status, _, body = self.request('POST', '/api/pod-metric-discovery', payload, {'X-Goo-Query-ID': query_id})
        self.assertEqual(status, 400)
        self.assertEqual(json.loads(body), {'error': 'backend refusal'})
        self.assertEqual(Upstream.calls[-1][1], '/pod-metric-discovery')
        self.assertEqual(Upstream.calls[-1][3], payload)
        self.assertEqual({k.lower(): v for k, v in Upstream.calls[-1][2].items()}['x-goo-query-id'], query_id)
        before = len(Upstream.calls)
        self.assertEqual(self.request('POST', '/api/pod-metric-discovery', payload+b' ')[0], 400)
        self.assertEqual(len(Upstream.calls), before)
        plan_payload = b'{}' + b' ' * (8389632-2)
        self.assertEqual(self.request('POST', '/api/pod-dashboard/plan', plan_payload)[0], 400)
        self.assertEqual(Upstream.calls[-1][1], '/pod-dashboard/plan')
        before = len(Upstream.calls)
        self.assertEqual(self.request('POST', '/api/pod-dashboard/plan', plan_payload+b' ')[0], 400)
        self.assertEqual(len(Upstream.calls), before)
        for asset in ('pod-metrics.js', 'resource-native-view.js'):
            conn = http.client.HTTPConnection('127.0.0.1', self.proxy.server_port, timeout=5)
            try:
                conn.request('GET', '/'+asset)
                response = conn.getresponse()
                self.assertEqual(response.status, 200)
                self.assertEqual(response.read(), Path('ui/'+asset).read_bytes())
            finally:
                conn.close()

    def test_get_errors_and_queries_are_reencoded(self):
        Upstream.mode = 'text'
        status, headers, body = self.request('GET', '/api/operations?package=example&minutes=5')
        self.assertEqual(status, 400)
        self.assertEqual(headers['Content-Type'], 'application/json')
        self.assertEqual(json.loads(body), {'error':'backend refusal'})
        self.assertEqual(Upstream.calls[-1][1], '/operations?package=example&minutes=5')
        Upstream.mode = 'malformed'
        for method, path, payload in [('GET', '/api/operations', None), ('POST', '/api/measurements', b'{}'), ('POST', '/api/tail', b'{}')]:
            with self.subTest(path=path):
                status, _, body = self.request(method, path, payload)
                self.assertEqual(status, 502)
                self.assertEqual(json.loads(body)['error'], 'upstream returned a non-JSON error response')

    def test_tail_early_error_is_json_no_store(self):
        status, headers, body = self.request('POST', '/api/tail', b'{}')
        self.assertEqual(status, 400)
        self.assertEqual(headers['Content-Type'], 'application/json')
        self.assertEqual(headers['Cache-Control'], 'no-store')
        self.assertEqual(json.loads(body), {'error':'backend refusal'})

    def test_tail_forwards_frame_and_closes_idle_upstream_on_abort(self):
        Upstream.mode = 'stream'
        conn = http.client.HTTPConnection('127.0.0.1', self.proxy.server_port, timeout=5)
        response = None
        try:
            conn.request('POST', '/api/tail', body=b'{"environment":"prod"}')
            response = conn.getresponse()
            self.assertEqual(response.status, 200)
            self.assertEqual(response.getheader('Content-Type'), 'text/event-stream')
            self.assertEqual(response.getheader('Cache-Control'), 'no-store')
            self.assertEqual(response.getheader('X-Accel-Buffering'), 'no')
            lines = [response.fp.readline() for _ in range(4)]
            self.assertEqual(lines[1], b'event: reset\n')
            self.assertEqual(json.loads(lines[2][6:])['continuity'], 'unknown')
            response.close()
            conn.close()
            self.assertTrue(Upstream.ended.wait(2), 'proxy failed to close idle upstream after browser abort')
        finally:
            if response:
                response.close()
            conn.close()


if __name__ == '__main__':
    unittest.main()
