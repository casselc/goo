package verification

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/casselc/goo/measurements"
	"github.com/casselc/goo/metrics"
	"hegel.dev/go/hegel"
)

func TestHegelContainerReadinessAuthorityAndZeroQuality(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		id := fmt.Sprintf("reported-%d", hegel.Draw(tc, hegel.Integers(1, 100000)))
		code := hegel.Draw(tc, hegel.Integers(-2, 3))
		stale := hegel.Draw(tc, hegel.Integers(0, 1)) == 1
		at := "103500000000"
		if stale {
			at = "102000000000"
		}
		stream := metrics.Stream{Resource: map[string]string{"container.id": id, "k8s.pod.uid": "p", "k8s.container.name": "worker", "deployment.environment.name": "prod"}, Attributes: map[string]string{}, Scope: "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/k8sclusterreceiver", Version: "0.137.0", Name: "k8s.container.ready"}
		q := measurements.Question{FromMS: 100000, ToMS: 104000, Definition: measurements.Definition{APIVersion: measurements.Version, Name: "ready", Title: "Reported readiness", Kind: "state", StateProfile: metrics.ContainerReadyProfile, Environment: "prod", LookbackMinutes: 1, MaxGapMS: 1000, Grain: "per-stream", Source: measurements.Binding{Services: []string{}, Resource: &measurements.ResourceBinding{Kind: "kubernetes-container", ID: id}, ExactStream: &stream, Scope: stream.Scope, Version: stream.Version, Metric: stream.Name}}}
		row := map[string]any{"resource": stream.Resource, "resource_schema_url": "", "scope": stream.Scope, "version": stream.Version, "schema_url": "", "name": stream.Name, "unit": "", "attributes": stream.Attributes, "end_ns": at, "flags": 0, "value": fmt.Sprint(code)}
		b, _ := json.Marshal(row)
		r, err := measurements.Evaluate(q, []json.RawMessage{b}, time.Unix(200, 0), "", nil)
		if err != nil || measurements.ValidateResult(r) != nil {
			tc.Errorf("readiness replay: %v", err)
			return
		}
		s := r.States.Series[0].Samples[0]
		quality := "observed"
		if stale {
			quality = "stale"
		}
		label := ""
		if code == 0 {
			label = "Not ready"
		}
		if code == 1 {
			label = "Ready"
		}
		if s.State != quality || s.Code == nil || *s.Code != float64(code) || s.Label != label || (s.CodeState == "supported") != (code == 0 || code == 1) {
			tc.Errorf("domain/quality conflated: %+v", s)
			return
		}
		// A restart is a different reported identity even with the same name/pod.
		changed := map[string]string{}
		for k, v := range stream.Resource {
			changed[k] = v
		}
		changed["container.id"] = id + "-restart"
		row["resource"] = changed
		b, _ = json.Marshal(row)
		if _, err := measurements.Evaluate(q, []json.RawMessage{b}, time.Unix(200, 0), "", nil); err == nil {
			tc.Errorf("replacement container accepted")
			return
		}
		row["resource"] = stream.Resource
		row["value"] = "1e-400"
		b, _ = json.Marshal(row)
		if _, err := measurements.Evaluate(q, []json.RawMessage{b}, time.Unix(200, 0), "", nil); err == nil {
			tc.Errorf("rounding manufactured Not ready")
			return
		}
		r.States.Series[0].Samples[0].Label = "Healthy pod"
		if measurements.ValidateResult(r) == nil {
			tc.Errorf("invented health accepted")
		}
	}, options(60)...)
}
