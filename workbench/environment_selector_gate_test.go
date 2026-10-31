package workbench

import (
	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/identity"
	"testing"
)

func TestSelectorAuthoringBoundaries(t *testing.T) {
	s, _ := identity.EffectiveEnvironment(identity.EnvironmentScope{Mode: "unreported"})
	wire, _ := s.Selector()
	q := exploration.Query{Signal: "logs", EnvironmentSelector: &wire, FromMS: 100000, ToMS: 160000, Limit: 100}
	v := ExplorationView{Version: 1, Title: "Scope", Queries: []exploration.Query{q}, Layout: "records", Columns: []string{"utc", "summary"}}
	if err := v.Validate(); err != nil {
		t.Fatal("exploration view rejected typed selector", err)
	}
	a := ActivityView{Version: 1, Title: "Scope", Selector: exploration.TailSelector{EnvironmentSelector: &wire}, Columns: []string{"utc", "summary"}}
	if a.Validate() == nil {
		t.Fatal("Activity view accepted selector")
	}
	step := Step{Version: 1, Kind: "record", Record: &q, Scope: StepScope{FromMS: q.FromMS, ToMS: q.ToMS}, View: "Records"}
	if step.Validate() == nil {
		t.Fatal("authored step accepted selector")
	}
}
