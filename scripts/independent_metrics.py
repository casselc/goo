"""Ordinary OTel producer: no goo manifest, annotations or instrumentation imports."""
from opentelemetry import metrics
from opentelemetry.sdk.metrics import MeterProvider
from opentelemetry.sdk.metrics.export import PeriodicExportingMetricReader
from opentelemetry.exporter.otlp.proto.http.metric_exporter import OTLPMetricExporter
from opentelemetry.sdk.resources import Resource

provider = MeterProvider(resource=Resource.create({
    "service.name": "independent-metrics", "service.instance.id": "one",
    "deployment.environment.name": "local",
}), metric_readers=[PeriodicExportingMetricReader(OTLPMetricExporter(
    endpoint="http://127.0.0.1:4318/v1/metrics"), export_interval_millis=60000)])
metrics.set_meter_provider(provider)
meter = metrics.get_meter("ordinary.application")  # optional version/schema omitted
counter = meter.create_counter("jobs.completed", unit="{job}")
for outcome in ("ok", "error"):
    counter.add(3, {"operation": "import", "outcome": outcome})
provider.force_flush()
for outcome in ("ok", "error"):
    counter.add(2, {"operation": "import", "outcome": outcome})
provider.force_flush()
provider.shutdown()
