package verification

import (
	"encoding/json"
	"os"
	"reflect"
	"testing"

	"github.com/casselc/goo/exploration"
)

func TestHistogramViewActorObserved407(t *testing.T) {
	mainPath, actorPath := os.Getenv("GOO_HISTOGRAM_407_REPORT"), os.Getenv("GOO_HISTOGRAM_407_ACTOR_REPORT")
	if mainPath == "" || actorPath == "" {
		t.Skip("actual407 main and actor reports required")
	}
	var main histogramLedger407
	wire, e := os.ReadFile(mainPath)
	if e != nil {
		t.Fatal(e)
	}
	if e = json.Unmarshal(wire, &main); e != nil {
		t.Fatal(e)
	}
	if e = admitHistogram407(main); e != nil {
		t.Fatal(e)
	}
	var actor struct {
		Version, SourceBank string
		Question            exploration.Query
		Setup               []customizeSaved394
		SQLBefore, SQLAfter int
		ProxyCalls          []literalWire393
		Actor               struct {
			Snapshots int
			Phases    []string
			Complete  struct{ Phase string }
			Observed  struct {
				Phase, Population string
				Calls             []struct {
					Path, Method      string
					Status            int
					Request, Response json.RawMessage
				}
				Histogram histogramFrame407
			}
		}
	}
	wire, e = os.ReadFile(actorPath)
	if e != nil {
		t.Fatal(e)
	}
	if e = json.Unmarshal(wire, &actor); e != nil {
		t.Fatal(e)
	}
	c := main.Cases[0]
	fail := func(s string) { t.Fatalf("histogram407 actor: %s", s) }
	if actor.Version != "goo.histogram407ActorReport/1" || actor.SourceBank != main.Fixture.Token || !sameTransport384(actor.Question, c.DestinationQuestion) || len(actor.Setup) != 2 || actor.Actor.Snapshots < 6 || !reflect.DeepEqual(actor.Actor.Phases, []string{"initial", "library", "preview", "applied", "read", "complete"}) || actor.Actor.Complete.Phase != "complete" || actor.Actor.Observed.Phase != "read" || actor.Actor.Observed.Population != "applied" {
		fail("incomplete actor/source association")
	}
	for _, saved := range actor.Setup {
		if e := templateStored404(saved); e != nil {
			t.Fatal(e)
		}
	}
	var read customizeRead394
	var expansion templateExpansion404
	reads, expands := 0, 0
	for _, call := range actor.Actor.Observed.Calls {
		if call.Method != "POST" {
			continue
		}
		switch call.Path {
		case "/api/explore":
			reads++
			read.Status = call.Status
			if json.Unmarshal(call.Request, &read.Request) != nil || json.Unmarshal(call.Response, &read.Response) != nil {
				fail("malformed source read")
			}
		case "/api/view-templates/expand":
			expands++
			expansion.Status = call.Status
			if json.Unmarshal(call.Request, &expansion.Request) != nil || json.Unmarshal(call.Response, &expansion.Response) != nil {
				fail("malformed expansion")
			}
		default:
			fail("unauthorized actor write/plan/source")
		}
	}
	if reads != 1 || expands != 1 || !sameTransport384(read.Request, c.DestinationQuestion) || expansion.Status != 200 || !sameTransport384(expansion.Request.Query, c.DestinationQuestion) || !sameTransport384(expansion.Response.View.Queries, []exploration.Query{c.DestinationQuestion}) || !sameTransport384(expansion.Response.View.Composition, c.Saved.Response.Composition) {
		fail("destination or presentation changed")
	}
	if e := histogramRead407(read, main.Raw.Spans); e != nil {
		t.Fatal(e)
	}
	if !sameTransport384(actor.Actor.Observed.Histogram.Snapshot, read.Response.HistogramDataset) {
		fail("DOM bins differ from source")
	}
	var display struct {
		Kind             string
		RequiresComplete bool
	}
	if json.Unmarshal(actor.Actor.Observed.Histogram.Display, &display) != nil || display.Kind != "table" || !display.RequiresComplete {
		fail("saved table/complete policy changed")
	}
	if actor.SQLBefore != 0 || actor.SQLAfter != len(actor.ProxyCalls) {
		fail("actor source effect budget changed")
	}
	if e := summaryPlans400([]customizeRead394{read}, actor.ProxyCalls); e != nil {
		t.Fatal(e)
	}
}
