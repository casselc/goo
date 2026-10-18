import base64
import json
import os
import time
import urllib.error
import urllib.parse
import urllib.request
from pathlib import Path
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer
from grammar import PACKAGE, compile_query, summarize


def clickhouse(sql, params):
    query = urllib.parse.urlencode({**{'param_' + key: value for key, value in params.items()},
                                   'max_execution_time': '10', 'max_result_rows': '1001',
                                   'result_overflow_mode': 'throw', 'readonly': '1'})
    endpoint = os.getenv('CLICKHOUSE_URL', 'http://localhost:8123')
    auth = (os.getenv('CLICKHOUSE_USER', 'lab') + ':' + os.getenv('CLICKHOUSE_PASSWORD', 'lab-local')).encode()
    request = urllib.request.Request(endpoint + '/?' + query, data=sql.encode(), headers={
        'Authorization': 'Basic ' + base64.b64encode(auth).decode(), 'Content-Type': 'text/plain'})
    with urllib.request.urlopen(request, timeout=15) as response:
        return [json.loads(line) for line in response.read().decode().splitlines() if line.strip()]


class Handler(BaseHTTPRequestHandler):
    def json_response(self, code, data):
        body = json.dumps(data).encode()
        self.send_response(code); self.send_header('Content-Type', 'application/json')
        self.send_header('Cache-Control', 'no-store'); self.end_headers(); self.wfile.write(body)

    def do_GET(self):
        if self.path == '/api/package': self.json_response(200, PACKAGE)
        elif self.path.startswith(('/api/http-package', '/api/operations', '/api/operation-packages', '/api/counter-rates', '/api/histograms', '/api/investigations', '/api/captures', '/api/telemetry-packages')):
            try:
                parsed = urllib.parse.urlsplit(self.path)
                if not (parsed.path.startswith('/api/captures/') or parsed.path.startswith('/api/investigations/') or parsed.path.startswith('/api/telemetry-packages/') or parsed.path in ('/api/http-package', '/api/operations', '/api/operation-packages', '/api/counter-rates', '/api/histograms', '/api/captures', '/api/telemetry-packages')): self.send_error(404); return
                endpoint = os.getenv('PACKAGES_URL', 'http://localhost:8089') + parsed.path.removeprefix('/api')
                with urllib.request.urlopen(endpoint + ('?' + parsed.query if parsed.query else ''), timeout=90) as response:
                    self.json_response(200, json.load(response))
            except urllib.error.HTTPError as exc:
                self.json_response(exc.code, json.loads(exc.read()))
            except Exception as exc: self.json_response(502, {'error': str(exc)})
        elif self.path == '/api/prebuilts':
            self.json_response(200, json.loads((Path(__file__).parent / 'definitions/prebuilts.json').read_text()))
        elif self.path == '/health':
            try:
                clickhouse('SELECT 1 AS ok FORMAT JSONEachRow', {})
                self.json_response(200, {'ok': True})
            except Exception as exc: self.json_response(503, {'ok': False, 'error': str(exc)})
        elif self.path in ('/', '/index.html'):
            self.send_response(200); self.send_header('Content-Type', 'text/html; charset=utf-8'); self.end_headers()
            self.wfile.write((Path(__file__).parent / 'ui/index.html').read_bytes())
        elif self.path == '/product.css':
            self.send_response(200); self.send_header('Content-Type', 'text/css; charset=utf-8'); self.end_headers()
            self.wfile.write((Path(__file__).parent / 'ui/product.css').read_bytes())
        elif self.path in ('/queries.js', '/namespace.js', '/captures.js', '/inventory.js', '/workspace.js', '/attributes.js', '/discovery.js', '/visualizations.js', '/workflows.js', '/operations.js', '/histograms.js', '/ratios.js', '/fleet.js', '/investigations.js', '/exploration.js', '/distributions.js', '/telemetry-packages.js'):
            self.send_response(200); self.send_header('Content-Type', 'application/javascript; charset=utf-8'); self.end_headers()
            self.wfile.write((Path(__file__).parent / 'ui' / self.path[1:]).read_bytes())
        else: self.send_error(404)

    def do_POST(self):
        if self.path.startswith('/api/queries/') or self.path in ('/api/captures','/api/investigations','/api/explore','/api/inventory', '/api/discovery','/api/compare','/api/telemetry-packages','/api/telemetry-packages/draft','/api/telemetry-packages/plan'):
            try:
                length = int(self.headers.get('Content-Length', 0))
                limit = 1048576 if self.path == '/api/compare' else 65536 if self.path.startswith('/api/telemetry-packages') else 16384
                if not 0 < length <= limit: raise ValueError('JSON request exceeds size limit')
                endpoint = os.getenv('PACKAGES_URL', 'http://localhost:8089') + self.path.removeprefix('/api')
                request = urllib.request.Request(endpoint, data=self.rfile.read(length), headers={'Content-Type': 'application/json', 'X-Goo-Query-ID': self.headers.get('X-Goo-Query-ID', '')})
                with urllib.request.urlopen(request, timeout=90) as response:
                    self.json_response(response.status, json.load(response))
            except urllib.error.HTTPError as exc: self.json_response(exc.code, json.loads(exc.read()))
            except ValueError as exc: self.json_response(400, {'error': str(exc)})
            except Exception as exc: self.json_response(502, {'error': str(exc)})
            return
        if self.path not in ('/api/query', '/api/explain'): self.send_error(404); return
        try:
            length = int(self.headers.get('Content-Length', 0))
            if not 0 < length <= 8192: raise ValueError('Expected JSON body, maximum 8 KiB')
            spec = json.loads(self.rfile.read(length))
            now = int(time.time() * 1000)
            sql, params, expression = compile_query(spec, now)
            result = {'expression': expression, 'sql': sql, 'parameters': params, 'evaluatedAt': now,
                      'evidence': 'demo-unverified', 'limit': 1000 if expression['view'] in ('completion', 'queue') else None if expression['view'] == 'heatmap' else 100 if expression['view'] == 'metrics' else 200}
            if self.path == '/api/query':
                rows = clickhouse(sql, params)
                result['rows'] = rows
                result['summary'] = summarize(rows) if expression['view'] == 'completion' else None
                result['possiblyTruncated'] = expression['view'] != 'heatmap' and len(rows) >= (1000 if expression['view'] in ('completion', 'queue') else 100 if expression['view'] == 'metrics' else 200)
            self.json_response(200, result)
        except (ValueError, TypeError) as exc: self.json_response(400, {'error': str(exc)})
        except urllib.error.HTTPError as exc:
            self.json_response(502, {'error': exc.read().decode()[:1500], 'hint': 'Collector tables appear after startup; wait 15 seconds and retry.'})
        except Exception as exc: self.json_response(502, {'error': str(exc)})

    def log_message(self, *_): pass


if __name__ == '__main__': ThreadingHTTPServer(('0.0.0.0', 8088), Handler).serve_forever()
