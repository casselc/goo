package verification

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"slices"
	"testing"
	"time"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/measurements"
	"github.com/casselc/observability/quintgo/qobs"
	"github.com/casselc/observability/quintgo/validate"
)

type rpcCapabilityCounter362 struct{ Calls, Source, Plans, SQL, Store int }
type rpcCapabilityFact362 struct {
	Name          string
	Before, After rpcCapabilityCounter362
}
type rpcCapabilityLedger362 struct {
	Version string
	Fixture struct {
		Token, Environment, Service, Namespace string
		Scope                                  exploration.InventoryQuestion
		Metrics                                map[string]any
	}
	Plan                                exploration.DashboardPlan
	ServiceDiscovery, MeasuredDiscovery measurements.DiscoveryResult
	Selected                            measurements.Instrument
	Build                               struct {
		Request  json.RawMessage
		Response struct {
			Definition measurements.Definition
			Question   measurements.Question
		}
	}
	Run struct {
		Request  measurements.Question
		Response measurements.Result
	}
	HealthyBefore, HealthyAfter []json.RawMessage
	Facts                       []rpcCapabilityFact362
	Calls                       []struct {
		Path, Method   string
		Body, Response json.RawMessage
		Status         int
	}
	Errors        []string
	FrozenCapture struct {
		Request                                                struct{ Question exploration.CaptureQuestion }
		Created, Reopened                                      exploration.CapturePage
		StoredBytesBefore, StoredBytesAfter, GetRequestURL     string
		BeforeCapture, AfterCapture, BeforeReopen, AfterReopen rpcCapabilityCounter362
		SourceRefusalProbe                                     struct {
			Status   int
			Response struct{ Refusal struct{ Kind string } }
		}
	}
}

func rpcObject362(v any) map[string]any { o, _ := v.(map[string]any); return o }
func rpcArray362(v any) []any           { a, _ := v.([]any); return a }
func rpcLabels362(v any) map[string]string {
	out := map[string]string{}
	for _, raw := range rpcArray362(v) {
		o := rpcObject362(raw)
		key, _ := o["key"].(string)
		text, _ := rpcObject362(o["value"])["stringValue"].(string)
		out[key] = text
	}
	return out
}
func rpcRawPoints362(l rpcCapabilityLedger362, i measurements.Instrument) ([]measurements.DiscoveryObservation, []map[string]any) {
	observations := []measurements.DiscoveryObservation{}
	points := []map[string]any{}
	for _, rawResource := range rpcArray362(l.Fixture.Metrics["resourceMetrics"]) {
		resource := rpcObject362(rawResource)
		labels := rpcLabels362(rpcObject362(resource["resource"])["attributes"])
		resourceSchema, _ := resource["schemaUrl"].(string)
		for _, rawScope := range rpcArray362(resource["scopeMetrics"]) {
			scope := rpcObject362(rawScope)
			identity := rpcObject362(scope["scope"])
			if identity["name"] != i.Scope || identity["version"] != i.Version || scope["schemaUrl"] != i.SchemaURL {
				continue
			}
			for _, rawMetric := range rpcArray362(scope["metrics"]) {
				metric := rpcObject362(rawMetric)
				if metric["name"] != i.Metric || metric["unit"] != i.Unit {
					continue
				}
				kind := rpcObject362(metric[i.Kind])
				if kind == nil {
					continue
				}
				if i.Kind == "histogram" && kind["aggregationTemporality"] != float64(i.Temporality) {
					continue
				}
				for _, rawPoint := range rpcArray362(kind["dataPoints"]) {
					point := rpcObject362(rawPoint)
					observations = append(observations, measurements.DiscoveryObservation{Resource: labels, Attributes: rpcLabels362(point["attributes"]), ResourceSchemaURL: resourceSchema})
					points = append(points, point)
				}
			}
		}
	}
	return observations, points
}
func rpcHistogramCheck362(l rpcCapabilityLedger362, r measurements.Result, q measurements.Question) error {
	if savedJSON360(r.Question) != savedJSON360(q) || r.Histograms == nil || len(r.Histograms.Series) != 4 || r.State != "unknown" {
		return fmt.Errorf("exact independently evaluated delta streams and unknown window coverage")
	}
	observations, points := rpcRawPoints362(l, l.Selected)
	expected := map[string]int{}
	for k, o := range observations {
		expected[rpcObservationKey362(o.Resource, o.Attributes, o.ResourceSchemaURL)] = k
	}
	seen := map[string]bool{}
	for _, s := range r.Histograms.Series {
		key := rpcObservationKey362(s.Stream.Resource, s.Stream.Attributes, s.Stream.ResourceSchemaURL)
		index, ok := expected[key]
		if !ok || seen[key] || s.Stream.Scope != l.Selected.Scope || s.Stream.Name != "rpc.server.call.duration" || s.Stream.Unit != "s" || s.Stream.SchemaURL != l.Selected.SchemaURL || s.Temporality != "delta" || len(s.Intervals) != 1 {
			return fmt.Errorf("native stream identity/units/temporality")
		}
		seen[key] = true
		p := points[index]
		interval := s.Intervals[0]
		d := interval.Distribution
		if interval.State != "ok" || interval.FromNS != p["startTimeUnixNano"] || interval.ToNS != p["timeUnixNano"] || d == nil || d.Unit != "s" || d.Count != p["count"] || savedJSON360(d.Bounds) != savedJSON360(p["explicitBounds"]) || savedJSON360(d.BucketCounts) != savedJSON360(p["bucketCounts"]) {
			return fmt.Errorf("raw OTLP retained seconds histogram counts/buckets")
		}
	}
	return nil
}
func rpcCapabilityAdmission362(l rpcCapabilityLedger362) error {
	if l.Version != "goo.currentRpcCapability362/1" || len(l.Errors) != 0 || len(l.Facts) != 6 || len(l.HealthyBefore) < 2 || savedJSON360(l.HealthyBefore) != savedJSON360(l.HealthyAfter) {
		return fmt.Errorf("complete native visit and retained service DOM/receipts")
	}
	scope := l.Fixture.Scope
	if l.Plan.Profile != "rpc-server-current" || l.Plan.Service != l.Fixture.Service || savedJSON360(l.Plan.Scope.EnvironmentSelector) != savedJSON360(scope.EnvironmentSelector) || !l.Plan.Scope.Namespace.Equal(scope.Namespace) || l.Plan.Scope.FromMS != scope.FromMS || l.Plan.Scope.ToMS != scope.ToMS {
		return fmt.Errorf("typed parent fixed service scope")
	}
	d := l.MeasuredDiscovery
	if savedJSON360(d) != savedJSON360(l.ServiceDiscovery) || d.Question.Environment != l.Fixture.Environment || d.Question.EnvironmentSelector != nil || d.Question.Service != l.Fixture.Service || !d.Question.Namespace.Equal(scope.Namespace) || d.Question.FromMS != scope.FromMS || d.Question.ToMS != scope.ToMS || len(d.Instruments) != 7 {
		return fmt.Errorf("same checked scalar-native discovery under typed current parent")
	}
	selectedFound := false
	for _, i := range d.Instruments {
		if i.CapabilityPlan == nil {
			return fmt.Errorf("missing capability metadata")
		}
		p := *i.CapabilityPlan
		shape := i.Metric == "rpc.server.call.duration" && i.Kind == "histogram" && i.Unit == "s"
		if p.Verification != "not-run" || p.Eligible != (i.Supported && i.Definition != nil) {
			return fmt.Errorf("metadata promoted evaluation or blocked generic eligibility")
		}
		if i.Metric == "rpc.server.call.duration" {
			if p.Profile != "rpc-server-call-duration" || p.RuleSet != "OpenTelemetry semantic conventions v1.45.0" || p.ShapeMatches != shape {
				return fmt.Errorf("literal current RPC shape/profile pin")
			}
			observations, _ := rpcRawPoints362(l, i)
			if len(observations) != i.Records {
				return fmt.Errorf("raw OTLP discovery records")
			}
			if err := rpcCoverageCheck362(i, observations, p); err != nil {
				return err
			}
		} else if p.Profile != "generic" {
			return fmt.Errorf("legacy/custom implicitly aliased")
		}
		if i.Scope == "goo.native.rpc362.delta" {
			selectedFound = savedJSON360(i) == savedJSON360(l.Selected)
		}
	}
	if !selectedFound || l.Selected.Definition == nil || !l.Selected.Supported {
		return fmt.Errorf("actual selected delta binding")
	}
	native := 0
	for _, w := range l.Plan.Widgets {
		if w.Measurement == nil || w.State != "supported" {
			continue
		}
		native++
		definition := w.Measurement.Definition
		if definition.Kind != "histogram" || definition.Source.Metric != "rpc.server.call.duration" || definition.Source.Unit != "s" || !slices.Contains([]string{"goo.native.rpc362.delta", "goo.native.rpc362.cumulative"}, definition.Source.Scope) {
			return fmt.Errorf("Service current RPC native eligibility disagrees")
		}
	}
	if native != 2 {
		return fmt.Errorf("Services current histograms only, generic bindings still listed")
	}
	q := l.Build.Response.Question
	if savedJSON360(l.Build.Response.Definition) != savedJSON360(l.Selected.Definition) || savedJSON360(q.Definition) != savedJSON360(l.Selected.Definition) || q.FromMS != scope.FromMS || q.ToMS != scope.ToMS || q.Definition.Environment != l.Fixture.Environment || q.Definition.EnvironmentSelector != nil || !q.Definition.Source.Namespace.Equal(scope.Namespace) || savedJSON360(q.Definition.Source.Services) != savedJSON360([]string{l.Fixture.Service}) || savedJSON360(l.Run.Request) != savedJSON360(q) {
		return fmt.Errorf("unaltered selected generic definition and fixed scalar-native question")
	}
	if err := rpcHistogramCheck362(l, l.Run.Response, q); err != nil {
		return err
	}
	names := []string{"discover-stage", "discover", "inspect", "build", "run", "return"}
	for index, f := range l.Facts {
		if f.Name != names[index] || f.Before.Calls < 0 || f.After.Calls < f.Before.Calls || f.After.Calls > len(l.Calls) {
			return fmt.Errorf("actual fact boundaries")
		}
		sources, plans := 0, 0
		for _, call := range l.Calls[f.Before.Calls:f.After.Calls] {
			if call.Path == "/api/measurement-packages/plan" {
				plans++
				if savedJSON360(call.Body) != savedJSON360(l.Build.Request) {
					return fmt.Errorf("actual metadata build request")
				}
			}
			if serviceTemplateSource356(serviceContinuationCall352{Path: call.Path, Method: call.Method}) {
				sources++
				if index == 1 {
					var request measurements.DiscoveryQuestion
					if call.Path != "/api/metric-discovery" || json.Unmarshal(call.Body, &request) != nil || savedJSON360(request) != savedJSON360(d.Question) {
						return fmt.Errorf("actual exact discovery request")
					}
				} else if index == 4 {
					var request measurements.Question
					if call.Path != "/api/measurements" || json.Unmarshal(call.Body, &request) != nil || savedJSON360(request) != savedJSON360(q) {
						return fmt.Errorf("actual explicit native Run")
					}
				} else {
					return fmt.Errorf("passive metadata inspection acquired telemetry")
				}
			}
		}
		want, metadata := 0, 0
		if index == 1 || index == 4 {
			want = 1
		}
		if index == 3 {
			metadata = 1
		}
		if sources != want || plans != metadata || f.After.Source-f.Before.Source != want || f.After.Plans-f.Before.Plans != metadata || want == 0 && f.After.SQL != f.Before.SQL {
			return fmt.Errorf("explicit reads separate from metadata plan/inspection")
		}
	}
	frozen := l.FrozenCapture
	if frozen.Created.Measurement == nil || frozen.Reopened.Measurement == nil || frozen.Request.Question.Kind != "measurement" || savedJSON360(frozen.Request.Question.Measurement) != savedJSON360(q) || savedJSON360(frozen.Created.Receipt.Question.Measurement) != savedJSON360(q) || savedJSON360(frozen.Created) != savedJSON360(frozen.Reopened) || frozen.StoredBytesBefore == "" || frozen.StoredBytesBefore != frozen.StoredBytesAfter {
		return fmt.Errorf("frozen measurement question/bytes changed")
	}
	var stored exploration.SavedCapture
	if json.Unmarshal([]byte(frozen.StoredBytesBefore), &stored) != nil || stored.Measurement == nil || savedJSON360(stored.CaptureReceipt) != savedJSON360(frozen.Created.Receipt) || savedJSON360(stored.Measurement) != savedJSON360(frozen.Created.Measurement) || !peLiveCaptureHash(exploration.CapturePage{Receipt: stored.CaptureReceipt, Measurement: stored.Measurement}) {
		return fmt.Errorf("own immutable native capture hash")
	}
	read, err := time.Parse(time.RFC3339Nano, frozen.Created.Measurement.ReadAt)
	if err != nil {
		return err
	}
	custody, err := time.Parse(time.RFC3339Nano, frozen.Created.Receipt.CapturedAt)
	if err != nil || custody.Before(read) {
		return fmt.Errorf("capture custody precedes its independent source read")
	}
	if err = rpcHistogramCheck362(l, *frozen.Created.Measurement, q); err != nil {
		return err
	}
	if frozen.AfterCapture.SQL <= frozen.BeforeCapture.SQL || frozen.SourceRefusalProbe.Status != 502 || frozen.SourceRefusalProbe.Response.Refusal.Kind != "unavailable" || frozen.AfterReopen.Source != frozen.BeforeReopen.Source || frozen.AfterReopen.SQL != frozen.BeforeReopen.SQL || frozen.AfterReopen.Store-frozen.BeforeReopen.Store != 1 {
		return fmt.Errorf("own capture read then source-refused metadata-only reopen")
	}
	u, err := url.Parse(frozen.GetRequestURL)
	if err != nil || u.Path != "/api/captures/"+frozen.Created.Receipt.ID {
		return fmt.Errorf("actual frozen capture GET")
	}
	return nil
}
func rpcCapabilityModel362(t *testing.T, label string, inspectReads int, coherent bool) bool {
	t.Helper()
	sink := &qobs.MemorySink{}
	rec := qobs.NewRecorder(sink, qobs.WithActor(label))
	rec.Record(context.Background(), "discover", nil, "obs.coherent", true, "obs.reads", 1, "obs.effects", 0)
	rec.Record(context.Background(), "inspect", nil, "obs.coherent", coherent, "obs.reads", inspectReads, "obs.effects", 0)
	checker, e := validate.NewChecker("model/instrument_presence.yaml", validate.Options{Quint: quintPath(t), Backend: "typescript", Dir: t.TempDir()})
	if e != nil {
		t.Fatal(e)
	}
	r, e := checker.Check(context.Background(), sink.Steps())
	if e != nil {
		t.Fatal(e)
	}
	artifact("rpc-capability362-"+label+".txt", []byte(r.Summary()))
	return r.OK()
}
func TestQuintCurrentRPCCapabilityProtocol362(t *testing.T) {
	if !rpcCapabilityModel362(t, "protocol-complete", 1, true) {
		t.Fatal("actual presence wrapper refused")
	}
	if rpcCapabilityModel362(t, "protocol-implicit-read", 2, true) {
		t.Fatal("inspection acquired telemetry")
	}
	if rpcCapabilityModel362(t, "protocol-forged-presence", 1, false) {
		t.Fatal("incoherent presence accepted")
	}
}
func TestQuintCurrentRPCCapabilityObserved362(t *testing.T) {
	path := os.Getenv("GOO_CURRENT_RPC_CAPABILITY_362_REPORT")
	if path == "" {
		t.Skip("explicit native current RPC capability report required")
	}
	raw, e := os.ReadFile(path)
	if e != nil {
		t.Fatal(e)
	}
	var l rpcCapabilityLedger362
	if json.Unmarshal(raw, &l) != nil {
		t.Fatal("invalid native capability report")
	}
	if e = rpcCapabilityAdmission362(l); e != nil {
		t.Fatal(e)
	}
	reads := l.Facts[2].After.Source - l.Facts[1].Before.Source
	if !rpcCapabilityModel362(t, "actual-discover-inspect", reads, true) {
		t.Fatal("actual discover/inspect trace refused")
	}
	mutants := []func(*rpcCapabilityLedger362){func(l *rpcCapabilityLedger362) { l.Selected.CapabilityPlan.Coverage.TotalStreams++ }, func(l *rpcCapabilityLedger362) { l.Selected.CapabilityPlan.Verification = "verified" }, func(l *rpcCapabilityLedger362) { l.Facts[2].After.Source++ }, func(l *rpcCapabilityLedger362) { l.FrozenCapture.Created.Measurement.Question.FromMS++ }, func(l *rpcCapabilityLedger362) {
		l.FrozenCapture.Created.Measurement.Histograms.Series[0].Stream.Unit = "ms"
	}, func(l *rpcCapabilityLedger362) { l.FrozenCapture.Reopened.Measurement.ReadAt = "2020-01-01T00:00:00Z" }}
	for i, mutate := range mutants {
		var fresh rpcCapabilityLedger362
		json.Unmarshal(raw, &fresh)
		mutate(&fresh)
		if rpcCapabilityAdmission362(fresh) == nil {
			t.Fatalf("native negative%d admitted", i)
		}
	}
}
