"""Small typed query subset: fixed sources, validated dimensions, CH parameters."""
import json
from pathlib import Path

PACKAGE = json.loads((Path(__file__).parent / 'definitions/reservations.json').read_text())
VIEWS = set(PACKAGE['views'])
DIMENSIONS = {'release': 'app.release', 'region': 'app.region', 'scenario': 'app.scenario'}


def validate(spec):
    if not isinstance(spec, dict) or set(spec) - {'view', 'minutes', 'groupBy', 'reservationId', 'scenario', 'spanFilter'}:
        raise ValueError('Allowed keys: view, minutes, groupBy, reservationId, scenario, spanFilter')
    view = spec.get('view', 'completion')
    if view not in VIEWS: raise ValueError('Unsupported view')
    minutes = spec.get('minutes', 30)
    if type(minutes) is not int or not 1 <= minutes <= 120: raise ValueError('minutes must be 1..120')
    group = spec.get('groupBy', 'release')
    if group not in DIMENSIONS: raise ValueError('groupBy must be release, region, or scenario')
    for key in ('reservationId', 'scenario'):
        if not isinstance(spec.get(key, ''), str) or len(spec.get(key, '')) > 100:
            raise ValueError(key + ' must be a string of at most 100 characters')
    if spec.get('scenario', '') not in ('', 'good', 'retry', 'late', 'missing', 'reject'):
        raise ValueError('Unknown scenario')
    selection = spec.get('spanFilter')
    if selection is not None:
        if view != 'traces' or not isinstance(selection, dict) or set(selection) != {'fromMs', 'toMs', 'minMs', 'maxMs'}:
            raise ValueError('spanFilter is a heatmap interval for the traces view')
        if any(type(selection[k]) is not int or not 0 <= selection[k] <= 9007199254740991 for k in selection):
            raise ValueError('spanFilter bounds must be nonnegative safe integers')
        if selection['toMs'] > 4102444800000:
            raise ValueError('spanFilter timestamp exceeds supported range (2100)')
        if not selection['fromMs'] < selection['toMs'] or not selection['minMs'] < selection['maxMs']:
            raise ValueError('spanFilter intervals must be nonempty and half-open')
    return {'view': view, 'minutes': minutes, 'groupBy': group,
            'reservationId': spec.get('reservationId', ''), 'scenario': spec.get('scenario', ''), **({'spanFilter': selection} if selection is not None else {})}


def compile_query(spec, now):
    s = validate(spec)
    params = {'now': str(now), 'minutes': str(s['minutes']),
              'id': s['reservationId'], 'scenario': s['scenario']}
    lower = 'Timestamp >= fromUnixTimestamp64Milli({now:Int64}) - toIntervalMinute({minutes:UInt32})'
    upper = 'Timestamp < fromUnixTimestamp64Milli({now:Int64})' if s['view'] == 'heatmap' else 'Timestamp <= fromUnixTimestamp64Milli({now:Int64})'
    if 'spanFilter' in s:
        lower = 'Timestamp >= fromUnixTimestamp64Milli({fromMs:Int64})'
        upper = 'Timestamp < fromUnixTimestamp64Milli({toMs:Int64})'
    attrs = 'SpanAttributes' if s['view'] in ('traces', 'latency', 'heatmap') else 'LogAttributes'
    filters = f"{lower} AND {upper} AND ({{id:String}} = '' OR {attrs}['app.reservation.id'] = {{id:String}})"
    # Filter the business population consistently across investigation views.
    filters += f" AND ({{scenario:String}} = '' OR {attrs}['app.scenario'] = {{scenario:String}})"
    if 'spanFilter' in s:
        params.update({key: str(value) for key, value in s['spanFilter'].items()})
        # Exact raw-span predicate shared with the heatmap, including overflow.
        filters += " AND ServiceName = 'lab-grpc' AND SpanKind = 'Server'"
        filters += ' AND Timestamp >= fromUnixTimestamp64Milli({fromMs:Int64}) AND Timestamp < fromUnixTimestamp64Milli({toMs:Int64})'
        filters += ' AND toUInt128(Duration) >= toUInt128({minMs:UInt64}) * 1000000 AND toUInt128(Duration) < toUInt128({maxMs:UInt64}) * 1000000'
    if s['view'] == 'queue':
        if s['reservationId'] or s['scenario']:
            raise ValueError('Queue resource telemetry cannot filter by reservation/scenario; clear selection')
        sql = """SELECT toUnixTimestamp64Milli(toDateTime64(toStartOfMinute(TimeUnix), 3)) AS minute_ms,
 MetricName AS metric, ServiceName AS service, ResourceAttributes['service.instance.id'] AS instance,
 MetricUnit AS unit, argMax(Value, TimeUnix) AS value,
 toUnixTimestamp64Milli(max(TimeUnix)) AS sample_ms, count() AS samples
 FROM otel_metrics_gauge
 WHERE TimeUnix >= fromUnixTimestamp64Milli({now:Int64}) - toIntervalMinute({minutes:UInt32})
 AND TimeUnix < fromUnixTimestamp64Milli({now:Int64}) AND ServiceName = 'lab-consumer'
 AND Attributes['messaging.destination.name'] = 'reservations'
 AND Attributes['messaging.consumer.group.name'] = 'inventory'
 AND MetricName IN ('lab.queue.depth','lab.queue.pending','lab.worker.active','lab.worker.capacity','lab.worker.utilization')
 GROUP BY minute_ms, metric, service, instance, unit
 UNION ALL
 SELECT toUnixTimestamp64Milli(toDateTime64(toStartOfMinute(Timestamp), 3)) AS minute_ms,
 'lab.queue.wait.p95' AS metric, ServiceName AS service, ResourceAttributes['service.instance.id'] AS instance,
 's' AS unit, quantileExact(0.95)(toUInt64OrNull(SpanAttributes['lab.queue.wait_ns']) / 1e9) AS value,
 toUnixTimestamp64Milli(max(Timestamp)) AS sample_ms, count() AS samples
 FROM otel_traces
 WHERE Timestamp >= fromUnixTimestamp64Milli({now:Int64}) - toIntervalMinute({minutes:UInt32})
 AND Timestamp < fromUnixTimestamp64Milli({now:Int64}) AND ServiceName = 'lab-consumer'
 AND SpanAttributes['messaging.destination.name'] = 'reservations'
 AND SpanAttributes['messaging.consumer.group.name'] = 'inventory'
 AND isNotNull(toUInt64OrNull(SpanAttributes['lab.queue.wait_ns']))
 GROUP BY minute_ms, service, instance
 ORDER BY minute_ms DESC, metric, instance LIMIT 1000"""
    elif s['view'] == 'completion':
        deadline = PACKAGE['deadlineSeconds']
        allowance = PACKAGE['ingestionAllowanceSeconds']
        sql = f"""
WITH subjects AS (
 SELECT LogAttributes['app.tenant.id'] AS tenant,
        LogAttributes['app.reservation.id'] AS reservation_id,
        minIf(toFloat64OrZero(LogAttributes['app.accepted_at']), Body = 'reservation.accepted') AS accepted_at,
        minIf(toFloat64OrZero(LogAttributes['app.completed_at']), Body = 'reservation.terminal') AS completed_at,
        countIf(Body = 'reservation.accepted') AS accept_records,
        countIf(Body = 'reservation.terminal') AS terminal_records,
        uniqExactIf(tuple(LogAttributes['app.completed_at'], LogAttributes['app.outcome']), Body = 'reservation.terminal') AS decisions,
        anyIf(LogAttributes['app.outcome'], Body = 'reservation.terminal') AS outcome,
        anyIf(LogAttributes['{DIMENSIONS[s['groupBy']]}'], Body = 'reservation.accepted') AS cohort
 FROM otel_logs WHERE {filters}
   AND Body IN ('reservation.accepted', 'reservation.terminal')
 GROUP BY tenant, reservation_id
)
SELECT tenant, reservation_id, cohort, accepted_at, completed_at,
 multiIf(decisions > 1, 'unknown',
         accepted_at > {{now:Int64}} / 1000.0 - {deadline + allowance}, 'pending',
         terminal_records = 0, 'bad',
         completed_at < accepted_at, 'unknown',
         outcome = 'confirmed' AND completed_at <= accepted_at + {deadline}, 'good', 'bad') AS classification,
 multiIf(decisions > 1, 'conflicting decisions', terminal_records = 0, 'no completion',
         completed_at < accepted_at, 'clock violation', outcome != 'confirmed', 'rejected',
         completed_at > accepted_at + {deadline}, 'late', 'timely') AS reason
FROM subjects WHERE accept_records > 0
ORDER BY accepted_at DESC LIMIT 1000"""
    elif s['view'] == 'logs':
        sql = f"SELECT Timestamp, ServiceName, Body, TraceId, SpanId, LogAttributes FROM otel_logs WHERE {filters} ORDER BY Timestamp DESC LIMIT 200"
    elif s['view'] == 'traces':
        sql = f"SELECT Timestamp, toString(toUnixTimestamp64Nano(Timestamp)) AS timestamp_ns, toString(Duration) AS duration_ns, ServiceName, SpanName, TraceId, SpanId, ParentSpanId, Duration / 1e6 AS duration_ms, StatusCode, SpanAttributes, `Links.TraceId` AS linked_traces, `Links.SpanId` AS linked_spans FROM otel_traces WHERE {filters} ORDER BY Timestamp DESC LIMIT 200"
    elif s['view'] == 'heatmap':
        # At most 121 intersecting minutes * 7 duration bands = 847 cells.
        # Aggregate the full matching population; never histogram a span LIMIT.
        sql = f"""SELECT toUnixTimestamp64Milli(toDateTime64(toStartOfMinute(Timestamp), 3)) AS bucket_ms,
 multiIf(Duration < 10000000, 0, Duration < 50000000, 1,
 Duration < 100000000, 2, Duration < 500000000, 3,
 Duration < 1000000000, 4, Duration < 10000000000, 5, 6) AS band,
 count() AS calls FROM otel_traces WHERE {filters}
 AND ServiceName = 'lab-grpc' AND SpanKind = 'Server'
 GROUP BY bucket_ms, band ORDER BY bucket_ms, band"""
    elif s['view'] == 'latency':
        sql = f"""SELECT toStartOfMinute(Timestamp) AS minute,
 SpanAttributes['{DIMENSIONS[s['groupBy']]}'] AS cohort,
 count() AS calls, quantile(0.50)(Duration/1e6) AS p50_ms,
 quantile(0.95)(Duration/1e6) AS p95_ms
 FROM otel_traces WHERE {filters} AND ServiceName = 'lab-grpc' AND SpanKind = 'Server'
 GROUP BY minute, cohort ORDER BY minute DESC LIMIT 200"""
    else:
        if s['reservationId'] or s['scenario']:
            raise ValueError('Native metrics cannot filter by reservation/scenario in this initial view; use a population view')
        # Exported point values, explicitly not a rate or sum across cumulative points.
        sql = """SELECT TimeUnix, ServiceName, MetricName, Value, Attributes,
 AggregationTemporality, IsMonotonic FROM otel_metrics_sum
 WHERE TimeUnix >= fromUnixTimestamp64Milli({now:Int64}) - toIntervalMinute({minutes:UInt32})
 AND TimeUnix <= fromUnixTimestamp64Milli({now:Int64})
 ORDER BY TimeUnix DESC LIMIT 100"""
    return sql + ' FORMAT JSONEachRow', params, s


def summarize(rows):
    totals = {name: 0 for name in ('good', 'bad', 'pending', 'unknown')}
    for row in rows: totals[row['classification']] += 1
    eligible = totals['good'] + totals['bad']
    ratio = totals['good'] / eligible if eligible and not totals['unknown'] else None
    burn = (1 - ratio) / (1 - PACKAGE['target']) if ratio is not None else None
    return {**totals, 'eligible': eligible, 'successRatio': ratio, 'burnRate': burn,
            'alarmPreview': 'unknown' if burn is None else ('would-fire' if burn > PACKAGE['alarm']['threshold'] else 'clear'),
            'evidence': 'demo-unverified', 'scope': 'returned rows (maximum 1000); not a production SLO'}
