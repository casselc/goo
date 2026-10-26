package verification

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/measurements"
	"github.com/casselc/goo/metrics"
	"math/big"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
	"testing"
)

type pvPair struct {
	Request  measurements.PoolViewRequest
	Response measurements.PoolView
}
type pvLive struct {
	Facts            json.RawMessage
	Primary, Partial pvPair
	Boundary         []struct {
		ComposeRequest  measurements.PoolViewRequest
		ComposeResponse measurements.PoolView
	}
	Individual                measurements.Result
	Saved, Reopened           exploration.CapturePage
	StoredBefore, StoredAfter struct {
		File, SHA256 string
		Bytes        int
	}
	Events []struct {
		Action                   string
		Source, PoolPlans, Views int
	}
}

func pvRaw(f psMap, r *measurements.Result) bool {
	if r == nil {
		return false
	}
	b, _ := json.Marshal(r)
	var m psMap
	_ = json.Unmarshal(b, &m)
	if r.Question.Definition.Kind == "histogram" {
		return psMeasurement(f, m)
	}
	if r.Stocks == nil || r.Question.Definition.StockDomain != "nonnegative-integer" {
		return false
	}
	q := psObj(m["question"])
	pts := psSelected(f, q, false)
	if len(pts) != len(r.Rows) || r.Count != strconv.Itoa(len(pts)) || len(r.Stocks.Series) != 1 {
		return false
	}
	sort.Slice(pts, func(i, j int) bool { return pts[i]["endNS"].(string) < pts[j]["endNS"].(string) })
	samples := r.Stocks.Series[0].Samples
	if len(samples) != len(pts) {
		return false
	}
	stream := r.Question.Definition.Source.ExactStream
	if stream == nil || !reflect.DeepEqual(r.Stocks.Series[0].Stream, *stream) {
		return false
	}
	_, to := r.Question.BoundsNS()
	hi, _ := strconv.ParseUint(to, 10, 64)
	for n, p := range pts {
		var row psMap
		_ = json.Unmarshal(r.Rows[n], &row)
		if row["start_ns"] != p["startNS"] || row["end_ns"] != p["endNS"] || row["value"] != p["value"] || !psEq(row["resource"], stream.Resource) || !psEq(row["attributes"], stream.Attributes) {
			return false
		}
		v, _ := strconv.ParseFloat(p["value"].(string), 64)
		s := samples[n]
		at, _ := strconv.ParseUint(p["endNS"].(string), 10, 64)
		state := "observed"
		if hi-at > uint64(r.Question.Definition.MaxGapMS)*1000000 {
			state = "stale"
		}
		if s.AtNS != p["endNS"] || s.StartNS != p["startNS"] || s.Value == nil || *s.Value != v || s.State != state {
			return false
		}
	}
	return true
}
func pvShared(s metrics.Stream) string {
	return siKey([]any{s.Resource, s.ResourceSchemaURL, s.Scope, s.Version, s.SchemaURL, s.Attributes})
}
func pvPairOK(raw json.RawMessage, p pvPair) bool {
	var facts piFacts
	_ = json.Unmarshal(raw, &facts)
	var f psMap
	_ = json.Unmarshal(raw, &f)
	q, v := p.Request, p.Response
	if !piResult(facts, &q.Count, true) || v.Version != "goo/pool-view/v1" || !reflect.DeepEqual(q.Stream, v.Stream) || !reflect.DeepEqual(q.Count.Question, v.CountQuestion) || v.CountEvidence.Basis != q.Count.Basis || v.CountEvidence.ReadAt != q.Count.ReadAt {
		return false
	}
	key := piIdentity(q.Stream)
	states := []measurements.PoolViewCountState{}
	var used *metrics.StockSample
	for _, s := range q.Count.Stocks.Series {
		if piIdentity(s.Stream) != key {
			continue
		}
		name := s.Stream.Attributes["db.client.connection.state"]
		if name == "" {
			continue
		}
		var latest *metrics.StockSample
		if len(s.Samples) > 0 {
			x := s.Samples[len(s.Samples)-1]
			latest = &x
		}
		states = append(states, measurements.PoolViewCountState{Name: name, Stream: s.Stream, Latest: latest})
		if name == "used" {
			used = latest
		}
	}
	sort.Slice(states, func(i, j int) bool { return states[i].Name < states[j].Name })
	if !reflect.DeepEqual(states, v.States) {
		return false
	}
	state, reason := "unavailable", "capacity-not-supplied"
	var capacity *metrics.StockSample
	for _, pair := range []struct {
		r    *measurements.Result
		kind string
	}{{q.Capacity, "capacity"}, {q.Wait, "wait"}} {
		r := pair.r
		if r == nil {
			continue
		}
		if !pvRaw(f, r) {
			return false
		}
		a, b := q.Count.Question.BoundsNS()
		c, d := r.Question.BoundsNS()
		if a != c || b != d {
			return false
		}
		s := r.Question.Definition.Source.ExactStream
		if s == nil || pvShared(*s) != pvShared(func() metrics.Stream {
			x := q.Stream
			x.Attributes = map[string]string{}
			for k, v := range q.Stream.Attributes {
				if k != "db.client.connection.state" {
					x.Attributes[k] = v
				}
			}
			return x
		}()) {
			return false
		}
		if pair.kind == "capacity" {
			if v.Capacity == nil || v.CapacityQuestion == nil || !reflect.DeepEqual(*v.CapacityQuestion, r.Question) || v.Capacity.Evidence.Basis != r.Basis || v.Capacity.Evidence.ReadAt != r.ReadAt {
				return false
			}
			samples := r.Stocks.Series[0].Samples
			if len(samples) > 0 {
				x := samples[len(samples)-1]
				capacity = &x
			}
			if !reflect.DeepEqual(v.Capacity.Latest, capacity) {
				return false
			}
		} else {
			if v.Wait == nil || v.WaitQuestion == nil || !reflect.DeepEqual(*v.WaitQuestion, r.Question) || v.Wait.Evidence.Basis != r.Basis || v.Wait.Evidence.ReadAt != r.ReadAt {
				return false
			}
			series := r.Histograms.Series
			if len(series) != 1 || len(series[0].Intervals) == 0 {
				return false
			}
			latest := series[0].Intervals[len(series[0].Intervals)-1]
			if !reflect.DeepEqual(v.Wait.Latest, &latest) {
				return false
			}
		}
	}
	if q.Capacity != nil {
		switch {
		case used == nil:
			reason = "used-not-reported"
		case used.State != "observed" || used.Value == nil:
			reason = "used-unavailable"
		case capacity == nil:
			reason = "capacity-no-data"
		case capacity.State != "observed" || capacity.Value == nil:
			reason = "capacity-unavailable"
		case used.AtNS != capacity.AtNS:
			reason = "observation-time-mismatch"
		case used.StartNS == "0" || capacity.StartNS == "0":
			reason = "epoch-not-reported"
		case used.StartNS != capacity.StartNS:
			reason = "epoch-mismatch"
		case *capacity.Value == 0:
			reason = "zero-capacity"
		case *used.Value > *capacity.Value:
			state, reason = "inconsistent", "used-exceeds-capacity"
		default:
			state, reason = "observed", "matching-reported-observation-and-epoch"
		}
	}
	if v.Utilization.State != state || v.Utilization.Reason != reason {
		return false
	}
	if state == "observed" {
		r := new(big.Rat).SetFrac(big.NewInt(int64(*used.Value)), big.NewInt(int64(*capacity.Value)))
		return v.Utilization.Numerator == r.Num().String() && v.Utilization.Denominator == r.Denom().String()
	}
	return v.Utilization.Numerator == "" && v.Utilization.Denominator == ""
}
func pvLiveOK(d pvLive) bool {
	if !pvPairOK(d.Facts, d.Primary) || !pvPairOK(d.Facts, d.Partial) || !reflect.DeepEqual(d.Saved, d.Reopened) {
		return false
	}
	for _, b := range d.Boundary {
		if !pvPairOK(d.Facts, pvPair{Request: b.ComposeRequest, Response: b.ComposeResponse}) {
			return false
		}
	}
	src := []int{9, 9, 9, 11, 11, 12, 12, 12, 13, 14, 14, 14}
	plans := []int{0, 1, 1, 1, 1, 2, 2, 2, 2, 2, 2, 2}
	views := []int{0, 0, 0, 1, 1, 2, 2, 2, 2, 2, 2, 2}
	if len(d.Events) != len(src) {
		return false
	}
	for n, e := range d.Events {
		if e.Source != src[n] || e.PoolPlans != plans[n] || e.Views != views[n] {
			return false
		}
	}
	return true
}
func TestPoolViewObserved(t *testing.T) {
	path := os.Getenv("GOO_POOL_VIEW_LEDGER")
	if path == "" {
		t.Skip("set GOO_POOL_VIEW_LEDGER")
	}
	b, e := os.ReadFile(path)
	if e != nil {
		t.Fatal(e)
	}
	var d pvLive
	if json.Unmarshal(b, &d) != nil || !pvLiveOK(d) {
		t.Fatal("raw input/composition/provenance/effect oracle")
	}
	a, e := os.ReadFile(filepath.Join(filepath.Dir(path), d.StoredBefore.File))
	if e != nil {
		t.Fatal(e)
	}
	z, e := os.ReadFile(filepath.Join(filepath.Dir(path), d.StoredAfter.File))
	h := sha256.Sum256(a)
	var stored exploration.SavedCapture
	if e != nil || string(a) != string(z) || len(a) != d.StoredBefore.Bytes || hex.EncodeToString(h[:]) != d.StoredBefore.SHA256 || d.StoredBefore.SHA256 != d.StoredAfter.SHA256 || json.Unmarshal(a, &stored) != nil || stored.ID != d.Saved.Receipt.ID || !reflect.DeepEqual(stored.CaptureReceipt, d.Saved.Receipt) || !peLiveCaptureHash(exploration.CapturePage{Receipt: stored.CaptureReceipt, Measurement: stored.Measurement}) {
		t.Fatal("complete individual capture custody")
	}
	var facts psMap
	_ = json.Unmarshal(d.Facts, &facts)
	if !pvRaw(facts, stored.Measurement) || !reflect.DeepEqual(stored.Measurement.Question, d.Individual.Question) {
		t.Fatal("frozen original individual measurement")
	}
	controls := []func(*pvLive){func(x *pvLive) { x.Primary.Response.Utilization.Numerator = "4" }, func(x *pvLive) { x.Primary.Response.States = x.Primary.Response.States[:1] }, func(x *pvLive) { x.Primary.Response.Capacity.Evidence.ReadAt = "changed" }, func(x *pvLive) { x.Primary.Response.CountEvidence.Basis = "changed" }, func(x *pvLive) { x.Events[1].Source++ }, func(x *pvLive) { x.Boundary[0].ComposeResponse.Utilization.State = "observed" }, func(x *pvLive) { x.Primary.Request.Capacity.Rows = nil }, func(x *pvLive) { x.Primary.Response.States[0].Stream.Resource["service.instance.id"] = "other" }}
	for n, change := range controls {
		var x pvLive
		_ = json.Unmarshal(b, &x)
		change(&x)
		if reflect.DeepEqual(x, d) || pvLiveOK(x) {
			t.Fatalf("changed control %d accepted", n)
		}
	}
}
func TestPoolViewOwnershipObserved(t *testing.T) {
	path := os.Getenv("GOO_POOL_VIEW_OWNER_LEDGER")
	if path == "" {
		t.Skip("set GOO_POOL_VIEW_OWNER_LEDGER")
	}
	raw, e := os.ReadFile(path)
	if e != nil {
		t.Fatal(e)
	}
	var d psMap
	_ = json.Unmarshal(raw, &d)
	check := func(x psMap) bool {
		cases := psArr(x["cases"])
		if len(cases) != 3 {
			return false
		}
		for _, v := range cases {
			c := psObj(v)
			a, z := psObj(c["beforeRelease"]), psObj(c["afterRelease"])
			if !psEq(a, z) || a["views"] != float64(0) || a["viewCount"] != float64(0) || a["source"] != psObj(c["beforeLoad"])["source"].(float64)+1 || psObj(c["held"])["status"] != float64(200) {
				return false
			}
		}
		return true
	}
	if !check(d) {
		t.Fatal("held real related-source result published after owner invalidation")
	}
	for _, key := range []string{"source", "views", "viewCount", "url"} {
		var x psMap
		_ = json.Unmarshal(raw, &x)
		psObj(psObj(psArr(x["cases"])[0])["afterRelease"])[key] = "changed"
		if check(x) {
			t.Fatal("owner control", key)
		}
	}
}
