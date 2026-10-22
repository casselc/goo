"""Exercise the actual handler's static docs without ClickHouse or a socket."""
import io
import json
import unittest

from server import Handler


class DocumentationRoutes(unittest.TestCase):
    def request(self, path):
        handler = object.__new__(Handler)
        handler.path = path
        handler.wfile = io.BytesIO()
        headers = {}
        statuses = []
        handler.send_response = statuses.append
        handler.send_header = lambda key, value: headers.update({key: value})
        handler.end_headers = lambda: None
        handler.do_GET()
        self.assertEqual(statuses, [200])
        self.assertEqual(headers['Cache-Control'], 'no-store')
        return headers, handler.wfile.getvalue()

    def test_spec_download(self):
        headers, body = self.request('/openapi.json')
        self.assertEqual(headers['Content-Type'], 'application/json')
        spec = json.loads(body)
        self.assertEqual(spec['openapi'], '3.1.0')
        self.assertIn('/tail', spec['paths'])

    def test_browsable_docs(self):
        headers, body = self.request('/api-docs')
        self.assertTrue(headers['Content-Type'].startswith('text/html'))
        self.assertIn(b"fetch('/openapi.json')", body)
        self.assertIn(b'id="search"', body)
        self.assertNotIn(b'cdn.', body)


if __name__ == '__main__':
    unittest.main()
