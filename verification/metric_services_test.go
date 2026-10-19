package verification

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/measurements"
	"github.com/casselc/observability/quintgo/qobs"
	"github.com/casselc/observability/quintgo/validate"
	"hegel.dev/go/hegel"
)

func TestHegelMetricServiceBounds(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		from := hegel.Draw(tc, hegel.Integers(1, 1000000))
		width := hegel.Draw(tc, hegel.Integers(1, 7200000))
		offset := hegel.Draw(tc, hegel.Integers(0, width-1))
		count := hegel.Draw(tc, hegel.Integers(1, 1000000))
		q := exploration.InventoryQuestion{Environment: "prod'\\n", FromMS: int64(from), ToMS: int64(from + width)}
		timestamp := from + offset
		bad := false
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			b, _ := io.ReadAll(r.Body)
			if strings.Contains(string(b), "WITH captured") {
				return
			}
			at := timestamp
			if bad {
				at = from + width
			}
			fmt.Fprintf(w, `{"service":"metrics-only","signal":"metrics","count":"%d","first_ms":"%d","last_ms":"%d"}`, count, at, at)
		}))
		defer server.Close()
		source := exploration.Source{Endpoint: server.URL}
		result, e := source.DiscoverServices(context.Background(), q)
		if e != nil || len(result.Signals) != 1 || len(result.Dependencies) != 0 || result.Question != q || result.Signals[0].Count != fmt.Sprint(count) || result.Signals[0].FirstMS != int64(timestamp) {
			tc.Fatal("metric signal scope/count changed", result, e)
		}
		bad = true
		if _, e = source.DiscoverServices(context.Background(), q); e == nil {
			tc.Fatal("exclusive upper bound accepted")
		}
	}, options(100)...)
}
func TestQuintMetricServicesLive(t *testing.T) {
	name := os.Getenv("GOO_METRIC_SERVICES_LEDGER")
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
		Discovery, BoundaryDiscovery exploration.ServiceDiscovery
		Browse                       []read
		Preview                      struct {
			Question   measurements.Question
			Definition measurements.Definition
		}
		Result                                                 measurements.Result
		BeforePreview, AfterPreview, BeforeReturn, AfterReturn int
	}
	check := func(mutation int) bool {
		var l ledger
		if json.Unmarshal(data, &l) != nil {
			t.Fatal("invalid ledger")
		}
		switch mutation {
		case 0:
			l.Discovery.Signals[0].Count = "5000"
		case 1:
			l.Preview.Question.ToMS++
		case 2:
			l.Preview.Definition.Source.Services = []string{"foreign"}
		case 3:
			l.AfterPreview++
		case 4:
			l.AfterReturn++
		case 5:
			l.Discovery.Dependencies = append(l.Discovery.Dependencies, exploration.Dependency{Parent: "fake", Child: "metrics-worker", Kind: "edge", Count: "1"})
		case 6:
			l.BoundaryDiscovery.Signals[0].LastMS = l.BoundaryDiscovery.Question.ToMS
		}
		sink := &qobs.MemorySink{}
		rec := qobs.NewRecorder(sink)
		record := func(a string, b bool) { rec.Record(context.Background(), a, nil, "obs.coherent", b) }
		d := l.Discovery
		q := d.Question
		b := l.BoundaryDiscovery
		coherent := len(d.Signals) == 2 && len(d.Dependencies) == 0 && d.Signals[0].Service == "metrics-worker" && d.Signals[0].Signal == "metrics" && d.Signals[0].Count == "4" && d.Signals[1].Count == "1" && len(b.Signals) == 1 && b.Signals[0].Count == "2" && b.Signals[0].FirstMS == b.Question.FromMS && b.Signals[0].LastMS == b.Question.ToMS-1
		record("discover", coherent)
		coherent = len(l.Browse) == 4
		metadata := false
		records := false
		for _, r := range l.Browse {
			var scope struct {
				Environment, Service string
				FromMS, ToMS         int64
			}
			json.Unmarshal(r.Question, &scope)
			coherent = coherent && scope.Environment == q.Environment && scope.FromMS == q.FromMS && scope.ToMS == q.ToMS && r.Status == "fulfilled"
			if r.Source == "/api/metric-discovery" {
				var result measurements.DiscoveryResult
				json.Unmarshal(r.Evidence, &result)
				metadata = scope.Service == "metrics-worker" && result.Records == 4 && len(result.Instruments) == 3
			}
			if r.Source == "/api/explore" {
				var result exploration.Result
				json.Unmarshal(r.Evidence, &result)
				records = result.Question.Signal == "spans" && result.Count == "0"
			}
		}
		record("browse", coherent && metadata && records)
		p := l.Preview.Question
		record("preview", p.Validate() == nil && p.Definition.Environment == q.Environment && p.FromMS == q.FromMS && p.ToMS == q.ToMS && reflect.DeepEqual(l.Preview.Definition.Source.Services, []string{"metrics-worker"}) && l.BeforePreview == l.AfterPreview)
		record("execute", measurements.ValidateResult(&l.Result) == nil && reflect.DeepEqual(l.Result.Question, p) && l.Result.Count == "2" && l.Result.Gauges != nil)
		record("returnLocal", l.BeforeReturn == l.AfterReturn)
		checker, e := validate.NewChecker("model/metric_services.yaml", validate.Options{Quint: quintPath(t), Backend: "typescript", Dir: t.TempDir()})
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
		t.Fatal("live metric service trace rejected")
	}
	for i := 0; i < 7; i++ {
		if check(i) {
			t.Fatal("mutated trace accepted", i)
		}
	}
}
