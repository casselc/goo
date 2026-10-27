package verification

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/measurements"
	"github.com/casselc/goo/metrics"
	"math"
	"math/big"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"testing"
)

func cpRole(s metrics.Stream) string {
	if s.Unit != "{cpu}" || s.Version != "0.137.0" || s.SchemaURL != "" {
		return ""
	}
	if s.Name == "container.cpu.usage" && s.Scope == "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/kubeletstatsreceiver" {
		return "usage"
	}
	if s.Scope != "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/k8sclusterreceiver" {
		return ""
	}
	if s.Name == "k8s.container.cpu_request" {
		return "request"
	}
	if s.Name == "k8s.container.cpu_limit" {
		return "limit"
	}
	return ""
}

type cpFact struct {
	cmFact
	cores      *big.Rat
	normalized string
}

func cpRawGauge(f psMap, r *measurements.Result) (cpFact, bool) {
	base, ok := rawContainerGauge(f, r, cpRole)
	out := cpFact{cmFact: base}
	if !ok {
		return out, false
	}
	if base.latest == nil || base.latest.Value == nil {
		return out, true
	}
	for _, raw := range r.Rows {
		var row psMap
		if json.Unmarshal(raw, &row) != nil {
			return out, false
		}
		if row["end_ns"] != base.latest.AtNS {
			continue
		}
		text, ok := row["value"].(string)
		if !ok {
			return out, false
		}
		value, normalized, valid := cpDecimal(text)
		if !valid || (out.cores != nil && out.cores.Cmp(value) != 0) {
			out.cores = nil
			out.normalized = ""
			return out, true
		}
		out.cores, out.normalized = value, normalized
	}
	return out, true
}

type cpPair struct {
	Request  measurements.ContainerCPUViewRequest
	Response measurements.ContainerCPUView
}

func cpPairOK(f psMap, p cpPair, origin exploration.PodDiscoveryQuestion) bool {
	q, v := p.Request, p.Response
	if q.Parent.ResourceSchemaURL == nil || q.Parent.Environment == "" || q.Parent.Resource["deployment.environment.name"] != q.Parent.Environment || v.Version != "goo/container-cpu-view/v1" || !reflect.DeepEqual(v.Parent, q.Parent) || q.Parent.Environment != origin.Environment || q.Parent.FromMS != origin.FromMS || q.Parent.ToMS != origin.ToMS || len(v.Signals) != 3 || !cmContext(q.Parent.Resource, q.Parent.Resource) {
		return false
	}
	facts := map[string]cpFact{}
	results := map[string]*measurements.Result{"usage": q.Usage, "request": q.Request, "limit": q.Limit}
	for n, role := range []string{"usage", "request", "limit"} {
		signal := v.Signals[n]
		if signal.Role != role {
			return false
		}
		r := results[role]
		if r == nil {
			if signal.State != "unavailable" || signal.Reason != "not-supplied" || signal.Value != "" || signal.AtNS != "" || signal.Question != nil || signal.ReadAt != "" || signal.Basis != "" || signal.ObservationState != "" {
				return false
			}
			continue
		}
		fact, ok := cpRawGauge(f, r)
		if !ok || cpRole(fact.stream) != role || !cmContext(q.Parent.Resource, fact.stream.Resource) || r.Question.Definition.Environment != q.Parent.Environment || r.Question.FromMS != q.Parent.FromMS || r.Question.ToMS != q.Parent.ToMS || signal.Question == nil || !aeJSONEqual(*signal.Question, r.Question) || signal.ReadAt != r.ReadAt || signal.Basis != r.Basis {
			return false
		}
		facts[role] = fact
		state, reason, at, observation, value := "no_data", "no-observations", "", "", ""
		if fact.latest != nil {
			state, reason, at, observation = fact.latest.State, fact.latest.Reason, fact.latest.AtNS, fact.latest.State
			if fact.latest.Value != nil {
				if fact.cores == nil {
					state, reason = "invalid", "invalid-core-domain"
				} else {
					value = fact.normalized
				}
			}
		}
		if signal.State != state || signal.Reason != reason || signal.AtNS != at || signal.ObservationState != observation || signal.Value != value {
			return false
		}
	}
	for _, role := range []string{"request", "limit"} {
		a, aok := facts["usage"]
		b, bok := facts[role]
		reason := ""
		var ratio *big.Rat
		switch {
		case !aok || !bok:
			reason = "input-not-supplied"
		case a.latest == nil || b.latest == nil:
			reason = "no-observations"
		case a.latest.State != "observed" || b.latest.State != "observed":
			reason = "observation-not-current"
		case a.cores == nil || b.cores == nil:
			reason = "invalid-core-domain"
		case !reflect.DeepEqual(a.stream.Resource, b.stream.Resource) || a.stream.ResourceSchemaURL != b.stream.ResourceSchemaURL || !psEq(a.stream.Attributes, b.stream.Attributes):
			reason = "resource-or-attribute-mismatch"
		case a.latest.AtNS != b.latest.AtNS:
			reason = "observation-time-mismatch"
		case results["usage"].Question.Definition.MaxGapMS != results[role].Question.Definition.MaxGapMS:
			reason = "freshness-policy-mismatch"
		case b.cores.Sign() == 0:
			reason = "zero-denominator"
		default:
			ratio = new(big.Rat).Mul(new(big.Rat).Quo(a.cores, b.cores), big.NewRat(100, 1))
		}
		if ratio != nil && len(ratio.RatString()) > 1024 {
			ratio = nil
			reason = "bounded-percentage-domain"
		}
		actual := v.Percentages.Request
		if role == "limit" {
			actual = v.Percentages.Limit
		}
		if ratio == nil {
			if actual.State != "unavailable" || actual.Reason != reason || actual.Value != "" || actual.Text != "" {
				return false
			}
		} else if actual.State != "observed" || actual.Reason != "" || actual.Value != ratio.RatString() || !cpTextOK(ratio, actual.Text) {
			return false
		}
	}
	return true
}

func cpSnapshot(v measurements.ContainerCPUView, snapshot psMap) bool {
	rows, percent := psArr(snapshot["signals"]), psArr(snapshot["percentages"])
	if len(rows) != 3 || len(percent) != 2 {
		return false
	}
	maximum := new(big.Rat)
	maxText := "0"
	for _, s := range v.Signals {
		if s.Value != "" {
			value, normalized, ok := cpDecimal(s.Value)
			if !ok || normalized != s.Value {
				return false
			}
			if value.Cmp(maximum) > 0 {
				maximum = value
				maxText = s.Value
			}
		}
	}
	for n, s := range v.Signals {
		row := psObj(rows[n])
		text, ok := row["text"].(string)
		if !ok {
			return false
		}
		quality := s.State
		if s.State == "invalid" && s.ObservationState != "" {
			quality += " · " + s.ObservationState
		}
		if row["role"] != s.Role || row["state"] != s.State || !strings.Contains(text, quality) {
			return false
		}
		expected, title := s.Value, s.Value+" {cpu} (exact exported value)"
		if s.Value == "" {
			expected, title = "unavailable", "Unavailable"
		}
		if row["title"] != title {
			return false
		}
		bars := psArr(row["bars"])
		if len(bars) != 1 {
			return false
		}
		bar := psObj(bars[0])
		if bar["value"] != expected || bar["scaleMax"] != maxText {
			return false
		}
		if s.Value != "" {
			value, _, _ := cpDecimal(s.Value)
			width := float64(0)
			if maximum.Sign() > 0 {
				scaled := new(big.Rat).Mul(new(big.Rat).Quo(value, maximum), big.NewRat(10000, 1))
				integer := new(big.Int).Quo(scaled.Num(), scaled.Denom())
				width = float64(integer.Int64()) / 100
			}
			fill, ok := bar["fill"].(string)
			if !ok || !strings.HasSuffix(fill, "%") {
				return false
			}
			actual, e := strconv.ParseFloat(strings.TrimSuffix(fill, "%"), 64)
			if e != nil || math.Abs(actual-width) > 0.00001 {
				return false
			}
		} else if bar["fill"] != nil {
			return false
		}
	}
	for n, p := range []measurements.ContainerCPUPercentage{v.Percentages.Request, v.Percentages.Limit} {
		role := []string{"request", "limit"}[n]
		row := psObj(percent[n])
		if p.State == "observed" {
			if row["text"] != "Usage / "+role+": "+p.Text || row["title"] != p.Value+"%" {
				return false
			}
		} else if row["title"] != "" || row["text"] != "Usage / "+role+" unavailable · "+strings.ReplaceAll(p.Reason, "-", " ") {
			return false
		}
	}
	return true
}

type cpLive struct {
	pdLive
	First, Second             pnRead
	Cluster                   exploration.PodResourceObservation
	Compatible                []string
	Independent               cpPair
	IndependentSnapshot       psMap
	Individual                measurements.Result
	Saved, Reopened           exploration.CapturePage
	StoredBefore, StoredAfter struct {
		File, SHA256 string
		Bytes        int
	}
	Boundary []struct {
		ID       string
		Row      exploration.PodResourceObservation
		Opened   pnRead
		Composed cpPair
		Snapshot psMap
	}
	Blocked []string
	Events  []struct {
		Action                      string
		Source, Plans, Compositions int
	}
}

func cpLiveOK(d cpLive) bool {
	if !pdOracle(pdObservedFacts(d.Facts, d.Initial.Request), &d.Initial.Response) || !ppMetadata(d.Facts, d.First) || !ppMetadata(d.Facts, d.Second) || len(d.First.Plan.Widgets) != 0 || len(d.Second.Plan.Widgets) != 0 || !reflect.DeepEqual(d.Saved, d.Reopened) || len(d.Blocked) != 0 {
		return false
	}
	if !reflect.DeepEqual(d.First.DiscoveryRequest.Resource, d.Primary.Resource) || !reflect.DeepEqual(d.First.DiscoveryRequest.ResourceSchemaURL, d.Primary.ResourceSchemaURL) || !reflect.DeepEqual(d.Second.DiscoveryRequest.Resource, d.Cluster.Resource) || !reflect.DeepEqual(d.Second.DiscoveryRequest.ResourceSchemaURL, d.Cluster.ResourceSchemaURL) {
		return false
	}
	for _, read := range []pnRead{d.First, d.Second} {
		q := read.DiscoveryRequest
		if q.Environment != d.Initial.Request.Environment || q.FromMS != d.Initial.Request.FromMS || q.ToMS != d.Initial.Request.ToMS {
			return false
		}
	}
	expected := []string{}
	primaryFound := false
	for _, row := range d.Initial.Response.Observations {
		if reflect.DeepEqual(row, d.Primary) {
			primaryFound = true
		}
		if row.ResourceSchemaURL != nil && cmContext(d.Primary.Resource, row.Resource) {
			expected = append(expected, row.Key)
		}
	}
	sort.Strings(expected)
	compatible := append([]string{}, d.Compatible...)
	sort.Strings(compatible)
	if !primaryFound || !reflect.DeepEqual(expected, compatible) {
		return false
	}
	network := func(endpoint string, want any) bool {
		for _, response := range d.Responses {
			if response.Endpoint != endpoint || response.Status != 200 {
				continue
			}
			var actual any
			if json.Unmarshal(response.Body, &actual) == nil && aeJSONEqual(actual, want) {
				return true
			}
		}
		return false
	}
	checkPair := func(p cpPair, snapshot psMap) bool {
		if !cpPairOK(d.Facts, p, d.Initial.Request) || !network("/api/container-cpu/view", p.Response) || !cpSnapshot(p.Response, snapshot) {
			return false
		}
		for _, r := range []*measurements.Result{p.Request.Usage, p.Request.Request, p.Request.Limit} {
			if r != nil && !network("/api/measurements", *r) {
				return false
			}
		}
		return true
	}
	if !reflect.DeepEqual(d.Independent.Request.Parent, d.First.DiscoveryRequest) {
		return false
	}
	if !checkPair(d.Independent, d.IndependentSnapshot) {
		return false
	}
	for _, p := range []measurements.ContainerCPUPercentage{d.Independent.Response.Percentages.Request, d.Independent.Response.Percentages.Limit} {
		if p.State != "unavailable" {
			return false
		}
	}
	for _, r := range []*measurements.Result{d.Independent.Request.Usage, d.Independent.Request.Request, d.Independent.Request.Limit} {
		if r == nil {
			return false
		}
		found := false
		for _, read := range []pnRead{d.First, d.Second} {
			for _, w := range read.Plan.Candidates {
				if aeJSONEqual(w.Question, r.Question) && w.Role == cpRole(*r.Question.Definition.Source.ExactStream) {
					found = true
				}
			}
		}
		if !found {
			return false
		}
	}
	if _, ok := cpRawGauge(d.Facts, &d.Individual); !ok || !network("/api/measurements", d.Individual) || !aeJSONEqual(d.Individual.Question, d.Independent.Request.Usage.Question) {
		return false
	}
	if len(d.Boundary) != 9 {
		return false
	}
	for _, b := range d.Boundary {
		if !reflect.DeepEqual(b.Composed.Request.Parent, b.Opened.DiscoveryRequest) || !ppMetadata(d.Facts, b.Opened) || len(b.Opened.Plan.Widgets) != 0 || !checkPair(b.Composed, b.Snapshot) || !reflect.DeepEqual(b.Row.Resource, b.Opened.DiscoveryRequest.Resource) {
			return false
		}
	}
	actions := []string{"inventory", "explicit-first-metadata-no-values", "explicit-second-metadata", "local-role-selection", "independent-three-signals", "native-staged", "retained-back", "native-run", "individual-capture", "capture-retained-return", "boundary-aligned", "boundary-over-100", "boundary-zero-denominator", "boundary-partial", "boundary-negative", "boundary-fractional", "boundary-stale", "boundary-conflict", "boundary-missing", "frozen-reopen"}
	sources := []int{1, 2, 3, 3, 6, 6, 6, 7, 8, 8, 12, 16, 20, 22, 26, 30, 34, 38, 42, 42}
	plans := []int{0, 1, 2, 2, 2, 3, 3, 4, 4, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 13}
	compositions := []int{0, 0, 0, 0, 1, 1, 1, 1, 1, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 10}
	if len(d.Events) != len(actions) {
		return false
	}
	for n, e := range d.Events {
		if e.Action != actions[n] || e.Source != sources[n] || e.Plans != plans[n] || e.Compositions != compositions[n] {
			return false
		}
	}
	source, pure, compose := 0, 0, 0
	endpointCounts := map[string]int{}
	for _, r := range d.Requests {
		switch r.Effect {
		case "source":
			source++
			endpointCounts[r.Endpoint]++
		case "plan":
			pure++
		case "composition":
			compose++
		}
	}
	return source == 42 && pure == 13 && compose == 10 && endpointCounts["/api/pod-discovery"] == 1 && endpointCounts["/api/pod-metric-discovery"] == 11 && endpointCounts["/api/measurements"] == 29 && endpointCounts["/api/captures"] == 1
}
func TestContainerCPUObserved(t *testing.T) {
	path := os.Getenv("GOO_CONTAINER_CPU_LEDGER")
	if path == "" {
		t.Skip("set GOO_CONTAINER_CPU_LEDGER to actual207 receipt")
	}
	raw, e := os.ReadFile(path)
	if e != nil {
		t.Fatal(e)
	}
	var d cpLive
	if e = json.Unmarshal(raw, &d); e != nil {
		t.Fatal(e)
	}
	if !cpLiveOK(d) {
		t.Fatal("raw container gauges/alignment/dom/source oracle")
	}
	a, e := os.ReadFile(filepath.Join(filepath.Dir(path), d.StoredBefore.File))
	if e != nil {
		t.Fatal(e)
	}
	z, e := os.ReadFile(filepath.Join(filepath.Dir(path), d.StoredAfter.File))
	sum := sha256.Sum256(a)
	var stored exploration.SavedCapture
	if e != nil || string(a) != string(z) || len(a) != d.StoredBefore.Bytes || hex.EncodeToString(sum[:]) != d.StoredBefore.SHA256 || json.Unmarshal(a, &stored) != nil || stored.ID != d.Saved.Receipt.ID || !reflect.DeepEqual(stored.CaptureReceipt, d.Saved.Receipt) || !peLiveCaptureHash(exploration.CapturePage{Receipt: stored.CaptureReceipt, Measurement: stored.Measurement}) || stored.Measurement == nil || !aeJSONEqual(stored.Measurement.Question, d.Individual.Question) {
		t.Fatal("complete immutable individual gauge custody")
	}
	if _, ok := cpRawGauge(d.Facts, stored.Measurement); !ok {
		t.Fatal("raw stored gauge replay")
	}
	controls := []func(*cpLive){func(x *cpLive) { x.Boundary[1].Composed.Response.Percentages.Request.Value = "100" }, func(x *cpLive) { x.Boundary[5].Composed.Response.Signals[0].Value = "0" }, func(x *cpLive) {
		x.Independent.Response.Percentages.Limit.State = "observed"
		x.Independent.Response.Percentages.Limit.Value = "25"
	}, func(x *cpLive) { x.Boundary[0].Composed.Response.Percentages.Request.Value = "51" }, func(x *cpLive) {
		x.Boundary[2].Composed.Response.Percentages.Request.State = "observed"
		x.Boundary[2].Composed.Response.Percentages.Request.Value = "0"
	}, func(x *cpLive) { x.Boundary[3].Composed.Response.Signals[1].Value = "0" }, func(x *cpLive) {
		x.Boundary[4].Composed.Response.Signals[0].State = "observed"
		x.Boundary[4].Composed.Response.Signals[0].Value = "0.125"
	}, func(x *cpLive) { x.Boundary[6].Composed.Response.Signals[0].State = "observed" }, func(x *cpLive) { x.Boundary[7].Composed.Request.Usage.Rows = nil }, func(x *cpLive) {
		x.Independent.Request.Usage.Question.Definition.Source.Resource.ID = "replacement"
	}, func(x *cpLive) {
		x.Independent.Request.Request.Question.Definition.Source.ExactStream.ResourceSchemaURL = "wrong"
	}, func(x *cpLive) { x.Individual.ReadAt = "wrong" }, func(x *cpLive) { x.Individual.Basis = "wrong" }, func(x *cpLive) { x.Compatible = nil }, func(x *cpLive) { x.Events[3].Source++ }, func(x *cpLive) { x.Events[3].Compositions++ }, func(x *cpLive) { psObj(psArr(x.IndependentSnapshot["signals"])[0])["title"] = "0 By (exact)" }, func(x *cpLive) { x.Reopened.Receipt.ID = "wrong" }}
	for n, change := range controls {
		var x cpLive
		if json.Unmarshal(raw, &x) != nil {
			t.Fatal("control decode")
		}
		change(&x)
		if reflect.DeepEqual(d, x) || cpLiveOK(x) {
			t.Fatalf("container CPU changed control %d accepted", n)
		}
	}
}
func TestContainerCPUOwnershipObserved(t *testing.T) {
	path := os.Getenv("GOO_CONTAINER_CPU_OWNER_LEDGER")
	if path == "" {
		t.Skip("set GOO_CONTAINER_CPU_OWNER_LEDGER")
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
			if held["status"] != float64(200) || !psEq(c["beforeRelease"], c["afterRelease"]) {
				return false
			}
			origin := psObj(psObj(c["initial"])["request"])
			primary := psObj(c["primary"])
			primaryResource := psObj(primary["resource"])
			switch c["phase"] {
			case "measurement":
				var result measurements.Result
				b, _ := json.Marshal(held["body"])
				if json.Unmarshal(b, &result) != nil {
					return false
				}
				if result.Question.Definition.Source.ExactStream == nil || result.Question.FromMS != int64(origin["fromMs"].(float64)) || result.Question.ToMS != int64(origin["toMs"].(float64)) || !psEq(result.Question.Definition.Source.ExactStream.Resource, primaryResource) {
					return false
				}
				if _, ok := cpRawGauge(psObj(x["facts"]), &result); !ok {
					return false
				}
			case "composition":
				var pair cpPair
				var origin exploration.PodDiscoveryQuestion
				b, _ := json.Marshal(map[string]any{"request": held["request"], "response": held["body"]})
				if json.Unmarshal(b, &pair) != nil {
					return false
				}
				b, _ = json.Marshal(psObj(c["initial"])["request"])
				if json.Unmarshal(b, &origin) != nil || !psEq(pair.Request.Parent.Resource, primaryResource) || !cpPairOK(psObj(x["facts"]), pair, origin) {
					return false
				}
			case "metadata":
				var q measurements.PodMetricDiscoveryQuestion
				var result measurements.PodMetricDiscovery
				b, _ := json.Marshal(held["request"])
				if json.Unmarshal(b, &q) != nil {
					return false
				}
				b, _ = json.Marshal(held["body"])
				if json.Unmarshal(b, &result) != nil || q.FromMS != int64(origin["fromMs"].(float64)) || q.ToMS != int64(origin["toMs"].(float64)) || !psEq(q.Resource, primaryResource) || !cmHeldMetadata(psObj(x["facts"]), q, result) {
					return false
				}
			default:
				return false
			}
		}
		return true
	}
	if !check(d) {
		t.Fatal("actual selected reads/compose published after withdrawal")
	}
	for _, key := range []string{"source", "plans", "compositions", "url", "panels", "metadata-count"} {
		var x psMap
		if json.Unmarshal(raw, &x) != nil {
			t.Fatal("control decode")
		}
		after := psObj(psObj(psArr(x["cases"])[0])["afterRelease"])
		if key == "metadata-count" {
			body := psObj(psObj(psObj(psArr(x["cases"])[0])["held"])["body"])
			body["records"] = body["records"].(float64) + 1
		} else if key == "source" || key == "plans" || key == "compositions" {
			after[key] = float64(999)
		} else if key == "panels" {
			after[key] = []any{map[string]any{"key": "wrong-late-panel", "view": 1}}
		} else {
			after[key] = "wrong"
		}
		if psEq(x, d) || check(x) {
			t.Fatalf("owner changed control %s accepted", key)
		}
	}
}
func TestContainerCPUDisclosureOwnershipObserved(t *testing.T) {
	path := os.Getenv("GOO_CONTAINER_CPU_DISCLOSURE_LEDGER")
	if path == "" {
		t.Skip("set GOO_CONTAINER_CPU_DISCLOSURE_LEDGER")
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
		before, after := psObj(x["beforeRelease"]), psObj(x["afterRelease"])
		held := psObj(x["held"])
		if held["status"] != float64(200) || !psEq(before, after) || before["source"] != float64(2) || before["plans"] != float64(0) || before["compositions"] != float64(0) || before["open"] != false || before["options"] != float64(0) || before["views"] != float64(0) {
			return false
		}
		var q measurements.PodMetricDiscoveryQuestion
		var result measurements.PodMetricDiscovery
		b, _ := json.Marshal(held["request"])
		if json.Unmarshal(b, &q) != nil {
			return false
		}
		b, _ = json.Marshal(held["body"])
		if json.Unmarshal(b, &result) != nil || !cmHeldMetadata(psObj(x["facts"]), q, result) {
			return false
		}
		origin := psObj(psObj(x["initial"])["request"])
		primary := psObj(x["primary"])
		if q.Environment != origin["environment"] || q.FromMS != int64(origin["fromMs"].(float64)) || q.ToMS != int64(origin["toMs"].(float64)) || !psEq(q.Resource, primary["resource"]) {
			return false
		}
		source, plan, composition := 0, 0, 0
		for _, v := range psArr(x["requests"]) {
			r := psObj(v)
			switch r["effect"] {
			case "source":
				source++
			case "plan":
				plan++
			case "composition":
				composition++
			}
		}
		return source == 2 && plan == 0 && composition == 0
	}
	if !check(d) {
		t.Fatal("closed disclosure accepted late metadata or follow-on intent")
	}
	for _, key := range []string{"source", "plans", "options", "views"} {
		var x psMap
		if json.Unmarshal(raw, &x) != nil {
			t.Fatal("control decode")
		}
		psObj(x["afterRelease"])[key] = float64(999)
		if psEq(x, d) || check(x) {
			t.Fatalf("closed disclosure control %s accepted", key)
		}
	}
}
