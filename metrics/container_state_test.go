package metrics

import (
	"reflect"
	"testing"
)

func readyPoint(at string, value float64) GaugePoint {
	p := phasePoint(at, value)
	p.Stream.Name = "k8s.container.ready"
	p.Stream.Resource["container.id"] = "reported-id"
	p.Stream.Resource["k8s.container.name"] = "worker"
	return p
}

func TestContainerReadinessDomainAndQuality(t *testing.T) {
	for _, c := range []struct {
		code          float64
		label, domain string
	}{{0, "Not ready", "supported"}, {1, "Ready", "supported"}, {-1, "", "unsupported"}, {0.5, "", "unsupported"}, {2, "", "unsupported"}} {
		p := readyPoint("90", c.code)
		r, err := EvaluateContainerReadiness("0", "100", "30", "reported-id", []GaugePoint{p})
		if err != nil {
			t.Fatal(err)
		}
		s := r.Series[0].Samples[0]
		if s.State != "observed" || s.Code == nil || *s.Code != c.code || s.Label != c.label || s.CodeState != c.domain {
			t.Fatalf("bad readiness: %+v", s)
		}
	}
	missing := readyPoint("95", 1)
	missing.NoRecordedValue = true
	points := []GaugePoint{readyPoint("10", 1), readyPoint("90", 0), readyPoint("90", 1), missing}
	r, err := EvaluateContainerReadiness("0", "100", "30", "reported-id", points)
	if err != nil {
		t.Fatal(err)
	}
	s := r.Series[0]
	if s.State != "unknown" || s.Samples[0].State != "stale" || s.Samples[0].Label != "Ready" || s.Samples[1].Code != nil || s.Samples[2].Code != nil {
		t.Fatal("quality lost", s)
	}
	for a, b := 0, len(points)-1; a < b; a, b = a+1, b-1 {
		points[a], points[b] = points[b], points[a]
	}
	again, err := EvaluateContainerReadiness("0", "100", "30", "reported-id", points)
	if err != nil || !reflect.DeepEqual(r, again) {
		t.Fatal("order changed readiness")
	}
	empty, err := EvaluateContainerReadiness("0", "100", "30", "reported-id", nil)
	if err != nil || len(empty.Series) != 0 {
		t.Fatal("absence became Not ready")
	}
}

func TestContainerReadinessAuthorityAndProfile(t *testing.T) {
	for _, edit := range []func(*GaugePoint){
		func(p *GaugePoint) { delete(p.Stream.Resource, "container.id") },
		func(p *GaugePoint) { delete(p.Stream.Resource, "k8s.pod.uid") },
		func(p *GaugePoint) { delete(p.Stream.Resource, "k8s.container.name") },
		func(p *GaugePoint) { p.Stream.Resource["container.id"] = "replacement-id" },
		func(p *GaugePoint) { p.Stream.Version = "v0.137.0" },
		func(p *GaugePoint) { p.Stream.Unit = "{container}" },
		func(p *GaugePoint) { p.Stream.Name = "k8s.pod.phase" },
	} {
		p := readyPoint("90", 1)
		edit(&p)
		if _, err := EvaluateContainerReadiness("0", "100", "30", "reported-id", []GaugePoint{p}); err == nil {
			t.Fatal("missing identity/profile admitted")
		}
	}
	a, b := readyPoint("90", 0), readyPoint("90", 1)
	b.Stream.ResourceSchemaURL = "other-schema"
	r, err := EvaluateContainerReadiness("0", "100", "30", "reported-id", []GaugePoint{a, b})
	if err != nil || len(r.Series) != 2 {
		t.Fatal("resource schemas merged")
	}
	if _, err := EvaluatePodGauges("0", "100", "30", "pod-a", []GaugePoint{a}); err == nil {
		t.Fatal("scalar unit contract relaxed")
	}
}
