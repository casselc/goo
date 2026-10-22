package exploration

import (
	"encoding/json"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestResultSnapshotCustodyExpiryAndCapacity(t *testing.T) {
	now := time.Unix(0, 0)
	s := ResultSnapshots{now: func() time.Time { return now }}
	r := &Result{Question: Query{Signal: "spans", Environment: "prod", FromMS: 1000, ToMS: 2000, Limit: 100, Offset: 100}, Rows: []json.RawMessage{json.RawMessage(`{"n":9007199254740993,"timestamp_ns":"1000000001"}`)}, Basis: "original", CapturedAt: "2026-10-22T00:00:00Z"}
	id, e := s.Freeze(r)
	if e != nil {
		t.Fatal(e)
	}
	r.Rows[0][2] = 'x'
	r.Question.Environment = "changed"
	c, e := s.Load(id)
	if e != nil || c.Question.Explore.Environment != "prod" || c.Question.Explore.Offset != 100 || !strings.Contains(string(c.Explore.Rows[0]), "9007199254740993") {
		t.Fatal(c, e)
	}
	c.Explore.Rows[0][2] = 'z'
	again, _ := s.Load(id)
	if string(again.Explore.Rows[0]) != `{"n":9007199254740993,"timestamp_ns":"1000000001"}` {
		t.Fatal("load aliases retained bytes")
	}
	now = now.Add(ResultSnapshotTTL)
	if _, e = s.Load(id); e == nil {
		t.Fatal("expiry accepted")
	}
	if s.bytes != 0 {
		t.Fatal("expired memory retained")
	}
	r.Rows = nil
	for i := 0; i < 64; i++ {
		if _, e = s.Freeze(r); e != nil {
			t.Fatal(e)
		}
	}
	if _, e = s.Freeze(r); e == nil {
		t.Fatal("unbounded entries")
	}
	now = now.Add(ResultSnapshotTTL)
	if _, e = s.Freeze(r); e != nil {
		t.Fatal("expired capacity not reclaimed", e)
	}
	r.Rows = []json.RawMessage{json.RawMessage(`"` + strings.Repeat("a", ResultSnapshotMaxBytes) + `"`)}
	if _, e = s.Freeze(r); e == nil {
		t.Fatal("oversized page accepted")
	}
}
func TestResultSnapshotsConcurrentReaders(t *testing.T) {
	s := new(ResultSnapshots)
	r := &Result{Question: Query{Signal: "logs", Environment: "prod", FromMS: 1, ToMS: 2, Limit: 100}, Rows: []json.RawMessage{json.RawMessage(`{"message":"same"}`)}}
	var wg sync.WaitGroup
	for i := 0; i < 32; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			id, e := s.Freeze(r)
			if e != nil {
				t.Error(e)
				return
			}
			if _, e = s.Load(id); e != nil {
				t.Error(e)
			}
		}()
	}
	wg.Wait()
}
