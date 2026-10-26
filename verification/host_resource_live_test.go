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
	"strconv"
	"testing"
)

type hrLive struct {
	Responses []struct {
		Endpoint string
		Status   int
		Body     json.RawMessage
	}

	Facts     psMap
	Inventory psMap
	Primary   psMap
	Plan      measurements.HostDashboardPlan
	Results   []measurements.Result
	Context   []struct {
		Signal   string
		Question exploration.Query
		Result   exploration.Result
	}
	Individual                measurements.Result
	Saved, Reopened           exploration.CapturePage
	StoredBefore, StoredAfter struct {
		File, SHA256 string
		Bytes        int
	}
	Events []struct {
		Action        string
		Source, Plans int
	}
}

func hrRaw(f psMap, r *measurements.Result) bool {
	q := r.Question
	d := q.Definition
	s := d.Source.ExactStream
	if d.Source.Resource == nil || d.Source.Resource.Kind != "host" || s == nil || d.Source.Resource.ID != s.Resource["host.id"] || len(d.Source.Services) != 0 || d.Source.Namespace != nil {
		return false
	}
	lo, hi := q.BoundsNS()
	expected := []psMap{}
	for _, v := range psArr(f["points"]) {
		p := psObj(v)
		if p["metric"] != s.Name || p["unit"] != s.Unit || p["scope"] != s.Scope || p["version"] != s.Version || p["schemaURL"] != s.SchemaURL || p["resourceSchemaURL"] != s.ResourceSchemaURL || !psEq(p["resource"], s.Resource) || !psEq(p["attributes"], s.Attributes) {
			continue
		}
		at := p["endNS"].(string)
		if at >= lo && at < hi {
			expected = append(expected, p)
		}
	}
	if len(expected) == 0 || len(expected) != len(r.Rows) || r.Count != strconv.Itoa(len(expected)) {
		return false
	}
	for n, raw := range r.Rows {
		var row psMap
		_ = json.Unmarshal(raw, &row)
		p := expected[n]
		if !psEq(row["resource"], p["resource"]) || !psEq(row["attributes"], p["attributes"]) || row["resource_schema_url"] != p["resourceSchemaURL"] || row["schema_url"] != p["schemaURL"] || row["scope"] != p["scope"] || row["version"] != p["version"] || row["name"] != p["metric"] || row["unit"] != p["unit"] || row["value"] != p["value"] || row["end_ns"] != p["endNS"] {
			return false
		}
	}

	if d.Kind == "gauge" {
		if r.Gauges == nil || len(r.Gauges.Series) != 1 || !reflect.DeepEqual(r.Gauges.Series[0].Stream, *s) || len(r.Gauges.Series[0].Samples) != len(expected) {
			return false
		}
		for n, p := range expected {
			value, e := strconv.ParseFloat(p["value"].(string), 64)
			v := r.Gauges.Series[0].Samples[n]
			if e != nil || v.State != "observed" || v.AtNS != p["endNS"] || v.Value == nil || *v.Value != value {
				return false
			}
		}
		return true
	}
	if d.Kind == "stock" {
		if r.Stocks == nil || len(r.Stocks.Series) != 1 || !reflect.DeepEqual(r.Stocks.Series[0].Stream, *s) || len(r.Stocks.Series[0].Samples) != len(expected) || d.StockDomain != "nonnegative-integer" {
			return false
		}
		for n, p := range expected {
			value, e := strconv.ParseFloat(p["value"].(string), 64)
			v := r.Stocks.Series[0].Samples[n]
			if e != nil || v.State != "observed" || v.StartNS != p["startNS"] || v.AtNS != p["endNS"] || v.Value == nil || *v.Value != value {
				return false
			}
		}
		return true
	}

	return false
}
func hrLiveOK(d hrLive) bool {
	f := d.Facts
	groups := map[string]int{}
	total := 0
	for _, v := range psArr(f["points"]) {
		p := psObj(v)
		res := psObj(p["resource"])
		if res["deployment.environment.name"] != f["environment"] || res["host.id"] == nil || res["host.id"] == "" {
			continue
		}
		groups[siKey([]any{res, p["resourceSchemaURL"]})]++
		total++
	}
	hosts := psArr(d.Inventory["hosts"])
	if len(hosts) != len(groups) || d.Inventory["records"] != float64(total) {
		return false
	}
	for _, v := range hosts {
		h := psObj(v)
		key := siKey([]any{h["resource"], h["resourceSchemaURL"]})
		if groups[key] == 0 || h["records"] != float64(groups[key]) {
			return false
		}
		delete(groups, key)
	}
	if len(groups) != 0 || len(d.Plan.Widgets) != 2 || len(d.Results) != 2 {
		return false
	}
	for n := range d.Results {
		r := &d.Results[n]
		observed := false
		for _, response := range d.Responses {
			if response.Endpoint != "/api/measurements" || response.Status != 200 {
				continue
			}
			var actual measurements.Result
			if json.Unmarshal(response.Body, &actual) == nil && aeJSONEqual(actual, *r) {
				observed = true
			}
		}
		if !observed {
			return false
		}

		if !hrRaw(f, r) {
			return false
		}
		found := false
		for _, w := range d.Plan.Widgets {
			if reflect.DeepEqual(w.Question, r.Question) {
				found = true
			}
		}
		if !found {
			return false
		}
	}
	primary := psObj(d.Primary["resource"])
	for _, c := range d.Context {
		q := c.Question
		if q.Environment != f["environment"] || q.FromMS != d.Plan.Question.FromMS || q.ToMS != d.Plan.Question.ToMS || !nhLiveSameQuery(q, c.Result.Question) {
			return false
		}
		b, _ := json.Marshal(q.Where)
		var where psMap
		_ = json.Unmarshal(b, &where)
		terms := psArr(where["all"])
		if len(terms) != len(primary) {
			return false
		}
		for _, v := range terms {
			p := psObj(v)
			field := psObj(p["field"])
			k, _ := field["key"].(string)
			if field["kind"] != "resource" || p["op"] != "eq" || primary[k] != p["value"] {
				return false
			}
		}
		expected := []psMap{}
		for _, v := range psArr(f["records"]) {
			p := psObj(v)
			if p["signal"] != c.Signal {
				continue
			}
			match := true
			for k, v := range primary {
				if psObj(p["resource"])[k] != v {
					match = false
				}
			}
			if match {
				expected = append(expected, p)
			}
		}
		if c.Result.Count != strconv.Itoa(len(expected)) || len(c.Result.Rows) != len(expected) {
			return false
		}
		for n, raw := range c.Result.Rows {
			var row psMap
			b, _ := json.Marshal(raw)
			_ = json.Unmarshal(b, &row)
			p := expected[n]
			if !psEq(row["ResourceAttributes"], p["resource"]) || row["TraceId"] != p["traceId"] || row["timestamp_ns"] != p["timeNS"] || row["ServiceName"] != psObj(p["resource"])["service.name"] {
				return false
			}
		}
	}
	src := []int{0, 1, 3, 3, 4, 4, 4, 5, 5, 5, 6, 7, 7}
	if len(d.Events) != len(src) {
		return false
	}
	for n, e := range d.Events {
		if e.Source != src[n] {
			return false
		}
	}
	return reflect.DeepEqual(d.Saved, d.Reopened) && hrRaw(f, &d.Individual)
}
func TestHostResourceObserved(t *testing.T) {
	path := os.Getenv("GOO_HOST_RESOURCE_LEDGER")
	if path == "" {
		t.Skip("set GOO_HOST_RESOURCE_LEDGER")
	}
	raw, e := os.ReadFile(path)
	if e != nil {
		t.Fatal(e)
	}
	var d hrLive
	if json.Unmarshal(raw, &d) != nil || !hrLiveOK(d) {
		t.Fatal("raw host identity/source/context/effect mapping")
	}
	a, e := os.ReadFile(filepath.Join(filepath.Dir(path), d.StoredBefore.File))
	if e != nil {
		t.Fatal(e)
	}
	z, e := os.ReadFile(filepath.Join(filepath.Dir(path), d.StoredAfter.File))
	h := sha256.Sum256(a)
	var s exploration.SavedCapture
	if e != nil || string(a) != string(z) || len(a) != d.StoredBefore.Bytes || hex.EncodeToString(h[:]) != d.StoredBefore.SHA256 || json.Unmarshal(a, &s) != nil || s.ID != d.Saved.Receipt.ID || !reflect.DeepEqual(s.CaptureReceipt, d.Saved.Receipt) || !peLiveCaptureHash(exploration.CapturePage{Receipt: s.CaptureReceipt, Measurement: s.Measurement}) || !hrRaw(d.Facts, s.Measurement) || !reflect.DeepEqual(s.Measurement.Question, d.Individual.Question) {
		t.Fatal("complete immutable host capture")
	}
	controls := []func(*hrLive){
		func(x *hrLive) { x.Results[0].Basis = "changed" }, func(x *hrLive) { x.Results[0].ReadAt = "changed" },
		func(x *hrLive) { x.Results[0].Question.Definition.Source.Resource.ID = "wrong" }, func(x *hrLive) {
			x.Results[0].Question.Definition.Source.ExactStream.Resource["cloud.account.id"] = "wrong"
		}, func(x *hrLive) { x.Results[0].Question.Definition.Source.ExactStream.Resource["service.name"] = "" }, func(x *hrLive) { x.Results[0].Question.Definition.Source.ExactStream.ResourceSchemaURL = "wrong" }, func(x *hrLive) { x.Results[0].Rows = nil }, func(x *hrLive) { x.Results[0].Question.Definition.Source.Services = []string{"fake"} }, func(x *hrLive) { x.Events[3].Source++ }, func(x *hrLive) { x.Reopened.Receipt.ID = "wrong" }}
	for n, change := range controls {
		var x hrLive
		_ = json.Unmarshal(raw, &x)
		change(&x)
		if reflect.DeepEqual(x, d) || hrLiveOK(x) {
			t.Fatalf("changed host control %d accepted", n)
		}
	}
}
func TestHostResourceOwnershipObserved(t *testing.T) {
	path := os.Getenv("GOO_HOST_RESOURCE_OWNER_LEDGER")
	if path == "" {
		t.Skip("set GOO_HOST_RESOURCE_OWNER_LEDGER")
	}
	raw, e := os.ReadFile(path)
	if e != nil {
		t.Fatal(e)
	}
	var d psMap
	_ = json.Unmarshal(raw, &d)
	check := func(x psMap) bool {
		a, z := psObj(x["before"]), psObj(x["after"])
		held := psObj(x["held"])
		var r measurements.Result
		b, _ := json.Marshal(held["response"])
		_ = json.Unmarshal(b, &r)
		return psEq(a, z) && a["source"] == float64(3) && held["status"] == float64(200) && hrRaw(psObj(x["facts"]), &r)
	}
	if !check(d) {
		t.Fatal("actual held host result/stale owner")
	}
	for _, k := range []string{"url", "source", "widgets"} {
		var x psMap
		_ = json.Unmarshal(raw, &x)
		psObj(x["after"])[k] = "changed"
		if check(x) {
			t.Fatal("owner changed control", k)
		}
	}
}
