package measurements

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strings"
	"testing"
	"time"

	"gopkg.in/yaml.v3"
)

func podMetricQuestion() PodMetricDiscoveryQuestion {
	schema := ""
	return PodMetricDiscoveryQuestion{Environment: "prod", FromMS: 100000, ToMS: 160000, Resource: map[string]string{"deployment.environment.name": "prod", "k8s.pod.uid": "pod-one", "k8s.pod.name": "same-name"}, ResourceSchemaURL: &schema}
}
func podMetadata(q PodMetricDiscoveryQuestion, metric, scope, version, schema string) json.RawMessage {
	return raw(map[string]any{"kind": "gauge", "scope": scope, "version": version, "schemaURL": schema, "metric": metric, "unit": "By", "temporality": 0, "monotonic": false, "resource": q.Resource, "resourceSchemaURL": *q.ResourceSchemaURL, "attributes": map[string]string{}})
}
func podDiscoveryFixture(t *testing.T) PodMetricDiscovery {
	t.Helper()
	q := podMetricQuestion()
	r, e := discoverPodMetricRows(q, []json.RawMessage{podMetadata(q, "k8s.pod.memory.usage", KubeletStatsScope, KubeletStatsVersion, ""), podMetadata(q, "k8s.pod.memory.working_set", KubeletStatsScope, KubeletStatsVersion, "")})
	if e != nil {
		t.Fatal(e)
	}
	return *r
}
func TestPodMetricDiscoveryAndExactNativeReplay(t *testing.T) {
	discovery := podDiscoveryFixture(t)
	plan, e := PlanPodDashboard(PodDashboardRequest{Discovery: discovery})
	if e != nil || len(plan.Widgets) != 2 {
		t.Fatal(plan, e)
	}
	for _, widget := range plan.Widgets {
		q := widget.Question
		stream := q.Definition.Source.ExactStream
		if q.Definition.Source.Resource.Kind != "kubernetes-pod" || len(q.Definition.Source.Services) != 0 || stream.Resource["service.name"] != "" {
			t.Fatal(q)
		}
		compiled, e := Compile(q)
		if e != nil || strings.Contains(compiled.SQL, "mapUpdate") || strings.Contains(compiled.SQL, "ServiceName IN") || !strings.Contains(compiled.SQL, "ResourceAttributes['k8s.pod.uid']") {
			t.Fatal(compiled, e)
		}
		row := metricRow{Resource: stream.Resource, ResourceSchemaURL: stream.ResourceSchemaURL, Scope: stream.Scope, Version: stream.Version, SchemaURL: stream.SchemaURL, Name: stream.Name, Unit: stream.Unit, Attributes: stream.Attributes, EndNS: ns(120000), Value: "2048"}
		r, e := Evaluate(q, []json.RawMessage{raw(row)}, time.Unix(200, 0))
		if e != nil || r.Gauges == nil {
			t.Fatal(r, e)
		}
		if e = ValidateResult(r); e != nil {
			t.Fatal(e)
		}
		encoded, e := yaml.Marshal(q.Definition)
		if e != nil {
			t.Fatal(e)
		}
		d, e := Read(string(encoded))
		if e != nil || Hash(d) != Hash(q.Definition) {
			t.Fatal("pod YAML identity drift", e)
		}
		if _, present := r.Gauges.Series[0].Stream.Resource["service.name"]; present {
			t.Fatal("fake service inserted")
		}
		row.Resource = cloneLabels(row.Resource)
		row.Resource["k8s.pod.uid"] = "other"
		if _, e = Evaluate(q, []json.RawMessage{raw(row)}, time.Unix(200, 0)); e == nil {
			t.Fatal("different pod accepted")
		}
		bad := q
		bad.Definition.Kind = "stock"
		if bad.Validate() == nil {
			t.Fatal("pod stock authority relaxed")
		}
	}
}
func TestPodProfilesRemainDistinctAndDeterministic(t *testing.T) {
	q := podMetricQuestion()
	rows := []json.RawMessage{podMetadata(q, "k8s.pod.memory.usage", "custom-scope", "actual-version", PodMetricsSchema), podMetadata(q, "k8s.pod.memory.working_set", "custom-scope", "actual-version", PodMetricsSchema)}
	discovery, e := discoverPodMetricRows(q, rows)
	if e != nil {
		t.Fatal(e)
	}
	plan, e := PlanPodDashboard(PodDashboardRequest{Discovery: *discovery})
	if e != nil || len(plan.Widgets) != 1 || plan.Widgets[0].Profile != PodSemconvProfile {
		t.Fatal(plan, e)
	}
	for _, version := range []string{"v0.137.0", "0.136.0"} {
		discovery, e := discoverPodMetricRows(q, []json.RawMessage{podMetadata(q, "k8s.pod.memory.usage", KubeletStatsScope, version, "")})
		if e != nil {
			t.Fatal(e)
		}
		plan, e := PlanPodDashboard(PodDashboardRequest{Discovery: *discovery})
		if e != nil || len(plan.Candidates) != 0 {
			t.Fatal("version alias", plan, e)
		}
	}
	discovery2 := podDiscoveryFixture(t)
	original, e := PlanPodDashboard(PodDashboardRequest{Discovery: discovery2})
	if e != nil {
		t.Fatal(e)
	}
	duplicate := discovery2.Instruments[0]
	d := *duplicate.Definition
	d.Name = "alias-name"
	d.Title = "Alias title"
	duplicate.Definition = &d
	discovery2.Instruments = append(discovery2.Instruments, duplicate)
	again, e := PlanPodDashboard(PodDashboardRequest{Discovery: discovery2})
	if e != nil || !reflect.DeepEqual(original, again) {
		t.Fatal("package labels multiplied source identity", e)
	}
}
func TestPodSourceParserRefusesUnknownSchemaAndMalformedMetadata(t *testing.T) {
	q := podMetricQuestion()
	good := podMetadata(q, "k8s.pod.memory.usage", KubeletStatsScope, KubeletStatsVersion, "")
	for _, key := range []string{"resourceSchemaURL", "schemaURL", "version", "scope", "kind", "unit", "temporality", "monotonic", "resource", "attributes"} {
		var object map[string]json.RawMessage
		json.Unmarshal(good, &object)
		delete(object, key)
		if _, e := discoverPodMetricRows(q, []json.RawMessage{raw(object)}); e == nil {
			t.Fatal("omitted source field accepted", key)
		}
		json.Unmarshal(good, &object)
		object[key] = json.RawMessage("null")
		if _, e := discoverPodMetricRows(q, []json.RawMessage{raw(object)}); e == nil {
			t.Fatal("unknown source field accepted as empty", key)
		}
	}
	unknown := q
	unknown.ResourceSchemaURL = nil
	if unknown.Validate() == nil {
		t.Fatal("trace unknown schema became reportedempty")
	}
	emptyScope := podMetadata(q, "k8s.pod.memory.usage", "", "actual-version", PodMetricsSchema)
	discovery, e := discoverPodMetricRows(q, []json.RawMessage{emptyScope})
	if e != nil {
		t.Fatal(e)
	}
	plan, e := PlanPodDashboard(PodDashboardRequest{Discovery: *discovery})
	if e != nil || len(plan.Widgets) != 0 {
		t.Fatal("scope omission admits failed default", plan, e)
	}
}
func TestPodMetricSourceOneExactReadAndCancel(t *testing.T) {
	q := podMetricQuestion()
	reads := 0
	source := Source{Endpoint: "http://source.invalid", Client: &http.Client{Transport: nativeWireTransport(func(r *http.Request) (*http.Response, error) {
		reads++
		b, _ := io.ReadAll(r.Body)
		sql := string(b)
		if strings.Count(sql, " UNION ALL ") != 2 || strings.Contains(sql, "mapUpdate") || !strings.Contains(sql, "ResourceSchemaUrl={resourceSchema:String}") || !strings.Contains(sql, "LIMIT 10001") {
			t.Fatal(sql)
		}
		return &http.Response{StatusCode: 200, Header: http.Header{}, Body: io.NopCloser(strings.NewReader(string(podMetadata(q, "k8s.pod.memory.usage", KubeletStatsScope, KubeletStatsVersion, ""))))}, nil
	})}}
	result, e := source.DiscoverPodMetrics(context.Background(), q)
	if e != nil || reads != 1 || len(result.Instruments) != 1 {
		t.Fatal(result, e, reads)
	}
	q.ResourceSchemaURL = nil
	if _, e := source.DiscoverPodMetrics(context.Background(), q); e == nil || reads != 1 {
		t.Fatal("unknownschema reachedsource", e, reads)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	q = podMetricQuestion()
	if _, e := source.DiscoverPodMetrics(ctx, q); e == nil || reads != 1 {
		t.Fatal("cancelled source", e, reads)
	}
}

func TestPodDashboardAllPreviewStreamsAndDefaultsBeforeCap(t *testing.T) {
	discovery := podDiscoveryFixture(t)
	for index := range discovery.Instruments {
		instrument := &discovery.Instruments[index]
		if instrument.Metric != "k8s.pod.memory.usage" {
			continue
		}
		original := instrument.CapabilityPlan.Coverage.Streams[0]
		instrument.CapabilityPlan.Coverage.Streams = nil
		for n := 0; n < 20; n++ {
			preview := original
			preview.Stream.Attributes = map[string]string{"reported.dimension": fmt.Sprint(n)}
			instrument.CapabilityPlan.Coverage.Streams = append(instrument.CapabilityPlan.Coverage.Streams, preview)
		}
	}
	plan, e := PlanPodDashboard(PodDashboardRequest{Discovery: discovery})
	if e != nil || len(plan.Candidates) != 16 || plan.CandidatesOmitted != 5 || len(plan.Widgets) != 2 {
		t.Fatal("preview streams lost or defaults clipped", plan, e)
	}
	families := map[string]bool{}
	attributes := map[string]bool{}
	for _, widget := range plan.Widgets {
		families[widget.Question.Definition.Source.Metric] = true
	}
	for _, candidate := range plan.Candidates {
		if candidate.Question.Definition.Source.Metric == "k8s.pod.memory.usage" {
			attributes[candidate.Question.Definition.Source.ExactStream.Attributes["reported.dimension"]] = true
		}
	}
	if !families["k8s.pod.memory.usage"] || !families["k8s.pod.memory.working_set"] || len(attributes) != 15 {
		t.Fatal("family or exact attribute identities merged", families, attributes)
	}
	// Explicit binding selectors remain constraints when fresh previews are pinned.
	for index := range discovery.Instruments {
		instrument := &discovery.Instruments[index]
		if instrument.Metric == "k8s.pod.memory.usage" {
			instrument.Definition.Source.Attributes = map[string]string{"required": "yes"}
			instrument.Definition.Source.ExactStream.Attributes = map[string]string{"required": "yes"}
		}
	}
	plan, e = PlanPodDashboard(PodDashboardRequest{Discovery: discovery})
	if e != nil {
		t.Fatal(e)
	}
	for _, candidate := range plan.Candidates {
		if candidate.Question.Definition.Source.Metric == "k8s.pod.memory.usage" {
			t.Fatal("explicit source selector was relaxed")
		}
	}
}
