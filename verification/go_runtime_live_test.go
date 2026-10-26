package verification

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/measurements"
	"math/big"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
	"testing"
)

type grLive struct {
	Requests []struct{ Endpoint, Effect string }

	Facts                     psMap
	Initial                   struct{ Plan exploration.DashboardPlan }
	Results                   []measurements.Result
	Individual, Goroutine     measurements.Result
	GoroutinePlan             exploration.DashboardPlan
	GoroutineOption           exploration.DashboardNativeOption
	Saved, Reopened           exploration.CapturePage
	StoredBefore, StoredAfter struct {
		File, SHA256 string
		Bytes        int
	}
	Events []struct {
		Action        string
		Source, Plans int
	}
	Responses []struct {
		Endpoint string
		Status   int
		Body     json.RawMessage
	}
}

func grRaw(f psMap, r *measurements.Result) bool {
	q, d := r.Question, r.Question.Definition
	s := d.Source.ExactStream
	if d.Source.Namespace == nil || d.Source.Namespace.Mode != "value" || d.Source.Namespace.Value != f["namespace"] {
		return false
	}
	if d.Source.Resource != nil || d.StockDomain != "nonnegative-integer" || s == nil || s.Resource["service.name"] != f["service"] || s.Resource["service.instance.id"] == "" || d.Environment != f["environment"] || d.Source.SchemaURL != "https://opentelemetry.io/schemas/1.37.0" || len(d.Source.Services) != 1 || d.Source.Services[0] != f["service"] {
		return false
	}
	unit := "By"
	if s.Name == "go.goroutine.count" {
		unit = "{goroutine}"
	} else if s.Name != "go.memory.used" && s.Name != "go.memory.gc.goal" {
		return false
	}
	if s.Unit != unit {
		return false
	}
	lo, hi := q.BoundsNS()
	pts := []psMap{}
	for _, v := range psArr(f["points"]) {
		p := psObj(v)
		if p["kind"] != "stock" || p["temporality"] != float64(2) || p["monotonic"] != false || p["metric"] != s.Name || p["unit"] != s.Unit || p["scope"] != s.Scope || p["version"] != s.Version || p["schemaURL"] != s.SchemaURL || p["resourceSchemaURL"] != s.ResourceSchemaURL || !psEq(p["resource"], s.Resource) || !psEq(p["attributes"], s.Attributes) {
			continue
		}
		if p["endNS"].(string) >= lo && p["endNS"].(string) < hi {
			pts = append(pts, p)
		}
	}
	sort.Slice(pts, func(i, j int) bool { return pts[i]["endNS"].(string) < pts[j]["endNS"].(string) })
	if len(pts) == 0 || r.Count != strconv.Itoa(len(pts)) || len(r.Rows) != len(pts) || r.Stocks == nil || len(r.Stocks.Series) != 1 || !reflect.DeepEqual(r.Stocks.Series[0].Stream, *s) || len(r.Stocks.Series[0].Samples) != len(pts) {
		return false
	}
	for n, p := range pts {
		var row psMap
		_ = json.Unmarshal(r.Rows[n], &row)
		if !psEq(row["resource"], p["resource"]) || !psEq(row["attributes"], p["attributes"]) || row["resource_schema_url"] != p["resourceSchemaURL"] || row["schema_url"] != p["schemaURL"] || row["scope"] != p["scope"] || row["version"] != p["version"] || row["name"] != p["metric"] || row["unit"] != p["unit"] || row["start_ns"] != p["startNS"] || row["end_ns"] != p["endNS"] || row["value"] != p["value"] || row["temporality"] != float64(2) || row["monotonic"] != false {
			return false
		}
		v, ok := new(big.Int).SetString(p["value"].(string), 10)
		if !ok || v.Sign() < 0 || v.Cmp(new(big.Int).Lsh(big.NewInt(1), 53)) >= 0 {
			return false
		}
		sample := r.Stocks.Series[0].Samples[n]
		if sample.State != "observed" || sample.StartNS != p["startNS"] || sample.AtNS != p["endNS"] || sample.Value == nil || *sample.Value != float64(v.Int64()) {
			return false
		}
	}
	return true
}
func grObserved(d grLive) bool {
	if len(d.Results) != 2 || !reflect.DeepEqual(d.Saved, d.Reopened) || !grRaw(d.Facts, &d.Individual) || !grRaw(d.Facts, &d.Goroutine) {
		return false
	}
	metrics := map[string]bool{}
	for n := range d.Results {
		r := &d.Results[n]
		if r.Question.FromMS != d.Initial.Plan.Scope.FromMS || r.Question.ToMS != d.Initial.Plan.Scope.ToMS {
			return false
		}
		if !grRaw(d.Facts, r) {
			return false
		}
		metrics[r.Question.Definition.Source.Metric] = true
		seen := false
		for _, response := range d.Responses {
			if response.Endpoint != "/api/measurements" || response.Status != 200 {
				continue
			}
			var actual measurements.Result
			if json.Unmarshal(response.Body, &actual) == nil && aeJSONEqual(actual, *r) {
				seen = true
			}
		}
		if !seen {
			return false
		}
	}
	if !metrics["go.memory.used"] || !metrics["go.memory.gc.goal"] || !reflect.DeepEqual(d.Results[0].Question.Definition.Source.ExactStream.Resource, d.Results[1].Question.Definition.Source.ExactStream.Resource) {
		return false
	}
	if d.Individual.Question.FromMS != d.Initial.Plan.Scope.FromMS || d.Individual.Question.ToMS != d.Initial.Plan.Scope.ToMS || d.Goroutine.Question.FromMS != d.Initial.Plan.Scope.FromMS || d.Goroutine.Question.ToMS != d.Initial.Plan.Scope.ToMS {
		return false
	}
	defaults := 0
	for _, o := range d.Initial.Plan.NativeOptions {
		if o.Default {
			defaults++
			if o.Instrument.Metric == "go.goroutine.count" || o.Instrument.Definition == nil || o.Instrument.Definition.Source.ExactStream == nil || o.Key != measurements.Hash(*o.Instrument.Definition) {
				return false
			}
		}
	}
	if defaults != 2 || d.Goroutine.Question.Definition.Source.Metric != "go.goroutine.count" || d.GoroutineOption.Key != measurements.Hash(d.Goroutine.Question.Definition) {
		return false
	}
	native := 0
	for _, w := range d.GoroutinePlan.Widgets {
		if w.Measurement != nil {
			native++
			if !aeJSONEqual(*w.Measurement, d.Goroutine.Question) {
				return false
			}
		}
	}
	if native != 1 {
		return false
	}
	events := map[string]int{}
	for _, e := range d.Events {
		events[e.Action] = e.Source
	}
	for _, pair := range [][2]string{{"overview", "native-staged"}, {"native-staged", "retained-back"}, {"individual-capture", "capture-retained-service-return"}, {"capture-retained-service-return", "goroutine-selection-staged"}, {"goroutine-alone-read", "immutable-reopen"}} {
		a, ok := events[pair[0]]
		b, ok2 := events[pair[1]]
		if !ok || !ok2 || a != b {
			return false
		}
	}
	source, measurement, captures := 0, 0, 0
	for _, request := range d.Requests {
		if request.Effect == "source" {
			source++
			if request.Endpoint == "/api/measurements" {
				measurement++
			}
			if request.Endpoint == "/api/captures" {
				captures++
			}
		}
	}
	if source != events["immutable-reopen"] || measurement != 4 || captures != 1 {
		return false
	}
	return events["native-run"] == events["retained-back"]+1 && events["individual-capture"] == events["native-run"]+1 && events["goroutine-alone-read"] > events["goroutine-selection-staged"]
}
func TestGoRuntimeObserved(t *testing.T) {
	path := os.Getenv("GOO_GO_RUNTIME_LEDGER")
	if path == "" {
		t.Skip("set GOO_GO_RUNTIME_LEDGER")
	}
	raw, e := os.ReadFile(path)
	if e != nil {
		t.Fatal(e)
	}
	var d grLive
	if json.Unmarshal(raw, &d) != nil || !grObserved(d) {
		t.Fatal("runtime raw populations/defaults/individual source discipline")
	}
	a, e := os.ReadFile(filepath.Join(filepath.Dir(path), d.StoredBefore.File))
	if e != nil {
		t.Fatal(e)
	}
	z, e := os.ReadFile(filepath.Join(filepath.Dir(path), d.StoredAfter.File))
	h := sha256.Sum256(a)
	var s exploration.SavedCapture
	if e != nil || string(a) != string(z) || len(a) != d.StoredBefore.Bytes || hex.EncodeToString(h[:]) != d.StoredBefore.SHA256 || json.Unmarshal(a, &s) != nil || s.ID != d.Saved.Receipt.ID || !reflect.DeepEqual(s.CaptureReceipt, d.Saved.Receipt) || !peLiveCaptureHash(exploration.CapturePage{Receipt: s.CaptureReceipt, Measurement: s.Measurement}) || !grRaw(d.Facts, s.Measurement) || !reflect.DeepEqual(s.Measurement.Question, d.Individual.Question) {
		t.Fatal("full individual runtime capture custody")
	}
	controls := []func(*grLive){func(x *grLive) {
		x.Results[0].Question.Definition.Source.ExactStream.Resource["service.instance.id"] = "other"
	}, func(x *grLive) {
		x.Results[0].Question.Definition.Source.ExactStream.Attributes["go.memory.type"] = "merged"
	}, func(x *grLive) { x.Results[0].Question.Definition.Source.SchemaURL = "other" }, func(x *grLive) { x.Results[0].Rows = nil }, func(x *grLive) { x.Results[0].Basis = "changed" }, func(x *grLive) { x.Results[0].ReadAt = "changed" }, func(x *grLive) { x.Events[1].Source++ }, func(x *grLive) { x.Goroutine.Question.Definition.Source.Unit = "1" }, func(x *grLive) { x.Reopened.Receipt.ID = "changed" }}
	for n, change := range controls {
		var x grLive
		_ = json.Unmarshal(raw, &x)
		change(&x)
		if reflect.DeepEqual(x, d) || grObserved(x) {
			t.Fatalf("runtime changed control %d accepted", n)
		}
	}
}
func TestGoRuntimeOwnershipObserved(t *testing.T) {
	path := os.Getenv("GOO_GO_RUNTIME_OWNER_LEDGER")
	if path == "" {
		t.Skip("set GOO_GO_RUNTIME_OWNER_LEDGER")
	}
	raw, e := os.ReadFile(path)
	if e != nil {
		t.Fatal(e)
	}
	var d psMap
	_ = json.Unmarshal(raw, &d)
	check := func(x psMap) bool {
		held := psObj(x["held"])
		var result measurements.Result
		b, _ := json.Marshal(held["response"])
		_ = json.Unmarshal(b, &result)
		after := psObj(x["after"])
		return psEq(x["before"], after) && after["detailHidden"] == true && held["status"] == float64(200) && grRaw(psObj(x["facts"]), &result)
	}
	if !check(d) {
		t.Fatal("held actual runtime response republished after departure")
	}
	for _, k := range []string{"url", "source", "rendered", "detailHidden"} {
		var x psMap
		_ = json.Unmarshal(raw, &x)
		psObj(x["after"])[k] = "changed"
		if check(x) {
			t.Fatal("owner changed control", k)
		}
	}
}
