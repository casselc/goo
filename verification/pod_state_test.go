package verification

import (
	"encoding/json"
	"fmt"
	"github.com/casselc/goo/measurements"
	"reflect"
	"testing"
	"time"

	"github.com/casselc/goo/metrics"
	"hegel.dev/go/hegel"
)

func TestHegelPodStatePermutationAndQuality(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		code := float64(hegel.Draw(tc, hegel.Integers(-10, 15)))
		at := "90"
		stale := hegel.Draw(tc, hegel.Integers(0, 1)) == 1
		if stale {
			at = "10"
		}
		stream := metrics.Stream{Scope: "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/k8sclusterreceiver", Version: "0.137.0", Name: "k8s.pod.phase", Resource: map[string]string{"k8s.pod.uid": "p"}}
		points := []metrics.GaugePoint{{Stream: stream, EndNS: at, Value: code}, {Stream: stream, EndNS: at, Value: code}}
		r, err := metrics.EvaluatePodPhases("0", "100", "30", "p", points)
		if err != nil {
			tc.Errorf("valid profile: %v", err)
			return
		}
		s := r.Series[0].Samples[0]
		quality := "observed"
		if stale {
			quality = "stale"
		}
		if s.State != quality || s.Code == nil || *s.Code != code || (s.CodeState == "supported") != (code >= 1 && code <= 5) || (s.Label != "") != (code >= 1 && code <= 5) {
			tc.Errorf("quality/domain conflated: %+v", s)
			return
		}
		points[1].Value = code + 1
		conflict, err := metrics.EvaluatePodPhases("0", "100", "30", "p", points)
		if err != nil || conflict.Series[0].Samples[0].State != "unknown" || conflict.Series[0].Samples[0].Code != nil {
			tc.Errorf("conflict became label")
			return
		}
		points[0], points[1] = points[1], points[0]
		reversed, err := metrics.EvaluatePodPhases("0", "100", "30", "p", points)
		if err != nil || !reflect.DeepEqual(conflict, reversed) {
			tc.Errorf("permutation changed result")
		}
	}, options(60)...)
}

func TestHegelPodStateNativeReplayRejectsLabelMutation(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		code := hegel.Draw(tc, hegel.Integers(-10, 15))
		s := metrics.Stream{Resource: map[string]string{"k8s.pod.uid": "p", "deployment.environment.name": "prod"}, Attributes: map[string]string{}, Scope: "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/k8sclusterreceiver", Version: "0.137.0", Name: "k8s.pod.phase"}
		q := measurements.Question{FromMS: 100000, ToMS: 104000, Definition: measurements.Definition{APIVersion: measurements.Version, Name: "pod-phase", Title: "Pod phase", Kind: "state", StateProfile: metrics.PodPhaseProfile, Environment: "prod", LookbackMinutes: 1, MaxGapMS: 60000, Grain: "per-stream", Source: measurements.Binding{Services: []string{}, Resource: &measurements.ResourceBinding{Kind: "kubernetes-pod", ID: "p"}, ExactStream: &s, Scope: s.Scope, Version: s.Version, Metric: s.Name}}}
		row, _ := json.Marshal(map[string]any{"resource": s.Resource, "resource_schema_url": "", "scope": s.Scope, "version": s.Version, "schema_url": "", "name": s.Name, "unit": "", "attributes": s.Attributes, "end_ns": "102000000000", "flags": 0, "value": fmt.Sprint(code)})
		r, err := measurements.Evaluate(q, []json.RawMessage{row}, time.Unix(200, 0))
		if err != nil || measurements.ValidateResult(r) != nil {
			tc.Errorf("native replay: %v", err)
			return
		}
		sample := r.States.Series[0].Samples[0]
		if sample.Code == nil || *sample.Code != float64(code) || (sample.CodeState == "supported") != (code >= 1 && code <= 5) {
			tc.Errorf("code changed")
			return
		}
		r.States.Series[0].Samples[0].Label = "Healthy"
		if measurements.ValidateResult(r) == nil {
			tc.Errorf("invented health label accepted")
		}
	}, options(40)...)
}

func TestHegelPodPhaseDiscoveryPreservesExactStreamAndReadBudget(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		q := pnQuestion(fmt.Sprintf("phase-%d", hegel.Draw(tc, hegel.Integers(1, 100000))))
		attr := fmt.Sprint(hegel.Draw(tc, hegel.Integers(0, 10000)))
		var row map[string]any
		_ = json.Unmarshal(pnMetadata(q, "k8s.pod.phase", "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/k8sclusterreceiver", "0.137.0", "", map[string]string{"reported.dimension": attr}), &row)
		row["unit"] = ""
		b, _ := json.Marshal(row)
		discovered, err := pnDiscover(q, []json.RawMessage{b})
		if err != nil {
			tc.Errorf("phase metadata: %v", err)
			return
		}
		plan, err := measurements.PlanPodDashboard(measurements.PodDashboardRequest{Discovery: *discovered})
		if err != nil || len(plan.Widgets) != 1 || len(plan.Candidates) != 1 {
			tc.Errorf("phase plan: %v", err)
			return
		}
		w := plan.Widgets[0]
		d := w.Question.Definition
		s := d.Source.ExactStream
		if d.Kind != "state" || d.StateProfile != metrics.PodPhaseProfile || s == nil || !reflect.DeepEqual(s.Resource, q.Resource) || s.Attributes["reported.dimension"] != attr || s.ResourceSchemaURL != *q.ResourceSchemaURL || len(d.Source.Services) != 0 || d.Source.Namespace != nil {
			tc.Errorf("phase authority lost")
			return
		}
		// Mixed metadata preserves memory defaults; state does not add an automatic read.
		mixed := []json.RawMessage{b, pnMetadata(q, "k8s.pod.memory.usage", measurements.KubeletStatsScope, "0.137.0", "", map[string]string{}), pnMetadata(q, "k8s.pod.memory.working_set", measurements.KubeletStatsScope, "0.137.0", "", map[string]string{})}
		discovered, err = pnDiscover(q, mixed)
		if err != nil {
			tc.Errorf("mixed metadata: %v", err)
			return
		}
		plan, err = measurements.PlanPodDashboard(measurements.PodDashboardRequest{Discovery: *discovered})
		if err != nil || len(plan.Widgets) != 2 || len(plan.Candidates) != 3 {
			tc.Errorf("mixed budget: %v", err)
			return
		}
		for _, w := range plan.Widgets {
			if w.Question.Definition.Kind != "gauge" {
				tc.Errorf("state grew default read budget")
			}
		}
	}, options(40)...)
}
