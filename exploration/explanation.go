package exploration

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"reflect"
	"sort"
	"time"
	"unicode/utf8"
)

const ExplanationVersion = "goo/explanation/v1"
const ExplanationFieldMaxBytes = 128 << 10

type ExplanationRequest struct {
	Version   string  `json:"version"`
	Baseline  Query   `json:"baseline"`
	Selection Query   `json:"selection"`
	Fields    []Field `json:"fields"`
}
type ExplanationPlan struct {
	Request   ExplanationRequest `json:"request"`
	Baseline  Query              `json:"baseline"`
	Selection Query              `json:"selection"`
}
type ExplanationPopulation struct {
	Count  string `json:"count"`
	ReadAt string `json:"readAt"`
	Basis  string `json:"basis"`
}
type ExplanationValueCount struct {
	Value *string `json:"value"`
	Count string  `json:"count"`
}
type ExplanationFieldRead struct {
	State  string                  `json:"state"`
	Reason string                  `json:"reason,omitempty"`
	ReadAt string                  `json:"readAt"`
	Basis  string                  `json:"basis,omitempty"`
	Values []ExplanationValueCount `json:"values"`
}
type ExplanationRow struct {
	Value            *string `json:"value"`
	BaselineCount    string  `json:"baselineCount"`
	SelectionCount   string  `json:"selectionCount"`
	BaselinePercent  string  `json:"baselinePercent,omitempty"`
	SelectionPercent string  `json:"selectionPercent,omitempty"`
	DifferencePP     string  `json:"differencePP,omitempty"`
}
type ExplanationFieldResult struct {
	Field     Field                `json:"field"`
	State     string               `json:"state"`
	Reason    string               `json:"reason,omitempty"`
	Baseline  ExplanationFieldRead `json:"baseline"`
	Selection ExplanationFieldRead `json:"selection"`
	Rows      []ExplanationRow     `json:"rows"`
}
type ExplanationLead struct {
	FieldIndex int `json:"fieldIndex"`
	RowIndex   int `json:"rowIndex"`
}
type ExplanationResult struct {
	Version     string                   `json:"version"`
	Request     ExplanationRequest       `json:"request"`
	State       string                   `json:"state"`
	ReadAt      string                   `json:"readAt"`
	Basis       string                   `json:"basis"`
	Baseline    ExplanationPopulation    `json:"baseline"`
	Selection   ExplanationPopulation    `json:"selection"`
	Fields      []ExplanationFieldResult `json:"fields"`
	Ranked      []ExplanationLead        `json:"ranked"`
	Warnings    []string                 `json:"warnings"`
	ResultToken string                   `json:"resultToken,omitempty"`
}

func (r ExplanationRequest) Validate() error {
	if r.Version != ExplanationVersion || len(r.Fields) < 1 || len(r.Fields) > 6 {
		return fmt.Errorf("versioned explanation requires 1..6 explicit fields")
	}
	if e := r.Baseline.Validate(); e != nil {
		return e
	}
	if e := r.Selection.Validate(); e != nil {
		return e
	}
	if r.Baseline.Signal != r.Selection.Signal || r.Baseline.Environment != r.Selection.Environment {
		return fmt.Errorf("explanation requires identical signal and environment; selectors and windows remain independent")
	}
	seen := map[Field]bool{}
	for _, f := range r.Fields {
		if _, e := f.Type(r.Baseline.Signal); e != nil {
			return e
		}
		if f.Kind == "column" && f.Key == "duration_bucket" {
			return fmt.Errorf("duration_bucket is not supported by field explanations")
		}
		if seen[f] {
			return fmt.Errorf("duplicate explanation field")
		}
		seen[f] = true
	}
	return nil
}
func explanationCountQuery(q Query) Query {
	v := detached(q)
	v.Numeric = nil
	v.NumericHistogram = nil
	v.Summary = ""
	v.SpanStatus = false
	v.BucketMS = 0
	v.GroupTimeSeries = false
	v.GroupBy = nil
	v.Offset = 0
	v.Limit = 100
	return v
}
func PlanExplanation(r ExplanationRequest) (*ExplanationPlan, error) {
	if e := r.Validate(); e != nil {
		return nil, e
	}
	return &ExplanationPlan{Request: detached(r), Baseline: explanationCountQuery(r.Baseline), Selection: explanationCountQuery(r.Selection)}, nil
}
func explanationDigest(v any) string {
	raw, _ := json.Marshal(v)
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:])
}
func explanationTime(v string) bool { _, e := time.Parse(time.RFC3339Nano, v); return e == nil }
func explanationPopulationValid(p ExplanationPopulation) bool {
	_, e := decimal(p.Count)
	return e == nil && explanationTime(p.ReadAt) && regexpRevision.MatchString(p.Basis)
}
func explanationReadValid(r ExplanationFieldRead, f Field) error {
	if !explanationTime(r.ReadAt) || r.Basis != "" && !regexpRevision.MatchString(r.Basis) {
		return fmt.Errorf("invalid explanation field read metadata")
	}
	switch r.State {
	case "complete":
		if r.Reason != "" || !regexpRevision.MatchString(r.Basis) || len(r.Values) > 1000 {
			return fmt.Errorf("invalid complete explanation groups")
		}
	case "partial":
		if r.Reason != "cardinality-limit" || !regexpRevision.MatchString(r.Basis) || len(r.Values) != 0 {
			return fmt.Errorf("partial explanation must withhold incomplete groups")
		}
	case "unavailable":
		if r.Reason != "source-read-failed" && r.Reason != "source-timeout" && r.Reason != "invalid-source-groups" && r.Reason != "field-bytes-limit" || len(r.Values) != 0 {
			return fmt.Errorf("unavailable explanation must withhold groups")
		}
	default:
		return fmt.Errorf("unsupported explanation field read state")
	}
	seen := map[string]bool{}
	for _, v := range r.Values {
		n, e := decimal(v.Count)
		if e != nil || n.Sign() == 0 {
			return fmt.Errorf("invalid explanation group count")
		}
		if v.Value != nil && (!utf8.ValidString(*v.Value) || len(*v.Value) > 4096) {
			return fmt.Errorf("invalid explanation group value")
		}
		if !f.CanBeMissing() && v.Value == nil {
			return fmt.Errorf("fixed explanation columns cannot be missing")
		}
		key := valueKey(v.Value)
		if seen[key] {
			return fmt.Errorf("duplicate explanation value")
		}
		seen[key] = true
	}
	return nil
}
func explanationGroupTotal(values []ExplanationValueCount) *big.Int {
	n := new(big.Int)
	for _, v := range values {
		x, _ := decimal(v.Count)
		n.Add(n, x)
	}
	return n
}
func explanationPercent(count, total string) string {
	if total == "0" {
		return ""
	}
	n, _ := decimal(count)
	d, _ := decimal(total)
	return new(big.Rat).Mul(new(big.Rat).SetFrac(n, d), big.NewRat(100, 1)).RatString()
}
func deriveExplanationField(field ExplanationFieldResult, a, b ExplanationPopulation) (ExplanationFieldResult, error) {
	if e := explanationReadValid(field.Baseline, field.Field); e != nil {
		return field, e
	}
	if e := explanationReadValid(field.Selection, field.Field); e != nil {
		return field, e
	}
	out := ExplanationFieldResult{Field: field.Field, Baseline: detached(field.Baseline), Selection: detached(field.Selection), State: "complete", Rows: []ExplanationRow{}}
	for _, side := range []*ExplanationFieldRead{&out.Baseline, &out.Selection} {
		sort.Slice(side.Values, func(i, j int) bool { return valueKey(side.Values[i].Value) < valueKey(side.Values[j].Value) })
	}
	if out.Baseline.State != "complete" || out.Selection.State != "complete" {
		out.State = "unavailable"
		out.Reason = out.Baseline.Reason
		if out.Baseline.State == "complete" {
			out.Reason = out.Selection.Reason
		}
		if out.Baseline.Reason == "field-bytes-limit" || out.Selection.Reason == "field-bytes-limit" {
			out.State = "unavailable"
			out.Reason = "field-bytes-limit"
		} else if out.Baseline.State == "partial" || out.Selection.State == "partial" {
			out.State = "partial"
			out.Reason = "cardinality-limit"
		}
	} else if explanationGroupTotal(out.Baseline.Values).String() != a.Count || explanationGroupTotal(out.Selection.Values).String() != b.Count {
		out.State = "changed_between_reads"
		out.Reason = "source-reads-disagree"
	} else {
		if a.Count == "0" || b.Count == "0" {
			out.State = "no_data"
			out.Reason = "empty-population"
		}
		av, bv := map[string]ExplanationValueCount{}, map[string]ExplanationValueCount{}
		keys := map[string]bool{}
		for _, v := range out.Baseline.Values {
			key := valueKey(v.Value)
			av[key] = v
			keys[key] = true
		}
		for _, v := range out.Selection.Values {
			key := valueKey(v.Value)
			bv[key] = v
			keys[key] = true
		}
		ordered := []string{}
		for k := range keys {
			ordered = append(ordered, k)
		}
		sort.Strings(ordered)
		for _, k := range ordered {
			x, xok := av[k]
			y, yok := bv[k]
			row := ExplanationRow{BaselineCount: "0", SelectionCount: "0"}
			if xok {
				row.Value = detached(x.Value)
				row.BaselineCount = x.Count
			}
			if yok {
				row.Value = detached(y.Value)
				row.SelectionCount = y.Count
			}
			row.BaselinePercent = explanationPercent(row.BaselineCount, a.Count)
			row.SelectionPercent = explanationPercent(row.SelectionCount, b.Count)
			if row.BaselinePercent != "" && row.SelectionPercent != "" {
				ap, _ := new(big.Rat).SetString(row.BaselinePercent)
				bp, _ := new(big.Rat).SetString(row.SelectionPercent)
				row.DifferencePP = new(big.Rat).Sub(bp, ap).RatString()
			}
			out.Rows = append(out.Rows, row)
		}
	}
	raw, _ := json.Marshal(out)
	if len(raw) > ExplanationFieldMaxBytes {
		out.State = "unavailable"
		out.Reason = "field-bytes-limit"
		out.Rows = []ExplanationRow{}
		for _, side := range []*ExplanationFieldRead{&out.Baseline, &out.Selection} {
			if side.State == "complete" {
				side.State = "unavailable"
				side.Reason = "field-bytes-limit"
			}
			side.Values = []ExplanationValueCount{}
		}
	}
	return out, nil
}

// BuildExplanation ranks complete observed group/count reads, not inspection rows.
// Read bases are source metadata; the result hash protects retained representation,
// not the authenticity or completeness of the upstream telemetry system.
func BuildExplanation(request ExplanationRequest, a, b ExplanationPopulation, fields []ExplanationFieldResult, readAt string) (*ExplanationResult, error) {
	if request.Validate() != nil || !explanationPopulationValid(a) || !explanationPopulationValid(b) || !explanationTime(readAt) || len(fields) != len(request.Fields) {
		return nil, fmt.Errorf("invalid explanation observations or request")
	}
	r := &ExplanationResult{Version: ExplanationVersion, Request: detached(request), State: "complete", ReadAt: readAt, Baseline: a, Selection: b, Fields: []ExplanationFieldResult{}, Ranked: []ExplanationLead{}, Warnings: []string{"Descriptive population-share differences, not causal attribution, health, statistical significance or proof of complete telemetry.", "Baseline, selection counts and each field are independent source reads with their own clocks and bases. Missing keys and present empty strings are distinct.", "Fields are explicitly chosen; no semantic convention, correlated identity, automatic remainder or original OTLP scalar type is inferred."}}
	counts := map[string]int{}
	for i, field := range fields {
		if field.Field != request.Fields[i] {
			return nil, fmt.Errorf("explanation field order differs from request")
		}
		out, e := deriveExplanationField(field, a, b)
		if e != nil {
			return nil, e
		}
		r.Fields = append(r.Fields, out)
		counts[out.State]++
		if out.State == "complete" {
			for j, row := range out.Rows {
				if row.DifferencePP != "0" {
					r.Ranked = append(r.Ranked, ExplanationLead{FieldIndex: i, RowIndex: j})
				}
			}
		}
	}
	if counts["complete"] != len(fields) {
		switch {
		case counts["no_data"] == len(fields):
			r.State = "no_data"
		case counts["unavailable"] == len(fields):
			r.State = "unavailable"
		case counts["changed_between_reads"] == len(fields):
			r.State = "changed_between_reads"
		default:
			r.State = "partial"
		}
	}
	sort.Slice(r.Ranked, func(i, j int) bool {
		x, y := r.Ranked[i], r.Ranked[j]
		xr, yr := r.Fields[x.FieldIndex].Rows[x.RowIndex], r.Fields[y.FieldIndex].Rows[y.RowIndex]
		a, _ := new(big.Rat).SetString(xr.DifferencePP)
		b, _ := new(big.Rat).SetString(yr.DifferencePP)
		a.Abs(a)
		b.Abs(b)
		if n := a.Cmp(b); n != 0 {
			return n > 0
		}
		xf, yf := r.Fields[x.FieldIndex].Field, r.Fields[y.FieldIndex].Field
		if xf.Kind != yf.Kind {
			return xf.Kind < yf.Kind
		}
		if xf.Key != yf.Key {
			return xf.Key < yf.Key
		}
		return valueKey(xr.Value) < valueKey(yr.Value)
	})
	r.Basis = explanationDigest(r)
	raw, _ := json.Marshal(r)
	if len(raw) > ResultSnapshotMaxBytes {
		return nil, fmt.Errorf("explanation exceeds result byte bound")
	}
	return r, nil
}
func ValidateExplanationResult(r *ExplanationResult) error {
	if r == nil {
		return fmt.Errorf("explanation required")
	}
	expected, e := BuildExplanation(r.Request, r.Baseline, r.Selection, r.Fields, r.ReadAt)
	if e != nil {
		return e
	}
	copy := *r
	copy.ResultToken = ""
	if !reflect.DeepEqual(copy, *expected) {
		return fmt.Errorf("explanation derived values or basis differ from observed reads")
	}
	return nil
}

func (s Source) Explain(ctx context.Context, request ExplanationRequest) (*ExplanationResult, error) {
	plan, e := PlanExplanation(request)
	if e != nil {
		return nil, e
	}
	read := func(sql string, params map[string]string) ([]json.RawMessage, error) {
		child, cancel := context.WithTimeout(ctx, 3*time.Second)
		defer cancel()
		return s.read(child, sql, params)
	}
	population := func(q Query) (ExplanationPopulation, error) {
		p, e := Compile(q)
		if e != nil {
			return ExplanationPopulation{}, e
		}
		rows, e := read(p.Count, p.Parameters)
		if e != nil {
			return ExplanationPopulation{}, e
		}
		var v struct {
			Count string `json:"count"`
		}
		if len(rows) != 1 || json.Unmarshal(rows[0], &v) != nil {
			return ExplanationPopulation{}, fmt.Errorf("invalid explanation population count")
		}
		if _, e := decimal(v.Count); e != nil {
			return ExplanationPopulation{}, e
		}
		return ExplanationPopulation{Count: v.Count, ReadAt: time.Now().UTC().Format(time.RFC3339Nano), Basis: explanationDigest(struct {
			Question Query
			Count    string
		}{q, v.Count})}, nil
	}
	a, e := population(plan.Baseline)
	if e != nil {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		return nil, fmt.Errorf("explanation baseline count unavailable")
	}
	b, e := population(plan.Selection)
	if e != nil {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		return nil, fmt.Errorf("explanation selection count unavailable")
	}
	fieldRead := func(q Query, f Field) ExplanationFieldRead {
		q.GroupBy = []Field{f}
		p, _ := Compile(q)
		rows, e := read(p.Groups, p.Parameters)
		out := ExplanationFieldRead{State: "complete", ReadAt: time.Now().UTC().Format(time.RFC3339Nano), Values: []ExplanationValueCount{}}
		if e != nil {
			out.State = "unavailable"
			out.Reason = "source-read-failed"
			if errors.Is(e, context.DeadlineExceeded) {
				out.Reason = "source-timeout"
			}
			return out
		}
		out.Basis = explanationDigest(struct {
			Question Query
			Rows     []json.RawMessage
		}{q, rows})
		if len(rows) > 1000 {
			out.State = "partial"
			out.Reason = "cardinality-limit"
			return out
		}
		for _, row := range rows {
			var v struct {
				Value *string `json:"g0"`
				Count string  `json:"count"`
			}
			var keys map[string]json.RawMessage
			if json.Unmarshal(row, &v) != nil || json.Unmarshal(row, &keys) != nil || keys["g0"] == nil || keys["count"] == nil {
				out.State = "unavailable"
				out.Reason = "invalid-source-groups"
				out.Values = []ExplanationValueCount{}
				return out
			}
			out.Values = append(out.Values, ExplanationValueCount{Value: v.Value, Count: v.Count})
		}
		if explanationReadValid(out, f) != nil {
			out.State = "unavailable"
			out.Reason = "invalid-source-groups"
			out.Values = []ExplanationValueCount{}
		}
		return out
	}
	fields := []ExplanationFieldResult{}
	for _, f := range plan.Request.Fields {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		x := fieldRead(plan.Baseline, f)
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		y := fieldRead(plan.Selection, f)
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		fields = append(fields, ExplanationFieldResult{Field: f, Baseline: x, Selection: y})
	}
	return BuildExplanation(plan.Request, a, b, fields, time.Now().UTC().Format(time.RFC3339Nano))
}

func PlanExplanationFilter(request ExplanationRequest, field Field, value *string, side, mode string) (Query, error) {
	if e := request.Validate(); e != nil {
		return Query{}, e
	}
	found := false
	for _, f := range request.Fields {
		found = found || f == field
	}
	if !found {
		return Query{}, fmt.Errorf("filter field was not requested")
	}
	if value != nil && (!utf8.ValidString(*value) || len(*value) > 4096) {
		return Query{}, fmt.Errorf("invalid explanation filter value")
	}
	var q Query
	switch side {
	case "baseline":
		q = detached(request.Baseline)
	case "selection":
		q = detached(request.Selection)
	default:
		return Query{}, fmt.Errorf("filter side must be baseline or selection")
	}
	var predicate Predicate
	switch mode {
	case "include":
		if value == nil {
			predicate = Predicate{Field: &field, Op: "missing"}
		} else {
			predicate = Predicate{Field: &field, Op: "eq", Value: *value}
		}
	case "exclude":
		if value == nil {
			predicate = Predicate{Field: &field, Op: "exists"}
		} else {
			predicate = Predicate{Any: []Predicate{{Field: &field, Op: "missing"}, {Field: &field, Op: "ne", Value: *value}}}
		}
	default:
		return Query{}, fmt.Errorf("filter mode must be include or exclude")
	}
	if q.Where == nil {
		q.Where = &predicate
	} else {
		q.Where = &Predicate{All: []Predicate{*q.Where, predicate}}
	}
	q.Offset = 0
	if e := q.Validate(); e != nil {
		return Query{}, e
	}
	return q, nil
}

// SaveExplanation publishes the already observed aggregates without a source read.
// HTTP callers promote only a server-held snapshot, never client-supplied counts.
func (s CaptureStore) SaveExplanation(result *ExplanationResult, title string) (*SavedCapture, error) {
	return s.saveExplanationAt(result, title, time.Now().UTC().Format(time.RFC3339Nano))
}

func (s CaptureStore) saveExplanationAt(result *ExplanationResult, title, custodyAt string) (*SavedCapture, error) {
	if e := ValidateCaptureTitle(title); e != nil {
		return nil, e
	}
	if e := ValidateExplanationResult(result); e != nil {
		return nil, e
	}
	copy := detached(result)
	copy.ResultToken = ""
	return s.publish(SavedCapture{Explanation: copy, CaptureReceipt: CaptureReceipt{Version: 1, Title: title, CapturedAt: custodyAt, Question: CaptureQuestion{Kind: "explanation", Explanation: detached(&copy.Request)}}})
}
