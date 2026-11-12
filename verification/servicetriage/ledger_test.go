package servicetriage

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/observability/quintgo/qobs"
	"github.com/casselc/observability/quintgo/validate"
)

type observedCounters struct{ Posts, SQL int }
type observedFact struct {
	Name          string
	Before, After observedCounters
}
type observedRead struct {
	Request  exploration.Query
	Response exploration.RecordAggregatesResult
	Status   int
}
type observedReport struct {
	Version string
	Scope   exploration.InventoryQuestion
	Reads   []observedRead
	Facts   []observedFact
	Errors  []string
}
type modelStep struct {
	Name                  string
	Reads, SQL, Adoptions int
}

func modelCheck(t *testing.T, steps []modelStep) bool {
	t.Helper()
	quint := os.Getenv("QUINT_BIN")
	if quint == "" {
		quint = "../../node_modules/.bin/quint"
	}
	quint, err := filepath.Abs(quint)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = os.Stat(quint); err != nil {
		t.Fatalf("Quint required, not skipped: %v", err)
	}
	sink := &qobs.MemorySink{}
	rec := qobs.NewRecorder(sink, qobs.WithActor(t.Name()))
	for _, step := range steps {
		rec.Record(context.Background(), step.Name, nil, "obs.reads", step.Reads, "obs.sql", step.SQL, "obs.adoptions", step.Adoptions)
	}
	checker, err := validate.NewChecker("binding.yaml", validate.Options{Quint: quint, Backend: "typescript", Dir: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	report, err := checker.Check(t.Context(), sink.Steps())
	if err != nil {
		t.Fatal(err)
	}
	return report.OK()
}
func TestQuintTriageReadAndRetirementProtocol(t *testing.T) {
	good := []modelStep{{"catalog", 0, 0, 0}, {"measure", 2, 4, 1}, {"sort", 2, 4, 1}, {"search", 2, 4, 1}, {"return_catalog", 2, 4, 1}, {"retire", 2, 4, 1}, {"late_result", 2, 4, 1}}
	if !modelCheck(t, good) {
		t.Fatal("valid read/local/retirement protocol refused")
	}
	for _, index := range []int{2, 3, 4} {
		bad := append([]modelStep(nil), good...)
		bad[index].Reads++
		bad[index].SQL += 2
		if modelCheck(t, bad) {
			t.Fatal("implicit local read accepted")
		}
	}
	late := append([]modelStep(nil), good...)
	late[len(late)-1].Adoptions++
	if modelCheck(t, late) {
		t.Fatal("retired response adoption accepted")
	}
	bad := append([]modelStep(nil), good...)
	bad[1].SQL = 8
	if modelCheck(t, bad) {
		t.Fatal("per-service fanout budget accepted")
	}
	if modelCheck(t, []modelStep{{"catalog", 0, 0, 0}, {"sort", 0, 0, 0}}) {
		t.Fatal("ranking before accepted measurement allowed")
	}
}
func admission(r observedReport) ([]modelStep, error) {
	if r.Version != "goo.serviceTriage/1" || len(r.Reads) != 2 || len(r.Errors) != 0 {
		return nil, fmt.Errorf("incomplete observed triage report")
	}
	current, prior := r.Reads[0], r.Reads[1]
	if current.Request.FromMS < prior.Request.FromMS {
		current, prior = prior, current
	}
	if current.Request.FromMS != r.Scope.FromMS || current.Request.ToMS != r.Scope.ToMS || prior.Request.ToMS != current.Request.FromMS || prior.Request.FromMS != current.Request.FromMS-(current.Request.ToMS-current.Request.FromMS) {
		return nil, fmt.Errorf("actual adjacent window scope changed")
	}
	for _, read := range []observedRead{current, prior} {
		q, dto := read.Request, read.Response
		if read.Status != 200 || !reflect.DeepEqual(q, dto.Question) || dto.State != "observed" && dto.State != "no_data" || dto.Version != exploration.RecordAggregatesVersion {
			return nil, fmt.Errorf("foreign or incoherent aggregate observation")
		}
		if err := exploration.ValidateRecordAggregatesQuery(q); err != nil {
			return nil, err
		}
		if !reflect.DeepEqual(q.GroupBy, []exploration.Field{{Kind: "column", Key: "service"}, {Kind: "resource", Key: "service.namespace"}}) || q.Summary != "span-duration" || !q.SpanStatus {
			return nil, fmt.Errorf("qualified server observation projection changed")
		}
		if q.Environment != r.Scope.Environment || !reflect.DeepEqual(q.EnvironmentSelector, r.Scope.EnvironmentSelector) {
			return nil, fmt.Errorf("environment policy changed")
		}
		role := exploration.Field{Kind: "column", Key: "kind"}
		expected := []exploration.Predicate{{Field: &role, Op: "eq", Value: "Server"}}
		if r.Scope.Namespace != nil {
			ns := exploration.Field{Kind: "resource", Key: "service.namespace"}
			p := exploration.Predicate{Field: &ns, Op: "missing"}
			if r.Scope.Namespace.Mode == "value" {
				p.Op = "eq"
				p.Value = r.Scope.Namespace.Value
			}
			expected = append(expected, p)
		}
		if !reflect.DeepEqual(q.Where, &exploration.Predicate{All: expected}) {
			return nil, fmt.Errorf("server role/namespace population changed")
		}
		if err := conserved(dto); err != nil {
			return nil, err
		}
	}
	p := prior.Request
	p.FromMS = current.Request.FromMS
	p.ToMS = current.Request.ToMS
	if !reflect.DeepEqual(current.Request, p) {
		return nil, fmt.Errorf("comparison predicate changed")
	}
	steps := []modelStep{{"catalog", 0, 0, 0}}
	names := map[string]bool{}
	reads, sql, adoptions := 0, 0, 0
	for _, fact := range r.Facts {
		action := fact.Name
		if action == "return" {
			action = "return_catalog"
		}
		if action != "measure" && action != "sort" && action != "search" && action != "return_catalog" {
			continue
		}
		if names[action] {
			return nil, fmt.Errorf("duplicate observed action %s", action)
		}
		names[action] = true
		posts := fact.After.Posts - fact.Before.Posts
		statements := fact.After.SQL - fact.Before.SQL
		wantPosts, wantSQL := 0, 0
		if action == "measure" {
			wantPosts, wantSQL = 2, 4
			adoptions = 1
		}
		if posts != wantPosts || statements != wantSQL {
			return nil, fmt.Errorf("%s source budget: posts%d sql%d", action, posts, statements)
		}
		reads += posts
		sql += statements
		steps = append(steps, modelStep{action, reads, sql, adoptions})
	}
	if len(names) != 4 {
		return nil, fmt.Errorf("measure/sort/search/return evidence required")
	}
	return steps, nil
}
func TestQuintObservedTriageLedger(t *testing.T) {
	path := os.Getenv("GOO_SERVICE_TRIAGE_REPORT")
	if path == "" {
		t.Skip("explicit actual browser report required; protocol unit test remains required")
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var r observedReport
	if err = json.Unmarshal(raw, &r); err != nil {
		t.Fatal(err)
	}
	steps, err := admission(r)
	if err != nil {
		t.Fatal(err)
	}
	if !modelCheck(t, steps) {
		t.Fatal("actual gesture/read ledger refused")
	}
	for _, mutate := range []func(*observedReport){func(r *observedReport) { r.Reads[0].Request.FromMS++ }, func(r *observedReport) { r.Reads[0].Response.State = "changed_between_reads" }, func(r *observedReport) { r.Reads[0].Request.GroupBy[1].Kind = "attribute" }, func(r *observedReport) {
		for i := range r.Facts {
			if r.Facts[i].Name == "sort" {
				r.Facts[i].After.SQL++
				return
			}
		}
	}} {
		var bad observedReport
		_ = json.Unmarshal(raw, &bad)
		mutate(&bad)
		if _, err = admission(bad); err == nil {
			t.Fatal("poisoned observed scope/read ledger accepted")
		}
	}
}

// Independent integer conservation over actual supplied aggregates. Complete
// populations cannot be made plausible by rounding, dropping or duplicating a group.
func conserved(dto exploration.RecordAggregatesResult) error {
	n, sum, errors, ok, unset, unknown := new(big.Int), new(big.Int), new(big.Int), new(big.Int), new(big.Int), new(big.Int)
	seen := map[string]bool{}
	if dto.Duration == nil || dto.SpanStatus == nil || len(dto.Groups) > 1000 {
		return fmt.Errorf("missing complete measures")
	}
	for _, raw := range dto.Groups {
		var carrier map[string]json.RawMessage
		if json.Unmarshal(raw, &carrier) != nil || carrier["g0"] == nil || carrier["g1"] == nil {
			return fmt.Errorf("missing qualified dimensions")
		}
		var g group
		if json.Unmarshal(raw, &g) != nil {
			return fmt.Errorf("invalid group carrier")
		}
		k := key(g.G0, g.G1)
		if seen[k] {
			return fmt.Errorf("duplicate qualified service")
		}
		seen[k] = true
		values := []string{g.Count, g.Duration.Sum, g.Status.Error, g.Status.OK, g.Status.Unset, g.Status.Unknown}
		ints := make([]*big.Int, len(values))
		for i, v := range values {
			var valid bool
			ints[i], valid = new(big.Int).SetString(v, 10)
			if !valid || ints[i].Sign() < 0 || ints[i].String() != v {
				return fmt.Errorf("invalid exact group measure")
			}
		}
		if ints[0].Sign() == 0 || g.Duration.Count != g.Count || g.Status.Count != g.Count {
			return fmt.Errorf("group denominator")
		}
		classes := new(big.Int)
		for _, v := range ints[2:] {
			classes.Add(classes, v)
		}
		if classes.Cmp(ints[0]) != 0 {
			return fmt.Errorf("group status conservation")
		}
		for i, total := range []*big.Int{n, sum, errors, ok, unset, unknown} {
			total.Add(total, ints[i])
		}
	}
	if dto.Count != n.String() || dto.Duration.Count != dto.Count || dto.Duration.SumNS != sum.String() || dto.SpanStatus.Count != dto.Count || dto.SpanStatus.Error != errors.String() || dto.SpanStatus.OK != ok.String() || dto.SpanStatus.Unset != unset.String() || dto.SpanStatus.Unknown != unknown.String() {
		return fmt.Errorf("total/group conservation")
	}
	return nil
}
