package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCLIContracts(t *testing.T) {
	definition := "../../examples/portable/reservations.yaml"
	for _, target := range []string{"clickhouse", "databricks"} {
		var out bytes.Buffer
		if err := run([]string{"compile", "-definition", definition, "-target", target, "-group-by", "release"}, &out); err != nil {
			t.Fatal(err)
		}
		var v map[string]any
		if err := json.Unmarshal(out.Bytes(), &v); err != nil || v["status"] != "requires-prepared-source" {
			t.Fatalf("bad CLI plan: %s", out.String())
		}
	}
	for _, args := range [][]string{nil, {"apply"}, {"compile", "-definition", definition, "-target", "unknown"}} {
		var out bytes.Buffer
		if err := run(args, &out); err == nil || strings.TrimSpace(out.String()) != "" {
			t.Fatal("invalid invocation should fail without publishing partial JSON")
		}
	}
}

func TestLifecycleCLI(t *testing.T) {
	var prepared bytes.Buffer
	if err := run([]string{"prepare", "-policy", "../../examples/lifecycle/policy.yaml",
		"-events", "../../examples/lifecycle/events.json"}, &prepared); err != nil {
		t.Fatal(err)
	}
	p := filepath.Join(t.TempDir(), "prepared.json")
	if err := os.WriteFile(p, prepared.Bytes(), 0600); err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	args := []string{"evaluate-prepared", "-definition", "../../examples/lifecycle/measurement.yaml", "-prepared", p}
	if err := run(args, &out); err != nil {
		t.Fatal(err)
	}
	var result struct {
		Receipt struct {
			State string `json:"state"`
		} `json:"receipt"`
	}
	if err := json.Unmarshal(out.Bytes(), &result); err != nil || result.Receipt.State != "unknown" {
		t.Fatal("unverified example must preserve unknown evidence")
	}
	out.Reset()
	if err := run(append(args, "-evaluation", "another-basis"), &out); err == nil || out.Len() != 0 {
		t.Fatal("receipt basis must not be overridden")
	}
}
