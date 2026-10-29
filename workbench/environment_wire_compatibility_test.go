package workbench_test

import (
	"bytes"
	"encoding/json"
	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/workbench"
	"gopkg.in/yaml.v3"
	"os"
	"testing"
)

func TestLegacyEnvironmentWires249(t *testing.T) {
	data, err := os.ReadFile("../testdata/protocol/environment-legacy-wires.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture map[string]json.RawMessage
	if err = json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	q := exploration.Query{Signal: "logs", Environment: "prod'雪", FromMS: 60000, ToMS: 120000, Limit: 100, Where: &exploration.Predicate{Field: &exploration.Field{Kind: "resource", Key: "service.namespace"}, Op: "eq", Value: ""}}
	tail := exploration.TailSelector{Signal: "logs", Environment: q.Environment, Where: q.Where}
	p, err := exploration.DraftTelemetryPackage(q, "sdk-logs", "SDK logs", 1)
	if err != nil {
		t.Fatal(err)
	}
	revision, err := p.Revision()
	if err != nil {
		t.Fatal(err)
	}
	v := workbench.ActivityView{Version: 1, Title: "SDK logs", Selector: tail, Columns: []string{"utc", "service", "summary"}}
	store := workbench.Store{Dir: t.TempDir()}
	saved, err := store.CreateActivityView(v)
	if err != nil {
		t.Fatal(err)
	}
	yamlBytes, err := yaml.Marshal(p)
	if err != nil {
		t.Fatal(err)
	}
	actual := map[string]any{"query": q, "tail": tail, "package": p, "packageRevision": revision, "packageYAML": string(yamlBytes), "captureQuestion": (exploration.CaptureQuestion{Kind: "explore", Explore: &q}).Normalized(), "activityView": saved}
	for name, value := range actual {
		t.Run(name, func(t *testing.T) {
			raw, err := json.Marshal(value)
			if err != nil {
				t.Fatal(err)
			}
			var compact bytes.Buffer
			if err = json.Compact(&compact, fixture[name]); err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(raw, compact.Bytes()) {
				t.Fatalf("old wire identity changed\n%s\n%s", raw, compact.Bytes())
			}
		})
	}
	reopened, err := store.LoadActivityView(saved.ID)
	if err != nil || reopened.Selector.Environment != q.Environment {
		t.Fatal("old view exact reopen failed", err)
	}
}
