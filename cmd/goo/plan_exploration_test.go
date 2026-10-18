package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPlanExplorationUsesSharedCompiler(t *testing.T) {
	var out bytes.Buffer
	if e := run([]string{"plan-exploration", "-query", "../../examples/exploration/errors.json"}, &out); e != nil {
		t.Fatal(e)
	}
	if !strings.Contains(out.String(), "mapContains") || !strings.Contains(out.String(), "groupBy") {
		t.Fatal("missing shared plan")
	}
	path := filepath.Join(t.TempDir(), "large.json")
	os.WriteFile(path, []byte(strings.Repeat(" ", 16385)), 0600)
	if e := run([]string{"plan-exploration", "-query", path}, &out); e == nil {
		t.Fatal("large input accepted")
	}
}
