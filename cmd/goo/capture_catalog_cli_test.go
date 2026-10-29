package main

import (
	"bytes"
	"encoding/json"
	"github.com/casselc/goo/exploration"
	"os"
	"path/filepath"
	"testing"
)

func TestCaptureCatalogRebuildCommand(t *testing.T) {
	dir := t.TempDir()
	b, e := os.ReadFile("../../testdata/storage/exploration-capture.json")
	if e != nil {
		t.Fatal(e)
	}
	var c exploration.SavedCapture
	if e = json.Unmarshal(b, &c); e != nil {
		t.Fatal(e)
	}
	file := filepath.Join(dir, c.ID+".json")
	if e = os.WriteFile(file, b, 0600); e != nil {
		t.Fatal(e)
	}
	var out bytes.Buffer
	if e = run([]string{"rebuild-capture-catalog", "-dir", dir}, &out); e != nil {
		t.Fatal(e)
	}
	var result exploration.CaptureCatalogRebuild
	if e = json.Unmarshal(out.Bytes(), &result); e != nil || result.Indexed != 1 {
		t.Fatal(result, e)
	}
	after, e := os.ReadFile(file)
	if e != nil || !bytes.Equal(b, after) {
		t.Fatal("payload altered", e)
	}
	if _, e = (exploration.CaptureStore{Dir: dir}).Search(t.Context(), exploration.CaptureSearchQuery{}, ""); e != nil {
		t.Fatal(e)
	}
	if e = run([]string{"rebuild-capture-catalog"}, &out); e == nil {
		t.Fatal("missing path accepted")
	}
}
