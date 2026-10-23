package exploration

import (
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"
)

func membershipFixture(t *testing.T) (*TailMembershipSnapshots, *time.Time, TailMembershipBatch) {
	t.Helper()
	now := time.UnixMilli(200000).UTC()
	tracker := TailTracker{}
	rows, err := tracker.Observe([]json.RawMessage{json.RawMessage(`{"timestamp_ns":"150000000000","Body":"same","exact":18446744073709551615}`), json.RawMessage(`{"Body":"same","exact":18446744073709551615,"timestamp_ns":"150000000000"}`)})
	if err != nil {
		t.Fatal(err)
	}
	return NewTailMembershipSnapshots(func() time.Time { return now }), &now, TailMembershipBatch{Connection: strings.Repeat("a", 64), Selector: TailSelector{Environment: "test"}, FromMS: 100000, ToMS: 160000, ObservedAt: now, Rows: rows}
}
func TestTailMembershipExactOccurrencesAndProvenance(t *testing.T) {
	s, now, batch := membershipFixture(t)
	cap, err := s.Freeze(batch)
	if err != nil {
		t.Fatal(err)
	}
	key1, key2 := batch.Rows[0].Key, batch.Rows[1].Key
	refs := []TailMembershipRef{{Token: cap.Token, Key: key1}, {Token: cap.Token, Key: key2}}
	draft, err := s.Resolve(batch.Connection, batch.Selector, refs)
	if err != nil || len(draft.Members) != 2 {
		t.Fatal(draft, err)
	}
	if draft.Members[0].FromMS != 100000 || draft.Members[0].ToMS != 160000 || !draft.Members[0].ObservedAt.Equal(*now) || draft.Members[0].Basis != cap.Basis || !strings.Contains(string(draft.Members[0].Record), "18446744073709551615") {
		t.Fatal("lost original provenance or numeric precision")
	}
	refs[0], refs[1] = refs[1], refs[0]
	reverse, err := s.Resolve(batch.Connection, TailSelector{Signal: "logs", Environment: "test"}, refs)
	if err != nil || !reflect.DeepEqual(draft, reverse) {
		t.Fatal("selection order or default signal changed intent", err)
	}
	// Input and returned objects cannot mutate the cache.
	batch.Rows[0].Record[0] = '!'
	draft.Members[0].Record[0] = '!'
	reverse.Members[0].Record[0] = '!'
	again, err := s.Resolve(batch.Connection, batch.Selector, refs)
	if err != nil || !json.Valid(again.Members[0].Record) {
		t.Fatal("cache aliased caller", err)
	}
	*now = now.Add(TailMembershipTTL)
	if got, err := s.Resolve(batch.Connection, batch.Selector, refs); got != nil || !errors.Is(err, ErrTailMembershipUnavailable) {
		t.Fatal("expiry did not fail closed", got, err)
	}
}
func TestTailMembershipRejectsScopeAndPartialSelections(t *testing.T) {
	s, now, batch := membershipFixture(t)
	cap, err := s.Freeze(batch)
	if err != nil {
		t.Fatal(err)
	}
	ref := TailMembershipRef{Token: cap.Token, Key: batch.Rows[0].Key}
	for _, tc := range []struct {
		connection string
		selector   TailSelector
		refs       []TailMembershipRef
	}{
		{strings.Repeat("b", 64), batch.Selector, []TailMembershipRef{ref}},
		{batch.Connection, TailSelector{Signal: "spans", Environment: "test"}, []TailMembershipRef{ref}},
		{batch.Connection, TailSelector{Environment: "other"}, []TailMembershipRef{ref}},
		{batch.Connection, TailSelector{Environment: "test", Where: &Predicate{Field: &Field{Kind: "attribute", Key: "tenant"}, Op: "eq", Value: ""}}, []TailMembershipRef{ref}},
		{batch.Connection, batch.Selector, nil},
		{batch.Connection, batch.Selector, make([]TailMembershipRef, 21)},
		{batch.Connection, batch.Selector, []TailMembershipRef{ref, {Token: cap.Token, Key: "forged"}}},
		{batch.Connection, batch.Selector, []TailMembershipRef{ref, ref}},
		{batch.Connection, batch.Selector, []TailMembershipRef{ref, {Token: strings.Repeat("c", 64), Key: ref.Key}}},
	} {
		if d, e := s.Resolve(tc.connection, tc.selector, tc.refs); e == nil || d != nil {
			t.Fatal("invalid or partial selection published", tc, d, e)
		}
	}
	alias, err := s.Freeze(batch)
	if err != nil {
		t.Fatal(err)
	}
	if d, e := s.Resolve(batch.Connection, batch.Selector, []TailMembershipRef{ref, {Token: alias.Token, Key: ref.Key}}); d != nil || e == nil {
		t.Fatal("same occurrence counted twice through alias")
	}
	changed := batch
	changed.Selector = TailSelector{Environment: "other"}
	if _, e := s.Freeze(changed); e == nil {
		t.Fatal("connection reused for changed scope")
	}
	*now = now.Add(TailMembershipTTL)
	if _, e := s.Freeze(batch); !errors.Is(e, ErrTailMembershipUnavailable) {
		t.Fatal("delayed issuance extended old read lifetime", e)
	}
}
func TestTailMembershipBoundsAndNoEviction(t *testing.T) {
	s, now, batch := membershipFixture(t)
	first, err := s.Freeze(batch)
	if err != nil {
		t.Fatal(err)
	}
	for i := 1; i < TailMembershipMaxBatches; i++ {
		if _, e := s.Freeze(batch); e != nil {
			t.Fatal(e)
		}
	}
	if _, e := s.Freeze(batch); !errors.Is(e, ErrTailMembershipCapacity) {
		t.Fatal("capacity not bounded", e)
	}
	if _, e := s.Resolve(batch.Connection, batch.Selector, []TailMembershipRef{{Token: first.Token, Key: batch.Rows[0].Key}}); e != nil {
		t.Fatal("capacity evicted a valid capability", e)
	}
	*now = now.Add(TailMembershipTTL)
	batch.ObservedAt = *now
	if _, e := s.Freeze(batch); e != nil {
		t.Fatal("expired capacity not reclaimed", e)
	}
	if len(s.entries) != 1 {
		t.Fatal("expired entries retained")
	}
	// Byte pressure refuses a new batch before count capacity and keeps old proof.
	byteStore, _, large := membershipFixture(t)
	body := strings.Repeat("x", 512<<10)
	raw, _ := json.Marshal(map[string]string{"timestamp_ns": "150000000000", "Body": body})
	tracker := TailTracker{}
	large.Rows, _ = tracker.Observe([]json.RawMessage{raw})
	count := 0
	for ; count < TailMembershipMaxBatches; count++ {
		_, e := byteStore.Freeze(large)
		if errors.Is(e, ErrTailMembershipCapacity) {
			break
		}
		if e != nil {
			t.Fatal(e)
		}
	}
	if count == 0 || count == TailMembershipMaxBatches {
		t.Fatal("byte pressure did not bound capacity independently", count)
	}
}
func TestTailMembershipRejectsUnprovenRows(t *testing.T) {
	for _, change := range []func(*TailMembershipBatch){
		func(b *TailMembershipBatch) { b.Rows[0].Key = "forged:1" },
		func(b *TailMembershipBatch) { b.Rows = append(b.Rows, b.Rows[0]) },
		func(b *TailMembershipBatch) { b.Rows[0].Record = json.RawMessage(`{"timestamp_ns":"160000000000"}`) },
		func(b *TailMembershipBatch) { b.Rows[0].Record = append(b.Rows[0].Record, []byte(" {}")...) },
		func(b *TailMembershipBatch) { b.FromMS++ },
		func(b *TailMembershipBatch) { b.ObservedAt = time.UnixMilli(150000) },
		func(b *TailMembershipBatch) { b.ObservedAt = time.UnixMilli(200001) },
		func(b *TailMembershipBatch) { b.Rows = nil },
	} {
		s, _, batch := membershipFixture(t)
		change(&batch)
		if _, e := s.Freeze(batch); e == nil {
			t.Fatal("invalid batch accepted")
		}
		if len(s.entries) != 0 || s.bytes != 0 {
			t.Fatal("invalid batch partially allocated")
		}
	}
	s, _, batch := membershipFixture(t)
	body := strings.Repeat("x", TailMembershipBatchBytes)
	raw, _ := json.Marshal(map[string]string{"timestamp_ns": "150000000000", "Body": body})
	tracker := TailTracker{}
	batch.Rows, _ = tracker.Observe([]json.RawMessage{raw})
	if _, e := s.Freeze(batch); e == nil {
		t.Fatal("oversize batch accepted")
	}
}
func TestTailMembershipKeepsPerPollProvenanceAndConcurrentReads(t *testing.T) {
	s, now, batch := membershipFixture(t)
	first, err := s.Freeze(batch)
	if err != nil {
		t.Fatal(err)
	}
	other := batch
	other.FromMS += 1000
	other.ToMS += 1000
	other.ObservedAt = now.Add(time.Second)
	other.Limited = true
	tracker := TailTracker{}
	other.Rows, _ = tracker.Observe([]json.RawMessage{json.RawMessage(`{"timestamp_ns":"149000000000","Body":"late"}`)})
	other.Rows[0].Late = true
	*now = other.ObservedAt
	second, err := s.Freeze(other)
	if err != nil {
		t.Fatal(err)
	}
	refs := []TailMembershipRef{{Token: first.Token, Key: batch.Rows[0].Key}, {Token: second.Token, Key: other.Rows[0].Key}}
	draft, err := s.Resolve(batch.Connection, batch.Selector, refs)
	if err != nil {
		t.Fatal(err)
	}
	windows := map[int64]bool{}
	for _, m := range draft.Members {
		windows[m.ToMS] = true
		if m.ToMS == other.ToMS && (!m.Limited || !m.Late || m.Basis != second.Basis) {
			t.Fatal("lost late/limited read provenance")
		}
	}
	if len(windows) != 2 {
		t.Fatal("different read windows collapsed")
	}
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, e := s.Freeze(batch); e != nil {
				t.Error("concurrent freeze failed", e)
			}
			got, e := s.Resolve(batch.Connection, batch.Selector, refs)
			if e != nil || got.Digest != draft.Digest {
				t.Error("concurrent resolve changed draft", e)
			}
		}()
	}
	wg.Wait()
}

func TestTailMembershipRestartAndDelayedIssuance(t *testing.T) {
	s, now, batch := membershipFixture(t)
	first, e := s.Freeze(batch)
	if e != nil {
		t.Fatal(e)
	}
	*now = now.Add(30 * time.Second)
	second, e := s.Freeze(batch)
	if e != nil || !second.ExpiresAt.Equal(first.ExpiresAt) {
		t.Fatal("delayed issuance renewed original read", e)
	}
	refs := []TailMembershipRef{{Token: first.Token, Key: batch.Rows[0].Key}}
	restarted := NewTailMembershipSnapshots(func() time.Time { return *now })
	if d, e := restarted.Resolve(batch.Connection, batch.Selector, refs); d != nil || !errors.Is(e, ErrTailMembershipUnavailable) {
		t.Fatal("restart fabricated an uncommitted proof", d, e)
	}
	if d, e := s.Resolve(batch.Connection, batch.Selector, refs); e != nil || d == nil {
		t.Fatal("new cache affected original proof", e)
	}
}
