package exploration

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"reflect"
	"strings"
	"testing"
)

func statusViewFixture(t *testing.T) *Result {
	t.Helper()
	r := viewFixture()
	r.Question.SpanStatus = true
	r.Question.BucketMS = 1000
	r.Question.ToMS = r.Question.FromMS + 3000
	r.Count = "10"
	r.SpanStatus = &SpanStatusSummary{Count: "10", Error: "2", OK: "3", Unset: "4", Unknown: "1"}
	r.Groups = nil
	for _, item := range []struct {
		category *string
		status   SpanStatusSummary
	}{{nil, SpanStatusSummary{"4", "1", "0", "2", "1"}}, {str(""), SpanStatusSummary{"6", "1", "3", "2", "0"}}} {
		b, err := json.Marshal(map[string]any{"g0": item.category, "count": item.status.Count, "spanStatus": item.status})
		if err != nil {
			t.Fatal(err)
		}
		r.Groups = append(r.Groups, b)
	}
	return r
}

func TestViewStatusUsesExactClassCountsAndSubsetFraction(t *testing.T) {
	r := statusViewFixture(t)
	groups, summary := groupViewDataset(r, true), summaryViewDataset(r, true)
	if groups == nil || summary == nil || groups.Completeness == nil || summary.Completeness == nil || groups.Partition.Total != "10" {
		t.Fatal("validated status did not produce canonical snapshots", groups, summary)
	}
	if *groups.Rows[0].Values["status.error_fraction"] != "1/4" || *groups.Rows[1].Values["status.error_fraction"] != "1/6" || *summary.Rows[0].Values["status.error_fraction"] != "1/5" {
		t.Fatal("status fractions were scaled or averaged instead of derived from each row", groups.Rows, summary.Rows)
	}
	if *groups.Rows[0].Values["status.ok"] != "0" || *summary.Rows[0].Values["status.unset"] != "4" || *summary.Rows[0].Values["status.unknown"] != "1" {
		t.Fatal("zero, Unset or unknown classes changed", groups.Rows, summary.Rows)
	}
	for _, d := range []*ViewDataset{groups, summary} {
		for _, row := range d.Rows {
			for _, key := range []string{"status.error", "status.ok", "status.unset", "status.unknown", "status.error_fraction"} {
				if row.Cells[key].State != "derived" || row.Cells[key].Support != nil {
					t.Fatal("status fabricated numeric-validity support", key, row.Cells[key])
				}
			}
		}
		for _, field := range d.Fields {
			if field.Key == "status.error_fraction" && (field.Type != "rational" || field.Unit != "1" || !field.Nullable || !reflect.DeepEqual(field.Aggregation, &ViewAggregation{Family: "ratio", Kind: "fraction", Method: "exact-rational", Numerator: "status.error", Denominator: "count"})) {
				t.Fatal("fraction metadata does not name its operands", field)
			}
		}
	}
}

func TestViewStatusZeroAndUnknownAreDifferent(t *testing.T) {
	r := statusViewFixture(t)
	r.State, r.Count, r.Groups = "no_data", "0", nil
	zero := emptySpanStatus()
	r.SpanStatus = &zero
	d := summaryViewDataset(r, true)
	if d == nil || *d.Rows[0].Values["status.error"] != "0" || d.Rows[0].Cells["status.error"].State != "derived" || d.Rows[0].Values["status.error_fraction"] != nil || d.Rows[0].Cells["status.error_fraction"].State != "no_data" {
		t.Fatal("empty denominator became an observed healthy zero fraction", d)
	}
	start := r.Question.FromMS
	r.TimeSeries = []TimePoint{
		{FromMS: start, ToMS: start + 1000, State: "unknown"},
		{FromMS: start + 1000, ToMS: start + 2000, State: "no_data", Count: "0", SpanStatus: &zero},
	}
	d = timeViewDataset(r, false)
	if d == nil || d.Completeness != nil {
		t.Fatal(d)
	}
	for _, key := range []string{"status.error", "status.ok", "status.unset", "status.unknown", "status.error_fraction"} {
		if d.Rows[0].Values[key] != nil || d.Rows[0].Cells[key].State != "unknown" || d.Rows[0].Cells[key].Support != nil {
			t.Fatal("unknown interval manufactured status evidence", key, d.Rows[0])
		}
	}
	if *d.Rows[1].Values["status.unknown"] != "0" || d.Rows[1].Cells["status.unknown"].State != "derived" || d.Rows[1].Cells["status.error_fraction"].State != "no_data" {
		t.Fatal("known-zero interval lost its status partition", d.Rows[1])
	}
}

func TestViewStatusRejectsMissingForeignOrMismatchedAccounting(t *testing.T) {
	for _, edit := range []func(*Result){
		func(r *Result) { r.SpanStatus = nil },
		func(r *Result) { r.SpanStatus.Count = "11"; r.SpanStatus.Unknown = "2" },
		func(r *Result) { r.SpanStatus.Error = "9" },
		func(r *Result) { r.Question.SpanStatus = false },
	} {
		r := statusViewFixture(t)
		edit(r)
		if summaryViewDataset(r, true) != nil {
			t.Fatal("invalid source status produced summary fields", r)
		}
	}
	for _, status := range []*SpanStatusSummary{nil, {Count: "5", Error: "1", OK: "1", Unset: "1", Unknown: "2"}, {Count: "4", Error: "1", OK: "1", Unset: "1", Unknown: "0"}} {
		r := statusViewFixture(t)
		r.Groups[0], _ = json.Marshal(map[string]any{"g0": nil, "count": "4", "spanStatus": status})
		if groupViewDataset(r, true) != nil {
			t.Fatal("invalid grouped status was adopted", status)
		}
	}
	r := statusViewFixture(t)
	zero := emptySpanStatus()
	r.TimeSeries = []TimePoint{{FromMS: r.Question.FromMS, ToMS: r.Question.FromMS + 1000, State: "unknown", SpanStatus: &zero}}
	if timeViewDataset(r, false) != nil {
		t.Fatal("unknown interval accepted a supplied status DTO")
	}
}

func TestViewStatusPreservesNumericAndDurationProfiles(t *testing.T) {
	binding := numericBinding()
	r := measureGroupFixture(t, &binding, numericValues("1", "2", "3"), []uint64{0, 0, 1})
	r.Question.SpanStatus = true
	r.SpanStatus = &SpanStatusSummary{"3", "1", "1", "0", "1"}
	var group map[string]any
	if err := json.Unmarshal(r.Groups[0], &group); err != nil {
		t.Fatal(err)
	}
	group["spanStatus"] = r.SpanStatus
	r.Groups[0], _ = json.Marshal(group)
	d := groupViewDataset(r, true)
	if d == nil || len(d.Fields) != 17 || *d.Rows[0].Values["numeric.mean"] != "2" || *d.Rows[0].Values["duration.mean"] != "1/3" || *d.Rows[0].Values["duration.p95"] != "1" || *d.Rows[0].Values["status.error_fraction"] != "1/3" || d.Rows[0].Cells["numeric.mean"].Support.Valid != "3" || d.Rows[0].Cells["status.error_fraction"].Support != nil {
		t.Fatal("status changed other profile semantics", d)
	}
}

func TestViewStatusComparisonRetainsLegacyPercentagePointsExplicitly(t *testing.T) {
	a := measuredCohort([]uint64{10, 20}, SpanStatusSummary{"2", "1", "0", "1", "0"})
	b := measuredCohort([]uint64{30, 50, 100}, SpanStatusSummary{"3", "1", "1", "0", "1"})
	result, err := CompareCohorts(ComparisonRequest{Baseline: a, Selection: b})
	if err != nil || result.ViewComparison != nil || result.Measures == nil || result.Measures.ErrorDifferencePP != "-16.67" {
		t.Fatal("status comparison was mislabeled as a bounded fraction or lost PP behavior", result, err)
	}
	withoutStatus := detached(a)
	withoutStatus.Question.SpanStatus, withoutStatus.SpanStatus = false, nil
	for i := range withoutStatus.Groups {
		withoutStatus.Groups[i].SpanStatus = nil
	}
	mixed, err := CompareCohorts(ComparisonRequest{Baseline: withoutStatus, Selection: b})
	if err != nil || mixed.ViewComparison != nil || mixed.Measures.ErrorDifferencePP != "" {
		t.Fatal("one status operand invented a comparative fraction", mixed, err)
	}
	held := heldSummaryView(a)
	if held == nil || held.Origin != "browser-held" || held.Completeness != nil || *held.Rows[0].Values["status.error_fraction"] != "1/2" {
		t.Fatal("held summary silently discarded its requested status profile", held)
	}
	a.SpanStatus = nil
	if _, err = CompareCohorts(ComparisonRequest{Baseline: a, Selection: b}); err == nil || !strings.Contains(err.Error(), "captured span status required") {
		t.Fatal("missing requested status profile was not explicitly rejected", err)
	}
}

func TestSourceSuppliesStatusViewsWithoutAdditionalReads(t *testing.T) {
	reads := 0
	source := Source{Endpoint: "http://source.invalid", Client: &http.Client{Transport: countedValuesTransport(func(req *http.Request) (*http.Response, error) {
		reads++
		body, _ := io.ReadAll(req.Body)
		sql, text := string(body), ""
		switch {
		case strings.Contains(sql, " AS from_ms"):
			text = `{"from_ms":"1000","count":"4","status_error":"1","status_ok":"1","status_unset":"1","status_unknown":"1"}`
		case strings.Contains(sql, "GROUP BY g0"):
			text = `{"g0":"a","count":"4","status_error":"1","status_ok":"1","status_unset":"1","status_unknown":"1"}`
		case strings.Contains(sql, "SELECT DISTINCT"):
		case strings.Contains(sql, "count()"):
			text = `{"count":"4","status_error":"1","status_ok":"1","status_unset":"1","status_unknown":"1"}`
		}
		return &http.Response{StatusCode: 200, Header: http.Header{}, Body: io.NopCloser(strings.NewReader(text))}, nil
	})}}
	q := basic()
	q.FromMS, q.ToMS, q.BucketMS, q.SpanStatus = 1000, 3000, 1000, true
	q.GroupBy = []Field{{Kind: "column", Key: "service"}}
	r, err := source.Fetch(context.Background(), q)
	if err != nil || reads != 5 || r == nil || r.ViewDataset == nil || r.TimeDataset == nil || r.SummaryDataset == nil {
		t.Fatal("status projection changed source execution", r, err, reads)
	}
	if *r.ViewDataset.Rows[0].Values["status.error_fraction"] != "1/4" || *r.SummaryDataset.Rows[0].Values["status.error_fraction"] != "1/4" || r.TimeDataset.Rows[1].Cells["status.error_fraction"].State != "no_data" || r.ViewDataset.Partition.Total != "4" {
		t.Fatal("source status accounting was not carried into every view", r)
	}
}

// Optional exact DTO emission lets the separate OpenAPI gate consume these same
// Go constructors without giving ordinary Go tests a Python/tooling dependency.
func TestViewStatusWireEncoding(t *testing.T) {
	r := statusViewFixture(t)
	start := r.Question.FromMS
	r.TimeSeries = []TimePoint{
		{FromMS: start, ToMS: start + 1000, State: "observed", Count: "4", SpanStatus: &SpanStatusSummary{"4", "1", "0", "2", "1"}},
		{FromMS: start + 1000, ToMS: start + 2000, State: "no_data", Count: "0", SpanStatus: &SpanStatusSummary{"0", "0", "0", "0", "0"}},
		{FromMS: start + 2000, ToMS: start + 3000, State: "observed", Count: "6", SpanStatus: &SpanStatusSummary{"6", "1", "3", "2", "0"}},
	}
	fixtures := map[string]*ViewDataset{"ranked": groupViewDataset(r, true), "summary": summaryViewDataset(r, true), "intervals": timeViewDataset(r, true)}
	r.TimeSeries[0].State, r.TimeSeries[0].Count, r.TimeSeries[0].SpanStatus = "unknown", "", nil
	fixtures["unknown-intervals"] = timeViewDataset(r, false)
	r.State, r.Count, r.Groups, r.TimeSeries = "no_data", "0", nil, nil
	zero := emptySpanStatus()
	r.SpanStatus = &zero
	fixtures["empty-ranked"], fixtures["empty-summary"] = groupViewDataset(r, true), summaryViewDataset(r, true)
	for name, d := range fixtures {
		if d == nil {
			t.Fatal("wire fixture unavailable", name)
		}
		wire, err := json.Marshal(d)
		if err != nil {
			t.Fatal(name, err)
		}
		var decoded struct {
			Rows []struct {
				Values map[string]any `json:"values"`
			} `json:"rows"`
		}
		if err = json.Unmarshal(wire, &decoded); err != nil {
			t.Fatal(name, err)
		}
		for _, row := range decoded.Rows {
			for _, key := range []string{"status.error", "status.ok", "status.unset", "status.unknown", "status.error_fraction"} {
				value, present := row.Values[key]
				if !present {
					t.Fatal(name, "lost explicit status value", key)
				}
				if value != nil {
					if _, ok := value.(string); !ok {
						t.Fatal(name, "coerced exact status value", key, value)
					}
				}
			}
		}
	}
	if file := os.Getenv("GOO_STATUS_VIEW_FIXTURES"); file != "" {
		wire, err := json.Marshal(fixtures)
		if err != nil {
			t.Fatal(err)
		}
		if err = os.WriteFile(file, wire, 0600); err != nil {
			t.Fatal(err)
		}
	}
}
