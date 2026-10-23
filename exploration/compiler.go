package exploration

import (
	"fmt"
	"strconv"
	"strings"
)

type Plan struct {
	GroupedTimeSeries string            `json:"groupedTimeSeries,omitempty"`
	TimeSeries        string            `json:"timeSeries,omitempty"`
	Rows              string            `json:"rows"`
	Count             string            `json:"count"`
	Groups            string            `json:"groups,omitempty"`
	Fields            string            `json:"fields"`
	Parameters        map[string]string `json:"parameters"`
}
type compiler struct {
	signal string
	params map[string]string
}

func (c *compiler) param(value, typ string) string {
	key := fmt.Sprintf("p%d", len(c.params))
	c.params[key] = value
	return "{" + key + ":" + typ + "}"
}
func (c *compiler) field(f Field) (value, present string) {
	if f.Kind == "column" {
		mapping := map[string]string{"service": "ServiceName", "trace_id": "TraceId", "span_id": "SpanId", "scope": "ScopeName", "scope_version": "ScopeVersion", "name": "SpanName", "kind": "SpanKind", "status": "StatusCode", "duration_ns": "Duration", "timestamp_ns": "toUInt64(toUnixTimestamp64Nano(Timestamp))", "body": "Body", "severity": "SeverityText"}
		if f.Key == "status_class" {
			return statusClassSQL(), "1"
		}
		if f.Key == "duration_bucket" {
			return durationBucketSQL(), "1"
		}
		return mapping[f.Key], "1"
	}
	column := "ResourceAttributes"
	if f.Kind == "attribute" {
		column = "SpanAttributes"
		if c.signal == "logs" {
			column = "LogAttributes"
		}
	}
	key := c.param(f.Key, "String")
	return column + "[" + key + "]", "mapContains(" + column + ", " + key + ")"
}
func (c *compiler) predicate(p *Predicate) string {
	if p == nil {
		return "1"
	}
	if len(p.All) > 0 || len(p.Any) > 0 {
		parts := []string{}
		items, op := p.All, " AND "
		if len(p.Any) > 0 {
			items, op = p.Any, " OR "
		}
		for i := range items {
			parts = append(parts, c.predicate(&items[i]))
		}
		return "(" + strings.Join(parts, op) + ")"
	}
	v, present := c.field(*p.Field)
	if p.Op == "exists" {
		return present
	}
	if p.Op == "missing" {
		return "NOT (" + present + ")"
	}
	typ, _ := p.Field.Type(c.signal)
	sqltype := "String"
	if typ == "uint64" {
		sqltype = "UInt64"
	}
	parameter := c.param(p.Value, sqltype)
	op := map[string]string{"eq": "=", "ne": "!=", "gte": ">=", "lte": "<="}[p.Op]
	expr := v + " " + op + " " + parameter
	if p.Op == "contains" {
		expr = "position(" + v + ", " + parameter + ") > 0"
	}
	return "(" + present + " AND " + expr + ")"
}
func (c *compiler) source(q Query) (string, string) {
	table, attrs := "otel_traces", "SpanAttributes"
	if q.Signal == "logs" {
		table, attrs = "otel_logs", "LogAttributes"
	}
	where := "Timestamp >= fromUnixTimestamp64Milli({from:Int64}) AND Timestamp < fromUnixTimestamp64Milli({to:Int64}) AND mapContains(ResourceAttributes, 'deployment.environment.name') AND ResourceAttributes['deployment.environment.name'] = {environment:String} AND " + c.predicate(q.Where)
	if q.ReportedEdge != nil {
		where += " AND " + c.reportedEdge(q)
	}
	return " FROM " + table + " WHERE " + where, attrs
}
func Compile(q Query) (Plan, error) { return compile(q, q.Limit) }

func compile(q Query, rowLimit int) (Plan, error) {
	if e := q.Validate(); e != nil {
		return Plan{}, e
	}
	c := compiler{signal: q.Signal, params: map[string]string{"from": strconv.FormatInt(q.FromMS, 10), "to": strconv.FormatInt(q.ToMS, 10), "environment": q.Environment}}
	base, attrs := c.source(q)
	common := "toString(toUnixTimestamp64Nano(Timestamp)) AS timestamp_ns, ServiceName, TraceId, SpanId, ResourceAttributes, ScopeName, ScopeVersion, " + attrs + " AS Attributes"
	selectRows := common + ", SpanName, SpanKind, ParentSpanId, " + durationBucketSQL() + " AS duration_bucket, toString(Duration) AS duration_ns, StatusCode, StatusMessage, SpanAttributes, arrayMap(t -> toString(toUnixTimestamp64Nano(t)), Events.Timestamp) AS event_times_ns, Events.Name AS event_names, Events.Attributes AS event_attributes, `Links.Attributes` AS linked_attributes, `Links.TraceId` AS linked_traces, `Links.SpanId` AS linked_spans"
	order := "Timestamp, TraceId, SpanId, ServiceName, toString(ResourceAttributes), toString(SpanAttributes), SpanName, Duration"
	if q.Signal == "logs" {
		selectRows = common + ", Body, SeverityText"
		order = "Timestamp, TraceId, SpanId, ServiceName, Body, SeverityText, toString(ResourceAttributes), toString(LogAttributes)"
	}
	p := Plan{Rows: "SELECT " + selectRows + base + " ORDER BY " + order + fmt.Sprintf(" LIMIT %d OFFSET %d FORMAT JSONEachRow", rowLimit+1, q.Offset), Count: "SELECT toString(count()) AS count" + base + " FORMAT JSONEachRow"}
	aggregate, e := recordAggregateSQL(q)
	if e != nil {
		return Plan{}, e
	}
	if q.Numeric != nil {
		aggregate += ", " + c.numericSQL(*q.Numeric)
	}
	p.Count = "SELECT " + aggregate + base + " FORMAT JSONEachRow"

	if q.BucketMS != 0 {
		c.params["bucketMs"] = strconv.FormatInt(q.BucketMS, 10)
		p.TimeSeries = "SELECT toString({from:Int64} + intDiv(toUnixTimestamp64Milli(Timestamp) - {from:Int64}, {bucketMs:Int64}) * {bucketMs:Int64}) AS from_ms, " + aggregate + base + " GROUP BY from_ms ORDER BY toInt64(from_ms) LIMIT 241 FORMAT JSONEachRow"
	}
	if len(q.GroupBy) > 0 {
		dims := []string{}
		aliases := []string{}
		for i, f := range q.GroupBy {
			v, present := c.field(f)
			alias := fmt.Sprintf("g%d", i)
			dims = append(dims, "if("+present+", CAST(toString("+v+") AS Nullable(String)), NULL) AS "+alias)
			aliases = append(aliases, alias)
		}
		p.Groups = "SELECT " + strings.Join(dims, ", ") + ", " + aggregate + base + " GROUP BY " + strings.Join(aliases, ", ") + " ORDER BY " + strings.Join(aliases, ", ") + " LIMIT 1001 FORMAT JSONEachRow"
		if q.GroupTimeSeries {
			time := "toString({from:Int64} + intDiv(toUnixTimestamp64Milli(Timestamp) - {from:Int64}, {bucketMs:Int64}) * {bucketMs:Int64}) AS from_ms"
			p.GroupedTimeSeries = "SELECT " + strings.Join(dims, ", ") + ", " + time + ", " + aggregate + base + " GROUP BY " + strings.Join(aliases, ", ") + ", from_ms ORDER BY " + strings.Join(aliases, ", ") + ", toInt64(from_ms) LIMIT 1001 FORMAT JSONEachRow"
		}

	}
	p.Fields = "SELECT DISTINCT arrayJoin(arrayConcat(arrayMap(k -> concat('resource:', k), mapKeys(ResourceAttributes)), arrayMap(k -> concat('attribute:', k), mapKeys(" + attrs + ")))) AS field" + base + " ORDER BY field LIMIT 129 FORMAT JSONEachRow"
	p.Parameters = c.params
	return p, nil
}
