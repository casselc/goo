package measurements

import (
	"encoding/json"
	"github.com/casselc/goo/identity"
	"strings"
	"testing"
	"time"
)

func TestNamespaceBindingIsolation(t *testing.T) {
	for _, selector := range []*identity.Namespace{{Mode: "missing"}, {Mode: "value"}, {Mode: "value", Value: "payments'"}} {
		d := definition(t, "queue-lag")
		d.Source.Namespace = selector
		q := Question{Definition: d, FromMS: 100000, ToMS: 220000}
		p, e := Compile(q)
		if e != nil || !strings.Contains(p.SQL, "mapContains") {
			t.Fatal(p, e)
		}
		row := gaugeRow(d, "one", 210000, "7")
		if selector.Mode == "value" {
			row.Resource[identity.NamespaceKey] = selector.Value
		}
		if _, e = Evaluate(q, []json.RawMessage{raw(row)}, time.UnixMilli(230000)); e != nil {
			t.Fatal(e)
		}
		row.Resource[identity.NamespaceKey] = "foreign"
		if _, e = Evaluate(q, []json.RawMessage{raw(row)}, time.UnixMilli(230000)); e == nil {
			t.Fatal("foreign namespace accepted")
		}
	}
	d := definition(t, "queue-lag")
	b, _ := json.Marshal(d)
	if strings.Contains(string(b), "namespace") {
		t.Fatal("legacy binding changed")
	}
}
