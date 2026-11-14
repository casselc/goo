package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/querycel"
)

func TestPlanCELSharedQueryCompiler(t *testing.T) {
	const text = `otel.severityBand(log.severity_number) == "error" && log.body.contains("x' OR 1=1 --")`
	parent := exploration.Query{Signal: "logs", Environment: "prod", FromMS: 1000, ToMS: 61000, Limit: 20, LogSeverity: true,
		Where: &exploration.Predicate{Any: []exploration.Predicate{
			{Field: &exploration.Field{Kind: "column", Key: "service"}, Op: "eq", Value: "api"},
			{Field: &exploration.Field{Kind: "attribute", Key: "team"}, Op: "eq", Value: "core"},
		}}}
	path := filepath.Join(t.TempDir(), "question.json")
	data, err := json.Marshal(parent)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, data, 0600); err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	if err := run([]string{"plan-cel", "-query", path, "-expression", text}, &out); err != nil {
		t.Fatal(err)
	}
	var result struct {
		Version, CELLibrary, Expression string
		Predicate                       *exploration.Predicate
		Question                        exploration.Query
		Plan                            exploration.Plan
	}
	if err := json.Unmarshal(out.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if result.Version != "goo.cel-plan/1" || result.CELLibrary != querycel.LibraryVersion || result.Expression != text {
		t.Fatal("planning provenance changed", result)
	}
	if !reflect.DeepEqual(&result.Question.Where.All[0], parent.Where) || !result.Question.LogSeverity || result.Question.FromMS != parent.FromMS || result.Question.ToMS != parent.ToMS || result.Question.Limit != parent.Limit {
		t.Fatal("CEL changed the parent question", result.Question)
	}
	if !reflect.DeepEqual(&result.Question.Where.All[1], result.Predicate) {
		t.Fatal("detached CEL predicate differs from applied refinement")
	}
	want, err := exploration.Compile(result.Question)
	if err != nil || !reflect.DeepEqual(want, result.Plan) {
		t.Fatal("plan does not use shared compiler", err)
	}
	if strings.Contains(result.Plan.Rows, "x' OR 1=1 --") || strings.Contains(result.Plan.Count, "x' OR 1=1 --") {
		t.Fatal("CEL text interpolated into SQL")
	}
	parameterized := false
	for _, value := range result.Plan.Parameters {
		parameterized = parameterized || value == "x' OR 1=1 --"
	}
	if !parameterized {
		t.Fatal("literal message parameter lost")
	}
}

func TestPlanCELExampleAndRefusals(t *testing.T) {
	var out bytes.Buffer
	if err := run([]string{"plan-cel", "-query", "../../examples/cel/logs.json", "-expression-file", "../../examples/cel/errors.cel"}, &out); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), `"celLibrary"`) || !strings.Contains(out.String(), `"value": "17"`) || !strings.Contains(out.String(), `"value": "20"`) {
		t.Fatal("example did not lower severity helper", out.String())
	}
	large := filepath.Join(t.TempDir(), "large.cel")
	if err := os.WriteFile(large, []byte(strings.Repeat(" ", querycel.MaxExpressionBytes+1)), 0600); err != nil {
		t.Fatal(err)
	}
	for _, args := range [][]string{
		{"plan-cel"},
		{"plan-cel", "-query", "../../examples/cel/logs.json"},
		{"plan-cel", "-query", "../../examples/cel/logs.json", "-expression", `log.body.matches(".*")`},
		{"plan-cel", "-query", "../../examples/cel/logs.json", "-expression", `false && log.body.matches(".*")`},
		{"plan-cel", "-query", "../../examples/cel/logs.json", "-expression", `log.severity_number >= 17`},
		{"plan-cel", "-query", "../../examples/cel/logs.json", "-expression-file", large},
		{"plan-cel", "-query", "../../examples/cel/logs.json", "-expression", `log.body == "timeout"`, "-expression-file", "../../examples/cel/errors.cel"},
	} {
		out.Reset()
		if err := run(args, &out); err == nil || out.Len() != 0 {
			t.Fatal("unsupported authoring emitted a plan", args, err, out.String())
		}
	}
}
