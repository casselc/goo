package measurements

// HostFilesystemInstrument pins reported filesystem byte stocks, not counter rates.
func HostFilesystemInstrument(i Instrument) bool {
	return i.Kind == "stock" && i.Unit == "By" && i.SchemaURL == HostMetricsSchema && i.Temporality == 2 && !i.Monotonic && (i.Metric == "system.filesystem.usage" || i.Metric == "system.filesystem.limit")
}
func HostFilesystemDefinition(d Definition) bool {
	return d.Kind == "stock" && d.StockDomain == "nonnegative-integer" && d.Source.Resource != nil && d.Source.Resource.Kind == "host" && d.Source.Unit == "By" && d.Source.SchemaURL == HostMetricsSchema && (d.Source.Metric == "system.filesystem.usage" || d.Source.Metric == "system.filesystem.limit")
}

// HostFilesystemUtilizationDefinition is a reported fraction, never a byte ratio.
func HostFilesystemUtilizationDefinition(d Definition) bool {
	return d.Kind == "gauge" && d.Source.Resource != nil && d.Source.Resource.Kind == "host" && d.Source.Metric == "system.filesystem.utilization" && d.Source.Unit == "1" && d.Source.SchemaURL == HostMetricsSchema
}
