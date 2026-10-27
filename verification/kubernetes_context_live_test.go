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
	"sort"
	"strconv"
	"testing"
	"time"
)

var workloadKinds = []string{"deployment", "replicaset", "statefulset", "daemonset", "job", "cronjob"}

func workloadKey(r map[string]string, kind string) string {
	state := func(k string) any {
		v, ok := r[k]
		if !ok {
			return map[string]string{"state": "missing"}
		}
		s := "present"
		if v == "" {
			s = "empty"
		}
		return map[string]string{"state": s, "value": v}
	}
	_, name := r["k8s."+kind+".name"]
	_, uid := r["k8s."+kind+".uid"]
	if !name && !uid {
		return ""
	}
	b, _ := json.Marshal([]any{kind, state("k8s.cluster.uid"), state("k8s.cluster.name"), state("k8s.namespace.name"), state("k8s." + kind + ".uid"), state("k8s." + kind + ".name")})
	return string(b)
}
func workloadRows(f psMap, primary map[string]string, origin exploration.PodDiscoveryQuestion, question exploration.Query, result exploration.Result) bool {
	if _, e := time.Parse(time.RFC3339Nano, result.CapturedAt); e != nil {
		return false
	}
	if b, e := hex.DecodeString(result.Basis); e != nil || len(b) != 32 {
		return false
	}
	if origin.Environment != f["environment"] || !workloadScope(question, primary, "deployment", origin.Environment, origin.FromMS, origin.ToMS) || !nhLiveSameQuery(question, result.Question) {
		return false
	}
	expected := map[string]psMap{}
	for _, v := range psArr(f["records"]) {
		p := psObj(v)
		if p["family"] != question.Signal {
			continue
		}
		r := psObj(p["resource"])
		ok := true
		for _, k := range []string{"deployment.environment.name", "k8s.cluster.uid", "k8s.cluster.name", "k8s.namespace.name", "k8s.deployment.uid", "k8s.deployment.name"} {
			value, present := primary[k]
			actual, has := r[k]
			if present != has || (present && actual != value) {
				ok = false
			}
		}
		ns, e := strconv.ParseUint(p["timeNS"].(string), 10, 64)
		if e != nil || ns < uint64(question.FromMS)*1000000 || ns >= uint64(question.ToMS)*1000000 {
			ok = false
		}
		if ok {
			expected[p["id"].(string)] = p
		}
	}
	if result.Count != strconv.Itoa(len(expected)) || len(result.Rows) != len(expected) || len(expected) != 3 {
		return false
	}
	for _, row := range result.Rows {
		raw, _ := json.Marshal(row)
		var value psMap
		if json.Unmarshal(raw, &value) != nil {
			return false
		}
		id, _ := psObj(value["Attributes"])["work.id"].(string)
		p, ok := expected[id]
		if !ok || !psEq(value["ResourceAttributes"], p["resource"]) || !psEq(value["Attributes"], p["attributes"]) || value["TraceId"] != p["traceId"] || value["timestamp_ns"] != p["timeNS"] {
			return false
		}
		delete(expected, id)
	}
	return len(expected) == 0
}

// The retained OTLP payload establishes omitted scalar Flags=0. Normalize a
// detached observer copy only; source rows and saved evidence remain untouched.
func workloadRaw(f psMap, r *measurements.Result) bool {
	points := psArr(f["points"])
	wire := psArr(psObj(psObj(f["rawOTLP"])["metrics"])["resourceMetrics"])
	if len(points) != len(wire) {
		return false
	}
	attributes := func(v any) (psMap, bool) {
		out := psMap{}
		for _, v := range psArr(v) {
			a := psObj(v)
			k, ok := a["key"].(string)
			if !ok {
				return nil, false
			}
			value, ok := psObj(a["value"])["stringValue"].(string)
			if !ok {
				return nil, false
			}
			if _, duplicate := out[k]; duplicate {
				return nil, false
			}
			out[k] = value
		}
		return out, true
	}
	normalized := make([]any, len(points))
	for n, v := range points {
		p := psObj(v)
		w := psObj(wire[n])
		resource, ok := attributes(psObj(w["resource"])["attributes"])
		if !ok || !psEq(resource, p["resource"]) || w["schemaUrl"] != p["resourceSchemaURL"] {
			return false
		}
		scopes := psArr(w["scopeMetrics"])
		if len(scopes) != 1 {
			return false
		}
		scope := psObj(scopes[0])
		if psObj(scope["scope"])["name"] != p["scope"] || psObj(scope["scope"])["version"] != p["version"] || scope["schemaUrl"] != p["schemaURL"] {
			return false
		}
		metrics := psArr(scope["metrics"])
		if len(metrics) != 1 {
			return false
		}
		metric := psObj(metrics[0])
		if metric["name"] != p["metric"] || metric["unit"] != p["unit"] {
			return false
		}
		samples := psArr(psObj(metric["gauge"])["dataPoints"])
		if len(samples) != 1 {
			return false
		}
		sample := psObj(samples[0])
		attrs, ok := attributes(sample["attributes"])
		if !ok || !psEq(attrs, p["attributes"]) || sample["startTimeUnixNano"] != p["startNS"] || sample["timeUnixNano"] != p["endNS"] || sample["asInt"] != p["value"] {
			return false
		}
		flags, exists := sample["flags"]
		if !exists {
			flags = float64(0)
		}
		copy := psMap{}
		for k, v := range p {
			copy[k] = v
		}
		if existing, has := copy["flags"]; has && existing != flags {
			return false
		}
		copy["flags"] = flags
		normalized[n] = copy
	}
	detached := psMap{}
	for k, v := range f {
		detached[k] = v
	}
	detached["points"] = normalized
	return crRaw(detached, r)
}

type workloadLive struct {
	pdLive
	Key            string
	NoWorkloadKeys []string
	FacetCases     []struct {
		Case, Kind, Key string
		ExpectedKeys    []string
	}
	Metadata                  measurements.PodMetricDiscovery
	Planned                   measurements.PodDashboardPlan
	Individual                measurements.Result
	Saved, Reopened           exploration.CapturePage
	StoredBefore, StoredAfter struct {
		File, SHA256 string
		Bytes        int
	}
	Blocked []string
}

func workloadLiveOK(d workloadLive) bool {
	if !pdOracle(pdObservedFacts(d.Facts, d.Initial.Request), &d.Initial.Response) || d.Key != workloadKey(d.Primary.Resource, "deployment") || len(d.Context) != 2 || len(d.Blocked) > 0 || !reflect.DeepEqual(d.Saved, d.Reopened) {
		return false
	}
	selectKeys := func(kind, key string) []string {
		out := []string{}
		for _, r := range d.Initial.Response.Observations {
			if workloadKey(r.Resource, kind) == key {
				out = append(out, r.Key)
			}
		}
		sort.Strings(out)
		return out
	}
	sorted := func(keys []string) []string { out := append([]string{}, keys...); sort.Strings(out); return out }
	if !reflect.DeepEqual(selectKeys("deployment", d.Key), sorted(d.FilteredKeys)) {
		return false
	}
	absent := []string{}
	for _, r := range d.Initial.Response.Observations {
		found := false
		for _, kind := range workloadKinds {
			found = found || workloadKey(r.Resource, kind) != ""
		}
		if !found {
			absent = append(absent, r.Key)
		}
	}
	if !reflect.DeepEqual(sorted(absent), sorted(d.NoWorkloadKeys)) {
		return false
	}
	if len(d.FacetCases) != 8 {
		return false
	}
	for _, c := range d.FacetCases {
		found := false
		for _, v := range psArr(d.Facts["points"]) {
			p := psObj(v)
			if p["case"] == c.Case {
				var r map[string]string
				b, _ := json.Marshal(p["resource"])
				if json.Unmarshal(b, &r) != nil {
					return false
				}
				found = workloadKey(r, c.Kind) == c.Key
				break
			}
		}
		if !found || !reflect.DeepEqual(selectKeys(c.Kind, c.Key), sorted(c.ExpectedKeys)) {
			return false
		}
	}
	network := func(endpoint string, want any) bool {
		for _, r := range d.Responses {
			if r.Endpoint == endpoint && r.Status == 200 {
				var actual any
				if json.Unmarshal(r.Body, &actual) == nil && aeJSONEqual(actual, want) {
					return true
				}
			}
		}
		return false
	}
	signals := map[string]bool{}
	for _, c := range d.Context {
		if (c.Signal != "logs" && c.Signal != "spans") || signals[c.Signal] {
			return false
		}
		signals[c.Signal] = true
		if c.Question.Signal != c.Signal || !workloadRows(d.Facts, d.Primary.Resource, d.Initial.Request, c.Question, c.Result) || !network("/api/explore", c.Result) {
			return false
		}
	}
	if !reflect.DeepEqual(d.Metadata.Question, d.Planned.Question) || !reflect.DeepEqual(d.Metadata.Question.Resource, d.Primary.Resource) || !reflect.DeepEqual(d.Metadata.Question.ResourceSchemaURL, d.Primary.ResourceSchemaURL) || len(d.Planned.Widgets) != 1 || !aeJSONEqual(d.Planned.Widgets[0].Question, d.Individual.Question) {
		return false
	}
	if !workloadRaw(d.Facts, &d.Individual) || !network("/api/measurements", d.Individual) || !network("/api/pod-metric-discovery", d.Metadata) || !network("/api/pod-dashboard/plan", d.Planned) || len(d.Planned.Widgets) != 1 || !reflect.DeepEqual(d.Primary.Resource, d.Individual.Question.Definition.Source.ExactStream.Resource) {
		return false
	}
	reads := []int{1, 1, 1, 2, 2, 2, 3, 3, 5, 6, 7, 7}
	plans := []int{0, 0, 0, 0, 0, 0, 0, 0, 1, 2, 2, 2}
	actions := []string{"inventory", "local-workload-facet", "logs-staged", "logs-run", "logs-retained-back", "spans-staged", "spans-run", "spans-retained-back", "exact-container-identity", "native-run", "individual-capture", "frozen-reopen"}
	if len(d.Events) != len(reads) {
		return false
	}
	for n, e := range d.Events {
		if e.Action != actions[n] || e.Source != reads[n] || e.Plans != plans[n] {
			return false
		}
	}
	source, pure := 0, 0
	for _, q := range d.Requests {
		if q.Effect == "source" {
			source++
		}
		if q.Effect == "plan" {
			pure++
		}
	}
	return source == 7 && pure == 2
}
func workloadEnvelope(raw []byte) bool {
	var d psMap
	if json.Unmarshal(raw, &d) != nil {
		return false
	}
	for _, v := range psArr(d["context"]) {
		q := psObj(psObj(v)["question"])
		for k := range q {
			switch k {
			case "groupBy":
				if a, ok := q[k].([]any); !ok || len(a) != 0 {
					return false
				}
			case "signal", "environment", "fromMs", "toMs", "where", "limit", "offset":
			default:
				return false
			}
		}
	}
	return true
}
func TestKubernetesWorkloadObserved(t *testing.T) {
	path := os.Getenv("GOO_WORKLOAD_LEDGER")
	if path == "" {
		t.Skip("set GOO_WORKLOAD_LEDGER to actual205 evidence")
	}
	raw, e := os.ReadFile(path)
	if e != nil {
		t.Fatal(e)
	}
	var d workloadLive
	if e = json.Unmarshal(raw, &d); e != nil {
		t.Fatal(e)
	}
	if !workloadEnvelope(raw) || !workloadLiveOK(d) {
		t.Fatal("workload qualifier/raw membership/facet/source oracle")
	}
	a, e := os.ReadFile(filepath.Join(filepath.Dir(path), d.StoredBefore.File))
	if e != nil {
		t.Fatal(e)
	}
	z, e := os.ReadFile(filepath.Join(filepath.Dir(path), d.StoredAfter.File))
	sum := sha256.Sum256(a)
	var stored exploration.SavedCapture
	if e != nil || string(a) != string(z) || len(a) != d.StoredBefore.Bytes || hex.EncodeToString(sum[:]) != d.StoredBefore.SHA256 || json.Unmarshal(a, &stored) != nil || stored.ID != d.Saved.Receipt.ID || !reflect.DeepEqual(stored.CaptureReceipt, d.Saved.Receipt) || !peLiveCaptureHash(exploration.CapturePage{Receipt: stored.CaptureReceipt, Measurement: stored.Measurement}) || stored.Measurement == nil || !workloadRaw(d.Facts, stored.Measurement) || !aeJSONEqual(stored.Measurement.Question, d.Individual.Question) {
		t.Fatal("complete individual readiness custody")
	}
	controls := []func(*workloadLive){func(x *workloadLive) {
		var row psMap
		_ = json.Unmarshal(x.Individual.Rows[0], &row)
		row["flags"] = float64(1)
		x.Individual.Rows[0], _ = json.Marshal(row)
	}, func(x *workloadLive) {
		var row psMap
		_ = json.Unmarshal(x.Individual.Rows[0], &row)
		delete(row, "flags")
		x.Individual.Rows[0], _ = json.Marshal(row)
	}, func(x *workloadLive) { x.Context[1] = x.Context[0] }, func(x *workloadLive) {
		x.Context[0].Question.Where.Any = []exploration.Predicate{{Op: "exists", Field: &exploration.Field{Kind: "resource", Key: "container.id"}}}
	}, func(x *workloadLive) {
		x.Context[0].Question.Where.Field = &exploration.Field{Kind: "resource", Key: "container.id"}
	}, func(x *workloadLive) {
		x.Context[0].Question.Where.Numeric = &exploration.NumericBinding{Field: exploration.Field{Kind: "attribute", Key: "x"}, Type: "int64", Parse: "canonical-decimal", Unit: "items"}
	}, func(x *workloadLive) { x.Context[0].Question.Environment = "wrong" }, func(x *workloadLive) {
		for n, p := range x.Context[0].Question.Where.All {
			if p.Field.Key == "k8s.deployment.uid" {
				x.Context[0].Question.Where.All[n].Op = "missing"
				x.Context[0].Question.Where.All[n].Value = ""
			}
		}
	}, func(x *workloadLive) { x.Context[0].Result.Rows = x.Context[0].Result.Rows[:2] }, func(x *workloadLive) { x.Context[0].Result.Rows[1] = x.Context[0].Result.Rows[0] }, func(x *workloadLive) { x.FilteredKeys = nil }, func(x *workloadLive) { x.FacetCases[0].Key = "wrong" }, func(x *workloadLive) { x.Primary.Resource["k8s.cluster.uid"] = "wrong" }, func(x *workloadLive) { x.Individual.Question.Definition.Source.Resource.ID = "replacement" }, func(x *workloadLive) { x.Individual.ReadAt = "wrong" }, func(x *workloadLive) { x.Events[1].Source++ }, func(x *workloadLive) { x.Reopened.Receipt.ID = "wrong" }}
	for _, key := range []string{"fromNS", "toNS"} {
		var x psMap
		if json.Unmarshal(raw, &x) != nil {
			t.Fatal("wire control decode")
		}
		psObj(psObj(psArr(x["context"])[0])["question"])[key] = "1"
		b, e := json.Marshal(x)
		if e != nil || string(b) == string(raw) || workloadEnvelope(b) {
			t.Fatalf("wire scope control %s accepted", key)
		}
	}
	for n, change := range controls {
		var x workloadLive
		if json.Unmarshal(raw, &x) != nil {
			t.Fatal("control decode")
		}
		change(&x)
		if reflect.DeepEqual(x, d) || workloadLiveOK(x) {
			t.Fatalf("workload changed control %d accepted", n)
		}
	}
}
func TestKubernetesWorkloadOwnershipObserved(t *testing.T) {
	path := os.Getenv("GOO_WORKLOAD_OWNER_LEDGER")
	if path == "" {
		t.Skip("set GOO_WORKLOAD_OWNER_LEDGER")
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
		held := psObj(x["held"])
		before, after := psObj(x["beforeRelease"]), psObj(x["afterRelease"])
		if held["status"] != float64(200) || !psEq(before, after) || before["focus"] != true || before["source"] != float64(2) {
			return false
		}
		var q exploration.Query
		var r exploration.Result
		var primary exploration.PodResourceObservation
		b, _ := json.Marshal(held["request"])
		if json.Unmarshal(b, &q) != nil {
			return false
		}
		b, _ = json.Marshal(held["body"])
		if json.Unmarshal(b, &r) != nil {
			return false
		}
		b, _ = json.Marshal(x["primary"])
		if json.Unmarshal(b, &primary) != nil {
			return false
		}
		var origin exploration.PodDiscoveryQuestion
		b, _ = json.Marshal(psObj(x["initial"])["request"])
		if json.Unmarshal(b, &origin) != nil {
			return false
		}
		return q.Signal == "logs" && workloadRows(psObj(x["facts"]), primary.Resource, origin, q, r)
	}
	if !check(d) {
		t.Fatal("held actual workload response/retained owner")
	}
	for _, key := range []string{"source", "url", "focus", "facet"} {
		var x psMap
		if json.Unmarshal(raw, &x) != nil {
			t.Fatal("control decode")
		}
		after := psObj(x["afterRelease"])
		if key == "source" {
			after[key] = float64(3)
		} else if key == "focus" {
			after[key] = false
		} else {
			after[key] = "changed"
		}
		if psEq(x, d) || check(x) {
			t.Fatalf("owner control %s accepted", key)
		}
	}
}
