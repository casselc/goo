package verification

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/observability/quintgo/qobs"
	"github.com/casselc/observability/quintgo/validate"
)

// Layout evidence is raw browser geometry, not a production "coherent" flag.
// The actual clipping intersection and source membership are admitted before
// replaying measured request counters through the existing read/history model.
type fixedRect384 struct{ Left, Right, Top, Bottom, Width, Height float64 }
type fixedViewport384 struct{ Width, Height, DocumentWidth float64 }
type fixedClip384 struct {
	Rect                                                                                               fixedRect384
	ClientWidth, ClientHeight, ClientLeft, ClientTop, ScrollLeft, ScrollTop, ScrollWidth, ScrollHeight float64
	Tag                                                                                                string
	OverflowX, OverflowY                                                                               string
}
type fixedGeometry384 struct {
	FirstSpanID, Phase                                string
	ScrollRoot, MainScroll                            fixedClip384
	WindowScroll                                      struct{ X, Y float64 }
	Viewport                                          fixedViewport384
	ScrollClips, ScopeScrollClips, ActionsScrollClips []fixedClip384
	Summary                                           struct {
		Heading string
		Facts   map[string]string
	}
	Scope, Actions, Primary fixedRect384
	PagePan                 bool
}
type fixedFrame384 struct {
	Width, Height int
	Kind          string
	Geometry      fixedGeometry384
}
type fixedCounters384 struct{ Source, SQL, Captures, Documents, StoreReads, Streams, CancelControls int }
type fixedFact384 struct {
	Name          string
	Before, After fixedCounters384
}
type fixedDraft384 struct {
	Viewport                   fixedViewport384
	ScrollClips                []fixedClip384
	ResultRows                 int
	EditorVisible, RunDisabled bool
	Raw, Error, Phase          string
	Expanded, HideDisabled     bool
	Rect                       fixedRect384
}
type fixedCase384 struct {
	RetainedActivityReturnVisible *bool
	Reopened                      *exploration.CapturePage
	StaleCapture                  *struct{ DialogOpen bool }
	Fixture                       struct{ Rows []traceSearchRawRow382 }
	Dirty, RefusedDraft           *fixedDraft384
	Name                          string
	Question, Request             exploration.Query
	Result                        exploration.Result
	Status                        int
	Frames                        []fixedFrame384
	Facts                         []fixedFact384
	HeldBefore, HeldAfter         json.RawMessage
	ReturnFocus                   *bool
}
type fixedReport384 struct {
	Capture, RawCapture                 *exploration.CapturePage
	StoredBytesBefore, StoredBytesAfter struct{ Search, Records string }
	BaselinePath, Phase, Status         string
	Version                             string
	Errors                              []string
	Fixture                             struct{ Rows []traceSearchRawRow382 }
	Cases                               []fixedCase384
}

func finiteRect384(r fixedRect384) bool {
	for _, v := range []float64{r.Left, r.Right, r.Top, r.Bottom, r.Width, r.Height} {
		if math.IsNaN(v) || math.IsInf(v, 0) {
			return false
		}
	}
	return r.Width > 0 && r.Height > 0 && r.Right > r.Left && r.Bottom > r.Top && math.Abs(r.Width-(r.Right-r.Left)) < 1 && math.Abs(r.Height-(r.Bottom-r.Top)) < 1
}
func clip384(g fixedGeometry384) (fixedRect384, error) {
	v := g.Viewport
	if v.Width <= 0 || v.Height <= 0 || v.DocumentWidth <= 0 || v.DocumentWidth > v.Width+1 || g.PagePan {
		return fixedRect384{}, fmt.Errorf("viewport/page or real scroll clip absent")
	}
	out := fixedRect384{Right: v.Width, Bottom: v.Height, Width: v.Width, Height: v.Height}
	if g.ScrollRoot.Tag != "" {
		root := g.ScrollRoot
		if root.ClientWidth <= 0 || root.ClientHeight <= 0 || root.ScrollWidth > root.ClientWidth+1 {
			return out, fmt.Errorf("invalid actual document scrollport")
		}
		out.Right = min(out.Right, root.ClientWidth)
		out.Bottom = min(out.Bottom, root.ClientHeight)
	}
	for _, c := range g.ScrollClips {
		if !finiteRect384(c.Rect) || c.ClientWidth <= 0 || c.ClientHeight <= 0 || c.ScrollTop < 0 || c.ScrollLeft < 0 {
			return out, fmt.Errorf("invalid actual scroll-container bounds")
		}
		if c.OverflowX == "auto" || c.OverflowX == "scroll" || c.OverflowX == "hidden" || c.OverflowX == "clip" {
			out.Left = max(out.Left, c.Rect.Left+c.ClientLeft)
			out.Right = min(out.Right, c.Rect.Left+c.ClientLeft+c.ClientWidth)
		}
		if c.OverflowY == "auto" || c.OverflowY == "scroll" || c.OverflowY == "hidden" || c.OverflowY == "clip" {
			out.Top = max(out.Top, c.Rect.Top+c.ClientTop)
			out.Bottom = min(out.Bottom, c.Rect.Top+c.ClientTop+c.ClientHeight)
		}
	}
	out.Width = out.Right - out.Left
	out.Height = out.Bottom - out.Top
	// A valid measured scroll container may be wholly below the viewport. Its
	// empty intersection is a real baseline failure, not malformed evidence.
	return out, nil
}
func inside384(r, clip fixedRect384) bool {
	return finiteRect384(r) && r.Left >= clip.Left-.5 && r.Right <= clip.Right+.5 && r.Top >= clip.Top-.5 && r.Bottom <= clip.Bottom+.5
}
func admitFrame384(f fixedFrame384, ids []string) error {
	g := f.Geometry
	if g.ScrollRoot.Tag == "" || g.ScrollRoot.ScrollTop != 0 || g.ScrollRoot.ScrollLeft != 0 || g.MainScroll.Tag != "MAIN" || g.MainScroll.ScrollTop != 0 || g.MainScroll.ScrollLeft != 0 || g.WindowScroll.X != 0 || g.WindowScroll.Y != 0 {
		return fmt.Errorf("frame was scrolled past secondary controls instead of fitting primary data")
	}
	clip, e := clip384(g)
	if e != nil {
		return e
	}
	scopeGeometry, actionsGeometry := g, g
	scopeGeometry.ScrollClips = g.ScopeScrollClips
	actionsGeometry.ScrollClips = g.ActionsScrollClips
	scopeClip, e := clip384(scopeGeometry)
	if e != nil {
		return e
	}
	actionsClip, e := clip384(actionsGeometry)
	if e != nil {
		return e
	}
	if (f.Width != 1280 && f.Width != 390) || float64(f.Width) != g.Viewport.Width || float64(f.Height) != g.Viewport.Height || f.Width == 1280 && f.Height != 1000 || f.Width == 390 && f.Height != 844 {
		return fmt.Errorf("unmatched actual viewport")
	}
	if !slices.Contains([]string{"record-table", "record-span-waterfall"}, f.Kind) {
		return fmt.Errorf("unknown supplied record kind")
	}
	if !inside384(g.Scope, scopeClip) || !inside384(g.Actions, actionsClip) || !finiteRect384(g.Primary) {
		return fmt.Errorf("applied scope/action row unavailable inside actual clipping ancestor")
	}
	// The documented frame must show the full first supplied record inside the
	// real scroll container, not just its viewport-clipped top edge.
	if g.Primary.Top < max(g.Scope.Bottom, g.Actions.Bottom)-1 || !inside384(g.Primary, clip) {
		return fmt.Errorf("first actual record/visual hidden by secondary chrome")
	}
	if !slices.Contains(ids, g.FirstSpanID) {
		return fmt.Errorf("visible first record is not from independently observed population")
	}
	return nil
}
func fixedMembers384(q exploration.Query, rows []traceSearchRawRow382) ([]string, error) {
	if q.Validate() != nil || q.Signal != "spans" || q.Where == nil || requiredTrace382(*q.Where) == "" {
		return nil, fmt.Errorf("fixed query must bound complete independent trace bank")
	}
	request := exploration.TraceSearchRequest{Environment: q.Environment, EnvironmentSelector: q.EnvironmentSelector, FromMS: q.FromMS, ToMS: q.ToMS}
	ids := []string{}
	for _, row := range traceSearchRecords382(rows) {
		if tsInScope(request, row) && tsPredicate(*q.Where, row.Values) {
			ids = append(ids, row.Values[exploration.Field{Kind: "column", Key: "span_id"}])
		}
	}
	slices.Sort(ids)
	return ids, nil
}
func admitFixed384(r fixedReport384) error {
	if r.Version != "goo.fixedResultLayout384/1" || r.Phase != "reviewed" || r.Status != "passed" || len(r.Errors) > 0 || len(r.Cases) != 4 {
		return fmt.Errorf("incomplete actual layout matrix")
	}
	if e := admitStoredBytes384(r); e != nil {
		return e
	}
	rows := r.Fixture.Rows
	if len(rows) == 0 {
		return fmt.Errorf("independent raw records absent")
	}
	names := map[string]bool{}
	for caseIndex, c := range r.Cases {
		if c.Name != []string{"activity", "ordinary", "saved-search", "saved-record"}[caseIndex] {
			return fmt.Errorf("actual finite context order changed")
		}
		if names[c.Name] || !slices.Contains([]string{"activity", "ordinary", "saved-search", "saved-record"}, c.Name) {
			return fmt.Errorf("wrong fixed-result contexts")
		}
		names[c.Name] = true
		if c.Status != 200 || !nhLiveSameQuery(c.Question, c.Request) || !nhLiveSameQuery(c.Question, c.Result.Question) {
			return fmt.Errorf("layout changed full fixed question")
		}
		bank := rows
		if len(c.Fixture.Rows) > 0 {
			bank = c.Fixture.Rows
		}
		ids, e := fixedMembers384(c.Question, bank)
		if e != nil {
			return e
		}
		actual, e := reportedResultIDs353(c.Result)
		if e != nil || len(ids) == 0 || !slices.Equal(ids, actual) || c.Result.Count != strconv.Itoa(len(ids)) || c.Result.HasMore || c.Result.Basis == "" || c.Result.State != "observed" {
			return fmt.Errorf("layout changed independent source membership/full count")
		}
		if _, e = time.Parse(time.RFC3339Nano, c.Result.CapturedAt); e != nil {
			return fmt.Errorf("source clock absent")
		}
		if c.Name == "saved-record" {
			page := c.Reopened
			if page == nil || page.Explore == nil || page.Receipt.Question.Kind != "explore" || page.Receipt.Question.Explore == nil || !nhLiveSameQuery(*page.Receipt.Question.Explore, c.Question) || !sameTransport384(*page.Explore, c.Result) || !sameTransport384(page, r.RawCapture) || !tcID(page.Receipt.ID, 64) {
				return fmt.Errorf("stored record page changed its immutable question/result/plan")
			}
			at, err := time.Parse(time.RFC3339Nano, page.Receipt.CapturedAt)
			read, readErr := time.Parse(time.RFC3339Nano, c.Result.CapturedAt)
			if err != nil || readErr != nil || at.Before(read) {
				return fmt.Errorf("stored custody/read clocks invalid")
			}
			// Historical plans belong to the physical capture bytes checked above;
			// opening a stored page must not require current compiler equivalence.
		} else {
			plan, err := exploration.Compile(c.Question)
			if err != nil || !reflect.DeepEqual(plan, c.Result.Plan) {
				return fmt.Errorf("layout changed live source plan")
			}
		}
		widths := map[int]bool{}
		modes := map[string]bool{}
		for _, f := range c.Frames {
			prefix, phase := "Applied question", "applied"
			if c.Name == "saved-record" {
				prefix, phase = "Stored evidence", "frozen"
			}
			if f.Geometry.Phase != phase {
				return fmt.Errorf("fixed result is not applied/frozen evidence")
			}
			if f.Geometry.Summary.Heading != fmt.Sprintf("%s · %s matching captured %s · %d displayed", prefix, c.Result.Count, c.Question.Signal, len(c.Result.Rows)) || f.Geometry.Summary.Facts["Source read"] != c.Result.CapturedAt || f.Geometry.Summary.Facts["Evidence basis"] != c.Result.Basis || f.Geometry.Summary.Facts["Signal"] != c.Question.Signal {
				return fmt.Errorf("visible shared summary changed count, signal, source clock or basis")
			}
			if e = admitFrame384(f, ids); e != nil {
				return fmt.Errorf("%s/%d/%s: %w", c.Name, f.Width, f.Kind, e)
			}
			widths[f.Width] = true
			modes[fmt.Sprintf("%s/%d", f.Kind, f.Width)] = true
		}
		wantFrames := 2
		if c.Name == "saved-record" {
			wantFrames = 4
			if !modes["record-span-waterfall/1280"] || !modes["record-span-waterfall/390"] {
				return fmt.Errorf("stored selected waterfall visibility not exercised")
			}
		}
		if len(c.Frames) != wantFrames || !modes["record-table/1280"] || !modes["record-table/390"] || !widths[390] || !widths[1280] {
			return fmt.Errorf("missing matched desktop/narrow frame")
		}
		if (c.Name == "activity" || c.Name == "saved-search") && (len(c.HeldBefore) == 0 || !reflect.DeepEqual(c.HeldBefore, c.HeldAfter) || c.ReturnFocus == nil || !*c.ReturnFocus) {
			return fmt.Errorf("held origin DOM/focus replaced on Back")
		}
		if c.Name == "ordinary" {
			if c.RetainedActivityReturnVisible == nil || *c.RetainedActivityReturnVisible {
				return fmt.Errorf("ordinary cold-entry question inherited Activity origin")
			}
			d := c.Dirty
			if c.StaleCapture == nil || c.StaleCapture.DialogOpen {
				return fmt.Errorf("retired capture callback opened new work")
			}
			for _, draft := range []*fixedDraft384{d, c.RefusedDraft} {
				if e := admitDraft384(draft); e != nil {
					return e
				}
			}
			if d.RunDisabled || c.RefusedDraft.RunDisabled || d.Raw != c.RefusedDraft.Raw || !strings.Contains(c.RefusedDraft.Error, "Duration") {
				return fmt.Errorf("invalid enabled Run was not visibly refused before reading source")
			}
		}
		if e = admitFacts384(c.Name, c.Facts); e != nil {
			return fmt.Errorf("%s: %w", c.Name, e)
		}
	}
	return nil
}
func admitDraft384(d *fixedDraft384) error {
	if d == nil || d.Raw != "not-a-duration" || d.Error == "" || d.Phase != "draft" || !d.Expanded || !d.HideDisabled || !d.EditorVisible || d.ResultRows != 0 {
		return fmt.Errorf("invalid raw draft hidden or unpublished evidence shown")
	}
	geometry := fixedGeometry384{Viewport: d.Viewport, ScrollClips: d.ScrollClips}
	clip, e := clip384(geometry)
	if e != nil || !inside384(d.Rect, clip) {
		return fmt.Errorf("dirty invalid editor outside real viewport")
	}
	return nil
}

// HTTP JSON.parse and report serialization may reorder object keys inside raw
// record messages. Transport comparison ignores key order, retaining json.Number
// so exact large integers do not pass through float64. Physical bytes/digests
// are still checked separately without canonicalization.
func transportValue384(v any) (any, error) {
	raw, e := json.Marshal(v)
	if e != nil {
		return nil, e
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	var out any
	e = decoder.Decode(&out)
	return out, e
}
func sameTransport384(a, b any) bool {
	x, e := transportValue384(a)
	y, f := transportValue384(b)
	return e == nil && f == nil && reflect.DeepEqual(x, y)
}
func transportDifference384(a, b any, path string) string {
	x, e := transportValue384(a)
	y, f := transportValue384(b)
	if e != nil || f != nil {
		return fmt.Sprintf("%s decode: %v / %v", path, e, f)
	}
	var visit func(any, any, string) string
	visit = func(x, y any, p string) string {
		if reflect.DeepEqual(x, y) {
			return ""
		}
		switch a := x.(type) {
		case map[string]any:
			if b, ok := y.(map[string]any); ok {
				keys := make([]string, 0, len(a))
				for key := range a {
					keys = append(keys, key)
				}
				slices.Sort(keys)
				for _, key := range keys {
					v, ok := b[key]
					if !ok {
						return p + "." + key + " missing"
					}
					if d := visit(a[key], v, p+"."+key); d != "" {
						return d
					}
				}
				return p + " additional keys"
			}
		case []any:
			if b, ok := y.([]any); ok {
				if len(a) != len(b) {
					return fmt.Sprintf("%s lengths %d/%d", p, len(a), len(b))
				}
				for i := range a {
					if d := visit(a[i], b[i], fmt.Sprintf("%s[%d]", p, i)); d != "" {
						return d
					}
				}
			}
		}
		return fmt.Sprintf("%s values %.100v / %.100v", p, x, y)
	}
	return visit(x, y, path)
}
func admitStoredBytes384(r fixedReport384) error {
	if r.StoredBytesBefore.Search == "" || r.StoredBytesBefore.Records == "" || r.StoredBytesBefore != r.StoredBytesAfter {
		return fmt.Errorf("layout altered immutable capture bytes")
	}
	for _, v := range []struct {
		raw  string
		page *exploration.CapturePage
	}{{r.StoredBytesBefore.Search, r.Capture}, {r.StoredBytesBefore.Records, r.RawCapture}} {
		var stored exploration.SavedCapture
		if v.page == nil || json.Unmarshal([]byte(v.raw), &stored) != nil || stored.ID != v.page.Receipt.ID {
			return fmt.Errorf("stored bytes do not own observed receipt")
		}
		id := stored.ID
		stored.ID = ""
		wire, e := json.Marshal(stored)
		if e != nil {
			return e
		}
		sum := sha256.Sum256(wire)
		if hex.EncodeToString(sum[:]) != id {
			return fmt.Errorf("immutable capture digest invalid")
		}
		stored.ID = id
		page, e := stored.Page("")
		if e != nil || !sameTransport384(page, v.page) {
			return fmt.Errorf("reopened %s page differs from custody bytes: %s", stored.Question.Kind, transportDifference384(page, v.page, "$"))
		}
	}
	return nil
}
func admitBaseline384(before, after fixedReport384) error {
	if before.Version != after.Version || before.Phase != "baseline" || before.Status != "baseline-observed" || len(before.Errors) > 0 || len(before.Cases) != 3 {
		return fmt.Errorf("actual paired original layout baseline absent")
	}
	for baselineIndex, old := range before.Cases {
		if old.Name != []string{"ordinary", "saved-search", "saved-record"}[baselineIndex] {
			return fmt.Errorf("paired original context order changed")
		}
		if old.Name != "ordinary" && old.Name != "saved-search" && old.Name != "saved-record" {
			return fmt.Errorf("unexpected original baseline context")
		}
		i := slices.IndexFunc(after.Cases, func(c fixedCase384) bool { return c.Name == old.Name })
		if i < 0 {
			return fmt.Errorf("original context missing after consolidation")
		}
		next := after.Cases[i]
		a, e := reportedResultIDs353(old.Result)
		b, f := reportedResultIDs353(next.Result)
		if e != nil || f != nil || !slices.Equal(a, b) || !nhLiveSameQuery(old.Request, next.Request) || !reflect.DeepEqual(old.Result.Plan, next.Result.Plan) || old.Result.Count != next.Result.Count {
			return fmt.Errorf("paired source question,plan,count or membership changed with layout")
		}
		if len(old.Frames) != 2 {
			return fmt.Errorf("missing paired original desktop/narrow frames")
		}
		widths := map[int]bool{}
		for _, frame := range old.Frames {
			clip, e := clip384(frame.Geometry)
			if e != nil {
				return e
			}
			if inside384(frame.Geometry.Primary, clip) {
				return fmt.Errorf("baseline did not independently reproduce below-fold first record")
			}
			widths[frame.Width] = true
		}
		if !widths[1280] || !widths[390] {
			return fmt.Errorf("unmatched original viewport pair")
		}
	}
	return nil
}
func action384(f fixedFact384) string {

	if strings.HasSuffix(f.Name, "-run") || f.Name == "activity-observe" {
		return "read"
	}
	return "history"
}
func admitFacts384(context string, facts []fixedFact384) error {
	if len(facts) == 0 {
		return fmt.Errorf("no measured actual gesture facts")
	}
	required := map[string][]string{
		"activity":     {"observe", "stage", "run", "geometry-desktop", "geometry-390", "back"},
		"ordinary":     {"stage", "run", "geometry-desktop", "geometry-390", "scope-disclosure", "dirty-invalid", "invalid-run-refused", "stale-capture"},
		"saved-search": {"open", "stage", "run", "geometry-desktop", "geometry-390", "back"},
		"saved-record": {"open", "geometry-desktop", "geometry-390", "geometry-desktop-waterfall", "geometry-390-waterfall", "kind-roundtrip", "scope-disclosure"},
	}[context]
	if len(required) != len(facts) {
		return fmt.Errorf("incomplete actual finite gesture sequence")
	}
	for i, name := range required {
		if facts[i].Name != context+"-"+name {
			return fmt.Errorf("actual gesture order changed")
		}
	}
	seenRead, seenHistory := false, false
	for i, f := range facts {
		a, b := f.Before, f.After
		store, streams := 0, 0
		if f.Name == "saved-search-open" || f.Name == "saved-record-open" {
			store = 1
		}
		if f.Name == "activity-observe" {
			streams = 1
		}
		if f.Name == "" || b.Source < a.Source || b.SQL < a.SQL || b.Captures != a.Captures || b.Documents != a.Documents || b.StoreReads-a.StoreReads != store || b.Streams-a.Streams != streams || b.CancelControls < a.CancelControls {
			return fmt.Errorf("layout gesture acquired custody/stream/source authority")
		}
		switch action384(f) {
		case "read":
			if b.Source-a.Source+b.Streams-a.Streams != 1 || b.SQL <= a.SQL {
				return fmt.Errorf("explicit read not independently observed")
			}
			seenRead = true
		case "history":
			if b.Source != a.Source || streams == 0 && b.SQL != a.SQL {
				return fmt.Errorf("kind/disclosure/Back acquired telemetry")
			}
			seenHistory = true
		default:
			return fmt.Errorf("unknown measured action")
		}
		if i > 0 && a != facts[i-1].After {
			return fmt.Errorf("unaccounted inter-gesture read")
		}
	}
	if context != "saved-record" && !seenRead || !seenHistory {
		return fmt.Errorf("missing read/history boundary")
	}
	return nil
}
func replayFixed384(t *testing.T, label string, cases []fixedCase384) bool {
	t.Helper()
	for _, c := range cases {
		sink := &qobs.MemorySink{}
		rec := qobs.NewRecorder(sink)
		base := c.Facts[0].Before.Source + c.Facts[0].Before.Streams
		for _, f := range c.Facts {
			rec.Record(t.Context(), action384(f), nil, "obs.reads", f.After.Source+f.After.Streams-base, "obs.captures", 0)
		}
		checker, e := validate.NewChecker("model/captured_trace_search_383.yaml", validate.Options{Quint: quintPath(t), Backend: "typescript", Dir: t.TempDir()})
		if e != nil {
			t.Fatal(e)
		}
		result, e := checker.Check(t.Context(), sink.Steps())
		if e != nil {
			t.Fatal(e)
		}
		if e = artifact("fixed-result-layout384-"+label+"-"+c.Name+"-model.txt", []byte(result.Summary())); e != nil {
			t.Fatal(e)
		}
		if !result.OK() {
			return false
		}
	}
	return true
}
func TestQuintFixedResultLayoutObserved384(t *testing.T) {
	path := os.Getenv("GOO_FIXED_RESULT_LAYOUT_384_REPORT")
	if path == "" {
		t.Skip("actual fixed-result layout report opt-in")
	}
	raw, e := os.ReadFile(path)
	if e != nil {
		t.Fatal(e)
	}
	var r fixedReport384
	if e = json.Unmarshal(raw, &r); e != nil {
		t.Fatal(e)
	}
	if e = admitFixed384(r); e != nil {
		t.Fatal(e)
	}
	baselinePath := r.BaselinePath
	if baselinePath == "" {
		t.Fatal("paired actual baseline report required")
	}
	if !filepath.IsAbs(baselinePath) {
		baselinePath = filepath.Join("..", baselinePath)
	}
	baselineRaw, e := os.ReadFile(baselinePath)
	if e != nil {
		t.Fatal(e)
	}
	var before fixedReport384
	if e = json.Unmarshal(baselineRaw, &before); e != nil {
		t.Fatal(e)
	}
	if e = admitBaseline384(before, r); e != nil {
		t.Fatal(e)
	}
	if !replayFixed384(t, "actual", r.Cases) {
		t.Fatal("actual request counters violate read/history model")
	}
	clone := func() fixedReport384 { var v fixedReport384; json.Unmarshal(raw, &v); return v }
	changes := []func(*fixedReport384){
		func(r *fixedReport384) { r.StoredBytesAfter.Records += " " },
		func(r *fixedReport384) {
			for i := range r.Cases {
				if r.Cases[i].Name == "ordinary" {
					r.Cases[i].StaleCapture.DialogOpen = true
				}
			}
		},
		func(r *fixedReport384) {
			g := &r.Cases[0].Frames[0].Geometry
			delta := g.Viewport.Height + 1 - g.Primary.Top
			g.Primary.Top += delta
			g.Primary.Bottom += delta
		},
		func(r *fixedReport384) { r.Cases[0].Frames[0].Geometry.Viewport.DocumentWidth += 1000 },
		func(r *fixedReport384) {
			clip := &r.Cases[0].Frames[0].Geometry.ScrollClips[0]
			clip.Rect.Bottom = clip.Rect.Top + 1
			clip.Rect.Height = 1
			clip.ClientHeight = 1
		},
		func(r *fixedReport384) { r.Cases[0].Request.ToMS++ },
		func(r *fixedReport384) { r.Cases[0].Result.Count = "0" },
		func(r *fixedReport384) {
			clock, _ := time.Parse(time.RFC3339Nano, r.Cases[0].Result.CapturedAt)
			r.Cases[0].Result.CapturedAt = clock.Add(time.Nanosecond).Format(time.RFC3339Nano)
		},
		func(r *fixedReport384) { r.Cases[0].Frames[0].Geometry.Summary.Facts["Evidence basis"] = "changed" },
		func(r *fixedReport384) {
			for i := range r.Cases {
				if r.Cases[i].Name == "ordinary" {
					rect := &r.Cases[i].Dirty.Rect
					delta := 10000 - rect.Top
					rect.Top += delta
					rect.Bottom += delta
				}
			}
		},
		func(r *fixedReport384) { r.Fixture.Rows = nil },
		func(r *fixedReport384) {
			for i := range r.Cases {
				if r.Cases[i].Name == "ordinary" {
					r.Cases[i].Dirty.Expanded = false
				}
			}
		},
		func(r *fixedReport384) { r.Cases[0].HeldAfter = json.RawMessage(`{"replaced":true}`) },
		func(r *fixedReport384) {
			for i := range r.Cases[0].Facts {
				if action384(r.Cases[0].Facts[i]) == "history" {
					r.Cases[0].Facts[i].After.StoreReads++
					return
				}
			}
		},
		func(r *fixedReport384) {
			for i := range r.Cases[0].Facts {
				if action384(r.Cases[0].Facts[i]) == "history" {
					r.Cases[0].Facts[i].After.Source++
					return
				}
			}
		},
	}
	for i, change := range changes {
		bad := clone()
		change(&bad)
		if admitFixed384(bad) == nil {
			t.Fatalf("semantic layout mutant %d admitted", i)
		}
	}
	bad := clone()
	last := &bad.Cases[0].Facts[len(bad.Cases[0].Facts)-1]
	last.After.Source++
	if replayFixed384(t, "implicit-read", bad.Cases) {
		t.Fatal("model admitted passive implicit read")
	}
	bad = clone()
	for i := range bad.Cases[0].Facts {
		if action384(bad.Cases[0].Facts[i]) == "read" {
			bad.Cases[0].Facts[i].After.Source--
			break
		}
	}
	if replayFixed384(t, "missing-read", bad.Cases) {
		t.Fatal("model admitted absent explicit read")
	}
	n := 0
	for _, c := range r.Cases {
		n += len(c.Facts)
	}
	t.Logf("checked %d measured facts, %d layout/source/ownership mutants and two model controls", n, len(changes))
}
