package workbench

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/identity"
)

func TestAllEnvironments437ExistingCarriersRetainTypedScope(t *testing.T) {
	wire, _ := identity.AllEnvironments().Selector()
	q := exploration.Query{Signal: "logs", EnvironmentSelector: &wire, FromMS: 1000, ToMS: 2000, Limit: 100, Where: &exploration.Predicate{Field: &exploration.Field{Kind: "column", Key: "body"}, Op: "contains", Value: "checkout"}}
	store := Store{Dir: t.TempDir()}
	view := ExplorationView{Version: 1, Title: "All log environments", Queries: []exploration.Query{q}, Layout: "records", Columns: []string{"utc", "service", "summary", "context"}}
	saved, err := store.CreateExplorationView(view)
	if err != nil {
		t.Fatal(err)
	}
	loaded, err := store.LoadExplorationView(saved.ID)
	if err != nil || !reflect.DeepEqual(loaded.Queries, view.Queries) {
		t.Fatal("saved view lost All scope", loaded, err)
	}
	activity := ActivityView{Version: 1, Title: "All current logs", Selector: ActivitySelector{Signal: "logs", EnvironmentSelector: &wire}, Columns: []string{"summary", "utc"}}
	kept, err := store.CreateActivityView(activity)
	if err != nil {
		t.Fatal(err)
	}
	reopened, err := store.LoadActivityView(kept.ID)
	if err != nil || !reflect.DeepEqual(reopened.Selector, activity.Selector) {
		t.Fatal("Activity definition lost All", reopened, err)
	}
	artifact := ViewArtifact{Version: ViewArtifactVersion, Namespace: "observability", Name: "all-logs", Revision: "one", View: view}
	encoded, err := EncodeViewArtifact(artifact)
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := DecodeViewArtifact(encoded.YAML)
	if err != nil || !reflect.DeepEqual(decoded.Artifact, artifact) {
		t.Fatal("YAML changed All into a label", err)
	}
	step := Step{Version: 1, Kind: "record", View: "Records", Scope: StepScope{EnvironmentSelector: &wire, FromMS: q.FromMS, ToMS: q.ToMS}, Record: &q}
	if err = step.Validate(); err != nil {
		t.Fatal(err)
	}
	raw, _ := json.Marshal(step)
	var copy Step
	if err = json.Unmarshal(raw, &copy); err != nil || copy.Validate() != nil || !copy.Record.SameEnvironment(q) {
		t.Fatal("typed question carrier lost All", err)
	}
	copy.Record.Signal = "spans"
	if err = copy.Validate(); err == nil {
		t.Fatal("record carrier bypassed unsupported signal gate")
	}
	view.Queries[0].Signal = "spans"
	if _, err = store.CreateExplorationView(view); err == nil {
		t.Fatal("saved view broadened All into spans")
	}
}
