package verification

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/measurements"
	"github.com/casselc/goo/metrics"
	"github.com/casselc/goo/scopedview"
	"github.com/casselc/goo/workbench"
	"github.com/casselc/observability/quintgo/qobs"
	"github.com/casselc/observability/quintgo/validate"
)

type clientRaw392 struct {
	entryRaw390
	Bounds []float64
}

// Finite source admission for the observed 392 bank. This is not a general
// histogram evaluator: it admits the bounded exported bucket/count differences
// and explicit epoch/flag/gap cases exercised by this journey. Existing kernel
// properties remain responsible for the general algorithm.
func admitClientRead392(read entryRead390, bank []clientRaw392) (*measurements.Result, error) {
	fail := func(s string) (*measurements.Result, error) { return nil, fmt.Errorf("native392 %s: %s", read.Name, s) }
	var result measurements.Result
	if read.Status != 200 || json.Unmarshal(read.Response, &result) != nil || !reflect.DeepEqual(result.Question, read.Request) || measurements.ValidateResult(&result) != nil || result.Histograms == nil || len(result.Histograms.Series) != 1 {
		return fail("actual native result/request refused")
	}
	selected := []clientRaw392{}
	fromText, toText := read.Request.BoundsNS()
	lo, e := strconv.ParseInt(fromText, 10, 64)
	if e != nil {
		return fail("invalid lower bound")
	}
	hi, e := strconv.ParseInt(toText, 10, 64)
	if e != nil {
		return fail("invalid upper bound")
	}
	for _, row := range bank {
		at, e := strconv.ParseInt(row.AtNS, 10, 64)
		if e != nil {
			return fail("invalid physical timestamp")
		}
		if rawIdentity390(read.Request, row.entryRaw390) && at >= lo && at <= hi {
			selected = append(selected, row)
		}
	}
	sort.Slice(selected, func(i, j int) bool { return selected[i].AtNS < selected[j].AtNS })
	if len(selected) < 2 || len(selected) > 6 || result.Count != strconv.Itoa(len(selected)) {
		return fail("finite exact source stream absent")
	}
	series := result.Histograms.Series[0]
	source := selected[0]
	stream := metrics.Stream{Resource: source.ResourceAttributes, Attributes: source.Attributes, Scope: source.ScopeName, Version: source.ScopeVersion, SchemaURL: source.ScopeSchemaUrl, ResourceSchemaURL: source.ResourceSchemaUrl, Name: source.MetricName, Unit: source.MetricUnit}
	if !reflect.DeepEqual(series.Stream, stream) {
		return fail("full stream identity changed")
	}
	cumulative := source.Temporality == "2"
	if !cumulative && source.Temporality != "1" {
		return fail("unsupported physical temporality")
	}
	if series.Temporality != map[bool]string{false: "delta", true: "cumulative"}[cumulative] {
		return fail("temporality relabelled")
	}
	expected := []metrics.HistogramInterval{}
	for index, p := range selected {
		count, e := strconv.ParseUint(p.Count, 10, 64)
		if e != nil || strconv.FormatUint(count, 10) != p.Count {
			return fail("noncanonical physical count")
		}
		buckets := make([]uint64, len(p.BucketCounts))
		sum := uint64(0)
		for j, text := range p.BucketCounts {
			n, e := strconv.ParseUint(text, 10, 64)
			if e != nil || strconv.FormatUint(n, 10) != text {
				return fail("noncanonical physical bucket")
			}
			buckets[j] = n
			sum += n
		}
		if sum != count || len(p.Bounds) != 2 || !slices.Equal(p.Bounds, []float64{0.01, 0.1}) || len(buckets) != 3 {
			return fail("physical bucket/count/bounds mismatch")
		}
		if cumulative && index == 0 {
			continue
		}
		start := p.StartNS
		reason := ""
		flags, e := flags390(p.Flags)
		if e != nil {
			return fail("physical flags invalid")
		}
		if cumulative {
			prior := selected[index-1]
			start = prior.AtNS
			if p.StartNS != prior.StartNS {
				reason = "histogram-epoch-change"
			} else {
				priorCount, e := strconv.ParseUint(prior.Count, 10, 64)
				if e != nil || count < priorCount {
					return fail("unexpected bounded cumulative decrease")
				}
				count -= priorCount
				for j := range buckets {
					previous, e := strconv.ParseUint(prior.BucketCounts[j], 10, 64)
					if e != nil || buckets[j] < previous {
						return fail("unexpected bounded bucket decrease")
					}
					buckets[j] -= previous
				}
			}
			priorFlags, e := flags390(prior.Flags)
			if e != nil {
				return fail("prior flags invalid")
			}
			flags |= priorFlags
		}
		if reason == "" && flags&1 != 0 {
			reason = "no-recorded-value"
		}
		a, e1 := strconv.ParseInt(start, 10, 64)
		b, e2 := strconv.ParseInt(p.AtNS, 10, 64)
		if e1 != nil || e2 != nil || a < lo || b > hi || b <= a {
			return fail("invalid physical export interval")
		}
		if reason == "" && b-a > read.Request.Definition.MaxGapMS*1000000 {
			reason = "excessive-gap"
		}
		interval := metrics.HistogramInterval{FromNS: start, ToNS: p.AtNS, State: "unknown", Reason: reason}
		if reason == "" {
			values := make([]string, len(buckets))
			for j, n := range buckets {
				values[j] = strconv.FormatUint(n, 10)
			}
			interval.State = "ok"
			interval.Distribution = &metrics.Distribution{Unit: p.MetricUnit, Bounds: p.Bounds, BucketCounts: values, Count: strconv.FormatUint(count, 10)}
		}
		expected = append(expected, interval)
	}
	if len(series.Intervals) != len(expected) {
		return fail("export interval population changed")
	}
	for i, want := range expected {
		got := series.Intervals[i]
		if got.FromNS != want.FromNS || got.ToNS != want.ToNS || got.State != want.State || got.Reason != want.Reason {
			return fail("reset/gap/unknown interval state changed")
		}
		if want.Distribution == nil {
			if got.Distribution != nil {
				return fail("unknown acquired numeric distribution")
			}
			continue
		}
		if got.Distribution == nil || got.Distribution.Unit != want.Distribution.Unit || got.Distribution.Count != want.Distribution.Count || !slices.Equal(got.Distribution.BucketCounts, want.Distribution.BucketCounts) || !slices.Equal(got.Distribution.Bounds, want.Distribution.Bounds) {
			return fail("raw bucket differences/count/unit changed")
		}
		// Verify the conservative P95 bucket, never interpolation or fleet averaging.
		total, _ := strconv.ParseUint(want.Distribution.Count, 10, 64)
		if total > 0 {
			target := (total*95 + 99) / 100
			running := uint64(0)
			bin := 0
			for j, text := range want.Distribution.BucketCounts {
				n, _ := strconv.ParseUint(text, 10, 64)
				running += n
				if running >= target {
					bin = j
					break
				}
			}
			p95 := got.Distribution.P95
			state := "bucket-bound"
			if bin == 0 {
				state = "unbounded-lower"
			}
			if bin == len(want.Distribution.Bounds) {
				state = "unbounded-upper"
			}
			if p95.State != state {
				return fail("known export lost percentile bucket range")
			}
			if bin == 0 {
				if p95.Lower != nil {
					return fail("first percentile bucket fabricated lower bound")
				}
			} else if p95.Lower == nil || *p95.Lower != want.Distribution.Bounds[bin-1] {
				return fail("percentile lower bucket bound changed")
			}
			if bin == len(want.Distribution.Bounds) {
				if p95.Upper != nil {
					return fail("overflow percentile bucket fabricated upper bound")
				}
			} else if p95.Upper == nil || *p95.Upper != want.Distribution.Bounds[bin] {
				return fail("percentile upper bucket bound changed")
			}
		}
	}
	return &result, nil
}

type clientCall392 struct {
	Path, Method   string
	Body, Response json.RawMessage
	Status         int
}
type clientCase392 struct {
	Role                                                   string
	Plan                                                   exploration.DashboardPlan
	Natives, Qualified                                     []exploration.DashboardWidget
	RecordProfileOptions, SourcePaths, ExpectedSourcePaths []string
	FromCall, ToCall, WidgetReads                          int
	StreamQuestion                                         measurements.Question
	Measurement                                            clientCall392
	Retained                                               struct {
		Snapshot scopedview.Dataset
		Display  json.RawMessage
	}
	SpanLens struct {
		Plan  exploration.DashboardPlan
		Reads []clientCall392
	}
}
type clientLedger392 struct {
	Version string
	Fixture struct {
		Environment, Namespace string
		Scope                  exploration.InventoryQuestion
		Services               map[string]string
		Streams                []struct {
			Role     string
			Question measurements.Question
		}
	}
	InventoryCapture                    exploration.CapturePage
	RawHistograms                       []clientRaw392
	RawSpans                            []httpRaw385
	Reads                               []entryRead390
	Cases                               []clientCase392
	SourceQuestion, DestinationQuestion measurements.Question
	SourceRun, DestinationRun           entryRead390
	Definitions                         []struct {
		Request  workbench.NativeViewInput
		Response workbench.NativeView
		Status   int
	}
	Captures []struct {
		ID, StoredBytesBefore, StoredBytesAfter string
		Response                                exploration.CapturePage
	}
	ViewBytes struct{ Before, After string }
	Reuse     struct {
		Request  workbench.NativeViewReuseRequest
		Response workbench.NativeView
		Status   int
	}
	Facts               []patternFact391
	Calls               []clientCall392
	Errors              []string
	Setup               struct{ NativeReads, InventoryCaptures int }
	DestinationSnapshot struct {
		Snapshot scopedview.Dataset
		Display  json.RawMessage
	}
	UnknownSelection struct {
		SelectedKey, Text string
		ChartAbsent       bool
	}
	Actor struct {
		Snapshots, SQLBefore, SQLAfter int
		Complete                       struct {
			Version, Phase string
			SelectedKey    string
			Chart          json.RawMessage
			Points         []entryPoint390
			Overflow       bool
			Requests       []struct {
				Path     string
				Request  measurements.Question
				Response json.RawMessage
				Status   int
			}
		}
	}
	ProxyCalls []struct {
		SQL        string
		Parameters map[string]string
		Status     int
	}
}

var clientActions392 = []string{"service-load-metricsOnly", "service-retained-local", "open-exact-stream-question", "service-load-spansOnly", "explicit-client-span-lens", "service-load-combined", "service-load-wrongUnit", "explicit-source-run", "save-display-configuration", "capture-source-evidence", "inspect-apply-destination", "explicit-destination-run", "destination-buckets-local", "unknown-then-known-local", "complete"}

func admitClientLedger392(r clientLedger392) error {
	fail := func(s string) error { return fmt.Errorf("native392: %s", s) }
	if r.Version != "goo.nativeHTTPClient392/1" || len(r.Errors) != 0 || len(r.RawHistograms) != 15 || len(r.RawSpans) != 6 || len(r.Reads) != 4 || len(r.Cases) != 4 || r.Setup.NativeReads != 4 || r.Setup.InventoryCaptures != 1 {
		return fail("finite actual bank/setup absent")
	}
	for _, raw := range r.RawHistograms {
		if raw.ResourceAttributes["deployment.environment.name"] != r.Fixture.Environment || raw.ResourceAttributes["service.namespace"] != r.Fixture.Namespace || raw.MetricName != "http.client.request.duration" {
			return fail("foreign physical metric row")
		}
	}
	for _, read := range r.Reads {
		if _, e := admitClientRead392(read, r.RawHistograms); e != nil {
			return e
		}
	}
	for i := range r.RawSpans {
		at, e := strconv.ParseInt(r.RawSpans[i].TimestampNS, 10, 64)
		if e != nil {
			return fail("raw span timestamp invalid")
		}
		r.RawSpans[i].At = at / 1000000
	}
	roles := map[string]bool{}
	sourceCosts := map[string]int{}
	for _, c := range r.Cases {
		if roles[c.Role] || r.Fixture.Services[c.Role] == "" || c.Plan.Service != r.Fixture.Services[c.Role] || !reflect.DeepEqual(c.Plan.Scope, r.Fixture.Scope) {
			return fail("service source case identity changed")
		}
		roles[c.Role] = true
		qualified := []exploration.DashboardWidget{}
		executable := 0
		expectedPaths := []string{"/api/discovery", "/api/discovery", "/api/explore", "/api/metric-discovery", "/api/inventory"}
		for _, w := range c.Plan.Widgets {
			if w.State == "supported" && (w.Query != nil || w.Measurement != nil) {
				executable++
				path := "/api/record-aggregates"
				if w.Measurement != nil {
					path = "/api/measurements"
				}
				expectedPaths = append(expectedPaths, path)
			}
			if w.Measurement != nil && w.Family == "http-client" {
				qualified = append(qualified, w)
				if w.Grain != "per-stream" || w.Measurement.Definition.Source.Metric != "http.client.request.duration" || w.Measurement.Definition.Source.Unit != "s" {
					return fail("native widget borrowed another profile or unit")
				}
			}
		}
		wantNative := c.Role == "metricsOnly" || c.Role == "combined"
		if (len(qualified) > 0) != wantNative || !reflect.DeepEqual(qualified, c.Qualified) || c.WidgetReads != executable || !slices.Equal(c.SourcePaths, expectedPaths) || !slices.Equal(c.ExpectedSourcePaths, expectedPaths) {
			return fail("native/profile or source acquisition boundary changed")
		}
		if c.FromCall < 0 || c.ToCall > len(r.Calls) || c.ToCall < c.FromCall {
			return fail("invalid actual service call boundary")
		}
		actualPaths := []string{}
		for _, call := range r.Calls[c.FromCall:c.ToCall] {
			if call.Method == "POST" && slices.Contains([]string{"/api/discovery", "/api/explore", "/api/explore/page", "/api/metric-discovery", "/api/inventory", "/api/record-aggregates", "/api/measurements", "/api/captures"}, call.Path) {
				actualPaths = append(actualPaths, call.Path)
			}
		}
		if !slices.Equal(actualPaths, expectedPaths) {
			return fail("actual service endpoints departed from source plan")
		}
		sourceCosts["service-load-"+c.Role] = 5 + executable
		if c.Role == "metricsOnly" {
			if c.Plan.Profile == "http-client" || slices.Contains(c.RecordProfileOptions, "http-client") {
				return fail("metric-only observations promoted Client spans")
			}
			var q measurements.Question
			if json.Unmarshal(c.Measurement.Body, &q) != nil {
				return fail("native service request absent")
			}
			read := entryRead390{Name: "service", Request: q, Response: c.Measurement.Response, Status: c.Measurement.Status}
			data, e := admitClientRead392(read, r.RawHistograms)
			if e != nil {
				return e
			}
			view, err := measurements.HistogramView(data, 0, 0)
			if err != nil || !passiveClientView392(view, c.Retained.Snapshot) {
				return fail("service retained buckets lost full source identity/data/clocks")
			}

			if len(data.Histograms.Series) != 1 || !reflect.DeepEqual(c.StreamQuestion.Definition.Source.ExactStream, &data.Histograms.Series[0].Stream) || !reflect.DeepEqual(c.StreamQuestion, r.SourceQuestion) {
				return fail("open stream question lost complete stream identity")
			}
		}
		if c.Role == "spansOnly" || c.Role == "combined" {
			if !slices.Contains(c.RecordProfileOptions, "http-client") {
				return fail("independent current Client spans unavailable")
			}
		}
		if c.Role == "spansOnly" {
			if len(c.SpanLens.Reads) != 5 || c.SpanLens.Plan.Profile != "http-client" {
				return fail("explicit independent Client lens missing")
			}
			sourceCosts["explicit-client-span-lens"] = len(c.SpanLens.Reads)
			f := httpFixture385{Mode: 1, Scope: r.Fixture.Scope, Service: r.Fixture.Services[c.Role], Rows: r.RawSpans}
			for _, call := range c.SpanLens.Reads {
				var q exploration.Query
				if json.Unmarshal(call.Body, &q) != nil {
					return fail("Client span request invalid")
				}
				if e := httpAdmitResult385(f, httpRead385{Request: q, Response: call.Response, Status: call.Status}, false); e != nil {
					return e
				}
				var result exploration.Result
				if json.Unmarshal(call.Response, &result) != nil {
					return fail("Client span result invalid")
				}
				wanted := map[string]int{}
				field := q.GroupBy[0]
				for _, row := range f.Rows {
					if !httpMember385(f, row) {
						continue
					}
					var value *string
					switch field.Kind {
					case "attribute":
						if text, ok := row.Attributes[field.Key]; ok {
							copy := text
							value = &copy
						}
					case "resource":
						if text, ok := row.Resource[field.Key]; ok {
							copy := text
							value = &copy
						}
					case "column":
						if field.Key == "name" {
							copy := row.Operation
							value = &copy
						} else if text, ok := httpValues385(row)[field]; ok {
							copy := text
							value = &copy
						}
					}
					wanted[continuationJSON352(value)]++
				}
				got := map[string]int{}
				for _, raw := range result.Groups {
					var g struct {
						G0    *string
						Count string
					}
					if json.Unmarshal(raw, &g) != nil {
						return fail("Client span groups invalid")
					}
					n, e := strconv.Atoi(g.Count)
					if e != nil {
						return e
					}
					got[continuationJSON352(g.G0)] = n
				}
				if !reflect.DeepEqual(got, wanted) {
					return fail("Client group counts differ from raw source")
				}
			}
		}
	}
	storePaths := []string{}
	for _, call := range r.Calls {
		if call.Method == "GET" && (strings.HasPrefix(call.Path, "/api/captures/") || strings.HasPrefix(call.Path, "/api/native-views/")) {
			storePaths = append(storePaths, call.Path)
		}
	}
	expectedStores := []string{}
	for range 4 {
		expectedStores = append(expectedStores, "/api/captures/"+r.InventoryCapture.Receipt.ID, "/api/captures/search")
	}
	if len(r.Definitions) != 1 {
		return fail("native saved definition absent")
	}
	expectedStores = append(expectedStores, "/api/captures/search", "/api/captures/search", "/api/native-views/"+r.Definitions[0].Response.ID)
	if !slices.Equal(storePaths, expectedStores) {
		return fail("actual artifact/catalog read endpoints changed")
	}
	if len(roles) != 4 {
		return fail("missing independent evidence mode")
	}
	if !reflect.DeepEqual(r.SourceRun.Request, r.SourceQuestion) || !reflect.DeepEqual(r.DestinationRun.Request, r.DestinationQuestion) || reflect.DeepEqual(r.SourceQuestion, r.DestinationQuestion) {
		return fail("source/destination question identity changed")
	}
	for _, read := range []entryRead390{r.SourceRun, r.DestinationRun} {
		if _, e := admitClientRead392(read, r.RawHistograms); e != nil {
			return e
		}
	}
	var destination measurements.Result
	if json.Unmarshal(r.DestinationRun.Response, &destination) != nil {
		return fail("destination result decode failed")
	}
	destinationView, err := measurements.HistogramView(&destination, 0, 0)
	if err != nil || !sameTransport384(destinationView, r.DestinationSnapshot.Snapshot) {
		return fail("destination histogram snapshot borrowed other source/clocks: " + fmt.Sprint(err) + " " + transportDifference384(destinationView, r.DestinationSnapshot.Snapshot, "snapshot"))
	}
	for _, call := range r.Calls {
		if call.Path == "/api/measurements" && call.Method == "POST" {
			var q measurements.Question
			if json.Unmarshal(call.Body, &q) != nil {
				return fail("actual native source call invalid")
			}
			if _, e := admitClientRead392(entryRead390{Name: "page-native", Request: q, Response: call.Response, Status: call.Status}, r.RawHistograms); e != nil {
				return e
			}
		}
	}
	if len(r.Definitions) != 1 || len(r.Captures) != 1 || r.ViewBytes.Before == "" || r.ViewBytes.Before != r.ViewBytes.After {
		return fail("immutable native configuration/capture missing")
	}
	definition := r.Definitions[0]
	if definition.Status != 201 || definition.Response.Validate() != nil || !reflect.DeepEqual(definition.Request.Question, r.SourceQuestion) || !reflect.DeepEqual(definition.Response.Question, r.SourceQuestion) {
		return fail("saved native definition changed source")
	}
	var storedView workbench.NativeView
	if json.Unmarshal([]byte(r.ViewBytes.Before), &storedView) != nil || !reflect.DeepEqual(storedView, definition.Response) {
		return fail("saved native file differs from HTTP artifact")
	}
	wire, _ := json.Marshal(definition.Request)
	sum := sha256.Sum256(wire)
	if hex.EncodeToString(sum[:]) != definition.Response.ID {
		return fail("native artifact content identity changed")
	}
	if r.Reuse.Status != 200 || r.Reuse.Request.ID != definition.Response.ID || !reflect.DeepEqual(r.Reuse.Request.Question, r.DestinationQuestion) || !reflect.DeepEqual(r.Reuse.Response.Question, r.DestinationQuestion) || !reflect.DeepEqual(r.Reuse.Response.Declaration, definition.Response.Declaration) || r.Reuse.Response.ID != "" {
		return fail("display reuse borrowed source observations or identity")
	}
	capture := r.Captures[0]
	if capture.StoredBytesBefore == "" || capture.StoredBytesBefore != capture.StoredBytesAfter {
		return fail("captured bytes changed")
	}
	var saved exploration.SavedCapture
	if json.Unmarshal([]byte(capture.StoredBytesBefore), &saved) != nil || saved.Measurement == nil {
		return fail("native capture absent")
	}
	id := saved.ID
	saved.ID = ""
	wire, _ = json.Marshal(saved)
	sum = sha256.Sum256(wire)
	if id != capture.ID || hex.EncodeToString(sum[:]) != id {
		return fail("capture identity changed")
	}
	saved.ID = id
	page, e := saved.Page("")
	if e != nil || !sameTransport384(page, capture.Response) || !reflect.DeepEqual(saved.Question.Measurement, &r.SourceQuestion) {
		return fail("frozen source receipt/page changed")
	}
	raw, _ := json.Marshal(saved.Measurement)
	if _, e := admitClientRead392(entryRead390{Name: "capture", Request: r.SourceQuestion, Response: raw, Status: 200}, r.RawHistograms); e != nil {
		return e
	}
	if r.UnknownSelection.SelectedKey != "1" || !r.UnknownSelection.ChartAbsent || !strings.Contains(r.UnknownSelection.Text, "Unknown is not zero") {
		return fail("unknown interval acquired numeric chart")
	}
	if len(r.Facts) != len(clientActions392) {
		return fail("unfinished measured journey")
	}
	for index, f := range r.Facts {
		if f.Name != clientActions392[index] || index > 0 && f.Before != r.Facts[index-1].After {
			return fail("measured action order/counters changed")
		}
		source, stores, configs, captures := sourceCosts[f.Name], 0, 0, 0
		if strings.HasPrefix(f.Name, "service-load-") {
			stores = 2
		}
		switch f.Name {
		case "explicit-source-run", "explicit-destination-run":
			source = 1
		case "save-display-configuration":
			configs = 1
		case "capture-source-evidence":
			source = 1
			captures = 1
			stores = 2
		case "inspect-apply-destination":
			stores = 1
		}
		if f.After.Source-f.Before.Source != source || f.After.Store-f.Before.Store != stores || f.After.ConfigWrites-f.Before.ConfigWrites != configs || f.After.CaptureWrites-f.Before.CaptureWrites != captures || source == 0 && f.After.SQL != f.Before.SQL {
			return fail("local/store/configuration/source authority widened")
		}
	}
	actor := r.Actor.Complete
	if r.Actor.Snapshots <= 0 || actor.Version != "goo.nativeHTTPClient392Actor/1" || actor.Phase != "complete" || actor.SelectedKey != "1" || string(actor.Chart) != "null" || len(actor.Points) < 2 || actor.Points[1].State != "unknown" || actor.Overflow || len(actor.Requests) != 1 || r.Actor.SQLAfter-r.Actor.SQLBefore != 1 {
		return fail("positive actual actor outcome absent")
	}
	request := actor.Requests[0]
	if request.Path != "/api/measurements" || !reflect.DeepEqual(request.Request, r.DestinationQuestion) {
		return fail("actor performed a foreign read")
	}
	if _, e := admitClientRead392(entryRead390{Name: "actor", Request: request.Request, Response: request.Response, Status: request.Status}, r.RawHistograms); e != nil {
		return e
	}
	if r.Actor.SQLBefore < 0 || r.Actor.SQLAfter > len(r.ProxyCalls) {
		return fail("actor physical source boundary invalid")
	}
	plan, e := measurements.Compile(r.DestinationQuestion)
	if e != nil {
		return e
	}
	call := r.ProxyCalls[r.Actor.SQLBefore]
	if call.SQL != plan.SQL || call.Status != 200 || call.Parameters["readonly"] != "1" {
		return fail("actor SQL/read-only response departed from exact source plan")
	}
	for key, value := range plan.Parameters {
		if call.Parameters["param_"+key] != value {
			return fail("actor physical typed parameter changed")
		}
	}
	return admitClientPhysical392(r)
}

func TestQuintHTTPClientNativeObserved392(t *testing.T) {
	path := os.Getenv("GOO_NATIVE_HTTP_CLIENT_392_REPORT")
	if path == "" {
		t.Skip("actual native HTTP-client journey required")
	}
	wire, e := os.ReadFile(path)
	if e != nil {
		t.Fatal(e)
	}
	var r clientLedger392
	if e = json.Unmarshal(wire, &r); e != nil {
		t.Fatal(e)
	}
	if e = admitClientLedger392(r); e != nil {
		t.Fatal(e)
	}
	if !replayClient392(t, r, false) {
		t.Fatal("actual measured request ledger violates model")
	}
	mutations := []func(*clientLedger392){
		func(r *clientLedger392) { r.RawHistograms[0].MetricUnit = "ms" },
		func(r *clientLedger392) { r.Captures[0].StoredBytesAfter += " " },
		func(r *clientLedger392) {
			r.DestinationQuestion.Definition.Source.ExactStream.Resource["service.instance.id"] = "foreign"
		},
		func(r *clientLedger392) { r.UnknownSelection.ChartAbsent = false },
		func(r *clientLedger392) { r.Facts[1].After.Source++ },
		func(r *clientLedger392) { r.DestinationSnapshot.Snapshot.CapturedAt = "2026-01-01T00:00:00Z" },
	}
	for index, mutate := range mutations {
		var bad clientLedger392
		json.Unmarshal(wire, &bad)
		mutate(&bad)
		if admitClientLedger392(bad) == nil {
			t.Fatalf("semantic mutant%d escaped", index)
		}
	}
	if replayClient392(t, r, true) {
		t.Fatal("local source-read model control escaped")
	}
	t.Logf("admitted%d native exports,%d raw spans,%d main facts; six semantic mutants and local-read control rejected", len(r.RawHistograms), len(r.RawSpans), len(r.Facts))
}

func replayClient392(t *testing.T, r clientLedger392, mutate bool) bool {
	t.Helper()
	if len(r.Facts) != len(clientActions392) {
		return false
	}
	reads, stores, configs, captures := make([]int, 15), make([]int, 15), make([]int, 15), make([]int, 15)
	for index, name := range clientActions392 {
		if strings.HasPrefix(name, "service-load-") {
			role := strings.TrimPrefix(name, "service-load-")
			for _, c := range r.Cases {
				if c.Role == role {
					reads[index] = 5 + c.WidgetReads
				}
			}
			stores[index] = 2
		}
		switch name {
		case "explicit-client-span-lens":
			for _, c := range r.Cases {
				if c.Role == "spansOnly" {
					reads[index] = len(c.SpanLens.Reads)
				}
			}
		case "explicit-source-run", "explicit-destination-run":
			reads[index] = 1
		case "save-display-configuration":
			configs[index] = 1
		case "capture-source-evidence":
			reads[index] = 1
			captures[index] = 1
			stores[index] = 2
		case "inspect-apply-destination":
			stores[index] = 1
		}
	}
	sink := &qobs.MemorySink{}
	rec := qobs.NewRecorder(sink)
	for index, f := range r.Facts {
		if f.Name != clientActions392[index] {
			return false
		}
		source := f.After.Source
		if mutate && f.Name == "service-retained-local" {
			source++
		}
		action := "step"
		if f.Name == "complete" {
			action = "finish"
		}
		rec.Record(context.Background(), action, nil, "obs.phase", index+1, "obs.reads", source, "obs.stores", f.After.Store, "obs.configs", f.After.ConfigWrites, "obs.captures", f.After.CaptureWrites, "obs.finished", f.Name == "complete")
	}
	spec, e := filepath.Abs("model/http_client_native392.qnt")
	if e != nil {
		t.Fatal(e)
	}
	list := func(v []int) string { wire, _ := json.Marshal(v); return string(wire) }
	binding := fmt.Sprintf("spec: %q\nmodule: http_client_native392\nconstants: {READS: '%s', STORES: '%s', CONFIGS: '%s', CAPTURES: '%s'}\ninvariants: [safety]\nactions:\n", spec, list(reads), list(stores), list(configs), list(captures))
	for _, name := range []string{"step", "finish"} {
		binding += "  " + name + ": {expect: 'phase == {{.obs.phase}} and reads == {{.obs.reads}} and stores == {{.obs.stores}} and configs == {{.obs.configs}} and captures == {{.obs.captures}} and finished == {{.obs.finished}}'}\n"
	}
	dir := t.TempDir()
	file := filepath.Join(dir, "binding.yaml")
	if e = os.WriteFile(file, []byte(binding), 0600); e != nil {
		t.Fatal(e)
	}
	checker, e := validate.NewChecker(file, validate.Options{Quint: quintPath(t), Backend: "typescript", Dir: dir})
	if e != nil {
		t.Fatal(e)
	}
	result, e := checker.Check(t.Context(), sink.Steps())
	if e != nil {
		t.Fatal(e)
	}
	label := "actual"
	if mutate {
		label = "local-read-mutant"
	}
	steps, _ := json.Marshal(sink.Steps())
	if e = artifact("native-http-client392-"+label+"-steps.json", steps); e != nil {
		t.Fatal(e)
	}
	if e = artifact("native-http-client392-"+label+"-model.txt", []byte(result.Summary())); e != nil {
		t.Fatal(e)
	}
	return result.OK()
}

func admitClientPhysical392(r clientLedger392) error {
	expected := map[string]int{}
	knownSQL := map[string]bool{}
	key := func(sql string, parameters map[string]string) string {
		wire, _ := json.Marshal(parameters)
		return sql + "\x00" + string(wire)
	}
	add := func(q measurements.Question) error {
		plan, e := measurements.Compile(q)
		if e != nil {
			return e
		}
		knownSQL[plan.SQL] = true
		expected[key(plan.SQL, plan.Parameters)]++
		return nil
	}
	for _, read := range r.Reads {
		if e := add(read.Request); e != nil {
			return e
		}
	}
	for _, call := range r.Calls {
		if call.Method == "POST" && call.Path == "/api/measurements" {
			var q measurements.Question
			if e := json.Unmarshal(call.Body, &q); e != nil {
				return e
			}
			if e := add(q); e != nil {
				return e
			}
		}
	}
	// Native capture is one explicit fresh source read, separate from UI Run.
	if e := add(r.SourceQuestion); e != nil {
		return e
	}
	if e := add(r.DestinationQuestion); e != nil {
		return e
	} // independent actor
	for _, call := range r.ProxyCalls {
		if !knownSQL[call.SQL] {
			continue
		}
		if call.Status != 200 || call.Parameters["readonly"] != "1" {
			return fmt.Errorf("native392 physical source status/readonly changed")
		}
		params := map[string]string{}
		for name, value := range call.Parameters {
			if strings.HasPrefix(name, "param_") {
				params[strings.TrimPrefix(name, "param_")] = value
			}
		}
		k := key(call.SQL, params)
		if expected[k] <= 0 {
			return fmt.Errorf("native392 physical SQL/typed parameters contain unadmitted read")
		}
		expected[k]--
	}
	for _, remaining := range expected {
		if remaining != 0 {
			return fmt.Errorf("native392 physical source read missing")
		}
	}
	return nil
}

// Public chart import deliberately strips source guarantees. Permit only that
// existing transport normalization; selectors, parent basis, clocks and all
// supplied values still have to equal the independently admitted projection.
func passiveClientView392(expected *scopedview.Dataset, actual scopedview.Dataset) bool {
	if expected == nil || !strings.HasPrefix(actual.ID, "goo.local/") || actual.Revision != actual.ID || actual.Origin != "browser-held" || actual.Completeness != nil || actual.Quality.Extent != "unknown" {
		return false
	}
	token := strings.TrimPrefix(actual.ID, "goo.local/")
	number, e := strconv.Atoi(token)
	if e != nil || number < 1 || strconv.Itoa(number) != token {
		return false
	}
	wire, _ := json.Marshal(expected)
	var normalized scopedview.Dataset
	if json.Unmarshal(wire, &normalized) != nil {
		return false
	}
	normalized.ID = actual.ID
	normalized.Revision = actual.Revision
	normalized.Origin = "browser-held"
	normalized.Completeness = nil
	normalized.Quality.Extent = "unknown"
	return sameTransport384(normalized, actual)
}
