package exploration

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

func podDiscoveryTestQuestion() PodDiscoveryQuestion {
	return PodDiscoveryQuestion{Environment: "prod", FromMS: 1000, ToMS: 2000}
}
func podDiscoveryRow(family string, resource map[string]string, schema *string, count string) json.RawMessage {
	b, _ := json.Marshal(map[string]any{"family": family, "resource": resource, "resource_schema_url": schema, "count": count, "first_ns": "1000000000", "last_ns": "1999999999"})
	return b
}
func TestPodDiscoveryVariantsSchemaAndTrueFamilyCounts(t *testing.T) {
	q := podDiscoveryTestQuestion()
	schema := ""
	resource := map[string]string{"deployment.environment.name": "prod", "k8s.pod.uid": "pod-1", "k8s.pod.name": "same-name", "k8s.namespace.name": "payments"}
	rows := []json.RawMessage{podDiscoveryRow("logs", resource, &schema, "250"), podDiscoveryRow("sum", resource, &schema, "300"), podDiscoveryRow("spans", resource, nil, "999")}
	result, e := decodePodDiscovery(q, rows, time.Unix(100, 0))
	if e != nil || len(result.Observations) != 2 || result.SourceGroups != 3 {
		t.Fatal(result, e)
	}
	for _, o := range result.Observations {
		if !o.Actionable || o.PodUIDState != "present" || o.ClusterUIDState != "missing" {
			t.Fatal(o)
		}
		if _, ok := o.Resource["service.name"]; ok {
			t.Fatal("fabricated service")
		}
	}
	reversed := []json.RawMessage{rows[2], rows[1], rows[0]}
	again, e := decodePodDiscovery(q, reversed, time.Unix(200, 0))
	if e != nil || result.Basis != again.Basis {
		t.Fatal("readclock or arrival order changes basis", e)
	}
	missing := map[string]string{"deployment.environment.name": "prod", "k8s.pod.name": "same-name"}
	empty := map[string]string{"deployment.environment.name": "prod", "k8s.pod.name": "same-name", "k8s.pod.uid": "", "service.name": ""}
	result, e = decodePodDiscovery(q, []json.RawMessage{podDiscoveryRow("logs", missing, &schema, "1"), podDiscoveryRow("logs", empty, &schema, "1")}, time.Unix(100, 0))
	if e != nil || len(result.Observations) != 2 {
		t.Fatal(result, e)
	}
	for _, o := range result.Observations {
		if o.Actionable {
			t.Fatal("missing/empty UID actionable")
		}
	}
}
func TestPodDiscoverySourceBudgetAndOutputOmissions(t *testing.T) {
	q := podDiscoveryTestQuestion()
	schema := ""
	rows := []json.RawMessage{}
	for n := 0; n < 101; n++ {
		rows = append(rows, podDiscoveryRow("logs", map[string]string{"deployment.environment.name": "prod", "k8s.pod.uid": fmt.Sprint(n)}, &schema, "5000"))
	}
	result, e := decodePodDiscovery(q, rows, time.Unix(100, 0))
	if e != nil || len(result.Observations) != 100 || result.ObservationsOmitted != 1 || result.SourceGroups != 101 {
		t.Fatal(result, e)
	}
	rows = make([]json.RawMessage, 1001)
	if _, e := decodePodDiscovery(q, rows, time.Unix(100, 0)); e == nil {
		t.Fatal("source-family sentinel accepted")
	}
}
func TestPodDiscoveryRefusesMalformedSource(t *testing.T) {
	q := podDiscoveryTestQuestion()
	schema := ""
	good := string(podDiscoveryRow("logs", map[string]string{"deployment.environment.name": "prod", "k8s.pod.uid": "one"}, &schema, "1"))
	for _, bad := range []string{strings.Replace(good, `"count":"1"`, `"count":"01"`, 1), strings.Replace(good, `"k8s.pod.uid":"one"`, `"k8s.pod.uid":null`, 1), strings.Replace(good, `"resource_schema_url":""`, `"resource_schema_url":null`, 1), strings.Replace(good, `"last_ns":"1999999999"`, `"last_ns":"2000000000"`, 1), strings.Replace(good, `"family":"logs"`, `"family":"future"`, 1), strings.Replace(good, `"deployment.environment.name":"prod"`, `"deployment.environment.name":"other"`, 1)} {
		if _, e := decodePodDiscovery(q, []json.RawMessage{json.RawMessage(bad)}, time.Unix(100, 0)); e == nil {
			t.Fatal("bad source admitted", bad)
		}
	}
	if _, e := decodePodDiscovery(q, []json.RawMessage{json.RawMessage(good), json.RawMessage(good)}, time.Unix(100, 0)); e == nil {
		t.Fatal("duplicate source-family admitted")
	}
}
func TestPodDiscoveryOneGroupedReadAndCancellation(t *testing.T) {
	q := podDiscoveryTestQuestion()
	reads := 0
	source := Source{Endpoint: "http://source.invalid", Client: &http.Client{Transport: recordWireTransport(func(r *http.Request) (*http.Response, error) {
		reads++
		b, _ := io.ReadAll(r.Body)
		sql := string(b)
		if strings.Count(sql, " UNION ALL ") != 4 || strings.Count(sql, " GROUP BY labels, resource_schema_url") != 5 || !strings.Contains(sql, "CAST(NULL AS Nullable(String))") || !strings.Contains(sql, "LIMIT 1001") || strings.Contains(sql, "mapUpdate") {
			t.Fatal(sql)
		}
		return &http.Response{StatusCode: 200, Header: http.Header{}, Body: io.NopCloser(strings.NewReader(""))}, nil
	})}}
	result, e := source.DiscoverPods(context.Background(), q)
	if e != nil || reads != 1 || len(result.Observations) != 0 {
		t.Fatal(result, e, reads)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, e := source.DiscoverPods(ctx, q); e == nil || reads != 1 {
		t.Fatal("cancelled source read", e, reads)
	}
	q.Environment = ""
	if _, e := source.DiscoverPods(context.Background(), q); e == nil || reads != 1 {
		t.Fatal("bad request reached source", e, reads)
	}
}
