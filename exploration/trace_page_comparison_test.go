package exploration

import (
	"encoding/json"
	"os"
	"reflect"
	"strings"
	"testing"
)

func comparisonRow(id, parent, start, duration, operation string) json.RawMessage {
	return tracePageEdit(tracePageRow(id, parent, start, duration), map[string]any{"SpanName": operation, "SpanKind": "Server"})
}

type traceComparisonFixture struct {
	Name               string               `json:"name"`
	Baseline           []json.RawMessage    `json:"baseline"`
	Selection          []json.RawMessage    `json:"selection"`
	BaselineTruncated  bool                 `json:"baselineTruncated"`
	SelectionTruncated bool                 `json:"selectionTruncated"`
	Expected           *TracePageComparison `json:"expected"`
}

func traceComparisonCases() []traceComparisonFixture {
	a := []json.RawMessage{comparisonRow(tRoot, "", "100", "100", "root"), comparisonRow(tChild, tRoot, "110", "40", "child")}
	b := []json.RawMessage{comparisonRow(tRoot, "", "1000", "140", "root"), comparisonRow(tChild, tRoot, "1020", "60", "child")}
	out := []traceComparisonFixture{{Name: "unique-different-relative-origins", Baseline: a, Selection: b}, {Name: "repeated-labels-no-zip", Baseline: []json.RawMessage{a[0], tracePageEdit(a[1], map[string]any{"SpanName": "root"})}, Selection: []json.RawMessage{b[0], tracePageEdit(b[1], map[string]any{"SpanName": "root"})}}, {Name: "absent-not-zero", Baseline: a, Selection: []json.RawMessage{b[0]}}, {Name: "unknown-residual-known-inclusive", Baseline: a, Selection: []json.RawMessage{b[0], comparisonRow(tChild, tRoot, "900", "60", "child")}}, {Name: "duplicate-identity", Baseline: []json.RawMessage{a[0], tracePageEdit(a[0], map[string]any{"SpanName": "other"})}, Selection: b}, {Name: "mixed-trace-page", Baseline: []json.RawMessage{a[0], tracePageEdit(a[1], map[string]any{"TraceId": "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"})}, Selection: b}, {Name: "empty-pages", Baseline: []json.RawMessage{}, Selection: []json.RawMessage{}}, {Name: "nullable-labels", Baseline: []json.RawMessage{tracePageEdit(a[0], map[string]any{"ServiceName": nil, "SpanName": nil})}, Selection: []json.RawMessage{tracePageEdit(b[0], map[string]any{"ServiceName": nil, "SpanName": nil})}}, {Name: "bad-label-types", Baseline: []json.RawMessage{tracePageEdit(a[0], map[string]any{"ServiceName": 4})}, Selection: []json.RawMessage{tracePageEdit(b[0], map[string]any{"ServiceName": 4})}}, {Name: "partial-page-only", Baseline: a, Selection: b, BaselineTruncated: true, SelectionTruncated: true}, {Name: "invalid-time-known-duration", Baseline: []json.RawMessage{tracePageEdit(a[0], map[string]any{"timestamp_ns": "01"})}, Selection: []json.RawMessage{b[0]}}, {Name: "above-safe-integer", Baseline: []json.RawMessage{comparisonRow(tRoot, "", "9007199254740993", "9007199254740993", "root")}, Selection: []json.RawMessage{comparisonRow(tRoot, "", "9007199254740995", "9007199254740995", "root")}}}
	for _, c := range []struct {
		name string
		a, b any
	}{{"absent", nil, nil}, {"empty", "", ""}, {"wrong-type", 3, 3}, {"different", "Client", "Server"}, {"unknown-same", "custom-kind", "custom-kind"}} {
		out = append(out, traceComparisonFixture{Name: "span-kind-" + c.name, Baseline: []json.RawMessage{tracePageEdit(a[0], map[string]any{"SpanKind": c.a})}, Selection: []json.RawMessage{tracePageEdit(b[0], map[string]any{"SpanKind": c.b})}})
	}
	out = append(out, traceComparisonFixture{Name: "surrogate-kind-json-normalization", Baseline: []json.RawMessage{json.RawMessage(strings.ReplaceAll(string(a[0]), `"SpanKind":"Server"`, `"SpanKind":"\ud800"`))}, Selection: []json.RawMessage{tracePageEdit(b[0], map[string]any{"SpanKind": "\ufffd"})}})
	out = append(out, traceComparisonFixture{Name: "reported-empty-labels", Baseline: []json.RawMessage{tracePageEdit(a[0], map[string]any{"ServiceName": "", "SpanName": ""})}, Selection: []json.RawMessage{tracePageEdit(b[0], map[string]any{"ServiceName": "", "SpanName": ""})}})
	return out
}
func TestTracePageComparisonKnownDifferencesAndNoZip(t *testing.T) {
	cases := traceComparisonCases()
	r, e := CompareTracePages(cases[0].Baseline, cases[0].Selection, false, false)
	if e != nil {
		t.Fatal(e)
	}
	for _, g := range r.Groups {
		if g.MatchState != "unique-label-match" || g.Match == nil {
			t.Fatal(g)
		}
		if *g.Operation == "child" && (*g.Match.RelativeStartDifferenceNS != "10" || *g.InclusiveDifferenceNS != "20") {
			t.Fatal("exact row deltas", g)
		}
	}
	r, _ = CompareTracePages(cases[1].Baseline, cases[1].Selection, false, false)
	if r.Groups[0].MatchState != "ambiguous" || r.Groups[0].Match != nil || r.Groups[0].InclusiveDifferenceNS == nil {
		t.Fatal("repeat zipped or aggregate lost", r.Groups)
	}
	r, _ = CompareTracePages(cases[2].Baseline, cases[2].Selection, false, false)
	for _, g := range r.Groups {
		if g.Selection == nil && (g.InclusiveDifferenceNS != nil || g.OutsideDifferenceNS != nil) {
			t.Fatal("absent synthesized zero", g)
		}
	}
	r, _ = CompareTracePages(cases[3].Baseline, cases[3].Selection, false, false)
	for _, g := range r.Groups {
		if *g.Operation == "root" && (g.OutsideDifferenceNS != nil || g.InclusiveDifferenceNS == nil) {
			t.Fatal("partial metric conflation", g)
		}
	}
	for _, i := range []int{4, 5, 8, 12, 13, 14, 15} {
		r, _ = CompareTracePages(cases[i].Baseline, cases[i].Selection, false, false)
		for _, g := range r.Groups {
			if g.Match != nil {
				t.Fatal("invalid eligibility matched", cases[i].Name, g)
			}
		}
	}
	r, _ = CompareTracePages(cases[len(cases)-2].Baseline, cases[len(cases)-2].Selection, false, false)
	if r.Groups[0].Match == nil {
		t.Fatal("exact unknown kind coerced/refused")
	}
	if _, e := CompareTracePages(make([]json.RawMessage, 1001), nil, false, false); e == nil {
		t.Fatal("unbounded comparison accepted")
	}
}
func TestTracePageComparisonSharedFixtures(t *testing.T) {
	fixtures := traceComparisonCases()
	for i := range fixtures {
		f := &fixtures[i]
		var e error
		f.Expected, e = CompareTracePages(f.Baseline, f.Selection, f.BaselineTruncated, f.SelectionTruncated)
		if e != nil {
			t.Fatal(e)
		}
	}
	raw, e := json.MarshalIndent(fixtures, "", "  ")
	if e != nil {
		t.Fatal(e)
	}
	path := "../tests/fixtures/trace-page-comparison.json"
	if os.Getenv("GOO_UPDATE_TRACE_COMPARISON_FIXTURES") == "1" {
		if e := os.WriteFile(path, append(raw, '\n'), 0644); e != nil {
			t.Fatal(e)
		}
	}
	stored, e := os.ReadFile(path)
	if e != nil {
		t.Fatal(e)
	}
	var x, y any
	json.Unmarshal(raw, &x)
	json.Unmarshal(stored, &y)
	if !reflect.DeepEqual(x, y) {
		t.Fatal("normative comparison fixture mismatch")
	}
}
