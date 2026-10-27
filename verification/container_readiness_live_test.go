package verification

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/measurements"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

type crLive struct {
	ppLive
	PairRead pnRead
	PairRow  exploration.PodResourceObservation
}

func crRaw(f psMap, r *measurements.Result) bool { return categoricalRaw(f, r, true) }
func crLiveOK(d crLive) bool {
	if !pdOracle(pdObservedFacts(d.Facts, d.Initial.Request), &d.Initial.Response) || !ppMetadata(d.Facts, d.MixedRead) || !ppMetadata(d.Facts, d.PrimaryRead) || len(d.MixedRead.Plan.Widgets) != 2 || len(d.PrimaryResults) != 1 || !reflect.DeepEqual(d.Saved, d.Reopened) || len(d.Blocked) != 0 {
		return false
	}
	for _, binding := range []struct {
		row  exploration.PodResourceObservation
		read pnRead
	}{{d.Primary, d.PrimaryRead}, {d.Mixed, d.MixedRead}, {d.PairRow, d.PairRead}} {
		if !reflect.DeepEqual(binding.row.Resource, binding.read.DiscoveryRequest.Resource) || !reflect.DeepEqual(binding.row.ResourceSchemaURL, binding.read.DiscoveryRequest.ResourceSchemaURL) {
			return false
		}
		found := false
		for _, row := range d.Initial.Response.Observations {
			if reflect.DeepEqual(row, binding.row) {
				found = true
			}
		}
		if !found {
			return false
		}
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
	if !ppMetadata(d.Facts, d.PairRead) || len(d.PairRead.Plan.Widgets) != 2 {
		return false
	}
	pairKinds := map[string]bool{}
	for _, w := range d.PairRead.Plan.Widgets {
		found := false
		for _, response := range d.Responses {
			if response.Endpoint != "/api/measurements" || response.Status != 200 {
				continue
			}
			var actual measurements.Result
			if json.Unmarshal(response.Body, &actual) != nil || !aeJSONEqual(actual.Question, w.Question) {
				continue
			}
			container := actual.Question.Definition.StateProfile == "k8scluster-container-ready/0.137.0"
			if !categoricalRaw(d.Facts, &actual, container) {
				return false
			}
			pairKinds[actual.Question.Definition.StateProfile] = true
			found = true
		}
		if !found {
			return false
		}
	}
	if len(pairKinds) != 2 {
		return false
	}
	r := d.PrimaryResults[0]
	if !crRaw(d.Facts, &r) || !ppQuestion(d.PrimaryRead, r) || !network(r) || !ppSnapshot(r, d.Retained) || !crRaw(d.Facts, &d.Individual) || !network(d.Individual) || !aeJSONEqual(r.Question, d.Individual.Question) {
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
	if len(d.Boundary) != 8 || len(d.Refusals) != 5 {
		return false
	}
	for _, b := range d.Boundary {
		if len(b.Results) != 1 || !ppMetadata(d.Facts, b.Read) || !ppQuestion(b.Read, b.Results[0]) || !crRaw(d.Facts, &b.Results[0]) || !network(b.Results[0]) || !ppSnapshot(b.Results[0], b.Snapshot) {
			return false
		}
	}
	for _, b := range d.Refusals {
		if !ppMetadata(d.Facts, b.Read) || len(b.Read.Plan.Widgets) != 0 || len(b.Read.Plan.Candidates) != 0 {
			return false
		}
	}
	sources := []int{1, 4, 4, 7, 9, 9, 9, 10, 11, 11, 13, 15, 17, 19, 21, 23, 25, 27, 28, 29, 30, 31, 32, 32}
	plans := []int{0, 1, 2, 3, 4, 5, 5, 6, 6, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 19}
	actions := []string{"inventory", "memory-defaults-readiness-candidate", "explicit-readiness-candidate-staged-back", "phase-readiness-pair", "readiness-only-default", "native-staged", "retained-back", "native-run", "individual-capture", "capture-retained-return", "boundary-reported-container-not-ready", "boundary-reported-container-replacement", "boundary-unsupported", "boundary-fractional", "boundary-stale", "boundary-stale-unsupported", "boundary-conflict", "boundary-missing-data", "refusal-wrong-version", "refusal-wrong-scope", "refusal-wrong-schema", "refusal-wrong-unit", "refusal-wrong-kind", "frozen-reopen-no-posts"}
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
	return source == 32 && metadata == 16 && native == 14 && capture == 1 && pure == 19
}
func TestContainerReadinessObserved(t *testing.T) {
	path := os.Getenv("GOO_CONTAINER_READY_LEDGER")
	if path == "" {
		t.Skip("set GOO_CONTAINER_READY_LEDGER to actual204b evidence")
	}
	raw, e := os.ReadFile(path)
	if e != nil {
		t.Fatal(e)
	}
	var d crLive
	if e = json.Unmarshal(raw, &d); e != nil {
		t.Fatal("ledger decode", e)
	}
	if !crLiveOK(d) {
		t.Fatal("categorical raw/domain/quality/profile/clock/source oracle")
	}

	a, e := os.ReadFile(filepath.Join(filepath.Dir(path), d.StoredBefore.File))
	if e != nil {
		t.Fatal(e)
	}
	z, e := os.ReadFile(filepath.Join(filepath.Dir(path), d.StoredAfter.File))
	sum := sha256.Sum256(a)
	var stored exploration.SavedCapture
	if e != nil || string(a) != string(z) || len(a) != d.StoredBefore.Bytes || hex.EncodeToString(sum[:]) != d.StoredBefore.SHA256 || json.Unmarshal(a, &stored) != nil || stored.ID != d.Saved.Receipt.ID || !reflect.DeepEqual(stored.CaptureReceipt, d.Saved.Receipt) || !peLiveCaptureHash(exploration.CapturePage{Receipt: stored.CaptureReceipt, Measurement: stored.Measurement}) || !crRaw(d.Facts, stored.Measurement) || !aeJSONEqual(stored.Measurement.Question, d.Individual.Question) {
		t.Fatal("full categorical capture bytes/content ID/raw replay")
	}
	controls := []func(*crLive){func(x *crLive) { x.PrimaryResults[0].States.Series[0].Samples[1].Label = "Healthy" }, func(x *crLive) { x.Boundary[0].Results[0].States.Series[0].Samples[0].State = "unknown" }, func(x *crLive) { x.Boundary[1].Results[0].States.Series[0].Samples[0].Label = "Failed" }, func(x *crLive) { x.Boundary[2].Results[0].States.Series[0].Samples[0].CodeState = "supported" }, func(x *crLive) { x.PrimaryResults[0].Question.Definition.Source.Resource.ID = "wrong" }, func(x *crLive) {
		x.PrimaryResults[0].Question.Definition.Source.ExactStream.ResourceSchemaURL = "wrong"
	}, func(x *crLive) { x.PrimaryResults[0].Rows = nil }, func(x *crLive) { x.PrimaryResults[0].Basis = "changed" }, func(x *crLive) { x.PrimaryResults[0].ReadAt = "changed" }, func(x *crLive) { x.Events[3].Source++ }, func(x *crLive) { x.Retained.Widgets[0].States[0].Label = "2" }, func(x *crLive) { x.Reopened.Receipt.ID = "changed" }}
	for n, change := range controls {
		var x crLive
		if e := json.Unmarshal(raw, &x); e != nil {
			t.Fatal(e)
		}
		change(&x)
		if reflect.DeepEqual(x, d) || crLiveOK(x) {
			t.Fatalf("categorical changed control %d accepted", n)
		}
	}
}
func TestContainerReadinessOwnershipObserved(t *testing.T) {
	path := os.Getenv("GOO_CONTAINER_READY_OWNER_LEDGER")
	if path == "" {
		t.Skip("set GOO_CONTAINER_READY_OWNER_LEDGER")
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
				if json.Unmarshal(b, &r) != nil || !crRaw(psObj(x["facts"]), &r) {
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
