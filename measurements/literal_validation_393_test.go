package measurements

import (
	"strings"
	"testing"
	"time"
)

func TestEvaluateRejectsMalformedLiteralBeforeClone393(t *testing.T) {
	malformed := string([]byte{'a', 0xff, 'b'})
	d := gauge319(selectors319()[0])
	d.Source.Attributes = map[string]string{"literal": malformed}
	q := Question{Definition: d, FromMS: 100000, ToMS: 160000}
	if _, err := Compile(q); err == nil {
		t.Fatal("malformed typed literal compiled")
	}
	if _, err := Evaluate(q, nil, time.UnixMilli(160000)); err == nil {
		t.Fatal("evaluation repaired malformed literal instead of rejecting it")
	}
	if got := q.Definition.Source.Attributes["literal"]; got != malformed {
		t.Fatal("evaluation changed caller literal bytes")
	}
	q.Definition.Source.Attributes["literal"] = strings.ToValidUTF8(malformed, "�")
	if _, err := Evaluate(q, nil, time.UnixMilli(160000)); err != nil {
		t.Fatalf("valid replacement-character literal was refused: %v", err)
	}
}
