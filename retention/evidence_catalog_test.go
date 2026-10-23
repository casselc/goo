package retention

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/workbench"
)

func catalogTailFixture(t *testing.T, n int, service *string) (Evidence, TailReceipt) {
	t.Helper()
	now := time.UnixMilli(200000)
	row := map[string]any{"timestamp_ns": "150000000000"}
	if service != nil {
		row["ServiceName"] = *service
	}
	raw, _ := json.Marshal(row)
	tracker := exploration.TailTracker{}
	rows, e := tracker.Observe([]json.RawMessage{raw})
	if e != nil {
		t.Fatal(e)
	}
	bank := exploration.NewTailMembershipSnapshots(func() time.Time { return now })
	connection := hash([]byte(fmt.Sprint(n)))
	selector := exploration.TailSelector{Environment: "test"}
	cap, e := bank.Freeze(exploration.TailMembershipBatch{Connection: connection, Selector: selector, FromMS: 100000, ToMS: 160000, ObservedAt: now, Rows: rows})
	if e != nil {
		t.Fatal(e)
	}
	draft, e := bank.Resolve(connection, selector, []exploration.TailMembershipRef{{Token: cap.Token, Key: rows[0].Key}})
	if e != nil {
		t.Fatal(e)
	}
	m := TailManifest{Version: TailEvidenceVersion, Selection: *draft}
	raw, _ = json.Marshal(m)
	m.ID = hash(raw)
	review, e := workbench.PrepareMeaning(workbench.MeaningInput{Target: workbench.MeaningTarget{Kind: "evidence", ID: m.ID}, Tags: []string{"review:tail"}, State: "active"})
	if e != nil {
		t.Fatal(e)
	}
	return Evidence{Origin: "tail-selection", Tail: &m}, TailReceipt{Manifest: m, Annotation: review, RetainedAt: time.UnixMilli(300000)}
}
func TestEvidenceCatalogClocksFiltersAndPagination(t *testing.T) {
	evidence := []Evidence{}
	receipts := []TailReceipt{}
	reviews := []workbench.MeaningRevision{}
	empty, api := "", "api"
	for i := 0; i < 25; i++ {
		var service *string
		if i == 0 {
			service = &empty
		}
		if i == 1 {
			service = &api
		}
		v, r := catalogTailFixture(t, i, service)
		evidence = append(evidence, v)
		receipts = append(receipts, r)
		reviews = append(reviews, *r.Annotation)
	}
	legacy, review := catalogFixture(50)
	evidence = append(evidence, Evidence{Origin: "result-page", Capture: &legacy})
	reviews = append(reviews, review)
	query := EvidenceCatalogQuery{Origin: "tail-selection"}
	page, e := evidenceCatalogPage(query, evidence, receipts, reviews)
	if e != nil || len(page.Items) != 20 || page.NextCursor == "" {
		t.Fatal("first page", e)
	}
	query.Cursor = page.NextCursor
	next, e := evidenceCatalogPage(query, evidence, receipts, reviews)
	if e != nil || len(next.Items) != 5 || next.NextCursor != "" || next.Items[0].ID <= page.Items[19].ID {
		t.Fatal("next page", e)
	}
	for _, item := range append(page.Items, next.Items...) {
		if item.Clock != "retained-at" || item.At != receipts[0].RetainedAt.Format(time.RFC3339Nano) {
			t.Fatal("fabricated capture clock")
		}
	}
	changed := query
	changed.Origin = "result-page"
	if _, e = evidenceCatalogPage(changed, evidence, receipts, reviews); e == nil {
		t.Fatal("cross-origin cursor accepted")
	}
	changed = query
	changed.Tag = "review:other"
	if _, e = evidenceCatalogPage(changed, evidence, receipts, reviews); e == nil {
		t.Fatal("cross-filter cursor accepted")
	}
	changedReviews := append([]workbench.MeaningRevision{}, reviews...)
	correction, _ := workbench.PrepareMeaning(workbench.MeaningInput{Target: reviews[0].Target, Parent: reviews[0].ID, State: "active", Tags: []string{"review:corrected"}})
	changedReviews = append(changedReviews, *correction)
	if _, e = evidenceCatalogPage(query, evidence, receipts, changedReviews); e == nil {
		t.Fatal("changed head reused catalog cursor")
	}
	all, e := evidenceCatalogPage(EvidenceCatalogQuery{Origin: "result-page"}, evidence, receipts, reviews)
	if e != nil || len(all.Items) != 1 || all.Items[0].Clock != "captured-at" || all.Items[0].At != legacy.Capture.CapturedAt {
		t.Fatal("legacy clock changed", e)
	}
	servicePage, e := evidenceCatalogPage(EvidenceCatalogQuery{CatalogQuery: CatalogQuery{Service: "api"}, Origin: "tail-selection"}, evidence, receipts, reviews)
	if e != nil || len(servicePage.Items) != 1 || servicePage.Items[0].ID != evidence[1].ID() {
		t.Fatal("exact service filter", e)
	}
	missingEmpty, e := evidenceCatalogPage(EvidenceCatalogQuery{Origin: "tail-selection"}, evidence[:3], receipts[:3], reviews[:3])
	if e != nil {
		t.Fatal(e)
	}
	for _, item := range missingEmpty.Items {
		if item.ID == evidence[2].ID() && len(item.Services) != 0 {
			t.Fatal("missing service became empty")
		}
		if item.ID == evidence[0].ID() && (len(item.Services) != 1 || item.Services[0] != "") {
			t.Fatal("empty service became missing")
		}
	}
	if _, e = evidenceCatalogPage(EvidenceCatalogQuery{}, evidence, nil, reviews); e == nil {
		t.Fatal("missing custody silently returned partial catalog")
	}
	if _, e = evidenceCatalogPage(EvidenceCatalogQuery{}, append(evidence, evidence[0]), receipts, reviews); e == nil {
		t.Fatal("duplicate identity hidden")
	}
	if (EvidenceCatalogQuery{Origin: "mystery"}).Validate() == nil {
		t.Fatal("unknown origin")
	}
}
