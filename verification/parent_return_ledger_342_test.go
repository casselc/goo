package verification

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"reflect"
	"testing"

	"github.com/casselc/observability/quintgo/qobs"
	"github.com/casselc/observability/quintgo/validate"
)

func TestQuintParentReturnObserved342(t *testing.T) {
	name := os.Getenv("GOO_PARENT_RETURN_LEDGER")
	if name == "" {
		t.Skip("requires observed browser ledger")
	}
	raw, err := os.ReadFile(name)
	if err != nil {
		t.Fatal(err)
	}
	type fact struct {
		Name                                                           string
		Before, After                                                  map[string]any
		BeforeSource, AfterSource, BeforeComparisons, AfterComparisons int
		BeforeMetadata, AfterMetadata                                  int
	}
	type event struct {
		Action       string
		Depth, Reads int
	}
	type ledger struct {
		Events                                              []event
		Facts                                               []fact
		InitialSourceReads, TotalSourceRequests, FinalDepth int
		Calls                                               []struct{ Path, Method string }
	}
	source := map[string]bool{"/api/explore": true, "/api/explore/page": true, "/api/record-aggregates": true, "/api/explore/counted-values": true, "/api/tail": true}
	check := func(mutation int) bool {
		var l ledger
		decoder := json.NewDecoder(bytes.NewReader(raw))
		decoder.UseNumber()
		if err := decoder.Decode(&l); err != nil {
			t.Fatal(err)
		}
		if len(l.Events) == 0 || len(l.Facts) < 16 {
			return false
		}
		switch mutation {
		case 0:
			l.Events[len(l.Events)-1].Reads++
		case 1:
			l.Facts[0].After["definition"] = "changed returned definition"
		case 2:
			for i := range l.Events {
				if l.Events[i].Action == "back" {
					l.Events[i].Reads++
					break
				}
			}
		case 3:
			for i := range l.Facts {
				if l.Facts[i].Name == "late-metadata-response" {
					l.Facts[i].After["statuses"] = "late response adopted"
					break
				}
			}
		case 4:
			l.Facts[0].AfterSource++
		case 5:
			l.Events = l.Events[:1]
		case 6:
			for i := range l.Facts {
				if l.Facts[i].Name == "old-fresh-window-0" {
					l.Facts[i].After["receipts"] = "changed supplied receipt"
					break
				}
			}
		case 7:
			for i := range l.Facts {
				if l.Facts[i].Name == "blocked-facet-draft-continuation" {
					l.Facts[i].After["inputs"] = "discarded unapplied facet draft"
					break
				}
			}
		}
		sourceReads := 0
		for _, call := range l.Calls {
			if source[call.Path] {
				sourceReads++
			}
		}
		last := l.Events[len(l.Events)-1]
		coherent := sourceReads == l.TotalSourceRequests && l.InitialSourceReads >= 1 && last.Reads == l.TotalSourceRequests-l.InitialSourceReads && last.Reads == 27 && last.Depth == l.FinalDepth && last.Depth == 1 && last.Action == "history"
		required := map[string]bool{
			"back-nonauthored":                  false,
			"back-ordinary-0":                   false,
			"back-ordinary-1":                   false,
			"back-pending-rest":                 false,
			"back-temporary-comparison":         false,
			"back-rest-origin":                  false,
			"back-authored-pair-facet":          false,
			"old-nonauthored-group":             false,
			"old-selection-open-0":              false,
			"old-selection-open-1":              false,
			"late-metadata-response":            false,
			"old-catalogue-handler":             false,
			"old-include-comparison":            false,
			"blocked-facet-draft-continuation":  false,
			"blocked-window-draft-continuation": false,
			"old-fresh-window-0":                false,
		}
		seen := map[string]bool{}
		for _, f := range l.Facts {
			if seen[f.Name] {
				coherent = false
			}
			seen[f.Name] = true
			if _, ok := required[f.Name]; ok {
				required[f.Name] = true
			}
			// Supplied snapshots are raw authority, never a UI "retained" flag.
			// Empty arrays are valid where a surface has no such renderer.
			for _, state := range []map[string]any{f.Before, f.After} {
				receipts, ok := state["receipts"].(map[string]any)
				coherent = coherent && ok
				for _, key := range []string{"charts", "comparisons", "facets"} {
					_, exists := receipts[key].([]any)
					coherent = coherent && exists
				}
				_, hasInputs := state["inputs"].([]any)
				coherent = coherent && hasInputs
			}
			coherent = coherent && f.Name != "" && f.Before != nil && f.After != nil && reflect.DeepEqual(f.Before, f.After) && f.BeforeSource == f.AfterSource && f.BeforeComparisons == f.AfterComparisons && f.BeforeMetadata == f.AfterMetadata
		}
		for _, present := range required {
			coherent = coherent && present
		}
		if !coherent {
			return false
		}
		sink := &qobs.MemorySink{}
		rec := qobs.NewRecorder(sink)
		for _, e := range l.Events {
			rec.Record(context.Background(), e.Action, nil, "obs.reads", e.Reads, "obs.depth", e.Depth, "obs.coherent", coherent)
		}
		checker, err := validate.NewChecker("model/population_drilldown.yaml", validate.Options{Quint: quintPath(t), Backend: "typescript", Dir: t.TempDir()})
		if err != nil {
			t.Fatal(err)
		}
		result, err := checker.Check(context.Background(), sink.Steps())
		if err != nil {
			t.Fatal(err)
		}
		return result.OK()
	}
	if !check(-1) {
		t.Fatal("observed parent-return ledger rejected")
	}
	for mutation := 0; mutation < 8; mutation++ {
		if check(mutation) {
			t.Fatalf("mutation %d accepted", mutation)
		}
	}
}
