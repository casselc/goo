package verification

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/observability/quintgo/qobs"
	"github.com/casselc/observability/quintgo/validate"
)

type sparseRead432 struct {
	Question   json.RawMessage `json:"question"`
	Count      string          `json:"count"`
	CapturedAt string          `json:"capturedAt"`
}
type sparseHeld432 struct {
	Selector   json.RawMessage `json:"selector"`
	ReadWindow struct {
		FromMS int64 `json:"fromMs"`
		ToMS   int64 `json:"toMs"`
	} `json:"readWindow"`
	ObservedAt string `json:"observedAt"`
	Retained   *int   `json:"retained"`
	Phase      string `json:"phase"`
}
type sparseFact432 struct {
	Name        string          `json:"name"`
	SourceReads *int            `json:"sourceReads"`
	Query       json.RawMessage `json:"query"`
}
type sparseCase432 struct {
	Signal   string          `json:"signal"`
	Selector json.RawMessage `json:"selector"`
	Poll     struct {
		FromMS     int64  `json:"fromMs"`
		ToMS       int64  `json:"toMs"`
		ObservedAt string `json:"observedAt"`
		Retained   *int   `json:"retained"`
		RowCount   *int   `json:"rowCount"`
		Limited    *bool  `json:"limited"`
	} `json:"poll"`
	Question   json.RawMessage `json:"question"`
	FirstRun   sparseRead432   `json:"firstRun"`
	SecondRun  sparseRead432   `json:"secondRun"`
	Steps      []sparseFact432 `json:"steps"`
	HeldBefore sparseHeld432   `json:"heldBefore"`
	HeldAfter  sparseHeld432   `json:"heldAfter"`
}
type sparseReport432 struct {
	Version string          `json:"version"`
	Status  string          `json:"status"`
	Errors  []string        `json:"errors"`
	Cases   []sparseCase432 `json:"cases"`
}
type sparseStep432 struct {
	Action                                string
	Reads                                 int
	ScopeOK, WindowOK, ClockFresh, HeldOK bool
}

var sparseActions432 = []struct{ name, action string }{{"completed", "complete_poll"}, {"dirty_refused", "refuse_draft"}, {"stage", "stage_wide"}, {"run", "read_wide"}, {"rerun", "reread_wide"}, {"return", "return_held"}, {"resume", "resume_recent"}, {"retire", "retire_context"}}

func sparseDecode432(raw json.RawMessage, out any) error {
	if len(raw) == 0 || bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
		return fmt.Errorf("missing observed source request")
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(out); err != nil {
		return err
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		return fmt.Errorf("unexpected observed request tail")
	}
	return nil
}
func sparseQuery432(raw json.RawMessage) (exploration.Query, error) {
	var q exploration.Query
	if err := sparseDecode432(raw, &q); err != nil {
		return q, err
	}
	return q, q.Validate()
}
func sparseSelector432(raw json.RawMessage, toMS int64) (exploration.TailSelector, error) {
	var selector exploration.TailSelector
	if err := sparseDecode432(raw, &selector); err != nil {
		return selector, err
	}
	return selector, selector.Query(toMS).Validate()
}
func sparseSame432(a, b any) bool {
	left, err := json.Marshal(a)
	if err != nil {
		return false
	}
	right, err := json.Marshal(b)
	return err == nil && bytes.Equal(left, right)
}
func sparseScope432(q, origin exploration.Query) bool {
	return q.Signal == origin.Signal && q.SameEnvironment(origin) && sparseSame432(q.Where, origin.Where) && q.Limit == 100 && q.Offset == 0 && len(q.GroupBy) == 0 && q.Numeric == nil && q.NumericHistogram == nil && q.ReportedEdge == nil && !q.GroupTimeSeries && !q.SpanStatus && q.BucketMS == 0 && q.Summary == ""
}
func sparseWindow432(q exploration.Query, toMS int64) bool {
	return q.ToMS == toMS && q.ToMS-q.FromMS == 15*60000
}
func sparseHeldEqual432(before, after sparseHeld432, selector exploration.TailSelector, c sparseCase432) bool {
	b, e1 := sparseSelector432(before.Selector, c.Poll.ToMS)
	a, e2 := sparseSelector432(after.Selector, c.Poll.ToMS)
	return e1 == nil && e2 == nil && sparseSame432(b, selector) && sparseSame432(a, selector) && before.ReadWindow.FromMS == c.Poll.FromMS && before.ReadWindow.ToMS == c.Poll.ToMS && after.ReadWindow == before.ReadWindow && before.ObservedAt == c.Poll.ObservedAt && after.ObservedAt == before.ObservedAt && before.Retained != nil && after.Retained != nil && *before.Retained == 0 && *after.Retained == 0 && after.Phase == "paused"
}
func sparseObservedSteps432(c sparseCase432) ([]sparseStep432, error) {
	if len(c.Steps) != len(sparseActions432) || c.Poll.Retained == nil || c.Poll.RowCount == nil || c.Poll.Limited == nil || *c.Poll.Retained != 0 || *c.Poll.RowCount != 0 || *c.Poll.Limited || c.Poll.ToMS-c.Poll.FromMS != exploration.TailWindowMS {
		return nil, fmt.Errorf("missing complete successful empty-buffer observation")
	}
	selector, err := sparseSelector432(c.Selector, c.Poll.ToMS)
	if err != nil {
		return nil, fmt.Errorf("original selector: %w", err)
	}
	origin := selector.Query(c.Poll.ToMS)
	if origin.Signal != c.Signal {
		return nil, fmt.Errorf("signal witness mismatch")
	}
	staged, err := sparseQuery432(c.Question)
	if err != nil {
		return nil, fmt.Errorf("staged admission: %w", err)
	}
	first, err := sparseQuery432(c.FirstRun.Question)
	if err != nil {
		return nil, fmt.Errorf("first source admission: %w", err)
	}
	second, err := sparseQuery432(c.SecondRun.Question)
	if err != nil {
		return nil, fmt.Errorf("second source admission: %w", err)
	}
	pollClock, err := time.Parse(time.RFC3339Nano, c.Poll.ObservedAt)
	if err != nil {
		return nil, fmt.Errorf("missing completed-poll clock")
	}
	firstClock, err := time.Parse(time.RFC3339Nano, c.FirstRun.CapturedAt)
	if err != nil {
		return nil, fmt.Errorf("missing first independent read clock")
	}
	secondClock, err := time.Parse(time.RFC3339Nano, c.SecondRun.CapturedAt)
	if err != nil {
		return nil, fmt.Errorf("missing second independent read clock")
	}
	if c.FirstRun.Count != "1" || c.SecondRun.Count != "2" {
		return nil, fmt.Errorf("late-record source oracle changed")
	}
	states := make([]sparseStep432, 0, len(c.Steps))
	reads := 0
	scopeOK, windowOK, clockFresh, heldOK := true, true, false, false
	for index, fact := range c.Steps {
		expected := sparseActions432[index]
		if fact.Name != expected.name || fact.SourceReads == nil || *fact.SourceReads < 0 {
			return nil, fmt.Errorf("incomplete observed step %d", index)
		}
		reads += *fact.SourceReads
		switch fact.Name {
		case "completed", "resume":
			observed, err := sparseSelector432(fact.Query, c.Poll.ToMS)
			if err != nil {
				return nil, fmt.Errorf("tail admission %s: %w", fact.Name, err)
			}
			scopeOK = scopeOK && sparseSame432(observed, selector)
		case "stage":
			scopeOK = sparseScope432(staged, origin)
			windowOK = sparseWindow432(staged, c.Poll.ToMS)
		case "run", "rerun":
			query, err := sparseQuery432(fact.Query)
			if err != nil {
				return nil, fmt.Errorf("actual POST body %s: %w", fact.Name, err)
			}
			responseQuestion := first
			if fact.Name == "rerun" {
				responseQuestion = second
			}
			scopeOK = sparseScope432(query, origin) && sparseSame432(query, responseQuestion) && sparseSame432(query, staged)
			windowOK = sparseWindow432(query, c.Poll.ToMS)
			clockFresh = firstClock.After(pollClock)
			if fact.Name == "rerun" {
				clockFresh = clockFresh && secondClock.After(firstClock)
			}
		case "return":
			heldOK = sparseHeldEqual432(c.HeldBefore, c.HeldAfter, selector, c)
		}
		states = append(states, sparseStep432{Action: expected.action, Reads: reads, ScopeOK: scopeOK, WindowOK: windowOK, ClockFresh: clockFresh, HeldOK: heldOK})
	}
	return states, nil
}
func sparseCheck432(t *testing.T, steps []sparseStep432) (bool, error) {
	t.Helper()
	quint, err := exec.LookPath("quint")
	if err != nil {
		return false, fmt.Errorf("native sparse recovery validation requires quint: %w", err)
	}
	spec, err := filepath.Abs("model/sparse_recovery432.qnt")
	if err != nil {
		return false, err
	}
	manifest := "spec: " + spec + "\nmodule: sparse_recovery432\nconstants: {INITIAL: '0'}\ninvariants: [safety]\nactions:\n"
	for _, entry := range sparseActions432 {
		manifest += "  " + entry.action + ":\n    expect: 'reads == {{.obs.reads}} and scopeOK == {{.obs.scopeOK}} and windowOK == {{.obs.windowOK}} and clockFresh == {{.obs.clockFresh}} and heldOK == {{.obs.heldOK}}'\n"
	}
	config := filepath.Join(t.TempDir(), "binding.yaml")
	if err = os.WriteFile(config, []byte(manifest), 0600); err != nil {
		return false, err
	}
	sink := &qobs.MemorySink{}
	rec := qobs.NewRecorder(sink)
	for _, step := range steps {
		rec.Record(context.Background(), step.Action, nil, "obs.reads", step.Reads, "obs.scopeOK", step.ScopeOK, "obs.windowOK", step.WindowOK, "obs.clockFresh", step.ClockFresh, "obs.heldOK", step.HeldOK)
	}
	checker, err := validate.NewChecker(config, validate.Options{Quint: quint, Backend: "typescript", Dir: t.TempDir()})
	if err != nil {
		return false, err
	}
	result, err := checker.Check(context.Background(), sink.Steps())
	if err != nil {
		return false, err
	}
	return result.OK(), nil
}
func TestQuintObservedSparseRecovery432(t *testing.T) {
	path := os.Getenv("GOO_SPARSE432_REPORT")
	if path == "" {
		t.Skip("set GOO_SPARSE432_REPORT to the completed real sparse-recovery journey report for mandatory milestone validation")
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var report sparseReport432
	if err = json.Unmarshal(raw, &report); err != nil {
		t.Fatal(err)
	}
	if report.Version != "goo.sparseData432/1" || report.Status != "passed" || len(report.Errors) > 0 || len(report.Cases) != 2 {
		t.Fatal("incomplete real sparse recovery report")
	}
	seen := map[string]bool{}
	for _, c := range report.Cases {
		t.Run(c.Signal, func(t *testing.T) {
			if seen[c.Signal] || c.Signal != "logs" && c.Signal != "spans" {
				t.Fatal("duplicate or unsupported signal")
			}
			seen[c.Signal] = true
			steps, err := sparseObservedSteps432(c)
			if err != nil {
				t.Fatal(err)
			}
			ok, err := sparseCheck432(t, steps)
			if err != nil || !ok {
				t.Fatal("real recovery trace rejected", ok, err)
			}
			for _, control := range []string{"window", "scope", "implicit-stage-read", "implicit-return-read", "stale-read-clock"} {
				t.Run(control, func(t *testing.T) {
					encoded, _ := json.Marshal(c)
					var bad sparseCase432
					_ = json.Unmarshal(encoded, &bad)
					switch control {
					case "window":
						query, e := sparseQuery432(bad.Question)
						if e != nil {
							t.Fatal(e)
						}
						query.ToMS++
						bad.Question, _ = json.Marshal(query)
					case "scope":
						query, e := sparseQuery432(bad.Question)
						if e != nil {
							t.Fatal(e)
						}
						query.Where = &exploration.Predicate{Field: &exploration.Field{Kind: "column", Key: "service"}, Op: "eq", Value: "counterfactual-service"}
						bad.Question, _ = json.Marshal(query)
					case "implicit-stage-read":
						*bad.Steps[2].SourceReads++
					case "implicit-return-read":
						*bad.Steps[5].SourceReads++
					case "stale-read-clock":
						bad.FirstRun.CapturedAt = bad.Poll.ObservedAt
					}
					observed, e := sparseObservedSteps432(bad)
					if e != nil {
						t.Fatal("counterfactual should retain valid encoding/admission", e)
					}
					accepted, e := sparseCheck432(t, observed)
					if e != nil {
						t.Fatal(e)
					}
					if accepted {
						t.Fatal("unsafe recovery trace accepted", control)
					}
				})
			}
		})
	}
	if !reflect.DeepEqual(seen, map[string]bool{"logs": true, "spans": true}) {
		t.Fatal("missing signal journey")
	}
}
