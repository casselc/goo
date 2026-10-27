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
	"regexp"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"
)

var cmDecimal = regexp.MustCompile(`^[+-]?([0-9]+(\.[0-9]*)?|\.[0-9]+)([eE][+-]?[0-9]+)?$`)

func cmSafeBytes(raw string) *big.Int {
	if len(raw) > 128 || !cmDecimal.MatchString(raw) {
		return nil
	}
	if i := strings.IndexAny(raw, "eE"); i >= 0 {
		e, err := strconv.Atoi(raw[i+1:])
		if err != nil || e < -400 || e > 400 {
			return nil
		}
	}
	r, ok := new(big.Rat).SetString(raw)
	if !ok || !r.IsInt() || r.Sign() < 0 || r.Num().BitLen() > 53 {
		return nil
	}
	return new(big.Int).Set(r.Num())
}
func cmRole(s metrics.Stream) string {
	if s.Unit != "By" || s.Version != "0.137.0" || s.SchemaURL != "" {
		return ""
	}
	if s.Name == "container.memory.working_set" && s.Scope == "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/kubeletstatsreceiver" {
		return "working-set"
	}
	if s.Scope != "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/k8sclusterreceiver" {
		return ""
	}
	if s.Name == "k8s.container.memory_request" {
		return "request"
	}
	if s.Name == "k8s.container.memory_limit" {
		return "limit"
	}
	return ""
}
func cmContext(a, b map[string]string) bool {
	for _, k := range []string{"container.id", "k8s.pod.uid", "k8s.container.name"} {
		if a[k] == "" || a[k] != b[k] {
			return false
		}
	}
	for _, k := range []string{"k8s.cluster.uid", "k8s.cluster.name", "k8s.namespace.name"} {
		av, ap := a[k]
		bv, bp := b[k]
		if ap != bp || av != bv {
			return false
		}
	}
	return true
}

type cmFact struct {
	stream metrics.Stream
	latest *metrics.GaugeSample
	bytes  *big.Int
}

// Raw fixture observations, not Evaluate or Compose, determine sample quality
// and exact byte certification. The ordinary gauge remains float-backed while
// descriptive percentages are certified from the original decimal strings.
func cmRawGauge(f psMap, r *measurements.Result) (cmFact, bool) {
	return rawContainerGauge(f, r, cmRole)
}
func rawContainerGauge(f psMap, r *measurements.Result, classify func(metrics.Stream) string) (cmFact, bool) {
	out := cmFact{}
	if r == nil {
		return out, false
	}
	q, d := r.Question, r.Question.Definition
	s := d.Source.ExactStream
	if s == nil || d.Kind != "gauge" || d.StateProfile != "" || d.StockDomain != "" || d.Source.Resource == nil || d.Source.Resource.Kind != "kubernetes-container" || d.Source.Resource.ID == "" || d.Source.Resource.ID != s.Resource["container.id"] || s.Resource["k8s.pod.uid"] == "" || s.Resource["k8s.container.name"] == "" || len(d.Source.Services) != 0 || d.Source.Namespace != nil || d.Environment != f["environment"] || s.Resource["deployment.environment.name"] != d.Environment || d.Source.Metric != s.Name || d.Source.Unit != s.Unit || d.Source.Scope != s.Scope || d.Source.Version != s.Version || d.Source.SchemaURL != s.SchemaURL || classify(*s) == "" || r.Gauges == nil || r.States != nil || r.Stocks != nil || r.Counters != nil || r.Histograms != nil {
		return out, false
	}
	if _, e := time.Parse(time.RFC3339Nano, r.ReadAt); e != nil {
		return out, false
	}
	if b, e := hex.DecodeString(r.Basis); e != nil || len(b) != 32 {
		return out, false
	}
	for k, v := range d.Source.Attributes {
		actual, has := s.Attributes[k]
		if !has || actual != v {
			return out, false
		}
	}
	lo, hi := uint64(q.FromMS)*1000000, uint64(q.ToMS)*1000000
	low, high := strconv.FormatUint(lo, 10), strconv.FormatUint(hi, 10)
	if q.FromMS < 0 || q.ToMS <= q.FromMS || (q.FromNS != "" && q.FromNS != low) || (q.ToNS != "" && q.ToNS != high) || r.Gauges.FromNS != low || r.Gauges.ToNS != high {
		return out, false
	}
	groups := map[string][]psMap{}
	physical := map[string]int{}
	count := 0
	for _, v := range psArr(f["points"]) {
		p := psObj(v)
		if p["kind"] != "gauge" || p["metric"] != s.Name || p["unit"] != s.Unit || p["scope"] != s.Scope || p["version"] != s.Version || p["schemaURL"] != s.SchemaURL || p["resourceSchemaURL"] != s.ResourceSchemaURL || !psEq(p["resource"], s.Resource) || !psEq(p["attributes"], s.Attributes) {
			continue
		}
		at, ok := p["endNS"].(string)
		if !ok {
			return out, false
		}
		ns, e := strconv.ParseUint(at, 10, 64)
		if e != nil {
			return out, false
		}
		if ns < lo || ns >= hi {
			continue
		}
		if _, ok := p["flags"].(float64); !ok {
			return out, false
		}
		groups[at] = append(groups[at], p)
		physical[siKey([]any{at, p["value"], p["flags"]})]++
		count++
	}
	if r.Count != strconv.Itoa(count) || len(r.Rows) != count {
		return out, false
	}
	for _, raw := range r.Rows {
		var row psMap
		if json.Unmarshal(raw, &row) != nil {
			return out, false
		}
		if !psEq(row["resource"], s.Resource) || !psEq(row["attributes"], s.Attributes) || row["resource_schema_url"] != s.ResourceSchemaURL || row["schema_url"] != s.SchemaURL || row["scope"] != s.Scope || row["version"] != s.Version || row["name"] != s.Name || row["unit"] != s.Unit {
			return out, false
		}
		key := siKey([]any{row["end_ns"], row["value"], row["flags"]})
		if physical[key] == 0 {
			return out, false
		}
		physical[key]--
	}
	out.stream = *s
	if count == 0 {
		return out, r.State == "no_data" && len(r.Gauges.Series) == 0
	}
	if len(r.Gauges.Series) != 1 || !reflect.DeepEqual(r.Gauges.Series[0].Stream, *s) {
		return out, false
	}
	series := r.Gauges.Series[0]
	times := []string{}
	for at := range groups {
		times = append(times, at)
	}
	sort.Slice(times, func(i, j int) bool {
		a, _ := strconv.ParseUint(times[i], 10, 64)
		b, _ := strconv.ParseUint(times[j], 10, 64)
		return a < b
	})
	if len(times) != len(series.Samples) {
		return out, false
	}
	for n, at := range times {
		points := groups[at]
		var value float64
		var missing bool
		variants := map[string]bool{}
		for i, p := range points {
			raw, ok := p["value"].(string)
			if !ok {
				return out, false
			}
			v, e := strconv.ParseFloat(raw, 64)
			if e != nil || math.IsNaN(v) || math.IsInf(v, 0) {
				return out, false
			}
			flag := int(p["flags"].(float64))&1 != 0
			variants[siKey([]any{v, flag})] = true
			if i == 0 {
				value, missing = v, flag
			}
		}
		quality, reason := "observed", ""
		ns, _ := strconv.ParseUint(at, 10, 64)
		if len(variants) > 1 {
			quality, reason = "unknown", "conflicting observations at one timestamp"
		} else if missing {
			quality, reason = "unknown", "no recorded value"
		} else if hi-ns > uint64(d.MaxGapMS)*1000000 {
			quality, reason = "stale", "older than declared freshness"
		}
		sample := series.Samples[n]
		if sample.AtNS != at || sample.State != quality || sample.Reason != reason {
			return out, false
		}
		if quality == "unknown" {
			if sample.Value != nil {
				return out, false
			}
		} else if sample.Value == nil || *sample.Value != value {
			return out, false
		}
		if n == len(times)-1 {
			out.latest = &series.Samples[n]
			if sample.Value != nil {
				for _, p := range points {
					integer := cmSafeBytes(p["value"].(string))
					if integer == nil || (out.bytes != nil && out.bytes.Cmp(integer) != 0) {
						out.bytes = nil
						break
					}
					out.bytes = integer
				}
			}
		}
	}
	overall := "observed"
	if out.latest.State != "observed" {
		overall = "unknown"
	}
	return out, series.State == out.latest.State && r.State == overall
}
func cmPercentText(r *big.Rat) string {
	scaled := new(big.Int).Mul(r.Num(), big.NewInt(10))
	q, rem := new(big.Int), new(big.Int)
	q.QuoRem(scaled, r.Denom(), rem)
	if new(big.Int).Mul(rem, big.NewInt(2)).Cmp(r.Denom()) >= 0 {
		q.Add(q, big.NewInt(1))
	}
	whole, fraction := new(big.Int), new(big.Int)
	whole.QuoRem(q, big.NewInt(10), fraction)
	return whole.String() + "." + fraction.String() + "%"
}

type cmPair struct {
	Request  measurements.ContainerMemoryViewRequest
	Response measurements.ContainerMemoryView
}

func cmPairOK(f psMap, p cmPair, origin exploration.PodDiscoveryQuestion) bool {
	q, v := p.Request, p.Response
	if q.Parent.ResourceSchemaURL == nil || q.Parent.Environment == "" || q.Parent.Resource["deployment.environment.name"] != q.Parent.Environment || v.Version != "goo/container-memory-view/v1" || !reflect.DeepEqual(v.Parent, q.Parent) || q.Parent.Environment != origin.Environment || q.Parent.FromMS != origin.FromMS || q.Parent.ToMS != origin.ToMS || len(v.Signals) != 3 || !cmContext(q.Parent.Resource, q.Parent.Resource) {
		return false
	}
	facts := map[string]cmFact{}
	results := map[string]*measurements.Result{"working-set": q.WorkingSet, "request": q.Request, "limit": q.Limit}
	for n, role := range []string{"working-set", "request", "limit"} {
		signal := v.Signals[n]
		if signal.Role != role {
			return false
		}
		r := results[role]
		if r == nil {
			if signal.State != "unavailable" || signal.Reason != "not-supplied" || signal.Value != "" || signal.AtNS != "" || signal.Question != nil || signal.ReadAt != "" || signal.Basis != "" {
				return false
			}
			continue
		}
		fact, ok := cmRawGauge(f, r)
		if !ok || cmRole(fact.stream) != role || !cmContext(q.Parent.Resource, fact.stream.Resource) || r.Question.Definition.Environment != q.Parent.Environment || r.Question.FromMS != q.Parent.FromMS || r.Question.ToMS != q.Parent.ToMS || signal.Question == nil || !aeJSONEqual(*signal.Question, r.Question) || signal.ReadAt != r.ReadAt || signal.Basis != r.Basis {
			return false
		}
		facts[role] = fact
		state, reason, at, observation, value := "no_data", "no-observations", "", "", ""
		if fact.latest != nil {
			state, reason, at, observation = fact.latest.State, fact.latest.Reason, fact.latest.AtNS, fact.latest.State
			if fact.latest.Value != nil {
				if fact.bytes == nil {
					state, reason = "invalid", "invalid-byte-domain"
				} else {
					value = fact.bytes.String()
				}
			}
		}
		if signal.State != state || signal.Reason != reason || signal.AtNS != at || signal.ObservationState != observation || signal.Value != value {
			return false
		}
	}
	for _, role := range []string{"request", "limit"} {
		a, aok := facts["working-set"]
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
		case a.bytes == nil || b.bytes == nil:
			reason = "invalid-byte-domain"
		case !reflect.DeepEqual(a.stream.Resource, b.stream.Resource) || a.stream.ResourceSchemaURL != b.stream.ResourceSchemaURL || !psEq(a.stream.Attributes, b.stream.Attributes):
			reason = "resource-or-attribute-mismatch"
		case a.latest.AtNS != b.latest.AtNS:
			reason = "observation-time-mismatch"
		case results["working-set"].Question.Definition.MaxGapMS != results[role].Question.Definition.MaxGapMS:
			reason = "freshness-policy-mismatch"
		case b.bytes.Sign() == 0:
			reason = "zero-denominator"
		default:
			ratio = new(big.Rat).SetFrac(new(big.Int).Mul(a.bytes, big.NewInt(100)), b.bytes)
		}
		actual := v.Percentages.Request
		if role == "limit" {
			actual = v.Percentages.Limit
		}
		if ratio == nil {
			if actual.State != "unavailable" || actual.Reason != reason || actual.Value != "" || actual.Text != "" {
				return false
			}
		} else if actual.State != "observed" || actual.Reason != "" || actual.Value != ratio.RatString() || actual.Text != cmPercentText(ratio) {
			return false
		}
	}
	return true
}
func cmSnapshot(v measurements.ContainerMemoryView, snapshot psMap) bool {
	rows := psArr(snapshot["signals"])
	percent := psArr(snapshot["percentages"])
	if len(rows) != 3 || len(percent) != 2 {
		return false
	}
	max := big.NewInt(0)
	for _, s := range v.Signals {
		if s.Value != "" {
			n, ok := new(big.Int).SetString(s.Value, 10)
			if !ok {
				return false
			}
			if n.Cmp(max) > 0 {
				max = n
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
		if !strings.Contains(text, quality) {
			return false
		}
		if row["role"] != s.Role || row["state"] != s.State {
			return false
		}
		expected, title := s.Value, s.Value+" By (exact)"
		if expected == "" {
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
		if bar["value"] != expected || bar["scaleMax"] != max.String() {
			return false
		}
		if s.Value != "" {
			n, _ := new(big.Int).SetString(s.Value, 10)
			width := float64(0)
			if max.Sign() > 0 {
				q := new(big.Int).Quo(new(big.Int).Mul(n, big.NewInt(10000)), max)
				width = float64(q.Int64()) / 100
			}
			text, ok := bar["fill"].(string)
			if !ok || !strings.HasSuffix(text, "%") {
				return false
			}
			actual, e := strconv.ParseFloat(strings.TrimSuffix(text, "%"), 64)
			if e != nil || math.Abs(actual-width) > 0.00001 {
				return false
			}
		} else if bar["fill"] != nil {
			return false
		}
	}
	for n, p := range []measurements.ContainerMemoryPercentage{v.Percentages.Request, v.Percentages.Limit} {
		role := []string{"request", "limit"}[n]
		row := psObj(percent[n])
		if p.State == "observed" {
			if row["text"] != "Working set / "+role+": "+p.Text || row["title"] != p.Value+"%" {
				return false
			}
		} else if row["title"] != "" || row["text"] != "Working set / "+role+" unavailable · "+strings.ReplaceAll(p.Reason, "-", " ") {
			return false
		}
	}
	return true
}

type cmLive struct {
	pdLive
	First, Second             pnRead
	Cluster                   exploration.PodResourceObservation
	Compatible                []string
	Independent               cmPair
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
		Composed cmPair
		Snapshot psMap
	}
	Blocked []string
	Events  []struct {
		Action                      string
		Source, Plans, Compositions int
	}
}

func cmLiveOK(d cmLive) bool {
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
	checkPair := func(p cmPair, snapshot psMap) bool {
		if !cmPairOK(d.Facts, p, d.Initial.Request) || !network("/api/container-memory/view", p.Response) || !cmSnapshot(p.Response, snapshot) {
			return false
		}
		for _, r := range []*measurements.Result{p.Request.WorkingSet, p.Request.Request, p.Request.Limit} {
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
	for _, p := range []measurements.ContainerMemoryPercentage{d.Independent.Response.Percentages.Request, d.Independent.Response.Percentages.Limit} {
		if p.State != "unavailable" {
			return false
		}
	}
	for _, r := range []*measurements.Result{d.Independent.Request.WorkingSet, d.Independent.Request.Request, d.Independent.Request.Limit} {
		if r == nil {
			return false
		}
		found := false
		for _, read := range []pnRead{d.First, d.Second} {
			for _, w := range read.Plan.Candidates {
				if aeJSONEqual(w.Question, r.Question) && w.Role == cmRole(*r.Question.Definition.Source.ExactStream) {
					found = true
				}
			}
		}
		if !found {
			return false
		}
	}
	if _, ok := cmRawGauge(d.Facts, &d.Individual); !ok || !network("/api/measurements", d.Individual) || !aeJSONEqual(d.Individual.Question, d.Independent.Request.WorkingSet.Question) {
		return false
	}
	if len(d.Boundary) != 7 {
		return false
	}
	for _, b := range d.Boundary {
		if !reflect.DeepEqual(b.Composed.Request.Parent, b.Opened.DiscoveryRequest) || !ppMetadata(d.Facts, b.Opened) || len(b.Opened.Plan.Widgets) != 0 || !checkPair(b.Composed, b.Snapshot) || !reflect.DeepEqual(b.Row.Resource, b.Opened.DiscoveryRequest.Resource) {
			return false
		}
	}
	actions := []string{"inventory", "explicit-first-metadata-no-values", "explicit-second-metadata", "local-role-selection", "independent-three-signals", "native-staged", "retained-back", "native-run", "individual-capture", "capture-retained-return", "boundary-aligned", "boundary-zero-denominator", "boundary-partial", "boundary-unsafe", "boundary-stale", "boundary-conflict", "boundary-missing", "frozen-reopen"}
	sources := []int{1, 2, 3, 3, 6, 6, 6, 7, 8, 8, 12, 16, 18, 22, 26, 30, 34, 34}
	plans := []int{0, 1, 2, 2, 2, 3, 3, 4, 4, 4, 5, 6, 7, 8, 9, 10, 11, 11}
	compositions := []int{0, 0, 0, 0, 1, 1, 1, 1, 1, 1, 2, 3, 4, 5, 6, 7, 8, 8}
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
	return source == 34 && pure == 11 && compose == 8 && endpointCounts["/api/pod-discovery"] == 1 && endpointCounts["/api/pod-metric-discovery"] == 9 && endpointCounts["/api/measurements"] == 23 && endpointCounts["/api/captures"] == 1
}
func TestContainerMemoryObserved(t *testing.T) {
	path := os.Getenv("GOO_CONTAINER_MEMORY_LEDGER")
	if path == "" {
		t.Skip("set GOO_CONTAINER_MEMORY_LEDGER to actual206 receipt")
	}
	raw, e := os.ReadFile(path)
	if e != nil {
		t.Fatal(e)
	}
	var d cmLive
	if e = json.Unmarshal(raw, &d); e != nil {
		t.Fatal(e)
	}
	if !cmLiveOK(d) {
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
	if _, ok := cmRawGauge(d.Facts, stored.Measurement); !ok {
		t.Fatal("raw stored gauge replay")
	}
	controls := []func(*cmLive){func(x *cmLive) {
		x.Independent.Response.Percentages.Limit.State = "observed"
		x.Independent.Response.Percentages.Limit.Value = "25"
	}, func(x *cmLive) { x.Boundary[0].Composed.Response.Percentages.Request.Value = "51" }, func(x *cmLive) {
		x.Boundary[1].Composed.Response.Percentages.Request.State = "observed"
		x.Boundary[1].Composed.Response.Percentages.Request.Value = "0"
	}, func(x *cmLive) { x.Boundary[2].Composed.Response.Signals[1].Value = "0" }, func(x *cmLive) {
		x.Boundary[3].Composed.Response.Signals[0].State = "observed"
		x.Boundary[3].Composed.Response.Signals[0].Value = "9007199254740992"
	}, func(x *cmLive) { x.Boundary[4].Composed.Response.Signals[0].State = "observed" }, func(x *cmLive) { x.Boundary[5].Composed.Request.WorkingSet.Rows = nil }, func(x *cmLive) {
		x.Independent.Request.WorkingSet.Question.Definition.Source.Resource.ID = "replacement"
	}, func(x *cmLive) {
		x.Independent.Request.Request.Question.Definition.Source.ExactStream.ResourceSchemaURL = "wrong"
	}, func(x *cmLive) { x.Individual.ReadAt = "wrong" }, func(x *cmLive) { x.Individual.Basis = "wrong" }, func(x *cmLive) { x.Compatible = nil }, func(x *cmLive) { x.Events[3].Source++ }, func(x *cmLive) { x.Events[3].Compositions++ }, func(x *cmLive) { psObj(psArr(x.IndependentSnapshot["signals"])[0])["title"] = "0 By (exact)" }, func(x *cmLive) { x.Reopened.Receipt.ID = "wrong" }}
	for n, change := range controls {
		var x cmLive
		if json.Unmarshal(raw, &x) != nil {
			t.Fatal("control decode")
		}
		change(&x)
		if reflect.DeepEqual(d, x) || cmLiveOK(x) {
			t.Fatalf("container memory changed control %d accepted", n)
		}
	}
}
func cmHeldMetadata(f psMap, q measurements.PodMetricDiscoveryQuestion, r measurements.PodMetricDiscovery) bool {
	if q.ResourceSchemaURL == nil || q.Environment != f["environment"] || !reflect.DeepEqual(q, r.Question) || !cmContext(q.Resource, q.Resource) {
		return false
	}
	groups := map[string]int{}
	count := 0
	for _, v := range psArr(f["points"]) {
		p := psObj(v)
		if !psEq(p["resource"], q.Resource) || p["resourceSchemaURL"] != *q.ResourceSchemaURL {
			continue
		}
		at, e := strconv.ParseUint(p["endNS"].(string), 10, 64)
		if e != nil {
			return false
		}
		if at < uint64(q.FromMS)*1000000 || at >= uint64(q.ToMS)*1000000 {
			continue
		}
		kind := p["kind"]
		tempo, mono := p["temporality"], p["monotonic"]
		if kind == "gauge" {
			tempo, mono = float64(0), false
		}
		if kind == "sum" {
			kind = "counter"
			if tempo == float64(2) && mono == false {
				kind = "stock"
			}
		}
		groups[siKey([]any{kind, p["scope"], p["version"], p["schemaURL"], p["metric"], p["unit"], tempo, mono})]++
		count++
	}
	if r.Version != "goo/pod-metric-discovery/v1" || r.Records != count || r.InstrumentsOmitted != 0 || len(r.Instruments) != len(groups) {
		return false
	}
	for _, i := range r.Instruments {
		key := siKey([]any{i.Kind, i.Scope, i.Version, i.SchemaURL, i.Metric, i.Unit, i.Temporality, i.Monotonic})
		if groups[key] != i.Records {
			return false
		}
		delete(groups, key)
	}
	return len(groups) == 0
}
func TestContainerMemoryOwnershipObserved(t *testing.T) {
	path := os.Getenv("GOO_CONTAINER_MEMORY_OWNER_LEDGER")
	if path == "" {
		t.Skip("set GOO_CONTAINER_MEMORY_OWNER_LEDGER")
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
				if _, ok := cmRawGauge(psObj(x["facts"]), &result); !ok {
					return false
				}
			case "composition":
				var pair cmPair
				var origin exploration.PodDiscoveryQuestion
				b, _ := json.Marshal(map[string]any{"request": held["request"], "response": held["body"]})
				if json.Unmarshal(b, &pair) != nil {
					return false
				}
				b, _ = json.Marshal(psObj(c["initial"])["request"])
				if json.Unmarshal(b, &origin) != nil || !psEq(pair.Request.Parent.Resource, primaryResource) || !cmPairOK(psObj(x["facts"]), pair, origin) {
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
func TestContainerMemoryDisclosureOwnershipObserved(t *testing.T) {
	path := os.Getenv("GOO_CONTAINER_MEMORY_DISCLOSURE_LEDGER")
	if path == "" {
		t.Skip("set GOO_CONTAINER_MEMORY_DISCLOSURE_LEDGER")
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
