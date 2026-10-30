package verification

import (
	"context"
	"encoding/json"
	"github.com/casselc/goo/exploration"
	"github.com/casselc/observability/quintgo/qobs"
	"github.com/casselc/observability/quintgo/validate"
	"os"
	"strings"
	"testing"
)

func TestQuintKeptDiscoveryLedger(t *testing.T) {
	file := os.Getenv("GOO_KEPT_DISCOVERY_LEDGER")
	if file == "" {
		t.Skip("actual cross-browser ledger required")
	}
	raw, e := os.ReadFile(file)
	if e != nil {
		t.Fatal(e)
	}
	check := func(mutant int) bool {
		var l struct {
			Page         exploration.KeptSearchPage
			Calls        []struct{ Path, Method string }
			Repairs      int
			CrossBrowser bool
		}
		if e = json.Unmarshal(raw, &l); e != nil {
			t.Fatal(e)
		}
		switch mutant {
		case 0:
			l.Calls = append(l.Calls, struct{ Path, Method string }{"/api/explore", "POST"})
		case 1:
			l.Calls = append(l.Calls, struct{ Path, Method string }{"/api/kept-observations/records", "POST"})
		case 2:
			l.Page.Items[0].Origin = "future"
		case 3:
			l.Repairs++
		case 4:
			l.CrossBrowser = false
		}
		coherent := l.CrossBrowser && l.Page.Version == "goo/kept-search/v1" && l.Page.Total == 65 && len(l.Page.Items) == 20
		for _, item := range l.Page.Items {
			coherent = coherent && item.Validate() == nil
		}
		discoveries, exact, repairs, source, payload, actualRepairs := 0, 0, 0, 0, 0, 0
		for _, c := range l.Calls {
			if c.Path == "/api/kept-observations/repair" {
				actualRepairs++
			}
			if c.Method == "POST" && c.Path != "/api/kept-observations/repair" {
				if c.Path == "/api/explore" {
					source++
				} else {
					payload++
				}
			}
		}
		coherent = coherent && actualRepairs == l.Repairs
		sink := &qobs.MemorySink{}
		rec := qobs.NewRecorder(sink)
		for _, c := range l.Calls {
			action := ""
			if c.Path == "/api/kept-observations/search" {
				discoveries++
				action = "discover"
			} else if c.Path == "/api/kept-observations/repair" {
				repairs++
				action = "repair"
			} else if strings.HasPrefix(c.Path, "/api/kept-observations/") {
				exact++
				action = "open"
			}
			if action != "" {
				rec.Record(context.Background(), action, nil, "obs.source", source, "obs.payload", payload, "obs.discovery", discoveries, "obs.exact", exact, "obs.repair", repairs, "obs.coherent", coherent && l.Repairs == 3)
			}
		}
		checker, e := validate.NewChecker("model/kept_discovery.yaml", validate.Options{Quint: quintPath(t), Backend: "typescript", Dir: t.TempDir()})
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
		t.Fatal("actual discovery effect ledger rejected")
	}
	for m := 0; m < 5; m++ {
		if check(m) {
			t.Fatal("mutated discovery accepted", m)
		}
	}
}
