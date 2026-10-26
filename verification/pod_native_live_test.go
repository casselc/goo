package verification

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/measurements"
	"math"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
	"testing"
)

type pnRead struct {
	DiscoveryRequest measurements.PodMetricDiscoveryQuestion
	Discovery        measurements.PodMetricDiscovery
	PlanRequest      measurements.PodDashboardRequest
	Plan             measurements.PodDashboardPlan
}
type pnSnapshot struct {
	Widgets []struct {
		Key, Value, Title string
		Timeline          struct {
			Points []struct {
				State string
				Value *string
			}
		}
	}
}
type pnLive struct {
	Facts   psMap
	Initial struct {
		Request  exploration.PodDiscoveryQuestion
		Response exploration.PodDiscovery
	}
	Primary                   exploration.PodResourceObservation
	PrimaryRead               pnRead
	PrimaryResults            []measurements.Result
	Retained                  pnSnapshot
	Individual                measurements.Result
	Saved, Reopened           exploration.CapturePage
	StoredBefore, StoredAfter struct {
		File, SHA256 string
		Bytes        int
	}
	Boundary []struct {
		UID, State string
		Value      *float64
		Row        exploration.PodResourceObservation
		Read       pnRead
		Results    []measurements.Result
		Snapshot   pnSnapshot
	}
	Blocked []string
	Events  []struct {
		Action        string
		Source, Plans int
	}
	Requests  []struct{ Endpoint, Effect string }
	Responses []struct {
		Endpoint string
		Status   int
		Body     json.RawMessage
	}
}

func pnRaw(f psMap, r *measurements.Result) bool {
	q, d := r.Question, r.Question.Definition
	s := d.Source.ExactStream
	if d.Kind != "gauge" || d.Source.Resource == nil || d.Source.Resource.Kind != "kubernetes-pod" || s == nil || d.Source.Resource.ID != s.Resource["k8s.pod.uid"] || d.Source.Resource.ID == "" || len(d.Source.Services) != 0 || d.Source.Namespace != nil || d.Environment != f["environment"] || r.Gauges == nil || len(r.Gauges.Series) != 1 || !reflect.DeepEqual(r.Gauges.Series[0].Stream, *s) {
		return false
	}
	lo, hi := q.BoundsNS()
	groups := map[string][]psMap{}
	count := 0
	expected := map[string]int{}
	for _, v := range psArr(f["points"]) {
		p := psObj(v)
		if p["kind"] != "gauge" || p["metric"] != s.Name || p["unit"] != s.Unit || p["scope"] != s.Scope || p["version"] != s.Version || p["schemaURL"] != s.SchemaURL || p["resourceSchemaURL"] != s.ResourceSchemaURL || !psEq(p["resource"], s.Resource) || !psEq(p["attributes"], s.Attributes) {
			continue
		}
		at := p["endNS"].(string)
		if at < lo || at >= hi {
			continue
		}
		groups[at] = append(groups[at], p)
		expected[siKey([]any{p["endNS"], p["value"], p["flags"]})]++
		count++
	}
	if count == 0 || r.Count != strconv.Itoa(count) || len(r.Rows) != count {
		return false
	}
	for _, raw := range r.Rows {
		var row psMap
		_ = json.Unmarshal(raw, &row)
		if !psEq(row["resource"], s.Resource) || !psEq(row["attributes"], s.Attributes) || row["resource_schema_url"] != s.ResourceSchemaURL || row["schema_url"] != s.SchemaURL || row["scope"] != s.Scope || row["version"] != s.Version || row["name"] != s.Name || row["unit"] != s.Unit {
			return false
		}
		key := siKey([]any{row["end_ns"], row["value"], row["flags"]})
		if expected[key] == 0 {
			return false
		}
		expected[key]--
	}
	times := []string{}
	for at := range groups {
		times = append(times, at)
	}
	sort.Strings(times)
	samples := r.Gauges.Series[0].Samples
	if len(samples) != len(times) {
		return false
	}
	to, _ := strconv.ParseUint(hi, 10, 64)
	for n, at := range times {
		pts := groups[at]
		sample := samples[n]
		if sample.AtNS != at {
			return false
		}
		missing := false
		values := map[float64]bool{}
		for _, p := range pts {
			if p["flags"] == float64(1) {
				missing = true
			}
			v, e := strconv.ParseFloat(p["value"].(string), 64)
			if e != nil {
				return false
			}
			values[v] = true
		}
		state := "observed"
		if missing || len(values) > 1 {
			state = "unknown"
		} else {
			ns, _ := strconv.ParseUint(at, 10, 64)
			if to-ns > uint64(d.MaxGapMS)*1000000 {
				state = "stale"
			}
		}
		if sample.State != state {
			return false
		}
		if state == "unknown" {
			if sample.Value != nil {
				return false
			}
		} else {
			if sample.Value == nil || !values[*sample.Value] {
				return false
			}
		}
	}
	return true
}
func pnReadOK(read pnRead, results []measurements.Result) bool {
	q := read.DiscoveryRequest
	if q.ResourceSchemaURL == nil || !reflect.DeepEqual(q, read.Discovery.Question) || !aeJSONEqual(read.PlanRequest.Discovery, read.Discovery) || !reflect.DeepEqual(read.Plan.Question, q) || len(read.Plan.Widgets) != len(results) || len(results) == 0 || len(results) > 2 {
		return false
	}
	for _, r := range results {
		found := false
		for _, w := range read.Plan.Widgets {
			if aeJSONEqual(w.Question, r.Question) {
				found = true
				s := r.Question.Definition.Source.ExactStream
				if s == nil || !reflect.DeepEqual(s.Resource, q.Resource) || s.ResourceSchemaURL != *q.ResourceSchemaURL || r.Question.FromMS != q.FromMS || r.Question.ToMS != q.ToMS || w.Key != measurements.Hash(r.Question) {
					return false
				}
				if w.Profile == "collector-kubeletstats-v0.137.0" {
					if s.Scope != measurements.KubeletStatsScope || s.Version != "0.137.0" || s.SchemaURL != "" || s.Unit != "By" || (s.Name != "k8s.pod.memory.usage" && s.Name != "k8s.pod.memory.working_set") {
						return false
					}
				} else if w.Profile == "otel-k8s-v1.37.0" {
					if s.SchemaURL != "https://opentelemetry.io/schemas/1.37.0" || s.Name != "k8s.pod.memory.usage" || s.Unit != "By" {
						return false
					}
				} else {
					return false
				}
			}
		}
		if !found {
			return false
		}
	}
	return true
}
func pnLiveOK(d pnLive) bool {
	if !pdOracle(pdObservedFacts(d.Facts, d.Initial.Request), &d.Initial.Response) || len(d.PrimaryResults) != 2 || !pnReadOK(d.PrimaryRead, d.PrimaryResults) || !reflect.DeepEqual(d.Saved, d.Reopened) || len(d.Blocked) != 0 || len(d.Boundary) != 8 {
		return false
	}
	for n := range d.PrimaryResults {
		r := &d.PrimaryResults[n]
		if !pnRaw(d.Facts, r) {
			return false
		}
		seen := false
		for _, response := range d.Responses {
			if response.Endpoint == "/api/measurements" && response.Status == 200 {
				var actual measurements.Result
				if json.Unmarshal(response.Body, &actual) == nil && aeJSONEqual(actual, *r) {
					seen = true
				}
			}
		}
		if !seen {
			return false
		}
	}
	if len(d.Retained.Widgets) != 2 {
		return false
	}
	for _, r := range d.PrimaryResults {
		samples := r.Gauges.Series[0].Samples
		latest := samples[len(samples)-1]
		found := false
		for _, w := range d.Retained.Widgets {
			if w.Key == measurements.Hash(r.Question) {
				found = true
				if latest.Value == nil || w.Title != strconv.FormatFloat(*latest.Value, 'f', -1, 64)+" By" {
					return false
				}
			}
		}
		if !found {
			return false
		}
	}
	networkObserved := func(r measurements.Result) bool {
		for _, response := range d.Responses {
			if response.Endpoint == "/api/measurements" && response.Status == 200 {
				var actual measurements.Result
				if json.Unmarshal(response.Body, &actual) == nil && aeJSONEqual(actual, r) {
					return true
				}
			}
		}
		return false
	}
	if !networkObserved(d.Individual) {
		return false
	}
	if !pnRaw(d.Facts, &d.Individual) {
		return false
	}
	for _, b := range d.Boundary {
		if len(b.Results) != 1 || len(b.Snapshot.Widgets) != 1 || !pnReadOK(b.Read, b.Results) || !pnRaw(d.Facts, &b.Results[0]) {
			return false
		}
		r := b.Results[0]
		if !networkObserved(r) {
			return false
		}
		samples := r.Gauges.Series[0].Samples
		if len(samples) == 0 {
			return false
		}
		latest := samples[len(samples)-1]
		widget := b.Snapshot.Widgets[0]
		invalid := latest.Value != nil && (math.IsNaN(*latest.Value) || math.IsInf(*latest.Value, 0) || *latest.Value < 0 || math.Trunc(*latest.Value) != *latest.Value || *latest.Value > 9007199254740991)
		state := latest.State
		if invalid {
			state = "invalid"
		}
		if state != b.State || b.Value != nil && (latest.Value == nil || *latest.Value != *b.Value) {
			return false
		}
		if invalid {
			if len(widget.Timeline.Points) != len(samples) {
				return false
			}
			if widget.Value != "Invalid reported byte count" {
				return false
			}
			for _, p := range widget.Timeline.Points {
				if p.State != "unknown" || p.Value != nil {
					return false
				}
			}
		} else if state == "unknown" && widget.Value != "Unavailable" {
			return false
		} else if (state == "observed" || state == "stale") && (latest.Value == nil || widget.Title != strconv.FormatFloat(*latest.Value, 'f', -1, 64)+" By") {
			return false
		}
	}
	sources := []int{1, 4, 4, 4, 5, 6, 6, 8, 10, 12, 14, 16, 18, 20, 22, 22}
	if len(d.Events) != len(sources) {
		return false
	}
	for n, e := range d.Events {
		if e.Source != sources[n] {
			return false
		}
	}
	source, metadata, native, capture := 0, 0, 0, 0
	for _, q := range d.Requests {
		if q.Effect == "source" {
			source++
			switch q.Endpoint {
			case "/api/pod-metric-discovery":
				metadata++
			case "/api/measurements":
				native++
			case "/api/captures":
				capture++
			}
		}
	}
	return source == 22 && metadata == 9 && native == 11 && capture == 1
}
func TestPodNativeObserved(t *testing.T) {
	path := os.Getenv("GOO_POD_NATIVE_LEDGER")
	if path == "" {
		t.Skip("set GOO_POD_NATIVE_LEDGER")
	}
	raw, e := os.ReadFile(path)
	if e != nil {
		t.Fatal(e)
	}
	var d pnLive
	if e := json.Unmarshal(raw, &d); e != nil {
		t.Fatal("ledger wire decode", e)
	}
	if !pnLiveOK(d) {
		t.Fatal("pod native raw/profile/mask/clock/source oracle")
	}

	a, e := os.ReadFile(filepath.Join(filepath.Dir(path), d.StoredBefore.File))
	if e != nil {
		t.Fatal(e)
	}
	z, e := os.ReadFile(filepath.Join(filepath.Dir(path), d.StoredAfter.File))
	h := sha256.Sum256(a)
	var s exploration.SavedCapture
	if e != nil || string(a) != string(z) || len(a) != d.StoredBefore.Bytes || hex.EncodeToString(h[:]) != d.StoredBefore.SHA256 || json.Unmarshal(a, &s) != nil || s.ID != d.Saved.Receipt.ID || !reflect.DeepEqual(s.CaptureReceipt, d.Saved.Receipt) || !peLiveCaptureHash(exploration.CapturePage{Receipt: s.CaptureReceipt, Measurement: s.Measurement}) || !pnRaw(d.Facts, s.Measurement) || !reflect.DeepEqual(s.Measurement.Question, d.Individual.Question) {
		t.Fatal("full raw pod gauge custody")
	}
	controls := []func(*pnLive){func(x *pnLive) { x.PrimaryResults[0].Question.Definition.Source.Resource.ID = "other" }, func(x *pnLive) {
		x.PrimaryResults[0].Question.Definition.Source.ExactStream.Resource["k8s.cluster.uid"] = ""
	}, func(x *pnLive) {
		x.PrimaryResults[0].Question.Definition.Source.ExactStream.ResourceSchemaURL = "other"
	}, func(x *pnLive) { x.PrimaryResults[0].Rows = nil }, func(x *pnLive) { x.PrimaryResults[0].Basis = "changed" }, func(x *pnLive) { x.PrimaryResults[0].ReadAt = "changed" }, func(x *pnLive) { x.Events[2].Source++ }, func(x *pnLive) { x.Boundary[5].Snapshot.Widgets[0].Value = "0 By" }, func(x *pnLive) { x.PrimaryRead.DiscoveryRequest.ResourceSchemaURL = nil }, func(x *pnLive) { x.Reopened.Receipt.ID = "changed" }}
	for n, change := range controls {
		var x pnLive
		_ = json.Unmarshal(raw, &x)
		change(&x)
		if reflect.DeepEqual(x, d) || pnLiveOK(x) {
			t.Fatalf("pod native changed control %d accepted", n)
		}
	}
}
func TestPodNativeOwnershipObserved(t *testing.T) {
	path := os.Getenv("GOO_POD_NATIVE_OWNER_LEDGER")
	if path == "" {
		t.Skip("set GOO_POD_NATIVE_OWNER_LEDGER")
	}
	raw, e := os.ReadFile(path)
	if e != nil {
		t.Fatal(e)
	}
	var d psMap
	_ = json.Unmarshal(raw, &d)
	check := func(x psMap) bool {
		cases := psArr(x["cases"])
		if len(cases) != 5 {
			return false
		}
		for _, v := range cases {
			c := psObj(v)
			held := psObj(c["held"])
			before, after := psObj(c["beforeRelease"]), psObj(c["afterRelease"])
			if held["status"] != float64(200) || !psEq(before, after) {
				return false
			}
			if c["phase"] == "native" {
				var r measurements.Result
				b, _ := json.Marshal(held["response"])
				if json.Unmarshal(b, &r) != nil || !pnRaw(psObj(x["facts"]), &r) {
					return false
				}
			} else {
				var q measurements.PodMetricDiscoveryQuestion
				var r measurements.PodMetricDiscovery
				b, _ := json.Marshal(held["request"])
				_ = json.Unmarshal(b, &q)
				b, _ = json.Marshal(held["response"])
				_ = json.Unmarshal(b, &r)
				if q.ResourceSchemaURL == nil || !reflect.DeepEqual(q, r.Question) || after["plans"] != psObj(c["beforeOpen"])["plans"] {
					return false
				}
			}
		}
		return true
	}
	if !check(d) {
		t.Fatal("real pod metadata/native response published after owner invalidation")
	}
	for _, key := range []string{"source", "plans", "panels", "url"} {
		var x psMap
		_ = json.Unmarshal(raw, &x)
		psObj(psObj(psArr(x["cases"])[0])["afterRelease"])[key] = "changed"
		if check(x) {
			t.Fatal("owner changed control", key)
		}
	}
}
