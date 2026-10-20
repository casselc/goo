package exploration

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestLiveSpanMeasuresUseOneAggregateAndFreezeFullPopulation(t *testing.T) {
	var aggregates atomic.Int64
	sourceServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		sql := string(raw)
		switch {
		case strings.Contains(sql, "count()"):
			if !strings.Contains(sql, "quantilesExactHigh") || !strings.Contains(sql, "status_error") {
				t.Error("split/missing aggregate measures", sql)
			}
			aggregates.Add(1)
			fmt.Fprintln(w, `{"count":"1200","sum_ns":"1200000000","min_ns":"1000000","max_ns":"1000000","quantiles_ns":["1000000","1000000","1000000"],"status_error":"200","status_ok":"800","status_unset":"200","status_unknown":"0"}`)
		case strings.Contains(sql, "SELECT DISTINCT"):
		default:
			for i := 0; i < 1001; i++ {
				status := "Ok"
				if i >= 800 {
					status = "Unset"
				}
				if i == 1000 {
					status = "Error"
				}
				fmt.Fprintf(w, "{\"timestamp_ns\":\"%d\",\"duration_ns\":\"1000000\",\"StatusCode\":\"%s\"}\n", i+1, status)
			}
		}
	}))
	defer sourceServer.Close()
	source := Source{Endpoint: sourceServer.URL, Client: sourceServer.Client()}
	store := CaptureStore{Dir: t.TempDir()}
	q := Query{Signal: "spans", Environment: "prod", FromMS: 1000, ToMS: 61000, Limit: 100, Summary: "span-duration", SpanStatus: true}
	baseline, e := store.Create(context.Background(), source, "Full span baseline", CaptureQuestion{Kind: "explore", Explore: &q}, "")
	if e != nil {
		t.Fatal(e)
	}
	snapshot, pair, e := source.LiveRecordComparison(context.Background(), baseline, time.UnixMilli(1800000000000))
	if e != nil {
		t.Fatal(e)
	}
	if aggregates.Load() != 2 || len(snapshot.Explore.Rows) != 1000 || pair.Selection.Count != "1200" || pair.Measures.SelectionDuration.Count != "1200" || pair.Measures.SelectionStatus.Error != "200" || pair.Measures.ErrorDifferencePP != "0.00" {
		t.Fatal("row bank became measure population", pair, aggregates.Load())
	}
	sourceServer.Close()
	frozen, e := store.FreezeSnapshot(snapshot)
	if e != nil {
		t.Fatal(e)
	}
	loaded, e := store.Load(frozen.ID)
	if e != nil || !reflect.DeepEqual(loaded.Explore.Duration, snapshot.Explore.Duration) || !reflect.DeepEqual(loaded.Explore.SpanStatus, snapshot.Explore.SpanStatus) || loaded.Explore.Basis != snapshot.Explore.Basis || aggregates.Load() != 2 {
		t.Fatal("frozen measures lost their basis", e)
	}
}
