package measurements

const CurrentRPCDurationMetric = "rpc.server.call.duration"
const CurrentRPCRuleSet = "OpenTelemetry semantic conventions v1.45.0"

// CurrentRPCDurationShape recognizes only the pinned instrument name, type and
// unit. It does not establish a valid binding, interval coverage or conformance.
func CurrentRPCDurationShape(i Instrument) bool {
	return i.Metric == CurrentRPCDurationMetric && i.Kind == "histogram" && i.Unit == "s"
}
