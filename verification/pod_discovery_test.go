package verification

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/casselc/goo/exploration"
	"hegel.dev/go/hegel"
	"io"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"testing"
)

type pdFact struct {
	Resource map[string]string
	Schema   *string
	Family   string
	At       string
}

func pdRows(facts []pdFact) []json.RawMessage {
	groups := map[string][]pdFact{}
	for _, f := range facts {
		k := siKey([]any{f.Resource, f.Schema, f.Family})
		groups[k] = append(groups[k], f)
	}
	rows := []json.RawMessage{}
	for _, g := range groups {
		first, last := g[0].At, g[0].At
		for _, f := range g {
			if f.At < first {
				first = f.At
			}
			if f.At > last {
				last = f.At
			}
		}
		b, _ := json.Marshal(map[string]any{"family": g[0].Family, "resource": g[0].Resource, "resource_schema_url": g[0].Schema, "count": strconv.Itoa(len(g)), "first_ns": first, "last_ns": last})
		rows = append(rows, b)
	}
	return rows
}
func pdFetch(rows []json.RawMessage) (*exploration.PodDiscovery, int, error) {
	reads := 0
	source := exploration.Source{Endpoint: "http://source.invalid", Client: &http.Client{Transport: cvTransport(func(req *http.Request) (*http.Response, error) {
		reads++
		return &http.Response{StatusCode: 200, Header: http.Header{}, Body: io.NopCloser(strings.NewReader(string(joinPDRows(rows))))}, nil
	})}}
	r, e := source.DiscoverPods(context.Background(), exploration.PodDiscoveryQuestion{Environment: "prod", FromMS: 1000, ToMS: 2000})
	return r, reads, e
}
func joinPDRows(rows []json.RawMessage) []byte {
	out := []byte{}
	for _, r := range rows {
		out = append(out, r...)
		out = append(out, '\n')
	}
	return out
}
func pdOracle(facts []pdFact, r *exploration.PodDiscovery) bool {
	expected := map[string]map[string][]pdFact{}
	for _, f := range facts {
		k := siKey([]any{f.Resource, f.Schema})
		if expected[k] == nil {
			expected[k] = map[string][]pdFact{}
		}
		expected[k][f.Family] = append(expected[k][f.Family], f)
	}
	if len(r.Observations) != len(expected) {
		return false
	}
	for _, o := range r.Observations {
		k := siKey([]any{o.Resource, o.ResourceSchemaURL})
		families, ok := expected[k]
		if !ok || len(families) != len(o.Families) || o.Actionable != (o.Resource["k8s.pod.uid"] != "") {
			return false
		}
		label := func(key string) string {
			v, ok := o.Resource[key]
			if !ok {
				return "missing"
			}
			if v == "" {
				return "empty"
			}
			return "present"
		}
		if o.PodUIDState != label("k8s.pod.uid") || o.ClusterUIDState != label("k8s.cluster.uid") {
			return false
		}
		for _, f := range o.Families {
			g := families[f.Family]
			if f.Count != strconv.Itoa(len(g)) {
				return false
			}
			first, last := g[0].At, g[0].At
			for _, v := range g {
				if v.At < first {
					first = v.At
				}
				if v.At > last {
					last = v.At
				}
			}
			if f.FirstNS != first || f.LastNS != last {
				return false
			}
		}
		delete(expected, k)
	}
	return len(expected) == 0
}
func TestHegelPodResourceVariantsAndFamilyCounts(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		n := hegel.Draw(tc, hegel.Integers(1, 12))
		facts := []pdFact{}
		schema := ""
		for v := 0; v < 3; v++ {
			res := map[string]string{"deployment.environment.name": "prod", "k8s.pod.name": "same-name", "reported": fmt.Sprint(n)}
			if v > 0 {
				res["k8s.pod.uid"] = ""
			}
			if v == 2 {
				res["k8s.pod.uid"] = "uid"
				res["k8s.cluster.uid"] = ""
				res["service.name"] = ""
			}
			for f, family := range []string{"logs", "spans", "gauge", "sum", "histogram"} {
				var s *string
				if family != "spans" {
					s = &schema
				}
				for j := 0; j < n+f; j++ {
					facts = append(facts, pdFact{Resource: res, Schema: s, Family: family, At: strconv.FormatInt(1000000000+int64(j), 10)})
				}
			}
		}
		before, _ := json.Marshal(facts)
		rows := pdRows(facts)
		r, reads, e := pdFetch(rows)
		if e != nil || reads != 1 || r.SourceGroups != len(rows) || !pdOracle(facts, r) {
			tc.Errorf("source/reference %v", e)
			return
		}
		for a, b := 0, len(rows)-1; a < b; a, b = a+1, b-1 {
			rows[a], rows[b] = rows[b], rows[a]
		}
		again, _, e := pdFetch(rows)
		if e != nil || r.Basis != again.Basis || !reflect.DeepEqual(r.Observations, again.Observations) {
			tc.Errorf("arrival/readclock changed identity basis %v", e)
			return
		}
		after, _ := json.Marshal(facts)
		if string(before) != string(after) {
			tc.Errorf("raw facts changed")
		}
	}, options(10)...)
}
func TestPodDiscoveryBudgetsAndWholeRefusal(t *testing.T) {
	schema := ""
	facts := []pdFact{}
	for n := 0; n < 101; n++ {
		facts = append(facts, pdFact{Resource: map[string]string{"deployment.environment.name": "prod", "k8s.pod.uid": strconv.Itoa(n)}, Schema: &schema, Family: "logs", At: "1000000000"})
	}
	rows := pdRows(facts)
	r, reads, e := pdFetch(rows)
	if e != nil || reads != 1 || len(r.Observations) != 100 || r.ObservationsOmitted != 1 || r.SourceGroups != 101 {
		t.Fatal("output cap", r, e)
	}
	tooMany := make([]json.RawMessage, 1001)
	for n := range tooMany {
		tooMany[n] = rows[0]
	}
	if _, _, e := pdFetch(tooMany); e == nil {
		t.Fatal("source-family sentinel admitted")
	}
	for _, edit := range []func(map[string]any){func(m map[string]any) { m["resource_schema_url"] = nil }, func(m map[string]any) { m["count"] = "01" }, func(m map[string]any) { m["last_ns"] = "2000000000" }, func(m map[string]any) { m["family"] = "future" }, func(m map[string]any) {
		r := m["resource"].(map[string]any)
		for n := 0; n < 65; n++ {
			r[fmt.Sprint(n)] = "x"
		}
	}, func(m map[string]any) { m["resource"].(map[string]any)["deployment.environment.name"] = "other" }} {
		var m map[string]any
		_ = json.Unmarshal(rows[0], &m)
		edit(m)
		bad, _ := json.Marshal(m)
		if _, _, e := pdFetch([]json.RawMessage{bad}); e == nil {
			t.Fatal("invalid source admitted")
		}
	}
	if _, _, e := pdFetch([]json.RawMessage{rows[0], rows[0]}); e == nil {
		t.Fatal("duplicate source-family accepted")
	}
}

func TestPodDiscoveryOmittedEvidenceBasisAndIdentityBytes(t *testing.T) {
	schema := ""
	facts := []pdFact{}
	for n := 0; n < 101; n++ {
		facts = append(facts, pdFact{Resource: map[string]string{"deployment.environment.name": "prod", "k8s.pod.uid": strconv.Itoa(n)}, Schema: &schema, Family: "logs", At: "1000000000"})
	}
	rows := pdRows(facts)
	r, _, e := pdFetch(rows)
	if e != nil {
		t.Fatal(e)
	}
	visible := map[string]bool{}
	for _, o := range r.Observations {
		visible[o.Resource["k8s.pod.uid"]] = true
	}
	changed := false
	for n, row := range rows {
		var v map[string]any
		_ = json.Unmarshal(row, &v)
		if !visible[v["resource"].(map[string]any)["k8s.pod.uid"].(string)] {
			v["count"] = "2"
			rows[n], _ = json.Marshal(v)
			changed = true
		}
	}
	again, _, e := pdFetch(rows)
	if e != nil || !changed || !reflect.DeepEqual(r.Observations, again.Observations) || r.Basis == again.Basis {
		t.Fatal("omitted evidence excluded from basis", e)
	}
	resource := map[string]string{"deployment.environment.name": "prod", "k8s.pod.uid": "bounded"}
	for n := 0; n < 9; n++ {
		resource[fmt.Sprint(n)] = strings.Repeat("x", 1900)
	}
	oversized := pdRows([]pdFact{{Resource: resource, Schema: &schema, Family: "logs", At: "1000000000"}})
	if _, _, e := pdFetch(oversized); e == nil {
		t.Fatal("identity >16KiB accepted below per-label/count limits")
	}
}
