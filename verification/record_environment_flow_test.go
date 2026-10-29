package verification

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/identity"
	"github.com/casselc/observability/quintgo/qobs"
	"github.com/casselc/observability/quintgo/validate"
	"hegel.dev/go/hegel"
)

func environmentFlow250(mode, name string, current bool) exploration.Query {
	scope := identity.EnvironmentScope{Mode: mode}
	if mode == "named" {
		scope.Value = name
	}
	selection, _ := identity.EffectiveEnvironment(scope)
	if current {
		selection, _ = identity.CurrentEnvironment(name)
	}
	wire, _ := selection.Selector()
	return exploration.Query{Signal: "logs", EnvironmentSelector: &wire, FromMS: 100000, ToMS: 160000, Limit: 100}
}

func TestHegelRecordEnvironmentFlow250(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		mode := hegel.Draw(tc, hegel.SampledFrom([]string{"named", "empty", "unreported"}))
		name := hegel.Draw(tc, hegel.SampledFrom([]string{"prod", "unreported", "雪", "prod' OR 1=1 --"}))
		current := mode == "named" && hegel.Draw(tc, hegel.Booleans())
		origin := environmentFlow250(mode, name, current)
		selected, err := origin.EnvironmentSelection()
		if err != nil {
			tc.Fatal(err)
		}
		tail := exploration.TailSelector{EnvironmentSelector: origin.EnvironmentSelector}.Query(160000)
		fixed := (exploration.Query{Signal: "logs", FromMS: 100000, ToMS: 160000, Limit: 10}).WithEnvironmentFrom(tail)
		normalized := (exploration.CaptureQuestion{Kind: "explore", Explore: &fixed}).Normalized()
		for _, q := range []exploration.Query{tail, fixed, *normalized.Explore} {
			actual, e := q.EnvironmentSelection()
			if e != nil || !actual.Equal(selected) || !q.SameEnvironment(origin) || q.Validate() != nil {
				tc.Fatal("constructor lost valid selection", e)
			}
		}
		// Mutate each caller after construction: all three boundaries must detach.
		origin.EnvironmentSelector.Scope.Mode = "broken"
		if tail.EnvironmentSelector.Scope.Mode != mode {
			tc.Fatal("tail aliases origin")
		}
		tail.EnvironmentSelector.Policy = "broken"
		if fixed.EnvironmentSelector.Policy == "broken" {
			tc.Fatal("fixed aliases tail")
		}
		fixed.EnvironmentSelector.Scope.Mode = "broken"
		actual, e := normalized.Explore.EnvironmentSelection()
		if e != nil || !actual.Equal(selected) {
			tc.Fatal("normalized aliases fixed", e)
		}
		effective := environmentFlow250("named", name, false)
		legacy := exploration.Query{Environment: name}
		if effective.SameEnvironment(legacy) {
			tc.Fatal("effective/current-only policy conflated")
		}
		explicitCurrent := environmentFlow250("named", name, true)
		if !explicitCurrent.SameEnvironment(legacy) {
			tc.Fatal("legacy current-only meaning changed")
		}
	}, options(200)...)
}

func TestQuintObservedRecordEnvironmentFlow250(t *testing.T) {
	// The model consumes observations of actual Go transitions, not fabricated
	// model outputs. Negative controls corrupt a constructor's returned selector.
	spec, err := filepath.Abs("model/record_environment_flow.qnt")
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	manifest := "spec: " + spec + "\nmodule: record_environment_flow\nconstants: {INITIAL: '0'}\ninvariants: [safety]\nactions:\n"
	for _, action := range []string{"construct_tail", "fixed", "normalized"} {
		manifest += fmt.Sprintf("  %s:\n    expect: 'preserved == {{.obs.preserved}} and detached == {{.obs.detached}} and distinct == {{.obs.distinct}}'\n", action)
	}
	config := filepath.Join(dir, "flow.yaml")
	if err := os.WriteFile(config, []byte(manifest), 0600); err != nil {
		t.Fatal(err)
	}
	for _, mode := range []string{"named", "empty", "unreported"} {
		for _, bad := range []int{-1, 0, 1} {
			t.Run(fmt.Sprintf("%s/control%d", mode, bad), func(t *testing.T) {
				origin := environmentFlow250(mode, "prod", false)
				sink := &qobs.MemorySink{}
				rec := qobs.NewRecorder(sink)
				record := func(action string, q, parent exploration.Query) {
					rec.Record(context.Background(), action, nil,
						"obs.preserved", q.SameEnvironment(origin),
						"obs.detached", q.EnvironmentSelector != nil && q.EnvironmentSelector != parent.EnvironmentSelector,
						"obs.distinct", !q.SameEnvironment(exploration.Query{Environment: "prod"}))
				}
				tail := exploration.TailSelector{EnvironmentSelector: origin.EnvironmentSelector}.Query(160000)
				record("construct_tail", tail, origin)
				fixed := (exploration.Query{Signal: "logs"}).WithEnvironmentFrom(tail)
				if bad == 0 {
					fixed.EnvironmentSelector = nil
				}
				if bad == 1 {
					replacement, _ := identity.CurrentEnvironment("prod")
					wire, _ := replacement.Selector()
					fixed.EnvironmentSelector = &wire
				}
				record("fixed", fixed, tail)
				normalized := (exploration.CaptureQuestion{Kind: "explore", Explore: &fixed}).Normalized()
				record("normalized", *normalized.Explore, fixed)
				checker, err := validate.NewChecker(config, validate.Options{Quint: quintPath(t), Backend: "typescript", Dir: t.TempDir()})
				if err != nil {
					t.Fatal(err)
				}
				result, err := checker.Check(context.Background(), sink.Steps())
				if err != nil {
					t.Fatal(err)
				}
				if result.OK() != (bad == -1) {
					t.Fatalf("unexpected trace acceptance: control=%d accepted=%v", bad, result.OK())
				}
			})
		}
	}
}
