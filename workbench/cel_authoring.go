package workbench

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"unicode"
	"unicode/utf8"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/identity"
	"github.com/casselc/goo/querycel"
)

const RecordAuthoringVersion = "goo.cel-authoring/1"

// RecordAuthoring preserves source beside ordinary executable meaning. The
// binding tokens describe an authoring revision, not authenticated attribution
// or execution authority. Validation always recompiles the paired typed query.
type RecordAuthoring struct {
	Version        string               `json:"version"`
	Representation string               `json:"representation"`
	Source         string               `json:"source"`
	Base           exploration.Query    `json:"base"`
	Binding        querycel.PlanBinding `json:"binding"`
	ParentRevision string               `json:"parentRevision"`
	LoweringDigest string               `json:"loweringDigest"`
}

// NewRecordAuthoring recomputes the checked result rather than trusting a client
// receipt. Exact source is preserved; mutable base members are detached.
func NewRecordAuthoring(request querycel.PlanRequest, parentRevision string) (RecordAuthoring, error) {
	if !authoringToken(parentRevision) {
		return RecordAuthoring{}, errors.New("authoring requires a bounded parent revision token")
	}
	plan, err := querycel.PlanRecord(request)
	if err != nil {
		return RecordAuthoring{}, err
	}
	base, err := exploration.CloneQuery(request.Base)
	if err != nil {
		return RecordAuthoring{}, err
	}
	return RecordAuthoring{RecordAuthoringVersion, "cel", request.Expression, base, request.Binding, parentRevision, plan.LoweringDigest}, nil
}

func authoringToken(value string) bool {
	if value == "" || len(value) > 128 || !utf8.ValidString(value) {
		return false
	}
	for _, r := range value {
		if unicode.IsSpace(r) || unicode.IsControl(r) {
			return false
		}
	}
	return true
}

// Validate checks the exact source/IR pair against the current compiler contract.
// It does not normalize a changed library, rewrite query policy or read telemetry.
func (a RecordAuthoring) Validate(question exploration.Query) error {
	if a.Version != RecordAuthoringVersion || a.Representation != "cel" || !authoringToken(a.ParentRevision) || !ID.MatchString(a.LoweringDigest) {
		return errors.New("invalid CEL authoring version, representation, parent revision or lowering digest")
	}
	plan, err := querycel.PlanRecord(querycel.PlanRequest{Base: a.Base, Expression: a.Source, Binding: a.Binding})
	if err != nil {
		return fmt.Errorf("CEL authoring: %w", err)
	}
	if plan.LoweringDigest != a.LoweringDigest {
		return errors.New("CEL authoring lowering digest does not match checked source")
	}
	if err := question.Validate(); err != nil {
		return err
	}
	checked, err := json.Marshal(plan.Question)
	if err != nil {
		return err
	}
	supplied, err := json.Marshal(question)
	if err != nil {
		return err
	}
	if !bytes.Equal(checked, supplied) {
		return errors.New("CEL authoring does not match the complete saved question")
	}
	return nil
}

// UnmarshalJSON keeps provenance closed, non-null and case-sensitive. Parsing
// before typed decode preserves JSON.parse's last-value object semantics rather
// than Go's merging of repeated nested struct objects.
func (a *RecordAuthoring) UnmarshalJSON(raw []byte) error {
	if !identity.WellFormedJSON(raw) {
		return errors.New("invalid CEL authoring JSON or Unicode")
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	var value any
	if err := decoder.Decode(&value); err != nil {
		return err
	}
	if decoder.Decode(new(any)) != io.EOF {
		return errors.New("require one CEL authoring object")
	}
	fields, ok := value.(map[string]any)
	if !ok {
		return errors.New("require a CEL authoring object")
	}
	required := []string{"version", "representation", "source", "base", "binding", "parentRevision", "loweringDigest"}
	if len(fields) != len(required) {
		return errors.New("CEL authoring requires exactly its declared fields")
	}
	for _, name := range required {
		if fields[name] == nil {
			return fmt.Errorf("CEL authoring requires non-null %s", name)
		}
	}
	binding, ok := fields["binding"].(map[string]any)
	if !ok || len(binding) != 10 {
		return errors.New("CEL authoring requires a complete checked binding")
	}
	for _, name := range []string{"operation", "signal", "baseIdentity", "slot", "epoch", "revision", "library", "compiler", "target", "profile"} {
		if _, ok := binding[name].(string); !ok {
			return fmt.Errorf("CEL authoring requires string binding.%s", name)
		}
	}
	canonical, err := json.Marshal(value)
	if err != nil {
		return err
	}
	type wire RecordAuthoring
	var out wire
	decoder = json.NewDecoder(bytes.NewReader(canonical))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&out); err != nil {
		return err
	}
	*a = RecordAuthoring(out)
	return nil
}

func (v ExplorationView) validateAuthoring() error {
	if v.Authoring == nil {
		return nil
	}
	if len(v.Authoring) == 0 || len(v.Authoring) > len(v.Queries) {
		return errors.New("authoring requires configured record signals")
	}
	for signal, authoring := range v.Authoring {
		found := false
		for _, q := range v.Queries {
			if q.Signal == signal {
				found = true
				if err := authoring.Validate(q); err != nil {
					return fmt.Errorf("%s authoring: %w", signal, err)
				}
				break
			}
		}
		if !found {
			return errors.New("authoring signal does not reference a configured question")
		}
	}
	return nil
}
