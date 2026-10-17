"""Run compiled SQL against embedded ClickHouse. Requires optional chdb."""
import json
import time
import unittest
from datetime import datetime, timezone
try:
    import chdb
except ImportError:
    chdb = None
from grammar import compile_query


@unittest.skipIf(chdb is None, 'Install requirements-dev.txt for embedded ClickHouse checks')
class SQLTests(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        from chdb import session
        cls.db = session.Session()
        cls.db.query("CREATE TABLE otel_logs (Timestamp DateTime64(9, 'UTC'), ServiceName String, Body String, TraceId String, SpanId String, LogAttributes Map(String,String)) ENGINE=Memory")
        cls.db.query("CREATE TABLE otel_traces (Timestamp DateTime64(9, 'UTC'), ServiceName String, SpanName String, TraceId String, SpanId String, ParentSpanId String, Duration UInt64, SpanKind String, StatusCode String, SpanAttributes Map(String,String), ResourceAttributes Map(String,String), `Links.TraceId` Array(String), `Links.SpanId` Array(String)) ENGINE=Memory")
        cls.db.query("CREATE TABLE otel_metrics_sum (TimeUnix DateTime64(9, 'UTC'), ServiceName String, MetricName String, Value Float64, Attributes Map(String,String), AggregationTemporality Int32, IsMonotonic Boolean) ENGINE=Memory")
        cls.db.query("CREATE TABLE otel_metrics_gauge (TimeUnix DateTime64(9, 'UTC'), ServiceName String, MetricName String, MetricUnit String, Value Float64, Attributes Map(String,String), ResourceAttributes Map(String,String)) ENGINE=Memory")
        cls.now = int(time.time() * 1000)
        now = cls.now / 1000
        records = []
        for id_, age, elapsed, outcome in [('good', 30, 1, 'confirmed'), ('late', 30, 14, 'confirmed'), ('missing', 30, None, None), ('pending', 2, None, None), ('reject', 30, 1, 'rejected')]:
            base = {'app.tenant.id': 'demo', 'app.reservation.id': id_, 'app.release': 'v1', 'app.region': 'west', 'app.scenario': id_ if id_ != 'pending' else 'good'}
            records.append({'Timestamp': now-age, 'Body': 'reservation.accepted', 'LogAttributes': {**base, 'app.accepted_at': str(now-age)}})
            if elapsed is not None:
                records.append({'Timestamp': now-age+elapsed, 'Body': 'reservation.terminal', 'LogAttributes': {**base, 'app.completed_at': str(now-age+elapsed), 'app.outcome': outcome}})
        # Duplicate exporter delivery must not double the subject contribution.
        records.append(records[0])
        for record in records:
            if isinstance(record['Timestamp'], (int, float)):
                record['Timestamp'] = datetime.fromtimestamp(record['Timestamp'], timezone.utc).strftime('%Y-%m-%d %H:%M:%S.%f')
        cls.db.query('INSERT INTO otel_logs FORMAT JSONEachRow\n' + '\n'.join(json.dumps(r) for r in records))

    @classmethod
    def tearDownClass(cls): cls.db.close()

    def query(self, spec):
        sql, params, _ = compile_query(spec, self.now)
        for name, value in params.items():
            typ = 'Int64' if name in ('now', 'fromMs', 'toMs') else 'UInt64' if name in ('minMs', 'maxMs') else 'UInt32' if name == 'minutes' else 'String'
            literal = value if typ != 'String' else "'" + value.replace('\\', '\\\\').replace("'", "\\'") + "'"
            sql = sql.replace('{' + name + ':' + typ + '}', literal)
        return [json.loads(line) for line in str(self.db.query(sql)).splitlines() if line]

    def test_deadline_counting_and_deduplication(self):
        rows = self.query({'view': 'completion'})
        self.assertEqual({r['reservation_id']: r['classification'] for r in rows},
                         {'good': 'good', 'late': 'bad', 'missing': 'bad', 'pending': 'pending', 'reject': 'bad'})

    def test_subject_drilldown(self):
        rows = self.query({'view': 'completion', 'reservationId': 'good'})
        self.assertEqual(len(rows), 1)

    def test_heatmap_counts_raw_spans_and_cell_drilldown_exactly(self):
        start = self.now - 60000
        records = []
        durations = [0, 9999999, 10000000, 49999999, 50000000, 99999999,
                     100000000, 499999999, 500000000, 999999999,
                     1000000000, 9999999999, 10000000000, 18446744073709551615]
        for i, duration in enumerate(durations):
            records.append({'Timestamp': datetime.fromtimestamp(start/1000, timezone.utc).strftime('%Y-%m-%d %H:%M:%S.%f'),
                            'ServiceName': 'lab-grpc', 'SpanKind': 'Server', 'SpanId': str(i),
                            'Duration': duration, 'SpanAttributes': {'app.reservation.id': 'heatmap-fixture'}})
        # Each band has two calls; an exporter/client span must not contribute.
        records.append({**records[0], 'SpanKind': 'Client'})
        self.db.query('INSERT INTO otel_traces FORMAT JSONEachRow\n' + '\n'.join(json.dumps(r) for r in records))
        heatmap = self.query({'view': 'heatmap', 'reservationId': 'heatmap-fixture'})
        self.assertEqual({int(r['band']): int(r['calls']) for r in heatmap}, {i: 2 for i in range(7)})
        edges = [0, 10, 50, 100, 500, 1000, 10000, 18446744073710]
        for band in range(7):
            selection = {'fromMs': start, 'toMs': start+1, 'minMs': edges[band], 'maxMs': edges[band+1]}
            rows = self.query({'view': 'traces', 'reservationId': 'heatmap-fixture', 'spanFilter': selection})
            self.assertEqual({r['SpanId'] for r in rows}, {str(band*2), str(band*2+1)})
        self.assertEqual(self.query({'view': 'traces', 'reservationId': 'heatmap-fixture',
                                    'spanFilter': {**selection, 'fromMs': start-1, 'toMs': start}}), [])

    def test_heatmap_population_exceeds_span_display_limit(self):
        self.db.query("INSERT INTO otel_traces (Timestamp, ServiceName, SpanKind, SpanId, Duration, SpanAttributes) SELECT fromUnixTimestamp64Milli(" + str(self.now-1000) + "), 'lab-grpc', 'Server', toString(number), 10000000, map('app.reservation.id','limit-fixture') FROM numbers(1101)")
        heatmap = self.query({'view': 'heatmap', 'reservationId': 'limit-fixture'})
        self.assertEqual(sum(int(r['calls']) for r in heatmap), 1101)
        self.assertEqual(len(self.query({'view': 'traces', 'reservationId': 'limit-fixture'})), 200)

    def test_queue_snapshots_keep_instances_and_exact_wait_population(self):
        from queue_telemetry import ATTRS
        records=[]
        for instance, value in [('one', 1), ('two', 5)]:
            for age, val in [(1001, value+1), (1000, value)]:
                records.append({'TimeUnix': datetime.fromtimestamp((self.now-age)/1000, timezone.utc).strftime('%Y-%m-%d %H:%M:%S.%f'),
                    'ServiceName': 'lab-consumer', 'MetricName': 'lab.queue.depth', 'MetricUnit': '{message}',
                    'Value': val, 'Attributes': ATTRS, 'ResourceAttributes': {'service.instance.id': instance}})
        self.db.query('INSERT INTO otel_metrics_gauge FORMAT JSONEachRow\n'+'\n'.join(json.dumps(r) for r in records))
        waits=[]
        for value in [0, 1000000000, 2000000000]:
            waits.append({'Timestamp': datetime.fromtimestamp((self.now-1000)/1000, timezone.utc).strftime('%Y-%m-%d %H:%M:%S.%f'),
                'ServiceName': 'lab-consumer', 'SpanAttributes': {**ATTRS, 'lab.queue.wait_ns': str(value)},
                'ResourceAttributes': {'service.instance.id': 'one'}})
        self.db.query('INSERT INTO otel_traces FORMAT JSONEachRow\n'+'\n'.join(json.dumps(r) for r in waits))
        rows=self.query({'view': 'queue'})
        self.assertEqual({r['instance']: float(r['value']) for r in rows if r['metric']=='lab.queue.depth'}, {'one':1, 'two':5})
        wait=next(r for r in rows if r['metric']=='lab.queue.wait.p95')
        self.assertEqual(float(wait['value']), 2)
        self.assertEqual(int(wait['samples']), 3)
        with self.assertRaises(ValueError): self.query({'view': 'queue', 'reservationId':'one'})

    def test_every_view_compiles_in_clickhouse(self):
        for view in ('logs', 'traces', 'latency', 'metrics', 'heatmap', 'queue'):
            with self.subTest(view=view): self.query({'view': view})


if __name__ == '__main__': unittest.main()
