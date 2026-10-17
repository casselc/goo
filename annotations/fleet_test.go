package annotations

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFleetPackageMembership(t *testing.T) {
	for _, mode := range []string{"duplicate", "reference", "missing-selected", "view", "too-many"} {
		t.Run(mode, func(t *testing.T) {
			p, e := LoadOperationPackage("../config/operations/checkout.yaml")
			if e != nil {
				t.Fatal(e)
			}
			d := p.Definition
			switch mode {
			case "duplicate":
				d.Fleet.Members = []string{"woven-demo", "woven-demo"}
			case "reference":
				d.Fleet.Ratio = "missing"
			case "missing-selected":
				d.Fleet.Members = []string{"two", "three"}
			case "view":
				d.Fleet = nil
			case "too-many":
				d.Fleet.Members = []string{"woven-demo", "two", "three", "four", "five"}
			}
			if _, e := d.Bind(p.Manifest); e == nil {
				t.Fatal("accepted invalid fleet")
			}
		})
	}
}
func TestFleetPeerFailureKeepsSelectedRates(t *testing.T) {
	p, e := LoadOperationPackage("../config/operations/checkout.yaml")
	if e != nil {
		t.Fatal(e)
	}
	source := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("param_instance") == "woven-peer" {
			w.WriteHeader(503)
			return
		}
		if r.URL.Query().Get("readonly") != "1" || r.URL.Query().Get("max_result_rows") != "3001" {
			t.Error("source guards missing")
		}
		for _, o := range []string{"ok", "error", "panic"} {
			for _, point := range []struct{ end, value string }{{"100000000", "0"}, {"200000000", "1"}} {
				fmt.Fprintf(w, `{"resource":{"service.name":"lab-go-checkout","deployment.environment.name":"local","service.instance.id":"woven-demo"},"scope":"github.com/casselc/goo/instrumentation","version":"0.1.0","schema_url":"goo://instrumentation/v1","name":"goo.operation.calls","unit":"{invocation}","attributes":{"goo.operation.name":"checkout.reserve","goo.operation.role":"server","goo.operation.outcome":%q},"start_ns":"1000000","end_ns":%q,"value":%q,"temporality":2,"monotonic":true}`+"\n", o, point.end, point.value)
			}
		}
	}))
	defer source.Close()
	result, e := p.FetchRates(context.Background(), source.Client(), source.URL, "test", "test", 100, 200)
	if e != nil {
		t.Fatal(e)
	}
	if len(result.Rates) != 2 || result.Ratios[0].Result.Series[0].Intervals[0].Calls != "3" {
		t.Fatal("selected rates lost")
	}
	fleet := result.Fleet.Result
	if fleet.Members[0].State != "observed" || fleet.Members[1].State != "unavailable" {
		t.Fatal(fleet.Members)
	}
	i := fleet.Series[0].Intervals[0]
	if i.State != "unknown" || i.Percent != nil || i.Supported != 1 {
		t.Fatal(i)
	}
}
