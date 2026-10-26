package verification

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/casselc/goo/measurements"
	"github.com/casselc/goo/metrics"
	"gopkg.in/yaml.v3"
	"hegel.dev/go/hegel"
	"io"
	"net/http"
	"reflect"
	"strings"
	"testing"
	"time"
)

func pnQuestion(uid string) measurements.PodMetricDiscoveryQuestion {
	schema := ""
	return measurements.PodMetricDiscoveryQuestion{Environment: "prod", FromMS: 100000, ToMS: 160000, Resource: map[string]string{"deployment.environment.name": "prod", "k8s.pod.uid": uid, "k8s.cluster.uid": "cluster-a"}, ResourceSchemaURL: &schema}
}
func pnMetadata(q measurements.PodMetricDiscoveryQuestion, metric, scope, version, schema string, attrs map[string]string) json.RawMessage {
	b, _ := json.Marshal(map[string]any{"kind": "gauge", "scope": scope, "version": version, "schemaURL": schema, "metric": metric, "unit": "By", "temporality": 0, "monotonic": false, "resource": q.Resource, "resourceSchemaURL": *q.ResourceSchemaURL, "attributes": attrs})
	return b
}
func pnDiscover(q measurements.PodMetricDiscoveryQuestion, rows []json.RawMessage) (*measurements.PodMetricDiscovery, error) {
	source := measurements.Source{Endpoint: "http://source.invalid", Client: &http.Client{Transport: cvTransport(func(r *http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 200, Header: http.Header{}, Body: io.NopCloser(strings.NewReader(string(joinPDRows(rows))))}, nil
	})}}
	return source.DiscoverPodMetrics(context.Background(), q)
}
func TestHegelPodNativeAuthorityAndRawReplay(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		q := pnQuestion(fmt.Sprintf("uid-%d", hegel.Draw(tc, hegel.Integers(1, 100000))))
		rows := []json.RawMessage{}
		for _, metric := range []string{"k8s.pod.memory.usage", "k8s.pod.memory.working_set"} {
			rows = append(rows, pnMetadata(q, metric, measurements.KubeletStatsScope, "0.137.0", "", map[string]string{}))
		}
		d, e := pnDiscover(q, rows)
		if e != nil {
			tc.Errorf("metadata %v", e)
			return
		}
		plan, e := measurements.PlanPodDashboard(measurements.PodDashboardRequest{Discovery: *d})
		if e != nil || len(plan.Widgets) != 2 {
			tc.Errorf("receiver defaults %v", e)
			return
		}
		for _, w := range plan.Widgets {
			native := w.Question
			s := native.Definition.Source.ExactStream
			if native.Definition.Source.Resource == nil || native.Definition.Source.Resource.Kind != "kubernetes-pod" || native.Definition.Source.Resource.ID != q.Resource["k8s.pod.uid"] || len(native.Definition.Source.Services) != 0 || native.Definition.Source.Namespace != nil || s == nil || !reflect.DeepEqual(s.Resource, q.Resource) || s.ResourceSchemaURL != "" || native.FromMS != q.FromMS || native.ToMS != q.ToMS {
				tc.Errorf("authority or exact scope drift")
				return
			}
			for _, value := range []string{"0", "2048", "-1", "0.5", "9007199254740992"} {
				row := hrRow(native, value)
				row["temporality"] = 0
				raw, _ := json.Marshal(row)
				r, e := measurements.Evaluate(native, []json.RawMessage{raw}, time.Unix(200, 0), "", nil)
				if e != nil || measurements.ValidateResult(r) != nil || !reflect.DeepEqual(r.Gauges.Series[0].Stream, *s) {
					tc.Errorf("raw gauge evidence filtered/fabricated %s %v", value, e)
					return
				}
			}
			for _, key := range []string{"uid", "cluster", "schema", "service"} {
				var bad measurements.Question
				raw, _ := json.Marshal(native)
				_ = json.Unmarshal(raw, &bad)
				row := hrRow(bad, "2048")
				row["temporality"] = 0
				switch key {
				case "uid":
					resource := map[string]string{}
					for k, v := range s.Resource {
						resource[k] = v
					}
					resource["k8s.pod.uid"] = "other"
					row["resource"] = resource
				case "cluster":
					resource := map[string]string{}
					for k, v := range s.Resource {
						resource[k] = v
					}
					resource["k8s.cluster.uid"] = "other"
					row["resource"] = resource
				case "schema":
					row["resource_schema_url"] = "other"
				case "service":
					bad.Definition.Source.Services = []string{"fake"}
				}
				raw, _ = json.Marshal(row)
				if _, e := measurements.Evaluate(bad, []json.RawMessage{raw}, time.Unix(200, 0), "", nil); e == nil {
					tc.Errorf("crossover/raw mismatch accepted %s", key)
					return
				}
			}
		}
	}, options(10)...)
}
func TestPodNativeProfilesPreviewCapAndAuthoringCustody(t *testing.T) {
	q := pnQuestion("one")
	rows := []json.RawMessage{}
	for n := 0; n < 20; n++ {
		rows = append(rows, pnMetadata(q, "k8s.pod.memory.usage", measurements.KubeletStatsScope, "0.137.0", "", map[string]string{"variant": fmt.Sprint(n)}))
	}
	rows = append(rows, pnMetadata(q, "k8s.pod.memory.working_set", measurements.KubeletStatsScope, "0.137.0", "", map[string]string{}))
	d, e := pnDiscover(q, rows)
	if e != nil {
		t.Fatal(e)
	}
	p, e := measurements.PlanPodDashboard(measurements.PodDashboardRequest{Discovery: *d})
	if e != nil || len(p.Candidates) != 16 || p.CandidatesOmitted != 5 || len(p.Widgets) != 2 {
		t.Fatal("default-before-cap", p, e)
	}
	metricsFound := map[string]bool{}
	for _, w := range p.Widgets {
		metricsFound[w.Question.Definition.Source.Metric] = true
	}
	if !metricsFound["k8s.pod.memory.usage"] || !metricsFound["k8s.pod.memory.working_set"] {
		t.Fatal("usage variants displaced working set")
	}
	sem, e := pnDiscover(q, []json.RawMessage{pnMetadata(q, "k8s.pod.memory.usage", "reported-scope", "v", measurements.PodMetricsSchema, map[string]string{}), pnMetadata(q, "k8s.pod.memory.working_set", "reported-scope", "v", measurements.PodMetricsSchema, map[string]string{})})
	if e != nil {
		t.Fatal(e)
	}
	sp, e := measurements.PlanPodDashboard(measurements.PodDashboardRequest{Discovery: *sem})
	if e != nil || len(sp.Candidates) != 1 || sp.Candidates[0].Profile != measurements.PodSemconvProfile {
		t.Fatal("profile alias union", sp, e)
	}
	unknown := q
	unknown.ResourceSchemaURL = nil
	if unknown.Validate() == nil {
		t.Fatal("unknown trace schema acquired native authority")
	}
	def := p.Widgets[0].Question.Definition
	def.Source.ExactStream.Attributes = nil
	before, _ := json.Marshal(def)
	authored, e := yaml.Marshal(def)
	if e != nil {
		t.Fatal(e)
	}
	fresh, e := measurements.Read(string(authored))
	if e != nil || fresh.Source.ExactStream.Attributes == nil {
		t.Fatal("fresh YAML did not canonicalize nil attributes", e)
	}
	after, _ := json.Marshal(def)
	if string(before) != string(after) || !strings.Contains(string(after), `"attributes":null`) {
		t.Fatal("direct JSON custody rewritten")
	}
	for _, kind := range []string{"stock", "counter", "histogram"} {
		bad := fresh
		bad.Kind = kind
		if bad.Validate() == nil {
			t.Fatal("pod gauge authority expanded", kind)
		}
	}
	s := *fresh.Source.ExactStream
	points := []metrics.GaugePoint{{Stream: s, EndNS: "10", Value: .5}}
	if _, e := metrics.EvaluateGauges("0", "20", "20", points); e == nil {
		t.Fatal("service gauge authority relaxed")
	}
	if _, e := metrics.EvaluateHostGauges("0", "20", "20", "host", points); e == nil {
		t.Fatal("host authority relaxed")
	}
}
func TestPodNativeMissingStaleConflictNoFallback(t *testing.T) {
	s := metrics.Stream{Resource: map[string]string{"k8s.pod.uid": "one"}, Scope: "reported", Name: "gauge", Unit: "By", Attributes: map[string]string{}}
	for _, c := range []struct {
		state  string
		points []metrics.GaugePoint
	}{{"stale", []metrics.GaugePoint{{Stream: s, EndNS: "20", Value: 2048}}}, {"unknown", []metrics.GaugePoint{{Stream: s, EndNS: "80", Value: 2048}, {Stream: s, EndNS: "90", NoRecordedValue: true}}}, {"unknown", []metrics.GaugePoint{{Stream: s, EndNS: "90", Value: 2048}, {Stream: s, EndNS: "90", Value: 1024}}}} {
		r, e := metrics.EvaluatePodGauges("0", "100", "20", "one", c.points)
		if e != nil {
			t.Fatal(e)
		}
		latest := r.Series[0].Samples[len(r.Series[0].Samples)-1]
		if latest.State != c.state || (c.state == "unknown" && latest.Value != nil) {
			t.Fatal("invalid fallback", latest)
		}
	}
}

func TestPodNativePositivePreviewAndRequiredMetadata(t *testing.T) {
	q := pnQuestion("one")
	row := pnMetadata(q, "k8s.pod.memory.usage", measurements.KubeletStatsScope, "0.137.0", "", map[string]string{})
	d, e := pnDiscover(q, []json.RawMessage{row})
	if e != nil {
		t.Fatal(e)
	}
	for _, at := range []int{24, 25} {
		encoded, _ := json.Marshal(d)
		var copy measurements.PodMetricDiscovery
		_ = json.Unmarshal(encoded, &copy)
		coverage := &copy.Instruments[0].CapabilityPlan.Coverage
		good := coverage.Streams[0]
		bad := good
		bad.Stream.Resource = map[string]string{}
		for k, v := range good.Stream.Resource {
			bad.Stream.Resource[k] = v
		}
		bad.Stream.Resource["k8s.pod.uid"] = "other"
		coverage.Streams = make([]measurements.ObservedStream, 26)
		for n := range coverage.Streams {
			coverage.Streams[n] = bad
		}
		coverage.Streams[at] = good
		p, e := measurements.PlanPodDashboard(measurements.PodDashboardRequest{Discovery: copy})
		want := 0
		if at == 24 {
			want = 1
		}
		if e != nil || len(p.Candidates) != want || p.StreamsOmitted != 1 {
			t.Fatal("positive first25 boundary", at, p, e)
		}
	}
	for _, field := range []string{"resourceSchemaURL", "schemaURL", "scope", "version", "unit", "attributes", "resource"} {
		var bad map[string]any
		_ = json.Unmarshal(row, &bad)
		delete(bad, field)
		raw, _ := json.Marshal(bad)
		if _, e := pnDiscover(q, []json.RawMessage{raw}); e == nil {
			t.Fatal("required source metadata missing", field)
		}
	}
}
