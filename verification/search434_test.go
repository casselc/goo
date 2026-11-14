package verification

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/observability/quintgo/qobs"
	"github.com/casselc/observability/quintgo/validate"
	"hegel.dev/go/hegel"
)

// Exercise the actual evaluator and compiler: SQL metacharacters are literal
// parameter values, the original OR remains an AND operand, and case is exact.
func TestHegelSourceSearch434(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		term := hegel.Draw(tc, hegel.SampledFrom([]string{" ERROR ", "A_%'\\[]", "(panic|error).*", "SELECT 'secret'", "雪Λ", "a\nb", "\tword\t"}))
		term += fmt.Sprint(hegel.Draw(tc, hegel.Integers(0, 999999)))
		service := hegel.Draw(tc, hegel.SampledFrom([]string{"api", "worker", "outside"}))
		body := hegel.Draw(tc, hegel.SampledFrom([]string{"before" + term + "after", strings.ToLower(term), "unrelated", term}))
		parent := exploration.Predicate{Any: []exploration.Predicate{{Field: &exploration.Field{Kind: "column", Key: "service"}, Op: "eq", Value: "api"}, {Field: &exploration.Field{Kind: "column", Key: "service"}, Op: "eq", Value: "worker"}}}
		message := exploration.Predicate{Field: &exploration.Field{Kind: "column", Key: "body"}, Op: "contains", Value: term}
		where := &exploration.Predicate{All: []exploration.Predicate{parent, message}}
		values := map[exploration.Field]string{{Kind: "column", Key: "service"}: service, {Kind: "column", Key: "body"}: body}
		got := exploration.Match(where, values)
		want := (service == "api" || service == "worker") && strings.Contains(body, term)
		if got != want || got && !exploration.Match(&parent, values) {
			tc.Fatal("literal search broadened or reinterpreted parent scope", service, body, term)
		}
		delete(values, exploration.Field{Kind: "column", Key: "body"})
		if exploration.Match(where, values) {
			tc.Fatal("missing message became a match")
		}
		to := int64(1800000 + hegel.Draw(tc, hegel.Integers(0, 999999)))
		q := exploration.Query{Signal: "logs", Environment: "prod", FromMS: to - 60000, ToMS: to, Limit: 100, Where: where}
		plan, err := exploration.Compile(q)
		if err != nil {
			tc.Fatal(err)
		}
		if strings.Contains(plan.Rows, term) || strings.Contains(plan.Count, term) || !strings.Contains(plan.Rows, "position(Body,") || !strings.Contains(plan.Count, "position(Body,") {
			tc.Fatal("literal text interpolated or contains operator changed", plan)
		}
		found := false
		for _, value := range plan.Parameters {
			found = found || value == term
		}
		if !found || plan.Parameters["from"] != fmt.Sprint(q.FromMS) || plan.Parameters["to"] != fmt.Sprint(q.ToMS) {
			tc.Fatal("exact literal or fixed event-time bounds lost", plan.Parameters)
		}
	}, options(150)...)
}

type searchHeld434 struct {
	Find       string `json:"find"`
	Selected   string `json:"selected"`
	ReadWindow struct {
		FromMS int64 `json:"fromMs"`
		ToMS   int64 `json:"toMs"`
	} `json:"readWindow"`
	Keys  []string `json:"keys"`
	Phase string   `json:"phase"`
}
type searchCase434 struct {
	Scenario     string          `json:"scenario"`
	LocalMatched int             `json:"localMatched"`
	Selector     json.RawMessage `json:"selector"`
	Term         string          `json:"term"`
	Poll         struct {
		FromMS int64 `json:"fromMs"`
		ToMS   int64 `json:"toMs"`
	} `json:"poll"`
	Question json.RawMessage `json:"question"`
	FirstRun struct {
		Question json.RawMessage `json:"question"`
		Count    string          `json:"count"`
		Result   struct {
			Rows []map[string]any `json:"rows"`
		} `json:"result"`
	} `json:"firstRun"`
	Steps      []sparseFact432 `json:"steps"`
	HeldBefore searchHeld434   `json:"heldBefore"`
	HeldAfter  searchHeld434   `json:"heldAfter"`
}
type searchReport434 struct {
	Version string          `json:"version"`
	Status  string          `json:"status"`
	Errors  []string        `json:"errors"`
	Cases   []searchCase434 `json:"cases"`
	Calls   []struct {
		Path   string          `json:"path"`
		Method string          `json:"method"`
		Body   json.RawMessage `json:"body"`
	} `json:"calls"`
}
type searchStep434 struct {
	Action                               string
	Reads                                int
	ScopeOK, WindowOK, LiteralOK, HeldOK bool
}

var searchActions434 = []string{"loaded", "find", "dirty_refused", "stage", "run", "return", "retire"}

func searchScope434(q, origin exploration.Query, term string) (bool, bool) {
	want := origin
	want.Limit, want.Offset = 100, 0
	message := exploration.Predicate{Field: &exploration.Field{Kind: "column", Key: "body"}, Op: "contains", Value: term}
	if origin.Where == nil {
		want.Where = &message
	} else {
		want.Where = &exploration.Predicate{All: []exploration.Predicate{*origin.Where, message}}
	}
	window := q.FromMS == origin.FromMS && q.ToMS == origin.ToMS
	q.FromMS, q.ToMS = want.FromMS, want.ToMS
	return sparseSame432(q, want), window
}

func searchObserved434(c searchCase434) ([]searchStep434, error) {
	if c.Term == "" || len(c.Steps) != len(searchActions434) || c.Poll.ToMS-c.Poll.FromMS != 60000 {
		return nil, fmt.Errorf("missing search/window observations")
	}
	selector, err := sparseSelector432(c.Selector, c.Poll.ToMS)
	if err != nil {
		return nil, err
	}
	origin := selector.Query(c.Poll.ToMS)
	if origin.Signal != "logs" {
		return nil, fmt.Errorf("source search must be logs")
	}
	staged, err := sparseQuery432(c.Question)
	if err != nil {
		return nil, err
	}
	first, err := sparseQuery432(c.FirstRun.Question)
	if err != nil {
		return nil, err
	}
	scopeOK, windowOK := searchScope434(staged, origin, c.Term)
	rows := c.FirstRun.Result.Rows
	expectedRows := 1
	if c.Scenario == "id-only" {
		expectedRows = 0
	}
	literalOK := len(rows) == expectedRows && c.FirstRun.Count == fmt.Sprint(expectedRows) && c.LocalMatched > 0
	for _, row := range rows {
		body, ok := row["Body"].(string)
		literalOK = literalOK && ok && strings.Contains(body, c.Term)
	}
	heldOK := len(c.HeldBefore.Keys) > 0 && slices.Equal(c.HeldBefore.Keys, c.HeldAfter.Keys) && c.HeldBefore.Find == c.Term && c.HeldAfter.Find == c.Term && c.HeldBefore.Selected == c.HeldAfter.Selected && c.HeldBefore.ReadWindow == c.HeldAfter.ReadWindow && c.HeldBefore.ReadWindow.FromMS == c.Poll.FromMS && c.HeldBefore.ReadWindow.ToMS == c.Poll.ToMS
	reads := 0
	steps := make([]searchStep434, 0, len(c.Steps))
	for index, fact := range c.Steps {
		if fact.Name != searchActions434[index] || fact.SourceReads == nil || *fact.SourceReads < 0 {
			return nil, fmt.Errorf("missing step %d", index)
		}
		reads += *fact.SourceReads
		action := fact.Name
		if action == "run" {
			action = "read_source"
		}
		if action == "return" {
			action = "return_held"
		}
		step := searchStep434{Action: action, Reads: reads, ScopeOK: true, WindowOK: true, LiteralOK: true, HeldOK: true}
		switch fact.Name {
		case "loaded":
			observed, e := sparseSelector432(fact.Query, c.Poll.ToMS)
			if e != nil {
				return nil, e
			}
			step.ScopeOK = tailSameScope433(observed, selector)
		case "stage":
			step.ScopeOK, step.WindowOK = scopeOK, windowOK
		case "run":
			q, e := sparseQuery432(fact.Query)
			if e != nil {
				return nil, e
			}
			step.ScopeOK, step.WindowOK = searchScope434(q, origin, c.Term)
			step.ScopeOK = step.ScopeOK && sparseSame432(q, staged) && sparseSame432(q, first)
			step.LiteralOK = literalOK
		case "return":
			step.HeldOK = heldOK
		}
		steps = append(steps, step)
	}
	return steps, nil
}

func searchCheck434(t *testing.T, steps []searchStep434) (bool, error) {
	t.Helper()
	quint, err := exec.LookPath("quint")
	if err != nil {
		return false, err
	}
	spec, err := filepath.Abs("model/search434.qnt")
	if err != nil {
		return false, err
	}
	manifest := "spec: " + spec + "\nmodule: search434\nconstants: {INITIAL: '0'}\ninvariants: [safety]\nactions:\n"
	for _, name := range searchActions434 {
		if name == "run" {
			name = "read_source"
		}
		if name == "return" {
			name = "return_held"
		}
		manifest += "  " + name + ":\n    expect: 'reads == {{.obs.reads}} and scopeOK == {{.obs.scopeOK}} and windowOK == {{.obs.windowOK}} and literalOK == {{.obs.literalOK}} and heldOK == {{.obs.heldOK}}'\n"
	}
	config := filepath.Join(t.TempDir(), "binding.yaml")
	if err = os.WriteFile(config, []byte(manifest), 0600); err != nil {
		return false, err
	}
	sink := &qobs.MemorySink{}
	rec := qobs.NewRecorder(sink)
	for _, s := range steps {
		rec.Record(context.Background(), s.Action, nil, "obs.reads", s.Reads, "obs.scopeOK", s.ScopeOK, "obs.windowOK", s.WindowOK, "obs.literalOK", s.LiteralOK, "obs.heldOK", s.HeldOK)
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

func TestQuintObservedSourceSearch434(t *testing.T) {
	file := os.Getenv("GOO_SEARCH434_REPORT")
	if file == "" {
		t.Skip("set GOO_SEARCH434_REPORT to the completed real activity-search journey report for mandatory milestone acceptance")
	}
	raw, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	var report searchReport434
	if err = json.Unmarshal(raw, &report); err != nil {
		t.Fatal(err)
	}
	if report.Version != "goo.activitySearch434/1" || report.Status != "passed" || len(report.Errors) > 0 || len(report.Cases) != 3 {
		t.Fatal("incomplete actual source-search report")
	}
	seen := map[string]bool{}
	for _, c := range report.Cases {
		if seen[c.Scenario] || (c.Scenario != "case" && c.Scenario != "literal" && c.Scenario != "id-only") {
			t.Fatal("missing or duplicate semantic scenario")
		}
		seen[c.Scenario] = true
		t.Run(c.Scenario, func(t *testing.T) {
			// Tie staged DOM observations to actual outgoing tail/explore requests.
			// The actor may make separate requests in its own uniquely scoped case.
			selector, e := sparseSelector432(c.Selector, c.Poll.ToMS)
			if e != nil {
				t.Fatal(e)
			}
			tailCount, queryCount := 0, 0
			for _, call := range report.Calls {
				if call.Method != "POST" {
					continue
				}
				if call.Path == "/api/tail" {
					s, e := sparseSelector432(call.Body, c.Poll.ToMS)
					if e == nil && tailSameScope433(s, selector) {
						tailCount++
						c.Steps[0].Query = call.Body
					}
				}
				if call.Path == "/api/explore" {
					q, e := sparseQuery432(call.Body)
					first, e2 := sparseQuery432(c.FirstRun.Question)
					if e == nil && e2 == nil && sparseSame432(q, first) {
						queryCount++
					}
				}
			}
			if tailCount != 1 || queryCount != 1 {
				t.Fatal("missing or repeated actual source requests", tailCount, queryCount)
			}
			steps, err := searchObserved434(c)
			if err != nil {
				t.Fatal(err)
			}
			ok, err := searchCheck434(t, steps)
			if err != nil || !ok {
				t.Fatal("actual search trace rejected", ok, err)
			}
			for _, control := range []string{"scope", "window", "implicit-stage-read", "implicit-return-read", "case-or-literal", "lost-held-row", "stale-owner"} {
				t.Run(control, func(t *testing.T) {
					encoded, _ := json.Marshal(c)
					var badCase searchCase434
					if err := json.Unmarshal(encoded, &badCase); err != nil {
						t.Fatal(err)
					}
					switch control {
					case "scope":
						q, e := sparseQuery432(badCase.Question)
						if e != nil {
							t.Fatal(e)
						}
						q.Where = &q.Where.All[1]
						badCase.Question, _ = json.Marshal(q)
					case "window":
						q, e := sparseQuery432(badCase.Question)
						if e != nil {
							t.Fatal(e)
						}
						q.FromMS--
						badCase.Question, _ = json.Marshal(q)
					case "implicit-stage-read":
						*badCase.Steps[3].SourceReads++
					case "implicit-return-read":
						*badCase.Steps[5].SourceReads++
					case "case-or-literal":
						badCase.FirstRun.Count = "999"
					case "lost-held-row":
						badCase.HeldAfter.Keys = badCase.HeldAfter.Keys[1:]
					case "stale-owner":
						*badCase.Steps[6].SourceReads++
					}
					bad, e := searchObserved434(badCase)
					if e != nil {
						t.Fatal("counterfactual malformed rather than semantically rejected", e)
					}
					ok, err := searchCheck434(t, bad)
					if err != nil {
						t.Fatal(err)
					}
					if ok {
						t.Fatal("unsafe counterfactual admitted", control)
					}
				})
			}
		})
	}
}
