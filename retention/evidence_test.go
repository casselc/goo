package retention

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/casselc/goo/workbench"
)

func TestEvidenceOriginViewAndReviewBranches(t *testing.T) {
	m, root := catalogFixture(100)
	view := Evidence{Origin: "result-page", Capture: &m}
	if view.Validate() != nil {
		t.Fatal("valid capture view")
	}
	for _, bad := range []Evidence{{Origin: "tail-selection", Capture: &m}, {Origin: "result-page"}, {Origin: "saved-capture", Capture: &m}, {Origin: "unknown", Capture: &m}} {
		if bad.Validate() == nil {
			t.Fatal("invalid origin view")
		}
	}
	left, _ := workbench.PrepareMeaning(workbench.MeaningInput{Target: root.Target, Parent: root.ID, State: "active", Tags: []string{"left"}})
	right, _ := workbench.PrepareMeaning(workbench.MeaningInput{Target: root.Target, Parent: root.ID, State: "active", Tags: []string{"right"}})
	withdrawal, _ := workbench.PrepareMeaning(workbench.MeaningInput{Target: root.Target, Parent: left.ID, State: "withdrawn"})
	page, e := reviewPage(m.ID, []workbench.MeaningRevision{*right, root, *left, root, *withdrawal})
	if e != nil || len(page.Items) != 4 {
		t.Fatal("merged roots/branches", e)
	}
	heads := 0
	active := 0
	for _, item := range page.Items {
		if item.Head {
			heads++
			if item.Revision.State == "active" {
				active++
			}
		}
	}
	if heads != 2 || active != 1 {
		t.Fatal("branch withdrawal rewrote another head")
	}
	if _, e = reviewPage(m.ID, []workbench.MeaningRevision{*left}); e == nil {
		t.Fatal("missing parent hidden")
	}
	foreign := root
	foreign.Target.ID = hash([]byte("foreign"))
	if _, e = reviewPage(m.ID, []workbench.MeaningRevision{foreign}); e == nil {
		t.Fatal("cross-target review admitted")
	}
}
func TestXTDBTailSharedReaders(t *testing.T) {
	url := os.Getenv("GOO_XTDB_TEST_URL")
	if url == "" {
		t.Skip("optional XTDB instance required")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	bank, now, in := tailFixture(t)
	in.Tags = []string{"review:origin-" + in.RequestID[:12]}
	s := Store{URL: url}
	receipt, e := s.PublishTail(ctx, bank, in)
	if e != nil {
		t.Fatal(e)
	}
	s.AwaitToken = receipt.AwaitToken
	original, _ := json.Marshal(receipt.Manifest)
	// Simulate the previous publisher's row shape: same immutable request/receipt,
	// with no manifest projection. Readers must recover it without a migration.
	c, e := s.connect(ctx)
	if e != nil {
		t.Fatal(e)
	}
	var before string
	if e = c.QueryRow(ctx, "SELECT payload FROM goo_tail_requests WHERE _id=$1", in.RequestID).Scan(&before); e != nil {
		t.Fatal(e)
	}
	if _, e = c.Exec(ctx, "UPDATE goo_tail_requests SET manifest_id=NULL WHERE _id=$1", in.RequestID); e != nil {
		t.Fatal(e)
	}
	c.Close(ctx)
	*now = now.Add(3 * time.Minute)
	read, e := s.ReadEvidence(ctx, receipt.Manifest.ID)
	if e != nil || read.Origin != "tail-selection" || read.Capture != nil || !reflect.DeepEqual(*read.Tail, receipt.Manifest) {
		t.Fatal("tail origin read", e)
	}
	page, e := s.Annotations(ctx, read.ID())
	if e != nil || len(page.Items) != 1 || page.Items[0].Revision.ID != receipt.Annotation.ID || !page.Items[0].Head {
		t.Fatal("legacy receipt root", e, page)
	}
	catalog, e := s.CatalogEvidence(ctx, EvidenceCatalogQuery{CatalogQuery: CatalogQuery{Tag: in.Tags[0]}})
	if e != nil || len(catalog.Items) != 1 || catalog.Items[0].ID != read.ID() || catalog.Items[0].Clock != "retained-at" || catalog.Items[0].At != receipt.RetainedAt.Format(time.RFC3339Nano) || catalog.Items[0].Observations != 2 {
		t.Fatal("tail catalog", e, catalog)
	}
	for _, q := range []EvidenceCatalogQuery{
		{CatalogQuery: CatalogQuery{Tag: in.Tags[0]}, Origin: "result-page"},
		{CatalogQuery: CatalogQuery{Tag: in.Tags[0], To: receipt.RetainedAt.Format(time.RFC3339Nano)}},
		{CatalogQuery: CatalogQuery{Tag: in.Tags[0], Environment: "other"}},
	} {
		p, e := s.CatalogEvidence(ctx, q)
		if e != nil || len(p.Items) != 0 {
			t.Fatal("catalog filter", e, p)
		}
	}
	correction := workbench.MeaningInput{Target: receipt.Annotation.Target, Parent: receipt.Annotation.ID, State: "active", Tags: []string{"review:later-" + in.RequestID[:12]}, Note: "Later interpretation"}
	key := hash([]byte(in.RequestID + "correct"))
	fixed, e := s.CorrectOnce(ctx, read.ID(), correction, key)
	if e != nil {
		t.Fatal(e)
	}
	s.AwaitToken = fixed.AwaitToken
	retry, e := s.CorrectOnce(ctx, read.ID(), correction, key)
	if e != nil || retry.ID != fixed.ID || retry.RecordedAt != fixed.RecordedAt {
		t.Fatal("correction retry", e)
	}
	changed := correction
	changed.Note = "different"
	if _, e = s.CorrectOnce(ctx, read.ID(), changed, key); e == nil {
		t.Fatal("changed correction intent")
	}
	old, e := s.Annotation(ctx, read.ID(), receipt.Annotation.ID)
	if e != nil || !reflect.DeepEqual(old, receipt.Annotation) {
		t.Fatal("pinned review changed", e)
	}
	page, e = s.Annotations(ctx, read.ID())
	if e != nil || len(page.Items) != 2 {
		t.Fatal("combined history", e)
	}
	catalog, e = s.CatalogEvidence(ctx, EvidenceCatalogQuery{CatalogQuery: CatalogQuery{Tag: in.Tags[0]}})
	if e != nil || len(catalog.Items) != 0 {
		t.Fatal("superseded root tag leaked", e)
	}
	catalog, e = s.CatalogEvidence(ctx, EvidenceCatalogQuery{CatalogQuery: CatalogQuery{Tag: correction.Tags[0]}, Origin: "tail-selection"})
	if e != nil || len(catalog.Items) != 1 {
		t.Fatal("corrected tail absent", e)
	}
	// Shared exact review resolution now supports file documents. File reads remain
	// independent of XTDB; no browser citation rendering is claimed by this check.
	files := workbench.Store{Dir: t.TempDir()}
	card := workbench.Card{Kind: "retained", Reference: read.ID(), Review: old.ID, ReadBasis: fixed.AwaitToken}
	doc, e := files.CreateDocument(workbench.DocumentInput{Title: "Observed tail evidence", Cards: []workbench.Card{card}}, "", nil, func(card workbench.Card) error { _, e := s.Annotation(ctx, card.Reference, card.Review); return e })
	if e != nil {
		t.Fatal(e)
	}
	if loaded, e := files.LoadDocument(doc.ID); e != nil || !reflect.DeepEqual(loaded.Cards[0], card) {
		t.Fatal("offline document identity", e)
	}
	if _, e = s.Annotation(ctx, read.ID(), hash([]byte("foreign-review"))); e == nil {
		t.Fatal("unknown exact review accepted")
	}
	foreign := correction
	foreign.Parent = hash([]byte("foreign-parent"))
	if _, e = s.CorrectOnce(ctx, read.ID(), foreign, hash([]byte(key+"bad"))); e == nil {
		t.Fatal("foreign parent accepted")
	}
	// New meaning never mutates the original manifest or publication receipt.
	restored, e := s.LoadTail(ctx, read.ID())
	if e != nil {
		t.Fatal(e)
	}
	raw, _ := json.Marshal(restored)
	if string(raw) != string(original) {
		t.Fatal("correction changed retained observations")
	}
	recovered, e := s.PublishTail(ctx, nil, in)
	if e != nil || recovered.Annotation.ID != old.ID || !recovered.RetainedAt.Equal(receipt.RetainedAt) {
		t.Fatal("publication retry followed correction head", e)
	}
	// Legacy origin remains readable and resolves through the same backend view.
	capture := sample()
	capture.ID = hash([]byte(in.RequestID + "capture"))
	manifest, e := Select(capture, []int{0})
	if e != nil {
		t.Fatal(e)
	}
	legacy, e := s.Publish(ctx, manifest, []string{"review:legacy"}, "", hash([]byte(in.RequestID+"legacy")))
	if e != nil {
		t.Fatal(e)
	}
	s.AwaitToken = legacy.AwaitToken
	v, e := s.ReadEvidence(ctx, manifest.ID)
	if e != nil || v.Tail != nil || v.Origin != "saved-capture" || !reflect.DeepEqual(*v.Capture, manifest) {
		t.Fatal("legacy view changed", e)
	}
	if _, e = s.Annotation(ctx, read.ID(), legacy.Annotation.ID); e == nil {
		t.Fatal("cross-evidence review resolved")
	}
	// The HTTP-owned legacy catalog deliberately does not expose the new origin.
	legacyPage, e := s.Catalog(ctx, CatalogQuery{Tag: correction.Tags[0]})
	if e != nil || len(legacyPage.Items) != 0 {
		t.Fatal("tail leaked into legacy catalog", e)
	}
	c, e = s.connect(ctx)
	if e != nil {
		t.Fatal(e)
	}
	defer c.Close(ctx)
	var after string
	if e = c.QueryRow(ctx, "SELECT payload FROM goo_tail_requests WHERE _id=$1", in.RequestID).Scan(&after); e != nil || after != before {
		t.Fatal("reader rewrote old receipt", e)
	}
	if _, e = s.ReadEvidence(ctx, hash([]byte(in.RequestID+"missing"))); e == nil {
		t.Fatal("missing origin accepted")
	}
	bad := Store{URL: "postgres://127.0.0.1:1/xtdb?connect_timeout=1"}
	if _, e = bad.Annotation(ctx, read.ID(), old.ID); e == nil || errors.Is(e, ErrRequestNotFound) {
		t.Fatal("outage hidden as absent request", e)
	}
}
