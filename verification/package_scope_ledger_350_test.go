package verification

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/observability/quintgo/qobs"
	"github.com/casselc/observability/quintgo/validate"
)

type packageScopeFact350 struct {
	Name                          string `json:"name"`
	Before, After                 int
	MetadataBefore, MetadataAfter *int
}
type packageScopeMatrix350 struct {
	Key                                                string
	Plan, YAML                                         exploration.PackagePlan
	Observed                                           exploration.Result
	Status, SourceBefore, SourceAfterPlan, SourceAfter int
	Rejected                                           map[string]json.RawMessage
}
type packageScopeLedger350 struct {
	Fixture                           struct{ Question exploration.Query }
	StoredBefore, StoredAfter         string
	Version                           string
	Parent, Preview, FixedRun, Run    exploration.Result
	Draft, Validated, Reopened, Stage exploration.PackagePlan
	Saved                             exploration.SavedPackage
	EditedYAML, RetirementYAML        string
	Facts                             []packageScopeFact350
	Matrix                            []packageScopeMatrix350
	LatePreview                       struct {
		Request                  exploration.Query
		Response                 exploration.Result
		StatusAfter, OutputAfter string
	}
	OldCallback struct {
		SourceBefore, SourceAfter int
		Before, After             string
	}
	Errors []string
}

func packageScopeCheck350(l packageScopeLedger350) error {
	if l.Version != "goo.packageScope350/1" || len(l.Errors) != 0 {
		return fmt.Errorf("missing genuine ledger or browser errors")
	}
	if l.Parent.Count != "5" || l.Parent.Question.EnvironmentSelector == nil || l.Parent.Question.EnvironmentSelector.Policy != "effective" {
		return fmt.Errorf("primary typed effective population missing")
	}
	if !reflect.DeepEqual(l.Parent.Question.Where, l.Fixture.Question.Where) || l.Parent.Question.Summary != "span-duration" || !l.Parent.Question.SpanStatus || l.Parent.Question.BucketMS != 1000 || !l.Parent.Question.GroupTimeSeries || !reflect.DeepEqual(l.Parent.Question.GroupBy, []exploration.Field{{Kind: "column", Key: "service"}}) {
		return fmt.Errorf("primary literal predicates or full source profile missing")
	}
	expected, err := exploration.DraftTelemetryPackage(l.Parent.Question, l.Draft.Definition.Name, l.Draft.Definition.Title, l.Draft.Definition.LookbackMinutes)
	if err != nil || !reflect.DeepEqual(expected, l.Draft.Definition) {
		return fmt.Errorf("whole original question did not survive drafting: %v", err)
	}
	sameQuestion := func(a, b exploration.Query, clock bool) bool {
		if !clock {
			a.FromMS = b.FromMS
			a.ToMS = b.ToMS
		}
		return reflect.DeepEqual(a, b)
	}
	if !sameQuestion(l.Draft.Question, l.Parent.Question, true) || !reflect.DeepEqual(l.Validated.Definition, l.Draft.Definition) || !sameQuestion(l.Validated.Question, l.Draft.Question, true) {
		return fmt.Errorf("fixed definition validation changed query or profile")
	}
	parsed, err := exploration.ReadTelemetryPackage(l.Draft.YAML)
	if err != nil || !reflect.DeepEqual(parsed, l.Draft.Definition) {
		return fmt.Errorf("draft YAML differs from JSON definition")
	}
	revision, err := parsed.Revision()
	if err != nil || revision != l.Draft.Revision || l.Saved.Revision != revision || !reflect.DeepEqual(l.Saved.Definition, l.Draft.Definition) {
		return fmt.Errorf("saved definition or immutable revision differs")
	}
	var stored exploration.TelemetryPackage
	if l.StoredBefore == "" || l.StoredAfter != l.StoredBefore || json.Unmarshal([]byte(l.StoredBefore), &stored) != nil || !reflect.DeepEqual(stored, l.Saved.Definition) {
		return fmt.Errorf("stored revision bytes changed or differ from admitted definition")
	}
	if l.EditedYAML == l.Draft.YAML || !strings.Contains(l.EditedYAML, "Unsaved edited ") {
		return fmt.Errorf("actual unsaved edit missing")
	}
	for _, p := range []exploration.PackagePlan{l.Reopened, l.Stage} {
		if p.Revision != revision || !reflect.DeepEqual(p.Definition, l.Saved.Definition) || !sameQuestion(p.Question, l.Parent.Question, false) {
			return fmt.Errorf("reopen/stage lost immutable definition or whole profile")
		}
	}
	if l.Stage.Question.ToMS <= l.Parent.Question.ToMS || l.Stage.Question.ToMS-l.Stage.Question.FromMS != int64(l.Saved.Definition.LookbackMinutes)*60000 || l.Stage.Question.ToMS < l.Reopened.Question.ToMS {
		return fmt.Errorf("fresh stage does not own a later relative window")
	}
	for _, r := range []exploration.Result{l.Preview, l.FixedRun} {
		if r.Count != "5" || !sameQuestion(r.Question, l.Parent.Question, true) || r.CapturedAt == "" || r.Basis == "" {
			return fmt.Errorf("fixed preview/run evidence lost original scope")
		}
	}
	if l.Run.Count != "5" || !sameQuestion(l.Run.Question, l.Stage.Question, true) || l.Run.CapturedAt == "" || l.Run.Basis == "" {
		return fmt.Errorf("explicit fresh run differs from staged question")
	}
	lastClock, err := time.Parse(time.RFC3339Nano, l.Parent.CapturedAt)
	if err != nil {
		return fmt.Errorf("missing parent read clock")
	}
	for _, r := range []exploration.Result{l.Preview, l.FixedRun, l.Run} {
		clock, e := time.Parse(time.RFC3339Nano, r.CapturedAt)
		if e != nil || !clock.After(lastClock) {
			return fmt.Errorf("independent explicit read clock did not advance")
		}
		lastClock = clock
	}
	if !sameQuestion(l.LatePreview.Request, l.Draft.Question, true) || l.LatePreview.Response.Count != "5" || l.LatePreview.StatusAfter != "Unvalidated changes." || l.LatePreview.OutputAfter != "" || l.RetirementYAML == l.Draft.YAML {
		return fmt.Errorf("withdrawn preview published or lacked an actual response")
	}
	if l.OldCallback.Before == "" || l.OldCallback.Before != l.OldCallback.After || l.OldCallback.SourceAfter != l.OldCallback.SourceBefore {
		return fmt.Errorf("retired actual callback changed local state or read source")
	}
	var readyCallback struct{ PreviewDisabled *bool }
	if json.Unmarshal([]byte(l.OldCallback.Before), &readyCallback) != nil || readyCallback.PreviewDisabled == nil || *readyCallback.PreviewDisabled {
		return fmt.Errorf("retired callback proof did not use current enabled control")
	}
	required := map[string]int{"draft": 0, "validate": 0, "preview": 1, "retired-preview": 1, "save": 0, "fixed-stage": 0, "open-saved-editor": 0, "edit": 0, "reopen": 0, "fresh-stage": 0, "fresh-run": 1}
	indices := map[string]int{}
	for i, f := range l.Facts {
		delta, ok := required[f.Name]
		if !ok {
			return fmt.Errorf("unknown fact %s", f.Name)
		}
		if _, dup := indices[f.Name]; dup || f.Before < 0 || f.After-f.Before != delta {
			return fmt.Errorf("duplicate fact or hidden source read: %s", f.Name)
		}
		if f.Name == "fixed-stage" || f.Name == "fresh-stage" {
			want := 0
			if f.Name == "fresh-stage" {
				want = 1
			}
			if f.MetadataBefore == nil || f.MetadataAfter == nil || *f.MetadataAfter-*f.MetadataBefore != want {
				return fmt.Errorf("stage metadata clock boundary differs")
			}
		}
		indices[f.Name] = i
	}
	if len(indices) != len(required) {
		return fmt.Errorf("missing operation source boundary")
	}
	last := -1
	for _, name := range []string{"draft", "save", "edit", "reopen", "fresh-stage"} {
		if indices[name] <= last {
			return fmt.Errorf("actual lifecycle order differs")
		}
		last = indices[name]
	}
	counts := map[string]string{"current-only": "3", "effective": "5", "empty": "3", "unreported": "2"}
	seen := map[string]bool{}
	rejected := map[string]bool{}
	for _, m := range l.Matrix {
		want, ok := counts[m.Key]
		if !ok {
			if m.Key != "null-selector" && m.Key != "mixed" && m.Key != "missing" {
				return fmt.Errorf("unknown selector case")
			}
			if rejected[m.Key] || m.Status != 400 || len(m.Rejected) == 0 || m.SourceAfter != m.SourceBefore {
				return fmt.Errorf("invalid union issued source IO or lacks rejection")
			}
			rejected[m.Key] = true
			continue
		}
		if m.SourceAfterPlan != m.SourceBefore || m.SourceAfter <= m.SourceAfterPlan {
			return fmt.Errorf("metadata/source boundary mismatch")
		}
		if seen[m.Key] || m.Observed.Count != want || !reflect.DeepEqual(m.Plan.Definition, m.YAML.Definition) || m.Plan.Revision != m.YAML.Revision || !sameQuestion(m.Plan.Question, m.YAML.Question, true) || !sameQuestion(m.Observed.Question, m.Plan.Question, true) {
			return fmt.Errorf("JSON/YAML or genuine population mismatch: %s", m.Key)
		}
		seen[m.Key] = true
	}
	if len(rejected) != 3 {
		return fmt.Errorf("missing preIO union rejections")
	}
	if len(seen) != len(counts) {
		return fmt.Errorf("missing genuine selector population")
	}
	return nil
}

func TestQuintPackageScopeObserved350(t *testing.T) {
	name := os.Getenv("GOO_PACKAGE_SCOPE_LEDGER")
	if name == "" {
		t.Skip("native package scope ledger opt-in")
	}
	data, err := os.ReadFile(name)
	if err != nil {
		t.Fatal(err)
	}
	var ledger packageScopeLedger350
	if err = json.Unmarshal(data, &ledger); err != nil {
		t.Fatal(err)
	}
	check := func(label string, l packageScopeLedger350, want bool) {
		t.Helper()
		coherent := packageScopeCheck350(l) == nil
		sink := &qobs.MemorySink{}
		rec := qobs.NewRecorder(sink, qobs.WithActor("packages350"), qobs.WithProcess("actual-browser"))
		for _, action := range []string{"draft", "save", "edit", "reopen", "apply"} {
			rec.Record(context.Background(), action, nil, "obs.reusable", coherent, "obs.immutable", coherent)
		}
		checker, e := validate.NewChecker("model/telemetry_packages.yaml", validate.Options{Quint: quintPath(t), Backend: "typescript", Dir: t.TempDir()})
		if e != nil {
			t.Fatal(e)
		}
		report, e := checker.Check(context.Background(), sink.Steps())
		if e != nil {
			t.Fatal(e)
		}
		artifact("package-scope350-"+label+".txt", []byte(report.Summary()))
		if report.OK() != want {
			t.Fatalf("%s coherent=%v admission=%v: %s", label, coherent, packageScopeCheck350(l), report.Summary())
		}
	}
	check("actual", ledger, true)
	mutations := []func(*packageScopeLedger350){func(l *packageScopeLedger350) { l.Draft.Definition.Source.EnvironmentSelector.Policy = "current-only" }, func(l *packageScopeLedger350) { l.Stage.Question.EnvironmentSelector = nil }, func(l *packageScopeLedger350) { l.Stage.Question.Where = nil }, func(l *packageScopeLedger350) { l.Stage.Question.SpanStatus = false }, func(l *packageScopeLedger350) { l.Stage.Question.ToMS = l.Parent.Question.ToMS }, func(l *packageScopeLedger350) { l.Run.Question.FromMS++ }, func(l *packageScopeLedger350) {
		for i := range l.Facts {
			if l.Facts[i].Name == "fresh-stage" {
				l.Facts[i].After++
			}
		}
	}, func(l *packageScopeLedger350) { l.Saved.Definition.Title = "mutated saved revision" }, func(l *packageScopeLedger350) { l.LatePreview.OutputAfter = "published stale preview" }}
	for i, mutate := range mutations {
		var bad packageScopeLedger350
		if err = json.Unmarshal(data, &bad); err != nil {
			t.Fatal(err)
		}
		mutate(&bad)
		check(fmt.Sprintf("negative-%d", i), bad, false)
	}
}
