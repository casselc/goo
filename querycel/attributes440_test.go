package querycel

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/casselc/goo/exploration"
)

func TestAttribute440PresenceSafeReferenceAndExactRepresentation(t *testing.T) {
	for _, signal := range []string{"logs", "spans"} {
		for _, target := range []struct{ name, kind string }{{"resource.attributes", "resource"}, {"record.attributes", "attribute"}} {
			key := "http.route"
			field := exploration.Field{Kind: target.kind, Key: key}
			for _, expression := range []string{
				`"http.route" in ` + target.name,
				`!("http.route" in ` + target.name + `)`,
				target.name + `.eq("http.route", "")`,
				target.name + `.ne("http.route", "")`,
				target.name + `.eq("http.route", "0")`,
				target.name + `.contains("http.route", "雪_%'\\")`,
				`!("http.route" in ` + target.name + `) || ` + target.name + `.ne("http.route", "0")`,
			} {
				program, err := Compile(signal, expression)
				if err != nil {
					t.Fatalf("%s: %v", expression, err)
				}
				printed, err := PrintPredicate(signal, program.Predicate())
				if err != nil {
					t.Fatal(err)
				}
				roundtrip, err := Compile(signal, printed)
				if err != nil || !reflect.DeepEqual(program.Predicate(), roundtrip.Predicate()) {
					t.Fatalf("%s roundtrip %s: %v", expression, printed, err)
				}
				for _, value := range []*string{nil, new(""), new("0"), new("7"), new("雪_%'\\ tail")} {
					values := map[exploration.Field]string{}
					if value != nil {
						values[field] = *value
					}
					got, err := program.Eval(values)
					if err != nil {
						t.Fatalf("%s/%v: %v", expression, value, err)
					}
					if want := exploration.Match(program.Predicate(), values); got != want {
						t.Fatalf("%s/%v CEL%t != AST%t", expression, value, got, want)
					}
				}
			}
		}
	}
	// Presence is independent of an empty string. Present-only inequality is not
	// boolean negation; missing-inclusive questions name their missing arm.
	for _, tc := range []struct {
		source         string
		missing, empty bool
	}{
		{`resource.attributes.eq("k", "")`, false, true},
		{`resource.attributes.ne("k", "x")`, false, true},
		{`"k" in resource.attributes`, false, true},
		{`!("k" in resource.attributes)`, true, false},
		{`!("k" in resource.attributes) || resource.attributes.ne("k", "x")`, true, true},
	} {
		p, err := Compile("logs", tc.source)
		if err != nil {
			t.Fatal(err)
		}
		for _, x := range []struct {
			values map[exploration.Field]string
			want   bool
		}{{map[exploration.Field]string{}, tc.missing}, {map[exploration.Field]string{{Kind: "resource", Key: "k"}: ""}, tc.empty}} {
			got, err := p.Eval(x.values)
			if err != nil || got != x.want {
				t.Fatalf("%s: %t %v", tc.source, got, err)
			}
		}
	}
}

func TestAttribute440UnsupportedFormsAndBudgets(t *testing.T) {
	for _, source := range []string{
		`resource.attributes["k"] == ""`, `resource.attributes["k"] != "x"`,
		`"k" in resource`, `has(resource.attributes.k)`,
		`!resource.attributes.eq("k", "x")`, `!(resource.attributes.eq("k", "x") || record.service == "worker")`,
		`resource.attributes.eq(record.service, "x")`, `resource.attributes.eq("k", record.service)`,
		`{"k":"v"}.eq("k", "v")`, `record.attributes.eq("k", 7)`,
		`int(record.attributes["k"]) >= 7`, `resource.attributes.exists(k, k == "x")`,
		`"k" in {"k":"v"}`, `resource.attributes.eq("", "v")`,
		`resource.attributes.eq("\u0000", "v")`, `resource.attributes.eq("k", "\u0000")`,
		`resource.attributes.contains("k", "` + strings.Repeat("x", 1025) + `")`,
	} {
		if _, err := Compile("logs", source); err == nil {
			t.Fatalf("admitted %s", source)
		}
	}
	for _, key := range []string{"record.service", "resource.attributes", "service.name", "雪_%'\\"} {
		quoted, _ := json.Marshal(key)
		p, err := Compile("logs", `record.attributes.eq(`+string(quoted)+`, "v")`)
		if err != nil {
			t.Fatal(err)
		}
		values := map[exploration.Field]string{{Kind: "attribute", Key: key}: "v"}
		got, err := p.Eval(values)
		if err != nil || !got {
			t.Fatalf("literal key alias %s %v", key, err)
		}
	}
	p, err := Compile("logs", `resource.attributes.eq("k", "x") || resource.attributes.eq("j", "v")`)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := p.Eval(map[exploration.Field]string{{Kind: "resource", Key: "k"}: "x", {Kind: "resource", Key: "j"}: string([]byte{0xff})}); err == nil {
		t.Fatal("malformed referenced text hidden by short circuit")
	}
	if got, err := p.Eval(map[exploration.Field]string{{Kind: "resource", Key: "k"}: "x"}); err != nil || !got {
		t.Fatal("absent attribute should be an ordinary false leaf", got, err)
	}
}
