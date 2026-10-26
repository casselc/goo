package verification

import (
	"encoding/json"
	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/measurements"
	"os"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"testing"
)

type pdLive struct {
	Facts   psMap
	Initial struct {
		Request  exploration.PodDiscoveryQuestion
		Response exploration.PodDiscovery
	}
	Primary      exploration.PodResourceObservation
	FilteredKeys []string
	Context      []struct {
		Signal          string
		Question        exploration.Query
		Result          exploration.Result
		ExpectedMembers []string
	}
	Events []struct {
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

func pdObservedFacts(f psMap, q exploration.PodDiscoveryQuestion) []pdFact {
	out := []pdFact{}
	lo, hi := uint64(q.FromMS)*1000000, uint64(q.ToMS)*1000000
	for _, group := range []string{"records", "points"} {
		for _, v := range psArr(f[group]) {
			p := psObj(v)
			res := psObj(p["resource"])
			if res["deployment.environment.name"] != q.Environment {
				continue
			}
			if _, uid := res["k8s.pod.uid"]; !uid {
				if _, name := res["k8s.pod.name"]; !name {
					continue
				}
			}
			clock := "endNS"
			if group == "records" {
				clock = "timeNS"
			}
			at, _ := p[clock].(string)
			n, e := strconv.ParseUint(at, 10, 64)
			if e != nil || n < lo || n >= hi {
				continue
			}
			labels := map[string]string{}
			valid := true
			for k, v := range res {
				s, ok := v.(string)
				if !ok {
					valid = false
				}
				labels[k] = s
			}
			if !valid {
				continue
			}
			family, _ := p["family"].(string)
			var schema *string
			if family != "spans" {
				s, _ := p["resourceSchemaURL"].(string)
				schema = &s
			}
			out = append(out, pdFact{Resource: labels, Schema: schema, Family: family, At: at})
		}
	}
	return out
}
func pdLiveOK(d pdLive) bool {
	facts := pdObservedFacts(d.Facts, d.Initial.Request)
	r := d.Initial.Response
	if r.Version != "goo/pod-discovery/v1" || r.Question != d.Initial.Request || r.SourceGroups != len(pdRows(facts)) || r.ObservationsOmitted != 0 || !pdOracle(facts, &r) {
		return false
	}
	observed := false
	for _, response := range d.Responses {
		if response.Endpoint == "/api/pod-discovery" && response.Status == 200 {
			var actual exploration.PodDiscovery
			if json.Unmarshal(response.Body, &actual) == nil && aeJSONEqual(actual, r) {
				observed = true
			}
		}
	}
	if !observed {
		return false
	}
	primaryFound := false
	filtered := []string{}
	for _, o := range r.Observations {
		if reflect.DeepEqual(o, d.Primary) {
			primaryFound = true
		}
		_, cluster := o.Resource["k8s.cluster.uid"]
		_, namespace := o.Resource["k8s.namespace.name"]
		search := false
		for _, v := range o.Resource {
			if strings.Contains(strings.ToLower(v), "uid-a") {
				search = true
			}
		}
		if !cluster && !namespace && search {
			filtered = append(filtered, o.Key)
		}
	}
	sort.Strings(filtered)
	keys := append([]string{}, d.FilteredKeys...)
	sort.Strings(keys)
	if !primaryFound || !d.Primary.Actionable || !reflect.DeepEqual(filtered, keys) || len(d.Context) != 2 {
		return false
	}
	for _, c := range d.Context {
		q := c.Question
		if q.Signal != c.Signal || q.Environment != d.Initial.Request.Environment || q.FromMS != d.Initial.Request.FromMS || q.ToMS != d.Initial.Request.ToMS || !nhLiveSameQuery(q, c.Result.Question) {
			return false
		}
		encoded, _ := json.Marshal(q.Where)
		var predicate psMap
		_ = json.Unmarshal(encoded, &predicate)
		terms := psArr(predicate["all"])
		if len(terms) != len(d.Primary.Resource)+2 {
			return false
		}
		seen := map[string]bool{}
		for _, v := range terms {
			p := psObj(v)
			field := psObj(p["field"])
			key, _ := field["key"].(string)
			if field["kind"] != "resource" || seen[key] {
				return false
			}
			seen[key] = true
			if key == "k8s.cluster.uid" || key == "k8s.namespace.name" {
				if p["op"] != "missing" {
					return false
				}
			} else {
				value, ok := d.Primary.Resource[key]
				if !ok || p["op"] != "eq" || p["value"] != value {
					return false
				}
			}
		}
		expected := map[string]psMap{}
		for _, v := range psArr(d.Facts["records"]) {
			p := psObj(v)
			if p["family"] != c.Signal {
				continue
			}
			res := psObj(p["resource"])
			match := true
			for k, v := range d.Primary.Resource {
				if res[k] != v {
					match = false
				}
			}
			if _, ok := res["k8s.cluster.uid"]; ok {
				match = false
			}
			if _, ok := res["k8s.namespace.name"]; ok {
				match = false
			}
			if match {
				expected[p["id"].(string)] = p
			}
		}
		if c.Result.Count != strconv.Itoa(len(expected)) || len(c.Result.Rows) != len(expected) {
			return false
		}
		ids := []string{}
		for _, raw := range c.Result.Rows {
			encoded, _ := json.Marshal(raw)
			var row psMap
			_ = json.Unmarshal(encoded, &row)
			id, _ := psObj(row["Attributes"])["work.id"].(string)
			p, ok := expected[id]
			if !ok || !psEq(row["ResourceAttributes"], p["resource"]) || row["TraceId"] != p["traceId"] || row["timestamp_ns"] != p["timeNS"] || !psEq(row["Attributes"], p["attributes"]) {
				return false
			}
			ids = append(ids, id)
			delete(expected, id)
		}
		if len(expected) != 0 {
			return false
		}
		sort.Strings(ids)
		wanted := append([]string{}, c.ExpectedMembers...)
		sort.Strings(wanted)
		if !reflect.DeepEqual(ids, wanted) {
			return false
		}
	}
	source := []int{1, 1, 1, 2, 2, 2, 3, 3}
	if len(d.Events) != len(source) {
		return false
	}
	for n, e := range d.Events {
		if e.Source != source[n] || e.Plans != 0 {
			return false
		}
	}
	reads := 0
	for _, q := range d.Requests {
		if q.Effect == "source" {
			reads++
		}
	}
	return reads == 3
}
func TestPodDiscoveryObserved(t *testing.T) {
	path := os.Getenv("GOO_POD_DISCOVERY_LEDGER")
	if path == "" {
		t.Skip("set GOO_POD_DISCOVERY_LEDGER")
	}
	raw, e := os.ReadFile(path)
	if e != nil {
		t.Fatal(e)
	}
	var d pdLive
	if json.Unmarshal(raw, &d) != nil || !pdLiveOK(d) {
		t.Fatal("ordinary raw five-family/variant/schema/context/source oracle")
	}
	controls := []func(*pdLive){func(x *pdLive) { x.Initial.Response.Observations[0].Families[0].Count = "999" }, func(x *pdLive) { x.Initial.Response.Observations[0].Resource["service.name"] = "" }, func(x *pdLive) {
		s := ""
		for n := range x.Initial.Response.Observations {
			if x.Initial.Response.Observations[n].ResourceSchemaURL == nil {
				x.Initial.Response.Observations[n].ResourceSchemaURL = &s
				return
			}
		}
	}, func(x *pdLive) { x.Initial.Response.Basis = "changed" }, func(x *pdLive) { x.Initial.Response.ReadAt = "changed" }, func(x *pdLive) { x.Events[1].Source++ }, func(x *pdLive) { x.Primary.Actionable = false }, func(x *pdLive) { x.Context[0].ExpectedMembers = nil }, func(x *pdLive) { x.Initial.Response.SourceGroups++ }}
	for n, change := range controls {
		var x pdLive
		_ = json.Unmarshal(raw, &x)
		change(&x)
		if reflect.DeepEqual(x, d) || pdLiveOK(x) {
			t.Fatalf("pod changed control %d accepted", n)
		}
	}
}
func TestPodDiscoveryOwnershipObserved(t *testing.T) {
	path := os.Getenv("GOO_POD_DISCOVERY_OWNER_LEDGER")
	if path == "" {
		t.Skip("set GOO_POD_DISCOVERY_OWNER_LEDGER")
	}
	raw, e := os.ReadFile(path)
	if e != nil {
		t.Fatal(e)
	}
	var d psMap
	_ = json.Unmarshal(raw, &d)
	check := func(x psMap) bool {
		cases := psArr(x["cases"])
		if len(cases) != 2 {
			return false
		}
		for n, v := range cases {
			c := psObj(v)
			held := psObj(c["held"])
			var q exploration.PodDiscoveryQuestion
			var r exploration.PodDiscovery
			b, _ := json.Marshal(held["request"])
			_ = json.Unmarshal(b, &q)
			b, _ = json.Marshal(held["response"])
			_ = json.Unmarshal(b, &r)
			after := psObj(c["after"])
			if held["status"] != float64(200) || !psEq(c["before"], after) || after["rows"] != float64(0) || after["source"] != float64(n+1) || r.Question != q || !pdOracle(pdObservedFacts(psObj(x["facts"]), q), &r) {
				return false
			}
		}
		return true
	}
	if !check(d) {
		t.Fatal("held raw pod discovery published after edit/departure")
	}
	for _, k := range []string{"source", "rows", "url"} {
		var x psMap
		_ = json.Unmarshal(raw, &x)
		psObj(psObj(psArr(x["cases"])[0])["after"])[k] = "changed"
		if check(x) {
			t.Fatal("owner changed control", k)
		}
	}
}

// The supporting and runtime responses below are genuine held upstream bodies;
// no positive substitute is supplied after withdrawing the service owner.
func TestPodServicesDepartureObserved(t *testing.T) {
	path := os.Getenv("GOO_POD_SERVICES_OWNER_LEDGER")
	if path == "" {
		t.Skip("set GOO_POD_SERVICES_OWNER_LEDGER")
	}
	raw, e := os.ReadFile(path)
	if e != nil {
		t.Fatal(e)
	}
	check := func(x psMap) bool {
		cases := psArr(x["cases"])
		if len(cases) != 2 {
			return false
		}
		facts := psObj(x["facts"])
		for _, v := range cases {
			c := psObj(v)
			held := psObj(c["held"])
			before, after := psObj(c["beforeRelease"]), psObj(c["afterRelease"])
			if held["status"] != float64(200) || !psEq(before, after) || after["detailHidden"] != true || after["runtimeCards"] != float64(0) {
				return false
			}
			body, _ := json.Marshal(held["response"])
			if c["phase"] == "supporting" {
				var r exploration.Result
				if json.Unmarshal(body, &r) != nil {
					return false
				}
				var f struct{ Records []hdLiveRecord }
				b, _ := json.Marshal(facts)
				_ = json.Unmarshal(b, &f)
				if hdLiveResult(f.Records, r) != nil {
					return false
				}
			} else if c["phase"] == "widget" {
				var r measurements.Result
				if json.Unmarshal(body, &r) != nil || !grRaw(facts, &r) {
					return false
				}
			} else {
				return false
			}
		}
		applied := psObj(x["applied"])
		var r measurements.Result
		body, _ := json.Marshal(applied["body"])
		if applied["status"] != float64(200) || json.Unmarshal(body, &r) != nil || !grRaw(facts, &r) || r.Question.Definition.Source.Metric != "go.goroutine.count" {
			return false
		}
		after := psObj(x["afterApply"])
		if after["detailHidden"] != false || after["runtimeCards"] != float64(1) || after["dashboardState"] != "loaded" || after["source"] != x["beforeTabs"].(float64)+2 {
			return false
		}
		source, plans := 0, 0
		for _, v := range psArr(x["requests"]) {
			q := psObj(v)
			if q["effect"] == "source" {
				source++
			}
			if q["effect"] == "plan" {
				plans++
			}
		}
		return source == int(after["source"].(float64)) && plans == int(after["plans"].(float64))
	}
	var d psMap
	_ = json.Unmarshal(raw, &d)
	if !check(d) {
		t.Fatal("service support/widget withdrawal, retained recovery or explicit Apply raw oracle")
	}
	for _, key := range []string{"source", "plans", "runtimeCards", "detailHidden"} {
		var x psMap
		_ = json.Unmarshal(raw, &x)
		psObj(psObj(psArr(x["cases"])[0])["afterRelease"])[key] = "changed"
		if check(x) {
			t.Fatal("Services changed owner control", key)
		}
	}
}
