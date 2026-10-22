package verification

import (
	"context"
	"encoding/json"
	"github.com/casselc/goo/identity"
	"github.com/casselc/goo/retention"
	"github.com/casselc/goo/workbench"
	"github.com/casselc/observability/quintgo/qobs"
	"github.com/casselc/observability/quintgo/validate"
	"hegel.dev/go/hegel"
	"os"
	"reflect"
	"testing"
)

type meaningMachine struct {
	store     workbench.Store
	revisions []*workbench.MeaningRevision
	edges     map[string]bool
}

func (m *meaningMachine) write(parent string, withdraw bool) {
	v := workbench.MeaningInput{Target: workbench.MeaningTarget{Kind: "service", Source: "clickstack", Environment: "prod", Service: "api", Namespace: &identity.Namespace{Mode: "missing"}}, Parent: parent, Tags: []string{"review:latency"}, State: "active", Applies: &workbench.MeaningWindow{FromMS: 100, ToMS: 200}}
	if withdraw {
		v.State = "withdrawn"
		v.Tags = []string{}
	}
	r, e := m.store.CreateMeaning(v, nil)
	if e != nil {
		panic(e)
	}
	m.revisions = append(m.revisions, r)
	if parent != "" {
		m.edges[parent] = true
	}
}
func (m *meaningMachine) RuleRoot(tc hegel.TestCase) { m.write("", false) }
func (m *meaningMachine) RuleCorrectOldest(tc hegel.TestCase) {
	if len(m.revisions) > 0 {
		m.write(m.revisions[0].ID, false)
	}
}
func (m *meaningMachine) RuleCorrectNewest(tc hegel.TestCase) {
	if len(m.revisions) > 0 {
		m.write(m.revisions[len(m.revisions)-1].ID, false)
	}
}
func (m *meaningMachine) RuleWithdraw(tc hegel.TestCase) {
	if len(m.revisions) > 0 {
		m.write(m.revisions[len(m.revisions)-1].ID, true)
	}
}
func (m *meaningMachine) InvariantHistoryAndHeads(tc hegel.TestCase) {
	want := map[string]bool{}
	for _, r := range m.revisions {
		got, e := m.store.LoadMeaning(r.ID)
		if e != nil || !reflect.DeepEqual(r, got) {
			panic("immutable assertion changed")
		}
		if !m.edges[r.ID] && r.State == "active" {
			want[r.ID] = true
		}
	}
	page, e := m.store.Meanings(nil, "review:latency", "", false)
	if e != nil {
		panic(e)
	}
	if len(page.Items) != len(want) {
		panic("branch lost or superseded label leaked")
	}
	for _, item := range page.Items {
		if !want[item.Revision.ID] || !item.Head {
			panic("invalid active head")
		}
	}
}
func TestHegelMeaningStateSwarm(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		m := &meaningMachine{store: workbench.Store{Dir: t.TempDir()}, edges: map[string]bool{}}
		hegel.RunStateful(tc, m, hegel.WithStatefulStepCount(18), hegel.WithAlwaysCheckInvariants("InvariantHistoryAndHeads"))
	}, options(60)...)
}
func TestQuintAnnotationLedger(t *testing.T) {
	name := os.Getenv("GOO_ANNOTATION_LEDGER")
	if name == "" {
		t.Skip("observed browser ledger required")
	}
	raw, e := os.ReadFile(name)
	if e != nil {
		t.Fatal(e)
	}
	check := func(mutation int) bool {
		var l struct {
			ManifestJSON                                   string
			Service, Left, Right, Correction, ViewRevision *workbench.MeaningRevision
			Manifest                                       *retention.Manifest
			Revisions                                      *workbench.MeaningPage
			Events                                         []struct {
				Action           string
				Reads, Published int
				Coherent         bool
			}
		}
		if json.Unmarshal(raw, &l) != nil {
			t.Fatal("invalid ledger")
		}
		if l.ManifestJSON != "" {
			var retained retention.Manifest
			if json.Unmarshal([]byte(l.ManifestJSON), &retained) != nil {
				t.Fatal("invalid canonical manifest")
			}
			l.Manifest = &retained
		}
		if mutation == 3 {
			if l.Correction != nil {
				l.Correction.Applies.FromMS++
			} else if l.Manifest != nil {
				l.Manifest.Members[0].Row = json.RawMessage(`{}`)
			}
		}
		if mutation == 4 {
			if l.Left != nil {
				l.Left.Parent = l.Right.ID
			} else if l.Revisions != nil {
				l.Revisions.Items = l.Revisions.Items[1:]
			}
		}
		coherent := true
		if l.Service != nil {
			for _, r := range []*workbench.MeaningRevision{l.Service, l.Left, l.Right, l.Correction, l.ViewRevision} {
				coherent = coherent && r != nil && workbench.ValidateMeaningRevision(*r) == nil
			}
			coherent = coherent && l.Left.Parent == l.Service.ID && l.Right.Parent == l.Service.ID && l.Correction.Parent == l.Left.ID && reflect.DeepEqual(l.Correction.Target, l.Service.Target) && reflect.DeepEqual(l.Correction.Applies, l.Service.Applies) && l.Service.Target.Namespace != nil
		}
		if l.Manifest != nil {
			coherent = coherent && l.Manifest.Validate() == nil && l.Revisions != nil
			ids := map[string]bool{}
			parents := map[string]bool{}
			for _, item := range l.Revisions.Items {
				r := item.Revision
				coherent = coherent && workbench.ValidateMeaningRevision(*r) == nil && r.Target.Kind == "evidence" && r.Target.ID == l.Manifest.ID
				ids[r.ID] = true
				if r.Parent != "" {
					parents[r.Parent] = true
				}
			}
			for _, item := range l.Revisions.Items {
				r := item.Revision
				coherent = coherent && (r.Parent == "" || ids[r.Parent]) && item.Head == !parents[r.ID]
			}
		}
		if l.Service == nil && l.Manifest == nil {
			coherent = false
		}
		if mutation >= 0 {
			for i := range l.Events {
				if l.Events[i].Action == "inspect" {
					switch mutation {
					case 0:
						l.Events[i].Reads++
					case 1:
						l.Events[i].Published++
					case 2:
						l.Events[i].Coherent = false
					}
					break
				}
			}
		}
		sink := &qobs.MemorySink{}
		rec := qobs.NewRecorder(sink)
		for _, event := range l.Events {
			rec.Record(context.Background(), event.Action, nil, "obs.reads", event.Reads, "obs.published", event.Published, "obs.coherent", event.Coherent && coherent)
		}
		c, e := validate.NewChecker("model/annotations.yaml", validate.Options{Quint: quintPath(t), Backend: "typescript", Dir: t.TempDir()})
		if e != nil {
			t.Fatal(e)
		}
		r, e := c.Check(context.Background(), sink.Steps())
		if e != nil {
			t.Fatal(e)
		}
		return r.OK()
	}
	if !check(-1) {
		t.Fatal("observed trace rejected")
	}
	for i := 0; i < 5; i++ {
		if check(i) {
			t.Fatal("negative control accepted", i)
		}
	}
}
