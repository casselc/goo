import logging
import os
import uuid
from opentelemetry import trace, metrics
from opentelemetry.sdk.resources import Resource
from opentelemetry.sdk.trace import TracerProvider
from opentelemetry.sdk.trace.sampling import ALWAYS_ON
from opentelemetry.sdk.trace.export import BatchSpanProcessor
from opentelemetry.exporter.otlp.proto.http.trace_exporter import OTLPSpanExporter
from opentelemetry.sdk.metrics import MeterProvider
from opentelemetry.sdk.metrics.export import PeriodicExportingMetricReader
from opentelemetry.exporter.otlp.proto.http.metric_exporter import OTLPMetricExporter
from opentelemetry.sdk._logs import LoggerProvider, LoggingHandler
from opentelemetry.sdk._logs.export import BatchLogRecordProcessor
from opentelemetry.exporter.otlp.proto.http._log_exporter import OTLPLogExporter


def setup(service):
    endpoint = os.getenv('OTEL_EXPORTER_OTLP_ENDPOINT', 'http://localhost:4318').rstrip('/')
    resource = Resource.create({'service.name': service, 'deployment.environment.name': 'local', 'service.instance.id': str(uuid.uuid4())})
    # Keep demo traces regardless of the host's OTEL_TRACES_SAMPLER setting.
    tp = TracerProvider(resource=resource, sampler=ALWAYS_ON)
    tp.add_span_processor(BatchSpanProcessor(OTLPSpanExporter(endpoint=endpoint + '/v1/traces'), schedule_delay_millis=500))
    trace.set_tracer_provider(tp)
    mp = MeterProvider(resource=resource, metric_readers=[PeriodicExportingMetricReader(
        OTLPMetricExporter(endpoint=endpoint + '/v1/metrics'), export_interval_millis=2000)])
    metrics.set_meter_provider(mp)
    lp = LoggerProvider(resource=resource)
    lp.add_log_record_processor(BatchLogRecordProcessor(OTLPLogExporter(endpoint=endpoint + '/v1/logs'), schedule_delay_millis=500))
    logger = logging.getLogger(service)
    logger.setLevel(logging.INFO)
    logger.addHandler(LoggingHandler(logger_provider=lp))
    logger.addHandler(logging.StreamHandler())
    meter = metrics.get_meter(service)
    return trace.get_tracer(service), logger, meter, (tp, mp, lp)


def emit(logger, name, item, **extra):
    attrs = {'app.event.name': name, 'app.reservation.id': item['id'],
             'app.tenant.id': item.get('tenant', 'demo'), 'app.release': item['release'],
             'app.region': item['region'], 'app.scenario': item['scenario'], **extra}
    logger.info(name, extra=attrs)
