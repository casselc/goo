package exploration

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/casselc/goo/identity"
)

func keptFixture(t *testing.T) (KeptObservationStore, *TailMembershipSnapshots, *time.Time, KeepTailIntent) {
	t.Helper()
	now := time.UnixMilli(200000).UTC()
	bank := NewTailMembershipSnapshots(func() time.Time { return now })
	tracker := TailTracker{}
	wire := json.RawMessage(`{"timestamp_ns":"150000000000","Body":"same","ResourceAttributes":{},"exact":18446744073709551615}`)
	rows, err := tracker.Observe([]json.RawMessage{wire, wire})
	if err != nil {
		t.Fatal(err)
	}
	scope, _ := identity.EffectiveEnvironment(identity.EnvironmentScope{Mode: "unreported"})
	dto, _ := scope.Selector()
	selector := TailSelector{Signal: "logs", EnvironmentSelector: &dto}
	cap, err := bank.Freeze(TailMembershipBatch{Connection: strings.Repeat("a", 64), Selector: selector, FromMS: 100000, ToMS: 160000, ObservedAt: now, Rows: rows})
	if err != nil {
		t.Fatal(err)
	}
	in := KeepTailIntent{Version: KeepTailVersion, RequestID: strings.Repeat("b", 64), Connection: strings.Repeat("a", 64), Selector: selector, Refs: []TailMembershipRef{{Token: cap.Token, Key: rows[0].Key}, {Token: cap.Token, Key: rows[1].Key}}}
	return KeptObservationStore{Dir: t.TempDir(), now: func() time.Time { return now }}, bank, &now, in
}
func TestKeepObservedSelectionRetriesAfterExpiryAndRestart(t *testing.T) {
	store, bank, now, in := keptFixture(t)
	v, err := store.Keep(context.Background(), bank, in)
	if err != nil {
		t.Fatal(err)
	}
	if len(v.Members) != 2 || v.Members[0].Key == v.Members[1].Key || !strings.Contains(string(v.Members[0].Record), "18446744073709551615") {
		t.Fatal(v)
	}
	wire, _ := json.Marshal(v)
	if strings.Contains(string(wire), in.Refs[0].Token) {
		t.Fatal("live proof token leaked")
	}
	*now = now.Add(3 * time.Minute)
	again, err := (KeptObservationStore{Dir: store.Dir}).Keep(context.Background(), nil, in)
	if err != nil {
		t.Fatal(err)
	}
	a, _ := json.Marshal(again)
	if string(a) != string(wire) {
		t.Fatal("retry changed kept identity/content/clock")
	}
	changed := in
	changed.Refs = changed.Refs[:1]
	if _, err = store.Keep(context.Background(), nil, changed); !errors.Is(err, ErrKeepConflict) {
		t.Fatal(err)
	}
	in.RequestID = strings.Repeat("c", 64)
	if _, err = store.Keep(context.Background(), bank, in); !errors.Is(err, ErrTailMembershipUnavailable) {
		t.Fatal(err)
	}
}
func TestKeepLostAcknowledgmentAndConcurrentWinners(t *testing.T) {
	store, bank, _, in := keptFixture(t)
	store.afterPublish = func() error { return errors.New("lost acknowledgment") }
	if _, err := store.Keep(context.Background(), bank, in); err == nil {
		t.Fatal("expected unknown acknowledgment")
	}
	store.afterPublish = nil
	var wg sync.WaitGroup
	errs := make(chan error, 20)
	for range 20 {
		wg.Go(func() {
			v, e := store.Keep(context.Background(), nil, in)
			if e == nil && v.ID != in.RequestID {
				e = errors.New("wrong winner")
			}
			errs <- e
		})
	}
	wg.Wait()
	close(errs)
	for e := range errs {
		if e != nil {
			t.Fatal(e)
		}
	}
	files, _ := os.ReadDir(store.Dir)
	if len(files) != 1 {
		t.Fatal(files)
	}
	// Also race first publication, not only already committed retries.
	in.RequestID = strings.Repeat("d", 64)
	errs = make(chan error, 20)
	for range 20 {
		wg.Go(func() { _, e := store.Keep(context.Background(), bank, in); errs <- e })
	}
	wg.Wait()
	close(errs)
	for e := range errs {
		if e != nil {
			t.Fatal(e)
		}
	}
}
func TestKeepRejectsForeignProofCorruptWinnerAndMalformedStorage(t *testing.T) {
	store, bank, _, in := keptFixture(t)
	bad := in
	bad.Connection = strings.Repeat("c", 64)
	if _, err := store.Keep(context.Background(), bank, bad); err == nil {
		t.Fatal("foreign connection accepted")
	}
	if _, err := store.Keep(context.Background(), bank, in); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(store.Dir, in.RequestID+".json")
	raw, _ := os.ReadFile(path)
	broken := strings.Replace(string(raw), "same", "changed", 1)
	os.WriteFile(path, []byte(broken), 0600)
	if _, err := store.Keep(context.Background(), bank, in); err == nil {
		t.Fatal("corrupt winner overwritten")
	}
	after, _ := os.ReadFile(path)
	if string(after) != broken {
		t.Fatal("corrupt data changed")
	}
	os.Remove(path)
	os.Symlink("missing", path)
	if _, err := store.Load(in.RequestID); err == nil {
		t.Fatal("symlink accepted")
	}
	bad = in
	bad.Refs = append(append([]TailMembershipRef{}, in.Refs...), in.Refs[0])
	if bad.Validate() == nil {
		t.Fatal("duplicate occurrence")
	}
}
