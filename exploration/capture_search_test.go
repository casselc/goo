package exploration

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/casselc/goo/measurements"
	"github.com/casselc/goo/metrics"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
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

func TestCaptureSummaryProjectionBoundsAndDeclaredScopes(t *testing.T) {
	c := searchFixture(t)
	// A legacy-valid timestamp may have arbitrarily many fractional digits.
	// Canonicalization bounds the summary without rewriting its receipt.
	c.CapturedAt = "2026-10-29T00:00:00." + strings.Repeat("0", 70000) + "Z"
	summary := summarizeCapture(c.CaptureReceipt)
	if summary.CapturedAt != "2026-10-29T00:00:00Z" {
		t.Fatal(summary.CapturedAt)
	}
	if c.CapturedAt == summary.CapturedAt {
		t.Fatal("fixture lost original time representation")
	}
	s := CaptureStore{Dir: t.TempDir()}
	saved, e := s.publish(c)
	if e != nil {
		t.Fatal(e)
	}
	page, e := s.Search(t.Context(), CaptureSearchQuery{}, "")
	if e != nil || page.Items[0].ID != saved.ID {
		t.Fatal(page, e)
	}
	// Maximum UTF-8 byte lengths with worst-case JSON escaping stay compact.
	c.Title = strings.Repeat("<", 120)
	c.Question.Explore.Environment = strings.Repeat("<", 128)
	summary = summarizeCapture(c.CaptureReceipt)
	b, e := json.Marshal(summary)
	if e != nil || len(b) > 2048 {
		t.Fatal("summary is not bounded", len(b), e)
	}
	// A valid pair retains its declared signal/environment without mining predicates.
	a, bq := *c.Question.Explore, *c.Question.Explore
	a.GroupBy = []Field{{Kind: "column", Key: "service"}}
	bq.GroupBy = a.GroupBy
	pair := CaptureReceipt{Question: CaptureQuestion{Kind: "comparison", Comparison: &ComparisonQuestion{Baseline: a, Selection: bq}}}
	if e := pair.Question.Validate(); e != nil {
		t.Fatal(e)
	}
	ps := summarizeCapture(pair)
	if ps.Kind != "comparison" || ps.Signal != "spans" || ps.Environment == nil || *ps.Environment != a.Environment {
		t.Fatal(ps)
	}
}

func TestCaptureSummaryBoundaryTimes(t *testing.T) {
	for _, value := range []string{"0000-01-01T00:00:00+01:00", "9999-12-31T23:59:59-01:00", "2026-10-28T20:00:00-04:00"} {
		got := canonicalCaptureTime(value)
		a, e := time.Parse(time.RFC3339Nano, got)
		b, _ := time.Parse(time.RFC3339Nano, value)
		if e != nil || !a.Equal(b) || len(got) > 35 {
			t.Fatal(value, got, e)
		}
	}
}

func TestAcceptedLargeNativeAndComparisonCaptures(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		if strings.Contains(string(b), "toString(count()) AS count") && !strings.Contains(string(b), "GROUP BY") {
			fmt.Fprintln(w, `{"count":"0"}`)
		}
	}))
	defer server.Close()
	store := CaptureStore{Dir: t.TempDir()}
	source := Source{Endpoint: server.URL}
	yaml, e := os.ReadFile("../config/measurements/worker-active.yaml")
	if e != nil {
		t.Fatal(e)
	}
	d, e := measurements.Read(string(yaml))
	if e != nil {
		t.Fatal(e)
	}
	d.Source.SchemaURL = strings.Repeat("<", 1024)
	d.Source.Scope = strings.Repeat("<", 256)
	d.Source.Metric = strings.Repeat("<", 256)
	d.Source.Version = strings.Repeat("<", 128)
	d.Source.Attributes = map[string]string{}
	for i := range 16 {
		d.Source.Attributes[fmt.Sprint(i)+strings.Repeat("<", 120)] = strings.Repeat("<", 512)
	}
	leaves := make([]Predicate, 31)
	for i := range leaves {
		leaves[i] = Predicate{Field: &Field{Kind: "column", Key: "service"}, Op: "eq", Value: strings.Repeat("<", 340)}
	}
	q := Query{Signal: "spans", Environment: "prod", FromMS: 1000, ToMS: 61000, Limit: 100, Where: &Predicate{All: leaves}, GroupBy: []Field{{Kind: "column", Key: "service"}}}
	questions := []CaptureQuestion{
		{Kind: "measurement", Measurement: &measurements.Question{Definition: d, FromMS: 1000, ToMS: 61000}},
		{Kind: "comparison", Comparison: &ComparisonQuestion{Baseline: q, Selection: q}},
	}
	for _, question := range questions {
		c, e := store.Create(t.Context(), source, "Large family", question, "")
		if e != nil {
			t.Fatal(question.Kind, e)
		}
		wire, _ := json.Marshal(c.CaptureReceipt)
		if len(wire) <= 64<<10 {
			t.Fatal("fixture does not cross old index limit", question.Kind, len(wire))
		}
		p, e := store.Search(t.Context(), CaptureSearchQuery{Kind: question.Kind}, "")
		if e != nil || p.Matched != 1 || p.Items[0].ID != c.ID {
			t.Fatal(question.Kind, p, e)
		}
		if question.Kind == "measurement" && (p.Items[0].MeasurementKind != "gauge" || p.Items[0].Signal != "metrics") {
			t.Fatal(p.Items[0])
		}
	}
	if e = os.RemoveAll(store.catalogDir()); e != nil {
		t.Fatal(e)
	}
	if _, e = store.RebuildCatalog(t.Context()); e != nil {
		t.Fatal(e)
	}
}

func TestCatalogMigrationReservesOnlyKnownDerivedDirectories(t *testing.T) {
	s := CaptureStore{Dir: t.TempDir()}
	saved, e := s.publish(searchFixture(t))
	if e != nil {
		t.Fatal(e)
	}
	if e = os.RemoveAll(s.catalogDir()); e != nil {
		t.Fatal(e)
	}
	if e = os.Mkdir(filepath.Join(s.Dir, ".catalog-v1"), 0700); e != nil {
		t.Fatal(e)
	}
	// 10,000 physical user entries including one authoritative capture.
	for i := range CaptureCatalogLimit - 1 {
		if e = os.WriteFile(filepath.Join(s.Dir, fmt.Sprint("unrelated-", i)), nil, 0600); e != nil {
			t.Fatal(e)
		}
	}
	r, e := s.RebuildCatalog(t.Context())
	if e != nil || r.Indexed != 1 {
		t.Fatal("v2 migration rejected admitted store", r, e)
	}
	p, e := s.Search(t.Context(), CaptureSearchQuery{}, "")
	if e != nil || p.Total != 1 || p.Items[0].ID != saved.ID {
		t.Fatal(p, e)
	}
	if _, e = s.List(""); e != nil {
		t.Fatal("migration broke legacy listing", e)
	}
	if e = os.Mkdir(filepath.Join(s.Dir, ".catalog-v3"), 0700); e != nil {
		t.Fatal(e)
	}
	if _, e = s.Search(t.Context(), CaptureSearchQuery{}, ""); !errors.Is(e, ErrCaptureSearchBudget) {
		t.Fatal("arbitrary directory bypassed budget", e)
	}
	if _, e = s.RebuildCatalog(t.Context()); !errors.Is(e, ErrCaptureSearchBudget) {
		t.Fatal(e)
	}
	if _, e = s.List(""); e == nil {
		t.Fatal("legacy listing bypassed budget")
	}
}

// Projection/read-back coverage for every current family/subtype. This deliberately
// tests only derived metadata; actual native/pair creation is exercised above.
func TestCaptureCatalogSummaryFamilyReadback(t *testing.T) {
	base := searchFixture(t)
	q := *base.Question.Explore
	questions := []CaptureQuestion{
		base.Question,
		{Kind: "inventory", Inventory: &InventoryQuestion{Environment: q.Environment, FromMS: q.FromMS, ToMS: q.ToMS}},
		{Kind: "explanation", Explanation: &ExplanationRequest{Version: ExplanationVersion, Baseline: q, Selection: q, Fields: []Field{{Kind: "column", Key: "service"}}}},
	}
	for _, name := range []string{"worker-active", "checkout-calls", "checkout-duration", "checkout-lifecycle"} {
		b, e := os.ReadFile("../config/measurements/" + name + ".yaml")
		if e != nil {
			t.Fatal(e)
		}
		d, e := measurements.Read(string(b))
		if e != nil {
			t.Fatal(e)
		}
		questions = append(questions, CaptureQuestion{Kind: "measurement", Measurement: &measurements.Question{Definition: d, FromMS: 1700000000000, ToMS: 1700000060000}})
	}
	stock := detached(*questions[3].Measurement)
	stock.Definition.Kind = "stock"
	questions = append(questions, CaptureQuestion{Kind: "measurement", Measurement: &stock})
	state := detached(stock)
	d := &state.Definition
	d.Kind = "state"
	d.StateProfile = metrics.PodPhaseProfile
	d.Source.Services = nil
	d.Source.Namespace = nil
	d.Source.Resource = &measurements.ResourceBinding{Kind: "kubernetes-pod", ID: "pod-a"}
	d.Source.Scope = "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/k8sclusterreceiver"
	d.Source.Version = "0.137.0"
	d.Source.SchemaURL = ""
	d.Source.Metric = "k8s.pod.phase"
	d.Source.Unit = ""
	d.Source.ExactStream = &metrics.Stream{Resource: map[string]string{"k8s.pod.uid": "pod-a", "deployment.environment.name": d.Environment}, Scope: d.Source.Scope, Version: d.Source.Version, SchemaURL: d.Source.SchemaURL, Name: d.Source.Metric, Attributes: d.Source.Attributes}
	questions = append(questions, CaptureQuestion{Kind: "measurement", Measurement: &state})
	for i, question := range questions {
		t.Run(fmt.Sprint(question.Kind, "-", i), func(t *testing.T) {
			if e := question.Validate(); e != nil {
				t.Fatal(e)
			}
			s := CaptureStore{Dir: t.TempDir()}
			c := base
			c.Question = question
			// A file stamp is sufficient for this derived-only reader test, not evidence.
			if e := os.WriteFile(filepath.Join(s.Dir, c.ID+".json"), []byte("projection test stamp only"), 0600); e != nil {
				t.Fatal(e)
			}
			if e := s.indexCapture(&c); e != nil {
				t.Fatal(e)
			}
			got, _, _, e := s.catalogEntry(c.ID, CaptureSearchMaxBytes)
			if e != nil {
				t.Fatal(e)
			}
			if !reflect.DeepEqual(got, summarizeCapture(c.CaptureReceipt)) {
				t.Fatal(got)
			}
			if got.Environment == nil || *got.Environment != q.Environment {
				t.Fatal("lost declared environment", got)
			}
			expectedSignal := "spans"
			if question.Measurement != nil {
				expectedSignal = "metrics"
				if question.Measurement.Definition.Kind == "lifecycle" {
					expectedSignal = "logs"
				}
				if got.MeasurementKind != question.Measurement.Definition.Kind {
					t.Fatal(got)
				}
			}
			if got.Signal != expectedSignal {
				t.Fatal(got)
			}
		})
	}
}
