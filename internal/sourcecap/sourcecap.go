// Package sourcecap checks the fixed ClickStack exporter profiles used by Goo's
// record and instrument discovery adapters. Readiness is not observed data,
// source completeness, or support for evaluating a particular instrument.
package sourcecap

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/casselc/goo/internal/sourcehttp"
)

const (
	Logs                 = "logs"
	Spans                = "spans"
	Gauge                = "gauge"
	Counter              = "counter"
	Histogram            = "histogram"
	ExponentialHistogram = "exponential-histogram"
)

// Read retains the source owner's transport, response budgets and cancellation.
// Inspect supplies only fixed SQL and never accepts a table name from a caller.
type Read func(context.Context, string, map[string]string) ([]json.RawMessage, error)

type Family struct {
	Kind   string `json:"kind"`
	Table  string `json:"table"`
	State  string `json:"state"`
	Detail string `json:"detail"`
}

func (f Family) Compatible() bool { return f.State == "compatible" }

type column struct{ name, shape string }
type spec struct {
	kind, table string
	columns     []column
}

func specs() []spec {
	record := []column{{"Timestamp", "time"}, {"ServiceName", "string"}, {"TraceId", "string"}, {"SpanId", "string"}, {"ResourceAttributes", "map"}, {"ScopeName", "string"}, {"ScopeVersion", "string"}}
	metric := []column{{"TimeUnix", "time"}, {"ServiceName", "string"}, {"ResourceAttributes", "map"}, {"ScopeName", "string"}, {"ScopeVersion", "string"}, {"ScopeSchemaUrl", "string"}, {"ResourceSchemaUrl", "string"}, {"MetricName", "string"}, {"MetricUnit", "string"}, {"Attributes", "map"}}
	plus := func(base []column, extra ...column) []column { return append(append([]column(nil), base...), extra...) }
	return []spec{
		{Logs, "otel_logs", plus(record, column{"LogAttributes", "map"}, column{"Body", "string"}, column{"SeverityText", "string"}, column{"SeverityNumber", "uint8"})},
		{Spans, "otel_traces", plus(record, column{"SpanAttributes", "map"}, column{"SpanName", "string"}, column{"SpanKind", "string"}, column{"ParentSpanId", "string"}, column{"Duration", "uint64"}, column{"StatusCode", "string"}, column{"StatusMessage", "string"}, column{"Events.Timestamp", "times"}, column{"Events.Name", "strings"}, column{"Events.Attributes", "maps"}, column{"Links.Attributes", "maps"}, column{"Links.TraceId", "strings"}, column{"Links.SpanId", "strings"})},
		{Gauge, "otel_metrics_gauge", plus(metric)},
		{Counter, "otel_metrics_sum", plus(metric, column{"AggregationTemporality", "int32"}, column{"IsMonotonic", "bool"})},
		{Histogram, "otel_metrics_histogram", plus(metric, column{"AggregationTemporality", "int32"})},
		{ExponentialHistogram, "otel_metrics_exp_histogram", plus(metric, column{"AggregationTemporality", "int32"})},
	}
}

// Inspect checks all six fixed families when kinds is empty, or only an explicit
// subset. Selection is validated before any source call. Every family is checked
// independently; cancellation/deadline stops the inspection rather than returning
// a successful partial capability list. No result is cached.
func Inspect(ctx context.Context, read Read, kinds ...string) ([]Family, error) {
	if read == nil {
		return nil, fmt.Errorf("source capability reader is required")
	}
	all := specs()
	selected := map[string]bool{}
	for _, kind := range kinds {
		if selected[kind] {
			return nil, fmt.Errorf("duplicate source family")
		}
		found := false
		for _, s := range all {
			if s.kind == kind {
				found = true
				break
			}
		}
		if !found {
			return nil, fmt.Errorf("unsupported source family")
		}
		selected[kind] = true
	}
	out := make([]Family, 0, len(all))
	for _, s := range all {
		if len(kinds) > 0 && !selected[s.kind] {
			continue
		}
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		f, err := inspect(ctx, read, s)
		if err != nil {
			return nil, err
		}
		out = append(out, f)
	}
	return out, nil
}

func inspect(ctx context.Context, read Read, s spec) (Family, error) {
	f := Family{Kind: s.kind, Table: s.table}
	rows, err := read(ctx, "DESCRIBE TABLE "+s.table+" FORMAT JSONEachRow", nil)
	if err := interrupted(ctx, err); err != nil {
		return Family{}, err
	}
	if err != nil {
		return failed(f, err), nil
	}
	columns, ok := decodeColumns(rows)
	if !ok {
		f.State = "unreadable"
		f.Detail = "Source schema response was rejected."
		return f, nil
	}
	for _, required := range s.columns {
		got, exists := columns[required.name]
		if !exists {
			f.State = "incompatible"
			f.Detail = "Required column " + required.name + " is missing."
			return f, nil
		}
		if !matches(got, required.shape) {
			f.State = "incompatible"
			f.Detail = "Required column " + required.name + " has an unsupported type."
			return f, nil
		}
	}
	names := make([]string, len(s.columns))
	for i, c := range s.columns {
		names[i] = "`" + c.name + "`"
	}
	rows, err = read(ctx, "SELECT "+strings.Join(names, ",")+" FROM "+s.table+" LIMIT 0 FORMAT JSONEachRow", nil)
	if err := interrupted(ctx, err); err != nil {
		return Family{}, err
	}
	if err != nil {
		return failed(f, err), nil
	}
	if len(rows) != 0 {
		f.State = "unreadable"
		f.Detail = "Source access response was rejected."
		return f, nil
	}
	f.State = "compatible"
	if s.kind == Logs || s.kind == Spans {
		f.Detail = "Record discovery columns are compatible and readable."
	} else {
		f.Detail = "Instrument discovery columns are compatible and readable; evaluation support was not checked."
	}
	return f, nil
}

func interrupted(ctx context.Context, err error) error {
	if e := ctx.Err(); e != nil {
		return e
	}
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return err
	}
	return nil
}
func failed(f Family, err error) Family {
	f.State = "unreadable"
	f.Detail = "Source schema or read access could not be checked."
	if response, ok := errors.AsType[*sourcehttp.HTTPError](err); ok && (response.Code == 60 || response.Code == 81) {
		f.State = "absent"
		f.Detail = "The source reported that this exporter table or database is absent."
	}
	return f
}
func decodeColumns(rows []json.RawMessage) (map[string]string, bool) {
	if len(rows) == 0 || len(rows) > 256 {
		return nil, false
	}
	out := make(map[string]string, len(rows))
	size := 0
	for _, raw := range rows {
		size += len(raw)
		if size > 1<<20 || len(raw) > 16<<10 {
			return nil, false
		}
		var row struct {
			Name *string `json:"name"`
			Type *string `json:"type"`
		}
		if json.Unmarshal(raw, &row) != nil || row.Name == nil || row.Type == nil || len(*row.Name) == 0 || len(*row.Name) > 256 || len(*row.Type) == 0 || len(*row.Type) > 1024 {
			return nil, false
		}
		if _, duplicate := out[*row.Name]; duplicate {
			return nil, false
		}
		out[*row.Name] = *row.Type
	}
	return out, true
}

// canonicalType parses the small metadata type grammar without accepting
// punctuation from the server as SQL. LowCardinality is representation-neutral
// for the adapter's String/Map/Array projections; Nullable is not.
func canonicalType(input string) (string, bool) {
	text := strings.TrimSpace(input)
	index := 0
	var parse func(int) (string, bool)
	skip := func() {
		for index < len(text) && (text[index] == ' ' || text[index] == '\t') {
			index++
		}
	}
	parse = func(depth int) (string, bool) {
		if depth > 12 {
			return "", false
		}
		skip()
		start := index
		for index < len(text) && ((text[index] >= 'a' && text[index] <= 'z') || (text[index] >= 'A' && text[index] <= 'Z') || text[index] == '_' || (text[index] >= '0' && text[index] <= '9')) {
			index++
		}
		if start == index {
			return "", false
		}
		name := text[start:index]
		skip()
		if index == len(text) || text[index] != '(' {
			return name, true
		}
		index++
		args := []string{}
		for {
			skip()
			var value string
			if index < len(text) && text[index] == '\'' {
				begin := index
				index++
				for index < len(text) && text[index] != '\'' {
					if text[index] == '\\' {
						return "", false
					}
					index++
				}
				if index == len(text) {
					return "", false
				}
				index++
				value = text[begin:index]
			} else {
				var ok bool
				value, ok = parse(depth + 1)
				if !ok {
					return "", false
				}
			}
			args = append(args, value)
			skip()
			if index >= len(text) {
				return "", false
			}
			separator := text[index]
			index++
			if separator == ')' {
				break
			}
			if separator != ',' {
				return "", false
			}
		}
		if name == "LowCardinality" {
			if len(args) != 1 {
				return "", false
			}
			return args[0], true
		}
		return name + "(" + strings.Join(args, ",") + ")", true
	}
	got, ok := parse(0)
	skip()
	return got, ok && index == len(text)
}
func isTime(t string) bool {
	if !strings.HasPrefix(t, "DateTime64(") || !strings.HasSuffix(t, ")") {
		return false
	}
	args := strings.Split(strings.TrimSuffix(strings.TrimPrefix(t, "DateTime64("), ")"), ",")
	if len(args) < 1 || len(args) > 2 {
		return false
	}
	precision, err := strconv.Atoi(args[0])
	if err != nil || precision < 0 || precision > 9 {
		return false
	}
	return len(args) == 1 || (len(args[1]) >= 3 && args[1][0] == '\'' && args[1][len(args[1])-1] == '\'')
}
func matches(raw, shape string) bool {
	t, ok := canonicalType(raw)
	if !ok {
		return false
	}
	switch shape {
	case "string":
		return t == "String"
	case "map":
		return t == "Map(String,String)"
	case "time":
		return isTime(t)
	case "times":
		return strings.HasPrefix(t, "Array(") && strings.HasSuffix(t, ")") && isTime(t[6:len(t)-1])
	case "strings":
		return t == "Array(String)"
	case "maps":
		return t == "Array(Map(String,String))"
	case "uint64":
		return t == "UInt64"
	case "uint8":
		return t == "UInt8"
	case "int32":
		return t == "Int32"
	case "bool":
		return t == "Bool" || t == "Boolean"
	default:
		return false
	}
}
