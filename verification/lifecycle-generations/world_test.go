package generations_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math"
	"reflect"
	"strings"

	"github.com/casselc/goo/lifecycle"
	"github.com/casselc/goo/subjectfacts"
	"github.com/chucklehead-dev/gotogether/qobs"
)

type operation struct {
	Action     string
	Tenant     int
	Successful bool
	Generation int
}
type counts struct{ Eligible, Good, Bad, Unknown int }
type frozen struct {
	value *lifecycle.Prepared
	bytes []byte
}
type world struct {
	events                             []lifecycle.Event
	snapshots                          []frozen
	phase, selected                    int
	aggregate                          counts
	mixedRejected, scopeRejected, done bool
}

func newWorld() *world {
	return &world{selected: -1, aggregate: counts{-1, -1, -1, -1}, mixedRejected: true, scopeRejected: true}
}
func policy(generation int) lifecycle.Policy {
	p := lifecycle.Policy{APIVersion: "goo/lifecycle/v1alpha1", EvaluationID: fmt.Sprint("g", generation),
		DefinitionRevision: "contract/v1", SourceRevision: "normalized/v1", EvidenceBasis: fmt.Sprint("basis", generation),
		CustodyClock: "bounded-input-ledger", EvaluationTimeNS: "100", EvidenceCutoffNS: "96", DeadlineNS: "10", AllowanceNS: "3", WindowNS: "30",
		SubjectField: "reservation_id", SuccessOutcome: "confirmed", TerminalOutcomes: []string{"confirmed", "rejected"},
		Scopes: []lifecycle.Scope{{Environment: "prod", TenantID: "a", ServiceName: "worker", Coverage: "complete", EvidenceRef: "synthetic-complete-ledger"},
			{Environment: "prod", TenantID: "b", ServiceName: "worker", Coverage: "complete", EvidenceRef: "synthetic-complete-ledger"}}}
	if generation == 1 {
		p.EvidenceCutoffNS = "100"
		p.SupersedesEvaluationID = "g0"
	}
	return p
}
func definition() *subjectfacts.Definition {
	fields := map[string]string{}
	for _, s := range []string{"evaluation_id", "environment", "tenant_id", "service_name", "reservation_id", "release", "classification"} {
		fields[s] = "string"
	}
	for _, s := range []string{"quality_impaired", "is_mature", "in_cohort"} {
		fields[s] = "bool"
	}
	count := func(field string, v any) subjectfacts.Measure {
		m := subjectfacts.Measure{Op: "count"}
		if field != "" {
			m.Where = &subjectfacts.Predicate{Field: field, Eq: v}
		}
		return m
	}
	return &subjectfacts.Definition{APIVersion: "goo/v1alpha1", Name: "scoped_generations",
		Source:     subjectfacts.Source{Kind: "subject-facts", Fields: fields, SubjectKey: []string{"environment", "tenant_id", "service_name", "reservation_id"}},
		Population: &subjectfacts.Predicate{Field: "in_cohort", Eq: true}, Dimensions: map[string]string{"tenant": "tenant_id"},
		Measures: map[string]subjectfacts.Measure{"eligible": count("", nil), "good": count("classification", "good"), "bad": count("classification", "bad"), "unknown": count("classification", "unknown"),
			"impaired": count("quality_impaired", true), "ratio": {Op: "ratio", Numerator: "good", Denominator: "eligible", UnknownWhen: []string{"unknown", "impaired"}}}}
}
func event(tenant int, kind string, successful bool, late bool) lifecycle.Event {
	e := lifecycle.Event{Environment: "prod", TenantID: []string{"a", "b"}[tenant], ServiceName: "worker", SubjectID: "shared", EventID: kind, Kind: kind, EventTimeNS: "60", ReceivedTimeNS: "61"}
	if kind == "accepted" {
		e.Release = "v1"
	} else {
		e.EventTimeNS = "70"
		e.ReceivedTimeNS = "95"
		e.Outcome = "rejected"
		if successful {
			e.Outcome = "confirmed"
		}
	}
	if late {
		e.ReceivedTimeNS = "97"
		if kind == "terminal" && !successful {
			e.ReceivedTimeNS = "98"
		}
	}
	return e
}
func copyPrepared(p *lifecycle.Prepared) (*lifecycle.Prepared, error) {
	b, err := json.Marshal(p)
	if err != nil {
		return nil, err
	}
	var out lifecycle.Prepared
	err = json.Unmarshal(b, &out)
	return &out, err
}
func (w *world) current() (*lifecycle.Prepared, error) {
	if w.phase == 0 {
		return policy(0).Prepare(w.events)
	}
	return w.snapshots[w.selected].value, nil
}
func selectedCounts(rows []subjectfacts.Result) (counts, error) {
	if len(rows) != 1 {
		return counts{}, fmt.Errorf("expected one ungrouped aggregate, got %d", len(rows))
	}
	out := counts{}
	for name, dest := range map[string]*int{"eligible": &out.Eligible, "good": &out.Good, "bad": &out.Bad, "unknown": &out.Unknown} {
		v := rows[0].Measures[name]
		if v.State != "ok" || v.Count == nil {
			return out, fmt.Errorf("invalid count %s: %+v", name, v)
		}
		*dest = int(*v.Count)
	}
	return out, nil
}
func (w *world) apply(ctx context.Context, rec *qobs.Recorder, op operation) error {
	switch op.Action {
	case "accept":
		w.events = append(w.events, event(op.Tenant, "accepted", false, w.phase > 0))
	case "terminal":
		w.events = append(w.events, event(op.Tenant, "terminal", op.Successful, w.phase > 0))
	case "duplicate":
		found := false
		for _, e := range w.events {
			if e.TenantID == []string{"a", "b"}[op.Tenant] {
				if w.phase > 0 {
					e.ReceivedTimeNS = "98"
				}
				w.events = append(w.events, e)
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("duplicate without a delivered event")
		}
	case "freeze", "correct":
		generation := 0
		if op.Action == "correct" {
			generation = 1
		}
		p := policy(generation)
		if generation == 1 {
			if err := p.ValidateCorrection(w.snapshots[0].value.Receipt); err != nil {
				return err
			}
		}
		prepared, err := p.Prepare(w.events)
		if err != nil {
			return err
		}
		b, err := json.Marshal(prepared)
		if err != nil {
			return err
		}
		w.snapshots = append(w.snapshots, frozen{prepared, b})
		w.phase = generation + 1
		w.selected = generation
	case "selectGeneration":
		w.selected = op.Generation
		var facts []subjectfacts.Fact
		for _, s := range w.snapshots {
			facts = append(facts, s.value.Facts...)
		}
		// The low-level evaluator explicitly selects one generation from coexisting
		// facts; Prepared.Evaluate has the stricter single-artifact contract.
		rows, err := definition().Evaluate(facts, policy(w.selected).EvaluationID, nil)
		if err != nil {
			return err
		}
		w.aggregate, err = selectedCounts(rows)
		if err != nil {
			return err
		}
	case "mixed":
		candidate, err := copyPrepared(w.snapshots[w.selected].value)
		if err != nil {
			return err
		}
		candidate.Facts = append(candidate.Facts, w.snapshots[1-w.selected].value.Facts...)
		_, err = candidate.Evaluate(definition(), nil)
		w.mixedRejected = err != nil
		if err != nil && !strings.Contains(err.Error(), "snapshots must not be mixed") {
			return fmt.Errorf("unexpected mixed-artifact error: %w", err)
		}
	case "changedScope":
		p := policy(1)
		p.Scopes[0].TenantID = "foreign"
		err := p.ValidateCorrection(w.snapshots[0].value.Receipt)
		w.scopeRejected = err != nil
		if err != nil && !strings.Contains(err.Error(), "correction changed subject scope") {
			return fmt.Errorf("unexpected correction error: %w", err)
		}
	case "finish":
		w.done = true
	case "idle":
	default:
		return fmt.Errorf("unknown action %q", op.Action)
	}
	return w.observe(ctx, rec, op)
}
func (w *world) observe(ctx context.Context, rec *qobs.Recorder, op operation) error {
	p, err := w.current()
	if err != nil {
		return err
	}
	classes, err := subjectClasses(p)
	if err != nil {
		return err
	}
	archived, replayStable := true, true
	for _, s := range w.snapshots {
		b, err := json.Marshal(s.value)
		if err != nil {
			return err
		}
		archived = archived && bytes.Equal(b, s.bytes)
	}
	if len(w.snapshots) > 0 {
		replay, err := policy(0).Prepare(w.events)
		if err != nil {
			return err
		}
		old := w.snapshots[0].value
		replayStable = reflect.DeepEqual(replay.Subjects, old.Subjects) && reflect.DeepEqual(replay.Facts, old.Facts) &&
			replay.Receipt.SnapshotDigest == old.Receipt.SnapshotDigest && replay.Receipt.EvidenceDigest == old.Receipt.EvidenceDigest
	}
	evaluated, err := p.Evaluate(definition(), nil)
	if err != nil {
		return err
	}
	if len(evaluated.Results) != 1 {
		return fmt.Errorf("expected one ratio result")
	}
	ratio := evaluated.Results[0].Measures["ratio"]
	ticks := -1
	if ratio.Ratio != nil {
		ticks = -99
		v := *ratio.Ratio * 2
		if !math.IsNaN(v) && !math.IsInf(v, 0) && v == math.Trunc(v) && v >= 0 && v <= 2 {
			ticks = int(v)
		}
	}
	args := []any{"obs.a", classes["a"], "obs.b", classes["b"], "obs.state", p.Receipt.State,
		"obs.eligible", p.Receipt.Counts["eligible"], "obs.good", p.Receipt.Counts["good"], "obs.bad", p.Receipt.Counts["bad"], "obs.unknown", p.Receipt.Counts["unknown"],
		"obs.phase", w.phase, "obs.selected", w.selected, "obs.aggregateEligible", w.aggregate.Eligible, "obs.aggregateGood", w.aggregate.Good, "obs.aggregateBad", w.aggregate.Bad, "obs.aggregateUnknown", w.aggregate.Unknown,
		"obs.mixedRejected", w.mixedRejected, "obs.scopeRejected", w.scopeRejected, "obs.archived", archived, "obs.replayStable", replayStable,
		"obs.ratioState", ratio.State, "obs.ratioTicks", ticks,
		"obs.evaluationID", p.Receipt.Policy.EvaluationID, "obs.evaluationTime", p.Receipt.Policy.EvaluationTimeNS,
		"obs.cutoff", p.Receipt.Policy.EvidenceCutoffNS, "obs.supersedes", p.Receipt.Policy.SupersedesEvaluationID,
		"obs.cohortStart", p.Receipt.CohortStartNS, "obs.cohortEnd", p.Receipt.CohortEndNS}
	for i, name := range []string{"original", "corrected"} {
		a, b, state := "absent", "absent", "no_data"
		if len(w.snapshots) > i {
			prepared := w.snapshots[i].value
			c, err := subjectClasses(prepared)
			if err != nil {
				return err
			}
			a, b, state = c["a"], c["b"], prepared.Receipt.State
		}
		args = append(args, "obs."+name+"A", a, "obs."+name+"B", b, "obs."+name+"State", state)
	}

	switch op.Action {
	case "accept", "duplicate":
		args = append(args, "tenant", op.Tenant)
	case "terminal":
		args = append(args, "tenant", op.Tenant, "successful", op.Successful)
	case "selectGeneration":
		args = append(args, "generation", op.Generation)
	}
	rec.Record(ctx, op.Action, nil, args...)
	return nil
}

func subjectClasses(p *lifecycle.Prepared) (map[string]string, error) {
	classes := map[string]string{"a": "absent", "b": "absent"}
	seen := map[string]bool{}
	for _, s := range p.Subjects {
		if s.Environment != "prod" || s.ServiceName != "worker" || s.SubjectID != "shared" || seen[s.TenantID] {
			return nil, fmt.Errorf("invalid/duplicate scoped subject: %+v", s)
		}
		if _, known := classes[s.TenantID]; !known {
			return nil, fmt.Errorf("unexpected subject tenant: %+v", s)
		}
		classes[s.TenantID] = s.Classification
		seen[s.TenantID] = true
	}
	return classes, nil
}
