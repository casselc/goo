package verification

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"reflect"
	"strconv"
	"testing"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/measurements"
	"github.com/casselc/observability/quintgo/qobs"
	"github.com/casselc/observability/quintgo/validate"
	"hegel.dev/go/hegel"
)

func TestHegelServiceSuggestionScope(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		from := hegel.Draw(tc, hegel.Integers(0, 1000000))
		width := hegel.Draw(tc, hegel.Integers(1, 7200000))
		alias := hegel.Draw(tc, hegel.Integers(0, 2))
		script := `const {suggestions}=require('../ui/service-browser.mjs');const [from,width,alias]=JSON.parse(process.argv[1]);const key=['db.system','db.system.name','unrelated'][alias];const scope={environment:'prod',fromMs:from,toMs:from+width};const s=suggestions(scope,"literal'\\n",[{kind:'attribute',key}])[1];process.stdout.write(JSON.stringify(s));`
		args, _ := json.Marshal([]int{from, width, alias})
		out, e := exec.Command("node", "-e", script, string(args)).Output()
		if e != nil {
			tc.Fatalf("actual JavaScript: %v", e)
		}
		var result struct {
			Available bool
			Question  *exploration.Query
		}
		if json.Unmarshal(out, &result) != nil {
			tc.Fatal("bad result")
		}
		if alias == 2 {
			if result.Available || result.Question != nil {
				tc.Fatal("unobserved convention offered")
			}
			return
		}
		q := result.Question
		if !result.Available || q == nil || q.Validate() != nil || q.FromMS != int64(from) || q.ToMS != int64(from+width) || q.Environment != "prod" || q.Where.All[0].Value != "literal'\\n" || q.Where.All[1].Op != "exists" {
			tc.Fatal("scope/convention mismatch")
		}
	}, options(100)...)
}
func TestQuintServiceBrowserLive(t *testing.T) {
	name := os.Getenv("GOO_SERVICE_BROWSER_LEDGER")
	if name == "" {
		t.Skip("live ledger opt-in")
	}
	data, e := os.ReadFile(name)
	if e != nil {
		t.Fatal(e)
	}
	type read struct {
		Source   string
		Question json.RawMessage
		Status   string
		Evidence json.RawMessage
	}
	type ledger struct {
		Catalog     exploration.ServiceDiscovery
		CatalogRows []struct {
			Service string
			Counts  []string
		}
		Inventory exploration.InventoryResult
		Browse    []read
		Drilled   exploration.Result
		Preview   struct {
			Question   measurements.Question
			Definition measurements.Definition
		}
		BeforeReturn, AfterReturn, BeforePreview, AfterPreview int
	}
	check := func(mutation int) bool {
		var l ledger
		if e := json.Unmarshal(data, &l); e != nil {
			t.Fatal(e)
		}
		switch mutation {
		case 0:
			l.Inventory.Question.Environment = "foreign"
		case 1:
			l.Drilled.Question.ToMS++
		case 2:
			l.Drilled.Question.Where.All[0].Value = "foreign"
		case 3:
			l.AfterReturn++
		case 4:
			l.AfterPreview++
		case 5:
			l.Preview.Question.FromMS++
		case 6:
			l.CatalogRows[0].Counts[0] = "9"
		case 7:
			l.Catalog.Question.ToMS++
		}
		sink := &qobs.MemorySink{}
		rec := qobs.NewRecorder(sink)
		record := func(a string, b bool) { rec.Record(context.Background(), a, nil, "obs.coherent", b) }
		q := l.Inventory.Question
		coherent := len(l.Browse) == 4 && len(l.CatalogRows) == 2 && l.CatalogRows[0].Service == "catalog" && reflect.DeepEqual(l.CatalogRows[0].Counts, []string{"6", "1", "2"}) && reflect.DeepEqual(l.CatalogRows[1].Counts, []string{"Not observed", "1", "Not observed"}) && l.Catalog.Question.Environment == q.Environment && l.Catalog.Question.FromMS == q.FromMS && l.Catalog.Question.ToMS == q.ToMS && q.Namespace != nil && q.Namespace.Mode == "missing"
		seen := map[string]bool{}
		for _, r := range l.Browse {
			var scope struct {
				Environment, Service string
				FromMS, ToMS         int64
			}
			json.Unmarshal(r.Question, &scope)
			coherent = coherent && scope.Environment == q.Environment && scope.FromMS == q.FromMS && scope.ToMS == q.ToMS && r.Status == "fulfilled"
			seen[r.Source] = true
			if r.Source == "/api/metric-discovery" {
				coherent = coherent && scope.Service == "catalog"
				var result measurements.DiscoveryResult
				json.Unmarshal(r.Evidence, &result)
				coherent = coherent && reflect.DeepEqual(result.Question, measurements.DiscoveryQuestion{Namespace: q.Namespace, Environment: q.Environment, Service: "catalog", FromMS: q.FromMS, ToMS: q.ToMS})
			}
		}
		record("browse", coherent && seen["/api/discovery"] && seen["/api/explore"] && seen["/api/metric-discovery"])
		d := l.Drilled.Question
		record("follow", d.Validate() == nil && d.Environment == q.Environment && d.FromMS == q.FromMS && d.ToMS == q.ToMS && l.Drilled.Count == "4" && len(d.Where.All) == 3 && d.Where.All[0].Field.Key == "service" && d.Where.All[0].Value == "catalog" && d.Where.All[1].Op == "missing" && d.Where.All[2].Op == "exists" && d.Where.All[2].Field.Key == "http.route" && len(d.GroupBy) == 1 && d.GroupBy[0].Key == "http.route")
		record("returnLocal", l.BeforeReturn == l.AfterReturn)
		p := l.Preview.Question
		record("preview", p.Validate() == nil && p.Definition.Environment == q.Environment && p.FromMS == q.FromMS && p.ToMS == q.ToMS && reflect.DeepEqual(l.Preview.Definition.Source.Services, []string{"catalog"}) && l.BeforePreview == l.AfterPreview)
		checker, e := validate.NewChecker("model/service_browser.yaml", validate.Options{Quint: quintPath(t), Backend: "typescript", Dir: t.TempDir()})
		if e != nil {
			t.Fatal(e)
		}
		r, e := checker.Check(context.Background(), sink.Steps())
		if e != nil {
			t.Fatal(e)
		}
		return r.OK()
	}
	if !check(-1) {
		t.Fatal("live service journey rejected")
	}
	for i := 0; i < 8; i++ {
		if check(i) {
			t.Fatal("mutated trace accepted", i)
		}
	}
}

func TestHegelServiceCatalogGrain(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		n := hegel.Draw(tc, hegel.Integers(1, 60))
		type signal struct {
			Service string `json:"service"`
			Signal  string `json:"signal"`
			Count   string `json:"count"`
			FirstMS int    `json:"firstMs"`
			LastMS  int    `json:"lastMs"`
		}
		input := []signal{}
		for i := 0; i < n; i++ {
			count := hegel.Draw(tc, hegel.Integers(1, 1000000))
			value := strconv.Itoa(count)
			if i%5 == 0 {
				value = "18446744073709551615"
			}
			input = append(input, signal{strconv.Itoa(i / 3), []string{"spans", "logs", "metrics"}[i%3], value, 100 + i, 1000 + i})
		}
		data, _ := json.Marshal(input)
		script := `const {catalogRows}=require('../ui/inventory.mjs');process.stdout.write(JSON.stringify(catalogRows(JSON.parse(process.argv[1]))));`
		out, e := exec.Command("node", "-e", script, string(data)).Output()
		if e != nil {
			tc.Fatal(e)
		}
		var rows []struct {
			Service         string
			Signals         map[string]signal
			FirstMS, LastMS int
		}
		if json.Unmarshal(out, &rows) != nil {
			tc.Fatal("invalid catalog")
		}
		byService := map[string]map[string]signal{}
		for _, r := range rows {
			byService[r.Service] = r.Signals
		}
		if len(rows) != (n+2)/3 {
			tc.Fatal("service duplication")
		}
		for _, s := range input {
			if !reflect.DeepEqual(byService[s.Service][s.Signal], s) {
				tc.Fatal("signal grain/precision changed")
			}
		}
	}, options(100)...)
}
