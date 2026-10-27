package verification

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"

	"github.com/casselc/goo/measurements"
	"hegel.dev/go/hegel"
)

// This exercises the physical Source decoder and planner, not a caller-created
// supported Instrument. Missing and empty context are independent refusals.
func TestHegelContainerReadinessDiscoveryAuthority(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		n := hegel.Draw(tc, hegel.Integers(1, 100000))
		q := pnQuestion(fmt.Sprintf("pod-%d", n))
		q.Resource["container.id"] = fmt.Sprintf("containerd://%d", n)
		q.Resource["k8s.container.name"] = fmt.Sprintf("worker-%d", n)
		q.Resource["reported.extra"] = fmt.Sprint(n)
		raw := func(metric string) json.RawMessage {
			b, _ := json.Marshal(map[string]any{"kind": "gauge", "scope": "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/k8sclusterreceiver", "version": "0.137.0", "schemaURL": "", "metric": metric, "unit": "", "temporality": 0, "monotonic": false, "resource": q.Resource, "resourceSchemaURL": *q.ResourceSchemaURL, "attributes": map[string]string{}})
			return b
		}
		rows := []json.RawMessage{raw("k8s.container.ready"), raw("k8s.pod.phase")}
		d, e := pnDiscover(q, rows)
		if e != nil {
			tc.Errorf("discovery: %v", e)
			return
		}
		before, _ := json.Marshal(d)
		p, e := measurements.PlanPodDashboard(measurements.PodDashboardRequest{Discovery: *d})
		after, _ := json.Marshal(d)
		if e != nil || len(p.Widgets) != 2 || string(before) != string(after) {
			tc.Errorf("two bounded defaults/immutability: %v", e)
			return
		}
		found := false
		for _, w := range p.Widgets {
			if w.Question.Definition.Source.Metric == "k8s.container.ready" {
				found = true
				s := w.Question.Definition.Source
				if s.Resource == nil || s.Resource.Kind != "kubernetes-container" || s.Resource.ID != q.Resource["container.id"] || s.ExactStream == nil || !reflect.DeepEqual(s.ExactStream.Resource, q.Resource) || s.ExactStream.ResourceSchemaURL != *q.ResourceSchemaURL || len(s.Services) != 0 || s.Namespace != nil {
					tc.Errorf("container authority broadened")
					return
				}
			}
		}
		if !found {
			tc.Errorf("readiness omitted")
			return
		}
		rows = append(rows, pnMetadata(q, "k8s.pod.memory.usage", measurements.KubeletStatsScope, "0.137.0", "", map[string]string{}), pnMetadata(q, "k8s.pod.memory.working_set", measurements.KubeletStatsScope, "0.137.0", "", map[string]string{}))
		mixed, e := pnDiscover(q, rows)
		if e != nil {
			tc.Errorf("mixed: %v", e)
			return
		}
		m, e := measurements.PlanPodDashboard(measurements.PodDashboardRequest{Discovery: *mixed})
		if e != nil || len(m.Widgets) != 2 || len(m.Candidates) != 4 {
			tc.Errorf("mixed budget: %v", e)
			return
		}
		for _, w := range m.Widgets {
			if w.Question.Definition.Kind != "gauge" {
				tc.Errorf("categorical candidate displaced memory")
				return
			}
		}
		for a, b := 0, len(rows)-1; a < b; a, b = a+1, b-1 {
			rows[a], rows[b] = rows[b], rows[a]
		}
		rev, e := pnDiscover(q, rows)
		if e != nil {
			tc.Errorf("reverse: %v", e)
			return
		}
		mp, e := measurements.PlanPodDashboard(measurements.PodDashboardRequest{Discovery: *rev})
		if e != nil || !reflect.DeepEqual(m, mp) {
			tc.Errorf("arrival order changes plan")
			return
		}
		for _, key := range []string{"container.id", "k8s.container.name"} {
			for _, empty := range []bool{false, true} {
				saved := q.Resource[key]
				if empty {
					q.Resource[key] = ""
				} else {
					delete(q.Resource, key)
				}
				bad, e := pnDiscover(q, []json.RawMessage{raw("k8s.container.ready")})
				q.Resource[key] = saved
				if e != nil {
					tc.Errorf("negative discovery: %v", e)
					return
				}
				bp, e := measurements.PlanPodDashboard(measurements.PodDashboardRequest{Discovery: *bad})
				if e != nil || len(bp.Candidates) != 0 {
					tc.Errorf("missing/empty %s inferred", key)
					return
				}
			}
		}
	}, options(10)...)
}
