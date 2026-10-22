package immutablefile

import (
	"bytes"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
)

func TestConcurrentPublicationHasOneCompleteWinner(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "objects")
	var wg sync.WaitGroup
	var winners atomic.Int32
	for i := 0; i < 24; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			data := bytes.Repeat([]byte{byte(i)}, 128*1024)
			created, e := Publish(dir, "object.json", ".draft-", data)
			if e != nil {
				t.Error(e)
				return
			}
			if created {
				winners.Add(1)
			}
			stored, e := os.ReadFile(filepath.Join(dir, "object.json"))
			if e != nil || len(stored) != len(data) || !bytes.Equal(stored, bytes.Repeat(stored[:1], len(data))) {
				t.Error("partial winner", e)
			}
		}(i)
	}
	wg.Wait()
	if winners.Load() != 1 {
		t.Fatal("winner count", winners.Load())
	}
	entries, e := os.ReadDir(dir)
	if e != nil || len(entries) != 1 {
		t.Fatal("scratch leak", e)
	}
	info, e := os.Stat(filepath.Join(dir, "object.json"))
	if e != nil || info.Mode().Perm() != 0600 {
		t.Fatal("file permissions", info, e)
	}
}
func TestPublicationFailureAndExistingObject(t *testing.T) {
	dir := t.TempDir()
	blocked := filepath.Join(dir, "blocked")
	if e := os.WriteFile(blocked, []byte("original"), 0600); e != nil {
		t.Fatal(e)
	}
	if _, e := Publish(blocked, "object", ".draft-", []byte("new")); e == nil {
		t.Fatal("file used as directory")
	}
	created, e := Publish(dir, "blocked", ".draft-", []byte("replacement"))
	if e != nil || created {
		t.Fatal("existing winner", created, e)
	}
	raw, e := os.ReadFile(blocked)
	if e != nil || string(raw) != "original" {
		t.Fatal("winner replaced", e)
	}
	// A missing target parent makes link fail after the temporary write succeeds.
	if _, e = Publish(dir, "missing/object", ".draft-", []byte("new")); e == nil {
		t.Fatal("link failure lost")
	}
	entries, e := os.ReadDir(dir)
	if e != nil || len(entries) != 1 {
		t.Fatal("failed link leaked scratch", e)
	}
}
