package metrics

import (
	"math"
	"reflect"
	"testing"
)

func phasePoint(at string, value float64) GaugePoint {
	return GaugePoint{EndNS: at, Value: value, Stream: Stream{
		Scope:      "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/k8sclusterreceiver",
		Attributes: map[string]string{}, Version: "0.137.0", Name: "k8s.pod.phase", Resource: map[string]string{"k8s.pod.uid": "pod-a"},
	}}
}

func TestPodPhaseQualityAndDomain(t *testing.T) {
	for code, label := range map[float64]string{1: "Pending", 2: "Running", 3: "Succeeded", 4: "Failed", 5: "Unknown"} {
		r, err := EvaluatePodPhases("0", "100", "30", "pod-a", []GaugePoint{phasePoint("90", code)})
		if err != nil {
			t.Fatal(err)
		}
		s := r.Series[0].Samples[0]
		if s.State != "observed" || s.Label != label || s.Code == nil || *s.Code != code {
			t.Fatalf("bad decoded phase: %+v", s)
		}
	}
	for _, code := range []float64{-1, 0, 1.5, 6, 9007199254740992} {
		r, err := EvaluatePodPhases("0", "100", "30", "pod-a", []GaugePoint{phasePoint("90", code)})
		if err != nil {
			t.Fatal(err)
		}
		s := r.Series[0].Samples[0]
		if s.State != "observed" || s.CodeState != "unsupported" || s.Label != "" || s.Code == nil || *s.Code != code {
			t.Fatalf("unsupported code hidden: %+v", s)
		}
	}
	missing := phasePoint("95", 2)
	missing.NoRecordedValue = true
	points := []GaugePoint{phasePoint("10", 1), phasePoint("90", 2), phasePoint("90", 3), missing}
	r, err := EvaluatePodPhases("0", "100", "30", "pod-a", points)
	if err != nil {
		t.Fatal(err)
	}
	s := r.Series[0]
	if s.State != "unknown" || s.Samples[0].State != "stale" || s.Samples[0].Label != "Pending" || s.Samples[1].Code != nil || s.Samples[2].Code != nil {
		t.Fatalf("quality lost: %+v", s)
	}
	for i, j := 0, len(points)-1; i < j; i, j = i+1, j-1 {
		points[i], points[j] = points[j], points[i]
	}
	other, err := EvaluatePodPhases("0", "100", "30", "pod-a", points)
	if err != nil || !reflect.DeepEqual(r, other) {
		t.Fatal("input ordering changed result")
	}
	empty, err := EvaluatePodPhases("0", "100", "30", "pod-a", nil)
	if err != nil || len(empty.Series) != 0 {
		t.Fatal("absence manufactured state")
	}
}

func TestPodPhaseProfileAndAuthority(t *testing.T) {
	mutations := []func(*GaugePoint){
		func(p *GaugePoint) { p.Stream.Version = "0.138.0" },
		func(p *GaugePoint) { p.Stream.Scope = "other" },
		func(p *GaugePoint) { p.Stream.SchemaURL = "other" },
		func(p *GaugePoint) { p.Stream.Name = "k8s.container.ready" },
		func(p *GaugePoint) { p.Stream.Unit = "1" },
		func(p *GaugePoint) { p.Stream.Resource["k8s.pod.uid"] = "other" },
		func(p *GaugePoint) { p.Value = math.NaN() },
		func(p *GaugePoint) { p.Value = math.Inf(1) },
	}
	for i, mutate := range mutations {
		p := phasePoint("90", 2)
		mutate(&p)
		if _, err := EvaluatePodPhases("0", "100", "30", "pod-a", []GaugePoint{p}); err == nil {
			t.Fatalf("mutation %d admitted", i)
		}
	}
	p := phasePoint("90", 2)
	if _, err := EvaluatePodGauges("0", "100", "30", "pod-a", []GaugePoint{p}); err == nil {
		t.Fatal("unitless scalar contract relaxed")
	}
	second := phasePoint("90", 3)
	second.Stream.Resource["k8s.namespace.name"] = "other"
	before := cloneStream(p.Stream)
	r, err := EvaluatePodPhases("0", "100", "30", "pod-a", []GaugePoint{p, second})
	if err != nil || len(r.Series) != 2 || !reflect.DeepEqual(p.Stream, before) {
		t.Fatal("resource variants merged or input mutated")
	}
	r.Series[0].Stream.Resource["new"] = "value"
	if _, ok := p.Stream.Resource["new"]; ok {
		t.Fatal("result aliases input")
	}
}

func TestPodPhaseRequestBounds(t *testing.T) {
	if _, err := EvaluatePodPhases("0", "100", "30", "", nil); err == nil {
		t.Fatal("empty UID admitted")
	}
	if _, err := EvaluatePodPhases("0", "100", "30", "p", make([]GaugePoint, 10001)); err == nil {
		t.Fatal("unbounded points admitted")
	}
	for _, window := range [][3]string{{"100", "0", "30"}, {"0", "100", "0"}, {"bad", "100", "30"}} {
		if _, err := EvaluatePodPhases(window[0], window[1], window[2], "p", nil); err == nil {
			t.Fatal("invalid time contract admitted")
		}
	}
}
