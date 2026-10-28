package measurements

import (
	"encoding/json"
	"fmt"
	"github.com/casselc/goo/internal/immutablefile"
	"github.com/casselc/goo/lifecycle"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

// ObserveLifecycle records first observations in custodyDir and evaluates the
// lifecycle at a fixed time. It may publish custody files, including before a
// later observation fails validation. Previous retains correction lineage.
func ObserveLifecycle(q Question, rows []json.RawMessage, at time.Time, custodyDir string, previous *Result) (*Result, error) {
	if q.Definition.Kind != "lifecycle" {
		return nil, fmt.Errorf("ObserveLifecycle requires a lifecycle definition")
	}
	r, err := prepareResult(q, rows, at)
	if err != nil {
		return nil, err
	}
	return evaluateLifecycle(r, at, custodyDir, previous)
}

// observe freezes first Goo observation of each normalized business payload. It
// is local custody, not warehouse ingestion, producer durability or completeness.
func observe(dir string, event lifecycle.Event, at time.Time) (lifecycle.Event, error) {
	if dir == "" {
		return event, fmt.Errorf("lifecycle requires persistent custody directory")
	}
	event.ReceivedTimeNS = ""
	id := Hash(event)
	if e := os.MkdirAll(dir, 0700); e != nil {
		return event, e
	}
	path := filepath.Join(dir, id+".json")
	load := func() (lifecycle.Event, error) {
		var prior lifecycle.Event
		b, e := os.ReadFile(path)
		if e != nil {
			return prior, e
		}
		e = json.Unmarshal(b, &prior)
		copy := prior
		copy.ReceivedTimeNS = ""
		if e != nil || Hash(copy) != id {
			return prior, fmt.Errorf("custody object integrity mismatch")
		}
		if _, e = strconv.ParseInt(prior.ReceivedTimeNS, 10, 64); e != nil {
			return prior, e
		}
		return prior, nil
	}
	if _, e := os.Stat(path); e == nil {
		return load()
	} else if !os.IsNotExist(e) {
		return event, e
	}
	event.ReceivedTimeNS = strconv.FormatInt(at.UnixNano(), 10)
	data, _ := json.Marshal(event)
	if _, e := immutablefile.PublishFileOnly(dir, id+".json", ".custody-", data); e != nil {
		return event, e
	}
	return load()
}
func evaluateLifecycle(r *Result, at time.Time, dir string, previous *Result) (*Result, error) {
	d := r.Question.Definition
	l := d.Lifecycle
	events := []lifecycle.Event{}
	for _, raw := range r.Rows {
		var row struct {
			Body       string            `json:"body"`
			Attributes map[string]string `json:"attributes"`
			Resource   map[string]string `json:"resource"`
			Service    string            `json:"service"`
		}
		if e := json.Unmarshal(raw, &row); e != nil {
			return nil, e
		}
		serviceOK := false
		for _, s := range d.Source.Services {
			serviceOK = serviceOK || s == row.Service
		}
		if !serviceOK || row.Resource["deployment.environment.name"] != d.Environment {
			return nil, fmt.Errorf("lifecycle source scope mismatch")
		}
		kind := ""
		if row.Body == l.Accepted {
			kind = "accepted"
		} else if row.Body == l.Terminal {
			kind = "terminal"
		} else {
			return nil, fmt.Errorf("lifecycle event binding mismatch")
		}
		a := row.Attributes
		// A later knowledge cutoff may reveal old business events, never future
		// business facts relative to the frozen evaluation time.
		if eventNS, err := strconv.ParseInt(a[l.EventTime], 10, 64); err == nil && eventNS > r.Question.ToMS*1000000 {
			return nil, fmt.Errorf("business event exceeds fixed evaluation time")
		}
		event := lifecycle.Event{EventID: a[l.EventID], Kind: kind, Environment: d.Environment, TenantID: a[l.Tenant], ServiceName: l.LogicalService, SubjectID: a[l.Subject], Release: a[l.Release], EventTimeNS: a[l.EventTime]}
		if kind == "terminal" {
			event.Outcome = a[l.Outcome]
		}
		var e error
		event, e = observe(dir, event, at)
		if e != nil {
			return nil, e
		}
		events = append(events, event)
	}
	scopes := []lifecycle.Scope{}
	for _, tenant := range l.Tenants {
		scopes = append(scopes, lifecycle.Scope{Environment: d.Environment, TenantID: tenant, ServiceName: l.LogicalService, Coverage: "unverified", EvidenceRef: r.Basis})
	}
	policy := lifecycle.Policy{APIVersion: "goo/lifecycle/v1alpha1", KnowledgeAfterEvaluation: true, EvaluationID: r.Basis, DefinitionRevision: r.Plan.Revision, SourceRevision: "clickhouse.otel-logs/goo-lifecycle-v1", EvidenceBasis: r.Basis, CustodyClock: "goo.local-first-observed/v1", EvaluationTimeNS: ns(r.Question.ToMS), EvidenceCutoffNS: strconv.FormatInt(at.UnixNano()+1, 10), DeadlineNS: ns(l.DeadlineMS), AllowanceNS: ns(l.AllowanceMS), WindowNS: ns(r.Question.ToMS - r.Question.FromMS), SubjectField: "subject_id", SuccessOutcome: l.Success, TerminalOutcomes: l.Outcomes, Scopes: scopes}
	if previous != nil && previous.Lifecycle != nil {
		policy.SupersedesEvaluationID = previous.Lifecycle.Receipt.Policy.EvaluationID
		if e := policy.ValidateCorrection(previous.Lifecycle.Receipt); e != nil {
			return nil, e
		}
	}
	var e error
	r.Events = events
	r.Lifecycle, e = policy.Prepare(events)
	if e != nil {
		return nil, e
	}
	r.State = r.Lifecycle.Receipt.State
	r.Warnings = append(r.Warnings, "Lifecycle absence stays unknown: local first-observation custody does not prove complete source delivery. Cohort/evaluation time is fixed; refresh may advance knowledge beyond that evaluation under the explicit policy.")
	return r, nil
}
