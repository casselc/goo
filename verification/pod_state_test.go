package verification

import (
	"reflect"
	"testing"

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
