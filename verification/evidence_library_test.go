package verification

import (
	"context"
	"encoding/json"
	"github.com/casselc/goo/retention"
	"github.com/casselc/goo/workbench"
	"github.com/casselc/observability/quintgo/qobs"
	"github.com/casselc/observability/quintgo/validate"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestQuintEvidenceLibraryLedger(t *testing.T) {
	name := os.Getenv("GOO_EVIDENCE_LIBRARY_LEDGER")
	if name == "" {
		t.Skip("observed browser ledger required")
	}
	raw, e := os.ReadFile(name)
	if e != nil {
		t.Fatal(e)
	}
	check := func(mutation int) bool {
		var l struct {
			ManifestJSON     string
			Initial, Current workbench.MeaningRevision
			Document         workbench.Document
			Catalog          retention.CatalogPage
			PendingCounts    []int
			InitialBody      struct{ RequestID, ResultToken string }
			CorrectionBody   struct{ RequestID string }
			Events           []struct {
				Action           string
				Reads, Published int
				Coherent         bool
			}
		}
		if e := json.Unmarshal(raw, &l); e != nil {
			t.Fatal(e)
		}
		var m retention.Manifest
		if e := json.Unmarshal([]byte(l.ManifestJSON), &m); e != nil {
			t.Fatal(e)
		}
		switch mutation {
		case 0:
			for i := range l.Events {
				if l.Events[i].Action == "inspect" {
					l.Events[i].Reads++
					break
				}
			}
		case 1:
			for i := range l.Events {
				if l.Events[i].Action == "retry" {
					l.Events[i].Published++
					break
				}
			}
		case 2:
			l.Document.Cards[0].Review = l.Current.ID
		case 3:
			l.Document.Cards[0].Reference = l.Initial.ID
		case 4:
			m.Members[0].Row = json.RawMessage(`{}`)
		case 5:
			l.PendingCounts[1] = 0
		case 6:
			l.Catalog.Items[0].Tags = l.Initial.Tags
		case 7:
			l.Current.Parent = ""
		}
		coherent := m.Validate() == nil && workbench.ValidateMeaningRevision(l.Initial) == nil && workbench.ValidateMeaningRevision(l.Current) == nil && l.Initial.Target.ID == m.ID && l.Current.Target.ID == m.ID && l.Current.Parent == l.Initial.ID && l.Initial.ID != l.Current.ID && m.Capture.ID == l.InitialBody.ResultToken && l.InitialBody.RequestID != l.CorrectionBody.RequestID && reflect.DeepEqual(l.PendingCounts, []int{1, 1, 0, 1, 1, 0})
		if len(l.Document.Cards) != 1 {
			coherent = false
		} else {
			card := l.Document.Cards[0]
			coherent = coherent && card.Kind == "retained" && card.Reference == m.ID && card.Review == l.Initial.ID && card.ReadBasis != ""
		}
		if len(l.Catalog.Items) != 1 {
			coherent = false
		} else {
			item := l.Catalog.Items[0]
			coherent = coherent && item.ID == m.ID && reflect.DeepEqual(item.Tags, l.Current.Tags)
		}
		dir := t.TempDir()
		if e := os.MkdirAll(filepath.Join(dir, "documents"), 0700); e != nil {
			t.Fatal(e)
		}
		docRaw, _ := json.Marshal(l.Document)
		if e := os.WriteFile(filepath.Join(dir, "documents", l.Document.ID+".json"), docRaw, 0600); e != nil {
			t.Fatal(e)
		}
		_, err := (workbench.Store{Dir: dir}).LoadDocument(l.Document.ID)
		coherent = coherent && err == nil
		sink := &qobs.MemorySink{}
		rec := qobs.NewRecorder(sink)
		for _, event := range l.Events {
			rec.Record(context.Background(), event.Action, nil, "obs.reads", event.Reads, "obs.published", event.Published, "obs.coherent", coherent && event.Coherent)
		}
		checker, e := validate.NewChecker("model/annotations.yaml", validate.Options{Quint: quintPath(t), Backend: "typescript", Dir: t.TempDir()})
		if e != nil {
			t.Fatal(e)
		}
		result, e := checker.Check(context.Background(), sink.Steps())
		if e != nil {
			t.Fatal(e)
		}
		return result.OK()
	}
	if !check(-1) {
		t.Fatal("observed library/recovery/citation trace rejected")
	}
	for i := 0; i < 8; i++ {
		if check(i) {
			t.Fatal("negative control accepted", i)
		}
	}
}
