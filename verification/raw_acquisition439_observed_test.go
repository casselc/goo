package verification

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/casselc/goo/measurements"
	"github.com/casselc/observability/quintgo/qobs"
	"github.com/casselc/observability/quintgo/validate"
)

type acquisition439API struct {
	Path, Method          string
	Status, ResponseBytes int
	Body, Response        json.RawMessage
	Cancelled             bool
}
type acquisition439SQL struct {
	Index                     int
	SQL                       string
	Status, ResponseBytes     int
	ReturnedRows              *int
	ElapsedMS                 float64
	Delayed, DownstreamClosed bool
}
type acquisition439Counters struct{ Source, Writes, SQL int }

type acquisition439Step struct {
	Before, After acquisition439Counters
	Name, Action  string
	SQL           []acquisition439SQL
	API           []acquisition439API
	RenderedRows  int
	ElapsedMS     float64
}
type acquisition439Bank struct {
	Name                       string
	Request                    measurements.RawGaugeQuestion
	Response                   measurements.RawGaugeResult
	ReadAt, Find, Page, Detail string
	Keys                       []string
}
type acquisition439Report struct {
	Version, Status          string
	Cleanup, DatabaseCleanup bool
	Errors                   []string
	Fixture                  struct {
		Token, Database, Service, Environment string
		Window                                struct{ FromMS, ToMS int64 }
		Points                                int
	}
	Steps        []acquisition439Step
	API          []acquisition439API
	SQL          []acquisition439SQL
	Banks        []acquisition439Bank
	ResponseHold struct {
		Status         int
		ReadAt         string
		FocusPreserved bool
		Declared       string
	}
}

// This acquisition replay does not interpret reported points as evaluated gauges.
// The existing model accounts only for source versus local authority; independent
// admission checks the fixed question, receipt, held display and meter ledger.
func TestQuintFirstUse439Observed(t *testing.T) {
	file := os.Getenv("GOO_FIRST_USE_439_REPORT")
	if file == "" {
		t.Skip("requires completed actual first-use acquisition browser report")
	}
	bytes, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	var report acquisition439Report
	if err = json.Unmarshal(bytes, &report); err != nil {
		t.Fatal(err)
	}
	if reason := acquisition439Admission(report); reason != "" {
		t.Fatal(reason)
	}
	checker, err := validate.NewChecker("model/population_drilldown.yaml", validate.Options{Quint: quintPath(t), Backend: "typescript", Dir: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	replay := func(r acquisition439Report) bool {
		sink := &qobs.MemorySink{}
		rec := qobs.NewRecorder(sink)
		reads := 0
		coherent := acquisition439Admission(r) == ""
		for _, step := range r.Steps {
			if len(step.SQL) == 0 {
				rec.Record(t.Context(), "history", nil, "obs.reads", reads, "obs.depth", 0, "obs.coherent", coherent)
			}
			for range step.SQL {
				reads++
				rec.Record(t.Context(), "read", nil, "obs.reads", reads, "obs.depth", 0, "obs.coherent", coherent)
			}
		}
		verdict, e := checker.Check(t.Context(), sink.Steps())
		if e != nil {
			t.Fatal(e)
		}
		return verdict.OK()
	}
	if !replay(report) {
		t.Fatal("actual acquisition replay rejected")
	}
	for mutant := range 7 {
		encoded, _ := json.Marshal(report)
		var bad acquisition439Report
		_ = json.Unmarshal(encoded, &bad)
		switch mutant {
		case 0:
			bad.Cleanup = false
		case 1:
			for i := range bad.Steps {
				if bad.Steps[i].Name == "local-find" {
					bad.Steps[i].SQL = append(bad.Steps[i].SQL, acquisition439SQL{SQL: "SELECT invented"})
				}
			}
		case 2:
			for i := range bad.Banks {
				if bad.Banks[i].Name == "after-return" {
					bad.Banks[i].ReadAt = "2000-01-01T00:00:00Z"
				}
			}
		case 3:
			for i := range bad.Banks {
				if bad.Banks[i].Name == "after-return" {
					bad.Banks[i].Request.Question.ToMS++
				}
			}
		case 4:
			for i := range bad.Banks {
				if bad.Banks[i].Name == "after-cancel" {
					bad.Banks[i].ReadAt = "2000-01-01T00:00:00Z"
				}
			}
		case 6:
			bad.ResponseHold.FocusPreserved = false
		case 5:
			for i := range bad.Steps {
				if bad.Steps[i].Name == "explicit-read" {
					bad.Steps[i].RenderedRows = 240
				}
			}
		}
		if replay(bad) {
			t.Fatalf("mutant %d accepted", mutant)
		}
	}
	t.Logf("admitted %d actual acquisition/local steps, %d source requests, %d reported points; seven controls rejected", len(report.Steps), len(report.SQL), report.Fixture.Points)
}

func acquisition439Admission(r acquisition439Report) string {
	if r.Version != "goo.firstUse439/1" || r.Status != "passed" || !r.Cleanup || !r.DatabaseCleanup || len(r.Errors) != 0 {
		return "unfinished acquisition run"
	}
	if r.Fixture.Points <= 100 || r.Fixture.Points > 10000 || r.Fixture.Token == "" || r.Fixture.Database != "goo439_"+r.Fixture.Token || r.Fixture.Window.FromMS <= 0 || r.Fixture.Window.ToMS <= r.Fixture.Window.FromMS {
		return "missing bounded cardinality fixture"
	}
	required := []string{"cold-entry", "scope-discovery", "stage-scope", "instrument-discovery", "explicit-read", "local-page", "local-find", "local-inspect", "retained-return", "explicit-reread", "cancel-pending-reread", "pending-departure", "late-success-ignored"}
	if len(r.Steps) != len(required) {
		return "incomplete phase ledger"
	}
	sourceCount := 0
	for i, s := range r.Steps {
		if s.Name != required[i] || s.ElapsedMS < 0 {
			return "unexpected acquisition phase"
		}
		local := s.Name == "stage-scope" || strings.HasPrefix(s.Name, "local-") || s.Name == "retained-return" || s.Name == "late-success-ignored"
		if s.Before.Writes != 0 || s.After.Writes != 0 || s.After.SQL-s.Before.SQL != len(s.SQL) || s.After.Source < s.Before.Source {
			return "source counter discontinuity or write authority"
		}
		if local {
			if s.Before != s.After {
				return "local phase started new source work"
			}
			if len(s.SQL) != 0 {
				return "local phase acquired source"
			}
			for _, call := range s.API {
				if call.Method == "POST" && s.Name != "late-success-ignored" {
					return "local phase sent mutation or source request"
				}
			}
		}
		if s.Name == "explicit-read" && (s.RenderedRows != 100 || len(s.SQL) == 0) {
			return "raw read does not distinguish source rows and bounded DOM"
		}
		if (s.Name == "scope-discovery" || s.Name == "instrument-discovery" || s.Name == "explicit-reread") && len(s.SQL) == 0 {
			return "explicit read phase missing source work"
		}
		for _, q := range s.SQL {
			if q.Index != sourceCount {
				return "source ledger gap or duplicate"
			}
			sourceCount++
			if q.Index >= len(r.SQL) || !reflect.DeepEqual(q, r.SQL[q.Index]) {
				return "phase/source meter mismatch"
			}
		}
	}
	if sourceCount != len(r.SQL) {
		return "unattributed source work"
	}
	delayed := false
	for _, q := range r.SQL {
		if q.ElapsedMS < 0 || q.ResponseBytes < 0 || q.Status < 200 {
			return "unfinished source measurement"
		}
		if q.Delayed {
			delayed = true
			if !q.DownstreamClosed {
				return "canceled source did not observe downstream closure"
			}
		}
		if strings.Contains(q.SQL, "otel_metrics_gauge") && q.Status == 200 && !q.Delayed && q.ReturnedRows == nil {
			return "gauge source row work unmeasured"
		}
	}
	if !delayed {
		return "missing actual canceled source read"
	}
	rawReads := 0
	for _, call := range r.API {
		if call.Method == "POST" && !slices.Contains([]string{"/api/metric-scopes", "/api/metric-discovery", "/api/metric-observations"}, call.Path) && !(strings.HasPrefix(call.Path, "/api/queries/") && strings.HasSuffix(call.Path, "/cancel") && len(strings.TrimSuffix(strings.TrimPrefix(call.Path, "/api/queries/"), "/cancel")) == 36) {
			return "acquisition acquired unrelated authority"
		}
		if call.Path != "/api/metric-observations" || call.Status != 200 || call.Cancelled {
			continue
		}
		var request measurements.RawGaugeQuestion
		var response measurements.RawGaugeResult
		if json.Unmarshal(call.Body, &request) != nil || request.Validate() != nil || json.Unmarshal(call.Response, &response) != nil {
			return "invalid successful raw acquisition"
		}
		if request.Question.FromMS != r.Fixture.Window.FromMS || request.Question.ToMS != r.Fixture.Window.ToMS || request.Question.Service != r.Fixture.Service {
			return "read changed fixed acquisition scope"
		}
		if !reflect.DeepEqual(response.Question, request.Question) || !reflect.DeepEqual(response.Instrument, request.Instrument) || len(response.Rows) != r.Fixture.Points || call.ResponseBytes <= 0 {
			return "raw acquisition lost scope, rows or bytes"
		}
		if _, err := time.Parse(time.RFC3339Nano, response.ReadAt); err != nil {
			return "raw acquisition lacks source clock"
		}
		rawReads++
	}
	if rawReads != 3 {
		return fmt.Sprintf("expected three successful raw responses including held obsolete read, got %d", rawReads)
	}
	banks := map[string]acquisition439Bank{}
	for _, b := range r.Banks {
		if _, ok := banks[b.Name]; ok {
			return "duplicate held bank"
		}
		banks[b.Name] = b
		if b.Request.Validate() != nil || b.ReadAt != b.Response.ReadAt || len(b.Keys) == 0 || len(b.Keys) > 100 {
			return "invalid held receipt"
		}
		known := map[string]bool{}
		for _, row := range b.Response.Rows {
			known[row.OccurrenceKey] = true
		}
		seen := map[string]bool{}
		for _, key := range b.Keys {
			if !known[key] || seen[key] {
				return "held display invented or duplicated occurrence"
			}
			seen[key] = true
		}
	}
	for _, name := range []string{"before-return", "after-return", "before-cancel", "after-cancel", "before-late", "after-late"} {
		if _, ok := banks[name]; !ok {
			return "missing held lifecycle bank"
		}
	}
	same := func(a, b acquisition439Bank) bool {
		return reflect.DeepEqual(a.Request, b.Request) && reflect.DeepEqual(a.Response, b.Response) && a.ReadAt == b.ReadAt && reflect.DeepEqual(a.Keys, b.Keys) && a.Find == b.Find && a.Page == b.Page && a.Detail == b.Detail
	}
	if !same(banks["before-return"], banks["after-return"]) {
		return "retained return replaced fixed held view"
	}
	if !same(banks["before-cancel"], banks["after-cancel"]) {
		return "canceled read replaced held view"
	}
	if !same(banks["before-late"], banks["after-late"]) {
		return "late successful response replaced held view"
	}
	if r.ResponseHold.Status != 200 || !r.ResponseHold.FocusPreserved || r.ResponseHold.Declared == "" || r.ResponseHold.ReadAt == "" || r.ResponseHold.ReadAt == banks["before-late"].ReadAt {
		return "successful obsolete response/focus not witnessed"
	}
	return ""
}
