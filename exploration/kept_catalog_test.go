package exploration

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

func populateKeeps(t *testing.T, n int) (KeptObservationStore, *ResultSnapshots, KeepRecordsIntent) {
	t.Helper()
	s, bank, _, in, _ := keptRecordFixture(t)
	for i := 0; i < n; i++ {
		in.RequestID = fmt.Sprintf("%064x", i+1)
		s.now = func() time.Time { return time.UnixMilli(300000 + int64(i)).UTC() }
		if _, e := s.KeepRecords(context.Background(), bank, CaptureStore{}, in); e != nil {
			t.Fatal(e)
		}
	}
	return s, bank, in
}
func TestKeptCatalogPagingRepairAndReadOnlyBoundary(t *testing.T) {
	s, bank, in := populateKeeps(t, 25)
	ctx := context.Background()
	page, e := s.Search(ctx, KeptSearchQuery{}, "")
	if e != nil || len(page.Items) != 20 || !page.HasMore || page.Total != 25 || page.Items[0].ID != fmt.Sprintf("%064x", 25) {
		t.Fatal(page, e)
	}
	if page.Items[0].Environment != nil || page.Items[0].EnvironmentSelector.Scope.Mode != "unreported" {
		t.Fatal("lost selector")
	}
	next, e := s.Search(ctx, KeptSearchQuery{}, page.NextCursor)
	if e != nil || len(next.Items) != 5 || next.HasMore || next.Items[0].ID != fmt.Sprintf("%064x", 5) {
		t.Fatal(next, e)
	}
	if _, e = s.Search(ctx, KeptSearchQuery{Signal: "spans"}, page.NextCursor); !errors.Is(e, ErrKeptCursor) {
		t.Fatal("cross-filter cursor", e)
	}
	in.RequestID = fmt.Sprintf("%064x", 26)
	if _, e = s.KeepRecords(ctx, bank, CaptureStore{}, in); e != nil {
		t.Fatal(e)
	}
	if _, e = s.Search(ctx, KeptSearchQuery{}, page.NextCursor); !errors.Is(e, ErrKeptCursor) {
		t.Fatal("changed-store cursor", e)
	}
	os.RemoveAll(s.catalogDir())
	before, _ := os.ReadDir(s.Dir)
	if _, e = s.Search(ctx, KeptSearchQuery{}, ""); !errors.Is(e, ErrKeptCatalog) {
		t.Fatal(e)
	}
	after, _ := os.ReadDir(s.Dir)
	if len(before) != len(after) {
		t.Fatal("GET repaired metadata")
	}
	if _, e = s.Open(in.RequestID); e != nil {
		t.Fatal("index outage broke exact open", e)
	}
	r, e := s.RepairCatalog(ctx, "")
	if e != nil || r.Indexed != 26 || r.HasMore {
		t.Fatal(r, e)
	}
	if _, e = s.Search(ctx, KeptSearchQuery{}, ""); e != nil {
		t.Fatal(e)
	}
}
func TestKeptCatalogMetadataIsNotRecordVerification(t *testing.T) {
	s, _, in := populateKeeps(t, 1)
	path := filepath.Join(s.Dir, in.RequestID+".json")
	before, _ := os.Stat(path)
	raw, _ := os.ReadFile(path)
	broken := strings.Replace(string(raw), "same", "oops", 1)
	os.WriteFile(path, []byte(broken), 0600)
	os.Chtimes(path, before.ModTime(), before.ModTime())
	if _, e := s.Search(context.Background(), KeptSearchQuery{}, ""); e != nil {
		t.Fatal("metadata-only boundary changed", e)
	}
	if _, e := s.Open(in.RequestID); e == nil {
		t.Fatal("corrupt exact payload accepted")
	}
	if _, e := s.RepairCatalog(context.Background(), ""); e == nil {
		t.Fatal("corrupt payload reindexed")
	}
	after, _ := os.ReadFile(path)
	if string(after) != broken {
		t.Fatal("repair changed keep bytes")
	}
}
func TestKeptCatalogRepairResumesBeyondWholeStoreBudget(t *testing.T) {
	s, bank, in := populateKeeps(t, 0)
	// Twelve individually valid near-3MiB receipts exceed a whole-store32MiB cap.
	huge := json.RawMessage(`{"timestamp_ns":"150000000000","Body":"` + strings.Repeat("x", 1500000) + `","ResourceAttributes":{}}`)
	r := &Result{Question: Query{Signal: "logs", FromMS: 100000, ToMS: 160000, Limit: 100}, CapturedAt: time.Now().UTC().Format(time.RFC3339Nano), Basis: strings.Repeat("a", 64), State: "observed", Rows: []json.RawMessage{huge, huge}}
	_, _, _, _, base := keptRecordFixture(t)
	r.Question.EnvironmentSelector = base.Question.EnvironmentSelector
	token, e := bank.Freeze(r)
	if e == nil {
		t.Fatal("expected original page snapshot bound")
	}
	// Durable bank can hold larger selections than the short-lived result bank.
	// Construct server-valid receipts directly as preexisting durable resources.
	for i := 0; i < 12; i++ {
		v := &KeptRecords{Version: KeptRecordsVersion, ID: fmt.Sprintf("%064x", i+1), IntentDigest: strings.Repeat("b", 64), KeptAt: time.Now().UTC(), Origin: "saved-capture", CaptureID: strings.Repeat("c", 64), CaptureAt: r.CapturedAt, Question: r.Question, ObservedAt: r.CapturedAt, Basis: r.Basis, State: r.State, ReturnedRecords: 2, Members: []KeptRecord{{0, tailMembershipHash(huge), huge}, {1, tailMembershipHash(huge), huge}}}
		v.Digest = tailMembershipHash(v)
		if e = v.Validate(); e != nil {
			t.Fatal(e)
		}
		raw, _ := json.Marshal(v)
		if len(raw) > KeptObservationsMaxBytes {
			t.Fatal("fixture too large")
		}
		os.WriteFile(filepath.Join(s.Dir, v.ID+".json"), raw, 0600)
	}
	_ = in
	_ = token
	ctx := context.Background()
	first, e := s.RepairCatalog(ctx, "")
	if e != nil || !first.HasMore || first.Bytes > KeptRepairBytes || first.Indexed == 0 {
		t.Fatal(first, e)
	}
	second, e := s.RepairCatalog(ctx, first.NextCursor)
	if e != nil || second.HasMore || second.Completed != 12 {
		t.Fatal(second, e)
	}
	if page, e := s.Search(ctx, KeptSearchQuery{}, ""); e != nil || page.Total != 12 {
		t.Fatal(page, e)
	}
}
func TestKeptCatalogCapacityCoordinatesIndependentPublishers(t *testing.T) {
	s, bank, in := populateKeeps(t, 1)
	// Junk consumes physical admission, so only one further slot remains.
	for i := 0; i < KeptCatalogLimit-2; i++ {
		os.WriteFile(filepath.Join(s.Dir, fmt.Sprintf("junk-%d", i)), nil, 0600)
	}
	var wg sync.WaitGroup
	errs := make(chan error, 8)
	for i := 0; i < 8; i++ {
		wg.Go(func() {
			candidate := in
			candidate.RequestID = fmt.Sprintf("%064x", i+2)
			_, e := (KeptObservationStore{Dir: s.Dir}).KeepRecords(context.Background(), bank, CaptureStore{}, candidate)
			errs <- e
		})
	}
	wg.Wait()
	close(errs)
	success := 0
	for e := range errs {
		if e == nil {
			success++
		} else if !errors.Is(e, ErrKeepCapacity) {
			t.Fatal(e)
		}
	}
	if success != 1 {
		t.Fatal("admission race", success)
	}
	if _, e := s.KeepRecords(context.Background(), nil, CaptureStore{}, in); e != nil {
		t.Fatal("duplicate at capacity", e)
	}
	if _, e := s.Search(context.Background(), KeptSearchQuery{}, ""); e != nil {
		t.Fatal("accepted last slot disabled discovery", e)
	}
	os.WriteFile(filepath.Join(s.Dir, "overflow"), nil, 0600)
	if _, e := s.Search(context.Background(), KeptSearchQuery{}, ""); !errors.Is(e, ErrKeepCapacity) {
		t.Fatal(e)
	}
	if _, e := s.Open(in.RequestID); e != nil {
		t.Fatal(e)
	}
}
func TestKeptCatalogCancellationAndMetadataBounds(t *testing.T) {
	s, _, in := populateKeeps(t, 1)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, e := s.Search(ctx, KeptSearchQuery{}, ""); !errors.Is(e, context.Canceled) {
		t.Fatal(e)
	}
	path := filepath.Join(s.catalogDir(), in.RequestID+".json")
	os.WriteFile(path, []byte(strings.Repeat("x", KeptCatalogEntryBytes+1)), 0600)
	if _, e := s.Search(context.Background(), KeptSearchQuery{}, ""); !errors.Is(e, ErrKeptCatalog) {
		t.Fatal(e)
	}
	if _, e := s.RepairCatalog(context.Background(), ""); e != nil {
		t.Fatal(e)
	}
	os.Remove(path)
	os.Symlink(filepath.Join(s.Dir, in.RequestID+".json"), path)
	if _, e := s.Search(context.Background(), KeptSearchQuery{}, ""); !errors.Is(e, ErrKeptCatalog) {
		t.Fatal(e)
	}
}

func TestKeptCatalogProcessWorker(t *testing.T) {
	root := os.Getenv("GOO_KEEP_WORKER_ROOT")
	if root == "" {
		t.Skip("subprocess helper")
	}
	_, e := (KeptObservationStore{Dir: root}).KeepRecords(context.Background(), nil, CaptureStore{Dir: os.Getenv("GOO_KEEP_WORKER_CAPTURES")}, KeepRecordsIntent{Version: KeepRecordsVersion, RequestID: os.Getenv("GOO_KEEP_WORKER_ID"), CaptureID: os.Getenv("GOO_KEEP_WORKER_CAPTURE"), Indexes: []int{0}})
	if errors.Is(e, ErrKeepCapacity) {
		os.Exit(10)
	}
	if e != nil {
		t.Fatal(e)
	}
}
func TestKeptCatalogCrossProcessLastSlot(t *testing.T) {
	s, _, _ := populateKeeps(t, 1)
	for i := 0; i < KeptCatalogLimit-2; i++ {
		os.WriteFile(filepath.Join(s.Dir, fmt.Sprintf("junk-%d", i)), nil, 0600)
	}
	_, _, _, _, r := keptRecordFixture(t)
	r.Question.Offset = 0
	captures := CaptureStore{Dir: t.TempDir()}
	saved, e := captures.publish(SavedCapture{CaptureReceipt: CaptureReceipt{Version: 1, Title: "worker source", CapturedAt: r.CapturedAt, Question: CaptureQuestion{Kind: "explore", Explore: &r.Question}}, Explore: r})
	if e != nil {
		t.Fatal(e)
	}
	binary, e := os.Executable()
	if e != nil {
		t.Fatal(e)
	}
	var wg sync.WaitGroup
	codes := make(chan int, 2)
	for i := 0; i < 2; i++ {
		wg.Go(func() {
			cmd := exec.Command(binary, "-test.run=^TestKeptCatalogProcessWorker$")
			cmd.Env = append(os.Environ(), "GOO_KEEP_WORKER_ROOT="+s.Dir, "GOO_KEEP_WORKER_CAPTURES="+captures.Dir, "GOO_KEEP_WORKER_CAPTURE="+saved.ID, "GOO_KEEP_WORKER_ID="+fmt.Sprintf("%064x", i+2))
			out, e := cmd.CombinedOutput()
			if e == nil {
				codes <- 0
				return
			}
			var exit *exec.ExitError
			if errors.As(e, &exit) && exit.ExitCode() == 10 {
				codes <- 10
				return
			}
			t.Errorf("worker %v %s", e, out)
			codes <- 99
		})
	}
	wg.Wait()
	close(codes)
	success, full := 0, 0
	for code := range codes {
		if code == 0 {
			success++
		}
		if code == 10 {
			full++
		}
	}
	if success != 1 || full != 1 {
		t.Fatal("cross-process admission", success, full)
	}
}

func TestKeptRepairReadsUseActualRemainingBudget(t *testing.T) {
	s, _, in := populateKeeps(t, 1)
	if _, n, e := s.loadBounded(in.RequestID, 1); e == nil || n != 0 {
		t.Fatal("stat size bypassed remaining actual read budget", n, e)
	}
	raw, e := s.Open(in.RequestID)
	if e != nil {
		t.Fatal(e)
	}
	if _, n, e := s.loadBounded(in.RequestID, int64(len(raw))); e != nil || n != int64(len(raw)) {
		t.Fatal("exact bounded read", n, e)
	}
}
