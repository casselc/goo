// Package lifecycle prepares normalized immutable business events into subject
// facts at a frozen evaluation basis. Collector-specific adapters are separate.
package lifecycle

import (
	"fmt"
	"io"
	"math"
	"regexp"
	"strconv"

	"github.com/casselc/goo/measurement"
	"gopkg.in/yaml.v3"
)

// Nanoseconds are decimal strings at the JSON/YAML boundary. This preserves
// exact epoch values across clients whose JSON numbers are IEEE-754 doubles.
type Policy struct {
	KnowledgeAfterEvaluation bool     `yaml:"knowledgeAfterEvaluation,omitempty" json:"knowledgeAfterEvaluation,omitempty"`
	APIVersion               string   `yaml:"apiVersion" json:"apiVersion"`
	EvaluationID             string   `yaml:"evaluationId" json:"evaluationId"`
	DefinitionRevision       string   `yaml:"definitionRevision" json:"definitionRevision"`
	SourceRevision           string   `yaml:"sourceRevision" json:"sourceRevision"`
	EvidenceBasis            string   `yaml:"evidenceBasis" json:"evidenceBasis"`
	CustodyClock             string   `yaml:"custodyClock" json:"custodyClock"`
	SupersedesEvaluationID   string   `yaml:"supersedesEvaluationId,omitempty" json:"supersedesEvaluationId,omitempty"`
	EvaluationTimeNS         string   `yaml:"evaluationTimeNs" json:"evaluationTimeNs"`
	EvidenceCutoffNS         string   `yaml:"evidenceCutoffNs" json:"evidenceCutoffNs"`
	DeadlineNS               string   `yaml:"deadlineNs" json:"deadlineNs"`
	AllowanceNS              string   `yaml:"allowanceNs" json:"allowanceNs"`
	WindowNS                 string   `yaml:"windowNs" json:"windowNs"`
	SubjectField             string   `yaml:"subjectField" json:"subjectField"`
	SuccessOutcome           string   `yaml:"successOutcome" json:"successOutcome"`
	TerminalOutcomes         []string `yaml:"terminalOutcomes" json:"terminalOutcomes"`
	Scopes                   []Scope  `yaml:"scopes" json:"scopes"`
}

type Scope struct {
	Environment string `yaml:"environment" json:"environment"`
	TenantID    string `yaml:"tenantId" json:"tenantId"`
	ServiceName string `yaml:"serviceName" json:"serviceName"`
	Coverage    string `yaml:"coverage" json:"coverage"`
	EvidenceRef string `yaml:"evidenceRef,omitempty" json:"evidenceRef,omitempty"`
}

type Event struct {
	EventID        string `json:"event_id"`
	Kind           string `json:"kind"`
	Environment    string `json:"environment"`
	TenantID       string `json:"tenant_id"`
	ServiceName    string `json:"service_name"`
	SubjectID      string `json:"subject_id"`
	Release        string `json:"release,omitempty"`
	EventTimeNS    string `json:"event_time_ns"`
	ReceivedTimeNS string `json:"received_time_ns"`
	Outcome        string `json:"outcome,omitempty"`
}

type Subject struct {
	Environment    string   `json:"environment"`
	TenantID       string   `json:"tenant_id"`
	ServiceName    string   `json:"service_name"`
	SubjectID      string   `json:"subject_id"`
	Release        string   `json:"release"`
	AcceptedTimeNS string   `json:"accepted_time_ns,omitempty"`
	TerminalTimeNS string   `json:"terminal_time_ns,omitempty"`
	Classification string   `json:"classification"`
	Reason         string   `json:"reason"`
	InCohort       bool     `json:"in_cohort"`
	IsMature       bool     `json:"is_mature"`
	EventIDs       []string `json:"event_ids"`
}

type ScopeReceipt struct {
	Scope          Scope    `json:"scope"`
	Quality        string   `json:"quality"`
	Reasons        []string `json:"reasons"`
	InvalidRecords int      `json:"invalid_records"`
	OrphanSubjects int      `json:"orphan_subjects"`
}

type Receipt struct {
	Policy              Policy         `json:"policy"`
	CohortStartNS       string         `json:"cohort_start_ns"`
	CohortEndNS         string         `json:"cohort_end_ns"`
	SnapshotDigest      string         `json:"snapshot_digest"`
	EvidenceDigest      string         `json:"evidence_digest"`
	InputManifestDigest string         `json:"input_manifest_digest"`
	State               string         `json:"state"`
	VisibleDeliveries   int            `json:"visible_deliveries"`
	AfterCutoff         int            `json:"after_cutoff"`
	OutsideScope        int            `json:"outside_scope"`
	DuplicateDeliveries int            `json:"duplicate_deliveries"`
	ConflictingEventIDs int            `json:"conflicting_event_ids"`
	InvalidRecords      int            `json:"invalid_records"`
	Counts              map[string]int `json:"counts"`
	Scopes              []ScopeReceipt `json:"scopes"`
}

type Prepared struct {
	Receipt  Receipt            `json:"receipt"`
	Subjects []Subject          `json:"subjects"`
	Facts    []measurement.Fact `json:"facts"`
}

type Evaluation struct {
	Receipt Receipt              `json:"receipt"`
	Results []measurement.Result `json:"results"`
}

type bounds struct{ evaluation, cutoff, deadline, allowance, start, end int64 }

var decimal = regexp.MustCompile(`^-?[0-9]+$`)
var simpleName = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

func parseNS(s string) (int64, error) {
	if !decimal.MatchString(s) {
		return 0, fmt.Errorf("nanoseconds require an exact decimal integer string")
	}
	return strconv.ParseInt(s, 10, 64)
}

func subtract(a, b int64) (int64, error) {
	if b < 0 || a < math.MinInt64+b {
		return 0, fmt.Errorf("nanosecond interval overflows int64")
	}
	return a - b, nil
}

func ReadPolicy(r io.Reader) (*Policy, error) {
	dec := yaml.NewDecoder(r)
	dec.KnownFields(true)
	var p Policy
	if err := dec.Decode(&p); err != nil {
		return nil, err
	}
	var extra any
	if err := dec.Decode(&extra); err != io.EOF {
		return nil, fmt.Errorf("policy requires exactly one YAML document")
	}
	if _, err := p.validate(); err != nil {
		return nil, err
	}
	return &p, nil
}

func (p Policy) validate() (bounds, error) {
	b := bounds{}
	if p.APIVersion != "goo/lifecycle/v1alpha1" || p.EvaluationID == "" || p.DefinitionRevision == "" || p.SourceRevision == "" || p.EvidenceBasis == "" || p.CustodyClock == "" {
		return b, fmt.Errorf("require lifecycle API version, evaluation ID, revisions and evidence basis")
	}
	if p.SupersedesEvaluationID == p.EvaluationID {
		return b, fmt.Errorf("a correction cannot supersede itself")
	}
	reserved := map[string]bool{"evaluation_id": true, "environment": true, "tenant_id": true,
		"service_name": true, "release": true, "classification": true, "quality_impaired": true,
		"is_mature": true, "in_cohort": true}
	if !simpleName.MatchString(p.SubjectField) || reserved[p.SubjectField] {
		return b, fmt.Errorf("subjectField must be a non-reserved simple identifier")
	}
	for _, pair := range []struct {
		text string
		dest *int64
	}{{p.EvaluationTimeNS, &b.evaluation}, {p.EvidenceCutoffNS, &b.cutoff},
		{p.DeadlineNS, &b.deadline}, {p.AllowanceNS, &b.allowance}} {
		v, err := parseNS(pair.text)
		if err != nil {
			return b, err
		}
		*pair.dest = v
	}
	window, err := parseNS(p.WindowNS)
	if err != nil || window <= 0 || b.deadline <= 0 || b.allowance < 0 || b.cutoff > b.evaluation && !p.KnowledgeAfterEvaluation {
		return b, fmt.Errorf("require positive deadline/window, nonnegative allowance, cutoff <= evaluation unless later knowledge is explicit")
	}
	b.end, err = subtract(b.evaluation, b.deadline)
	if err == nil {
		b.end, err = subtract(b.end, b.allowance)
	}
	if err == nil {
		b.start, err = subtract(b.end, window)
	}
	if err != nil {
		return b, err
	}
	outcomes := map[string]bool{}
	for _, outcome := range p.TerminalOutcomes {
		if outcome == "" || outcomes[outcome] {
			return b, fmt.Errorf("terminal outcomes must be nonempty and unique")
		}
		outcomes[outcome] = true
	}
	if !outcomes[p.SuccessOutcome] {
		return b, fmt.Errorf("successOutcome must be a declared terminal outcome")
	}
	if len(p.Scopes) == 0 {
		return b, fmt.Errorf("at least one explicit scope is required, including for empty input")
	}
	seen := map[scopeKey]bool{}
	for _, s := range p.Scopes {
		k := scopeOf(s.Environment, s.TenantID, s.ServiceName)
		if s.Environment == "" || s.TenantID == "" || s.ServiceName == "" || seen[k] {
			return b, fmt.Errorf("scopes must have unique, nonempty environment/tenant/service identities")
		}
		if s.Coverage != "complete" && s.Coverage != "unverified" && s.Coverage != "impaired" {
			return b, fmt.Errorf("scope coverage must be complete, unverified or impaired")
		}
		if s.Coverage == "complete" && s.EvidenceRef == "" {
			return b, fmt.Errorf("complete coverage requires an external evidence reference; unsampled is insufficient")
		}
		seen[k] = true
	}
	return b, nil
}
