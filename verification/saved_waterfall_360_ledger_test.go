package verification

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"slices"
	"testing"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/observability/quintgo/qobs"
	"github.com/casselc/observability/quintgo/validate"
)

type savedState360 struct {
	Calls, Source, Store, SQL, Offset, RetainedRecords int
	Kind, Basis                                        string
	RecordsLimited                                     bool
	Receipt                                            exploration.CaptureReceipt
	InspectorObserved                                  json.RawMessage
	Rendered                                           []struct {
		Index                                    int
		CapturedOccurrence, Timing, Parent, Zero string
	}
	WaterfallPage *struct {
		Question          exploration.Query
		Basis, CapturedAt string
		Page              struct{ Offset, Limit, Returned int }
		Quality           struct{ Extent, Upstream, Consistency string }
	}
}
type savedFact360 struct {
	Name          string
	Before, After savedState360
}
type savedCase360 struct {
	Key                                 string
	Question                            exploration.Query
	Capture                             exploration.CapturePage
	StoredBytesBefore, StoredBytesAfter string
	Pages                               []exploration.CapturePage
	Selections                          []struct {
		Offset, Index          int
		Row, InspectorObserved json.RawMessage
		CapturedOccurrence     string
	}
	OldInspect struct {
		Pending, Returned struct{ Before, After savedState360 }
	}
	Facts []savedFact360
}
type savedLedger360 struct {
	Version string
	Fixture struct {
		Environment, Namespace, Token string
		FromMS, ToMS                  int64
		Rows                          []struct {
			Occurrence, TraceID, SpanID, ParentSpanID, Service, Namespace, StartNS, DurationNS string
			Env                                                                                map[string]string
		}
	}
	Cases []savedCase360
	Calls []struct {
		Path, URL, Method string
		Body              json.RawMessage
	}
	Errors []string
}

var savedActions360 = []string{"reopen", "waterfall", "columns", "inspect_occurrence", "close_inspector", "table", "page_next", "page_previous", "return_inspect"}
var savedNames360 = []string{"reopen", "waterfall", "columns", "inspect-occurrence", "close-inspector", "table", "page-next", "page-previous", "return-inspect"}
var savedStores360 = []int{1, 1, 1, 1, 1, 1, 2, 3, 3}

func savedJSON360(v any) string {
	raw, _ := json.Marshal(v)
	var anyValue any
	if json.Unmarshal(raw, &anyValue) != nil {
		return "invalid"
	}
	return continuationJSON352(anyValue)
}
func savedPage360(l savedLedger360, c savedCase360, p exploration.CapturePage) error {
	if p.Explore == nil || savedJSON360(p.Receipt) != savedJSON360(c.Capture.Receipt) || p.RetainedRecords != 1000 || !p.RecordsLimited || p.Explore.Basis != c.Capture.Explore.Basis || p.Explore.CapturedAt != c.Capture.Explore.CapturedAt || p.Explore.Question.Limit != 100 || p.Explore.Question.Offset%100 != 0 {
		return fmt.Errorf("immutable stored page custody/clock/delivery")
	}
	q := p.Explore.Question
	expected := c.Question
	expected.Offset = q.Offset
	if savedJSON360(q) != savedJSON360(expected) {
		return fmt.Errorf("stored page changed whole query")
	}
	mode := 0
	if c.Key == "typed" {
		mode = 2
	}
	eligible := []int{}
	for i, r := range l.Fixture.Rows {
		n, ok := new(big.Int).SetString(r.StartNS, 10)
		if !ok {
			return fmt.Errorf("raw time")
		}
		if r.Namespace == l.Fixture.Namespace && continuationEnvironment352(mode, l.Fixture.Environment, r.Env) && n.Cmp(new(big.Int).Mul(big.NewInt(q.FromMS), big.NewInt(1000000))) >= 0 && n.Cmp(new(big.Int).Mul(big.NewInt(q.ToMS), big.NewInt(1000000))) < 0 {
			eligible = append(eligible, i)
		}
	}
	slices.SortFunc(eligible, func(a, b int) int {
		x, _ := new(big.Int).SetString(l.Fixture.Rows[a].StartNS, 10)
		y, _ := new(big.Int).SetString(l.Fixture.Rows[b].StartNS, 10)
		return x.Cmp(y)
	})
	if len(eligible) != 1007 && len(eligible) != 1014 || p.Explore.Count != fmt.Sprint(len(eligible)) {
		return fmt.Errorf("independent environment/namespace population count")
	}
	eligible = eligible[:1000]
	start := q.Offset
	end := min(start+100, len(eligible))
	if len(p.Explore.Rows) != end-start {
		return fmt.Errorf("stored page position count")
	}
	for local, index := range eligible[start:end] {
		var actual map[string]any
		if json.Unmarshal(p.Explore.Rows[local], &actual) != nil {
			return fmt.Errorf("raw stored record")
		}
		r := l.Fixture.Rows[index]
		for key, want := range map[string]string{"TraceId": r.TraceID, "SpanId": r.SpanID, "ParentSpanId": r.ParentSpanID, "ServiceName": r.Service, "timestamp_ns": r.StartNS, "duration_ns": r.DurationNS} {
			if actual[key] != want {
				return fmt.Errorf("original immutable occurrence%d field%s", start+local, key)
			}
		}
		attrs, _ := actual["Attributes"].(map[string]any)
		resource, _ := actual["ResourceAttributes"].(map[string]any)
		if attrs["sample.occurrence"] != r.Occurrence || attrs["run.id"] != l.Fixture.Token || resource["service.namespace"] != r.Namespace {
			return fmt.Errorf("original occurrence attributes/qualified namespace")
		}
	}
	return nil
}
func savedAdmission360(l savedLedger360, c savedCase360) error {
	if c.Capture.Explore == nil || c.Question.Validate() != nil || savedJSON360(c.Question) != savedJSON360(c.Capture.Receipt.Question.Explore) || c.StoredBytesBefore == "" || c.StoredBytesBefore != c.StoredBytesAfter || len(c.Pages) != 4 || len(c.Selections) < 2 {
		return fmt.Errorf("complete stored snapshot/bytes/selected occurrences")
	}
	var stored exploration.SavedCapture
	if json.Unmarshal([]byte(c.StoredBytesBefore), &stored) != nil || !peLiveCaptureHash(exploration.CapturePage{Receipt: stored.CaptureReceipt, Explore: stored.Explore}) || stored.ID != c.Capture.Receipt.ID || stored.Explore == nil || len(stored.Explore.Rows) != 1000 {
		return fmt.Errorf("genuine immutable capture hash/retention")
	}
	for _, p := range c.Pages {
		if err := savedPage360(l, c, p); err != nil {
			return err
		}
	}
	if err := savedPage360(l, c, c.Capture); err != nil {
		return err
	}
	pages := map[int]exploration.CapturePage{}
	for _, p := range c.Pages {
		pages[p.Explore.Question.Offset] = p
	}
	for _, s := range c.Selections {
		p, ok := pages[s.Offset]
		if !ok || s.Index < 0 || s.Index >= len(p.Explore.Rows) || savedJSON360(s.Row) != savedJSON360(p.Explore.Rows[s.Index]) || savedJSON360(s.InspectorObserved) != savedJSON360(s.Row) || s.CapturedOccurrence != fmt.Sprint(s.Offset+s.Index+1) {
			return fmt.Errorf("original page occurrence inspector")
		}
	}
	if savedJSON360(c.OldInspect.Pending.Before) != savedJSON360(c.OldInspect.Pending.After) || savedJSON360(c.OldInspect.Returned.Before) != savedJSON360(c.OldInspect.Returned.After) {
		return fmt.Errorf("pending-page or ABA old inspector acquired authority")
	}
	if len(c.Facts) != len(savedNames360) {
		return fmt.Errorf("incomplete observed stored lifecycle")
	}
	baseSource, baseStore := c.Facts[0].Before.Source, c.Facts[0].Before.Store
	for i, f := range c.Facts {
		if f.Name != savedNames360[i] || f.Before.Calls < 0 || f.After.Calls < f.Before.Calls || f.After.Calls > len(l.Calls) {
			return fmt.Errorf("observed gesture bounds")
		}
		gets, source := 0, 0
		for _, call := range l.Calls[f.Before.Calls:f.After.Calls] {
			if call.Method == "GET" && call.Path == "/api/captures/"+c.Capture.Receipt.ID {
				gets++
			}
			if call.Method == "POST" && slices.Contains([]string{"/api/explore", "/api/explore/page", "/api/record-aggregates", "/api/measurements", "/api/tail", "/api/discovery", "/api/inventory", "/api/metric-discovery"}, call.Path) {
				source++
			}
		}
		want := 0
		if i == 0 || i == 6 || i == 7 {
			want = 1
		}
		if source != 0 || f.After.Source != baseSource || f.After.SQL != f.Before.SQL || gets != want || f.After.Store-f.Before.Store != want || f.After.Store-baseStore != savedStores360[i] {
			return fmt.Errorf("passive display/source vs store GET budget")
		}
		p, ok := pages[f.After.Offset]
		if !ok || f.After.Basis != p.Explore.Basis || savedJSON360(f.After.Receipt) != savedJSON360(p.Receipt) || f.After.RetainedRecords != 1000 || !f.After.RecordsLimited {
			return fmt.Errorf("actual display page custody")
		}
		if f.After.WaterfallPage != nil {
			w := f.After.WaterfallPage
			if savedJSON360(w.Question) != savedJSON360(p.Explore.Question) || w.Basis != p.Explore.Basis || w.CapturedAt != p.Explore.CapturedAt || w.Page.Offset != f.After.Offset || w.Page.Limit != 100 || w.Page.Returned != len(p.Explore.Rows) || w.Quality.Extent != "bounded-page" || w.Quality.Upstream != "unknown" || w.Quality.Consistency != "independent-reads" {
				return fmt.Errorf("waterfall conferred whole trace authority or changed source")
			}
			rawRows := []map[string]any{}
			for _, raw := range p.Explore.Rows {
				var row map[string]any
				json.Unmarshal(raw, &row)
				rawRows = append(rawRows, row)
			}
			oracle := waterfallOracle360(rawRows)
			seen := map[int]bool{}
			for _, r := range f.After.Rendered {
				if r.Index < 0 || r.Index >= len(oracle) || seen[r.Index] || r.CapturedOccurrence != fmt.Sprint(f.After.Offset+r.Index+1) {
					return fmt.Errorf("visual order replaced original occurrence")
				}
				seen[r.Index] = true
				o := oracle[r.Index]
				timing := "unavailable"
				if o["timingValid"] == true {
					timing = "observed"
				}
				if r.Timing != timing || r.Parent != o["parentState"] {
					return fmt.Errorf("unknown timing/parent invented")
				}
			}
			if len(seen) != len(rawRows) {
				return fmt.Errorf("display dropped duplicate/unknown occurrences")
			}
		}
	}
	return nil
}
func savedModel360(t *testing.T, label string, stores, sources []int, coherent bool) bool {
	t.Helper()
	sink := &qobs.MemorySink{}
	rec := qobs.NewRecorder(sink, qobs.WithActor(label))
	for i, a := range savedActions360 {
		rec.Record(context.Background(), a, nil, "obs.stores", stores[i], "obs.source", sources[i], "obs.coherent", coherent || i != len(savedActions360)-1)
	}
	checker, e := validate.NewChecker("model/saved_waterfall_360.yaml", validate.Options{Quint: quintPath(t), Backend: "typescript", Dir: t.TempDir()})
	if e != nil {
		t.Fatal(e)
	}
	r, e := checker.Check(context.Background(), sink.Steps())
	if e != nil {
		t.Fatal(e)
	}
	artifact("saved-waterfall360-"+label+".txt", []byte(r.Summary()))
	return r.OK()
}
func TestQuintSavedWaterfallProtocol360(t *testing.T) {
	sources := make([]int, 9)
	if !savedModel360(t, "protocol-complete", savedStores360, sources, true) {
		t.Fatal("stored protocol refused")
	}
	source := slices.Clone(sources)
	source[1] = 1
	if savedModel360(t, "protocol-implicit-source", savedStores360, source, true) {
		t.Fatal("display acquired source")
	}
	stores := slices.Clone(savedStores360)
	stores[1]++
	if savedModel360(t, "protocol-speculative-store", stores, sources, true) {
		t.Fatal("display loaded another stored page")
	}
}
func TestQuintSavedWaterfallObserved360(t *testing.T) {
	path := os.Getenv("GOO_SAVED_WATERFALL_360_REPORT")
	if path == "" {
		t.Skip("explicit genuine capture visit report required")
	}
	raw, e := os.ReadFile(path)
	if e != nil {
		t.Fatal(e)
	}
	var l savedLedger360
	if json.Unmarshal(raw, &l) != nil || l.Version != "goo.capturedWaterfall360/1" || len(l.Cases) != 2 || len(l.Errors) != 0 {
		t.Fatal("incomplete observed visit")
	}
	sources := make([]int, 9)
	check := func(label string, c savedCase360, want bool) {
		t.Helper()
		err := savedAdmission360(l, c)
		if (err == nil) != want {
			t.Fatalf("%s independent admission: %v", label, err)
		}
		stores := slices.Clone(savedStores360)
		source := slices.Clone(sources)
		if len(c.Facts) == 9 {
			baseStore, baseSource := c.Facts[0].Before.Store, c.Facts[0].Before.Source
			for i, f := range c.Facts {
				stores[i] = f.After.Store - baseStore
				source[i] = f.After.Source - baseSource
			}
		}
		if savedModel360(t, label, stores, source, err == nil) != want {
			t.Fatalf("%s model verdict", label)
		}
	}
	for _, c := range l.Cases {
		check("actual-"+c.Key, c, true)
	}
	mutants := []func(*savedCase360){func(c *savedCase360) { c.Pages[1].Explore.Question.FromMS++ }, func(c *savedCase360) { c.Pages[2].Explore.Question.Offset = 0 }, func(c *savedCase360) { c.Pages[1].Explore.Rows[0] = json.RawMessage(`{}`) }, func(c *savedCase360) { c.Pages[1].Explore.CapturedAt = "2020-01-01T00:00:00Z" }, func(c *savedCase360) { c.Facts[1].After.Source++ }, func(c *savedCase360) { c.StoredBytesAfter += " " }}
	for i, mutate := range mutants {
		var fresh savedLedger360
		json.Unmarshal(raw, &fresh)
		c := fresh.Cases[0]
		mutate(&c)
		check(fmt.Sprintf("poison-%d", i), c, false)
	}
}
