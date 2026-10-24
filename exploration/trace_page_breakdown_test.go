package exploration

import (
	"encoding/json"
	"os"
	"reflect"
	"strings"
	"testing"
)

func tracePageRow(id, parent, start, duration string) json.RawMessage {
	v := map[string]any{"TraceId": strings.Repeat("a", 32), "SpanId": id, "ParentSpanId": parent, "ServiceName": "svc", "SpanName": "op", "timestamp_ns": start, "duration_ns": duration}
	raw, _ := json.Marshal(v)
	return raw
}
func tracePageEdit(raw json.RawMessage, changes map[string]any) json.RawMessage {
	v := map[string]any{}
	json.Unmarshal(raw, &v)
	for k, x := range changes {
		if x == nil {
			delete(v, k)
		} else {
			v[k] = x
		}
	}
	out, _ := json.Marshal(v)
	return out
}

var tRoot = "0000000000000001"
var tChild = "0000000000000002"
var tOther = "0000000000000003"

func tracePageCases() []struct {
	Name      string
	Rows      []json.RawMessage
	Truncated bool
} {
	base := []json.RawMessage{tracePageRow(tRoot, "", "0", "100"), tracePageRow(tChild, tRoot, "10", "50"), tracePageRow(tOther, tRoot, "40", "40"), tracePageRow("0000000000000004", tChild, "20", "10"), tracePageRow("0000000000000005", tRoot, "50", "0")}
	return []struct {
		Name      string
		Rows      []json.RawMessage
		Truncated bool
	}{
		{"overlap-nested-zero", base, false},
		{"child-own-residual-unknown", []json.RawMessage{base[0], base[1], tracePageRow(tOther, tChild, "5", "70")}, false},
		{"outside-parent", []json.RawMessage{base[0], tracePageRow(tChild, tRoot, "90", "20")}, true},
		{"duplicates", []json.RawMessage{base[0], base[1], tracePageRow(tChild, tRoot, "20", "20")}, false},
		{"cycle", []json.RawMessage{tracePageRow(tRoot, tChild, "0", "100"), tracePageRow(tChild, tRoot, "10", "20")}, false},
		{"missing-parent", []json.RawMessage{tracePageRow(tRoot, tOther, "9007199254740993", "10")}, true},
		{"invalid-intervals", []json.RawMessage{base[0], tracePageRow(tChild, tRoot, "01", "3"), tracePageRow(tOther, tRoot, "18446744073709551615", "1"), tracePageRow("bad", tRoot, "4", "2")}, false},
		{"nullable-labels", []json.RawMessage{tracePageEdit(base[0], map[string]any{"ServiceName": nil, "SpanName": nil}), tracePageEdit(base[1], map[string]any{"ServiceName": "", "SpanName": ""}), tracePageEdit(base[2], map[string]any{"ServiceName": "(missing)", "SpanName": "(missing)"})}, false},
		{"malformed-row-salvage", []json.RawMessage{base[0], tracePageEdit(base[1], map[string]any{"duration_ns": 2}), json.RawMessage("null"), tracePageEdit(base[2], map[string]any{"ServiceName": 2})}, false},
		{"wide-sums", []json.RawMessage{tracePageRow(tRoot, "", "0", "18446744073709551615"), tracePageRow(tChild, "", "0", "18446744073709551615")}, false},
		{"empty", []json.RawMessage{}, false},
	}
}
func TestTracePageDirectChildUnionAndLocalUnknown(t *testing.T) {
	cases := tracePageCases()
	r, e := AnalyzeTracePage(cases[0].Rows, false)
	if e != nil {
		t.Fatal(e)
	}
	if *r.Rows[0].ChildUnionNS != "70" || *r.Rows[0].OutsideChildrenNS != "30" || *r.Rows[1].OutsideChildrenNS != "40" {
		t.Fatal("union or direct child semantics", r.Rows)
	}
	r, _ = AnalyzeTracePage(cases[1].Rows, false)
	if r.Rows[1].OutsideChildrenNS != nil || r.Rows[0].OutsideChildrenNS == nil || *r.Rows[0].OutsideChildrenNS != "50" {
		t.Fatal("child residual uncertainty propagated beyond interval", r.Rows)
	}
	r, _ = AnalyzeTracePage(cases[2].Rows, true)
	if r.Rows[0].OutsideChildrenNS != nil || r.Rows[0].ChildUnionNS != nil || r.Rows[0].InclusiveNS == nil || !r.Truncated {
		t.Fatal("outside child clipped or inclusive lost", r)
	}
	r, _ = AnalyzeTracePage(cases[5].Rows, true)
	if r.Rows[0].OutsideChildrenNS == nil || *r.Rows[0].OutsideChildrenNS != "10" || r.State != "observed" {
		t.Fatal("missing ancestry blocks local evidence", r)
	}
}
func TestTracePageAmbiguityLabelsAndExactSums(t *testing.T) {
	cases := tracePageCases()
	for _, i := range []int{3, 4, 6, 8} {
		r, e := AnalyzeTracePage(cases[i].Rows, false)
		if e != nil || r.State != "partial" || r.Rows[0].OutsideChildrenNS != nil {
			t.Fatal(cases[i].Name, r, e)
		}
	}
	r, _ := AnalyzeTracePage(cases[7].Rows, false)
	if len(r.Groups) != 3 || r.Groups[0].Service != nil || r.Groups[1].Service == nil || *r.Groups[1].Service != "" {
		t.Fatal("nullable label groups collapsed", r.Groups)
	}
	r, _ = AnalyzeTracePage(cases[9].Rows, false)
	if *r.Groups[0].InclusiveNS != "36893488147419103230" {
		t.Fatal("group sum overflow", r.Groups)
	}
	r, _ = AnalyzeTracePage([]json.RawMessage{json.RawMessage("null")}, false)
	if r.Groups[0].InclusiveNS != nil || r.Groups[0].OutsideChildrenNS != nil || r.Groups[0].InclusiveUnknown != 1 || r.Groups[0].OutsideUnknown != 1 {
		t.Fatal("unknown fabricated zero", r)
	}
	rows := make([]json.RawMessage, 1001)
	if _, e := AnalyzeTracePage(rows, false); e == nil {
		t.Fatal("unbounded page accepted")
	}
	original := append([]json.RawMessage(nil), cases[0].Rows...)
	AnalyzeTracePage(cases[0].Rows, false)
	if !reflect.DeepEqual(original, cases[0].Rows) {
		t.Fatal("input mutation")
	}
}
func TestTracePageSharedNormativeFixtures(t *testing.T) {
	type fixture struct {
		Name      string              `json:"name"`
		Rows      []json.RawMessage   `json:"rows"`
		Truncated bool                `json:"truncated"`
		Expected  *TracePageBreakdown `json:"expected"`
	}
	fixtures := []fixture{}
	for _, tc := range tracePageCases() {
		out, e := AnalyzeTracePage(tc.Rows, tc.Truncated)
		if e != nil {
			t.Fatal(e)
		}
		fixtures = append(fixtures, fixture{tc.Name, tc.Rows, tc.Truncated, out})
	}
	raw, e := json.MarshalIndent(fixtures, "", "  ")
	if e != nil {
		t.Fatal(e)
	}
	path := "../tests/fixtures/trace-page-breakdown.json"
	if os.Getenv("GOO_UPDATE_TRACE_PAGE_FIXTURES") == "1" {
		if e := os.MkdirAll("../tests/fixtures", 0755); e != nil {
			t.Fatal(e)
		}
		if e := os.WriteFile(path, append(raw, '\n'), 0644); e != nil {
			t.Fatal(e)
		}
	}
	stored, e := os.ReadFile(path)
	if e != nil {
		t.Fatal(e)
	}
	var actual, want any
	json.Unmarshal(stored, &actual)
	json.Unmarshal(raw, &want)
	if !reflect.DeepEqual(actual, want) {
		t.Fatal("Go normative fixtures differ; review semantic delta before regeneration")
	}
}
