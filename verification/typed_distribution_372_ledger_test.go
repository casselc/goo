package verification

import (
	"context"
	"encoding/json"
	"math/big"
	"os"
	"reflect"
	"slices"
	"strconv"
	"testing"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/observability/quintgo/qobs"
	"github.com/casselc/observability/quintgo/validate"
)

type distributionCounters372 struct{ Source, SQL, Plans, Writes int }
type distributionFact372 struct {
	Name          string
	Before, After distributionCounters372
}
type distributionRecord372 struct {
	ID, TraceID, Service, Namespace string
	TimeMS                          int64
	Raw                             *string
	Env, Attributes                 map[string]string
}
type distributionFixture372 struct {
	Question     exploration.Query
	Binding      exploration.NumericBinding
	Records      []distributionRecord372
	Lower, Upper string
}
type distributionOperation372 struct {
	Request  exploration.Query
	Response exploration.Result
	Status   int
}
type distributionHeld372 struct {
	DOMID    int
	Text     string
	Snapshot json.RawMessage
}
type distributionCase372 struct {
	Fixture       distributionFixture372
	Parent, Child distributionOperation372
	Plan          struct {
		Request struct {
			Parent         exploration.Query
			FromBin, ToBin int
		}
		Response exploration.NumericHistogramSelection
		Status   int
	}
	HeldBefore, HeldAfter distributionHeld372
	Range                 struct {
		FromBin, ToBin int
		From, To       string
	}
	RawOracle struct{ ParentIDs, SelectedIDs []string }
	Canceled  *distributionOperation372
	Facts     []distributionFact372
}

func distributionMatch372(p *exploration.Predicate, r distributionRecord372, b exploration.NumericBinding) bool {
	if p == nil {
		return true
	}
	if len(p.All) > 0 {
		for i := range p.All {
			if !distributionMatch372(&p.All[i], r, b) {
				return false
			}
		}
		return true
	}
	if len(p.Any) > 0 {
		for i := range p.Any {
			if distributionMatch372(&p.Any[i], r, b) {
				return true
			}
		}
		return false
	}
	var raw *string
	if p.Numeric != nil {
		if !reflect.DeepEqual(*p.Numeric, b) {
			return false
		}
		raw = r.Raw
		n, state := decLiveScalar(raw, b)
		switch p.Op {
		case "valid", "missing", "empty", "invalid":
			return state == p.Op
		}
		v, vs := decLiveScalar(&p.Value, b)
		if state != "valid" || vs != "valid" {
			return false
		}
		c := n.Cmp(v)
		switch p.Op {
		case "eq":
			return c == 0
		case "ne":
			return c != 0
		case "lt":
			return c < 0
		case "lte":
			return c <= 0
		case "gt":
			return c > 0
		case "gte":
			return c >= 0
		}
		return false
	}
	if p.Field == nil {
		return false
	}
	if p.Field.Kind == "resource" {
		if p.Field.Key == "service.namespace" {
			raw = &r.Namespace
		} else if p.Field.Key == "service.name" {
			raw = &r.Service
		} else if v, ok := r.Env[p.Field.Key]; ok {
			raw = &v
		}
	}
	if p.Field.Kind == "attribute" {
		if v, ok := r.Attributes[p.Field.Key]; ok {
			raw = &v
		}
	}
	switch p.Op {
	case "missing":
		return raw == nil
	case "exists":
		return raw != nil
	case "eq":
		return raw != nil && *raw == p.Value
	case "ne":
		return raw != nil && *raw != p.Value
	}
	return false
}
func distributionMembers372(q exploration.Query, f distributionFixture372) []distributionRecord372 {
	var out []distributionRecord372
	for _, r := range f.Records {
		if !typedTriageEnvironment370(q, r.Env) || r.TimeMS < q.FromMS || r.TimeMS >= q.ToMS || !distributionMatch372(q.Where, r, f.Binding) {
			continue
		}
		out = append(out, r)
	}
	return out
}
func distributionIDs372(rows []distributionRecord372) []string {
	ids := make([]string, len(rows))
	for i, r := range rows {
		ids[i] = r.ID
	}
	slices.Sort(ids)
	return ids
}
func distributionResult372(op distributionOperation372, q exploration.Query, f distributionFixture372) bool {
	if op.Status != 200 || !nhLiveSameQuery(op.Request, q) || !nhLiveSameQuery(op.Response.Question, q) {
		return false
	}
	rows := distributionMembers372(q, f)
	if op.Response.Count != strconv.Itoa(len(rows)) {
		return false
	}
	got, e := reportedResultIDs353(op.Response)
	if e != nil || !slices.Equal(got, distributionIDs372(rows)) {
		return false
	}
	if op.Response.NumericHistogram == nil {
		return false
	}
	// Validate every independently constructed signed power-of-two bin, retaining
	// coefficients beyond binary64 precision and canonical decimal scale.
	edges := []*big.Int{}
	for i := 63; i >= 0; i-- {
		edges = append(edges, new(big.Int).Neg(new(big.Int).Lsh(big.NewInt(1), uint(i))))
	}
	edges = append(edges, big.NewInt(0))
	for i := 0; i <= 63; i++ {
		edges = append(edges, new(big.Int).Lsh(big.NewInt(1), uint(i)))
	}
	bins := op.Response.NumericHistogram.Bins
	if len(bins) != len(edges)-1 {
		return false
	}
	for i, bin := range bins {
		count := 0
		for _, r := range rows {
			n, state := decLiveScalar(r.Raw, f.Binding)
			if state == "valid" && n.Cmp(edges[i]) >= 0 && n.Cmp(edges[i+1]) < 0 {
				count++
			}
		}
		if bin.From != edges[i].String() || bin.To != edges[i+1].String() || bin.Count != strconv.Itoa(count) {
			return false
		}
	}
	return true
}
func distributionAdmission372(c distributionCase372, primary bool) bool {
	f := c.Fixture
	p := c.Plan.Response
	if c.Plan.Status != 200 || !nhLiveSameQuery(c.Plan.Request.Parent, f.Question) || !nhLiveSameQuery(p.Parent, f.Question) || p.From != f.Lower || p.To != f.Upper || p.FromBin != c.Range.FromBin || p.ToBin != c.Range.ToBin || !distributionResult372(c.Parent, f.Question, f) || !distributionResult372(c.Child, p.Selected, f) {
		return false
	}
	if !reflect.DeepEqual(p.Selected.Numeric, f.Question.Numeric) || !reflect.DeepEqual(p.Selected.NumericHistogram, f.Question.NumericHistogram) || !reflect.DeepEqual(p.Selected.EnvironmentSelector, f.Question.EnvironmentSelector) || p.Selected.FromMS != f.Question.FromMS || p.Selected.ToMS != f.Question.ToMS {
		return false
	}
	parent := distributionMembers372(f.Question, f)
	selected := distributionMembers372(p.Selected, f)
	rest := distributionMembers372(p.Rest, f)
	lo, ok := new(big.Int).SetString(f.Lower, 10)
	if !ok {
		return false
	}
	hi, ok := new(big.Int).SetString(f.Upper, 10)
	if !ok {
		return false
	}
	want := []distributionRecord372{}
	remaining := []distributionRecord372{}
	for _, r := range parent {
		n, state := decLiveScalar(r.Raw, f.Binding)
		if state == "valid" && n.Cmp(lo) >= 0 && n.Cmp(hi) < 0 {
			want = append(want, r)
		} else {
			remaining = append(remaining, r)
		}
	}
	oracleParent := slices.Clone(c.RawOracle.ParentIDs)
	oracleSelected := slices.Clone(c.RawOracle.SelectedIDs)
	slices.Sort(oracleParent)
	slices.Sort(oracleSelected)
	if !slices.Equal(distributionIDs372(parent), oracleParent) || !slices.Equal(distributionIDs372(want), oracleSelected) || !slices.Equal(distributionIDs372(selected), oracleSelected) || !slices.Equal(distributionIDs372(rest), distributionIDs372(remaining)) {
		return false
	}
	var held struct{ Revision string }
	if json.Unmarshal(c.HeldBefore.Snapshot, &held) != nil || held.Revision == "" || held.Revision != c.Parent.Response.Basis || c.HeldBefore.DOMID == 0 || c.HeldBefore.DOMID != c.HeldAfter.DOMID || savedJSON360(c.HeldBefore.Snapshot) != savedJSON360(c.HeldAfter.Snapshot) {
		return false
	}
	expected := []string{"parent-run", "brush", "child-stage", "child-run", "held-back", "retired-inspect-callback"}
	if primary {
		expected = append(expected, "explicit-parent-reactivation", "pending-brush", "pending-child-stage", "pending-child-run", "pending-held-back")
		if c.Canceled == nil || !distributionResult372(*c.Canceled, p.Selected, f) {
			return false
		}
	}
	if len(c.Facts) != len(expected) {
		return false
	}
	for i, fact := range c.Facts {
		if fact.Name != expected[i] || fact.Before.Writes != fact.After.Writes {
			return false
		}
		if i > 0 && fact.Before != c.Facts[i-1].After {
			return false
		}
		isRead := fact.Name == "parent-run" || fact.Name == "child-run" || fact.Name == "explicit-parent-reactivation" || fact.Name == "pending-child-run"
		if !isRead && fact.Before.SQL != fact.After.SQL {
			return false
		}
		plans := 0
		if fact.Name == "child-stage" || fact.Name == "pending-child-stage" {
			plans = 1
		}
		if fact.After.Plans-fact.Before.Plans != plans {
			return false
		}
	}
	return true
}
func distributionReplay372(t *testing.T, label string, facts []distributionFact372) bool {
	t.Helper()
	if len(facts) == 0 {
		t.Fatal("missing observed distribution facts")
	}
	sink := &qobs.MemorySink{}
	rec := qobs.NewRecorder(sink)
	baseline := facts[0].Before.Source
	depth := 0
	for _, f := range facts {
		action := "history"
		switch f.Name {
		case "parent-run", "child-run", "explicit-parent-reactivation", "pending-child-run":
			action = "read"
		case "child-stage", "pending-child-stage":
			action = "stage"
			depth++
		case "held-back", "pending-held-back":
			action = "back"
			depth--
		}
		rec.Record(context.Background(), action, nil, "obs.reads", f.After.Source-baseline, "obs.depth", depth, "obs.coherent", true)
	}
	raw, e := json.Marshal(sink.Steps())
	if e != nil {
		t.Fatal(e)
	}
	if e = artifact("typed-distribution372-"+label+"-steps.json", raw); e != nil {
		t.Fatal(e)
	}
	checker, e := validate.NewChecker("model/population_drilldown.yaml", validate.Options{Quint: quintPath(t), Backend: "typescript", Dir: t.TempDir()})
	if e != nil {
		t.Fatal(e)
	}
	report, e := checker.Check(t.Context(), sink.Steps())
	if e != nil {
		t.Fatal(e)
	}
	if e = artifact("typed-distribution372-"+label+"-model.txt", []byte(report.Summary())); e != nil {
		t.Fatal(e)
	}
	return report.OK()
}
func TestQuintTypedDistributionObserved372(t *testing.T) {
	path := os.Getenv("GOO_TYPED_DISTRIBUTION_372_REPORT")
	if path == "" {
		t.Skip("actual typed distribution journey opt-in")
	}
	raw, e := os.ReadFile(path)
	if e != nil {
		t.Fatal(e)
	}
	var r struct {
		Version string
		Errors  []string
		Cases   []distributionCase372
	}
	if json.Unmarshal(raw, &r) != nil || r.Version != "goo.typedDistributionPopulation372/1" || len(r.Errors) != 0 || len(r.Cases) != 2 {
		t.Fatal("missing typed distribution actual report")
	}
	for i, c := range r.Cases {
		if i == 0 && (c.Fixture.Question.Signal != "spans" || c.Fixture.Binding.Type != "int64" || c.Fixture.Binding.Field.Kind != "attribute") {
			t.Fatal("missing exact integer attribute case")
		}
		if i == 1 && (c.Fixture.Question.Signal != "logs" || c.Fixture.Binding.Type != "decimal64" || c.Fixture.Binding.Scale != 2 || c.Fixture.Binding.Field.Kind != "resource") {
			t.Fatal("missing qualified decimal resource case")
		}
		if !distributionAdmission372(c, i == 0) {
			t.Fatalf("case %d typed provenance/admission", i)
		}
		if !distributionReplay372(t, strconv.Itoa(i), c.Facts) {
			t.Fatal("actual distribution read lifecycle", i)
		}
	}
	c := r.Cases[0]
	read := slices.Clone(c.Facts)
	read[2].After.Source++
	if distributionReplay372(t, "bad-stage-read", read) {
		t.Fatal("stage-read mutant survived")
	}
	order := slices.Clone(c.Facts)
	order[2], order[4] = order[4], order[2]
	if distributionReplay372(t, "bad-return-order", order) {
		t.Fatal("return-before-stage mutant survived")
	}
	wrong := c
	wrong.Child.Request = c.Fixture.Question
	if distributionAdmission372(wrong, true) {
		t.Fatal("lost selected predicate mutant survived")
	}
}
