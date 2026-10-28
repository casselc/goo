"""Profile startup and real local HTTP boundaries without sample assets or sources."""
import http.client
import importlib.util
import json
import os
from pathlib import Path
import shutil
import subprocess
import sys
import tempfile
import threading
import unittest
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer
from unittest.mock import patch

import server


class GenericUpstream(BaseHTTPRequestHandler):
    calls = []
    capability_status = 200

    def handle_request(self):
        length = int(self.headers.get('Content-Length', '0'))
        body = self.rfile.read(length) if length else b''
        type(self).calls.append((self.command, self.path, body))
        if self.path == '/capabilities':
            code = self.capability_status
            result = {'version': 'goo/runtime-capabilities/v1', 'profile': 'generic',
                      'preparedHTTP': {'state': 'unconfigured'},
                      'declaredOperations': {'state': 'unconfigured'}} if code == 200 else {'error': 'backend metadata unavailable'}
        elif self.path.startswith('/http-package'):
            code, result = 503, {'error': 'Prepared HTTP is not configured', 'code': 'capability_unconfigured', 'capability': 'prepared-http'}
        else:
            code, result = 200, {'state': 'no_data', 'count': '0', 'request': json.loads(body) if body else None}
        self.send_response(code)
        self.send_header('Content-Type', 'application/json')
        self.end_headers()
        self.wfile.write(json.dumps(result).encode())

    do_GET = do_POST = handle_request

    def log_message(self, *_):
        pass


class GenericStartup(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        cls.tree = tempfile.TemporaryDirectory(prefix='goo-generic-ui-', dir='/tmp')
        cls.root = Path(cls.tree.name)
        source = Path(__file__).resolve().parents[1]
        shutil.copy(source / 'server.py', cls.root / 'server.py')
        shutil.copytree(source / 'ui', cls.root / 'ui')
        shutil.copytree(source / 'api', cls.root / 'api')
        spec = importlib.util.spec_from_file_location('generic_ui_fixture', cls.root / 'server.py')
        cls.module = importlib.util.module_from_spec(spec)
        with patch.dict(os.environ, {'GOO_PROFILE': 'generic'}):
            spec.loader.exec_module(cls.module)
        cls.module.clickhouse = lambda *_: (_ for _ in ()).throw(AssertionError('unexpected source read'))
        cls.upstream = ThreadingHTTPServer(('127.0.0.1', 0), GenericUpstream)
        cls.http = ThreadingHTTPServer(('127.0.0.1', 0), cls.module.Handler)
        for http in (cls.upstream, cls.http):
            threading.Thread(target=http.serve_forever, daemon=True).start()
        cls.config = patch.dict(os.environ, {'PACKAGES_URL': 'http://127.0.0.1:' + str(cls.upstream.server_port)})
        cls.config.start()

    @classmethod
    def tearDownClass(cls):
        cls.config.stop()
        for http in (cls.http, cls.upstream):
            http.shutdown()
            http.server_close()
        cls.tree.cleanup()

    def setUp(self):
        GenericUpstream.calls = []
        GenericUpstream.capability_status = 200

    def request(self, method, path, body=None):
        connection = http.client.HTTPConnection('127.0.0.1', self.http.server_port, timeout=5)
        try:
            connection.request(method, path, json.dumps(body) if body is not None else None,
                               {'Content-Type': 'application/json'})
            response = connection.getresponse()
            return response.status, response.getheader('Content-Type'), response.read()
        finally:
            connection.close()

    def test_standalone_import_without_grammar_or_any_sample_files(self):
        self.assertFalse((self.root / 'grammar.py').exists())
        self.assertFalse((self.root / 'definitions').exists())
        self.assertFalse((self.root / 'examples').exists())
        env = {**os.environ, 'GOO_PROFILE': 'generic', 'PYTHONPATH': ''}
        env.pop('GOO_HOST', None)
        env.pop('GOO_PORT', None)
        result = subprocess.run([sys.executable, '-c',
                                 "import server,sys; assert 'grammar' not in sys.modules; assert server.BIND_HOST=='127.0.0.1'; print(server.PROFILE)"],
                                cwd=self.root, env=env, text=True, capture_output=True, timeout=10)
        self.assertEqual(result.returncode, 0, result.stderr)
        self.assertEqual(result.stdout.strip(), 'generic')
        self.assertFalse(hasattr(self.module, 'PACKAGE'))
        self.assertFalse(hasattr(self.module, 'compile_query'))

    def test_invalid_startup_refuses_before_binding_and_missing_lab_grammar_is_actionable(self):
        for settings, message in [({'GOO_PROFILE': ''}, 'GOO_PROFILE'),
                                  ({'GOO_PROFILE': 'generic', 'PACKAGES_URL': 'http://user:do-not-print@host:invalid'}, 'PACKAGES_URL'),
                                  ({'GOO_PROFILE': 'lab'}, 'use GOO_PROFILE=generic')]:
            env = {**os.environ, 'PYTHONPATH': '', **settings}
            result = subprocess.run([sys.executable, '-c', 'import server'], cwd=self.root,
                                    env=env, text=True, capture_output=True, timeout=10)
            self.assertNotEqual(result.returncode, 0)
            self.assertIn(message, result.stderr)
            self.assertNotIn('do-not-print', result.stderr)

    def test_ui_config_static_runtime_and_docs_read_no_upstream(self):
        code, _, body = self.request('GET', '/api/ui-config')
        self.assertEqual(code, 200)
        self.assertEqual(json.loads(body), {'version': 'goo/ui-config/v1', 'profile': 'generic', 'examples': {'state': 'unconfigured'}})
        for path in ('/', '/ui-components', '/display.mjs', '/app.mjs', '/components.mjs', '/attributes.mjs', '/record-presentation.mjs', '/record-columns.mjs', '/exception-detail.mjs', '/aligned-histogram.mjs', '/tag-list.mjs', '/numeric-bins.mjs', '/trace-page-breakdown-view.mjs', '/trace-comparison-view.mjs', '/service-dashboard.mjs', '/product.css', '/api-docs', '/openapi.json'):
            with self.subTest(path=path):
                code, content_type, body = self.request('GET', path)
                self.assertEqual(code, 200)
                self.assertTrue(body)
                if path.endswith('.mjs'):
                    self.assertIn('javascript', content_type)
        for retired in ('/app-display.mjs', '/workspace.js'):
            self.assertEqual(self.request('GET', retired)[0], 404)
        self.assertEqual(GenericUpstream.calls, [])

    def test_disabled_examples_refuse_instead_of_empty_inventory_or_grammar_import(self):
        for method, path, capability in [('GET', '/api/package', 'example-grammar'),
                                         ('POST', '/api/query', 'example-grammar'),
                                         ('POST', '/api/explain', 'example-grammar'),
                                         ('GET', '/api/example-questions', 'example-catalog'),
                                         ('GET', '/api/prebuilts', 'example-catalog')]:
            with self.subTest(path=path):
                code, _, body = self.request(method, path, {} if method == 'POST' else None)
                value = json.loads(body)
                self.assertEqual(code, 503)
                self.assertEqual(value['code'], 'capability_unconfigured')
                self.assertEqual(value['capability'], capability)
                self.assertTrue(value['error'])
        self.assertEqual(GenericUpstream.calls, [])

    def test_capabilities_and_generic_execution_are_proxied_without_example_data(self):
        code, _, body = self.request('GET', '/api/capabilities')
        self.assertEqual(code, 200)
        self.assertEqual(json.loads(body)['declaredOperations']['state'], 'unconfigured')
        request = {'signal': 'spans', 'environment': 'prod', 'fromMs': 1, 'toMs': 2}
        code, _, body = self.request('POST', '/api/explore', request)
        self.assertEqual(code, 200)
        self.assertEqual(json.loads(body)['request'], request)
        self.assertEqual([path for _, path, _ in GenericUpstream.calls], ['/capabilities', '/explore'])

    def test_unconfigured_prepared_and_failed_capability_metadata_remain_errors(self):
        code, _, body = self.request('GET', '/api/http-package?minutes=2')
        self.assertEqual(code, 503)
        self.assertEqual(json.loads(body)['capability'], 'prepared-http')
        GenericUpstream.capability_status = 502
        code, _, body = self.request('GET', '/api/capabilities')
        self.assertEqual(code, 502)
        self.assertEqual(json.loads(body), {'error': 'backend metadata unavailable'})


class StartupConfiguration(unittest.TestCase):
    def test_default_lab_and_explicit_generic_bind_contract(self):
        self.assertEqual(server.startup_config({}), ('lab', '0.0.0.0', 8088))
        self.assertEqual(server.startup_config({'GOO_PROFILE': 'generic'}), ('generic', '127.0.0.1', 8088))
        self.assertEqual(server.startup_config({'GOO_PROFILE': 'generic', 'GOO_HOST': '0.0.0.0', 'GOO_PORT': '18088'}), ('generic', '0.0.0.0', 18088))
        self.assertEqual(server.ui_config()['examples']['state'], 'configured')
        self.assertIn('views', server.PACKAGE)

    def test_explicit_invalid_configuration_fails_without_disclosing_credentials(self):
        for setting in ({'GOO_PROFILE': ''}, {'GOO_PROFILE': 'typo'}, {'GOO_HOST': ''},
                        {'GOO_HOST': '::1'}, {'GOO_HOST': 'bad host'}, {'GOO_PORT': ''},
                        {'GOO_PORT': '0'}, {'GOO_PORT': '65536'}, {'GOO_PORT': '1\n'},
                        {'PACKAGES_URL': ''}, {'PACKAGES_URL': 'file:///tmp/store'},
                        {'PACKAGES_URL': 'http://user:do-not-print@host:bad'},
                        {'CLICKHOUSE_URL': 'http://host?query=private'}, {'CLICKHOUSE_URL': ' http://host'}):
            with self.subTest(keys=list(setting)):
                with self.assertRaises(ValueError) as caught:
                    server.startup_config(setting)
                self.assertNotIn('do-not-print', str(caught.exception))
                self.assertNotIn('private', str(caught.exception))


if __name__ == '__main__':
    unittest.main()
