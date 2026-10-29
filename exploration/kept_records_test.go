package exploration

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/casselc/goo/identity"
)

func keptRecordFixture(t *testing.T) (KeptObservationStore, *ResultSnapshots, *time.Time, KeepRecordsIntent, *Result) {
	t.Helper()
	now := time.UnixMilli(200000).UTC()
	bank := &ResultSnapshots{now: func() time.Time { return now }}
	scope, _ := identity.EffectiveEnvironment(identity.EnvironmentScope{Mode: "unreported"})
	dto, _ := scope.Selector()
	q := Query{Signal: "logs", EnvironmentSelector: &dto, FromMS: 100000, ToMS: 160000, Limit: 100, Offset: 100}
	row := json.RawMessage(`{"timestamp_ns":"150000000000","Body":"same","ResourceAttributes":{},"exact":18446744073709551615}`)
	r := &Result{Question: q, Rows: []json.RawMessage{row, row}, CapturedAt: now.Format(time.RFC3339Nano), Basis: strings.Repeat("a", 64), State: "changed_between_reads", HasMore: true, Count: "700"}
	token, e := bank.Freeze(r)
	if e != nil {
		t.Fatal(e)
	}
	return KeptObservationStore{Dir: t.TempDir()}, bank, &now, KeepRecordsIntent{Version: KeepRecordsVersion, RequestID: strings.Repeat("c", 64), ResultToken: token, Indexes: []int{1, 0}}, r
}
func TestKeepRecordsPageExactDuplicatesRetryAndCrossOrigin(t *testing.T) {
	s, bank, now, in, r := keptRecordFixture(t)
	v, e := s.KeepRecords(context.Background(), bank, CaptureStore{}, in)
	if e != nil {
		t.Fatal(e)
	}
	if !reflect.DeepEqual(v.Question, r.Question) || v.ObservedAt != r.CapturedAt || v.State != r.State || !v.RecordsLimited || v.Members[0].Index != 0 || v.Members[1].Index != 1 || v.Members[0].Digest != v.Members[1].Digest || !strings.Contains(string(v.Members[0].Record), "18446744073709551615") {
		t.Fatal("lost original page semantics")
	}
	wire, _ := s.Open(v.ID)
	if strings.Contains(string(wire), in.ResultToken) {
		t.Fatal("live capability leaked")
	}
	*now = now.Add(ResultSnapshotTTL)
	again, e := (KeptObservationStore{Dir: s.Dir}).KeepRecords(context.Background(), nil, CaptureStore{}, in)
	if e != nil || !reflect.DeepEqual(v, again) {
		t.Fatal("restart changed winner", e)
	}
	changed := in
	changed.Indexes = []int{0}
	if _, e = s.KeepRecords(context.Background(), nil, CaptureStore{}, changed); !errors.Is(e, ErrKeepConflict) {
		t.Fatal(e)
	}
	changed = in
	changed.RequestID = strings.Repeat("d", 64)
	if _, e = s.KeepRecords(context.Background(), bank, CaptureStore{}, changed); !errors.Is(e, ErrKeptRecordsUnavailable) {
		t.Fatal(e)
	}
	tail := KeepTailIntent{Version: KeepTailVersion, RequestID: in.RequestID, Connection: strings.Repeat("e", 64), Selector: TailSelector{Environment: "prod"}, Refs: []TailMembershipRef{{Token: strings.Repeat("f", 64), Key: strings.Repeat("a", 64) + ":1"}}}
	if _, e = s.Keep(context.Background(), nil, tail); !errors.Is(e, ErrKeepConflict) {
		t.Fatal("cross-origin overwrite", e)
	}
}
func TestKeepSavedRecordsAbsoluteIndexAndDeletedSourceRecovery(t *testing.T) {
	s, _, _, _, r := keptRecordFixture(t)
	r.Question.Offset = 0
	r.Question.Limit = 100
	r.Rows = make([]json.RawMessage, 102)
	for i := range r.Rows {
		r.Rows[i] = json.RawMessage(`{"timestamp_ns":"150000000000","Body":"same","ResourceAttributes":{}}`)
	}
	captures := CaptureStore{Dir: t.TempDir()}
	c, e := captures.publish(SavedCapture{CaptureReceipt: CaptureReceipt{Version: 1, Title: "original", CapturedAt: "2026-10-29T00:00:00Z", Question: CaptureQuestion{Kind: "explore", Explore: &r.Question}}, Explore: r})
	if e != nil {
		t.Fatal(e)
	}
	page, e := c.Page(captureCursor(c.ID, 100))
	if e != nil || page.Explore.Question.Offset != 100 {
		t.Fatal(e)
	}
	in := KeepRecordsIntent{Version: KeepRecordsVersion, RequestID: strings.Repeat("b", 64), CaptureID: c.ID, Indexes: []int{101, 100}}
	v, e := s.KeepRecords(context.Background(), nil, captures, in)
	if e != nil {
		t.Fatal(e)
	}
	if v.Origin != "saved-capture" || v.CaptureID != c.ID || v.CaptureAt != c.CapturedAt || v.ObservedAt != r.CapturedAt || v.Members[0].Index != 100 || v.Question.Offset != 0 {
		t.Fatal("capture index/clock drift")
	}
	os.Remove(filepath.Join(captures.Dir, c.ID+".json"))
	if _, e = s.KeepRecords(context.Background(), nil, captures, in); e != nil {
		t.Fatal("deleted origin blocked winner", e)
	}
}
func TestKeepRecordsConcurrentLostAckAndInputRefusal(t *testing.T) {
	s, bank, _, in, _ := keptRecordFixture(t)
	s.afterPublish = func() error { return errors.New("lost acknowledgment") }
	if _, e := s.KeepRecords(context.Background(), bank, CaptureStore{}, in); e == nil {
		t.Fatal("expected lost ack")
	}
	s.afterPublish = nil
	var wg sync.WaitGroup
	errs := make(chan error, 20)
	for range 20 {
		wg.Go(func() { _, e := s.KeepRecords(context.Background(), nil, CaptureStore{}, in); errs <- e })
	}
	wg.Wait()
	close(errs)
	for e := range errs {
		if e != nil {
			t.Fatal(e)
		}
	}
	for _, mutate := range []func(*KeepRecordsIntent){func(v *KeepRecordsIntent) { v.CaptureID = strings.Repeat("a", 64) }, func(v *KeepRecordsIntent) { v.ResultToken = "" }, func(v *KeepRecordsIntent) { v.Indexes = []int{0, 0} }, func(v *KeepRecordsIntent) { v.Indexes = []int{-1} }, func(v *KeepRecordsIntent) { v.Indexes = make([]int, 21) }} {
		bad := in
		mutate(&bad)
		if bad.Validate() == nil {
			t.Fatal("invalid selection accepted")
		}
	}
}

func TestKeepRecordsExpectedScopePreconditionCannotReplaceAttestedRows(t *testing.T) {
	s, bank, _, in, r := keptRecordFixture(t)
	in.Expected = &KeepRecordExpectation{Question: r.Question, Basis: r.Basis}
	in.Expected.Question.Offset = 0
	if _, e := s.KeepRecords(context.Background(), bank, CaptureStore{}, in); e == nil {
		t.Fatal("substituted expected question accepted")
	}
	in.Expected.Question = r.Question
	in.Expected.Basis = strings.Repeat("d", 64)
	if _, e := s.KeepRecords(context.Background(), bank, CaptureStore{}, in); e == nil {
		t.Fatal("substituted expected basis accepted")
	}
	if files, _ := os.ReadDir(s.Dir); len(files) != 0 {
		t.Fatal("failed precondition published")
	}
	in.Expected.Basis = r.Basis
	if _, e := s.KeepRecords(context.Background(), bank, CaptureStore{}, in); e != nil {
		t.Fatal(e)
	}
}
