package exploration

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// Generated with the pre-250 catalog/summary implementation in a Go overlay,
// using the unchanged stored-capture fixture and a fixed filesystem timestamp.
func TestLegacyCatalogEntryBytes250(t *testing.T) {
	raw, err := os.ReadFile("../testdata/storage/exploration-capture.json")
	if err != nil {
		t.Fatal(err)
	}
	c := searchFixture(t)
	store := CaptureStore{Dir: t.TempDir()}
	path := filepath.Join(store.Dir, c.ID+".json")
	if err = os.WriteFile(path, raw, 0600); err != nil {
		t.Fatal(err)
	}
	at := time.Unix(1735689600, 0)
	if err = os.Chtimes(path, at, at); err != nil {
		t.Fatal(err)
	}
	if _, err = store.RebuildCatalog(t.Context()); err != nil {
		t.Fatal(err)
	}
	actual, err := os.ReadFile(filepath.Join(store.catalogDir(), c.ID+".json"))
	if err != nil {
		t.Fatal(err)
	}
	expected, err := os.ReadFile("../testdata/storage/capture-catalog-legacy-250.json")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(actual, expected) {
		t.Fatal("old catalog bytes/digest changed", string(actual), string(expected))
	}
}
