package exploration

import (
	"encoding/json"
	"sort"
	"strings"
)

// A standard exporter schema fixture independent of sourcecap's implementation.
// Extra known columns are harmless; missing/type tests remove individual fields.
func sourceCapabilityResponse431(sql string) (string, bool) {
	if strings.HasPrefix(sql, "SELECT ") && strings.HasSuffix(sql, " LIMIT 0 FORMAT JSONEachRow") {
		return "", true
	}
	if !strings.HasPrefix(sql, "DESCRIBE TABLE ") {
		return "", false
	}
	columns := map[string]string{}
	for _, name := range []string{"ServiceName", "TraceId", "SpanId", "ScopeName", "ScopeVersion", "Body", "SeverityText", "SpanName", "SpanKind", "ParentSpanId", "StatusCode", "StatusMessage", "ScopeSchemaUrl", "ResourceSchemaUrl", "MetricName", "MetricUnit"} {
		columns[name] = "String"
	}
	for _, name := range []string{"Timestamp", "TimeUnix"} {
		columns[name] = "DateTime64(9)"
	}
	for _, name := range []string{"ResourceAttributes", "LogAttributes", "SpanAttributes", "Attributes"} {
		columns[name] = "Map(String,String)"
	}
	columns["Duration"] = "UInt64"
	columns["SeverityNumber"] = "UInt8"
	columns["Events.Timestamp"] = "Array(DateTime64(9))"
	for _, name := range []string{"Events.Name", "Links.TraceId", "Links.SpanId"} {
		columns[name] = "Array(String)"
	}
	for _, name := range []string{"Events.Attributes", "Links.Attributes"} {
		columns[name] = "Array(Map(String,String))"
	}
	columns["AggregationTemporality"] = "Int32"
	columns["IsMonotonic"] = "Bool"
	names := make([]string, 0, len(columns))
	for name := range columns {
		names = append(names, name)
	}
	sort.Strings(names)
	var output strings.Builder
	for _, name := range names {
		row, _ := json.Marshal(map[string]string{"name": name, "type": columns[name]})
		output.Write(row)
		output.WriteByte('\n')
	}
	return output.String(), true
}
