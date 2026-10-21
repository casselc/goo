package measurements

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestHistogramTimeDefinitionAndReplay(t *testing.T) {
	d := definition(t, "checkout-duration")
	d.Kind = "histogram"
	d.Source.Unit = "s"
	width := int64(2000)
	d.HistogramBucketMS = &width
	q := Question{Definition: d, FromMS: 100000, ToMS: 104000}
	if e := q.Validate(); e != nil {
		t.Fatal(e)
	}
	row := gaugeRow(d, "one", 101000, "")
	row.StartNS = ns(100000)
	row.Temporality = 1
	row.Count = "2"
	row.Bounds = []float64{1}
	row.BucketCounts = []string{"1", "1"}
	next := row
	next.StartNS = ns(101000)
	next.EndNS = ns(102000)
	next.Count = "3"
	next.BucketCounts = []string{"0", "3"}
	r, e := Evaluate(q, []json.RawMessage{raw(row), raw(next)}, time.Unix(200, 0), "", nil)
	if e != nil {
		t.Fatal(e)
	}
	if r.HistogramTime == nil || r.HistogramTime.Series[0].Buckets[0].Distribution.Count != "5" || r.HistogramTime.Series[0].Buckets[1].State != "unavailable" {
		t.Fatalf("bad buckets: %+v", r.HistogramTime)
	}
	if e = ValidateResult(r); e != nil {
		t.Fatal(e)
	}
	encoded, _ := json.Marshal(r)
	var restored Result
	json.Unmarshal(encoded, &restored)
	if !reflect.DeepEqual(r, &restored) || ValidateResult(&restored) != nil {
		t.Fatal("bucketed result did not replay")
	}
	restored.HistogramTime.Series[0].Buckets[0].Distribution.Count = "6"
	if ValidateResult(&restored) == nil {
		t.Fatal("accepted corrupt stored time bucket")
	}
	noWidth := d
	noWidth.HistogramBucketMS = nil
	b, _ := json.Marshal(noWidth)
	if strings.Contains(string(b), "histogramBucketMs") || Hash(noWidth) == Hash(d) {
		t.Fatal("absent width changed legacy shape or width not hashed")
	}
	for _, kind := range []string{"gauge", "counter", "lifecycle"} {
		bad := d
		bad.Kind = kind
		if bad.Validate() == nil {
			t.Fatal("width accepted on", kind)
		}
	}
	for _, n := range []int64{0, -1, 7200001} {
		bad := d
		bad.HistogramBucketMS = &n
		if bad.Validate() == nil {
			t.Fatal("accepted width", n)
		}
	}
	narrow := int64(1)
	q.Definition.HistogramBucketMS = &narrow
	if q.Validate() == nil {
		t.Fatal("accepted >200 bins")
	}
	// Exact nanosecond question origin is preserved, not rounded to milliseconds.
	q.Definition = d
	q.FromNS = "100000000123"
	q.ToNS = "104000000123"
	q.ToMS = 104001
	if q.Validate() != nil {
		t.Fatal("exact origin rejected")
	}
}
