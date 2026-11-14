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
	for snapshot, source := range map[string]string{"model/acquisition.qnt": "model/acquisition.qnt", "acquisition.binding.yaml": binding, "scenario.go.txt": "shadow.go"} {
		a, err := os.ReadFile(filepath.Join(dir, snapshot))
		if err != nil {
			t.Fatal(err)
		}
		b, err := os.ReadFile(source)
		if err != nil {
			t.Fatal(err)
		}
		if string(a) != string(b) {
			t.Fatalf("refresh CW1 snapshot from %s", source)
		}
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
