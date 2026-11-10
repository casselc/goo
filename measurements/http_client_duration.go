package measurements

const CurrentHTTPClientDurationMetric = "http.client.request.duration"
const CurrentHTTPClientRuleSet = "OpenTelemetry semantic conventions v1.45.0"
const CurrentHTTPClientSchema = "https://opentelemetry.io/schemas/1.45.0"

// CurrentHTTPClientDurationShape recognizes the pinned instrument shape only.
// Reported schema and dimension presence do not establish conformance or coverage.
func CurrentHTTPClientDurationShape(i Instrument) bool {
	return i.Metric == CurrentHTTPClientDurationMetric && i.Kind == "histogram" && i.Unit == "s"
}
