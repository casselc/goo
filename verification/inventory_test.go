package verification

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/casselc/goo/exploration"
	"github.com/casselc/observability/quintgo/qobs"
	"github.com/casselc/observability/quintgo/validate"
	"hegel.dev/go/hegel"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"strconv"
	"testing"
)

type inventoryMachine struct {
	count, errors int
	buckets       [8]int
}

func (m *inventoryMachine) RuleObserve(tc hegel.TestCase) {
	i := hegel.Draw(tc, hegel.Integers(0, 7))
	m.count++
	m.buckets[i]++
	if hegel.Draw(tc, hegel.Integers(0, 1)) == 1 {
		m.errors++
	}
}
func (m *inventoryMachine) RuleReset(tc hegel.TestCase) {
	m.count = 0
	m.errors = 0
	m.buckets = [8]int{}
}
func (m *inventoryMachine) InvariantAccounting(tc hegel.TestCase) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if m.count == 0 {
			return
		}
		row := exploration.InventoryRow{Service: "service", Name: "work", Count: strconv.Itoa(m.count), Errors: strconv.Itoa(m.errors)}
		for _, n := range m.buckets {
			row.Buckets = append(row.Buckets, strconv.Itoa(n))
		}
		json.NewEncoder(w).Encode(row)
	}))
	defer server.Close()
	r, e := (exploration.Source{Endpoint: server.URL}).Inventory(context.Background(), exploration.InventoryQuestion{Environment: "prod", FromMS: 1000, ToMS: 100000})
	if e != nil || r.Count != strconv.Itoa(m.count) || (m.count == 0) != (r.State == "no_data") {
		panic(fmt.Sprint("inventory lost population: ", e))
	}
	if m.count > 0 && r.Rows[0].Errors != strconv.Itoa(m.errors) {
		panic("errors changed")
	}
}
func TestHegelInventoryStateSwarm(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		hegel.RunStateful(tc, &inventoryMachine{}, hegel.WithStatefulStepCount(20), hegel.WithAlwaysCheckInvariants("InvariantAccounting"))
	}, options(100)...)
}

type inventoryLedger struct {
	Capture  exploration.InventoryResult `json:"capture"`
	Filtered exploration.InventoryResult `json:"filtered"`
	Row      exploration.InventoryRow    `json:"row"`
	Open     exploration.Query           `json:"open"`
	Package  exploration.Query           `json:"package"`
}

func TestQuintInventoryLiveLedger(t *testing.T) {
	name := os.Getenv("GOO_INVENTORY_LEDGER")
	if name == "" {
		t.Skip("live browser ledger opt-in")
	}
	data, e := os.ReadFile(name)
	if e != nil {
		t.Fatal(e)
	}
	var l inventoryLedger
	if e = json.Unmarshal(data, &l); e != nil {
		t.Fatal(e)
	}
	check := func(l inventoryLedger) (bool, string) {
		sink := &qobs.MemorySink{}
		rec := qobs.NewRecorder(sink, qobs.WithActor("inventory"), qobs.WithProcess("live-browser"))
		sum := uint64(0)
		coherent := true
		for _, row := range l.Capture.Rows {
			n, e := strconv.ParseUint(row.Count, 10, 64)
			if e != nil {
				coherent = false
			}
			sum += n
			parts := uint64(0)
			for _, b := range row.Buckets {
				v, e := strconv.ParseUint(b, 10, 64)
				coherent = coherent && e == nil
				parts += v
			}
			errs, e := strconv.ParseUint(row.Errors, 10, 64)
			coherent = coherent && e == nil && parts == n && errs <= n
		}
		rec.Record(context.Background(), "capture", nil, "obs.coherent", coherent && strconv.FormatUint(sum, 10) == l.Capture.Count)
		rec.Record(context.Background(), "search", nil, "obs.coherent", reflect.DeepEqual(l.Capture, l.Filtered))
		expected := l.Capture.Question.Query()
		expected.GroupBy = nil
		expected.Where = &exploration.Predicate{All: []exploration.Predicate{{Field: &exploration.Field{Kind: "column", Key: "service"}, Op: "eq", Value: l.Row.Service}, {Field: &exploration.Field{Kind: "column", Key: "name"}, Op: "eq", Value: l.Row.Name}}}
		if n := exploration.NamespacePredicate(l.Capture.Question.Namespace); n != nil {
			expected.Where.All = append(expected.Where.All[:1], append([]exploration.Predicate{*n}, expected.Where.All[1:]...)...)
		}
		rec.Record(context.Background(), "open", nil, "obs.coherent", reflect.DeepEqual(expected, l.Open))
		rec.Record(context.Background(), "package", nil, "obs.coherent", l.Package.Signal == l.Open.Signal && l.Package.Environment == l.Open.Environment && reflect.DeepEqual(l.Package.Where, l.Open.Where) && len(l.Package.GroupBy) == 0)
		checker, e := validate.NewChecker("model/inventory.yaml", validate.Options{Quint: quintPath(t), Backend: "typescript", Dir: t.TempDir()})
		if e != nil {
			t.Fatal(e)
		}
		report, e := checker.Check(context.Background(), sink.Steps())
		if e != nil {
			t.Fatal(e)
		}
		return report.OK(), report.Summary()
	}
	ok, summary := check(l)
	artifact("inventory-live.txt", []byte(summary))
	if !ok {
		t.Fatal(summary)
	}
	for i := 0; i < 3; i++ {
		bad := l
		switch i {
		case 0:
			bad.Filtered.Count = "0"
		case 1:
			bad.Open.FromMS--
		case 2:
			bad.Package.Environment = "wrong"
		}
		ok, summary := check(bad)
		artifact(fmt.Sprintf("inventory-negative-%d.txt", i), []byte(summary))
		if ok {
			t.Fatal("negative control accepted", i)
		}
	}
}
