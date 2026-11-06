package exploration

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/casselc/goo/identity"
	"gopkg.in/yaml.v3"
)

func packageQuestion350(mode string) Query {
	q := Query{Signal: "spans", Environment: "prod-'雪", FromMS: 1000, ToMS: 61000, Limit: 7, Offset: 42, GroupBy: []Field{{Kind: "attribute", Key: "cohort"}}, Where: &Predicate{All: []Predicate{{Field: &Field{Kind: "column", Key: "service"}, Op: "eq", Value: "api-'雪"}, {Field: &Field{Kind: "resource", Key: "service.namespace"}, Op: "eq", Value: ""}}}}
	if mode != "scalar" {
		q.Environment = ""
		q.EnvironmentSelector = &identity.EnvironmentSelector{Version: identity.EnvironmentSelectorVersion, Policy: "effective", Scope: identity.EnvironmentScope{Mode: "named", Value: "prod-'雪"}}
		switch mode {
		case "current":
			q.EnvironmentSelector.Policy = "current-only"
		case "empty":
			q.EnvironmentSelector.Scope = identity.EnvironmentScope{Mode: "empty"}
		case "unreported":
			q.EnvironmentSelector.Scope = identity.EnvironmentScope{Mode: "unreported"}
		}
	}
	return q
}
func TestTelemetryPackageTypedScopeProfiles350(t *testing.T) {
	for _, mode := range []string{"scalar", "current", "effective", "empty", "unreported"} {
		for _, profile := range []string{"count", "duration-status-time", "numeric-histogram"} {
			t.Run(mode+"/"+profile, func(t *testing.T) {
				q := packageQuestion350(mode)
				switch profile {
				case "duration-status-time":
					q.Summary = "span-duration"
					q.SpanStatus = true
					q.BucketMS = 10000
					q.GroupTimeSeries = true
				case "numeric-histogram":
					q.Numeric = &NumericBinding{Field: Field{Kind: "attribute", Key: "amount"}, Type: "uint64", Unit: "1", Parse: "canonical-decimal"}
					q.NumericHistogram = &NumericHistogramSpec{Binning: "powers-of-two"}
				}
				original := detached(q)
				p, err := DraftTelemetryPackage(q, "typed-package", "Typed package", 5)
				if err != nil {
					t.Fatal(err)
				}
				expected := detached(original)
				expected.Offset = 0
				expected.Limit = 100
				if !reflect.DeepEqual(p.Question(q.FromMS, q.ToMS), expected) {
					t.Fatal("draft lost full profile or policy")
				}
				q.Where.All[0].Value = "caller mutation"
				q.GroupBy[0].Key = "mutated"
				if q.EnvironmentSelector != nil {
					q.EnvironmentSelector.Scope = identity.EnvironmentScope{Mode: "named", Value: "mutated"}
				}
				if q.Numeric != nil {
					q.Numeric.Unit = "mutated"
				}
				if !reflect.DeepEqual(p.Question(original.FromMS, original.ToMS), expected) {
					t.Fatal("draft aliases caller")
				}
				plan, err := PlanTelemetryPackage(p, original.FromMS, original.ToMS, nil)
				if err != nil {
					t.Fatal(err)
				}
				wire, err := json.Marshal(p)
				if err != nil {
					t.Fatal(err)
				}
				decoded, err := DecodePackageJSON(wire)
				if err != nil || !reflect.DeepEqual(decoded, p) {
					t.Fatal("JSON scope/profile roundtrip changed", err)
				}
				yamlPackage, err := ReadTelemetryPackage(plan.YAML)
				if err != nil || !reflect.DeepEqual(yamlPackage, p) {
					t.Fatal("YAML scope/profile roundtrip changed", err, plan.YAML)
				}
				later, err := PlanTelemetryPackage(yamlPackage, 100000, 160000, nil)
				if err != nil || later.Revision != plan.Revision || later.Question.FromMS != 100000 || later.Question.ToMS != 160000 || later.Question.FromMS == plan.Question.FromMS {
					t.Fatal("fixed preview and new window authority confused", err)
				}
				store := PackageStore{Dir: t.TempDir()}
				saved, err := store.Save(p)
				if err != nil {
					t.Fatal(err)
				}
				loaded, err := store.Load(saved.Revision)
				if err != nil || !reflect.DeepEqual(loaded.Definition, p) || saved.Revision != plan.Revision {
					t.Fatal("stored definition changed scope", err)
				}
				snapshot := detached(*plan)
				p.Where.All[0].Value = "changed after plan"
				p.GroupBy[0].Key = "changed"
				p.Views[0] = "changed"
				if p.Source.EnvironmentSelector != nil {
					p.Source.EnvironmentSelector.Scope = identity.EnvironmentScope{Mode: "named", Value: "changed"}
				}
				if p.Numeric != nil {
					p.Numeric.Unit = "changed"
				}
				if !reflect.DeepEqual(*plan, snapshot) {
					t.Fatal("plan retains mutable definition aliases")
				}
				question := plan.Definition.Question(original.FromMS, original.ToMS)
				question.Where.All[0].Value = "changed question"
				question.GroupBy[0].Key = "changed"
				if question.EnvironmentSelector != nil {
					question.EnvironmentSelector.Scope = identity.EnvironmentScope{Mode: "named", Value: "changed"}
				}
				if question.Numeric != nil {
					question.Numeric.Unit = "changed"
				}
				if !reflect.DeepEqual(*plan, snapshot) {
					t.Fatal("Question output aliases package or plan")
				}
				target := &AnnotationTarget{Name: "operation.run", Role: "server"}
				targetPlan, targetErr := PlanTelemetryPackage(yamlPackage, original.FromMS, original.ToMS, target)
				if mode == "scalar" || mode == "current" {
					if targetErr != nil || targetPlan.Handoff.GoExample == "" || !strings.Contains(strings.Join(targetPlan.Handoff.Notes, " "), "does not encode the package environment") {
						t.Fatal("current-only optional template lost explicit non-equivalence", targetErr)
					}
				} else if targetErr == nil {
					t.Fatal("external typed scope projected into function template")
				}
			})
		}
	}
}
func TestPackageSourceClosedJSONYAML350(t *testing.T) {
	valid := `{"adapter":"clickhouse.otel/v1","signal":"spans","environmentSelector":{"version":"goo/environment-selector/v1","policy":"effective","scope":{"mode":"named","value":"prod"}}}`
	invalid := []string{
		strings.Replace(valid, `"environmentSelector":`, `"environment":"", "environmentSelector":`, 1),
		strings.Replace(valid, `"environmentSelector":`, `"environment":"prod", "environmentSelector":`, 1),
		strings.Replace(valid, `{"version":"goo/environment-selector/v1","policy":"effective","scope":{"mode":"named","value":"prod"}}`, `null`, 1),
		strings.Replace(valid, `"policy":"effective"`, `"policy":null`, 1),
		strings.Replace(valid, `"scope":{"mode":"named","value":"prod"}`, `"scope":{"mode":"named","value":"prod","unknown":true}`, 1),
		strings.Replace(valid, `"value":"prod"`, `"value":null`, 1),
		strings.Replace(valid, `"value":"prod"`, `"value":"\ud800"`, 1),
		`{"adapter":"clickhouse.otel/v1","signal":"spans","environment":null}`,
		`{"adapter":"clickhouse.otel/v1","signal":"spans","environment":"\ud800"}`,
		`{"adapter":"clickhouse.otel/v1","signal":"spans","environment":"prod","unknown":true}`,
	}
	for _, raw := range invalid {
		var source PackageSource
		if json.Unmarshal([]byte(raw), &source) == nil {
			t.Fatal("invalid JSON source admitted", raw)
		}
		var values map[string]any
		if json.Unmarshal([]byte(raw), &values) == nil {
			encoded, _ := yaml.Marshal(values)
			if strings.Contains(raw, `\ud800`) {
				continue
			}
			if yaml.Unmarshal(encoded, &source) == nil {
				t.Fatal("YAML weakened source admission", string(encoded))
			}
		}
	}
	for _, raw := range []string{valid, strings.Replace(valid, `"value":"prod"`, `"value":"\ufffd"`, 1), strings.Replace(valid, `"value":"prod"`, `"value":"\ud83d\ude42"`, 1)} {
		var s PackageSource
		if err := json.Unmarshal([]byte(raw), &s); err != nil {
			t.Fatal("valid Unicode refused", err)
		}
		encoded, _ := yaml.Marshal(s)
		var again PackageSource
		if err := yaml.Unmarshal(encoded, &again); err != nil || !reflect.DeepEqual(s, again) {
			t.Fatal("valid YAML source changed", err)
		}
	}
	for _, text := range []string{`adapter: clickhouse.otel/v1
signal: spans
environment: !!binary /w==
`, `adapter: clickhouse.otel/v1
signal: spans
environmentSelector:
  version: goo/environment-selector/v1
  policy: effective
  scope:
    mode: named
    value: prod
    value: other
`, `adapter: clickhouse.otel/v1
signal: spans
environment: &value prod
unknown: *value
`} {
		var s PackageSource
		if yaml.Unmarshal([]byte(text), &s) == nil {
			t.Fatal("unsafe YAML source accepted", text)
		}
	}
}
func TestPackageStoredUnicodeAndInvalidCaller350(t *testing.T) {
	q := packageQuestion350("effective")
	q.EnvironmentSelector.Scope.Value = "�"
	p, err := DraftTelemetryPackage(q, "unicode-package", "Unicode", 1)
	if err != nil {
		t.Fatal(err)
	}
	store := PackageStore{Dir: t.TempDir()}
	saved, err := store.Save(p)
	if err != nil {
		t.Fatal(err)
	}
	filename := filepath.Join(store.Dir, saved.Revision+".json")
	raw, err := os.ReadFile(filename)
	if err != nil {
		t.Fatal(err)
	}
	altered := strings.Replace(string(raw), `"value":"�"`, `"value":"\ud800"`, 1)
	if altered == string(raw) {
		t.Fatal("Unicode mutation did not apply")
	}
	if err = os.WriteFile(filename, []byte(altered), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err = store.Load(saved.Revision); err == nil {
		t.Fatal("lossy Unicode repaired before revision check")
	}
	p.Source.EnvironmentSelector.Scope.Value = string([]byte{0xff})
	if _, err = p.Revision(); err == nil {
		t.Fatal("invalid caller selector repaired before hashing")
	}
	if _, err = store.Save(p); err == nil {
		t.Fatal("invalid caller selector repaired before publication")
	}
	if _, err = PlanTelemetryPackage(p, 1000, 61000, nil); err == nil {
		t.Fatal("invalid caller selector repaired before planning")
	}
	if p.Question(1000, 61000).EnvironmentSelector.Scope.Value != string([]byte{0xff}) {
		t.Fatal("Question silently repaired invalid identity")
	}
}

func TestPackageSourceParsedMapDuplicate350(t *testing.T) {
	selector := `{"version":"goo/environment-selector/v1","policy":"effective","scope":{"mode":"named","value":"prod"}}`
	prefix := `{"adapter":"clickhouse.otel/v1","signal":"spans","environmentSelector":`
	var source PackageSource
	if err := json.Unmarshal([]byte(prefix+`{"version":"invalid"},"environmentSelector":`+selector+`}`), &source); err != nil || source.EnvironmentSelector.Scope.Value != "prod" {
		t.Fatal("parsed final selector not authoritative", err)
	}
	if json.Unmarshal([]byte(prefix+selector+`,"environmentSelector":{"version":"invalid"}}`), &source) == nil {
		t.Fatal("invalid final selector accepted")
	}
	if json.Unmarshal([]byte(`{"adapter":"clickhouse.otel/v1","SIGNAL":"spans","environment":"prod"}`), &source) == nil {
		t.Fatal("unknown case-variant source key accepted")
	}
}
func TestPackagePredicateUnicodePredecode350(t *testing.T) {
	q := packageQuestion350("effective")
	q.Where.All[0].Value = "�"
	p, err := DraftTelemetryPackage(q, "predicate-unicode", "Unicode predicate", 1)
	if err != nil {
		t.Fatal(err)
	}
	raw, _ := json.Marshal(p)
	if _, err = DecodePackageJSON(raw); err != nil {
		t.Fatal("valid replacement character predicate refused", err)
	}
	altered := strings.Replace(string(raw), `"value":"�"`, `"value":"\ud800"`, 1)
	if altered == string(raw) {
		t.Fatal("predicate mutation did not apply")
	}
	if _, err = DecodePackageJSON([]byte(altered)); err == nil {
		t.Fatal("predicate repaired before definition validation")
	}
	store := PackageStore{Dir: t.TempDir()}
	saved, err := store.Save(p)
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(filepath.Join(store.Dir, saved.Revision+".json"), []byte(altered), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err = store.Load(saved.Revision); err == nil {
		t.Fatal("predicate repaired before stored revision validation")
	}
}
