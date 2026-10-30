package exploration

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/casselc/goo/identity"
	"io"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"testing"
)

func viewFixture() *Result {
	q := basic()
	q.GroupBy = []Field{{Kind: "attribute", Key: "team"}}
	return &Result{Question: q, Basis: strings.Repeat("a", 64), Count: "18446744073709551615", State: "observed", HasMore: true, CapturedAt: "2026-10-30T00:00:00Z", Groups: []json.RawMessage{json.RawMessage(`{"g0":null,"count":"18446744073709551614"}`), json.RawMessage(`{"g0":"","count":"1"}`)}}
}
func TestViewDatasetPreservesExactSupportAndScope(t *testing.T) {
	r := viewFixture()
	d := groupViewDataset(r, true)
	if d == nil || d.Version != "goo.dataset/1" || d.Quality.Extent != "complete" || d.Quality.Sampling != "unknown" || d.Completeness == nil || !reflect.DeepEqual(d.Completeness.Aspects, []string{"groups"}) {
		t.Fatal(d)
	}
	if d.Rows[0].Values["category"] != nil || d.Rows[1].Values["category"] == nil || *d.Rows[1].Values["category"] != "" || d.Rows[0].Key == d.Rows[1].Key || *d.Rows[0].Values["count"] != "18446744073709551614" {
		t.Fatal(d.Rows)
	}
	var q Query
	if json.Unmarshal(d.Scope.Selector, &q) != nil || !reflect.DeepEqual(q, r.Question) {
		t.Fatal("lost selector", string(d.Scope.Selector))
	}
	before := string(d.Scope.Selector)
	r.Question.GroupBy[0].Key = "mutated"
	if string(d.Scope.Selector) != before {
		t.Fatal("caller mutated snapshot")
	}
	if d.Revision != r.Basis || d.Scope.Generation != r.Basis || d.Completeness.ScopeID != d.Scope.ID || d.Completeness.Revision != d.Revision {
		t.Fatal("unbound attestation", d)
	}
}
func TestViewDatasetCompletenessComesFromReadNotCount(t *testing.T) {
	r := viewFixture()
	// Even perfect count agreement is insufficient without an exhaustive read.
	d := groupViewDataset(r, false)
	if d.Completeness != nil || d.Quality.Extent != "unknown" {
		t.Fatal(d)
	}
	// A read can be complete independently of a bounded/hasMore record page.
	d = groupViewDataset(r, true)
	if d.Completeness == nil || !r.HasMore {
		t.Fatal(d)
	}
	r.State = "changed_between_reads"
	d = groupViewDataset(r, true)
	if d.Completeness != nil || d.Quality.Extent != "unknown" {
		t.Fatal(d)
	}
	r.State = "no_data"
	r.Groups = nil
	r.Count = "0"
	d = groupViewDataset(r, true)
	if d.Completeness == nil || len(d.Rows) != 0 {
		t.Fatal(d)
	}
}
func TestViewDatasetRejectsUnsupportedOrMalformedGroups(t *testing.T) {
	cases := []func(*Result){func(r *Result) { r.Groups = []json.RawMessage{json.RawMessage(`{"g0":1,"count":"1"}`)} }, func(r *Result) { r.Groups = []json.RawMessage{json.RawMessage(`{"g0":"x","count":"01"}`)} }, func(r *Result) { r.Groups = []json.RawMessage{json.RawMessage(`{"count":"1"}`)} }, func(r *Result) { r.Groups = append(r.Groups, r.Groups[0]) }, func(r *Result) {
		r.Question.GroupBy = append(r.Question.GroupBy, Field{"column", "service"}, Field{"column", "name"})
	}, func(r *Result) { r.Question.Summary = "span-duration" }, func(r *Result) { r.Basis = "unbound" }}
	for i, edit := range cases {
		r := viewFixture()
		edit(r)
		if groupViewDataset(r, true) != nil {
			t.Fatal("admitted", i)
		}
	}
}
func TestSourceSuppliesDatasetWithoutAdditionalReads(t *testing.T) {
	reads := 0
	source := Source{Endpoint: "http://source.invalid", Client: &http.Client{Transport: countedValuesTransport(func(req *http.Request) (*http.Response, error) {
		reads++
		body, _ := io.ReadAll(req.Body)
		sql := string(body)
		text := ""
		switch {
		case strings.Contains(sql, "GROUP BY"):
			if !strings.Contains(sql, "LIMIT 1001") || req.URL.Query().Get("result_overflow_mode") != "throw" {
				t.Fatal("no exhaustive group budget")
			}
			text = `{"g0":"alpha","count":"3"}`
		case strings.Contains(sql, "SELECT DISTINCT"):
			text = ""
		case strings.Contains(sql, "count()"):
			text = `{"count":"3"}`
		default:
			text = "{\"timestamp_ns\":\"1\",\"duration_ns\":\"0\"}\n{\"timestamp_ns\":\"2\",\"duration_ns\":\"0\"}\n{\"timestamp_ns\":\"3\",\"duration_ns\":\"0\"}"
		}
		return &http.Response{StatusCode: 200, Header: http.Header{}, Body: io.NopCloser(strings.NewReader(text))}, nil
	})}}
	q := basic()
	q.GroupBy = []Field{{Kind: "column", Key: "service"}}
	result, err := source.Fetch(context.Background(), q)
	if err != nil || reads != 4 || result.ViewDataset == nil || result.ViewDataset.Completeness == nil {
		t.Fatal(result, err, reads)
	}
	if *result.ViewDataset.Rows[0].Values["count"] != "3" || !result.HasMore {
		t.Fatal(result)
	}
}
func TestViewDatasetStableKeysAndIdentity(t *testing.T) {
	for n := 0; n < 100; n++ {
		r := viewFixture()
		r.Groups = nil
		for i := 0; i < n; i++ {
			raw, _ := json.Marshal(map[string]string{"g0": fmt.Sprintf("category-%d", i), "count": fmt.Sprint(i + 1)})
			r.Groups = append(r.Groups, raw)
		}
		a := groupViewDataset(r, true)
		b := groupViewDataset(r, true)
		if !reflect.DeepEqual(a, b) || len(a.Rows) != n {
			t.Fatal("unstable", n)
		}
		r.Question.ToMS++
		c := groupViewDataset(r, true)
		if c.Scope.ID == a.Scope.ID {
			t.Fatal("scope identity reused", n)
		}
	}
}

func TestViewDatasetCategoryByteBudgetAndHashedKeys(t *testing.T) {
	r := viewFixture()
	// JSON escaping may exceed the key budget although the actual category fits.
	value := strings.Repeat("\"", 1024)
	raw, _ := json.Marshal(map[string]string{"g0": value, "count": "1"})
	r.Groups = []json.RawMessage{raw}
	d := groupViewDataset(r, true)
	if d == nil || len(d.Rows[0].Key) != len("category:")+64 || *d.Rows[0].Values["category"] != value {
		t.Fatal("valid escaped category lost", d)
	}
	// Do not truncate identities or silently drop an over-budget group.
	for _, value := range []string{strings.Repeat("x", 1025), strings.Repeat("界", 342)} {
		raw, _ := json.Marshal(map[string]string{"g0": value, "count": "1"})
		r.Groups = []json.RawMessage{raw}
		if groupViewDataset(r, true) != nil {
			t.Fatal("over-byte-budget snapshot admitted")
		}
	}
}

func timeViewFixture(t *testing.T, q Query, raw []json.RawMessage) *Result {
	t.Helper()
	points, err := DecodeTimeSeries(q, raw)
	if err != nil {
		t.Fatal(err)
	}
	return &Result{Question: q, TimeSeries: points, Basis: strings.Repeat("b", 64), State: "observed", CapturedAt: "2026-10-30T00:00:00Z"}
}
func TestTimeDatasetExactIntervalsEmptyAndPartialLastBin(t *testing.T) {
	q := basic()
	q.FromMS = 1000
	q.ToMS = 3500
	q.BucketMS = 1000
	r := timeViewFixture(t, q, []json.RawMessage{json.RawMessage(`{"from_ms":"2000","count":"18446744073709551615"}`)})
	d := timeViewDataset(r, true)
	if d == nil || d.Shape != "time-series" || d.Completeness == nil || !reflect.DeepEqual(d.Completeness.Aspects, []string{"intervals"}) || len(d.Rows) != 3 {
		t.Fatal(d)
	}
	if *d.Rows[0].Values["count"] != "0" || d.Rows[0].State != "no_data" || *d.Rows[1].Values["count"] != "18446744073709551615" || *d.Rows[2].Values["toMs"] != "3500" {
		t.Fatal(d.Rows)
	}
	if *d.Rows[2].Values["fromMs"] != "3000" || d.Quality.Sampling != "unknown" || d.Interval.FromMS != "1000" || d.Interval.ToMS != "3500" || d.Interval.BucketMS != "1000" {
		t.Fatal(d)
	}
	empty := timeViewDataset(timeViewFixture(t, q, nil), true)
	for _, row := range empty.Rows {
		if row.State != "no_data" || *row.Values["count"] != "0" {
			t.Fatal(row)
		}
	}
	if empty.Completeness == nil {
		t.Fatal("exhaustive empty query lost interval coverage")
	}
}
func TestTimeDatasetDoesNotInventMissingOrUnknownIntervals(t *testing.T) {
	q := basic()
	q.FromMS = 1000
	q.ToMS = 4000
	q.BucketMS = 1000
	r := timeViewFixture(t, q, nil)
	r.TimeSeries = []TimePoint{{FromMS: 1000, ToMS: 2000, Count: "1", State: "observed"}, {FromMS: 3000, ToMS: 4000, State: "unknown"}}
	d := timeViewDataset(r, false)
	if d == nil || d.Completeness != nil || d.Quality.Extent != "unknown" || len(d.Rows) != 2 || d.Rows[1].Values["count"] != nil || d.Rows[1].State != "unknown" {
		t.Fatal(d)
	}
	if timeViewDataset(r, true).Completeness != nil {
		t.Fatal("unknown/gapped intervals marked complete")
	}
	r.TimeSeries[1].Count = "0"
	if timeViewDataset(r, false) != nil {
		t.Fatal("unknown interval zero-filled")
	}
	r = timeViewFixture(t, q, nil)
	r.State = "changed_between_reads"
	if timeViewDataset(r, true).Completeness != nil {
		t.Fatal("diagnostic read marked complete")
	}
	r.State = "observed"
	if timeViewDataset(r, false).Completeness != nil {
		t.Fatal("separate count/page manufactured exhaustivity")
	}
}
func TestTimeDatasetRejectsInvalidBinsAndStates(t *testing.T) {
	q := basic()
	q.FromMS = 1000
	q.ToMS = 3500
	q.BucketMS = 1000
	for _, edit := range []func(*Result){
		func(r *Result) { r.TimeSeries[0].FromMS++ },
		func(r *Result) { r.TimeSeries[0].ToMS++ },
		func(r *Result) { r.TimeSeries[1].FromMS = r.TimeSeries[0].FromMS },
		func(r *Result) { r.TimeSeries[0].State = "observed" },
		func(r *Result) { r.TimeSeries[0].Count = "01" },
		func(r *Result) { r.Question.Summary = "span-duration" },
	} {
		r := timeViewFixture(t, q, nil)
		edit(r)
		if timeViewDataset(r, true) != nil {
			t.Fatal("invalid interval adapted", r)
		}
	}
}
func TestTimeDatasetGeneratedPartitions(t *testing.T) {
	for n := 1; n <= 100; n++ {
		q := basic()
		q.FromMS = int64(1000 + n)
		q.BucketMS = 1000
		q.ToMS = q.FromMS + int64(n)*q.BucketMS - int64(n%997)
		r := timeViewFixture(t, q, nil)
		d := timeViewDataset(r, true)
		if d == nil || d.Completeness == nil || len(d.Rows) != n {
			t.Fatal("partition", n, d)
		}
		next := q.FromMS
		for i, row := range d.Rows {
			lo, _ := strconv.ParseInt(*row.Values["fromMs"], 10, 64)
			hi, _ := strconv.ParseInt(*row.Values["toMs"], 10, 64)
			if lo != next || hi <= lo || hi > q.ToMS || row.State != "no_data" || *row.Values["count"] != "0" {
				t.Fatal(n, i, row)
			}
			next = hi
		}
		if next != q.ToMS {
			t.Fatal("uncovered end", n)
		}
	}
}
func TestViewDatasetMultiFieldTupleIdentity(t *testing.T) {
	r := viewFixture()
	r.Question.GroupBy = append(r.Question.GroupBy, Field{Kind: "column", Key: "service"})
	r.Groups = []json.RawMessage{json.RawMessage(`{"g0":null,"g1":"","count":"1"}`), json.RawMessage(`{"g0":"","g1":null,"count":"2"}`), json.RawMessage(`{"g0":"a,b","g1":"c","count":"3"}`), json.RawMessage(`{"g0":"a","g1":"b,c","count":"4"}`)}
	d := groupViewDataset(r, true)
	if d == nil || len(d.Fields) != 3 || d.Fields[1].Key != "category1" || len(d.Rows) != 4 || d.Completeness == nil {
		t.Fatal(d)
	}
	keys := map[string]bool{}
	for _, row := range d.Rows {
		if keys[row.Key] || len(row.Values) != 3 {
			t.Fatal("tuple collision", row)
		}
		keys[row.Key] = true
	}
	if d.Rows[0].Values["category"] != nil || *d.Rows[0].Values["category1"] != "" || d.Rows[1].Values["category1"] != nil {
		t.Fatal("tuple null/empty lost")
	}
	r.Groups = append(r.Groups, r.Groups[0])
	if groupViewDataset(r, true) != nil {
		t.Fatal("duplicate tuple accepted")
	}
}

func TestSourceSuppliesTimeDatasetWithoutAdditionalReads(t *testing.T) {
	reads := 0
	source := Source{Endpoint: "http://source.invalid", Client: &http.Client{Transport: countedValuesTransport(func(req *http.Request) (*http.Response, error) {
		reads++
		body, _ := io.ReadAll(req.Body)
		sql := string(body)
		text := ""
		switch {
		case strings.Contains(sql, "GROUP BY from_ms"):
			if !strings.Contains(sql, "LIMIT 241") || req.URL.Query().Get("result_overflow_mode") != "throw" {
				t.Fatal("unbounded intervals")
			}
			text = `{"from_ms":"1000","count":"3"}`
		case strings.Contains(sql, "SELECT DISTINCT"):
			text = ""
		case strings.Contains(sql, "count()"):
			text = `{"count":"3"}`
		default:
			text = "{\"timestamp_ns\":\"1\",\"duration_ns\":\"0\"}\n{\"timestamp_ns\":\"2\",\"duration_ns\":\"0\"}\n{\"timestamp_ns\":\"3\",\"duration_ns\":\"0\"}"
		}
		return &http.Response{StatusCode: 200, Header: http.Header{}, Body: io.NopCloser(strings.NewReader(text))}, nil
	})}}
	q := basic()
	q.FromMS = 1000
	q.ToMS = 3500
	q.BucketMS = 1000
	result, err := source.Fetch(context.Background(), q)
	if err != nil || reads != 4 || result.TimeDataset == nil || result.TimeDataset.Completeness == nil || result.ViewDataset != nil {
		t.Fatal(result, err, reads)
	}
	if len(result.TimeDataset.Rows) != 3 || *result.TimeDataset.Rows[0].Values["count"] != "3" || *result.TimeDataset.Rows[2].Values["toMs"] != "3500" {
		t.Fatal(result.TimeDataset)
	}
}

func TestViewDatasetMultiFieldCardinalityBudget(t *testing.T) {
	r := viewFixture()
	r.Question.GroupBy = append(r.Question.GroupBy, Field{Kind: "attribute", Key: "region"})
	r.Groups = nil
	for i := 0; i < 1000; i++ {
		raw, _ := json.Marshal(map[string]string{"g0": fmt.Sprint(i / 13), "g1": fmt.Sprint(i % 13), "count": "1"})
		r.Groups = append(r.Groups, raw)
	}
	d := groupViewDataset(r, true)
	if d == nil || len(d.Rows) != 1000 || !reflect.DeepEqual(d.Dimensions, []string{"category", "category1"}) {
		t.Fatal("bounded multidimensional result refused")
	}
	keys := map[string]bool{}
	for _, row := range d.Rows {
		if keys[row.Key] || len(row.Values) != 3 {
			t.Fatal("tuple identity lost")
		}
		keys[row.Key] = true
	}
	raw, _ := json.Marshal(map[string]string{"g0": "extra", "g1": "extra", "count": "1"})
	r.Groups = append(r.Groups, raw)
	if groupViewDataset(r, true) != nil {
		t.Fatal("oversized grouped snapshot silently admitted")
	}
}

func TestViewDatasetCheckedEnvironmentPolicySurvives(t *testing.T) {
	r := viewFixture()
	r.Question.Environment = ""
	r.Question.EnvironmentSelector = &identity.EnvironmentSelector{Version: identity.EnvironmentSelectorVersion, Policy: "effective", Scope: identity.EnvironmentScope{Mode: "unreported"}}
	r.Question.Where = &Predicate{Field: &Field{Kind: "attribute", Key: "tenant"}, Op: "eq", Value: "team-a"}
	d := groupViewDataset(r, true)
	if d == nil {
		t.Fatal("checked scope refused")
	}
	var q Query
	if json.Unmarshal(d.Scope.Selector, &q) != nil || q.EnvironmentSelector == nil || q.EnvironmentSelector.Policy != "effective" || q.EnvironmentSelector.Scope.Mode != "unreported" || !reflect.DeepEqual(q, r.Question) {
		t.Fatal("checked selector narrowed or lost", string(d.Scope.Selector))
	}
	r.Question.EnvironmentSelector.Policy = "current-only"
	if groupViewDataset(r, true) != nil {
		t.Fatal("invalid checked selector admitted")
	}
}
