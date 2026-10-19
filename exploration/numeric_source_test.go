package exploration

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
)

func TestNumericSQLBindingAndPackageRoundTrip(t *testing.T) {
	q := basic()
	b := numericBinding()
	b.Field.Key = "value'\\x"
	q.Numeric = &b
	q.BucketMS = 10000
	q.GroupTimeSeries = true
	q.GroupBy = []Field{{Kind: "resource", Key: "service.name"}}
	p, e := Compile(q)
	if e != nil {
		t.Fatal(e)
	}
	for _, sql := range []string{p.Count, p.Groups, p.TimeSeries, p.GroupedTimeSeries} {
		if strings.Contains(sql, b.Field.Key) || !strings.Contains(sql, "toInt64OrNull") || !strings.Contains(sql, "toInt256") || !strings.Contains(sql, " AS numeric") {
			t.Fatal("unsafe or missing numeric lowering", sql)
		}
	}
	found := false
	for _, v := range p.Parameters {
		found = found || v == b.Field.Key
	}
	if !found {
		t.Fatal("key not parameterized")
	}
	pkg, e := DraftTelemetryPackage(q, "numeric", "Numeric", 1)
	if e != nil {
		t.Fatal(e)
	}
	plan, e := PlanTelemetryPackage(pkg, q.FromMS, q.ToMS, nil)
	if e != nil {
		t.Fatal(e)
	}
	parsed, e := ReadTelemetryPackage(plan.YAML)
	if e != nil || !reflect.DeepEqual(parsed.Numeric, q.Numeric) || !reflect.DeepEqual(parsed.Question(0, 60000).Numeric, q.Numeric) {
		t.Fatal(parsed, e)
	}
	measures, e := q.RecordMeasures()
	if e != nil || measures.Numeric == nil || len(measures.Measures) != 6 || measures.Measures[5].Unit != "{record}" {
		t.Fatal(measures, e)
	}
	measures.Numeric.Unit = "changed"
	pkg.Numeric.Unit = "changed"
	if q.Numeric.Unit != "{row}" {
		t.Fatal("planner/package aliased caller")
	}
	q.Numeric.Type = "float64"
	if _, e = Compile(q); e == nil {
		t.Fatal("unsupported type lowered")
	}
}

func TestNumericValidationRejectsCorruptSummaries(t *testing.T) {
	b := numericBinding()
	s, _ := EvaluateNumericValues(b, numericValues("-5", "10"))
	if s.Validate(b) != nil {
		t.Fatal(s)
	}
	for _, mutate := range []func(*NumericSummary){func(s *NumericSummary) { s.Valid = "3" }, func(s *NumericSummary) { s.Sum = "99" }, func(s *NumericSummary) { s.Unit = "s" }, func(s *NumericSummary) { s.Min = "01" }, func(s *NumericSummary) { s.Max = "9223372036854775808" }, func(s *NumericSummary) { s.State = "no_data" }, func(s *NumericSummary) { s.Count = "02" }} {
		bad := s
		mutate(&bad)
		if bad.Validate(b) == nil {
			t.Fatal("corrupt numeric summary accepted", bad)
		}
	}
}

func TestNumericSourceCaptureAndSameCountDisagreement(t *testing.T) {
	q := basic()
	q.Numeric = new(NumericBinding)
	*q.Numeric = numericBinding()
	q.Limit = 1
	q.BucketMS = 10000
	q.GroupBy = []Field{{Kind: "column", Key: "service"}}
	q.GroupTimeSeries = true
	total, _ := EvaluateNumericValues(*q.Numeric, numericValues("1", "3"))
	different, _ := EvaluateNumericValues(*q.Numeric, numericValues("2", "4"))
	changed := false
	bad := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data, _ := io.ReadAll(r.Body)
		sql := string(data)
		switch {
		case strings.Contains(sql, " AS timestamp_ns"):
			fmt.Fprintln(w, `{"timestamp_ns":"1","duration_ns":"1"}`)
		case strings.Contains(sql, "SELECT DISTINCT"):
			return
		default:
			row := map[string]any{"count": "2", "numeric": total}
			if strings.Contains(sql, "GROUP BY") {
				row["g0"] = "ordinary"
				if changed {
					row["numeric"] = different
				}
				if strings.Contains(sql, "from_ms") {
					row["from_ms"] = "0"
				}
			}
			if bad {
				row["numeric"] = nil
			}
			json.NewEncoder(w).Encode(row)
		}
	}))
	defer server.Close()
	src := Source{Endpoint: server.URL}
	r, e := src.Fetch(context.Background(), q)
	if e != nil || r.State != "observed" || r.Numeric.Sum != "4" || len(r.Rows) != 1 || r.TimeSeries[1].Numeric.State != "no_data" {
		t.Fatal(r, e)
	}
	store := CaptureStore{Dir: t.TempDir()}
	saved, e := store.Create(context.Background(), src, "Numeric", CaptureQuestion{Kind: "explore", Explore: &q}, "")
	if e != nil {
		t.Fatal(e)
	}
	loaded, e := store.Load(saved.ID)
	if e != nil || !reflect.DeepEqual(loaded.Explore.Numeric, saved.Explore.Numeric) {
		t.Fatal(e)
	}
	loaded.Explore.Numeric.Valid = "0"
	if _, e = store.publish(*loaded); e == nil {
		t.Fatal("invalid coverage capture published")
	}
	changed = true
	r, e = src.Fetch(context.Background(), q)
	if e != nil || r.State != "changed_between_reads" {
		t.Fatal("same-count numeric disagreement hidden", r, e)
	}
	saved, e = store.Create(context.Background(), src, "Changed", CaptureQuestion{Kind: "explore", Explore: &q}, "")
	if e != nil {
		t.Fatal(e)
	}
	if _, e = store.Load(saved.ID); e != nil {
		t.Fatal("disagreement not preserved", e)
	}
	bad = true
	if _, e = src.Fetch(context.Background(), q); e == nil {
		t.Fatal("missing numeric state treated as zero")
	}
}
