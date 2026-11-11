package exploration

import (
	"encoding/json"
	"testing"
)

func TestCapturedRecordRateProof(t *testing.T) {
	q := basic()
	q.FromMS, q.ToMS, q.BucketMS = 1000, 3501, 1000
	r := timeViewFixture(t, q, []json.RawMessage{json.RawMessage(`{"from_ms":"3000","count":"2"}`)})
	d := timeViewDataset(r, true)
	if d == nil || ValidateRecordRateDataset(d) != nil {
		t.Fatal("source intervals lack valid rate proof")
	}
	if got := *d.Rows[0].Values[RecordRateKey]; got != "0" {
		t.Fatal("known zero rate", got)
	}
	last := &d.Rows[2]
	if got := *last.Values[RecordRateKey]; got != "2000/501" || last.Cells[RecordRateKey].Rate.ElapsedMS != "501" {
		t.Fatal("clipped interval exposure", got)
	}
	for _, mutate := range []func(*ViewDataset){
		func(d *ViewDataset) { d.Rows[2].Cells[RecordRateKey].Rate.ElapsedMS = "1000" },
		func(d *ViewDataset) { d.Rows[2].Cells[RecordRateKey].Rate.Count = "3" },
		func(d *ViewDataset) { d.Rows[2].Values[RecordRateKey] = viewText("4000/1002") },
		func(d *ViewDataset) { d.Quality.Freshness = "live" },
		func(d *ViewDataset) {
			d.Scope.Selector = json.RawMessage(`{"signal":"logs","fromMs":"1000","toMs":3501}`)
		},
	} {
		raw, _ := json.Marshal(d)
		var bad ViewDataset
		if err := json.Unmarshal(raw, &bad); err != nil {
			t.Fatal(err)
		}
		mutate(&bad)
		if ValidateRecordRateDataset(&bad) == nil {
			t.Fatal("forged rate accepted")
		}
	}
}

func TestCapturedRecordRateExactAndUnknown(t *testing.T) {
	value, cell, err := RecordRateCell("18446744073709551615", 501, false)
	if err != nil || value == nil || *value != "6148914691236517205000/167" || cell.Rate.Count != "18446744073709551615" {
		t.Fatal(value, cell, err)
	}
	value, cell, err = RecordRateCell("", 501, true)
	if err != nil || value != nil || cell.State != "unknown" || cell.Rate != nil || cell.Support != nil {
		t.Fatal(value, cell, err)
	}
	for _, count := range []string{"", "01", "-1", "18446744073709551616"} {
		if _, _, err := RecordRateCell(count, 501, false); err == nil {
			t.Fatal("bad count", count)
		}
	}
	q := basic()
	q.GroupBy = []Field{{Kind: "resource", Key: "service.name"}}
	if RecordRateQuery(q) {
		t.Fatal("grouped rate admitted")
	}
}
