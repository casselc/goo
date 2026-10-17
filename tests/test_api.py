import json
import threading
import unittest
import urllib.error
import urllib.request
from http.server import ThreadingHTTPServer
from unittest.mock import patch
from server import Handler


class APITests(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        cls.server = ThreadingHTTPServer(('127.0.0.1', 0), Handler)
        cls.thread = threading.Thread(target=cls.server.serve_forever, daemon=True)
        cls.thread.start()
        cls.url = 'http://127.0.0.1:' + str(cls.server.server_address[1])

    @classmethod
    def tearDownClass(cls): cls.server.shutdown(); cls.server.server_close()

    def post(self, path, body):
        req = urllib.request.Request(self.url + path, data=json.dumps(body).encode(), headers={'Content-Type': 'application/json'})
        return urllib.request.urlopen(req)

    def test_explain_does_not_execute_query(self):
        with patch('server.clickhouse', side_effect=AssertionError('should not execute')):
            with self.post('/api/explain', {'view': 'logs'}) as response:
                data = json.load(response)
            self.assertIn('otel_logs', data['sql'])

    def test_query_returns_explicit_evidence_status(self):
        with patch('server.clickhouse', return_value=[{'classification': 'good'}]):
            with self.post('/api/query', {'view': 'completion'}) as response: data = json.load(response)
        self.assertEqual(data['summary']['good'], 1)
        self.assertEqual(data['evidence'], 'demo-unverified')

    def test_invalid_query_returns_400(self):
        with self.assertRaises(urllib.error.HTTPError) as exc: self.post('/api/query', {'view': 'unknown'})
        self.assertEqual(exc.exception.code, 400)


if __name__ == '__main__': unittest.main()
