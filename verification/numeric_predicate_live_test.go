package verification

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/observability/quintgo/qobs"
	"github.com/casselc/observability/quintgo/validate"
)

type typedFilterRecord struct {
	Raw         *string `json:"raw"`
	ResourceRaw *string `json:"resourceRaw"`
	TimeMS      int64   `json:"timeMs"`
}
type typedFilterObservation struct {
	Action  string                   `json:"action"`
	Reads   int                      `json:"reads"`
	Query   *exploration.Query       `json:"query,omitempty"`
	Result  *exploration.Result      `json:"result,omitempty"`
	Capture *exploration.CapturePage `json:"capture,omitempty"`
}
type typedFilterLedger struct {
	Version         int                        `json:"version"`
	Binding         exploration.NumericBinding `json:"binding"`
	Threshold       string                     `json:"threshold"`
	Records         []typedFilterRecord        `json:"records"`
	Original        exploration.Query          `json:"original"`
	MainResult      exploration.Result         `json:"mainResult"`
	Saved, Reopened exploration.CapturePage
	Reused          exploration.Result       `json:"reused"`
	Steps           []typedFilterObservation `json:"steps"`
}

// This fixture oracle uses the exact records emitted by journey 171 and the
// independent stdlib/roundtrip oracle from numeric_predicate_test.go. No SQL or
// production Match implementation predicts the result it is checking.
func typedFilterReference(p *exploration.Predicate, row typedFilterRecord, key string) (bool, error) {
	if p == nil {
		return true, nil
	}
	if len(p.All) > 0 {
		for i := range p.All {
			ok, e := typedFilterReference(&p.All[i], row, key)
			if e != nil || !ok {
				return false, e
			}
		}
		return true, nil
	}
	if len(p.Any) > 0 {
		for i := range p.Any {
			ok, e := typedFilterReference(&p.Any[i], row, key)
			if e != nil {
				return false, e
			}
			if ok {
				return true, nil
			}
		}
		return false, nil
	}
	var raw *string
	field := p.Field
	if p.Numeric != nil {
		field = &p.Numeric.Field
	}
	if field == nil {
		return false, fmt.Errorf("missing fixture field")
	}
	if field.Kind == "column" && field.Key == "timestamp_ns" {
		v := strconv.FormatInt(row.TimeMS*1000000, 10)
		raw = &v
	} else {
		if field.Key != key {
			return false, fmt.Errorf("fixture does not define field %s", field.Key)
		}
		switch field.Kind {
		case "attribute":
			raw = row.Raw
		case "resource":
			raw = row.ResourceRaw
		default:
			return false, fmt.Errorf("unsupported fixture field kind")
		}
	}
	if p.Numeric != nil {
		return predicateOracle(raw, *p.Numeric, p.Op, p.Value), nil
	}
	switch p.Op {
	case "exists":
		return raw != nil, nil
	case "missing":
		return raw == nil, nil
	}
	if raw == nil {
		return false, nil
	}
	switch p.Op {
	case "eq":
		return *raw == p.Value, nil
	case "ne":
		return *raw != p.Value, nil
	case "contains":
		return strings.Contains(*raw, p.Value), nil
	case "gte", "lte":
		a, e := strconv.ParseUint(*raw, 10, 64)
		b, e2 := strconv.ParseUint(p.Value, 10, 64)
		return e == nil && e2 == nil && (p.Op == "gte" && a >= b || p.Op == "lte" && a <= b), nil
	}
	return false, fmt.Errorf("unsupported fixture operator")
}
func typedFilterCount(q exploration.Query, l typedFilterLedger) (int, error) {
	n := 0
	for _, row := range l.Records {
		if row.TimeMS < q.FromMS || row.TimeMS >= q.ToMS {
			continue
		}
		ok, e := typedFilterReference(q.Where, row, l.Binding.Field.Key)
		if e != nil {
			return 0, e
		}
		if ok {
			n++
		}
	}
	return n, nil
}

// Match Go transport defaults only: omitempty treats nil and empty grouping alike.
// Numeric predicate bindings and decimal string thresholds are never rewritten.
func typedFilterSameQuery(a, b exploration.Query) bool {
	left, e1 := json.Marshal(a)
	right, e2 := json.Marshal(b)
	return e1 == nil && e2 == nil && string(left) == string(right)
}
func typedFilterValidResult(r *exploration.Result, q *exploration.Query, l typedFilterLedger) bool {
	if r == nil || q == nil || q.Validate() != nil || !typedFilterSameQuery(r.Question, *q) || q.Environment != l.Original.Environment || q.Signal != l.Original.Signal {
		return false
	}
	want, e := typedFilterCount(*q, l)
	if e != nil || r.Count != strconv.Itoa(want) || r.Basis == "" || r.CapturedAt == "" {
		return false
	}
	if (want > 0 && r.State != "observed") || (want == 0 && r.State != "no_data") {
		return false
	}
	pageSize := want - q.Offset
	if pageSize < 0 {
		pageSize = 0
	}
	if pageSize > q.Limit {
		pageSize = q.Limit
	}
	if len(r.Rows) != pageSize {
		return false
	}
	fixture := map[int64]typedFilterRecord{}
	for _, record := range l.Records {
		if _, exists := fixture[record.TimeMS]; exists {
			return false
		}
		fixture[record.TimeMS] = record
	}
	seen := map[int64]bool{}
	for _, rawRow := range r.Rows {
		var row struct {
			Timestamp  string            `json:"timestamp_ns"`
			Attributes map[string]string `json:"Attributes"`
			Resource   map[string]string `json:"ResourceAttributes"`
		}
		if json.Unmarshal(rawRow, &row) != nil {
			return false
		}
		ns, err := strconv.ParseInt(row.Timestamp, 10, 64)
		if err != nil || strconv.FormatInt(ns, 10) != row.Timestamp || ns%1000000 != 0 {
			return false
		}
		ms := ns / 1000000
		emitted, exists := fixture[ms]
		if !exists || seen[ms] || ms < q.FromMS || ms >= q.ToMS {
			return false
		}
		seen[ms] = true
		for _, pair := range []struct {
			values   map[string]string
			expected *string
		}{{row.Attributes, emitted.Raw}, {row.Resource, emitted.ResourceRaw}} {
			value, present := pair.values[l.Binding.Field.Key]
			if (pair.expected == nil && present) || (pair.expected != nil && (!present || value != *pair.expected)) {
				return false
			}
		}
		matches, err := typedFilterReference(q.Where, emitted, l.Binding.Field.Key)
		if err != nil || !matches {
			return false
		}
	}
	return true
}
func typedFilterNumeric(p *exploration.Predicate) []exploration.Predicate {
	if p == nil {
		return nil
	}
	if p.Numeric != nil {
		return []exploration.Predicate{*p}
	}
	out := []exploration.Predicate{}
	for i := range p.All {
		out = append(out, typedFilterNumeric(&p.All[i])...)
	}
	for i := range p.Any {
		out = append(out, typedFilterNumeric(&p.Any[i])...)
	}
	return out
}
func checkTypedFilterLedger(t *testing.T, l typedFilterLedger) bool {
	t.Helper()
	sink := &qobs.MemorySink{}
	rec := qobs.NewRecorder(sink)
	record := func(action string, ok bool) {
		if !ok {
			t.Logf("typed filter ledger rejected phase %s", action)
		}
		rec.Record(context.Background(), action, nil, "obs.coherent", ok)
	}
	original := typedFilterNumeric(l.Original.Where)
	planned := l.Version == 1 && len(l.Records) > 0 && len(l.Steps) > 0 && l.Original.Validate() == nil && len(original) == 1 && reflect.DeepEqual(*original[0].Numeric, l.Binding) && original[0].Value == l.Threshold && l.Threshold == "9007199254740993" && l.Binding.Type == "uint64" && l.Binding.Unit == "items" && l.Binding.Parse == "canonical-decimal" && l.Binding.Field.Kind == "attribute" && original[0].Op == "gte"
	record("plan", planned)
	reads := 0
	coherent := true
	actions := map[string]int{}
	operators := map[string]bool{}
	types := map[string]bool{}
	kinds := map[string]bool{}
	for _, step := range l.Steps {
		actions[step.Action]++
		delta := step.Reads - reads
		switch step.Action {
		case "edit", "invalid", "reopen":
			coherent = coherent && delta == 0
		case "run", "brush", "reuse", "legacy":
			if !typedFilterValidResult(step.Result, step.Query, l) {
				t.Logf("rejected result observation action=%s reads=%d", step.Action, step.Reads)
			}
			coherent = coherent && delta == 1 && typedFilterValidResult(step.Result, step.Query, l)
			if step.Query != nil {
				leaves := typedFilterNumeric(step.Query.Where)
				for _, leaf := range leaves {
					operators[leaf.Op] = true
					types[leaf.Numeric.Type] = true
					kinds[leaf.Numeric.Field.Kind] = true
				}
				if step.Action == "brush" || step.Action == "reuse" {
					coherent = coherent && reflect.DeepEqual(step.Query.Where, l.Original.Where)
				}
				if step.Action == "brush" {
					coherent = coherent && step.Query.FromMS >= l.Original.FromMS && step.Query.ToMS <= l.Original.ToMS && (step.Query.FromMS != l.Original.FromMS || step.Query.ToMS != l.Original.ToMS)
				}
				if step.Action == "legacy" {
					coherent = coherent && len(leaves) == 0 && step.Query.Where != nil
				}
			}
		case "capture":
			coherent = coherent && delta == 1 && step.Capture != nil
		default:
			coherent = false
		}
		if step.Reads < 0 {
			coherent = false
		}
		reads = step.Reads
		if step.Action == "invalid" && step.Query != nil {
			coherent = coherent && step.Query.Validate() != nil
		}
		if step.Action == "reopen" {
			coherent = coherent && step.Capture != nil && reflect.DeepEqual(*step.Capture, l.Saved)
		}
	}
	for _, action := range []string{"edit", "invalid", "run", "brush", "capture", "reuse", "reopen", "legacy"} {
		coherent = coherent && actions[action] > 0
	}
	for _, op := range []string{"gte", "ne", "valid", "missing", "empty", "invalid"} {
		coherent = coherent && operators[op]
	}
	coherent = coherent && types["int64"] && types["uint64"] && kinds["attribute"] && kinds["resource"]
	record("execute", coherent && typedFilterValidResult(&l.MainResult, &l.Original, l))
	record("inspect", actions["edit"] > 0 && actions["invalid"] > 0)
	saved := l.Saved
	var capturedQuestion *exploration.Query
	for _, step := range l.Steps {
		if step.Action == "brush" {
			capturedQuestion = step.Query
		}
	}
	captureOK := capturedQuestion != nil && saved.Receipt.Question.Explore != nil && reflect.DeepEqual(saved.Receipt.Question.Explore, capturedQuestion) && typedFilterValidResult(saved.Explore, capturedQuestion, l) && len(saved.Receipt.ID) == 64
	for _, step := range l.Steps {
		if step.Action == "capture" {
			captureOK = captureOK && step.Capture != nil && reflect.DeepEqual(*step.Capture, saved)
		}
	}
	record("capture", captureOK)
	record("reuse", reflect.DeepEqual(l.Reused.Question.Where, l.Original.Where) && typedFilterValidResult(&l.Reused, &l.Reused.Question, l))
	record("reopen", reflect.DeepEqual(l.Saved, l.Reopened))
	checker, e := validate.NewChecker("model/grouped_series.yaml", validate.Options{Quint: quintPath(t), Backend: "typescript", Dir: t.TempDir()})
	if e != nil {
		t.Fatal(e)
	}
	result, e := checker.Check(context.Background(), sink.Steps())
	if e != nil {
		t.Fatal(e)
	}
	return result.OK()
}
func TestQuintTypedNumericFilterLiveLedger(t *testing.T) {
	name := os.Getenv("GOO_TYPED_NUMERIC_LEDGER")
	if name == "" {
		t.Skip("actual journey171 ledger opt-in")
	}
	raw, e := os.ReadFile(name)
	if e != nil {
		t.Fatal(e)
	}
	read := func() typedFilterLedger {
		var l typedFilterLedger
		if e := json.Unmarshal(raw, &l); e != nil {
			t.Fatal(e)
		}
		return l
	}
	if !checkTypedFilterLedger(t, read()) {
		t.Fatal("actual typed numeric filter ledger rejected")
	}
	mutations := []func(*typedFilterLedger){
		func(l *typedFilterLedger) {
			for i := range l.Steps {
				if l.Steps[i].Action == "run" && l.Steps[i].Query != nil && l.Steps[i].Result != nil {
					l.Steps[i].Query.Environment = "other-environment"
					l.Steps[i].Result.Question.Environment = "other-environment"
					return
				}
			}
		},
		func(l *typedFilterLedger) {
			if len(l.MainResult.Rows) > 0 {
				l.MainResult.Rows = l.MainResult.Rows[1:]
			}
		},
		func(l *typedFilterLedger) {
			if len(l.MainResult.Rows) > 0 {
				var row map[string]json.RawMessage
				if json.Unmarshal(l.MainResult.Rows[0], &row) == nil {
					var attrs map[string]string
					if json.Unmarshal(row["Attributes"], &attrs) == nil {
						attrs[l.Binding.Field.Key] = "01"
						row["Attributes"], _ = json.Marshal(attrs)
						l.MainResult.Rows[0], _ = json.Marshal(row)
					}
				}
			}
		},
		func(l *typedFilterLedger) { l.MainResult.State = "no_data" },
		func(l *typedFilterLedger) { l.MainResult.Count = "999999" },
		func(l *typedFilterLedger) { l.Original.Where.Value = "9007199254740992" },
		func(l *typedFilterLedger) { l.Binding.Field.Kind = "resource" },
		func(l *typedFilterLedger) {
			for i := range l.Steps {
				if l.Steps[i].Action == "edit" {
					l.Steps[i].Reads++
					break
				}
			}
		},
		func(l *typedFilterLedger) {
			for i := range l.Steps {
				if l.Steps[i].Action == "invalid" {
					l.Steps[i].Reads++
					break
				}
			}
		},
		func(l *typedFilterLedger) { l.Reopened.Receipt.ID = strings.Repeat("f", 64) },
		func(l *typedFilterLedger) { l.Reused.Question.Where = nil },
		func(l *typedFilterLedger) {
			for i := range l.Steps {
				if l.Steps[i].Action == "brush" && l.Steps[i].Query != nil {
					l.Steps[i].Query.Where = nil
					break
				}
			}
		},
		func(l *typedFilterLedger) {
			for i := range l.Steps {
				if l.Steps[i].Action == "run" && l.Steps[i].Result != nil {
					l.Steps[i].Result.Count = "999999"
					break
				}
			}
		},
		func(l *typedFilterLedger) { l.Threshold = "9007199254740992" },
		func(l *typedFilterLedger) {
			for i := range l.Steps {
				if l.Steps[i].Action == "run" && l.Steps[i].Query != nil {
					for _, leaf := range typedFilterNumeric(l.Steps[i].Query.Where) {
						if leaf.Op == "invalid" {
							l.Steps[i].Result.Count = "0"
							return
						}
					}
				}
			}
		},
		func(l *typedFilterLedger) {
			for i := range l.Steps {
				if l.Steps[i].Action == "reopen" {
					l.Steps[i].Reads++
					return
				}
			}
		},
	}
	for i, mutate := range mutations {
		l := read()
		mutate(&l)
		if checkTypedFilterLedger(t, l) {
			t.Fatalf("typed numeric live mutation accepted %d", i)
		}
	}
}

func TestTypedNumericLedgerQueryDefaults(t *testing.T) {
	binding := predicateBinding("uint64", "attribute")
	q := predicateQuery(predicateLeaf(binding, "gte", "9007199254740993"))
	explicit := q
	explicit.GroupBy = []exploration.Field{}
	if !typedFilterSameQuery(q, explicit) {
		t.Fatal("Go omitted grouping transport default treated as changed intent")
	}
	changed := predicateQuery(predicateLeaf(binding, "gte", "9007199254740992"))
	if typedFilterSameQuery(q, changed) {
		t.Fatal("large numeric threshold rounded by transport comparison")
	}
	changed = q
	changed.GroupBy = []exploration.Field{{Kind: "column", Key: "service"}}
	if typedFilterSameQuery(q, changed) {
		t.Fatal("real grouping change erased by transport comparison")
	}
}
