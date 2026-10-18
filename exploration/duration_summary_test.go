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

func TestDurationSummaryExactAndInvalid(t *testing.T) {
	d := ReferenceDurationSummary([]uint64{0, 1, 2, 3})
	if d.P50NS != "2" || d.SumNS != "6" || d.P95NS != "3" || d.Validate() != nil {
		t.Fatal(d)
	}
	d = ReferenceDurationSummary([]uint64{18446744073709551615, 18446744073709551615})
	if d.SumNS != "36893488147419103230" || d.Validate() != nil {
		t.Fatal(d)
	}
	if ReferenceDurationSummary(nil).Validate() != nil {
		t.Fatal("empty")
	}
	for _, raw := range []string{`{"count":"0","sum_ns":"0","min_ns":"1","max_ns":"0","quantiles_ns":["0","0","0"]}`, `{"count":"1","sum_ns":"9","min_ns":"1","max_ns":"2","quantiles_ns":["1","1","2"]}`, `{"count":"1","sum_ns":"1","min_ns":"1","max_ns":"1","quantiles_ns":["1","1"]}`, `{"count":"01","sum_ns":"1","min_ns":"1","max_ns":"1","quantiles_ns":["1","1","1"]}`} {
		if _, e := DecodeDurationSummary(json.RawMessage(raw)); e == nil {
			t.Fatal(raw)
		}
	}
}
func TestSummaryPlanAndTimeSeries(t *testing.T) {
	q := Query{Signal: "spans", Environment: "prod", FromMS: 1001, ToMS: 4501, Limit: 1, Summary: "span-duration", BucketMS: 1000}
	p, e := Compile(q)
	if e != nil || !strings.Contains(p.Count, "sum(toUInt128(Duration))") || !strings.Contains(p.TimeSeries, "quantilesExactHigh") || strings.Contains(p.Count, "LIMIT") {
		t.Fatal(p, e)
	}
	raw := []json.RawMessage{json.RawMessage(`{"from_ms":"2001","count":"2","sum_ns":"3","min_ns":"1","max_ns":"2","quantiles_ns":["2","2","2"]}`)}
	series, e := DecodeTimeSeries(q, raw)
	if e != nil || len(series) != 4 || series[0].State != "no_data" || series[1].Count != "2" || series[3].ToMS != 4501 {
		t.Fatal(series, e)
	}
	if _, e = DecodeTimeSeries(q, append(raw, raw[0])); e == nil {
		t.Fatal("duplicate")
	}
	q.Signal = "logs"
	if q.Validate() == nil {
		t.Fatal("log duration")
	}
	q.Summary = ""
	q.ToMS = 500001
	if q.Validate() == nil {
		t.Fatal("budget")
	}
}
func TestSummarySourcePersistenceAndFailure(t *testing.T) {
	q := Query{Signal: "spans", Environment: "prod", FromMS: 1000, ToMS: 3000, Limit: 1, Summary: "span-duration", BucketMS: 1000}
	bad, dropSeries := false, false
	source := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		sql := string(b)
		switch {
		case strings.Contains(sql, "GROUP BY from_ms"):
			if dropSeries {
				return
			}
			fmt.Fprintln(w, `{"from_ms":"1000","count":"2","sum_ns":"3","min_ns":"1","max_ns":"2","quantiles_ns":["2","2","2"]}`)
		case strings.Contains(sql, "SELECT toString(count())"):
			if bad {
				fmt.Fprintln(w, `{"count":"2"}`)
			} else {
				fmt.Fprintln(w, `{"count":"2","sum_ns":"3","min_ns":"1","max_ns":"2","quantiles_ns":["2","2","2"]}`)
			}
		case strings.Contains(sql, " AS timestamp_ns"):
			fmt.Fprintln(w, `{"timestamp_ns":"1000000000","duration_ns":"1"}`)
		}
	}))
	defer source.Close()
	store := CaptureStore{Dir: t.TempDir()}
	c, e := store.Create(context.Background(), Source{Endpoint: source.URL}, "Durations", CaptureQuestion{Kind: "explore", Explore: &q}, "")
	if e != nil {
		t.Fatal(e)
	}
	loaded, e := store.Load(c.ID)
	if e != nil || !reflect.DeepEqual(loaded.Explore.Duration, c.Explore.Duration) || len(loaded.Explore.TimeSeries) != 2 {
		t.Fatal(loaded, e)
	}
	dropSeries = true
	changed, e := store.Create(context.Background(), Source{Endpoint: source.URL}, "Changed between reads", c.Question, "")
	if e != nil || changed.Explore.State != "changed_between_reads" || changed.Explore.Duration.Count != "2" || changed.Explore.TimeSeries[0].Count != "0" {
		t.Fatal(changed, e)
	}
	loaded.Explore.TimeSeries[0].ToMS++
	if _, e = store.publish(*loaded); e == nil {
		t.Fatal("corrupt interval published")
	}
	bad = true
	if _, e = store.Create(context.Background(), Source{Endpoint: source.URL}, "Bad", c.Question, c.ID); e == nil {
		t.Fatal("missing summary became valid")
	}
	p, e := DraftTelemetryPackage(q, "durations", "Duration question", 1)
	if e != nil || p.Question(1000, 3000).Summary != q.Summary || p.Question(1000, 3000).BucketMS != q.BucketMS {
		t.Fatal(p, e)
	}
}
