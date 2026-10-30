package exploration

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"reflect"
	"strings"
	"testing"
)

func measureGroupFixture(t *testing.T, b *NumericBinding, values []*string, durations []uint64) *Result {
	t.Helper()
	r := viewFixture()
	r.Question.Numeric = b
	row := map[string]any{"g0": "svc", "count": fmt.Sprint(len(values))}
	if b != nil {
		n, e := EvaluateNumericValues(*b, values)
		if e != nil {
			t.Fatal(e)
		}
		row["numeric"] = &n
		r.Numeric = &n
	}
	if durations != nil {
		r.Question.Summary = "span-duration"
		d := ReferenceDurationSummary(durations)
		row["duration"] = &d
		r.Duration = &d
		row["count"] = d.Count
	}
	raw, _ := json.Marshal(row)
	r.Groups = []json.RawMessage{raw}
	r.Count = row["count"].(string)
	return r
}
func TestViewMeasuresGeneratedScaleRationalsAndSupport(t *testing.T) {
	for scale := 1; scale <= 9; scale++ {
		for seed := 1; seed <= 20; seed++ {
			b := decimalBinding(scale)
			texts := []string{formatFixedCoefficient(big.NewInt(int64(seed)), scale), formatFixedCoefficient(big.NewInt(int64(-seed-2)), scale), formatFixedCoefficient(big.NewInt(9007199254740993), scale)}
			values := []*string{nil, viewText(""), viewText("bad")}
			sum := new(big.Rat)
			for _, s := range texts {
				values = append(values, viewText(s))
				v, ok := new(big.Rat).SetString(s)
				if !ok {
					t.Fatal(s)
				}
				sum.Add(sum, v)
			}
			r := measureGroupFixture(t, &b, values, nil)
			d := groupViewDataset(r, true)
			if d == nil || d.Completeness == nil {
				t.Fatal(scale, seed, d)
			}
			row := d.Rows[0]
			mean := new(big.Rat).Quo(sum, big.NewRat(3, 1)).RatString()
			if *row.Values["numeric.mean"] != mean || *row.Values["numeric.sum"] != sum.RatString() {
				t.Fatal(scale, seed, row.Values, mean, sum)
			}
			for _, key := range []string{"numeric.mean", "numeric.sum", "numeric.min", "numeric.max"} {
				cell := row.Cells[key]
				if cell.State != "derived" || !reflect.DeepEqual(cell.Support, &ViewSupport{Total: "6", Valid: "3", Missing: "1", Empty: "1", Invalid: "1"}) {
					t.Fatal(key, cell)
				}
				value, _ := new(big.Rat).SetString(*row.Values[key])
				if value.RatString() != *row.Values[key] {
					t.Fatal("noncanonical rational")
				}
			}
			for _, f := range d.Fields[2:] {
				if f.Type != "rational" || f.Unit != "USD" || f.Aggregation == nil || f.Aggregation.Family != "numeric" {
					t.Fatal(f)
				}
			}
		}
	}
}
func TestViewMeasuresUnavailableSumsAreNotZero(t *testing.T) {
	b := numericBinding()
	for _, values := range [][]*string{nil, {nil, viewText(""), viewText("bad")}} {
		r := measureGroupFixture(t, &b, values, nil)
		d := groupViewDataset(r, true)
		if d == nil || d.Completeness == nil {
			t.Fatal(d)
		}
		state := "no_valid_values"
		if len(values) == 0 {
			state = "no_data"
		}
		for _, key := range []string{"numeric.mean", "numeric.sum", "numeric.min", "numeric.max"} {
			if d.Rows[0].Values[key] != nil || d.Rows[0].Cells[key].State != state {
				t.Fatal(key, d.Rows[0])
			}
		}
	}
	// Exact cancellation is an observed zero, unlike absence of valid values.
	r := measureGroupFixture(t, &b, numericValues("-1", "1"), nil)
	d := groupViewDataset(r, true)
	if d == nil || *d.Rows[0].Values["numeric.sum"] != "0" || d.Rows[0].Cells["numeric.sum"].State != "derived" {
		t.Fatal(d)
	}
}
func TestViewMeasuresDurationExactQuantileMethod(t *testing.T) {
	b := numericBinding()
	r := measureGroupFixture(t, &b, numericValues("1", "2", "3"), []uint64{0, 0, 1})
	d := groupViewDataset(r, true)
	if d == nil || len(d.Fields) != 12 || *d.Rows[0].Values["duration.mean"] != "1/3" || *d.Rows[0].Values["duration.p95"] != "1" {
		t.Fatal(d)
	}
	for _, f := range d.Fields {
		if f.Key == "duration.p95" && (f.Unit != "ns" || f.Type != "rational" || !reflect.DeepEqual(f.Aggregation, &ViewAggregation{Family: "duration", Kind: "quantile", Method: "quantilesExactHigh", Quantile: "0.95"})) {
			t.Fatal(f)
		}
	}
	for _, key := range []string{"duration.mean", "duration.min", "duration.max", "duration.p50", "duration.p95", "duration.p99"} {
		if d.Rows[0].Cells[key].State != "derived" || !reflect.DeepEqual(d.Rows[0].Cells[key].Support, &ViewSupport{Total: "3", Valid: "3", Missing: "0", Empty: "0", Invalid: "0"}) {
			t.Fatal(key)
		}
	}
	// Captured selector, support and exact values do not alias mutable source models.
	before, _ := json.Marshal(d)
	r.Question.Numeric.Unit = "changed"
	r.Question.Numeric.Scale = 7
	r.Numeric.Valid = "0"
	r.Duration.Count = "0"
	r.State = "changed_between_reads"
	after, _ := json.Marshal(d)
	if string(before) != string(after) {
		t.Fatal("source mutation changed supplied snapshot")
	}
}
func TestViewMeasuresValidateSummariesBeforeProjection(t *testing.T) {
	for _, edit := range []func(*NumericSummary){func(n *NumericSummary) { n.Valid = "9" }, func(n *NumericSummary) { n.Sum = "bad" }, func(n *NumericSummary) { n.Unit = "wrong" }, func(n *NumericSummary) { n.State = "no_data" }} {
		b := numericBinding()
		r := measureGroupFixture(t, &b, numericValues("1", "2"), nil)
		var g map[string]json.RawMessage
		json.Unmarshal(r.Groups[0], &g)
		var n NumericSummary
		json.Unmarshal(g["numeric"], &n)
		edit(&n)
		g["numeric"], _ = json.Marshal(n)
		r.Groups[0], _ = json.Marshal(g)
		if groupViewDataset(r, true) != nil {
			t.Fatal("malformed numeric support admitted")
		}
	}
	r := measureGroupFixture(t, nil, nil, []uint64{1, 2})
	var g map[string]json.RawMessage
	json.Unmarshal(r.Groups[0], &g)
	d := *r.Duration
	d.P95NS = "99"
	g["duration"], _ = json.Marshal(d)
	r.Groups[0], _ = json.Marshal(g)
	if groupViewDataset(r, true) != nil {
		t.Fatal("malformed duration admitted")
	}
}
func TestViewMeasuresTimeClipNoDataUnknownAndCoherence(t *testing.T) {
	b := decimalBinding(2)
	q := basic()
	q.Numeric = &b
	q.Summary = "span-duration"
	q.FromMS = 1000
	q.ToMS = 3500
	q.BucketMS = 1000
	n, _ := EvaluateNumericValues(b, numericValues("0.01", "0.02", "0.03"))
	d := ReferenceDurationSummary([]uint64{1, 2, 3})
	raw, _ := json.Marshal(map[string]any{"from_ms": "1000", "count": "3", "numeric": n, "sum_ns": d.SumNS, "min_ns": d.MinNS, "max_ns": d.MaxNS, "quantiles_ns": []string{d.P50NS, d.P95NS, d.P99NS}})
	r := timeViewFixture(t, q, []json.RawMessage{raw})
	snapshot := timeViewDataset(r, true)
	if snapshot == nil || snapshot.Completeness == nil || len(snapshot.Fields) != 13 || *snapshot.Rows[0].Values["numeric.mean"] != "1/50" || *snapshot.Rows[0].Values["duration.mean"] != "2" || *snapshot.Rows[2].Values["toMs"] != "3500" {
		t.Fatal(snapshot)
	}
	for _, key := range []string{"numeric.sum", "duration.p95"} {
		if snapshot.Rows[2].Values[key] != nil || snapshot.Rows[2].Cells[key].State != "no_data" || snapshot.Rows[2].Cells[key].Support.Total != "0" {
			t.Fatal(key)
		}
	}
	r.State = "changed_between_reads"
	snapshot = timeViewDataset(r, true)
	if snapshot.Completeness != nil || snapshot.Quality.Extent != "unknown" || snapshot.Rows[0].Cells["duration.p95"].State != "derived" {
		t.Fatal("coverage conflated with per-read value support")
	}
	r.TimeSeries = []TimePoint{{FromMS: 3000, ToMS: 3500, State: "unknown"}}
	snapshot = timeViewDataset(r, false)
	if snapshot == nil || len(snapshot.Rows) != 1 || snapshot.Rows[0].Values["numeric.sum"] != nil || snapshot.Rows[0].Cells["numeric.sum"].State != "unknown" || snapshot.Rows[0].Cells["numeric.sum"].Support != nil {
		t.Fatal(snapshot)
	}
}
func TestSourceSuppliesMeasureDatasetsWithoutAdditionalReads(t *testing.T) {
	b := decimalBinding(2)
	q := basic()
	q.Numeric = &b
	q.Summary = "span-duration"
	q.FromMS = 1000
	q.ToMS = 3500
	q.BucketMS = 1000
	q.GroupBy = []Field{{Kind: "column", Key: "service"}}
	n, _ := EvaluateNumericValues(b, numericValues("0.01", "0.02", "0.03"))
	d := ReferenceDurationSummary([]uint64{1, 2, 3})
	reads := 0
	source := Source{Endpoint: "http://source.invalid", Client: &http.Client{Transport: countedValuesTransport(func(req *http.Request) (*http.Response, error) {
		reads++
		body, _ := io.ReadAll(req.Body)
		sql := string(body)
		row := map[string]any{"count": "3", "numeric": n, "sum_ns": d.SumNS, "min_ns": d.MinNS, "max_ns": d.MaxNS, "quantiles_ns": []string{d.P50NS, d.P95NS, d.P99NS}}
		text := ""
		switch {
		case strings.Contains(sql, " AS timestamp_ns"):
			text = "{\"timestamp_ns\":\"1\",\"duration_ns\":\"1\"}\n{\"timestamp_ns\":\"2\",\"duration_ns\":\"2\"}\n{\"timestamp_ns\":\"3\",\"duration_ns\":\"3\"}"
		case strings.Contains(sql, "SELECT DISTINCT"):
		case strings.Contains(sql, "GROUP BY from_ms"):
			row["from_ms"] = "1000"
			raw, _ := json.Marshal(row)
			text = string(raw)
		case strings.Contains(sql, "GROUP BY g0"):
			row["g0"] = "svc"
			raw, _ := json.Marshal(row)
			text = string(raw)
		default:
			raw, _ := json.Marshal(row)
			text = string(raw)
		}
		return &http.Response{StatusCode: 200, Header: http.Header{}, Body: io.NopCloser(strings.NewReader(text))}, nil
	})}}
	result, err := source.Fetch(context.Background(), q)
	if err != nil || reads != 5 || result.ViewDataset == nil || result.TimeDataset == nil || result.SummaryDataset == nil || result.State != "observed" {
		t.Fatal(result, err, reads)
	}
	if *result.ViewDataset.Rows[0].Values["numeric.mean"] != "1/50" || *result.TimeDataset.Rows[0].Values["duration.p95"] != "3" {
		t.Fatal(result)
	}
}
