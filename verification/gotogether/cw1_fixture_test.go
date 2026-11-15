package shadow

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestCW1Fixture replays a portable consumer snapshot, not its original SQL gate.
func TestCW1Fixture(t *testing.T) {
	const dir = "testdata/cw1"
	raw, err := os.ReadFile(filepath.Join(dir, "manifest.json"))
	if err != nil {
		t.Fatal(err)
	}
	var manifest struct {
		SHA256 map[string]string `json:"sha256"`
	}
	if err = json.Unmarshal(raw, &manifest); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"contract.md", "model/acquisition.qnt", "acquisition.binding.yaml", "scenario.go.txt", "requests.jsonl"} {
		data, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			t.Fatal(err)
		}
		sum := sha256.Sum256(data)
		if manifest.SHA256[name] != hex.EncodeToString(sum[:]) {
			t.Fatalf("fixture hash drift: %s", name)
		}
	}
	// The frozen handoff remains intact; only the live binding's spec path moved.
	for snapshot, source := range map[string]string{"acquisition.binding.yaml": binding, "scenario.go.txt": "shadow.go"} {
		a, err := os.ReadFile(filepath.Join(dir, snapshot))
		if err != nil {
			t.Fatal(err)
		}
		b, err := os.ReadFile(source)
		if err != nil {
			t.Fatal(err)
		}
		if snapshot == "acquisition.binding.yaml" {
			a = []byte(strings.Replace(string(a), "spec: model/acquisition.qnt", "spec: model/acquisition.md", 1))
		}
		if string(a) != string(b) {
			t.Fatalf("review CW1 snapshot drift from %s", source)
		}
	}
	model, err := os.ReadFile(filepath.Join(dir, "model/acquisition.qnt"))
	if err != nil {
		t.Fatal(err)
	}
	authored, err := os.ReadFile("model/acquisition.md")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(authored), "```quint\n"+string(model)+"```\n") {
		t.Fatal("authored contract changed: review the frozen CW1 replay basis")
	}
	document, err := os.ReadFile(filepath.Join(dir, "contract.md"))
	if err != nil {
		t.Fatal(err)
	}
	for language, name := range map[string]string{"quint": "model/acquisition.qnt", "yaml": "acquisition.binding.yaml"} {
		data, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(document), "```"+language+"\n"+string(data)+"```\n") {
			t.Fatalf("document fence drift: %s", name)
		}
	}
	t.Setenv("GOO_GOTOGETHER_REQUESTS", filepath.Join(dir, "requests.jsonl"))
	TestRecordedJourney(t)
}
