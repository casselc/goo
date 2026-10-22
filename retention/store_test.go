package retention

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/workbench"
	"os"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"
)

func sample() *exploration.SavedCapture {
	return &exploration.SavedCapture{CaptureReceipt: exploration.CaptureReceipt{Version: 1, ID: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", CapturedAt: "2026-10-22T00:00:00Z", Question: exploration.CaptureQuestion{Kind: "explore", Explore: &exploration.Query{Signal: "spans", Environment: "prod", FromMS: 1000, ToMS: 2000, Limit: 100}}}, Explore: &exploration.Result{Basis: "original-basis", Count: "100", Rows: []json.RawMessage{json.RawMessage(`{"timestamp_ns":"1792629000000000001","empty":"","n":9007199254740993}`), json.RawMessage(`{"timestamp_ns":"1792629000000000001","empty":"","n":9007199254740993}`)}}}
}
func TestSelectExactRows(t *testing.T) {
	c := sample()
	m, e := Select(c, []int{1, 0})
	if e != nil {
		t.Fatal(e)
	}
	n, e := Select(c, []int{0, 1})
	if e != nil || n.ID != m.ID {
		t.Fatal("unstable selection", e)
	}
	if len(m.Members) != 2 || m.Members[0].Hash != m.Members[1].Hash || m.Members[0].Index == m.Members[1].Index {
		t.Fatal("collapsed equal occurrences")
	}
	c.CaptureReceipt.Question.Explore.Environment = "foreign"
	if m.Capture.Question.Explore.Environment != "prod" {
		t.Fatal("aliased original question")
	}
	c.Explore.Rows[0][0] = ' '
	if string(m.Members[0].Row)[0] != '{' {
		t.Fatal("aliased row")
	}
	for _, indexes := range [][]int{nil, {0, 0}, {-1}, {2}} {
		if _, e = Select(sample(), indexes); e == nil {
			t.Fatal("invalid indexes", indexes)
		}
	}
}
func TestXTDBRetainedEvidenceIntegration(t *testing.T) {
	url := os.Getenv("GOO_XTDB_TEST_URL")
	if url == "" {
		t.Skip("optional XTDB instance required")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()
	s := Store{URL: url}
	capture := sample()
	capture.ID = hash([]byte(t.Name() + time.Now().String()))
	m, e := Select(capture, []int{0, 1})
	if e != nil {
		t.Fatal(e)
	}
	requestID := hash([]byte(t.Name() + time.Now().String()))
	r, e := s.Publish(ctx, m, []string{"review:latency"}, "Initial review", requestID)
	if e != nil {
		t.Fatal(e)
	}
	retry, e := s.Publish(ctx, m, []string{"review:latency"}, "Initial review", requestID)
	if e != nil || retry.Annotation.ID != r.Annotation.ID {
		t.Fatal("retry changed receipt", e)
	}
	if _, e = s.Publish(ctx, m, []string{"review:latency"}, "Different intent", requestID); e == nil {
		t.Fatal("request key reused for different intent")
	}
	c, e := s.connect(ctx)
	if e != nil {
		t.Fatal(e)
	}
	defer c.Close(ctx)
	var cutoff time.Time
	if e = c.QueryRow(ctx, "SELECT _system_from FROM goo_annotations WHERE _id=$1", r.Annotation.ID).Scan(&cutoff); e != nil {
		t.Fatal(e)
	}
	got, e := s.Load(ctx, m.ID)
	if e != nil || !reflect.DeepEqual(*got, m) {
		t.Fatal(got, e)
	}
	if r.AwaitToken == "" {
		t.Fatal("missing committed basis token")
	}
	in := r.Annotation.MeaningInput
	in.Parent = r.Annotation.ID
	in.Note = "Corrected review"
	left, e := s.Correct(ctx, m.ID, in)
	if e != nil {
		t.Fatal(e)
	}
	in.Note = "Independent review"
	if _, e = s.Correct(ctx, m.ID, in); e != nil {
		t.Fatal(e)
	}
	p, e := s.Annotations(ctx, m.ID)
	if e != nil {
		t.Fatal(e)
	}
	heads := 0
	found := false
	for _, item := range p.Items {
		if item.Head && item.Revision.State == "active" {
			heads++
		}
		found = found || item.Revision.ID == r.Annotation.ID && item.Revision.Note == "Initial review"
	}
	if heads != 2 || !found || len(p.Items) != 3 {
		t.Fatal("history or branches lost", p)
	}
	var historic int64
	if e = c.QueryRow(ctx, "SELECT count(*) FROM goo_annotations FOR SYSTEM_TIME AS OF $1 WHERE target=$2", cutoff, m.ID).Scan(&historic); e != nil || historic != 1 {
		t.Fatal("native historical cutoff", historic, e)
	}
	in.Parent = left.ID
	in.Tags = []string{}
	in.Note = ""
	in.State = "withdrawn"
	if _, e = s.Correct(ctx, m.ID, in); e != nil {
		t.Fatal(e)
	}
	in.Target = workbench.MeaningTarget{Kind: "evidence", ID: m.ID}
	in.Parent = "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
	if _, e = s.Correct(ctx, m.ID, in); e == nil {
		t.Fatal("nonexistent parent")
	}
}

func TestXTDBInterruptedPublicationAndRetry(t *testing.T) {
	url := os.Getenv("GOO_XTDB_TEST_URL")
	if url == "" {
		t.Skip("optional XTDB instance required")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	capture := sample()
	capture.ID = hash([]byte(t.Name() + time.Now().String()))
	m, e := Select(capture, []int{0})
	if e != nil {
		t.Fatal(e)
	}
	key := hash([]byte(m.ID + "request"))
	broken := Store{URL: url, beforeCommit: func() error { return errors.New("simulated interruption after staging") }}
	if _, e = broken.Publish(ctx, m, []string{"review:latency"}, "Draft", key); e == nil {
		t.Fatal("interruption ignored")
	}
	s := Store{URL: url}
	if _, e = s.Load(ctx, m.ID); e == nil {
		t.Fatal("partially published manifest")
	}
	r, e := s.Publish(ctx, m, []string{"review:latency"}, "Draft", key)
	if e != nil {
		t.Fatal(e)
	}
	p, e := s.Annotations(ctx, m.ID)
	if e != nil || len(p.Items) != 1 || p.Items[0].Revision.ID != r.Annotation.ID {
		t.Fatal("partial assertion or duplicate", p, e)
	}
	// An indexed projection is queryable without decoding the manifest blob.
	c, e := s.connect(ctx)
	if e != nil {
		t.Fatal(e)
	}
	defer c.Close(ctx)
	var n int64
	if e = c.QueryRow(ctx, "SELECT count(*) FROM goo_evidence_members WHERE manifest_id=$1", m.ID).Scan(&n); e != nil || n != 1 {
		t.Fatal(n, e)
	}
}

func TestXTDBConcurrentRequestRecovery(t *testing.T) {
	url := os.Getenv("GOO_XTDB_TEST_URL")
	if url == "" {
		t.Skip("optional XTDB instance required")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	capture := sample()
	capture.ID = hash([]byte(t.Name() + time.Now().String()))
	m, e := Select(capture, []int{0})
	if e != nil {
		t.Fatal(e)
	}
	s := Store{URL: url}
	key := hash([]byte(m.ID + "request"))
	var wg sync.WaitGroup
	results := make([]*Receipt, 2)
	errs := make([]error, 2)
	for i := range results {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			results[i], errs[i] = s.Publish(ctx, m, []string{"review:latency"}, "Review", key)
		}(i)
	}
	wg.Wait()
	for _, e := range errs {
		if e != nil {
			t.Fatal(e)
		}
	}
	if results[0].Annotation.ID != results[1].Annotation.ID {
		t.Fatal("concurrent request produced two annotations")
	}
	p, e := s.Annotations(ctx, m.ID)
	if e != nil || len(p.Items) != 1 {
		t.Fatal(p, e)
	}
}

func TestCanonicalManifestSurvivesDisplayDecode(t *testing.T) {
	m, e := Select(sample(), []int{0})
	if e != nil {
		t.Fatal(e)
	}
	raw, _ := json.Marshal(m)
	envelope, _ := json.Marshal(Receipt{Manifest: m, ManifestJSON: string(raw)})
	var display map[string]any
	if json.Unmarshal(envelope, &display) != nil {
		t.Fatal("display decode")
	}
	rewritten, _ := json.Marshal(display["manifest"])
	var lossy Manifest
	json.Unmarshal(rewritten, &lossy)
	if lossy.Validate() == nil {
		t.Fatal("negative control did not expose display reserialization")
	}
	var exact Manifest
	if json.Unmarshal([]byte(display["manifestJSON"].(string)), &exact) != nil || exact.Validate() != nil || !reflect.DeepEqual(exact, m) {
		t.Fatal("canonical artifact changed across display boundary")
	}
}

func TestResultPageManifestProvenance(t *testing.T) {
	c := sample()
	c.Question.Explore.Offset = 100
	m, e := SelectResult(c, []int{1})
	if e != nil || m.Validate() != nil {
		t.Fatal(e, m.Validate())
	}
	if m.Origin != "result-page" || m.Capture.Question.Explore.Offset != 100 || m.Members[0].Index != 1 {
		t.Fatal("lost exact page provenance")
	}
	saved, _ := Select(c, []int{1})
	if saved.ID == m.ID {
		t.Fatal("live and saved origins conflated")
	}
	m.Origin = ""
	if m.Validate() == nil {
		t.Fatal("altered provenance accepted")
	}
}
func TestXTDBCorrectionRetryAndReadBasis(t *testing.T) {
	url := os.Getenv("GOO_XTDB_TEST_URL")
	if url == "" {
		t.Skip("optional XTDB instance required")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()
	s := Store{URL: url}
	capture := sample()
	capture.ID = hash([]byte(t.Name() + time.Now().String()))
	m, _ := SelectResult(capture, []int{0})
	key := hash([]byte(t.Name() + time.Now().String()))
	initial, e := s.Publish(ctx, m, []string{"review"}, "first", key)
	if e != nil {
		t.Fatal(e)
	}
	reader := Store{URL: url, AwaitToken: initial.AwaitToken}
	if _, e = reader.Load(ctx, m.ID); e != nil {
		t.Fatal("new connection did not observe commit", e)
	}
	in := workbench.MeaningInput{Target: workbench.MeaningTarget{Kind: "evidence", ID: m.ID}, Parent: initial.Annotation.ID, State: "active", Tags: []string{"review"}, Note: "corrected"}
	correctionKey := hash([]byte(key + "correction"))
	var wg sync.WaitGroup
	results := make(chan *CorrectionReceipt, 2)
	errs := make(chan error, 2)
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			r, e := reader.CorrectOnce(ctx, m.ID, in, correctionKey)
			results <- r
			errs <- e
		}()
	}
	wg.Wait()
	close(results)
	close(errs)
	for e := range errs {
		if e != nil {
			t.Fatal(e)
		}
	}
	var id, token string
	for r := range results {
		if id != "" && r.ID != id {
			t.Fatal("concurrent retry duplicated revision")
		}
		id = r.ID
		token = r.AwaitToken
	}
	next := Store{URL: url, AwaitToken: token}
	page, e := next.Annotations(ctx, m.ID)
	if e != nil || len(page.Items) != 2 {
		t.Fatal("read basis or retry failed", e, page)
	}
	in.Note = "changed"
	if _, e = next.CorrectOnce(ctx, m.ID, in, correctionKey); e == nil {
		t.Fatal("changed intent reused key")
	}
	in.Note = "interrupted"
	interrupted := Store{URL: url, AwaitToken: token, beforeCommit: func() error { return errors.New("interrupted") }}
	retryKey := hash([]byte(key + "interrupted"))
	if _, e = interrupted.CorrectOnce(ctx, m.ID, in, retryKey); e == nil {
		t.Fatal("fault ignored")
	}
	page, e = next.Annotations(ctx, m.ID)
	if e != nil || len(page.Items) != 2 {
		t.Fatal("partial correction committed", e)
	}
	if _, e = next.CorrectOnce(ctx, m.ID, in, retryKey); e != nil {
		t.Fatal("retry after rollback failed", e)
	}
	invalid := Store{URL: url, AwaitToken: strings.Repeat("x", 16385)}
	if _, e = invalid.Load(ctx, m.ID); e == nil {
		t.Fatal("oversized basis accepted")
	}
}
func TestXTDBRecoverExpiredPageSelection(t *testing.T) {
	url := os.Getenv("GOO_XTDB_TEST_URL")
	if url == "" {
		t.Skip("optional XTDB instance required")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	s := Store{URL: url}
	m, _ := SelectResult(sample(), []int{1, 0})
	key := hash([]byte(t.Name() + time.Now().String()))
	r, e := s.Publish(ctx, m, []string{"review"}, "exact", key)
	if e != nil {
		t.Fatal(e)
	}
	again, e := s.RecoverSelection(ctx, key, m.Capture.ID, true, []int{0, 1}, []string{"review"}, "exact")
	if e != nil || again.Annotation.ID != r.Annotation.ID {
		t.Fatal("cannot recover expired result", e)
	}
	for _, test := range []struct {
		source  string
		live    bool
		indexes []int
		note    string
	}{{m.Capture.ID, false, []int{0, 1}, "exact"}, {m.Capture.ID, true, []int{0, 0}, "exact"}, {m.Capture.ID, true, []int{0, 1}, "changed"}, {strings.Repeat("b", 64), true, []int{0, 1}, "exact"}} {
		if _, e = s.RecoverSelection(ctx, key, test.source, test.live, test.indexes, []string{"review"}, test.note); e == nil {
			t.Fatal("changed selection recovered", test)
		}
	}
}
