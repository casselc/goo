package exploration

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"
)

func searchFixture(t *testing.T) SavedCapture {
	t.Helper()
	b, e := os.ReadFile("../testdata/storage/exploration-capture.json")
	if e != nil {
		t.Fatal(e)
	}
	var c SavedCapture
	if e = json.Unmarshal(b, &c); e != nil {
		t.Fatal(e)
	}
	return c
}
func TestCaptureSearchWholeCatalogAndBoundCursors(t *testing.T) {
	s := CaptureStore{Dir: t.TempDir()}
	c := searchFixture(t)
	ids := []string{}
	for i := range 47 {
		c.Title = "Same title"
		if i == 46 {
			c.Title = "Needle [雪].*"
		}
		c.CapturedAt = time.Unix(int64(i), 0).UTC().Format(time.RFC3339Nano)
		v, e := s.publish(c)
		if e != nil {
			t.Fatal(e)
		}
		ids = append(ids, v.ID)
	}
	q := CaptureSearchQuery{}
	p, e := s.Search(t.Context(), q, "")
	if e != nil || len(p.Items) != 20 || !p.HasMore || p.Items[0].ID != ids[46] || p.Total != 47 {
		t.Fatal(p, e)
	}
	found := []string{}
	cursor := ""
	for {
		p, e = s.Search(t.Context(), q, cursor)
		if e != nil {
			t.Fatal(e)
		}
		for _, r := range p.Items {
			found = append(found, r.ID)
		}
		if !p.HasMore {
			break
		}
		cursor = p.NextCursor
	}
	slices.Reverse(ids)
	if !slices.Equal(found, ids) {
		t.Fatal("wrong global ordering", found, ids)
	}
	q.Title = "[雪].*"
	p, e = s.Search(t.Context(), q, "")
	if e != nil || p.Matched != 1 {
		t.Fatal(p, e)
	}
	q.Title = "same"
	q.Order = "oldest"
	p, e = s.Search(t.Context(), q, "")
	if e != nil || p.Matched != 46 || p.Items[0].ID != ids[46] {
		t.Fatal(p, e)
	}
	cursor = p.NextCursor
	q.Title = "Other"
	if _, e = s.Search(t.Context(), q, cursor); !errors.Is(e, ErrCaptureSearchCursor) {
		t.Fatal("foreign query", e)
	}
	q.Title = "same"
	c.Title = "addition"
	c.CapturedAt = "2026-10-29T00:00:00Z"
	if _, e = s.publish(c); e != nil {
		t.Fatal(e)
	}
	if _, e = s.Search(t.Context(), q, cursor); !errors.Is(e, ErrCaptureSearchCursor) {
		t.Fatal("changed catalog", e)
	}
	if _, e = s.Search(t.Context(), q, ids[0]); !errors.Is(e, ErrCaptureSearchCursor) {
		t.Fatal("legacy cursor accepted", e)
	}
}
func TestCaptureSearchIntegrityBudgetsAndCancellation(t *testing.T) {
	s := CaptureStore{Dir: t.TempDir()}
	c, e := s.publish(searchFixture(t))
	if e != nil {
		t.Fatal(e)
	}
	env := ""
	for _, q := range []CaptureSearchQuery{{Title: "no match"}, {Environment: &env}, {Signal: "logs"}} {
		p, e := s.Search(t.Context(), q, "")
		if e != nil || p.Matched != 0 {
			t.Fatal(p, e)
		}
	}
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	if _, e = s.Search(ctx, CaptureSearchQuery{}, ""); !errors.Is(e, context.Canceled) {
		t.Fatal(e)
	}
	file := filepath.Join(s.Dir, c.ID+".json")
	b, _ := os.ReadFile(file)
	if e = os.WriteFile(file, []byte("broken"), 0600); e != nil {
		t.Fatal(e)
	}
	if _, e = s.Search(t.Context(), CaptureSearchQuery{Title: "does not match"}, ""); e == nil {
		t.Fatal("excluded corruption concealed")
	}
	if _, e = s.Load(c.ID); e == nil {
		t.Fatal("reopen corruption concealed")
	}
	if e = os.WriteFile(file, b, 0600); e != nil {
		t.Fatal(e)
	}
	f, e := os.Create(filepath.Join(s.Dir, strings.Repeat("0", 64)+".json"))
	if e != nil {
		t.Fatal(e)
	}
	if e = f.Truncate(CaptureSearchMaxBytes + 1); e != nil {
		t.Fatal(e)
	}
	f.Close()
	if _, e = s.Search(t.Context(), CaptureSearchQuery{}, ""); !errors.Is(e, ErrCaptureCatalogUnavailable) {
		t.Fatal("missing metadata accepted", e)
	}
	metadataFile := filepath.Join(s.catalogDir(), strings.Repeat("0", 64)+".json")
	large, e := os.Create(metadataFile)
	if e != nil {
		t.Fatal(e)
	}
	if e = large.Truncate(CaptureSearchMaxBytes + 1); e != nil {
		t.Fatal(e)
	}
	large.Close()
	if _, e = s.Search(t.Context(), CaptureSearchQuery{}, ""); !errors.Is(e, ErrCaptureSearchBudget) {
		t.Fatal(e)
	}
	if _, e = s.Load(c.ID); e != nil {
		t.Fatal("search limit broke direct reopen", e)
	}
}
func TestCaptureSearchTieInstantsAndScope(t *testing.T) {
	s := CaptureStore{Dir: t.TempDir()}
	c := searchFixture(t)
	ids := []string{}
	for _, at := range []string{"2026-10-29T00:00:00Z", "2026-10-28T20:00:00-04:00"} {
		c.CapturedAt = at
		v, e := s.publish(c)
		if e != nil {
			t.Fatal(e)
		}
		ids = append(ids, v.ID)
	}
	slices.Sort(ids)
	env := c.Question.Explore.Environment
	p, e := s.Search(t.Context(), CaptureSearchQuery{Environment: &env, Signal: "spans", Kind: "explore"}, "")
	if e != nil || p.Matched != 2 || p.Items[0].ID != ids[0] {
		t.Fatal(p, e)
	}
	if _, _, known := captureScope(CaptureQuestion{Kind: "measurement"}); known {
		t.Fatal("invented measurement scope")
	}
	for _, q := range []CaptureSearchQuery{{Title: strings.Repeat("a", 121)}, {Title: "\xff"}, {Environment: ptrSearch("\n")}, {Kind: "wat"}, {Order: "mtime"}, {Signal: "metricz"}} {
		if q.Validate() == nil {
			t.Fatal("invalid query accepted", q)
		}
	}
}
func ptrSearch(s string) *string { return &s }

func TestCaptureCatalogRebuildAndExplicitIntegrityBoundary(t *testing.T) {
	s := CaptureStore{Dir: t.TempDir()}
	c, e := s.publish(searchFixture(t))
	if e != nil {
		t.Fatal(e)
	}
	file := filepath.Join(s.Dir, c.ID+".json")
	original, _ := os.ReadFile(file)
	stamp, _ := os.Stat(file)
	if e = os.RemoveAll(s.catalogDir()); e != nil {
		t.Fatal(e)
	}
	if _, e = s.Search(t.Context(), CaptureSearchQuery{}, ""); !errors.Is(e, ErrCaptureCatalogUnavailable) {
		t.Fatal(e)
	}
	rebuilt, e := s.RebuildCatalog(t.Context())
	if e != nil || rebuilt.Indexed != 1 || rebuilt.Bytes != int64(len(original)) {
		t.Fatal(rebuilt, e)
	}
	// A same-size mutation with restored mtime is intentionally not a payload check.
	damaged := append([]byte{}, original...)
	damaged[0] = '!'
	if e = os.WriteFile(file, damaged, 0600); e != nil {
		t.Fatal(e)
	}
	if e = os.Chtimes(file, stamp.ModTime(), stamp.ModTime()); e != nil {
		t.Fatal(e)
	}
	p, e := s.Search(t.Context(), CaptureSearchQuery{}, "")
	if e != nil || p.Matched != 1 {
		t.Fatal("metadata incorrectly promises payload verification", p, e)
	}
	if _, e = s.Load(c.ID); e == nil {
		t.Fatal("corrupt payload reopened")
	}
	if _, e = s.RebuildCatalog(t.Context()); e == nil {
		t.Fatal("rebuild accepted corrupt payload")
	}
	if e = os.WriteFile(file, original, 0600); e != nil {
		t.Fatal(e)
	}
	if _, e = s.RebuildCatalog(t.Context()); e != nil {
		t.Fatal(e)
	}
	meta := filepath.Join(s.catalogDir(), c.ID+".json")
	if e = os.WriteFile(meta, []byte("bad"), 0600); e != nil {
		t.Fatal(e)
	}
	if _, e = s.Search(t.Context(), CaptureSearchQuery{}, ""); e == nil {
		t.Fatal("bad metadata accepted")
	}
	if _, e = s.RebuildCatalog(t.Context()); e != nil {
		t.Fatal(e)
	}
	if _, e = s.Search(t.Context(), CaptureSearchQuery{}, ""); e != nil {
		t.Fatal(e)
	}
	now, _ := os.ReadFile(file)
	if string(now) != string(original) {
		t.Fatal("rebuild rewrote capture")
	}
}
func TestCommittedCaptureSurvivesMetadataFailure(t *testing.T) {
	s := CaptureStore{Dir: t.TempDir()}
	if e := os.WriteFile(s.catalogDir(), []byte("blocks derived directory"), 0600); e != nil {
		t.Fatal(e)
	}
	c, e := s.publish(searchFixture(t))
	if e != nil {
		t.Fatal("committed capture disguised as failed save", e)
	}
	if _, e = s.Load(c.ID); e != nil {
		t.Fatal(e)
	}
	if _, e = s.Search(t.Context(), CaptureSearchQuery{}, ""); !errors.Is(e, ErrCaptureCatalogUnavailable) {
		t.Fatal("missing catalog concealed", e)
	}
}
