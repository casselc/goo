package exploration

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSpanStatusAccountingAndSourceDrift(t *testing.T) {
	q := basic()
	q.SpanStatus = true
	q.GroupBy = []Field{{"column", "service"}}
	p, e := Compile(q)
	if e != nil || !strings.Contains(p.Groups, "status_unknown") || strings.Contains(p.Count, "LIMIT") {
		t.Fatal(p, e)
	}
	mode := "valid"
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		sql := string(b)
		switch {
		case strings.Contains(sql, "GROUP BY g0"):
			value := "1"
			if mode == "changed" {
				value = "0"
			}
			if mode == "missing" {
				fmt.Fprintln(w, `{"g0":"a","count":"4"}`)
				return
			}
			fmt.Fprintf(w, `{"g0":"a","count":"4","status_error":"%s","status_ok":"1","status_unset":"1","status_unknown":"%s"}`, value, map[bool]string{true: "2", false: "1"}[mode == "changed"])
		case strings.Contains(sql, "SELECT toString(count())"):
			fmt.Fprintln(w, `{"count":"4","status_error":"1","status_ok":"1","status_unset":"1","status_unknown":"1"}`)
		}
	}))
	defer srv.Close()
	source := Source{Endpoint: srv.URL}
	r, e := source.Fetch(context.Background(), q)
	if e != nil || r.State != "observed" || r.SpanStatus.Unknown != "1" || !strings.Contains(string(r.Groups[0]), `"spanStatus"`) {
		t.Fatal(r, e)
	}
	mode = "changed"
	r, e = source.Fetch(context.Background(), q)
	if e != nil || r.State != "changed_between_reads" {
		t.Fatal(r, e)
	}
	mode = "missing"
	if _, e = source.Fetch(context.Background(), q); e == nil {
		t.Fatal("missing status became success")
	}
	for _, raw := range []string{`{"count":"1","status_error":"0","status_ok":"0","status_unset":"0","status_unknown":"0"}`, `{"count":"0","status_error":"0","status_ok":"0","status_unset":"0"}`} {
		if _, e := DecodeSpanStatus([]byte(raw)); e == nil {
			t.Fatal("invalid accounting accepted")
		}
	}
	if SpanStatusClass("") != "unknown" || SpanStatusClass("Unset") != "unset" || SpanStatusClass("Error") != "error" || SpanStatusClass("OK") != "unknown" {
		t.Fatal("invented status")
	}
	q.Signal = "logs"
	if q.Validate() == nil {
		t.Fatal("log status accepted")
	}
}
func measuredCohort(values []uint64, status SpanStatusSummary) Cohort {
	d := ReferenceDurationSummary(values)
	c := cohort(d.Count)
	c.Question.Summary = "span-duration"
	c.Question.SpanStatus = true
	c.Duration = &d
	c.SpanStatus = &status
	if d.Count == "0" {
		c.State = "no_data"
	} else {
		c.Groups = []ValueCount{{Value: str("a"), Count: d.Count, Duration: &d, SpanStatus: &status}}
	}
	return c
}
func TestMeasureComparisonAndMissingGroup(t *testing.T) {
	a := measuredCohort([]uint64{10, 20}, SpanStatusSummary{"2", "1", "0", "1", "0"})
	b := measuredCohort([]uint64{30, 50, 100}, SpanStatusSummary{"3", "1", "1", "0", "1"})
	b.Question.FromMS += 60000
	b.Question.ToMS += 60000
	r, e := CompareCohorts(ComparisonRequest{Baseline: a, Selection: b})
	if e != nil || r.Measures.MeanDifferenceNS != "45.000" || r.Measures.P95DifferenceNS != "80" || r.Measures.ErrorDifferencePP != "-16.67" {
		t.Fatal(r, e)
	}
	b.Groups[0].Value = nil
	r, e = CompareCohorts(ComparisonRequest{Baseline: a, Selection: b})
	if e != nil || len(r.Rows) != 2 || r.Rows[0].Measures.MeanDifferenceNS != "" {
		t.Fatal(r, e)
	}
	b.SpanStatus = &SpanStatusSummary{"3", "0", "1", "0", "2"}
	if _, e = CompareCohorts(ComparisonRequest{Baseline: a, Selection: b}); e == nil {
		t.Fatal("incoherent grouped status accepted")
	}
	b = measuredCohort(nil, emptySpanStatus())
	r, e = CompareCohorts(ComparisonRequest{Baseline: a, Selection: b})
	if e != nil || r.State != "no_data" || r.Measures != nil {
		t.Fatal(r, e)
	}
	a.State = "changed_between_reads"
	r, e = CompareCohorts(ComparisonRequest{Baseline: a, Selection: a})
	if e != nil || r.State != "unknown" || r.Measures != nil {
		t.Fatal(r, e)
	}
}

func TestSavedComparisonRetainsServerEvidenceAndReopensOffline(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		sql := string(b)
		row := `"count":"2","sum_ns":"30","min_ns":"10","max_ns":"20","quantiles_ns":["20","20","20"],"status_error":"1","status_ok":"0","status_unset":"1","status_unknown":"0"`
		switch {
		case strings.Contains(sql, "GROUP BY g0"):
			fmt.Fprintln(w, `{"g0":"a",`+row+`}`)
		case strings.Contains(sql, "SELECT toString(count())"):
			fmt.Fprintln(w, `{`+row+`}`)
		}
	}))
	source := Source{Endpoint: srv.URL}
	a := basic()
	a.GroupBy = []Field{{"column", "service"}}
	a.Summary = "span-duration"
	a.SpanStatus = true
	b := a
	b.FromMS += 60000
	b.ToMS += 60000
	q := CaptureQuestion{Kind: "comparison", Comparison: &ComparisonQuestion{Baseline: a, Selection: b}}
	store := CaptureStore{Dir: t.TempDir()}
	saved, e := store.Create(context.Background(), source, "Error regression", q, "")
	if e != nil {
		t.Fatal(e)
	}
	if saved.Comparison.Result.Measures.ErrorDifferencePP != "0.00" || saved.Comparison.BaselineID == saved.Comparison.SelectionID {
		t.Fatal(saved)
	}
	first, e := store.Load(saved.Comparison.BaselineID)
	if e != nil || len(first.Explore.Groups) != 1 || first.Explore.SpanStatus.Error != "1" {
		t.Fatal(first, e)
	}

	successor, e := store.Create(context.Background(), source, "Error regression", q, saved.ID)
	if e != nil || successor.Parent != saved.ID || successor.Comparison.BaselineID == saved.Comparison.BaselineID {
		t.Fatal("fixed pair refresh", successor, e)
	}
	child, e := store.Load(successor.Comparison.BaselineID)
	if e != nil || child.Parent != saved.Comparison.BaselineID {
		t.Fatal("child lineage", child, e)
	}
	srv.Close()
	reopened, e := store.Load(saved.ID)
	if e != nil || reopened.Comparison.Result.Basis != saved.Comparison.Result.Basis {
		t.Fatal(reopened, e)
	}
	page, e := reopened.Page("")
	if e != nil || page.Comparison == nil || page.Explore != nil {
		t.Fatal(page, e)
	}
	if _, e = store.Create(context.Background(), source, "Error regression", q, saved.ID); e == nil {
		t.Fatal("outage became empty pair")
	}
	if _, e = store.Load(saved.ID); e != nil {
		t.Fatal("failed refresh lost pair", e)
	}
}

func TestStatusPackageAndIntervalContracts(t *testing.T) {
	q := basic()
	q.SpanStatus = true
	q.BucketMS = 1000
	q.GroupBy = []Field{{"column", "status_class"}}
	p, e := DraftTelemetryPackage(q, "span-status", "Recorded span status", 1)
	if e != nil {
		t.Fatal(e)
	}
	plan, e := PlanTelemetryPackage(p, q.FromMS, q.ToMS, nil)
	if e != nil || !plan.Question.SpanStatus || !strings.Contains(plan.YAML, "spanStatus: true") || !strings.Contains(plan.Plan.Groups, statusClassSQL()) {
		t.Fatal(plan, e)
	}
	decoded, e := ReadTelemetryPackage(plan.YAML)
	if e != nil || !decoded.SpanStatus {
		t.Fatal(decoded, e)
	}
	raw := json.RawMessage(fmt.Sprintf(`{"from_ms":"%d","count":"4","status_error":"1","status_ok":"1","status_unset":"1","status_unknown":"1"}`, q.FromMS))
	points, e := DecodeTimeSeries(q, []json.RawMessage{raw})
	if e != nil || points[0].SpanStatus.Unknown != "1" || points[1].SpanStatus.Count != "0" || ValidateTimeSeries(q, points) != nil {
		t.Fatal(points, e)
	}
	points[0].SpanStatus.Error = "0"
	if ValidateTimeSeries(q, points) == nil {
		t.Fatal("invalid stored interval accepted")
	}
}
