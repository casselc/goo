package exploration

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"sort"
	"strconv"
	"strings"
	"time"
)

const TraceSearchVersion = "goo/trace-search/v1"

type TraceSearchRequest struct {
	Version     string      `json:"version"`
	Environment string      `json:"environment"`
	FromMS      int64       `json:"fromMs"`
	ToMS        int64       `json:"toMs"`
	Conditions  []Predicate `json:"conditions"`
	Limit       int         `json:"limit"`
}
type TraceSearchPlan struct {
	SQL        string            `json:"sql"`
	Parameters map[string]string `json:"parameters"`
}
type TraceSearchFieldValue struct {
	Field Field   `json:"field"`
	Value *string `json:"value"`
}
type TraceSearchWitness struct {
	FieldValues     []TraceSearchFieldValue `json:"fieldValues"`
	ConditionIndex  int                     `json:"conditionIndex"`
	MatchingRecords string                  `json:"matchingRecords"`
	SpanID          string                  `json:"spanId"`
	ParentSpanID    string                  `json:"parentSpanId"`
	Service         string                  `json:"service"`
	Operation       string                  `json:"operation"`
	Kind            string                  `json:"kind"`
	TimestampNS     string                  `json:"timestampNs"`
	DurationNS      string                  `json:"durationNs"`
}
type TraceSearchItem struct {
	TraceID   string               `json:"traceId"`
	Witnesses []TraceSearchWitness `json:"witnesses"`
}
type TraceSearchResult struct {
	Request       TraceSearchRequest `json:"request"`
	ReadAt        string             `json:"readAt"`
	Basis         string             `json:"basis"`
	State         string             `json:"state"`
	MatchedTraces string             `json:"matchedTraces"`
	HasMore       bool               `json:"hasMore"`
	Items         []TraceSearchItem  `json:"items"`
	Plan          TraceSearchPlan    `json:"plan"`
	Warnings      []string           `json:"warnings"`
}

func (r TraceSearchRequest) Validate() error {
	if r.Version != TraceSearchVersion || len(r.Conditions) < 1 || len(r.Conditions) > 4 || r.Limit < 1 || r.Limit > 100 {
		return fmt.Errorf("trace search requires v1, one to four conditions and limit 1..100")
	}
	nodes := 0
	var count func(Predicate)
	count = func(p Predicate) {
		nodes++
		for _, c := range p.All {
			count(c)
		}
		for _, c := range p.Any {
			count(c)
		}
	}
	for i := range r.Conditions {
		q := Query{Signal: "spans", Environment: r.Environment, FromMS: r.FromMS, ToMS: r.ToMS, Where: &r.Conditions[i], Limit: 1}
		if e := q.Validate(); e != nil {
			return fmt.Errorf("condition %d: %w", i, e)
		}
		count(r.Conditions[i])
	}
	if nodes > 64 {
		return fmt.Errorf("trace search exceeds 64 predicate nodes")
	}
	return nil
}
func CompileTraceSearch(r TraceSearchRequest) (TraceSearchPlan, error) {
	if e := r.Validate(); e != nil {
		return TraceSearchPlan{}, e
	}
	c := compiler{signal: "spans", params: map[string]string{"from": strconv.FormatInt(r.FromMS, 10), "to": strconv.FormatInt(r.ToMS, 10), "environment": r.Environment}}
	base, _ := c.source(Query{Signal: "spans"})
	base += " AND length(TraceId) = 32 AND match(TraceId, '^[0-9a-f]{32}$') AND TraceId != '00000000000000000000000000000000'"
	columns := []string{"TraceId AS trace_id", "toString(count() OVER ()) AS matched_traces"}
	having := []string{}
	for i, p := range r.Conditions {
		condition := c.predicate(&p)
		n := fmt.Sprint(i)
		columns = append(columns, "toString(countIf("+condition+")) AS c"+n)
		having = append(having, "countIf("+condition+") > 0")
		fields := traceSearchFields(p)
		values := []string{}
		for _, field := range fields {
			v, present := c.field(field)
			values = append(values, "tuple(toUInt8("+present+"), toString("+v+"))")
		}
		tuple := "tuple(toString(toUnixTimestamp64Nano(Timestamp)), SpanId, ParentSpanId, ServiceName, SpanName, SpanKind, toString(Duration), [" + strings.Join(values, ", ") + "])"
		order := "tuple(Timestamp, SpanId, ParentSpanId, ServiceName, SpanName, SpanKind, Duration, [" + strings.Join(values, ", ") + "])"
		columns = append(columns, "argMinIf("+tuple+", "+order+", "+condition+") AS w"+n)
	}
	sql := "SELECT " + strings.Join(columns, ", ") + base + " GROUP BY TraceId HAVING " + strings.Join(having, " AND ") + " ORDER BY TraceId LIMIT " + strconv.Itoa(r.Limit+1) + " FORMAT JSONEachRow"
	if len(sql) > 256<<10 {
		return TraceSearchPlan{}, fmt.Errorf("trace search compiled SQL exceeds 256 KiB; simplify conditions")
	}
	return TraceSearchPlan{SQL: sql, Parameters: c.params}, nil
}
func newTraceSearchResult(r TraceSearchRequest, plan TraceSearchPlan) *TraceSearchResult {
	return &TraceSearchResult{Request: detached(r), State: "no_data", MatchedTraces: "0", Items: []TraceSearchItem{}, Plan: plan, Warnings: []string{"Distinct valid reported trace IDs within this environment and fixed window; not unique invocations or complete traces.", "Conditions are independent EXISTS tests. One observed span may satisfy several conditions; no ordering, parent topology or causal relationship is implied.", "Each witness is one representative matching record, not every matching span. Duplicate records remain in matching-record counts.", "This is one source aggregate read. Sampling, late arrivals, source coverage and records outside the fixed window remain unknown."}}
}
func (s Source) SearchTraces(ctx context.Context, r TraceSearchRequest) (*TraceSearchResult, error) {
	plan, e := CompileTraceSearch(r)
	if e != nil {
		return nil, e
	}
	rows, e := s.read(ctx, plan.SQL, plan.Parameters)
	if e != nil {
		return nil, e
	}
	out := newTraceSearchResult(r, plan)
	if len(rows) > r.Limit+1 {
		return nil, fmt.Errorf("trace search source row bound exceeded")
	}
	last := ""
	total := ""
	for _, raw := range rows {
		var object map[string]json.RawMessage
		if json.Unmarshal(raw, &object) != nil || object == nil {
			return nil, fmt.Errorf("invalid trace search source row")
		}
		var id, n string
		if json.Unmarshal(object["trace_id"], &id) != nil || json.Unmarshal(object["matched_traces"], &n) != nil || !exampleID(id, 32) || id <= last {
			return nil, fmt.Errorf("invalid trace search source identity")
		}
		count, e := decimal(n)
		if e != nil || count.Sign() == 0 || total != "" && total != n {
			return nil, fmt.Errorf("invalid trace search source total")
		}
		last, total = id, n
		item := TraceSearchItem{TraceID: id, Witnesses: []TraceSearchWitness{}}
		for i := range r.Conditions {
			var count string
			var rawTuple []json.RawMessage
			var tuple []string
			if json.Unmarshal(object["c"+fmt.Sprint(i)], &count) != nil || json.Unmarshal(object["w"+fmt.Sprint(i)], &rawTuple) != nil || len(rawTuple) != 8 {
				return nil, fmt.Errorf("invalid trace search source witness")
			}
			for _, rawValue := range rawTuple[:7] {
				var value string
				if !traceSearchString(rawValue, &value) {
					return nil, fmt.Errorf("invalid trace search witness value")
				}
				tuple = append(tuple, value)
			}
			c, e := decimal(count)
			if e != nil || c.Sign() == 0 {
				return nil, fmt.Errorf("invalid trace search matching count")
			}
			if timestamp, ok := canonicalNS(tuple[0]); !ok || timestamp < uint64(r.FromMS)*1000000 || timestamp >= uint64(r.ToMS)*1000000 {
				return nil, fmt.Errorf("invalid trace search witness timestamp")
			}
			if _, ok := canonicalNS(tuple[6]); !ok {
				return nil, fmt.Errorf("invalid trace search witness duration")
			}
			for _, v := range tuple {
				if len(v) > 4096 {
					return nil, fmt.Errorf("trace search witness field bound exceeded")
				}
			}
			fields, e := decodeTraceSearchValues(rawTuple[7], r.Conditions[i])
			if e != nil {
				return nil, e
			}
			expectedColumns := map[string]string{"service": tuple[3], "name": tuple[4], "kind": tuple[5], "duration_ns": tuple[6], "timestamp_ns": tuple[0], "span_id": tuple[1], "trace_id": id}
			for _, fieldValue := range fields {
				if fieldValue.Field.Kind == "column" {
					if expected, known := expectedColumns[fieldValue.Field.Key]; known && (fieldValue.Value == nil || *fieldValue.Value != expected) {
						return nil, fmt.Errorf("trace witness field and record disagree")
					}
				}
				if fieldValue.Field == (Field{Kind: "resource", Key: "deployment.environment.name"}) && (fieldValue.Value == nil || *fieldValue.Value != r.Environment) {
					return nil, fmt.Errorf("trace witness environment disagrees")
				}
			}
			item.Witnesses = append(item.Witnesses, TraceSearchWitness{ConditionIndex: i, MatchingRecords: count, SpanID: tuple[1], ParentSpanID: tuple[2], Service: tuple[3], Operation: tuple[4], Kind: tuple[5], TimestampNS: tuple[0], DurationNS: tuple[6], FieldValues: fields})
		}
		out.Items = append(out.Items, item)
	}
	if len(rows) > 0 {
		n, _ := decimal(total)
		if n.Cmp(big.NewInt(int64(len(rows)))) < 0 || len(rows) < r.Limit+1 && n.Cmp(big.NewInt(int64(len(rows)))) != 0 {
			return nil, fmt.Errorf("trace search source list and total disagree")
		}
		out.MatchedTraces = total
		out.State = "observed"
	}
	out.HasMore = len(out.Items) > r.Limit
	if out.HasMore {
		out.Items = out.Items[:r.Limit]
	}
	out.ReadAt = time.Now().UTC().Format(time.RFC3339Nano)
	material, _ := json.Marshal(struct {
		Request TraceSearchRequest
		Rows    []json.RawMessage
	}{r, rows})
	sum := sha256.Sum256(material)
	out.Basis = hex.EncodeToString(sum[:])
	encoded, e := json.Marshal(out)
	if e != nil || len(encoded) > 8<<20 {
		return nil, fmt.Errorf("trace search response exceeds 8 MiB; narrow conditions or limit")
	}
	return out, nil
}

// TraceSearchRecord is a logical captured OTel record for independent reference
// execution. Fixed columns and map keys use the ordinary Predicate field model.
type TraceSearchRecord struct {
	Environment string
	Fields      map[Field]string
}

func ReferenceTraceSearch(r TraceSearchRequest, records []TraceSearchRecord) ([]TraceSearchItem, string, error) {
	if e := r.Validate(); e != nil {
		return nil, "", e
	}
	groups := map[string][][]TraceSearchWitness{}
	for _, record := range records {
		v := record.Fields
		id := v[Field{Kind: "column", Key: "trace_id"}]
		ts := v[Field{Kind: "column", Key: "timestamp_ns"}]
		timeNS, ok := canonicalNS(ts)
		if record.Environment != r.Environment || !exampleID(id, 32) || !ok || timeNS < uint64(r.FromMS)*1000000 || timeNS >= uint64(r.ToMS)*1000000 {
			continue
		}
		if _, ok := canonicalNS(v[Field{Kind: "column", Key: "duration_ns"}]); !ok {
			return nil, "", fmt.Errorf("reference record requires canonical duration_ns")
		}
		if groups[id] == nil {
			groups[id] = make([][]TraceSearchWitness, len(r.Conditions))
		}
		for i := range r.Conditions {
			if Match(&r.Conditions[i], v) {
				value := func(key string) string { return v[Field{Kind: "column", Key: key}] }
				groups[id][i] = append(groups[id][i], TraceSearchWitness{ConditionIndex: i, SpanID: value("span_id"), ParentSpanID: value("parent_span_id"), Service: value("service"), Operation: value("name"), Kind: value("kind"), TimestampNS: ts, DurationNS: value("duration_ns"), FieldValues: referenceTraceSearchValues(r.Conditions[i], v)})
			}
		}
	}
	items := []TraceSearchItem{}
	for id, conditions := range groups {
		item := TraceSearchItem{TraceID: id, Witnesses: []TraceSearchWitness{}}
		valid := true
		for _, witnesses := range conditions {
			if len(witnesses) == 0 {
				valid = false
				break
			}
			sort.Slice(witnesses, func(i, j int) bool { return traceWitnessLess(witnesses[i], witnesses[j]) })
			w := witnesses[0]
			w.MatchingRecords = strconv.Itoa(len(witnesses))
			item.Witnesses = append(item.Witnesses, w)
		}
		if valid {
			items = append(items, item)
		}
	}
	sort.Slice(items, func(i, j int) bool { return items[i].TraceID < items[j].TraceID })
	total := strconv.Itoa(len(items))
	if len(items) > r.Limit {
		items = items[:r.Limit]
	}
	return items, total, nil
}
func traceWitnessLess(a, b TraceSearchWitness) bool {
	at, _ := canonicalNS(a.TimestampNS)
	bt, _ := canonicalNS(b.TimestampNS)
	if at != bt {
		return at < bt
	}
	x, y := []string{a.SpanID, a.ParentSpanID, a.Service, a.Operation, a.Kind}, []string{b.SpanID, b.ParentSpanID, b.Service, b.Operation, b.Kind}
	for i := range x {
		if x[i] != y[i] {
			return x[i] < y[i]
		}
	}
	ad, _ := canonicalNS(a.DurationNS)
	bd, _ := canonicalNS(b.DurationNS)
	if ad != bd {
		return ad < bd
	}
	for i := range a.FieldValues {
		x, y := a.FieldValues[i].Value, b.FieldValues[i].Value
		if x == nil && y != nil {
			return true
		}
		if x != nil && y == nil {
			return false
		}
		if x != nil && *x != *y {
			return *x < *y
		}
	}
	return false
}

func traceSearchFields(p Predicate) []Field {
	set := map[Field]bool{}
	var walk func(Predicate)
	walk = func(p Predicate) {
		if p.Field != nil {
			set[*p.Field] = true
		}
		if p.Numeric != nil {
			set[p.Numeric.Field] = true
		}
		for _, c := range p.All {
			walk(c)
		}
		for _, c := range p.Any {
			walk(c)
		}
	}
	walk(p)
	out := []Field{}
	for f := range set {
		out = append(out, f)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Kind != out[j].Kind {
			return out[i].Kind < out[j].Kind
		}
		return out[i].Key < out[j].Key
	})
	return out
}
func referenceTraceSearchValues(p Predicate, values map[Field]string) []TraceSearchFieldValue {
	out := []TraceSearchFieldValue{}
	for _, f := range traceSearchFields(p) {
		entry := TraceSearchFieldValue{Field: f}
		if value, ok := values[f]; ok {
			v := value
			entry.Value = &v
		}
		out = append(out, entry)
	}
	return out
}
func decodeTraceSearchValues(raw json.RawMessage, p Predicate) ([]TraceSearchFieldValue, error) {
	fields := traceSearchFields(p)
	var rows [][]json.RawMessage
	if json.Unmarshal(raw, &rows) != nil || len(rows) != len(fields) {
		return nil, fmt.Errorf("invalid trace search field evidence")
	}
	out := []TraceSearchFieldValue{}
	values := map[Field]string{}
	for i, row := range rows {
		var present uint8
		var value string
		if len(row) != 2 || json.Unmarshal(row[0], &present) != nil || present > 1 || !traceSearchString(row[1], &value) || len(value) > 4096 {
			return nil, fmt.Errorf("invalid trace search field evidence")
		}
		entry := TraceSearchFieldValue{Field: fields[i]}
		if present == 1 {
			v := value
			entry.Value = &v
			values[fields[i]] = value
		}
		if present == 0 && !fields[i].CanBeMissing() {
			return nil, fmt.Errorf("fixed trace witness field missing")
		}
		out = append(out, entry)
	}
	if !Match(&p, values) {
		return nil, fmt.Errorf("trace witness does not satisfy its condition")
	}
	return out, nil
}

func traceSearchString(raw json.RawMessage, value *string) bool {
	return len(raw) > 0 && string(raw) != "null" && json.Unmarshal(raw, value) == nil
}
