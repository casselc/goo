package exploration

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
)

func captureFixture(t *testing.T) (Source, *atomic.Int64, func()) {
	t.Helper()
	total := &atomic.Int64{}
	total.Store(1200)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		sql := string(b)
		n := total.Load()
		switch {
		case strings.Contains(sql, "GROUP BY"):
			fmt.Fprintf(w, "{\"g0\":\"worker\",\"count\":\"%d\"}\n", n)
		case strings.Contains(sql, "SELECT DISTINCT"):
		case strings.Contains(sql, "toString(count()) AS count"):
			fmt.Fprintf(w, "{\"count\":\"%d\"}\n", n)
		default:
			if !strings.Contains(sql, "LIMIT 1001 OFFSET 0") {
				t.Error("capture must use one bounded row bank", sql)
			}
			if n > 1001 {
				n = 1001
			}
			for i := int64(0); i < n; i++ {
				fmt.Fprintf(w, "{\"timestamp_ns\":\"%d\",\"duration_ns\":\"0\",\"SpanId\":\"%016x\"}\n", i+1, i+1)
			}
		}
	}))
	return Source{Endpoint: server.URL}, total, server.Close
}
func savedQuestion() CaptureQuestion {
	q := basic()
	q.Offset = 100
	q.Limit = 2
	q.GroupBy = []Field{{"column", "service"}}
	return CaptureQuestion{Kind: "explore", Explore: &q}
}
func TestCaptureRetainsFullCountsAndStableBoundedPages(t *testing.T) {
	source, total, close := captureFixture(t)
	defer close()
	s := CaptureStore{Dir: t.TempDir()}
	c, e := s.Create(context.Background(), source, "Captured worker", savedQuestion(), "")
	if e != nil {
		t.Fatal(e)
	}
	if c.Explore.Count != "1200" || len(c.Explore.Rows) != 1000 || !c.Explore.HasMore || c.Question.Explore.Offset != 0 || c.Question.Explore.Limit != 100 {
		t.Fatal("page became population", c)
	}
	first, e := c.Page("")
	if e != nil || len(first.Explore.Rows) != 100 || !first.RecordsLimited || first.PreviousCursor != "" || first.NextCursor == "" {
		t.Fatal(first, e)
	}
	before, _ := json.Marshal(first)
	total.Store(1201)
	child, e := s.Create(context.Background(), source, c.Title, c.Question, c.ID)
	if e != nil || child.Parent != c.ID || child.Explore.Count != "1201" {
		t.Fatal(child, e)
	}
	if _, e := child.Page(first.NextCursor); e == nil {
		t.Fatal("cross-revision cursor accepted")
	}
	changed := c.Question
	v := *changed.Explore
	v.Environment = "other"
	changed.Explore = &v
	if _, e = s.Create(context.Background(), source, c.Title, changed, c.ID); e == nil {
		t.Fatal("changed refresh question accepted")
	}

	altered := *c
	changedEvidence := *c.Explore
	changedEvidence.Plan.Count += " /* older adapter */"
	altered.Explore = &changedEvidence
	previous, e := s.publish(altered)
	if e != nil {
		t.Fatal(e)
	}
	if _, e = s.Create(context.Background(), source, c.Title, c.Question, previous.ID); !errors.Is(e, ErrCapturePlanChanged) {
		t.Fatal("refresh silently changed compiler plan", e)
	}
	close()
	fresh := CaptureStore{Dir: s.Dir}
	original, e := fresh.Load(c.ID)
	if e != nil {
		t.Fatal(e)
	}
	again, _ := original.Page("")
	after, _ := json.Marshal(again)
	if string(before) != string(after) {
		t.Fatal("predecessor changed")
	}
	all := []string{}
	cursor := ""
	for {
		p, e := original.Page(cursor)
		if e != nil {
			t.Fatal(e)
		}
		if p.Explore.Count != "1200" || p.Explore.Basis != first.Explore.Basis || !reflect.DeepEqual(p.Explore.Groups, first.Explore.Groups) {
			t.Fatal("page changed captured aggregates")
		}
		for _, row := range p.Explore.Rows {
			var v struct {
				ID string `json:"SpanId"`
			}
			json.Unmarshal(row, &v)
			all = append(all, v.ID)
		}
		if p.NextCursor == "" {
			if p.Explore.Question.Offset != 900 {
				t.Fatal("unexpected last page")
			}
			break
		}
		cursor = p.NextCursor
	}
	if len(all) != 1000 {
		t.Fatal(len(all))
	}
	for i, v := range all {
		if v != fmt.Sprintf("%016x", i+1) {
			t.Fatal("skip or duplicate", i, v)
		}
	}
	for _, cur := range []string{"?", captureCursor(c.ID, 1), captureCursor(c.ID, 1000), captureCursor(child.ID, 100), strings.Repeat("x", 129)} {
		if _, e := original.Page(cur); e == nil {
			t.Fatal("invalid cursor accepted", cur)
		}
	}
}
func TestCaptureIntegrityCatalogAndNoData(t *testing.T) {
	source, total, close := captureFixture(t)
	defer close()
	total.Store(0)
	s := CaptureStore{Dir: t.TempDir()}
	c, e := s.Create(context.Background(), source, "Empty", savedQuestion(), "")
	if e != nil {
		t.Fatal(e)
	}
	p, e := c.Page("")
	if e != nil || p.Explore.State != "no_data" || p.RetainedRecords != 0 || p.RecordsLimited || p.NextCursor != "" {
		t.Fatal(p, e)
	}
	if _, e := c.Page(captureCursor(c.ID, 100)); e == nil {
		t.Fatal("nonexistent page accepted")
	}
	for i := 0; i < CaptureListLimit+1; i++ {
		copy := *c
		copy.Title = "Capture " + strconv.Itoa(i)
		if _, e = s.publish(copy); e != nil {
			t.Fatal(e)
		}
	}
	catalog, e := s.List("")
	if e != nil || len(catalog.Items) != CaptureListLimit || !catalog.HasMore {
		t.Fatal(catalog, e)
	}
	next, e := s.List(catalog.NextCursor)
	if e != nil || len(next.Items) != 2 || next.HasMore {
		t.Fatal(next, e)
	}
	for _, bad := range []string{"../escape", strings.Repeat("a", 63)} {
		if _, e = s.Load(bad); e == nil {
			t.Fatal("invalid ID")
		}
		if _, e = s.List(bad); e == nil {
			t.Fatal("invalid catalog cursor")
		}
	}
	badID := catalog.Items[0].ID
	path := filepath.Join(s.Dir, badID+".json")
	b, _ := os.ReadFile(path)
	var corrupt SavedCapture
	json.Unmarshal(b, &corrupt)
	corrupt.CapturedAt = "tampered"
	b, _ = json.Marshal(corrupt)
	os.WriteFile(path, b, 0600)
	if _, e = s.Load(badID); e == nil {
		t.Fatal("tampered receipt accepted")
	}
	if _, e = s.List(""); e == nil {
		t.Fatal("catalog hid corrupt saved evidence")
	}
}
func TestConcurrentCapturePublicationAndValidation(t *testing.T) {
	source, total, close := captureFixture(t)
	defer close()
	total.Store(1)
	s := CaptureStore{Dir: t.TempDir()}
	var wg sync.WaitGroup
	for range 8 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			c, e := s.Create(context.Background(), source, "Concurrent", savedQuestion(), "")
			if e != nil {
				t.Error(e)
				return
			}
			if _, e = s.Load(c.ID); e != nil {
				t.Error(e)
			}
		}()
	}
	wg.Wait()
	for _, title := range []string{"", strings.Repeat("a", 121), "bad\nname"} {
		if _, e := s.Create(context.Background(), source, title, savedQuestion(), ""); e == nil {
			t.Fatal("invalid title accepted")
		}
	}
	close()
	if _, e := s.Create(context.Background(), source, "Unavailable", savedQuestion(), ""); e == nil {
		t.Fatal("source outage became saved no-data")
	}
}

func TestCaptureRefreshCanonicalizesEmptyGroupingAndPresentation(t *testing.T) {
	source, total, close := captureFixture(t)
	defer close()
	total.Store(10)
	q := savedQuestion()
	q.Explore.GroupBy = []Field{}
	store := CaptureStore{Dir: t.TempDir()}
	first, e := store.Create(context.Background(), source, "No grouping", q, "")
	if e != nil {
		t.Fatal(e)
	}
	q.Explore.Offset = 500
	q.Explore.Limit = 1
	child, e := store.Create(context.Background(), source, "No grouping", q, first.ID)
	if e != nil || child.Parent != first.ID || child.Question.Explore.GroupBy != nil || child.Question.Explore.Offset != 0 {
		t.Fatal("equivalent presentation/grouping changed the question", e)
	}
}
