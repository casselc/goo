package exploration

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"testing"
)

func groupedCountFixture395(t *testing.T) *Result {
	t.Helper()
	q := Query{Signal: "logs", Environment: "prod", FromMS: 1000, ToMS: 4500, Limit: 1, GroupBy: []Field{{Kind: "attribute", Key: "route"}}, BucketMS: 1000, GroupTimeSeries: true}
	cells := []json.RawMessage{json.RawMessage(`{"g0":null,"from_ms":"1000","count":"1"}`), json.RawMessage(`{"g0":"","from_ms":"2000","count":"2"}`), json.RawMessage(`{"g0":"雪","from_ms":"4000","count":"3"}`)}
	grouped, e := DecodeGroupedTimeSeries(q, cells)
	if e != nil {
		t.Fatal(e)
	}
	overall, e := DecodeTimeSeries(q, []json.RawMessage{json.RawMessage(`{"from_ms":"1000","count":"1"}`), json.RawMessage(`{"from_ms":"2000","count":"2"}`), json.RawMessage(`{"from_ms":"4000","count":"3"}`)})
	if e != nil {
		t.Fatal(e)
	}
	return &Result{Question: q, Basis: strings.Repeat("a", 64), CapturedAt: "2026-11-10T00:00:00Z", Count: "6", State: "observed", GroupedTimeSeries: grouped, TimeSeries: overall, Groups: []json.RawMessage{json.RawMessage(`{"g0":null,"count":"1"}`), json.RawMessage(`{"g0":"","count":"2"}`), json.RawMessage(`{"g0":"雪","count":"3"}`)}}
}
func TestGroupedCountDataset395IdentityZeroAndPartial(t *testing.T) {
	r := groupedCountFixture395(t)
	d := groupedTimeViewDataset(r, true)
	if d == nil || d.Shape != "grouped-time-series" || len(d.Rows) != 12 || d.Completeness == nil || !reflect.DeepEqual(d.Completeness.Aspects, []string{"grouped-intervals"}) || d.Partition != nil || d.Quality.Sampling != "unknown" {
		t.Fatal("grouped source guarantees changed", d)
	}
	seen := map[string]bool{}
	positive, zero := 0, 0
	for _, row := range d.Rows {
		from, to := *row.Values["fromMs"], *row.Values["toMs"]
		key, _ := json.Marshal([]any{row.Values["category"], from, to})
		hash := sha256.Sum256(key)
		if row.Key != "group-interval:"+hex.EncodeToString(hash[:]) || seen[row.Key] {
			t.Fatal("group×interval identity collision")
		}
		seen[row.Key] = true
		if from == "4000" && to != "4500" {
			t.Fatal("partial final interval widened")
		}
		if row.State == "observed" {
			positive++
			if *row.Values["count"] == "0" {
				t.Fatal("positive cell lost count")
			}
		} else if row.State == "no_data" {
			zero++
			if *row.Values["count"] != "0" {
				t.Fatal("known empty cell changed")
			}
		} else {
			t.Fatal("coherent cell unavailable")
		}
	}
	if positive != 3 || zero != 9 {
		t.Fatal("wrong exact dense matrix", positive, zero)
	}
	before, _ := json.Marshal(d)
	for _, g := range r.GroupedTimeSeries {
		if g.Values[0] != nil {
			*g.Values[0] = "caller-mutated"
		}
	}
	after, _ := json.Marshal(d)
	if string(before) != string(after) {
		t.Fatal("projection retained caller category pointers")
	}
}
func TestGroupedCountDataset395ChangedAndRefusals(t *testing.T) {
	for _, change := range []func(*Result){func(r *Result) { r.State = "changed_between_reads" }, func(r *Result) { r.Count = "7" }, func(r *Result) { r.Groups = r.Groups[:2] }} {
		r := groupedCountFixture395(t)
		change(r)
		d := groupedTimeViewDataset(r, true)
		if d == nil || d.Completeness != nil || d.Quality.Extent != "unknown" {
			t.Fatal("disagreement acquired guarantees")
		}
		positives, unknown := 0, 0
		for _, row := range d.Rows {
			if row.State == "observed" && row.Values["count"] != nil {
				positives++
			} else if row.State == "unknown" && row.Values["count"] == nil {
				unknown++
			} else {
				t.Fatal("unattested absence became zero")
			}
		}
		if positives != 3 || unknown != 9 {
			t.Fatal("diagnostic count population changed")
		}
	}
	r := groupedCountFixture395(t)
	if d := groupedTimeViewDataset(r, false); d == nil || d.Completeness != nil {
		t.Fatal("nonexhaustive source promoted")
	}
	r = groupedCountFixture395(t)
	r.State = "no_data"
	r.Count = "0"
	r.Groups = nil
	r.GroupedTimeSeries = []TimeSeriesGroup{}
	r.TimeSeries, _ = DecodeTimeSeries(r.Question, nil)
	d := groupedTimeViewDataset(r, true)
	if d == nil || len(d.Rows) != 0 || d.Completeness == nil {
		t.Fatal("empty source fabricated category")
	}
	r = groupedCountFixture395(t)
	for index := range r.GroupedTimeSeries[0].Points {
		r.GroupedTimeSeries[0].Points[index].Count = "0"
		r.GroupedTimeSeries[0].Points[index].State = "no_data"
	}
	if groupedTimeViewDataset(r, true) != nil {
		t.Fatal("invented all-zero category accepted")
	}
	for _, change := range []func(*Query){func(q *Query) { q.Summary = "span-duration" }, func(q *Query) { q.Numeric = &NumericBinding{} }, func(q *Query) { q.GroupBy = append(q.GroupBy, Field{Kind: "attribute", Key: "extra"}) }, func(q *Query) { q.GroupBy[0] = Field{Kind: "column", Key: "body_pattern_v1"} }} {
		r = groupedCountFixture395(t)
		change(&r.Question)
		if groupedTimeViewDataset(r, true) != nil {
			t.Fatal("unsupported grouped profile projected")
		}
	}
}
func TestGroupedCountDataset395SourceAddsNoRead(t *testing.T) {
	fixture := groupedCountFixture395(t)
	plan, e := Compile(fixture.Question)
	if e != nil {
		t.Fatal(e)
	}
	reads := 0
	source := Source{Endpoint: "http://source.invalid", Client: &http.Client{Transport: countedValuesTransport(func(req *http.Request) (*http.Response, error) {
		reads++
		wire, _ := io.ReadAll(req.Body)
		body := ""
		switch string(wire) {
		case plan.Count:
			body = `{"count":"6"}`
		case plan.Rows:
			body = "{\"timestamp_ns\":\"1001000000\"}\n{\"timestamp_ns\":\"2001000000\"}"
		case plan.Fields:
		case plan.Groups:
			body = strings.Join([]string{string(fixture.Groups[0]), string(fixture.Groups[1]), string(fixture.Groups[2])}, "\n")
		case plan.TimeSeries:
			body = "{\"from_ms\":\"1000\",\"count\":\"1\"}\n{\"from_ms\":\"2000\",\"count\":\"2\"}\n{\"from_ms\":\"4000\",\"count\":\"3\"}"
		case plan.GroupedTimeSeries:
			body = "{\"g0\":null,\"from_ms\":\"1000\",\"count\":\"1\"}\n{\"g0\":\"\",\"from_ms\":\"2000\",\"count\":\"2\"}\n{\"g0\":\"雪\",\"from_ms\":\"4000\",\"count\":\"3\"}"
		default:
			t.Error("unexpected additional source query")
		}
		return &http.Response{StatusCode: 200, Header: http.Header{}, Body: io.NopCloser(strings.NewReader(body))}, nil
	})}}
	result, e := source.Fetch(t.Context(), fixture.Question)
	if e != nil || reads != 6 || result.GroupedTimeDataset == nil || result.GroupedTimeDataset.Completeness == nil {
		t.Fatal("grouped projection changed acquisition", reads, e)
	}
}
func TestGroupedCountDataset395DenseBudgetAndSnapshotRefusal(t *testing.T) {
	r := groupedCountFixture395(t)
	q := r.Question
	q.ToMS = q.FromMS + 240000
	r.Question = q
	r.Count = "1000"
	r.Groups = nil
	r.GroupedTimeSeries = nil
	raw := []json.RawMessage{}
	for group := 0; group < 8; group++ {
		category := strconv.Itoa(group) + strings.Repeat("x", 1023)
		total, _ := json.Marshal(map[string]string{"g0": category, "count": "125"})
		r.Groups = append(r.Groups, total)
		for bin := 0; bin < 125; bin++ {
			cell, _ := json.Marshal(map[string]string{"g0": category, "from_ms": strconv.FormatInt(q.FromMS+int64(bin)*q.BucketMS, 10), "count": "1"})
			raw = append(raw, cell)
		}
	}
	var e error
	r.GroupedTimeSeries, e = DecodeGroupedTimeSeries(q, raw)
	if e != nil {
		t.Fatal(e)
	}
	overall := []json.RawMessage{}
	for bin := 0; bin < 125; bin++ {
		cell, _ := json.Marshal(map[string]string{"from_ms": strconv.FormatInt(q.FromMS+int64(bin)*q.BucketMS, 10), "count": "8"})
		overall = append(overall, cell)
	}
	r.TimeSeries, e = DecodeTimeSeries(q, overall)
	if e != nil {
		t.Fatal(e)
	}
	r.GroupedTimeDataset = groupedTimeViewDataset(r, true)
	if r.GroupedTimeDataset == nil || len(r.GroupedTimeDataset.Rows) != 1920 {
		t.Fatal("existing observed/dense limits changed")
	}
	if token, e := new(ResultSnapshots).Freeze(r); e == nil || token != "" || !strings.Contains(e.Error(), "snapshot bound") {
		t.Fatal("large derived payload widened existing snapshot budget")
	}
	if _, e := DecodeGroupedTimeSeries(q, append(raw, raw[0])); e == nil {
		t.Fatal("more than1000 observed cells admitted")
	}
}
