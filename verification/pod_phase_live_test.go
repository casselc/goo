package verification

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/metrics"
	"math/big"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
	"testing"
	"time"

	"github.com/casselc/goo/measurements"
)

// ppRaw derives categorical domain and observation quality independently from
// ordinary raw fixture records; it never calls the production phase decoder.
func ppRaw(f psMap, r *measurements.Result) bool {
	return categoricalRaw(f, r, false)
}

func categoricalRaw(f psMap, r *measurements.Result, container bool) bool {
	profile, kind, identity, metric := "k8scluster-pod-phase/0.137.0", "kubernetes-pod", "k8s.pod.uid", "k8s.pod.phase"
	labels := map[int64]string{1: "Pending", 2: "Running", 3: "Succeeded", 4: "Failed", 5: "Unknown"}
	if container {
		profile, kind, identity, metric = "k8scluster-container-ready/0.137.0", "kubernetes-container", "container.id", "k8s.container.ready"
		labels = map[int64]string{0: "Not ready", 1: "Ready"}
	}
	q, d := r.Question, r.Question.Definition
	if _, e := time.Parse(time.RFC3339Nano, r.ReadAt); e != nil {
		return false
	}
	if b, e := hex.DecodeString(r.Basis); e != nil || len(b) != 32 {
		return false
	}

	s := d.Source.ExactStream
	if d.Kind != "state" || d.StateProfile != profile || d.Source.Resource == nil || d.Source.Resource.Kind != kind || s == nil || d.Source.Resource.ID == "" || d.Source.Resource.ID != s.Resource[identity] || len(d.Source.Services) != 0 || d.Source.Namespace != nil || d.Environment != f["environment"] || r.States == nil || r.Gauges != nil || r.Stocks != nil || r.Counters != nil || r.Histograms != nil {
		return false
	}
	if d.Grain != "per-stream" || d.StockDomain != "" || d.HistogramBucketMS != nil || s.Resource["deployment.environment.name"] != d.Environment || d.Source.Metric != s.Name || d.Source.Unit != s.Unit || d.Source.Scope != s.Scope || d.Source.Version != s.Version || d.Source.SchemaURL != s.SchemaURL {
		return false
	}
	for k, v := range d.Source.Attributes {
		if s.Attributes[k] != v {
			return false
		}
	}
	if container && (s.Resource["k8s.pod.uid"] == "" || s.Resource["k8s.container.name"] == "") {
		return false
	}
	if s.Name != metric || s.Unit != "" || s.Scope != "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/k8sclusterreceiver" || s.Version != "0.137.0" || s.SchemaURL != "" {
		return false
	}
	lo, hi := q.BoundsNS()
	low, e := strconv.ParseUint(lo, 10, 64)
	if e != nil {
		return false
	}
	high, e := strconv.ParseUint(hi, 10, 64)
	if e != nil {
		return false
	}
	groups := map[string][]psMap{}
	physical := map[string]int{}
	count := 0
	for _, v := range psArr(f["points"]) {
		p := psObj(v)
		if p["kind"] != "gauge" || p["metric"] != s.Name || p["unit"] != s.Unit || p["scope"] != s.Scope || p["version"] != s.Version || p["schemaURL"] != s.SchemaURL || p["resourceSchemaURL"] != s.ResourceSchemaURL || !psEq(p["resource"], s.Resource) || !psEq(p["attributes"], s.Attributes) {
			continue
		}
		at, _ := p["endNS"].(string)
		ns, e := strconv.ParseUint(at, 10, 64)
		if e != nil || ns < low || ns >= high {
			continue
		}
		groups[at] = append(groups[at], p)
		physical[siKey([]any{at, p["value"], p["flags"]})]++
		count++
	}
	if r.Count != strconv.Itoa(count) || len(r.Rows) != count || r.States.FromNS != lo || r.States.ToNS != hi {
		return false
	}
	for _, raw := range r.Rows {
		var row psMap
		if json.Unmarshal(raw, &row) != nil {
			return false
		}
		if !psEq(row["resource"], s.Resource) || !psEq(row["attributes"], s.Attributes) || row["resource_schema_url"] != s.ResourceSchemaURL || row["schema_url"] != s.SchemaURL || row["scope"] != s.Scope || row["version"] != s.Version || row["name"] != s.Name || row["unit"] != "" {
			return false
		}
		key := siKey([]any{row["end_ns"], row["value"], row["flags"]})
		if physical[key] == 0 {
			return false
		}
		physical[key]--
	}
	if count == 0 {
		return r.State == "no_data" && len(r.States.Series) == 0
	}
	if len(r.States.Series) != 1 || !reflect.DeepEqual(r.States.Series[0].Stream, *s) {
		return false
	}
	series := r.States.Series[0]
	times := []string{}
	for at := range groups {
		times = append(times, at)
	}
	sort.Slice(times, func(i, j int) bool {
		a, _ := strconv.ParseUint(times[i], 10, 64)
		b, _ := strconv.ParseUint(times[j], 10, 64)
		return a < b
	})
	if len(series.Samples) != len(times) {
		return false
	}
	latestQuality, latestDomain := "", ""
	for n, at := range times {
		points := groups[at]
		missing := false
		values := map[string]*big.Rat{}
		for _, p := range points {
			flags, ok := p["flags"].(float64)
			if !ok {
				return false
			}
			if int(flags)&1 != 0 {
				missing = true
			}
			raw, ok := p["value"].(string)
			if !ok {
				return false
			}
			v, ok := new(big.Rat).SetString(raw)
			if !ok {
				return false
			}
			values[v.RatString()] = v
		}
		quality := "observed"
		ns, _ := strconv.ParseUint(at, 10, 64)
		if missing || len(values) > 1 {
			quality = "unknown"
		} else if high-ns > uint64(d.MaxGapMS)*1000000 {
			quality = "stale"
		}
		sample := series.Samples[n]
		if sample.AtNS != at || sample.State != quality {
			return false
		}
		domain := ""
		label := ""
		if quality == "unknown" {
			if sample.Code != nil || sample.Label != "" || sample.CodeState != "" {
				return false
			}
		} else {
			if sample.Code == nil {
				return false
			}
			var value *big.Rat
			for _, v := range values {
				value = v
			}
			expected, _ := value.Float64()
			if *sample.Code != expected {
				return false
			}
			domain = "unsupported"
			if value.IsInt() && value.Num().IsInt64() {
				if name, ok := labels[value.Num().Int64()]; ok {
					domain = "supported"
					label = name
				}
			}
			if sample.CodeState != domain || sample.Label != label {
				return false
			}
		}
		latestQuality, latestDomain = quality, domain
	}
	overall := "observed"
	if latestQuality != "observed" || latestDomain == "unsupported" {
		overall = "unknown"
	}
	return series.State == latestQuality && r.State == overall
}

// ppRead binds the existing exact-resource metadata/plan protocol; phase uses
// the physical gauge family but an explicitly categorical native definition.
type ppStateSnapshot struct {
	Widgets []struct {
		Key    string
		States []struct {
			Data                             struct{ Samples []metrics.StateSample }
			Current                          *metrics.StateSample
			State, CodeState, Label, Quality string
		}
	}
}
type ppLive struct {
	Facts   psMap
	Initial struct {
		Request  exploration.PodDiscoveryQuestion
		Response exploration.PodDiscovery
	}
	Mixed, Primary            exploration.PodResourceObservation
	MixedRead, PrimaryRead    pnRead
	PrimaryResults            []measurements.Result
	Retained                  ppStateSnapshot
	Individual                measurements.Result
	Saved, Reopened           exploration.CapturePage
	StoredBefore, StoredAfter struct {
		File, SHA256 string
		Bytes        int
	}
	Boundary []struct {
		UID, Quality, Label, Domain string
		Code                        *float64
		Row                         exploration.PodResourceObservation
		Read                        pnRead
		Results                     []measurements.Result
		Snapshot                    ppStateSnapshot
	}
	Refusals []struct {
		UID  string
		Row  exploration.PodResourceObservation
		Read pnRead
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

func ppMetadata(f psMap, read pnRead) bool {
	q := read.DiscoveryRequest
	if q.ResourceSchemaURL == nil || !reflect.DeepEqual(q, read.Discovery.Question) || !aeJSONEqual(read.PlanRequest.Discovery, read.Discovery) || !reflect.DeepEqual(q, read.Plan.Question) {
		return false
	}
	lo, hi := uint64(q.FromMS)*1000000, uint64(q.ToMS)*1000000
	groups := map[string][]psMap{}
	count := 0
	for _, v := range psArr(f["points"]) {
		p := psObj(v)
		if !psEq(p["resource"], q.Resource) || p["resourceSchemaURL"] != *q.ResourceSchemaURL {
			continue
		}
		at, e := strconv.ParseUint(p["endNS"].(string), 10, 64)
		if e != nil || at < lo || at >= hi {
			continue
		}
		kind := p["kind"].(string)
		tempo, mono := p["temporality"], p["monotonic"]
		if kind == "gauge" {
			tempo = float64(0)
			mono = false
		}
		if kind == "sum" {
			kind = "counter"
			if tempo == float64(2) && mono == false {
				kind = "stock"
			}
		}
		key := siKey([]any{kind, p["scope"], p["version"], p["schemaURL"], p["metric"], p["unit"], tempo, mono})
		groups[key] = append(groups[key], p)
		count++
	}
	if read.Discovery.Records != count || read.Discovery.InstrumentsOmitted != 0 || len(read.Discovery.Instruments) != len(groups) {
		return false
	}
	for _, i := range read.Discovery.Instruments {
		key := siKey([]any{i.Kind, i.Scope, i.Version, i.SchemaURL, i.Metric, i.Unit, i.Temporality, i.Monotonic})
		points, ok := groups[key]
		if !ok || i.Records != len(points) || i.CapabilityPlan == nil {
			return false
		}
		for _, preview := range i.CapabilityPlan.Coverage.Streams {
			s := preview.Stream
			found := false
			for _, p := range points {
				if psEq(p["resource"], s.Resource) && psEq(p["attributes"], s.Attributes) && p["resourceSchemaURL"] == s.ResourceSchemaURL && p["scope"] == s.Scope && p["version"] == s.Version && p["schemaURL"] == s.SchemaURL && p["metric"] == s.Name && p["unit"] == s.Unit {
					found = true
				}
			}
			if !found {
				return false
			}
		}
		delete(groups, key)
	}
	return len(groups) == 0
}
func ppQuestion(read pnRead, r measurements.Result) bool {
	if len(read.Plan.Widgets) != 1 {
		return false
	}
	w := read.Plan.Widgets[0]
	q := read.DiscoveryRequest
	s := r.Question.Definition.Source.ExactStream
	return s != nil && w.Profile == "collector-k8scluster-v0.137.0" && w.Key == measurements.Hash(r.Question) && aeJSONEqual(w.Question, r.Question) && reflect.DeepEqual(s.Resource, q.Resource) && s.ResourceSchemaURL == *q.ResourceSchemaURL && r.Question.FromMS == q.FromMS && r.Question.ToMS == q.ToMS
}
func ppSnapshot(r measurements.Result, snapshot ppStateSnapshot) bool {
	if len(snapshot.Widgets) != 1 || len(snapshot.Widgets[0].States) != 1 || snapshot.Widgets[0].Key != measurements.Hash(r.Question) {
		return false
	}
	view := snapshot.Widgets[0].States[0]
	series := r.States.Series[0]
	if !aeJSONEqual(view.Data.Samples, series.Samples) || len(series.Samples) == 0 {
		return false
	}
	latest := series.Samples[len(series.Samples)-1]
	if view.Current == nil || !aeJSONEqual(*view.Current, latest) {
		return false
	}
	label, domain, quality := "Unavailable", "unavailable", "Observation unavailable"
	if latest.State == "observed" || latest.State == "stale" {
		domain = latest.CodeState
		quality = "Observed"
		if latest.State == "stale" {
			quality = "Stale"
		}
		if latest.CodeState == "supported" {
			label = latest.Label
		} else {
			if latest.Code == nil {
				return false
			}
			label = "Unsupported code " + strconv.FormatFloat(*latest.Code, 'f', -1, 64)
		}
	}
	return view.State == latest.State && view.CodeState == domain && view.Label == label && view.Quality == quality
}
func ppLiveOK(d ppLive) bool {
	if !pdOracle(pdObservedFacts(d.Facts, d.Initial.Request), &d.Initial.Response) || !ppMetadata(d.Facts, d.MixedRead) || !ppMetadata(d.Facts, d.PrimaryRead) || len(d.MixedRead.Plan.Widgets) != 2 || len(d.PrimaryResults) != 1 || !reflect.DeepEqual(d.Saved, d.Reopened) || len(d.Blocked) != 0 {
		return false
	}
	candidate := false
	for _, w := range d.MixedRead.Plan.Candidates {
		if w.Question.Definition.Kind == "state" {
			candidate = true
		}
	}
	if !candidate {
		return false
	}
	for _, w := range d.MixedRead.Plan.Widgets {
		if w.Question.Definition.Kind != "gauge" || w.Profile != "collector-kubeletstats-v0.137.0" {
			return false
		}
	}
	network := func(r measurements.Result) bool {
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
	r := d.PrimaryResults[0]
	if !ppRaw(d.Facts, &r) || !ppQuestion(d.PrimaryRead, r) || !network(r) || !ppSnapshot(r, d.Retained) || !ppRaw(d.Facts, &d.Individual) || !network(d.Individual) || !aeJSONEqual(r.Question, d.Individual.Question) {
		return false
	}
	// The mixed resource still executes only the existing memory default pair.
	for _, w := range d.MixedRead.Plan.Widgets {
		found := false
		for _, response := range d.Responses {
			if response.Endpoint != "/api/measurements" || response.Status != 200 {
				continue
			}
			var actual measurements.Result
			if json.Unmarshal(response.Body, &actual) == nil && aeJSONEqual(actual.Question, w.Question) && pnRaw(d.Facts, &actual) {
				found = true
			}
		}
		if !found {
			return false
		}
	}
	if len(d.Boundary) != 7 || len(d.Refusals) != 5 {
		return false
	}
	for _, b := range d.Boundary {
		if len(b.Results) != 1 || !ppMetadata(d.Facts, b.Read) || !ppQuestion(b.Read, b.Results[0]) || !ppRaw(d.Facts, &b.Results[0]) || !network(b.Results[0]) || !ppSnapshot(b.Results[0], b.Snapshot) {
			return false
		}
	}
	for _, b := range d.Refusals {
		if !ppMetadata(d.Facts, b.Read) || len(b.Read.Plan.Widgets) != 0 || len(b.Read.Plan.Candidates) != 0 {
			return false
		}
	}
	sources := []int{1, 4, 4, 6, 6, 6, 7, 8, 8, 10, 12, 14, 16, 18, 20, 22, 23, 24, 25, 26, 27, 27}
	plans := []int{0, 1, 2, 3, 4, 4, 5, 5, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 17}
	actions := []string{"inventory", "memory-defaults-phase-candidate", "explicit-phase-candidate-staged-back", "phase-only-default", "native-staged", "retained-back", "native-run", "individual-capture", "capture-retained-return", "boundary-reported-unknown", "boundary-unsupported", "boundary-fractional", "boundary-stale-phase", "boundary-stale-unsupported", "boundary-phase-conflict", "boundary-phase-missing", "refusal-wrong-version", "refusal-wrong-scope", "refusal-wrong-schema", "refusal-wrong-unit", "refusal-wrong-kind", "frozen-reopen-no-posts"}
	if len(d.Events) != len(sources) {
		return false
	}
	for n, e := range d.Events {
		if e.Source != sources[n] || e.Plans != plans[n] || e.Action != actions[n] {
			return false
		}
	}
	source, metadata, native, capture, pure := 0, 0, 0, 0, 0
	for _, q := range d.Requests {
		if q.Effect == "plan" {
			pure++
		}
		if q.Effect != "source" {
			continue
		}
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
	return source == 27 && metadata == 14 && native == 11 && capture == 1 && pure == 17
}
func TestPodPhaseObserved(t *testing.T) {
	path := os.Getenv("GOO_POD_PHASE_LEDGER")
	if path == "" {
		t.Skip("set GOO_POD_PHASE_LEDGER to actual203c evidence")
	}
	raw, e := os.ReadFile(path)
	if e != nil {
		t.Fatal(e)
	}
	var d ppLive
	if e = json.Unmarshal(raw, &d); e != nil {
		t.Fatal("ledger decode", e)
	}
	if !ppLiveOK(d) {
		t.Fatal("categorical raw/domain/quality/profile/clock/source oracle")
	}

	a, e := os.ReadFile(filepath.Join(filepath.Dir(path), d.StoredBefore.File))
	if e != nil {
		t.Fatal(e)
	}
	z, e := os.ReadFile(filepath.Join(filepath.Dir(path), d.StoredAfter.File))
	sum := sha256.Sum256(a)
	var stored exploration.SavedCapture
	if e != nil || string(a) != string(z) || len(a) != d.StoredBefore.Bytes || hex.EncodeToString(sum[:]) != d.StoredBefore.SHA256 || json.Unmarshal(a, &stored) != nil || stored.ID != d.Saved.Receipt.ID || !reflect.DeepEqual(stored.CaptureReceipt, d.Saved.Receipt) || !peLiveCaptureHash(exploration.CapturePage{Receipt: stored.CaptureReceipt, Measurement: stored.Measurement}) || !ppRaw(d.Facts, stored.Measurement) || !aeJSONEqual(stored.Measurement.Question, d.Individual.Question) {
		t.Fatal("full categorical capture bytes/content ID/raw replay")
	}
	controls := []func(*ppLive){func(x *ppLive) { x.PrimaryResults[0].States.Series[0].Samples[1].Label = "Healthy" }, func(x *ppLive) { x.Boundary[0].Results[0].States.Series[0].Samples[0].State = "unknown" }, func(x *ppLive) { x.Boundary[1].Results[0].States.Series[0].Samples[0].Label = "Failed" }, func(x *ppLive) { x.Boundary[2].Results[0].States.Series[0].Samples[0].CodeState = "supported" }, func(x *ppLive) { x.PrimaryResults[0].Question.Definition.Source.Resource.ID = "wrong" }, func(x *ppLive) {
		x.PrimaryResults[0].Question.Definition.Source.ExactStream.ResourceSchemaURL = "wrong"
	}, func(x *ppLive) { x.PrimaryResults[0].Rows = nil }, func(x *ppLive) { x.PrimaryResults[0].Basis = "changed" }, func(x *ppLive) { x.PrimaryResults[0].ReadAt = "changed" }, func(x *ppLive) { x.Events[3].Source++ }, func(x *ppLive) { x.Retained.Widgets[0].States[0].Label = "2" }, func(x *ppLive) { x.Reopened.Receipt.ID = "changed" }}
	for n, change := range controls {
		var x ppLive
		if e := json.Unmarshal(raw, &x); e != nil {
			t.Fatal(e)
		}
		change(&x)
		if reflect.DeepEqual(x, d) || ppLiveOK(x) {
			t.Fatalf("categorical changed control %d accepted", n)
		}
	}
}
func TestPodPhaseOwnershipObserved(t *testing.T) {
	path := os.Getenv("GOO_POD_PHASE_OWNER_LEDGER")
	if path == "" {
		t.Skip("set GOO_POD_PHASE_OWNER_LEDGER")
	}
	raw, e := os.ReadFile(path)
	if e != nil {
		t.Fatal(e)
	}
	var d psMap
	if e = json.Unmarshal(raw, &d); e != nil {
		t.Fatal(e)
	}
	check := func(x psMap) bool {
		cases := psArr(x["cases"])
		if len(cases) != 5 {
			return false
		}
		for _, v := range cases {
			c := psObj(v)
			held := psObj(c["held"])
			after := psObj(c["afterRelease"])
			if held["status"] != float64(200) || !psEq(c["beforeRelease"], after) {
				return false
			}
			if c["phase"] == "native" {
				var r measurements.Result
				b, _ := json.Marshal(held["response"])
				if json.Unmarshal(b, &r) != nil || !ppRaw(psObj(x["facts"]), &r) {
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
		t.Fatal("actual phase metadata/native response published after owner withdrawal")
	}
	for _, k := range []string{"source", "plans", "panels", "url"} {
		var x psMap
		_ = json.Unmarshal(raw, &x)
		psObj(psObj(psArr(x["cases"])[0])["afterRelease"])[k] = "changed"
		if check(x) {
			t.Fatal("categorical owner control", k)
		}
	}
}
