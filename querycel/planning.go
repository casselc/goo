package querycel

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"unicode"
	"unicode/utf8"

	"github.com/casselc/goo/exploration"
)

const (
	PlanVersion     = "goo.cel-plan/1"
	CompilerVersion = "goo.record-cel/1"
	RecordOperation = "record.refine"
	RecordTarget    = "clickhouse"
	RecordProfile   = "fixed-record/1"
)

// PlanBinding identifies one check in an owning editor. These are correlation
// tokens, not authorization or live readiness. The owner must match every field
// before accepting either a success or failure; no server session is created.
// Revision is a string so JavaScript cannot round an integer revision.
type PlanBinding struct {
	Operation    string `json:"operation"`
	Signal       string `json:"signal"`
	BaseIdentity string `json:"baseIdentity"`
	Slot         string `json:"slot"`
	Epoch        string `json:"epoch"`
	Revision     string `json:"revision"`
	Library      string `json:"library"`
	Compiler     string `json:"compiler"`
	Target       string `json:"target"`
	Profile      string `json:"profile"`
}

type PlanRequest struct {
	Base       exploration.Query `json:"base"`
	Expression string            `json:"expression"`
	Binding    PlanBinding       `json:"binding"`
}

// RecordPlan is a source-free planning receipt, never captured telemetry.
type RecordPlan struct {
	LoweringDigest string                 `json:"loweringDigest"`
	Version        string                 `json:"version"`
	CELLibrary     string                 `json:"celLibrary"`
	Expression     string                 `json:"expression"`
	Predicate      *exploration.Predicate `json:"predicate"`
	Question       exploration.Query      `json:"question"`
	Plan           exploration.Plan       `json:"plan"`
	Binding        PlanBinding            `json:"binding"`
}

// BaseIdentity hashes validated typed JSON, preserving predicate order and
// nesting. This is content identity, not algebraic or population equivalence.
func BaseIdentity(base exploration.Query) (string, error) {
	if err := base.Validate(); err != nil {
		return "", err
	}
	data, err := json.Marshal(base)
	if err != nil {
		return "", err
	}
	digest := sha256.Sum256(data)
	return hex.EncodeToString(digest[:]), nil
}

// NewPlanRequest prepares a binding for CLI or in-process callers. It detaches
// the base so caller edits cannot silently rebase a request already prepared.
func NewPlanRequest(base exploration.Query, expression, slot, epoch, revision string) (PlanRequest, error) {
	detached, err := exploration.CloneQuery(base)
	if err != nil {
		return PlanRequest{}, planningError("invalid_base", "binding", "base", err)
	}
	id, err := BaseIdentity(detached)
	if err != nil {
		return PlanRequest{}, planningError("invalid_base", "binding", "base", err)
	}
	request := PlanRequest{Base: detached, Expression: expression, Binding: PlanBinding{RecordOperation, base.Signal, id, slot, epoch, revision, LibraryVersion, CompilerVersion, RecordTarget, RecordProfile}}
	if err := validateBinding(request); err != nil {
		return PlanRequest{}, err
	}
	return request, nil
}

func planningError(code, phase, path string, err error) error {
	return &Diagnostic{Code: code, Phase: phase, Path: path, Message: err.Error()}
}

func validateBinding(request PlanRequest) error {
	b := request.Binding
	for _, token := range []struct{ name, value string }{{"slot", b.Slot}, {"epoch", b.Epoch}, {"revision", b.Revision}} {
		if token.value == "" || len(token.value) > 128 || !utf8.ValidString(token.value) {
			return planningError("invalid_binding", "binding", "binding."+token.name, fmt.Errorf("%s requires 1–128 UTF-8 bytes", token.name))
		}
		for _, r := range token.value {
			if unicode.IsSpace(r) || unicode.IsControl(r) {
				return planningError("invalid_binding", "binding", "binding."+token.name, fmt.Errorf("%s contains whitespace or control characters", token.name))
			}
		}
	}
	for _, check := range []struct{ name, got, want string }{{"operation", b.Operation, RecordOperation}, {"signal", b.Signal, request.Base.Signal}, {"library", b.Library, LibraryVersion}, {"compiler", b.Compiler, CompilerVersion}, {"target", b.Target, RecordTarget}, {"profile", b.Profile, RecordProfile}} {
		if check.got != check.want {
			return planningError("binding_mismatch", "binding", "binding."+check.name, fmt.Errorf("unsupported or mismatched %s", check.name))
		}
	}
	id, err := BaseIdentity(request.Base)
	if err != nil {
		return planningError("invalid_base", "binding", "base", err)
	}
	if b.BaseIdentity != id {
		return planningError("binding_mismatch", "binding", "binding.baseIdentity", fmt.Errorf("base identity does not match supplied question"))
	}
	return nil
}

// PlanRecord validates the immutable base and complete result independently.
// It compiles only the owned additional condition; it never converts or drops
// parent predicates. Replacements must use the original base, not the last result.
// No source, clock, readiness probe or store is accepted by this pure use case.
func PlanRecord(request PlanRequest) (*RecordPlan, error) {
	if err := validateBinding(request); err != nil {
		return nil, err
	}
	program, err := Compile(request.Base.Signal, request.Expression)
	if err != nil {
		return nil, err
	}
	question, err := program.Apply(request.Base)
	if err != nil {
		return nil, planningError("invalid_result", "admission", "question", err)
	}
	plan, err := exploration.Compile(question)
	if err != nil {
		return nil, planningError("unsupported_target", "target", "question", err)
	}
	predicate := program.Predicate()
	loweringDigest, err := PredicateDigest(request.Base.Signal, predicate)
	if err != nil {
		return nil, err
	}
	return &RecordPlan{LoweringDigest: loweringDigest, Version: PlanVersion, CELLibrary: LibraryVersion, Expression: request.Expression, Predicate: predicate, Question: question, Plan: plan, Binding: request.Binding}, nil
}

// PredicateDigest identifies validated typed lowering, preserving literal spelling,
// branch order and nesting. It is separate from exact authored source identity.
func PredicateDigest(signal string, predicate *exploration.Predicate) (string, error) {
	if err := exploration.ValidatePredicate(predicate, signal); err != nil {
		return "", err
	}
	data, err := json.Marshal(predicate)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:]), nil
}
