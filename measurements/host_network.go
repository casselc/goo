package measurements

import "github.com/casselc/goo/metrics"

// HostNetworkInstrument pins one cumulative monotonic OTel byte instrument.
func HostNetworkInstrument(i Instrument) bool {
	return i.Kind == "counter" && i.Metric == "system.network.io" && i.Unit == "By" && i.SchemaURL == HostMetricsSchema && i.Temporality == 2 && i.Monotonic
}
func HostNetworkDefinition(d Definition) bool {
	return d.Kind == "counter" && d.CounterDomain == metrics.DecimalCounterDomain && d.Source.Resource != nil && d.Source.Resource.Kind == "host" && d.Source.Metric == "system.network.io" && d.Source.Unit == "By" && d.Source.SchemaURL == HostMetricsSchema
}
func HostIntervalCounterDefinition(d Definition) bool {
	return HostByteCounterDefinition(d) || d.Kind == "counter" && d.CounterDomain == metrics.DecimalCounterDomain && d.Source.Resource != nil && d.Source.Resource.Kind == "host" && d.Source.Metric == "system.cpu.time" && d.Source.Unit == "s" && d.Source.SchemaURL == HostMetricsSchema
}

// HostDiskInstrument pins the cumulative monotonic disk byte instrument.
func HostDiskInstrument(i Instrument) bool {
	return i.Kind == "counter" && i.Metric == "system.disk.io" && i.Unit == "By" && i.SchemaURL == HostMetricsSchema && i.Temporality == 2 && i.Monotonic
}
func HostDiskDefinition(d Definition) bool {
	return d.Kind == "counter" && d.CounterDomain == metrics.DecimalCounterDomain && d.Source.Resource != nil && d.Source.Resource.Kind == "host" && d.Source.Metric == "system.disk.io" && d.Source.Unit == "By" && d.Source.SchemaURL == HostMetricsSchema
}
func HostByteCounterDefinition(d Definition) bool {
	return HostNetworkDefinition(d) || HostDiskDefinition(d)
}
