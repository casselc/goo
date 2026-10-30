package exploration

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func TestViewHistogramExactGeometryCoverageAndWire(t *testing.T) {
	fixtures := map[string]*ViewDataset{}
	for _, typ := range []string{"uint64", "int64", "decimal64"} {
		q := histogramQuestion(typ)
		if typ == "decimal64" {
			b := decimalBinding(2)
			q.Numeric = &b
		}
		for _, empty := range []bool{false, true} {
			values := histogramValues()
			if empty {
				values = nil
			}
			n, err := EvaluateNumericValues(*q.Numeric, values)
			if err != nil {
				t.Fatal(err)
			}
			h, err := EvaluateNumericHistogram(*q.Numeric, values)
			if err != nil {
				t.Fatal(err)
			}
			state := "observed"
			if empty {
				state = "no_data"
			}
			r := &Result{Question: q, Count: n.Count, Numeric: &n, NumericHistogram: &h, State: state, Basis: strings.Repeat("a", 64), CapturedAt: "2026-10-30T03:04:05Z"}
			d := histogramViewDataset(r)
			want := 128
			if typ == "uint64" {
				want = 65
			}
			if d == nil || len(d.Rows) != want || d.Histogram.Support.Valid != n.Valid || d.Completeness == nil || d.Completeness.Aspects[0] != "bins" {
				t.Fatalf("invalid %s view: %+v", typ, d)
			}
			if d.Rows[0].Key != "bin:0" || d.Rows[0].Cells != nil {
				t.Fatal("bin contract")
			}
			if typ == "uint64" && *d.Rows[64].Values["to"] != "18446744073709551616" {
				t.Fatal("max+1 lost")
			}
			if typ == "decimal64" && *d.Rows[0].Values["from"] != "-2305843009213693952/25" {
				t.Fatal("decimal coefficient not scaled", *d.Rows[0].Values["from"])
			}
			name := typ
			if empty {
				name += "-empty"
			}
			fixtures[name] = d
			r.State = "changed_between_reads"
			if histogramViewDataset(r).Completeness != nil {
				t.Fatal("incoherent result attested")
			}
			r.Count = "999"
			if histogramViewDataset(r) != nil {
				t.Fatal("foreign coverage accepted")
			}
			r.Count = n.Count
			h.Bins[0].Count = "18446744073709551616"
			if histogramViewDataset(r) != nil {
				t.Fatal("bad count accepted")
			}
		}
	}
	if path := os.Getenv("GOO_HISTOGRAM_VIEW_FIXTURES"); path != "" {
		b, e := json.Marshal(fixtures)
		if e != nil {
			t.Fatal(e)
		}
		if e = os.WriteFile(path, b, 0600); e != nil {
			t.Fatal(e)
		}
	}
}

func TestViewHistogramNoValidValuesAndUnavailableRefusals(t *testing.T) {
	q := histogramQuestion("uint64")
	bad, empty := "bad", ""
	values := []*string{nil, &bad, &empty}
	n, e := EvaluateNumericValues(*q.Numeric, values)
	if e != nil {
		t.Fatal(e)
	}
	h, e := EvaluateNumericHistogram(*q.Numeric, values)
	if e != nil {
		t.Fatal(e)
	}
	r := &Result{Question: q, Count: n.Count, Numeric: &n, NumericHistogram: &h, State: "observed", Basis: strings.Repeat("a", 64), CapturedAt: "2026-10-30T03:04:05Z"}
	d := histogramViewDataset(r)
	if d == nil || d.Histogram.Support.Valid != "0" || d.Histogram.Support.Total != "3" || d.Completeness == nil {
		t.Fatal("no-valid population lost", d)
	}
	for _, row := range d.Rows {
		if row.State != "no_data" || *row.Values["count"] != "0" {
			t.Fatal("invalid values entered bins")
		}
	}
	r.State = "unknown"
	if histogramViewDataset(r) != nil {
		t.Fatal("unknown source fabricated bins")
	}
	r.State = "observed"
	r.NumericHistogram = nil
	if histogramViewDataset(r) != nil {
		t.Fatal("missing histogram accepted")
	}
	r.NumericHistogram = &h
	r.Question.NumericHistogram = nil
	if histogramViewDataset(r) != nil {
		t.Fatal("unrequested histogram accepted")
	}
}
