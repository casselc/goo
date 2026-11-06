package verification

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"testing"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/identity"
	"hegel.dev/go/hegel"
)

func packageJSON350(value any) string {
	raw, err := json.Marshal(value)
	if err != nil {
		panic(err)
	}
	return string(raw)
}

func copyQuestion350(q exploration.Query) exploration.Query {
	var copy exploration.Query
	if err := json.Unmarshal([]byte(packageJSON350(q)), &copy); err != nil {
		panic(err)
	}
	return copy
}

func mutateQuestion350(q *exploration.Query) {
	q.Where.All[0].Value = "caller mutation"
	if q.EnvironmentSelector != nil {
		q.EnvironmentSelector.Scope = identity.EnvironmentScope{Mode: "named", Value: "caller mutation"}
	}
	if len(q.GroupBy) > 0 {
		q.GroupBy[0].Key = "caller mutation"
	}
	if q.Numeric != nil {
		q.Numeric.Unit = "caller mutation"
	}
	if q.NumericHistogram != nil {
		q.NumericHistogram.Binning = "caller mutation"
	}
	if q.ReportedEdge != nil {
		q.ReportedEdge.Parent.Namespace.Mode = "value"
		q.ReportedEdge.Parent.Namespace.Value = "caller mutation"
	}
	for i := range q.Where.All {
		if q.Where.All[i].Numeric != nil {
			q.Where.All[i].Numeric.Unit = "caller mutation"
		}
	}
}

// These cases cross the actual browser admission code and Go package/compiler,
// with an independent original-question oracle. Page delivery and fixed clocks
// are deliberately different from reusable configuration.
func TestHegelPackageScope350(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		seed := hegel.Draw(tc, hegel.Integers(1, 1000000))
		envMode := hegel.Draw(tc, hegel.Integers(0, 4))
		profile := hegel.Draw(tc, hegel.Integers(0, 4))
		namespaceMode := hegel.Draw(tc, hegel.Integers(0, 2))
		literal := fmt.Sprintf("literal-'\\雪-%d", seed)
		from := int64(1800000000000 + seed*1000)
		q := exploration.Query{Signal: "spans", Environment: "prod-'雪", FromMS: from, ToMS: from + 60000, Limit: 1 + seed%99, Offset: seed % 1000,
			GroupBy: []exploration.Field{{Kind: "column", Key: "service"}},
			Where: &exploration.Predicate{All: []exploration.Predicate{
				{Field: &exploration.Field{Kind: "column", Key: "service"}, Op: "eq", Value: literal},
				{Field: &exploration.Field{Kind: "resource", Key: identity.NamespaceKey}, Op: "missing"},
				{Any: []exploration.Predicate{{Field: &exploration.Field{Kind: "attribute", Key: "tenant"}, Op: "eq", Value: literal}, {Field: &exploration.Field{Kind: "attribute", Key: "tenant"}, Op: "missing"}}},
			}}}
		if namespaceMode > 0 {
			q.Where.All[1].Op = "eq"
			if namespaceMode == 2 {
				q.Where.All[1].Value = literal
			}
		}
		if envMode > 0 {
			q.Environment = ""
			q.EnvironmentSelector = &identity.EnvironmentSelector{Version: identity.EnvironmentSelectorVersion, Policy: "effective", Scope: identity.EnvironmentScope{Mode: "named", Value: "prod-'雪"}}
			switch envMode {
			case 1:
				q.EnvironmentSelector.Policy = "current-only"
			case 3:
				q.EnvironmentSelector.Scope = identity.EnvironmentScope{Mode: "empty"}
			case 4:
				q.EnvironmentSelector.Scope = identity.EnvironmentScope{Mode: "unreported"}
			}
		}
		switch profile {
		case 0:
			q.Signal = "logs"
			q.GroupBy = append(q.GroupBy, exploration.Field{Kind: "column", Key: "severity"})
		case 1:
			q.Summary, q.SpanStatus, q.GroupTimeSeries, q.BucketMS = "span-duration", true, true, 60000
		case 2, 3:
			q.Numeric = &exploration.NumericBinding{Field: exploration.Field{Kind: "attribute", Key: "custom.quantity"}, Type: "int64", Unit: "items", Parse: "canonical-decimal"}
			q.Where.All = append(q.Where.All, exploration.Predicate{Numeric: q.Numeric, Op: "gte", Value: "0"})
			if profile == 3 {
				q.NumericHistogram = &exploration.NumericHistogramSpec{Binning: "powers-of-two"}
			}
		case 4:
			q.GroupBy = nil
			if envMode == 0 {
				q.ReportedEdge = &exploration.ReportedEdge{Parent: exploration.EdgeService{Service: "upstream", Namespace: &identity.Namespace{Mode: "missing"}}, Child: exploration.EdgeService{Service: literal, Namespace: &identity.Namespace{Mode: "value", Value: literal}}}
			}
		}
		original := copyQuestion350(q)
		if err := original.Validate(); err != nil {
			tc.Fatal("invalid generated question", err)
			return
		}
		minutes := 1 + seed%120
		p, err := exploration.DraftTelemetryPackage(q, fmt.Sprintf("scope-%d", seed), "Typed scope", minutes)
		if err != nil {
			tc.Fatal(err)
			return
		}
		definition := packageJSON350(p)
		wanted := copyQuestion350(original)
		wanted.Limit, wanted.Offset = 100, 0
		if packageJSON350(p.Question(from, from+60000)) != packageJSON350(wanted) {
			tc.Fatal("definition lost original scope, predicate or profile")
			return
		}
		mutateQuestion350(&q)
		if packageJSON350(p) != definition {
			tc.Fatal("draft aliases caller question")
			return
		}
		output := p.Question(from, from+60000)
		mutateQuestion350(&output)
		if packageJSON350(p) != definition {
			tc.Fatal("Question aliases its reusable definition")
			return
		}
		plan, err := exploration.PlanTelemetryPackage(p, from, from+60000, nil)
		if err != nil {
			tc.Fatal(err)
			return
		}
		parsed, err := exploration.ReadTelemetryPackage(plan.YAML)
		if err != nil || packageJSON350(parsed) != definition {
			tc.Fatal("YAML changed typed scope/profile", err)
			return
		}
		decoded, err := exploration.DecodePackageJSON([]byte(definition))
		if err != nil || packageJSON350(decoded) != definition {
			tc.Fatal("JSON changed typed scope/profile", err)
			return
		}
		digest := sha256.Sum256([]byte(definition))
		if plan.Revision != hex.EncodeToString(digest[:]) {
			tc.Fatal("revision does not identify exact canonical definition")
			return
		}
		later, err := exploration.PlanTelemetryPackage(parsed, from+120000, from+180000, nil)
		if err != nil || later.Revision != plan.Revision || later.Question.FromMS != from+120000 || later.Question.ToMS != from+180000 {
			tc.Fatal("preview clock entered configuration identity", err)
			return
		}
		relativeFrom := from + 600000
		relativeTo := relativeFrom + int64(minutes)*60000
		relativePlan, err := exploration.PlanTelemetryPackage(parsed, relativeFrom, relativeTo, nil)
		if err != nil || relativePlan.Revision != plan.Revision {
			tc.Fatal("relative reuse changed configuration identity", err)
			return
		}
		changed, _ := exploration.DecodePackageJSON([]byte(definition))
		if changed.Source.EnvironmentSelector == nil {
			changed.Source.Environment += "-changed"
		} else if changed.Source.EnvironmentSelector.Scope.Mode == "named" {
			if changed.Source.EnvironmentSelector.Policy == "current-only" {
				changed.Source.EnvironmentSelector.Policy = "effective"
			} else {
				changed.Source.EnvironmentSelector.Policy = "current-only"
			}
		} else {
			changed.Source.EnvironmentSelector.Scope = identity.EnvironmentScope{Mode: "named", Value: "replacement"}
		}
		changedRevision, err := changed.Revision()
		if err != nil || changedRevision == plan.Revision {
			tc.Fatal("changing policy/scope did not change revision", err)
			return
		}
		store := exploration.PackageStore{Dir: t.TempDir()}
		saved, err := store.Save(p)
		if err != nil {
			tc.Fatal(err)
			return
		}
		loaded, err := store.Load(saved.Revision)
		if err != nil || packageJSON350(loaded.Definition) != definition {
			tc.Fatal("stored definition lost typed scope", err)
			return
		}
		loaded.Definition.Where.All[0].Value = "caller edits loaded definition"
		again, err := store.Load(saved.Revision)
		if err != nil || packageJSON350(again.Definition) != definition {
			tc.Fatal("caller mutation changed stored bytes", err)
			return
		}

		input, _ := json.Marshal(map[string]any{"plan": plan, "relativePlan": relativePlan, "original": original})
		cmd := exec.CommandContext(t.Context(), "node", "package_scope_350_fixture.mjs")
		cmd.Stdin = bytes.NewReader(input)
		raw, err := cmd.CombinedOutput()
		if err != nil {
			tc.Fatal("actual browser package guard", err, string(raw))
			return
		}
		var got struct {
			Constructed, Staged, RelativeStaged exploration.Query
			Accepted, Detached                  bool
			Label                               string
			Negatives                           []struct {
				Name     string
				Rejected bool
			}
		}
		if err = json.Unmarshal(raw, &got); err != nil || !got.Accepted || !got.Detached || got.Label == "" || strings.Contains(got.Label, "undefined") || len(got.Negatives) != 17 {
			tc.Fatal("browser package adoption or detachment failed", err, string(raw))
			return
		}
		if packageJSON350(got.Constructed) != packageJSON350(wanted) || packageJSON350(got.Staged) != packageJSON350(wanted) {
			tc.Fatal("actual browser projection differs from independent original question", got, wanted)
			return
		}
		wantedRelative := copyQuestion350(wanted)
		wantedRelative.FromMS, wantedRelative.ToMS = relativeFrom, relativeTo
		if packageJSON350(got.RelativeStaged) != packageJSON350(wantedRelative) {
			tc.Fatal("relative browser adoption lost the known definition or returned fixed window", got.RelativeStaged, wantedRelative)
			return
		}
		browserDefinition, err := exploration.DraftTelemetryPackage(got.Constructed, p.Name, p.Title, minutes)
		if err != nil || packageJSON350(browserDefinition) != definition {
			tc.Fatal("actual browser question changed Go draft semantics", err)
			return
		}
		browserPlan, err := exploration.Compile(got.Staged)
		if err != nil || packageJSON350(browserPlan) != packageJSON350(plan.Plan) {
			tc.Fatal("actual staged browser question changed compiler scope/profile", err)
			return
		}
		for _, negative := range got.Negatives {
			if !negative.Rejected {
				tc.Fatal("browser accepted poisoned plan", negative.Name)
				return
			}
		}
		planBefore := packageJSON350(plan)
		p.Where.All[0].Value = "caller changes definition after planning"
		if p.Source.EnvironmentSelector != nil {
			p.Source.EnvironmentSelector.Scope.Value = "caller mutation"
		}
		if packageJSON350(plan) != planBefore {
			tc.Fatal("returned plan aliases caller definition")
			return
		}
		planDefinition := packageJSON350(plan.Definition)
		mutateQuestion350(&plan.Question)
		if packageJSON350(plan.Definition) != planDefinition {
			tc.Fatal("plan question aliases plan definition/hash")
		}
	}, options(40)...)
}
