package verification

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"net/http/httptest"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/casselc/goo/exploration"
	"hegel.dev/go/hegel"
)

type raFact struct {
	Value    *string
	Duration uint64
	Status   string
	TimeMS   int64
}

func raSummary(facts []raFact) (map[string]any, exploration.DurationSummary, exploration.SpanStatusSummary) {
	n := len(facts)
	count := strconv.Itoa(n)
	sum := new(big.Int)
	values := []uint64{}
	classes := map[string]int{"Error": 0, "Ok": 0, "Unset": 0, "unknown": 0}
	for _, f := range facts {
		sum.Add(sum, new(big.Int).SetUint64(f.Duration))
		values = append(values, f.Duration)
		key := f.Status
		if _, ok := classes[key]; !ok {
			key = "unknown"
		}
		classes[key]++
	}
	sort.Slice(values, func(i, j int) bool { return values[i] < values[j] })
	min, max := "0", "0"
	quantiles := []string{"0", "0", "0"}
	d := exploration.DurationSummary{State: "no_data", Count: count, SumNS: sum.String()}
	if n > 0 {
		min = strconv.FormatUint(values[0], 10)
		max = strconv.FormatUint(values[n-1], 10)
		for i, p := range []int{50, 95, 99} {
			rank := new(big.Int).Div(new(big.Int).Mul(big.NewInt(int64(n)), big.NewInt(int64(p))), big.NewInt(100)).Int64()
			if rank >= int64(n) {
				rank = int64(n) - 1
			}
			quantiles[i] = strconv.FormatUint(values[rank], 10)
		}
		d.State = "observed"
		d.MinNS = min
		d.MaxNS = max
		d.P50NS = quantiles[0]
		d.P95NS = quantiles[1]
		d.P99NS = quantiles[2]
	}
	status := exploration.SpanStatusSummary{Count: count, Error: strconv.Itoa(classes["Error"]), OK: strconv.Itoa(classes["Ok"]), Unset: strconv.Itoa(classes["Unset"]), Unknown: strconv.Itoa(classes["unknown"])}
	raw := map[string]any{"count": count, "sum_ns": sum.String(), "min_ns": min, "max_ns": max, "quantiles_ns": quantiles, "status_error": status.Error, "status_ok": status.OK, "status_unset": status.Unset, "status_unknown": status.Unknown}
	return raw, d, status
}
func raJSON(v any) string { b, _ := json.Marshal(v); return string(b) }
func raFixture(facts []raFact) (string, []string, []string) {
	raw, _, _ := raSummary(facts)
	groups := map[string][]raFact{}
	labels := map[string]*string{}
	for _, f := range facts {
		key := "missing"
		if f.Value != nil {
			key = "present:" + *f.Value
		}
		groups[key] = append(groups[key], f)
		labels[key] = f.Value
	}
	keys := []string{}
	for k := range groups {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	groupRows := []string{}
	for _, k := range keys {
		row, _, _ := raSummary(groups[k])
		row["g0"] = labels[k]
		groupRows = append(groupRows, raJSON(row))
	}
	series := []string{}
	for start := int64(1000); start < 4000; start += 1000 {
		in := []raFact{}
		for _, f := range facts {
			if f.TimeMS >= start && f.TimeMS < start+1000 {
				in = append(in, f)
			}
		}
		if len(in) > 0 {
			row, _, _ := raSummary(in)
			row["from_ms"] = strconv.FormatInt(start, 10)
			series = append(series, raJSON(row))
		}
	}
	return raJSON(raw), groupRows, series
}
func raSource(rows [][]string, newProjection bool) (exploration.Source, *atomic.Int32) {
	calls := &atomic.Int32{}
	client := &http.Client{Transport: cvTransport(func(r *http.Request) (*http.Response, error) {
		i := int(calls.Add(1)) - 1
		body, _ := io.ReadAll(r.Body)
		w := httptest.NewRecorder()
		if i >= len(rows) || r.Method != "POST" || r.URL.Query().Get("readonly") != "1" {
			http.Error(w, "unexpected read", 500)
			return w.Result(), nil
		}
		if newProjection && (strings.Contains(string(body), " AS timestamp_ns") || strings.Contains(string(body), "mapKeys(")) {
			http.Error(w, "hidden row or field read", 500)
			return w.Result(), nil
		}
		for _, row := range rows[i] {
			fmt.Fprintln(w, row)
		}
		return w.Result(), nil
	})}
	return exploration.Source{Endpoint: "http://aggregate.invalid", Client: client}, calls
}
func TestHegelRecordAggregatesLegacyProjectionParity(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		seed := hegel.Draw(tc, hegel.Integers(0, 30))
		facts := []raFact{}
		for i := 0; i < 120+seed; i++ {
			var v *string
			switch i % 5 {
			case 0:
				v = nil
			case 1:
				v = lpcPointer("")
			default:
				v = lpcPointer("/orders")
			}
			duration := uint64(100 + i + seed)
			if i%19 == 0 {
				duration = 9007199254740993 + uint64(i)
			}
			facts = append(facts, raFact{v, duration, []string{"Error", "Ok", "Unset", "error"}[i%4], 1000 + int64(i*10)})
		}
		countRow, groups, series := raFixture(facts)
		_, expectedDuration, expectedStatus := raSummary(facts)
		for _, bucket := range []int64{0, 1000} {
			q := exploration.Query{Signal: "spans", Environment: "ordinary", FromMS: 1000, ToMS: 4000, Limit: 5, Summary: "span-duration", SpanStatus: true, BucketMS: bucket, GroupBy: []exploration.Field{{Kind: "attribute", Key: "route"}}}
			before, _ := json.Marshal(q)
			newRows := [][]string{{countRow}, groups}
			if bucket != 0 {
				newRows = append(newRows, series)
			}
			newSource, newCalls := raSource(newRows, true)
			aggregate, e := newSource.FetchRecordAggregates(context.Background(), q)
			if e != nil {
				tc.Errorf("aggregate source %v", e)
				return
			}
			inspection := []string{}
			for _, f := range facts[:6] {
				inspection = append(inspection, raJSON(map[string]string{"timestamp_ns": strconv.FormatInt(f.TimeMS*1000000, 10), "duration_ns": strconv.FormatUint(f.Duration, 10)}))
			}
			oldRows := [][]string{{countRow}, inspection, groups, {}}
			if bucket != 0 {
				oldRows = append(oldRows, series)
			}
			legacySource, oldCalls := raSource(oldRows, false)
			legacy, e := legacySource.Fetch(context.Background(), q)
			if e != nil {
				tc.Errorf("legacy source %v", e)
				return
			}
			after, _ := json.Marshal(q)
			if string(before) != string(after) || newCalls.Load() != int32(len(newRows)) || oldCalls.Load() != int32(len(oldRows)) || aggregate.Version != "goo/record-aggregates/v1" || aggregate.State != "observed" || !nhLiveSameQuery(aggregate.Question, q) {
				tc.Errorf("projection/read contract")
				return
			}
			if aggregate.Count != strconv.Itoa(len(facts)) || !reflect.DeepEqual(aggregate.Duration, &expectedDuration) || !reflect.DeepEqual(aggregate.SpanStatus, &expectedStatus) || aggregate.Count != legacy.Count || !reflect.DeepEqual(aggregate.Groups, legacy.Groups) || !reflect.DeepEqual(aggregate.TimeSeries, legacy.TimeSeries) || !reflect.DeepEqual(aggregate.Duration, legacy.Duration) || !reflect.DeepEqual(aggregate.SpanStatus, legacy.SpanStatus) {
				tc.Errorf("independent totals/legacy projection parity")
				return
			}
			if !cvLiveReceipt(aggregate.Reads.Count) || !cvLiveReceipt(aggregate.Reads.Groups) || (bucket != 0) != (aggregate.Reads.TimeSeries != nil) {
				tc.Errorf("read receipts")
				return
			}
			var wire map[string]json.RawMessage
			json.Unmarshal([]byte(raJSON(aggregate)), &wire)
			for _, forbidden := range []string{"rows", "fields", "plan", "resultToken"} {
				if _, ok := wire[forbidden]; ok {
					tc.Errorf("inspection leaked into aggregate DTO")
					return
				}
			}
		}
	}, options(18)...)
}
func TestRecordAggregatesChangedDurationRetainsDiagnostics(t *testing.T) {
	facts := []raFact{{lpcPointer("a"), 100, "Ok", 1000}, {lpcPointer("b"), 200, "Error", 2000}}
	count, groups, series := raFixture(facts)
	q := exploration.Query{Signal: "spans", Environment: "ordinary", FromMS: 1000, ToMS: 4000, Limit: 100, Summary: "span-duration", SpanStatus: true, BucketMS: 1000, GroupBy: []exploration.Field{{Kind: "attribute", Key: "route"}}}
	for _, which := range []string{"group-duration", "series-duration"} {
		g := append([]string(nil), groups...)
		s := append([]string(nil), series...)
		var row map[string]any
		target := &g[0]
		if which == "series-duration" {
			target = &s[0]
		}
		json.Unmarshal([]byte(*target), &row)
		// A separately read population can retain its count/status while durations differ.
		row["sum_ns"] = "101"
		row["min_ns"] = "101"
		row["max_ns"] = "101"
		row["quantiles_ns"] = []string{"101", "101", "101"}
		*target = raJSON(row)
		source, _ := raSource([][]string{{count}, g, s}, true)
		r, e := source.FetchRecordAggregates(context.Background(), q)
		if e != nil || r.State != "changed_between_reads" || r.Count != "2" || len(r.Groups) != 2 || len(r.TimeSeries) != 3 {
			t.Fatal(which, "changed duration diagnostics lost", e)
		}
	}
}
