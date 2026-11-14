package querycel

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"cel.dev/cel-go/cel"

	"github.com/casselc/goo/exploration"
)

func parentQuestion() exploration.Query {
	return exploration.Query{Signal: "logs", Environment: "prod", FromMS: 1000, ToMS: 61000, Limit: 100, Offset: 0, LogSeverity: true,
		Where: &exploration.Predicate{Any: []exploration.Predicate{{Field: &exploration.Field{Kind: "column", Key: "service"}, Op: "eq", Value: "worker"}, {Field: &exploration.Field{Kind: "column", Key: "service"}, Op: "eq", Value: "api"}}}}
}
func fields(body string, severity string) map[exploration.Field]string {
	return map[exploration.Field]string{{Kind: "column", Key: "body"}: body, {Kind: "column", Key: "severity_number"}: severity}
}

func TestCheckedOverloadsAndLiteralParity(t *testing.T) {
	for _, expression := range []string{
		`log.body.contains("雪_%'\\.*") && log.severity_number >= 17u`,
		`"雪_%'\\.*" == log.body || log.severity_number != 9u`,
		`24u >= log.severity_number && 17u <= log.severity_number`,
		`otel.severityBand(log.severity_number) == "error" && log.body != ""`,
		`"invalid" == otel.severityBand(log.severity_number)`,
	} {
		program, err := Compile("logs", expression)
		if err != nil {
			t.Fatalf("%s: %v", expression, err)
		}
		for _, body := range []string{"", "literal 雪_%'\\.* content", "雪_%'\\.*", "雪_%'\\.* ", "雪_%'\\.*\x00tail"} {
			for _, severity := range []string{"0", "9", "17", "20", "21", "24", "25", "255", "18446744073709551615"} {
				values := fields(body, severity)
				got, err := program.Eval(values)
				if err != nil {
					t.Fatal(err)
				}
				if want := exploration.Match(program.Predicate(), values); got != want {
					t.Fatalf("%s: %q/%s CEL=%t Goo=%t", expression, body, severity, got, want)
				}
			}
		}
	}
}

func TestAllSeverityBandsAndBoundaries(t *testing.T) {
	for _, band := range []string{"trace", "debug", "info", "warn", "error", "fatal", "unspecified", "invalid"} {
		program, err := Compile("logs", `otel.severityBand(log.severity_number) == "`+band+`"`)
		if err != nil {
			t.Fatal(err)
		}
		for n := uint64(0); n < 256; n++ {
			value := json.Number("")
			wire, _ := json.Marshal(n)
			value = json.Number(wire)
			values := fields("", value.String())
			got, err := program.Eval(values)
			if err != nil {
				t.Fatal(err)
			}
			want := exploration.LogSeverityClass(n) == band
			if got != want || exploration.Match(program.Predicate(), values) != want {
				t.Fatalf("%s/%d", band, n)
			}
		}
	}
}

func TestUnsupportedExpressionsRefuseBeforeReferenceExecution(t *testing.T) {
	for _, expression := range []string{
		`true`, `false && log.body.matches(".*")`, `!log.body.contains("x")`, `log.body.matches("x")`,
		`log.severity_number > 17u`, `log.severity_number < 20u`, `log.severity_number >= 17`,
		`log.severity_number + 1u >= 17u`, `uint("17") == log.severity_number`,
		`log.body == record.service`, `[1,2].exists(x, x > 0)`, `has(log.body)`,
		`resource["service.name"] == "worker"`, `log.body.startsWith("x")`,
		`otel.severityBand(record.timestamp_ns) == "error"`, `otel.severityBand(log.severity_number) != "error"`,
		`otel.severityBand(log.severity_number) == "ERROR"`, `log.body == null`,
		`log.body.contains("\u0000")`,
	} {
		if _, err := Compile("logs", expression); err == nil {
			t.Fatalf("admitted %s", expression)
		}
	}
	for _, input := range []string{"", strings.Repeat("x", MaxExpressionBytes+1), "log.body == \"\x00\"", string([]byte{0xff})} {
		if _, err := Compile("logs", input); err == nil {
			t.Fatal("admitted invalid input")
		}
	}
	if _, err := Compile("spans", `log.body == "x"`); err == nil {
		t.Fatal("log field leaked into spans")
	}
	if _, err := Compile("metrics", `record.service == "x"`); err == nil {
		t.Fatal("unknown signal admitted")
	}
	if _, err := Compile("spans", `span.duration_ns >= 10u && span.name.contains("work")`); err != nil {
		t.Fatal(err)
	}
}

func TestApplyRetainsWholeParentAndDetachedOwnership(t *testing.T) {
	program, err := Compile("logs", `otel.severityBand(log.severity_number) == "error"`)
	if err != nil {
		t.Fatal(err)
	}
	parent := parentQuestion()
	before, _ := json.Marshal(parent)
	query, err := program.Apply(parent)
	if err != nil {
		t.Fatal(err)
	}
	if query.FromMS != parent.FromMS || query.ToMS != parent.ToMS || query.Environment != parent.Environment || query.LogSeverity != parent.LogSeverity || !reflect.DeepEqual(query.Where.All[0], *parent.Where) {
		t.Fatal("parent intent changed", query)
	}
	plan, err := exploration.Compile(query)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(plan.Count, "SeverityNumber >=") || !strings.Contains(plan.Count, " OR ") {
		t.Fatal("SQL lost numeric band or parent OR", plan.Count)
	}
	query.Where.All[0].Any[0].Value = "mutated"
	predicate := program.Predicate()
	predicate.All[0].Value = "mutated"
	after, _ := json.Marshal(parent)
	if string(after) != string(before) {
		t.Fatal("parent aliased result")
	}
	if program.Predicate().All[0].Value != "17" {
		t.Fatal("predicate aliased program")
	}
	if _, err := program.Apply(exploration.Query{Signal: "spans", Environment: "prod", FromMS: 1000, ToMS: 61000, Limit: 100}); err == nil {
		t.Fatal("cross-signal apply admitted")
	}
	parent.Where = &exploration.Predicate{All: make([]exploration.Predicate, 30)}
	for i := range parent.Where.All {
		parent.Where.All[i] = exploration.Predicate{Field: &exploration.Field{Kind: "column", Key: "service"}, Op: "eq", Value: "api"}
	}
	if err := parent.Validate(); err != nil {
		t.Fatal(err)
	}
	if _, err := program.Apply(parent); err == nil {
		t.Fatal("combined expanded AST budget not enforced")
	}
}

func TestBoundedReferenceRequiresOnlyUsedWellTypedFields(t *testing.T) {
	program, err := Compile("logs", `log.body.contains("x")`)
	if err != nil {
		t.Fatal(err)
	}
	if got, err := program.Eval(fields("x", "not consulted")); err != nil || !got {
		t.Fatal("unused field was read", got, err)
	}
	for _, values := range []map[exploration.Field]string{nil, fields(strings.Repeat("x", MaxFieldBytes+1), "0"), fields(string([]byte{0xff}), "0")} {
		if _, err := program.Eval(values); err == nil {
			t.Fatal("invalid fixed field admitted")
		}
	}
	numeric, _ := Compile("logs", `log.severity_number == 17u`)
	for _, value := range []string{"017", "-1", "17.0", "18446744073709551616", ""} {
		if _, err := numeric.Eval(fields("", value)); err == nil {
			t.Fatal("noncanonical number admitted", value)
		}
	}
	for _, program := range []*Program{nil, {}} {
		if _, err := program.Eval(nil); err == nil {
			t.Fatal("uninitialized evaluator admitted")
		}
		if _, err := program.Apply(parentQuestion()); err == nil {
			t.Fatal("uninitialized apply admitted")
		}
	}
}

func TestLibraryIsUsableAsAnOrdinaryCELExtension(t *testing.T) {
	env, err := cel.NewEnv(cel.Lib(OTelLibrary{}), cel.Variable("number", cel.UintType))
	if err != nil {
		t.Fatal(err)
	}
	checked, issues := env.Compile(`otel.severityBand(number) == "error"`)
	if issues.Err() != nil {
		t.Fatal(issues.Err())
	}
	program, err := env.Program(checked)
	if err != nil {
		t.Fatal(err)
	}
	value, _, err := program.Eval(map[string]any{"number": uint64(17)})
	if err != nil || value.Value() != true {
		t.Fatal(value, err)
	}
}

func TestExpandedPredicateAndStaticWorkBudgets(t *testing.T) {
	// A CEL expression can typecheck and fit its syntax limits while its helper
	// expansion no longer fits the ordinary Goo predicate budget.
	leaf := `otel.severityBand(log.severity_number) == "error"`
	if _, err := Compile("logs", strings.Join([]string{leaf, leaf, leaf, leaf, leaf, leaf, leaf, leaf, leaf, leaf, leaf}, " || ")); err == nil {
		t.Fatal("expanded predicate budget was bypassed")
	}
	text := strings.Repeat("x", 1024)
	leaf = `log.body.contains("` + text + `")`
	if _, err := Compile("logs", leaf); err != nil {
		t.Fatal("bounded single traversal refused", err)
	}
	if _, err := Compile("logs", strings.Join([]string{leaf, leaf, leaf}, " && ")); err == nil || !strings.Contains(err.Error(), "work budget") {
		t.Fatal("static traversal budget was not enforced", err)
	}
}
