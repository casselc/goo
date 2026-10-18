package measurements

import (
	"encoding/json"
	"fmt"
	"github.com/casselc/goo/lifecycle"
	"reflect"
	"sort"
	"strconv"
	"time"
)

// Canonicalize source JSON before hashing. ClickHouse may escape slashes or use
// different number spellings; a browser round trip must preserve evidence identity.
// Exact timestamps and counts are strings in the source contract.
func canonicalRows(rows []json.RawMessage) ([]json.RawMessage, error) {
	if rows == nil {
		return nil, nil
	}
	out := make([]json.RawMessage, len(rows))
	for i, row := range rows {
		var value any
		if e := json.Unmarshal(row, &value); e != nil {
			return nil, e
		}
		b, e := json.Marshal(value)
		if e != nil {
			return nil, e
		}
		out[i] = b
	}
	return out, nil
}

// ValidateResult verifies stored evidence without querying telemetry or custody.
func ValidateResult(r *Result) error {
	if r == nil {
		return fmt.Errorf("missing measurement result")
	}
	e := r.Question.Validate()
	if e != nil || r.Plan.Revision != Hash(r.Question.Definition) || r.Plan.SQL == "" || len(r.Plan.SQL) > 65536 || len(r.Plan.Parameters) > 128 || len(r.Rows) > RowLimit || r.Count != strconv.Itoa(len(r.Rows)) {
		return fmt.Errorf("invalid stored measurement scope/plan/count")
	}
	at, e := time.Parse(time.RFC3339Nano, r.ReadAt)
	if e != nil {
		return e
	}
	basis := Hash(struct {
		Q    Question
		At   string
		Rows []json.RawMessage
	}{r.Question, r.ReadAt, r.Rows})
	canonicalInput, e := canonicalRows(r.Rows)
	if e != nil {
		return e
	}
	canonicalBasis := Hash(struct {
		Q    Question
		At   string
		Rows []json.RawMessage
	}{r.Question, r.ReadAt, canonicalInput})
	if basis != r.Basis && canonicalBasis != r.Basis {
		return fmt.Errorf("measurement basis mismatch")
	}
	if r.Question.Definition.Kind != "lifecycle" {
		expected, e := Evaluate(r.Question, r.Rows, at, "", nil)
		if expected != nil {
			expected.Plan = r.Plan
			expected.Rows = r.Rows
			expected.Basis = r.Basis
		} // Stored plans are evidence, not instructions to execute.
		if e != nil || !reflect.DeepEqual(expected, r) {
			return fmt.Errorf("stored metric replay mismatch: %v", e)
		}
		return nil
	}
	if r.Lifecycle == nil || r.Gauges != nil || r.Counters != nil || r.Histograms != nil || len(r.Events) != len(r.Rows) {
		return fmt.Errorf("invalid lifecycle evidence envelope")
	}
	expected, e := r.Lifecycle.Receipt.Policy.Prepare(r.Events)
	if e != nil || !reflect.DeepEqual(expected, r.Lifecycle) || r.State != expected.Receipt.State {
		return fmt.Errorf("lifecycle replay mismatch")
	}
	// Bind the normalized policy/events back to this package and raw source, not
	// merely to internally consistent caller-supplied facts.
	d := r.Question.Definition
	l := d.Lifecycle
	pcy := r.Lifecycle.Receipt.Policy
	outcomes := append([]string{}, l.Outcomes...)
	sort.Strings(outcomes)
	actualOutcomes := append([]string{}, pcy.TerminalOutcomes...)
	sort.Strings(actualOutcomes)
	scopes := []lifecycle.Scope{}
	for _, tenant := range l.Tenants {
		scopes = append(scopes, lifecycle.Scope{Environment: d.Environment, TenantID: tenant, ServiceName: l.LogicalService, Coverage: "unverified", EvidenceRef: r.Basis})
	}
	canonical := func(v any) string { b, _ := json.Marshal(v); return string(b) }
	scopeKeys := []string{}
	actualScopes := []string{}
	for _, s := range scopes {
		scopeKeys = append(scopeKeys, canonical(s))
	}
	for _, s := range pcy.Scopes {
		actualScopes = append(actualScopes, canonical(s))
	}
	sort.Strings(scopeKeys)
	sort.Strings(actualScopes)
	if pcy.APIVersion != "goo/lifecycle/v1alpha1" || pcy.SourceRevision != "clickhouse.otel-logs/goo-lifecycle-v1" || pcy.SubjectField != "subject_id" || !reflect.DeepEqual(scopeKeys, actualScopes) {
		return fmt.Errorf("lifecycle scope binding mismatch")
	}
	if pcy.EvaluationTimeNS != ns(r.Question.ToMS) || pcy.WindowNS != ns(r.Question.ToMS-r.Question.FromMS) || pcy.EvidenceCutoffNS != strconv.FormatInt(at.UnixNano()+1, 10) || pcy.DefinitionRevision != r.Plan.Revision || pcy.EvidenceBasis != r.Basis || pcy.EvaluationID != r.Basis || !pcy.KnowledgeAfterEvaluation || pcy.CustodyClock != "goo.local-first-observed/v1" || pcy.DeadlineNS != ns(l.DeadlineMS) || pcy.AllowanceNS != ns(l.AllowanceMS) || pcy.SuccessOutcome != l.Success || !reflect.DeepEqual(actualOutcomes, outcomes) {
		return fmt.Errorf("lifecycle policy binding mismatch")
	}
	for i, event := range r.Events {
		var row struct {
			Body       string            `json:"body"`
			Attributes map[string]string `json:"attributes"`
			Resource   map[string]string `json:"resource"`
			Service    string            `json:"service"`
		}
		if json.Unmarshal(r.Rows[i], &row) != nil {
			return fmt.Errorf("invalid lifecycle row")
		}
		serviceOK := false
		for _, s := range d.Source.Services {
			serviceOK = serviceOK || s == row.Service
		}
		if !serviceOK || row.Resource["deployment.environment.name"] != d.Environment || (row.Body != l.Accepted && row.Body != l.Terminal) {
			return fmt.Errorf("lifecycle source binding mismatch")
		}
		a := row.Attributes
		kind := "accepted"
		if row.Body == l.Terminal {
			kind = "terminal"
		}
		eventNS, eventErr := strconv.ParseInt(event.EventTimeNS, 10, 64)
		if eventErr == nil && eventNS > r.Question.ToMS*1000000 {
			return fmt.Errorf("business event exceeds fixed evaluation time")
		}
		received, e := strconv.ParseInt(event.ReceivedTimeNS, 10, 64)
		if e != nil || received > at.UnixNano() || event.EventID != a[l.EventID] || event.SubjectID != a[l.Subject] || event.TenantID != a[l.Tenant] || event.Environment != d.Environment || event.ServiceName != l.LogicalService || event.Kind != kind || event.EventTimeNS != a[l.EventTime] || event.Release != a[l.Release] || (kind == "terminal" && event.Outcome != a[l.Outcome]) {
			return fmt.Errorf("custody event binding mismatch")
		}
	}
	return nil
}
