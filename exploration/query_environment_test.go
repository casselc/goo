package exploration

import (
	"encoding/json"
	"fmt"
	"github.com/casselc/goo/identity"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func scopedQuestion250(mode, value string) Query {
	selection, _ := identity.EffectiveEnvironment(identity.EnvironmentScope{Mode: mode, Value: value})
	wire, _ := selection.Selector()
	return Query{Signal: "logs", EnvironmentSelector: &wire, FromMS: 100000, ToMS: 160000, Limit: 100}
}
func TestQuerySelectorValidationAndCopy250(t *testing.T) {
	for _, q := range []Query{scopedQuestion250("unreported", ""), scopedQuestion250("empty", ""), scopedQuestion250("named", "prod")} {
		if err := q.Validate(); err != nil {
			t.Fatal(err)
		}
		plan, err := Compile(q)
		if err != nil {
			t.Fatal(err)
		}
		if q.EnvironmentSelector.Scope.Mode != "named" {
			if _, ok := plan.Parameters["environment"]; ok {
				t.Fatal("dummy binding")
			}
		}
		derived, err := SurroundingQuestion(q, "130000000000", "api", 1000, 0)
		if err != nil || !derived.SameEnvironment(q) {
			t.Fatal("context dropped scope", err)
		}
		tail := TailSelector{EnvironmentSelector: q.EnvironmentSelector}
		fromTail := tail.Query(160000)
		normalized := (CaptureQuestion{Kind: "explore", Explore: &q}).Normalized()
		scope, err := (CaptureQuestion{Kind: "explore", Explore: &q}).Scope()
		if err != nil {
			t.Fatal(err)
		}
		original := *q.EnvironmentSelector
		q.EnvironmentSelector.Scope.Mode = "broken"
		for _, copy := range []*identity.EnvironmentSelector{derived.EnvironmentSelector, fromTail.EnvironmentSelector, normalized.Explore.EnvironmentSelector, scope.EnvironmentSelector} {
			if copy == nil || *copy != original {
				t.Fatal("selector aliased caller")
			}
		}
	}
	q := scopedQuestion250("named", "prod")
	legacy := q
	legacy.Environment = "prod"
	legacy.EnvironmentSelector = nil
	if q.SameEnvironment(legacy) {
		t.Fatal("effective and current-only conflated")
	}
	q.Environment = "prod"
	if q.Validate() == nil {
		t.Fatal("ambiguous selector accepted")
	}
	if (Query{Signal: "logs", FromMS: 1, ToMS: 1000, Limit: 1}).Validate() == nil {
		t.Fatal("missing scope wildcard")
	}
}
func TestSelectorCompatibilityAndRefusals250(t *testing.T) {
	a, b := scopedQuestion250("unreported", ""), scopedQuestion250("empty", "")
	field := Field{Kind: "column", Key: "service"}
	a.GroupBy = []Field{field}
	b.GroupBy = []Field{field}
	if (CaptureQuestion{Kind: "comparison", Comparison: &ComparisonQuestion{Baseline: a, Selection: b}}).Validate() == nil {
		t.Fatal("cross-scope comparison accepted")
	}
	if _, err := PlanExplanation(ExplanationRequest{Version: ExplanationVersion, Baseline: a, Selection: b, Fields: []Field{field}}); err == nil {
		t.Fatal("cross-scope explanation accepted")
	}
	b.Signal = "spans"
	if _, err := PlanPopulation(PopulationRequest{Question: a, Companion: &b}); err == nil {
		t.Fatal("cross-scope companion accepted")
	}
	a.Signal = "spans"
	a.ReportedEdge = &ReportedEdge{}
	if err := a.Validate(); err == nil || !strings.Contains(err.Error(), "environmentSelector") {
		t.Fatal("reported-edge bridge accepted selector", err)
	}
	a.Signal = "logs"
	a.ReportedEdge = nil
	if _, err := DraftTelemetryPackage(a, "sdk", "SDK", 1); err == nil {
		t.Fatal("package dropped scope")
	}
}
func TestSelectorTailMembershipAndSnapshots250(t *testing.T) {
	q := scopedQuestion250("named", "prod")
	row := json.RawMessage(`{"timestamp_ns":"130000000000","Body":"hello","ResourceAttributes":{"deployment.environment":"prod"}}`)
	now := time.UnixMilli(160000)
	bank := NewTailMembershipSnapshots(func() time.Time { return now })
	tracker := TailTracker{}
	rows, err := tracker.Observe([]json.RawMessage{row})
	if err != nil {
		t.Fatal(err)
	}
	selector := TailSelector{Signal: "logs", EnvironmentSelector: copyEnvironmentSelector(q.EnvironmentSelector)}
	batch := TailMembershipBatch{Connection: strings.Repeat("a", 64), Selector: selector, FromMS: 100000, ToMS: 160000, ObservedAt: now, Rows: rows}
	cap, err := bank.Freeze(batch)
	if err != nil {
		t.Fatal(err)
	}
	refs := []TailMembershipRef{{Token: cap.Token, Key: rows[0].Key}}
	if _, err = bank.Resolve(batch.Connection, TailSelector{Signal: "logs", Environment: "prod"}, refs); err == nil {
		t.Fatal("capability reused across policies")
	}
	draft, err := bank.Resolve(batch.Connection, selector, refs)
	if err != nil || !draft.Selector.Query(160000).SameEnvironment(q) {
		t.Fatal("tail scope lost", err)
	}
	draft.Selector.EnvironmentSelector.Scope.Value = "other"
	if _, err = bank.Resolve(batch.Connection, selector, refs); err != nil {
		t.Fatal("returned draft aliased bank", err)
	}
	snapshots := &ResultSnapshots{}
	token, err := snapshots.Freeze(&Result{Question: q, Rows: []json.RawMessage{row}})
	if err != nil {
		t.Fatal(err)
	}
	q.EnvironmentSelector.Scope.Value = "other"
	held, err := snapshots.Load(token)
	if err != nil || held.Question.Explore.EnvironmentSelector.Scope.Value != "prod" {
		t.Fatal("snapshot scope lost/aliased", err)
	}
	bad := scopedQuestion250("unreported", "")
	if _, err = snapshots.Freeze(&Result{Question: bad, Rows: []json.RawMessage{row}}); err == nil {
		t.Fatal("wrong-scope snapshot issued")
	}
}
func TestSelectorCaptureCatalogAndSourceGuards250(t *testing.T) {
	q := scopedQuestion250("named", "prod")
	row := `{"timestamp_ns":"130000000000","Body":"hello","ServiceName":"api","TraceId":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","ResourceAttributes":{"deployment.environment":"prod"}}`
	var currentRow atomic.Value
	currentRow.Store(row)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		sql := string(b)
		if strings.Contains(sql, "SELECT DISTINCT") {
			return
		}
		if strings.Contains(sql, "count()") {
			w.Write([]byte(`{"count":"1"}`))
			return
		}
		w.Write([]byte(currentRow.Load().(string)))
	}))
	defer server.Close()
	source := Source{Endpoint: server.URL, Client: server.Client()}
	store := CaptureStore{Dir: t.TempDir()}
	capture, err := store.Create(t.Context(), source, "Scoped", CaptureQuestion{Kind: "explore", Explore: &q}, "")
	if err != nil {
		t.Fatal(err)
	}
	reopened, err := store.Load(capture.ID)
	if err != nil || !reopened.Question.Explore.SameEnvironment(q) {
		t.Fatal("capture reopen lost policy", err)
	}
	page, err := store.Search(t.Context(), CaptureSearchQuery{}, "")
	if err != nil || len(page.Items) != 1 || page.Items[0].Environment != nil || page.Items[0].EnvironmentSelector == nil {
		t.Fatal("catalog erased scope", page, err)
	}
	name := "prod"
	filtered, err := store.Search(t.Context(), CaptureSearchQuery{Environment: &name}, "")
	if err != nil || len(filtered.Items) != 0 {
		t.Fatal("legacy catalog filter matches effective name", filtered, err)
	}
	followups, err := reopened.Followups(0)
	if err != nil {
		t.Fatal(err)
	}
	for _, action := range followups.Actions {
		if action.Discovery != nil || action.Query != nil && !action.Query.SameEnvironment(q) {
			t.Fatal("followup dropped policy", action)
		}
	}
	currentRow.Store(`{"timestamp_ns":"130000000000","Body":"hello","ResourceAttributes":{"deployment.environment.name":"prod"}}`)
	current, _ := identity.CurrentEnvironment("prod")
	wire, _ := current.Selector()
	explicit := q
	explicit.EnvironmentSelector = &wire
	explicitCapture, err := store.Create(t.Context(), source, "Explicit current", CaptureQuestion{Kind: "explore", Explore: &explicit}, "")
	if err != nil {
		t.Fatal(err)
	}
	legacy := explicit
	legacy.Environment = "prod"
	legacy.EnvironmentSelector = nil
	legacyCapture, err := store.Create(t.Context(), source, "Legacy current", CaptureQuestion{Kind: "explore", Explore: &legacy}, "")
	if err != nil {
		t.Fatal(err)
	}
	comparison, err := CompareEvidence(legacyCapture, explicitCapture)
	if err != nil || comparison.State != "comparable" {
		t.Fatal("equivalent current policies not comparable", comparison, err)
	}
	if _, err = store.RebuildCatalog(t.Context()); err != nil {
		t.Fatal(err)
	}
	mixed, err := store.Search(t.Context(), CaptureSearchQuery{}, "")
	if err != nil || len(mixed.Items) != 3 {
		t.Fatal("mixed catalog lost entries", err)
	}
	for i := 0; i < 21; i++ {
		base := capture
		if i%2 == 0 {
			base = legacyCapture
		}
		copy := detached(*base)
		copy.ID = ""
		copy.Title = fmt.Sprintf("Mixed page %02d", i)
		if _, err = store.publish(copy); err != nil {
			t.Fatal(err)
		}
	}
	first, err := store.Search(t.Context(), CaptureSearchQuery{}, "")
	if err != nil || len(first.Items) != 20 || !first.HasMore {
		t.Fatal("mixed first page", err)
	}
	second, err := store.Search(t.Context(), CaptureSearchQuery{}, first.NextCursor)
	if err != nil || len(second.Items) != 4 || second.HasMore {
		t.Fatal("mixed next page", err)
	}
	seen := map[string]bool{}
	for _, item := range append(first.Items, second.Items...) {
		if seen[item.ID] {
			t.Fatal("duplicate paged entry")
		}
		seen[item.ID] = true
		if item.EnvironmentSelector != nil && item.Environment != nil {
			t.Fatal("catalog conflates selector with legacy name")
		}
	}
	malformed := detached(*capture)
	malformed.Explore.Rows = []json.RawMessage{json.RawMessage(`{"timestamp_ns":"130000000000","Body":"wrong","ResourceAttributes":{"deployment.environment.name":"other"}}`)}
	id, _, err := captureDigest(malformed)
	if err != nil {
		t.Fatal(err)
	}
	malformed.ID = id
	data, err := json.Marshal(malformed)
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(filepath.Join(store.Dir, id+".json"), data, 0600); err != nil {
		t.Fatal(err)
	}
	if _, err = store.Load(id); err == nil || !strings.Contains(err.Error(), "different environment") {
		t.Fatal("correctly rehashed foreign-scope capture reopened")
	}
	if _, err = store.RebuildCatalog(t.Context()); err == nil {
		t.Fatal("foreign-scope capture indexed")
	}
	row = `{"timestamp_ns":"130000000000","Body":"wrong","ResourceAttributes":{"deployment.environment.name":"other"}}`
	currentRow.Store(row)
	if _, err = source.Fetch(t.Context(), q); err == nil {
		t.Fatal("wrong-scope source accepted")
	}
}
