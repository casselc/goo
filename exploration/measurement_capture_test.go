package exploration

import (
	"context"
	"encoding/json"
	"github.com/casselc/goo/measurements"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestMeasurementCaptureSourceFreeReopenAndFixedRefresh(t *testing.T) {
	b, e := os.ReadFile("../config/measurements/worker-active.yaml")
	if e != nil {
		t.Fatal(e)
	}
	d, e := measurements.Read(string(b))
	if e != nil {
		t.Fatal(e)
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{"resource": map[string]string{"service.name": d.Source.Services[0], "service.instance.id": "worker", "deployment.environment.name": d.Environment}, "scope": d.Source.Scope, "version": d.Source.Version, "schema_url": d.Source.SchemaURL, "attributes": d.Source.Attributes, "name": d.Source.Metric, "unit": d.Source.Unit, "end_ns": "210000000000", "value": "1"})
	}))
	defer server.Close()
	s := CaptureStore{Dir: t.TempDir()}
	q := CaptureQuestion{Kind: "measurement", Measurement: &measurements.Question{Definition: d, FromMS: 100000, ToMS: 220000}}
	source := Source{Endpoint: server.URL}
	c, e := s.Create(context.Background(), source, "Worker evidence", q, "")
	if e != nil {
		t.Fatal(e)
	}
	p, e := c.Page("")
	if e != nil || p.RetainedRecords != 1 || p.Measurement == nil {
		t.Fatal("measurement paging envelope", e)
	}
	// Empty selectors are valid and normalize identically to an omitted map.
	empty := q
	emptyQuestion := *q.Measurement
	emptyDefinition := emptyQuestion.Definition
	emptyDefinition.Source.Attributes = map[string]string{}
	emptyQuestion.Definition = emptyDefinition
	empty.Measurement = &emptyQuestion
	if _, e = s.Create(context.Background(), source, "Empty selectors", empty, ""); e != nil {
		t.Fatal("valid empty selectors failed capture", e)
	}
	child, e := s.Create(context.Background(), source, c.Title, c.Question, c.ID)
	if e != nil || child.Parent != c.ID {
		t.Fatal("refresh", e)
	}
	changed := q
	copy := *q.Measurement
	copy.ToMS++
	changed.Measurement = &copy
	if _, e = s.Create(context.Background(), source, c.Title, changed, c.ID); e == nil {
		t.Fatal("refresh scope changed")
	}
	altered := *c
	evidence := *c.Measurement
	evidence.Plan.SQL += " /* previous compiler */"
	altered.Measurement = &evidence
	old, e := s.publish(altered)
	if e != nil {
		t.Fatal(e)
	}
	if _, e = s.Load(old.ID); e != nil {
		t.Fatal("old plan must reopen", e)
	}
	if _, e = s.Create(context.Background(), source, c.Title, c.Question, old.ID); e != ErrCapturePlanChanged {
		t.Fatal("old plan refreshed silently", e)
	}
	server.Close()
	loaded, e := s.Load(c.ID)
	if e != nil || loaded.Measurement.Count != "1" {
		t.Fatal("source-free reopen", e)
	}
	if _, e = s.Create(context.Background(), source, c.Title, c.Question, c.ID); e == nil {
		t.Fatal("source outage accepted")
	}
	if _, e = s.Load(c.ID); e != nil {
		t.Fatal("failed refresh altered original", e)
	}
}
