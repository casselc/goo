package retention

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/casselc/goo/workbench"
	"os"
	"testing"
	"time"
)

func catalogFixture(i int) (Manifest, workbench.MeaningRevision) {
	c := sample()
	c.ID = hash([]byte(fmt.Sprint(i)))
	c.Explore.Rows = []json.RawMessage{json.RawMessage(`{"ServiceName":"api","timestamp_ns":"1792629000000000001"}`)}
	m, e := SelectResult(c, []int{0})
	if e != nil {
		panic(e)
	}
	r, e := workbench.PrepareMeaning(workbench.MeaningInput{Target: workbench.MeaningTarget{Kind: "evidence", ID: m.ID}, State: "active", Tags: []string{"review:old"}})
	if e != nil {
		panic(e)
	}
	return m, *r
}
func TestCatalogFiltersBranchesTimeAndCursor(t *testing.T) {
	m, root := catalogFixture(1)
	left, _ := workbench.PrepareMeaning(workbench.MeaningInput{Target: root.Target, Parent: root.ID, State: "active", Tags: []string{"review:new"}})
	right, _ := workbench.PrepareMeaning(workbench.MeaningInput{Target: root.Target, Parent: root.ID, State: "active", Tags: []string{"review:parallel"}})
	withdraw, _ := workbench.PrepareMeaning(workbench.MeaningInput{Target: root.Target, Parent: left.ID, State: "withdrawn"})
	all := []workbench.MeaningRevision{root, *left, *right, *withdraw}
	parents := map[string]bool{root.ID: true, left.ID: true}
	for _, test := range []struct {
		query CatalogQuery
		n     int
	}{{CatalogQuery{Tag: "review:old"}, 0}, {CatalogQuery{Tag: "review:new"}, 0}, {CatalogQuery{Tag: "review:parallel", Service: "api", Environment: "prod", From: m.Capture.CapturedAt}, 1}, {CatalogQuery{Service: "ap"}, 0}, {CatalogQuery{Environment: "other"}, 0}, {CatalogQuery{To: m.Capture.CapturedAt}, 0}, {CatalogQuery{To: "0001-01-01T00:00:00Z"}, 0}} {
		p, e := catalogPage(test.query, []Manifest{m}, all, parents)
		if e != nil || len(p.Items) != test.n {
			t.Fatal(test, e, p)
		}
	}
	manifests := []Manifest{}
	revisions := []workbench.MeaningRevision{}
	for i := 0; i < 25; i++ {
		m, r := catalogFixture(i)
		manifests = append(manifests, m)
		revisions = append(revisions, r)
	}
	p, e := catalogPage(CatalogQuery{}, manifests, revisions, nil)
	if e != nil || len(p.Items) != 20 || p.NextCursor == "" {
		t.Fatal(e, p)
	}
	next, e := catalogPage(CatalogQuery{Cursor: p.NextCursor}, manifests, revisions, nil)
	if e != nil || len(next.Items) != 5 || next.Items[0].ID <= p.Items[19].ID {
		t.Fatal(e, next)
	}
	if _, e = catalogPage(CatalogQuery{Cursor: p.NextCursor, Service: "api"}, manifests, revisions, nil); e == nil {
		t.Fatal("cross-filter cursor accepted")
	}
	revisions[0].Tags = []string{"changed"}
	if _, e = catalogPage(CatalogQuery{Cursor: p.NextCursor}, manifests, revisions, nil); e == nil {
		t.Fatal("changed review catalog reused cursor")
	}
	for _, q := range []CatalogQuery{{From: "0001-01-01T00:00:00Z", To: "0001-01-01T00:00:00Z"}, {Tag: "UPPER"}, {From: "not a date"}, {From: "2026-10-22T01:00:00Z", To: "2026-10-22T00:00:00Z"}} {
		if q.Validate() == nil {
			t.Fatal("invalid filter accepted", q)
		}
	}
}
func TestXTDBCatalogAndPinnedDocument(t *testing.T) {
	url := os.Getenv("GOO_XTDB_TEST_URL")
	if url == "" {
		t.Skip("optional native instance required")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 40*time.Second)
	defer cancel()
	s := Store{URL: url}
	c := sample()
	c.ID = hash([]byte(t.Name() + time.Now().String()))
	c.Question.Explore.Environment = "catalog-" + c.ID[:12]
	c.Explore.Rows = []json.RawMessage{json.RawMessage(`{"ServiceName":"api","timestamp_ns":"1792629000000000001"}`)}
	m, _ := SelectResult(c, []int{0})
	receipt, e := s.Publish(ctx, m, []string{"review:catalog"}, "Original review", hash([]byte(c.ID+"request")))
	if e != nil {
		t.Fatal(e)
	}
	s.AwaitToken = receipt.AwaitToken
	p, e := s.Catalog(ctx, CatalogQuery{Environment: c.Question.Explore.Environment, Service: "api", Tag: "review:catalog"})
	if e != nil || len(p.Items) != 1 || p.Items[0].ID != m.ID {
		t.Fatal(e, p)
	}
	in := workbench.MeaningInput{Target: receipt.Annotation.Target, Parent: receipt.Annotation.ID, State: "active", Tags: []string{"review:later"}, Note: "Later correction"}
	r, e := s.CorrectOnce(ctx, m.ID, in, hash([]byte(c.ID+"correct")))
	if e != nil {
		t.Fatal(e)
	}
	s.AwaitToken = r.AwaitToken
	p, e = s.Catalog(ctx, CatalogQuery{Environment: c.Question.Explore.Environment, Tag: "review:catalog"})
	if e != nil || len(p.Items) != 0 {
		t.Fatal("superseded tag leaked", e, p)
	}
	old, e := s.Annotation(ctx, m.ID, receipt.Annotation.ID)
	if e != nil || old.Note != "Original review" {
		t.Fatal("pinned revision changed", e)
	}
	files := workbench.Store{Dir: t.TempDir()}
	card := workbench.Card{Kind: "retained", Reference: m.ID, Review: old.ID, ReadBasis: r.AwaitToken}
	check := func(card workbench.Card) error { _, e := s.Annotation(ctx, card.Reference, card.Review); return e }
	d, e := files.CreateDocument(workbench.DocumentInput{Title: "Pinned investigation", Cards: []workbench.Card{card}}, "", nil, check)
	if e != nil {
		t.Fatal(e)
	}
	loaded, e := files.LoadDocument(d.ID)
	if e != nil || loaded.Cards[0].Review != old.ID {
		t.Fatal("document lost review identity", e)
	}
	if _, e = files.CreateDocument(d.DocumentInput, "", nil); e == nil {
		t.Fatal("unresolved retained citation accepted")
	}
	card.Review = hash([]byte("other"))
	if _, e = files.CreateDocument(workbench.DocumentInput{Title: "Bad citation", Cards: []workbench.Card{card}}, "", nil, check); e == nil {
		t.Fatal("nonexistent pinned review accepted")
	}
	if _, e = files.CreateDocument(d.DocumentInput, "", nil, func(workbench.Card) error { return errors.New("native unavailable") }); e == nil {
		t.Fatal("unavailable resolver accepted")
	}
	if _, e = files.LoadDocument(d.ID); e != nil {
		t.Fatal("offline narrative read requires native DB", e)
	}
}

// Frozen from the pre-origin-aware implementation at b8f6ddc, before extracting
// shared pagination. Review creation clocks are absent from this catalog cursor.
func TestLegacyCatalogCursorBytes(t *testing.T) {
	manifests := []Manifest{}
	reviews := []workbench.MeaningRevision{}
	for i := 0; i < 25; i++ {
		m, r := catalogFixture(i)
		manifests = append(manifests, m)
		reviews = append(reviews, r)
	}
	p, e := catalogPage(CatalogQuery{}, manifests, reviews, nil)
	const expected = "eyJBZnRlciI6ImMxNzA5MzVhYThmMjFkYWJkN2MyM2Q1YTY2MzYxYjM2ZTMxNjA1YjU2NmM0MTI3OTQxNDhlYzQ5ODU0NTZjMmYiLCJGaW5nZXJwcmludCI6IjI4MzQ0NjVlNzc1M2JlZDJmN2JhNTNiZDY4YTFmNjVjOGVmNDFlNzc1M2NiNDI0N2M3NTZjNzFmMWE5YzAyMDIifQ"
	if e != nil || p.NextCursor != expected {
		t.Fatal("legacy cursor bytes changed", e, p)
	}
}
