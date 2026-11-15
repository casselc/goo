package measurements

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestRawInstrument448ExactFields(t *testing.T) {
	for _, tc := range []struct {
		name, input   string
		newInstrument func() json.Unmarshaler
	}{
		{"gauge", `{"kind":"gauge","scope":"","version":"","schemaURL":"","metric":"queue.depth","unit":""}`, func() json.Unmarshaler { return &RawGaugeInstrument{} }},
		{"sum", `{"kind":"counter","scope":"","version":"","schemaURL":"","metric":"jobs.total","unit":"","temporality":0,"monotonic":false}`, func() json.Unmarshaler { return &RawSumInstrument{} }},
		{"histogram", `{"kind":"histogram","scope":"","version":"","schemaURL":"","metric":"work.payload","unit":"","temporality":0}`, func() json.Unmarshaler { return &RawHistogramInstrument{} }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			value := tc.newInstrument()
			if err := json.Unmarshal([]byte(tc.input), value); err != nil {
				t.Fatal(err)
			}
			raw, err := json.Marshal(value)
			if err != nil {
				t.Fatal(err)
			}
			var before, after any
			if json.Unmarshal([]byte(tc.input), &before) != nil || json.Unmarshal(raw, &after) != nil {
				t.Fatal("round trip did not produce JSON")
			}
			canonicalBefore, _ := json.Marshal(before)
			canonicalAfter, _ := json.Marshal(after)
			if string(canonicalBefore) != string(canonicalAfter) {
				t.Fatalf("exact empty/zero/false values changed: %s", raw)
			}
			for name, input := range map[string]string{
				"case-alias-after":    strings.TrimSuffix(tc.input, "}") + `,"Scope":"overridden"}`,
				"case-alias-before":   `{"Scope":"overridden",` + strings.TrimPrefix(tc.input, "{"),
				"wrong-case-required": strings.Replace(tc.input, `"scope"`, `"Scope"`, 1),
				"schema-case-alias":   strings.TrimSuffix(tc.input, "}") + `,"schemaUrl":"overridden"}`,
				"unknown":             strings.TrimSuffix(tc.input, "}") + `,"cursor":"retained"}`,
				"null-required":       strings.Replace(tc.input, `"scope":""`, `"scope":null`, 1),
				"missing-required":    strings.Replace(tc.input, `"scope":"",`, ``, 1),
				"wrong-type":          strings.Replace(tc.input, `"scope":""`, `"scope":false`, 1),
				"trailing-object":     tc.input + `{}`,
				"array":               `[]`, "null": `null`,
			} {
				t.Run(name, func(t *testing.T) {
					if err := json.Unmarshal([]byte(input), tc.newInstrument()); err == nil {
						t.Fatalf("accepted non-exact source tuple: %s", input)
					}
				})
			}
		})
	}
}
