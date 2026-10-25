package verification

import (
	"encoding/json"
	"math/big"
	"sort"
	"strconv"
	"strings"

	"github.com/casselc/goo/exploration"
)

type tsRecord struct {
	Environment string
	Values      map[exploration.Field]string
}

func tsInteger(raw string, typ string) *big.Int {
	if raw == "" || raw[0] == '+' {
		return nil
	}
	negative := raw[0] == '-'
	digits := raw
	if negative {
		if typ != "int64" {
			return nil
		}
		digits = raw[1:]
	}
	if digits == "" || (len(digits) > 1 && digits[0] == '0') || negative && digits == "0" {
		return nil
	}
	for _, c := range digits {
		if c < '0' || c > '9' {
			return nil
		}
	}
	n, ok := new(big.Int).SetString(raw, 10)
	if !ok {
		return nil
	}
	if typ == "int64" {
		lo := new(big.Int).Neg(new(big.Int).Lsh(big.NewInt(1), 63))
		hi := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 63), big.NewInt(1))
		if n.Cmp(lo) < 0 || n.Cmp(hi) > 0 {
			return nil
		}
	} else if typ != "uint64" || n.Sign() < 0 || n.BitLen() > 64 {
		return nil
	}
	return n
}
func tsNumeric(raw *string, b exploration.NumericBinding) (*big.Int, string) {
	if b.Type == "decimal64" {
		return decCoefficient(raw, b.Scale)
	}
	if raw == nil {
		return nil, "missing"
	}
	if *raw == "" {
		return nil, "empty"
	}
	n := tsInteger(*raw, b.Type)
	if n == nil {
		return nil, "invalid"
	}
	return n, "valid"
}
func tsPredicate(p exploration.Predicate, values map[exploration.Field]string) bool {
	if p.All != nil {
		if len(p.All) == 0 {
			return false
		}
		for _, child := range p.All {
			if !tsPredicate(child, values) {
				return false
			}
		}
		return true
	}
	if p.Any != nil {
		for _, child := range p.Any {
			if tsPredicate(child, values) {
				return true
			}
		}
		return false
	}
	if p.Numeric != nil {
		text, present := values[p.Numeric.Field]
		var raw *string
		if present {
			raw = &text
		}
		n, state := tsNumeric(raw, *p.Numeric)
		switch p.Op {
		case "missing", "empty", "invalid", "valid":
			return state == p.Op
		}
		threshold, tstate := tsNumeric(&p.Value, *p.Numeric)
		if state != "valid" || tstate != "valid" {
			return false
		}
		cmp := n.Cmp(threshold)
		switch p.Op {
		case "eq":
			return cmp == 0
		case "ne":
			return cmp != 0
		case "gt":
			return cmp > 0
		case "gte":
			return cmp >= 0
		case "lt":
			return cmp < 0
		case "lte":
			return cmp <= 0
		}
		return false
	}
	if p.Field == nil {
		return false
	}
	text, present := values[*p.Field]
	if p.Op == "exists" {
		return present
	}
	if p.Op == "missing" {
		return !present
	}
	if !present {
		return false
	}
	if p.Field.Kind == "column" && (p.Field.Key == "timestamp_ns" || p.Field.Key == "duration_ns" || p.Field.Key == "severity_number") {
		a, ae := strconv.ParseUint(text, 10, 64)
		b, be := strconv.ParseUint(p.Value, 10, 64)
		if ae != nil || be != nil {
			return false
		}
		switch p.Op {
		case "eq":
			return a == b
		case "ne":
			return a != b
		case "gte":
			return a >= b
		case "lte":
			return a <= b
		}
		return false
	}
	switch p.Op {
	case "eq":
		return text == p.Value
	case "ne":
		return text != p.Value
	case "contains":
		return strings.Contains(text, p.Value)
	}
	return false
}
func tsFields(p exploration.Predicate) []exploration.Field {
	fields := map[exploration.Field]bool{}
	var visit func(exploration.Predicate)
	visit = func(p exploration.Predicate) {
		if p.Field != nil {
			fields[*p.Field] = true
		}
		if p.Numeric != nil {
			fields[p.Numeric.Field] = true
		}
		for _, c := range p.All {
			visit(c)
		}
		for _, c := range p.Any {
			visit(c)
		}
	}
	visit(p)
	out := []exploration.Field{}
	for f := range fields {
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

type tsFieldValue struct {
	Field exploration.Field `json:"field"`
	Value *string           `json:"value"`
}
type tsWitness struct {
	ConditionIndex  int            `json:"conditionIndex"`
	MatchingRecords string         `json:"matchingRecords"`
	SpanID          string         `json:"spanId"`
	ParentSpanID    string         `json:"parentSpanId"`
	Service         string         `json:"service"`
	Operation       string         `json:"operation"`
	Kind            string         `json:"kind"`
	TimestampNS     string         `json:"timestampNs"`
	DurationNS      string         `json:"durationNs"`
	FieldValues     []tsFieldValue `json:"fieldValues"`
}
type tsItem struct {
	TraceID   string      `json:"traceId"`
	Witnesses []tsWitness `json:"witnesses"`
}

func tsWitnessFor(r tsRecord, condition exploration.Predicate, index int, count int) tsWitness {
	v := func(key string) string { return r.Values[exploration.Field{Kind: "column", Key: key}] }
	w := tsWitness{ConditionIndex: index, MatchingRecords: strconv.Itoa(count), SpanID: v("span_id"), ParentSpanID: v("parent_span_id"), Service: v("service"), Operation: v("name"), Kind: v("kind"), TimestampNS: v("timestamp_ns"), DurationNS: v("duration_ns"), FieldValues: []tsFieldValue{}}
	for _, field := range tsFields(condition) {
		var value *string
		if raw, ok := r.Values[field]; ok {
			copy := raw
			value = &copy
		}
		w.FieldValues = append(w.FieldValues, tsFieldValue{Field: field, Value: value})
	}
	return w
}
func tsWitnessOrder(a, b tsWitness) bool {
	x, y := tsInteger(a.TimestampNS, "uint64"), tsInteger(b.TimestampNS, "uint64")
	if c := x.Cmp(y); c != 0 {
		return c < 0
	}
	left, right := []string{a.SpanID, a.ParentSpanID, a.Service, a.Operation, a.Kind}, []string{b.SpanID, b.ParentSpanID, b.Service, b.Operation, b.Kind}
	for i := range left {
		if left[i] != right[i] {
			return left[i] < right[i]
		}
	}
	if c := tsInteger(a.DurationNS, "uint64").Cmp(tsInteger(b.DurationNS, "uint64")); c != 0 {
		return c < 0
	}
	for i, v := range a.FieldValues {
		w := b.FieldValues[i]
		if c := tcOptionalCompare(v.Value, w.Value); c != 0 {
			return c < 0
		}
	}
	return false
}
func tsInScope(request exploration.TraceSearchRequest, r tsRecord) bool {
	id := r.Values[exploration.Field{Kind: "column", Key: "trace_id"}]
	start := tsInteger(r.Values[exploration.Field{Kind: "column", Key: "timestamp_ns"}], "uint64")
	if r.Environment != request.Environment || !tcID(id, 32) || start == nil {
		return false
	}
	lo := new(big.Int).Mul(big.NewInt(request.FromMS), big.NewInt(1000000))
	hi := new(big.Int).Mul(big.NewInt(request.ToMS), big.NewInt(1000000))
	return start.Cmp(lo) >= 0 && start.Cmp(hi) < 0
}
func tsOracle(request exploration.TraceSearchRequest, records []tsRecord) ([]tsItem, string) {
	// Build each existential set independently, then intersect sets. Witnesses
	// are derived afterward from all matching records, independent of list limit.
	sets := make([]map[string]bool, len(request.Conditions))
	for i, p := range request.Conditions {
		sets[i] = map[string]bool{}
		for _, r := range records {
			if tsInScope(request, r) && tsPredicate(p, r.Values) {
				sets[i][r.Values[exploration.Field{Kind: "column", Key: "trace_id"}]] = true
			}
		}
	}
	ids := []string{}
	if len(sets) > 0 {
		for id := range sets[0] {
			present := true
			for _, s := range sets[1:] {
				present = present && s[id]
			}
			if present {
				ids = append(ids, id)
			}
		}
	}
	sort.Strings(ids)
	total := strconv.Itoa(len(ids))
	if len(ids) > request.Limit {
		ids = ids[:request.Limit]
	}
	items := []tsItem{}
	for _, id := range ids {
		item := tsItem{TraceID: id, Witnesses: []tsWitness{}}
		for index, p := range request.Conditions {
			matches := []tsWitness{}
			for _, r := range records {
				if tsInScope(request, r) && r.Values[exploration.Field{Kind: "column", Key: "trace_id"}] == id && tsPredicate(p, r.Values) {
					matches = append(matches, tsWitnessFor(r, p, index, 0))
				}
			}
			sort.SliceStable(matches, func(i, j int) bool { return tsWitnessOrder(matches[i], matches[j]) })
			w := matches[0]
			w.MatchingRecords = strconv.Itoa(len(matches))
			item.Witnesses = append(item.Witnesses, w)
		}
		items = append(items, item)
	}
	return items, total
}
func tsLogical(records []tsRecord) []exploration.TraceSearchRecord {
	out := make([]exploration.TraceSearchRecord, len(records))
	for i, r := range records {
		values := map[exploration.Field]string{}
		for k, v := range r.Values {
			values[k] = v
		}
		out[i] = exploration.TraceSearchRecord{Environment: r.Environment, Fields: values}
	}
	return out
}
func tsItemsEqual(items []exploration.TraceSearchItem, want []tsItem) bool {
	a, _ := json.Marshal(items)
	b, _ := json.Marshal(want)
	return aeJSONEqual(json.RawMessage(a), json.RawMessage(b))
}

func tsSnapshot(records []tsRecord) string {
	type value struct {
		Environment string
		Fields      []tsFieldValue
	}
	out := []value{}
	for _, r := range records {
		keys := []exploration.Field{}
		for f := range r.Values {
			keys = append(keys, f)
		}
		sort.Slice(keys, func(i, j int) bool {
			if keys[i].Kind != keys[j].Kind {
				return keys[i].Kind < keys[j].Kind
			}
			return keys[i].Key < keys[j].Key
		})
		v := value{Environment: r.Environment, Fields: []tsFieldValue{}}
		for _, f := range keys {
			copy := r.Values[f]
			v.Fields = append(v.Fields, tsFieldValue{Field: f, Value: &copy})
		}
		out = append(out, v)
	}
	raw, _ := json.Marshal(out)
	return string(raw)
}
