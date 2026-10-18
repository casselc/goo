// Package exploration implements bounded queries over captured OTel records.
// It does not interpret missing records as application health or completeness.
package exploration

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"
)

type Field struct {
	Kind string `json:"kind" yaml:"kind"`
	Key  string `json:"key" yaml:"key"`
}
type Predicate struct {
	All   []Predicate `json:"all,omitempty" yaml:"all,omitempty"`
	Any   []Predicate `json:"any,omitempty" yaml:"any,omitempty"`
	Field *Field      `json:"field,omitempty" yaml:"field,omitempty"`
	Op    string      `json:"op,omitempty" yaml:"op,omitempty"`
	Value string      `json:"value,omitempty" yaml:"value,omitempty"`
}
type Query struct {
	GroupTimeSeries bool       `json:"groupTimeSeries,omitempty" yaml:"groupTimeSeries,omitempty"`
	SpanStatus      bool       `json:"spanStatus,omitempty" yaml:"spanStatus,omitempty"`
	BucketMS        int64      `json:"bucketMs,omitempty" yaml:"bucketMs,omitempty"`
	Summary         string     `json:"summary,omitempty" yaml:"summary,omitempty"`
	Signal          string     `json:"signal" yaml:"signal"`
	Environment     string     `json:"environment" yaml:"environment"`
	FromMS          int64      `json:"fromMs" yaml:"fromMs"`
	ToMS            int64      `json:"toMs" yaml:"toMs"`
	Where           *Predicate `json:"where,omitempty" yaml:"where,omitempty"`
	GroupBy         []Field    `json:"groupBy,omitempty" yaml:"groupBy,omitempty"`
	Limit           int        `json:"limit" yaml:"limit"`
	Offset          int        `json:"offset" yaml:"offset"`
}

var columns = map[string]string{"service": "string", "trace_id": "string", "span_id": "string", "scope": "string", "scope_version": "string", "name": "string", "kind": "string", "status": "string", "duration_ns": "uint64", "duration_bucket": "string", "status_class": "string", "body": "string", "severity": "string"}

func (f Field) Type(signal string) (string, error) {
	if f.Kind == "column" {
		typ, ok := columns[f.Key]
		if !ok || signal == "logs" && (f.Key == "name" || f.Key == "kind" || f.Key == "status" || f.Key == "duration_ns" || f.Key == "duration_bucket" || f.Key == "status_class") || signal == "spans" && (f.Key == "body" || f.Key == "severity") {
			return "", fmt.Errorf("unsupported %s field %s", signal, f.Key)
		}
		return typ, nil
	}
	if (f.Kind != "resource" && f.Kind != "attribute") || !utf8.ValidString(f.Key) || len(f.Key) == 0 || len(f.Key) > 128 || strings.ContainsRune(f.Key, 0) {
		return "", fmt.Errorf("invalid attribute field")
	}
	return "string", nil
}
func (q Query) Validate() error {
	if q.GroupTimeSeries && (q.BucketMS == 0 || len(q.GroupBy) == 0) {
		return fmt.Errorf("split chart requires a time bucket and breakdown; at most 8 series and 1000 observed cells")
	}

	if q.SpanStatus && q.Signal != "spans" {
		return fmt.Errorf("spanStatus requires spans")
	}
	if q.BucketMS != 0 {
		if q.BucketMS != 1000 && q.BucketMS != 10000 && q.BucketMS != 60000 && q.BucketMS != 300000 {
			return fmt.Errorf("bucketMs must be 1000, 10000, 60000 or 300000")
		}
		if q.ToMS > q.FromMS && (q.ToMS-q.FromMS+q.BucketMS-1)/q.BucketMS > 240 {
			return fmt.Errorf("time chart exceeds 240 intervals; increase bucket width or narrow window")
		}
	}

	if q.Summary != "" && (q.Summary != "span-duration" || q.Signal != "spans") {
		return fmt.Errorf("summary must be span-duration over spans")
	}
	if q.Signal != "spans" && q.Signal != "logs" {
		return fmt.Errorf("signal must be spans or logs")
	}
	if q.Environment == "" || len(q.Environment) > 128 || !utf8.ValidString(q.Environment) || strings.ContainsRune(q.Environment, 0) {
		return fmt.Errorf("explicit environment required")
	}
	if q.FromMS < 0 || q.ToMS <= q.FromMS || q.ToMS-q.FromMS > 120*60000 || q.ToMS > 4102444800000 {
		return fmt.Errorf("require half-open window of at most 120 minutes before 2100")
	}
	if q.Limit < 1 || q.Limit > 100 || q.Offset < 0 || q.Offset > 1000 {
		return fmt.Errorf("page limit 1..100 and offset 0..1000 required")
	}
	if len(q.GroupBy) > 2 {
		return fmt.Errorf("at most two grouping fields")
	}
	seen := map[Field]bool{}
	for _, f := range q.GroupBy {
		if _, e := f.Type(q.Signal); e != nil {
			return e
		}
		if seen[f] {
			return fmt.Errorf("duplicate grouping field")
		}
		seen[f] = true
	}
	count := 0
	return validatePredicate(q.Where, q.Signal, 0, &count)
}
func validatePredicate(p *Predicate, signal string, depth int, count *int) error {
	if p == nil {
		return nil
	}
	*count++
	if depth > 4 || *count > 32 {
		return fmt.Errorf("predicate exceeds depth/node budget")
	}
	branches := p.All != nil || p.Any != nil
	if branches {
		if p.All != nil && p.Any != nil || len(p.All) == 0 && len(p.Any) == 0 || p.Field != nil || p.Op != "" || p.Value != "" {
			return fmt.Errorf("predicate must be one branch or leaf")
		}
		items := p.All
		if len(p.Any) > 0 {
			items = p.Any
		}
		for _, child := range items {
			if e := validatePredicate(&child, signal, depth+1, count); e != nil {
				return e
			}
		}
		return nil
	}
	if p.Field == nil {
		return fmt.Errorf("predicate leaf requires field")
	}
	typ, e := p.Field.Type(signal)
	if e != nil {
		return e
	}
	if len(p.Value) > 1024 || !utf8.ValidString(p.Value) || strings.ContainsRune(p.Value, 0) {
		return fmt.Errorf("invalid predicate value")
	}
	switch p.Op {
	case "exists", "missing":
		if p.Value != "" {
			return fmt.Errorf("existence predicate has no value")
		}
	case "eq", "ne":
	case "contains":
		if typ != "string" {
			return fmt.Errorf("contains requires string")
		}
	case "gte", "lte":
		if typ != "uint64" {
			return fmt.Errorf("ordering requires uint64 field")
		}
	default:
		return fmt.Errorf("unsupported predicate operator")
	}
	if typ == "uint64" && p.Op != "exists" && p.Op != "missing" {
		if _, e := strconv.ParseUint(p.Value, 10, 64); e != nil {
			return fmt.Errorf("require uint64 decimal")
		}
	}
	return nil
}

// Match is the independent reference predicate interpreter. Missing != empty and
// ne never matches missing fields; an explicit missing predicate is required.
// The logical record must include all fixed columns; only map attributes may
// be absent in a valid bound source row.
func Match(p *Predicate, values map[Field]string) bool {
	if p == nil {
		return true
	}
	if len(p.All) > 0 {
		for i := range p.All {
			if !Match(&p.All[i], values) {
				return false
			}
		}
		return true
	}
	if len(p.Any) > 0 {
		for i := range p.Any {
			if Match(&p.Any[i], values) {
				return true
			}
		}
		return false
	}
	if p.Field == nil {
		return false
	}
	v, exists := values[*p.Field]
	switch p.Op {
	case "exists":
		return exists
	case "missing":
		return !exists
	}
	if !exists {
		return false
	}
	if p.Field.Kind == "column" && p.Field.Key == "duration_ns" && (p.Op == "eq" || p.Op == "ne") {
		a, e := strconv.ParseUint(v, 10, 64)
		b, e2 := strconv.ParseUint(p.Value, 10, 64)
		return e == nil && e2 == nil && (p.Op == "eq" && a == b || p.Op == "ne" && a != b)
	}
	switch p.Op {
	case "eq":
		return v == p.Value
	case "ne":
		return v != p.Value
	case "contains":
		return strings.Contains(v, p.Value)
	case "gte", "lte":
		a, e := strconv.ParseUint(v, 10, 64)
		b, e2 := strconv.ParseUint(p.Value, 10, 64)
		return e == nil && e2 == nil && (p.Op == "gte" && a >= b || p.Op == "lte" && a <= b)
	}
	return false
}

var hexTrace = regexp.MustCompile(`^[0-9a-f]{32}$`)
var hexSpan = regexp.MustCompile(`^[0-9a-f]{16}$`)

func TraceQuestion(environment string, from, to int64, trace, span, signal string, offset int) (Query, error) {
	q := Query{Signal: signal, Environment: environment, FromMS: from, ToMS: to, Limit: 100, Offset: offset}
	if !hexTrace.MatchString(trace) || strings.Trim(trace, "0") == "" {
		return q, fmt.Errorf("valid nonzero trace ID required")
	}
	predicates := []Predicate{{Field: &Field{"column", "trace_id"}, Op: "eq", Value: trace}}
	if span != "" {
		if !hexSpan.MatchString(span) || strings.Trim(span, "0") == "" {
			return q, fmt.Errorf("valid nonzero span ID required")
		}
		predicates = append(predicates, Predicate{Field: &Field{"column", "span_id"}, Op: "eq", Value: span})
	}
	q.Where = &Predicate{All: predicates}
	return q, q.Validate()
}
