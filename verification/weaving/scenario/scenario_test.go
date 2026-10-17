package scenario

import (
	"context"
	"github.com/casselc/goo/prebuilt"
	"github.com/casselc/observability/quintgo/qobs"
	"github.com/casselc/observability/quintgo/qtest"
	"github.com/casselc/observability/quintgo/validate"
	"os"
	"path/filepath"
	"testing"
)

func TestWovenRealHTTPPreparation(t *testing.T) {
	directory := os.Getenv("QUINTGO_KEEP")
	if directory == "" {
		directory = t.TempDir()
	}
	inf, err := os.Open("../../../examples/packages/inventory.json")
	if err != nil {
		t.Fatal(err)
	}
	defer inf.Close()
	in, err := prebuilt.ReadInventory(inf)
	if err != nil {
		t.Fatal(err)
	}
	bf, err := os.Open("../../../examples/packages/http.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer bf.Close()
	binding, err := prebuilt.ReadBinding(bf)
	if err != nil {
		t.Fatal(err)
	}
	expanded, err := prebuilt.Expand(in, binding)
	if err != nil {
		t.Fatal(err)
	}
	status := 200
	requests := []prebuilt.HTTPRequest{{Kind: "server", Resource: in.Scopes[0].Resource, TraceID: "t1", SpanID: "s1", Method: "POST", Route: "/checkout", Status: &status}, {Kind: "server", Resource: in.Scopes[0].Resource, TraceID: "t2", SpanID: "s2", Method: "POST"}}
	observed := qtest.Record(t, "../binding.yaml", func(*qobs.Recorder) {
		p, e := expanded.PrepareHTTP(requests, "woven-http")
		if e != nil {
			t.Fatal(e)
		}
		if p.Facts[1]["classification"] != "unknown" {
			t.Fatal("unknown request lost")
		}
	}, qtest.WithOptions(validate.Options{Backend: "typescript"}), qtest.KeepIn(filepath.Join(directory, "faithful")))
	if len(observed.ITF) == 0 {
		t.Fatal("no retained model-level trace")
	}
	if len(observed.Steps) != 1 || !observed.OK() {
		t.Fatal("woven trace missing or invalid", observed.Report.Summary())
	}
	// Negative control: a one-row loss must violate the same model.
	bad := qtest.Record(t, "../binding.yaml", func(rec *qobs.Recorder) {
		rec.Record(context.Background(), "prepare_http", nil, "actor", "bad", "obs.selected", 2, "obs.facts", 1)
	}, qtest.WithOptions(validate.Options{Backend: "typescript"}), qtest.KeepIn(filepath.Join(directory, "negative-control")))
	if bad.OK() {
		t.Fatal("model accepted dropped request")
	}
}
