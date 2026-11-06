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

type fixedState358 struct {
	Calls, Source                    int
	Phase, Label, RawDuration, Error string
	Expanded                         any
	Disabled                         bool
	Evidence                         json.RawMessage
	Results                          []json.RawMessage
}
type fixedFact358 struct {
	Name          string
	Before, After fixedState358
}
type fixedRead358 struct {
	Request  exploration.Query
	Response exploration.Result
}
type fixedLedger358 struct {
	Version, Phase string
	Fixture        struct {
		Environment, Service, Namespace string
		Question, Numeric               exploration.Query
		Rows                            []struct {
			ID         string
			At         int64
			Env        map[string]string
			DurationNS string
		}
	}
	Cases []struct {
		Key                  string
		Width, Before, After int
		Request              exploration.Query
		Result               exploration.Result
	}
	Calls    []serviceContinuationCall352
	Errors   []string
	Protocol struct {
		Question            exploration.Query
		Initial             exploration.Result
		ValidRun, Recovered fixedRead358
		Held                struct {
			Request  exploration.Query
			Response exploration.Result
			Abort    json.RawMessage
		}
		Facts     []fixedFact358
		OldEditor struct{ Before, After fixedState358 }
	}
}

var fixedNames358 = []string{"edit-clean", "hide-clean", "edit-invalid", "invalid-run-blocked", "draft-return", "valid-fix", "run-valid", "edit-before-cancel", "cancel-read", "cancel", "late-result", "run-recover", "inspect", "applied-return"}
var fixedActions358 = []string{"edit_clean", "hide_clean", "edit_invalid", "invalid_run_blocked", "draft_return", "valid_fix", "run_valid", "edit_before_cancel", "cancel_read", "cancel_reading", "late_result", "run_recover", "inspect", "applied_return"}
var fixedReads358 = []int{0, 0, 0, 0, 0, 0, 1, 1, 2, 2, 2, 3, 3, 3}
var fixedOpen358 = []bool{true, false, true, true, true, true, false, true, true, true, true, false, false, false}
var fixedPhases358 = []string{"applied", "applied", "draft", "draft", "draft", "draft", "applied", "applied", "running", "cancelled", "cancelled", "applied", "applied", "applied"}

func fixedExpanded358(value any) (bool, bool) {
	switch v := value.(type) {
	case bool:
		return v, true
	case string:
		if v == "true" {
			return true, true
		}
		if v == "false" {
			return false, true
		}
	}
	return false, false
}
func fixedResult358(l fixedLedger358, r exploration.Result, q exploration.Query) error {
	if continuationJSON352(r.Question) != continuationJSON352(q) {
		return fmt.Errorf("whole fixed question")
	}
	ids := []string{}
	sum := new(big.Int)
	for _, raw := range l.Fixture.Rows {
		if raw.At >= q.FromMS && raw.At < q.ToMS && continuationEnvironment352(2, l.Fixture.Environment, raw.Env) {
			ids = append(ids, raw.ID)
			d, ok := new(big.Int).SetString(raw.DurationNS, 10)
			if !ok {
				return fmt.Errorf("fixture duration")
			}
			sum.Add(sum, d)
		}
	}
	slices.Sort(ids)
	actual, e := reportedResultIDs353(r)
	if e != nil {
		return e
	}
	if len(ids) != 8 || !slices.Equal(ids, actual) || r.Count != fmt.Sprint(len(ids)) {
		return fmt.Errorf("independent typed fixed population")
	}
	if q.Summary != "" && (r.Duration == nil || r.Duration.Validate() != nil || r.Duration.SumNS != sum.String()) {
		return fmt.Errorf("independent duration aggregate")
	}
	return nil
}
func fixedAdmission358(l fixedLedger358) error {
	p := l.Protocol
	q := l.Fixture.Question
	if l.Version != "goo.fixedQuestion358/1" || l.Phase != "final" || len(l.Errors) != 0 || len(l.Cases) != 4 || continuationJSON352(q) != continuationJSON352(p.Question) || q.Validate() != nil || q.Summary != "span-duration" || !q.SpanStatus || !q.GroupTimeSeries || q.BucketMS != 1000 || q.EnvironmentSelector == nil || q.EnvironmentSelector.Policy != "effective" {
		return fmt.Errorf("complete final fixed grouped-duration protocol")
	}
	seen := map[string]bool{}
	for _, c := range l.Cases {
		key := fmt.Sprintf("%s/%d", c.Key, c.Width)
		if seen[key] || c.Width != 1280 && c.Width != 390 || c.After != c.Before+1 {
			return fmt.Errorf("finite native primary/numeric matrix")
		}
		seen[key] = true
		expected := q
		if c.Key == "numeric" {
			expected = l.Fixture.Numeric
		} else if c.Key != "package" {
			return fmt.Errorf("unknown native profile")
		}
		if continuationJSON352(c.Request) != continuationJSON352(expected) {
			return fmt.Errorf("presentation converted profile")
		}
		if err := fixedResult358(l, c.Result, expected); err != nil {
			return err
		}
	}
	for _, r := range []exploration.Result{p.Initial, p.ValidRun.Response, p.Held.Response, p.Recovered.Response} {
		if err := fixedResult358(l, r, q); err != nil {
			return err
		}
	}
	for _, request := range []exploration.Query{p.ValidRun.Request, p.Held.Request, p.Recovered.Request} {
		if continuationJSON352(request) != continuationJSON352(q) {
			return fmt.Errorf("edit/rerun changed fixed profile or window")
		}
	}
	if len(p.Held.Abort) == 0 || string(p.Held.Abort) == "null" {
		return fmt.Errorf("no genuine held read abort")
	}
	if len(p.Facts) != len(fixedNames358) {
		return fmt.Errorf("incomplete actual presentation lifecycle")
	}
	base := p.Facts[0].Before.Source
	for i, f := range p.Facts {
		opened, ok := fixedExpanded358(f.After.Expanded)
		if !ok || opened != fixedOpen358[i] || f.After.Phase != fixedPhases358[i] || f.Name != fixedNames358[i] {
			return fmt.Errorf("fact%d actual phase/editor state", i)
		}
		required := i >= 2 && i <= 5 || i == 9 || i == 10
		if f.After.Disabled != required {
			return fmt.Errorf("dirty/terminal editor hidden affordance")
		}
		if f.Before.Calls < 0 || f.After.Calls < f.Before.Calls || f.After.Calls > len(l.Calls) {
			return fmt.Errorf("raw source boundaries")
		}
		source := 0
		for _, call := range l.Calls[f.Before.Calls:f.After.Calls] {
			if !serviceTemplateSource356(call) {
				continue
			}
			source++
			var actual exploration.Query
			if call.Path != "/api/explore" || json.Unmarshal(call.Body, &actual) != nil || continuationJSON352(actual) != continuationJSON352(q) {
				return fmt.Errorf("unexpected source or changed Run question")
			}
		}
		want := 0
		if i == 6 || i == 8 || i == 11 {
			want = 1
		}
		if source != want || f.After.Source-f.Before.Source != want || f.After.Source-base != fixedReads358[i] {
			return fmt.Errorf("presentation acquired telemetry")
		}
		if i > 0 && f.Before.Source != p.Facts[i-1].After.Source {
			return fmt.Errorf("unaccounted inter-gesture read")
		}
		if i == 2 || i == 3 || i == 4 {
			if f.After.RawDuration != "abc" {
				return fmt.Errorf("raw invalid draft lost")
			}
		}
		if slices.Contains([]int{0, 1, 3, 4, 7, 10, 12, 13}, i) {
			if continuationJSON352(f.Before.Evidence) != continuationJSON352(f.After.Evidence) || continuationJSON352(f.Before.Results) != continuationJSON352(f.After.Results) {
				return fmt.Errorf("passive edit/back/late-result replaced evidence DOM")
			}
		}
	}
	if continuationJSON352(p.OldEditor.Before) != continuationJSON352(p.OldEditor.After) || p.OldEditor.Before.Phase != "applied" {
		return fmt.Errorf("retired editor action renewed its lease")
	}
	return nil
}
func fixedModel358(t *testing.T, label string, reads []int, opens []bool, coherent bool) bool {
	t.Helper()
	sink := &qobs.MemorySink{}
	rec := qobs.NewRecorder(sink, qobs.WithActor(label))
	for i, a := range fixedActions358 {
		rec.Record(context.Background(), a, nil, "obs.reads", reads[i], "obs.opened", opens[i], "obs.coherent", coherent || i != len(fixedActions358)-1)
	}
	checker, e := validate.NewChecker("model/fixed_question_358.yaml", validate.Options{Quint: quintPath(t), Backend: "typescript", Dir: t.TempDir()})
	if e != nil {
		t.Fatal(e)
	}
	r, e := checker.Check(context.Background(), sink.Steps())
	if e != nil {
		t.Fatal(e)
	}
	artifact("fixed-question358-"+label+".txt", []byte(r.Summary()))
	return r.OK()
}
func TestQuintFixedQuestionProtocol358(t *testing.T) {
	if !fixedModel358(t, "protocol-complete", fixedReads358, fixedOpen358, true) {
		t.Fatal("complete display protocol refused")
	}
	wrong := slices.Clone(fixedReads358)
	for i := range wrong {
		wrong[i]++
	}
	if fixedModel358(t, "protocol-implicit-read", wrong, fixedOpen358, true) {
		t.Fatal("Edit source acquisition accepted")
	}
	closed := slices.Clone(fixedOpen358)
	closed[3] = false
	if fixedModel358(t, "protocol-invalid-hidden", fixedReads358, closed, true) {
		t.Fatal("invalid raw draft hidden")
	}
}
func TestQuintFixedQuestionObserved358(t *testing.T) {
	path := os.Getenv("GOO_FIXED_QUESTION_358_REPORT")
	if path == "" {
		t.Skip("explicit native final report required")
	}
	raw, e := os.ReadFile(path)
	if e != nil {
		t.Fatal(e)
	}
	var l fixedLedger358
	if json.Unmarshal(raw, &l) != nil {
		t.Fatal("invalid native report")
	}
	if e = fixedAdmission358(l); e != nil {
		t.Fatal(e)
	}
	reads := make([]int, len(l.Protocol.Facts))
	opens := make([]bool, len(reads))
	base := l.Protocol.Facts[0].Before.Source
	for i, f := range l.Protocol.Facts {
		reads[i] = f.After.Source - base
		opens[i], _ = fixedExpanded358(f.After.Expanded)
	}
	if !fixedModel358(t, "actual", reads, opens, true) {
		t.Fatal("actual native model refused")
	}
	mutants := []func(*fixedLedger358){func(l *fixedLedger358) { l.Protocol.ValidRun.Request.FromMS++ }, func(l *fixedLedger358) { l.Protocol.Facts[0].After.Source++ }, func(l *fixedLedger358) { l.Protocol.Facts[3].After.RawDuration = "" }, func(l *fixedLedger358) { l.Protocol.Facts[3].After.Expanded = "false" }, func(l *fixedLedger358) { l.Protocol.Recovered.Response.Rows = nil }, func(l *fixedLedger358) { l.Protocol.OldEditor.After.Error += "changed" }}
	for i, mutate := range mutants {
		var fresh fixedLedger358
		json.Unmarshal(raw, &fresh)
		mutate(&fresh)
		if fixedAdmission358(fresh) == nil {
			t.Fatalf("negative%d admitted", i)
		}
		if fixedModel358(t, fmt.Sprintf("poison-%d", i), reads, opens, false) {
			t.Fatalf("negative%d model accepted", i)
		}
	}
}
