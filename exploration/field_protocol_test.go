package exploration

import (
	"encoding/json"
	"os"
	"reflect"
	"testing"
)

// These frozen wire examples are also read by the browser presentation tests.
func TestRecordFieldProtocolFixtures(t *testing.T) {
	raw, err := os.ReadFile("../testdata/protocol/record-fields.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixtures struct {
		Version int
		Cases   []struct {
			Name      string
			Predicate Predicate
			Canonical json.RawMessage
			Observed  []struct {
				Field Field
				Value string
			}
			Match bool
		}
	}
	if err = json.Unmarshal(raw, &fixtures); err != nil {
		t.Fatal(err)
	}
	if fixtures.Version != 1 || len(fixtures.Cases) != 10 {
		t.Fatal("unexpected protocol fixture version/coverage")
	}
	for _, c := range fixtures.Cases {
		t.Run(c.Name, func(t *testing.T) {
			q := Query{Signal: "spans", Environment: "fixture", FromMS: 1000, ToMS: 2000, Limit: 100, Where: &c.Predicate}
			if err := q.Validate(); err != nil {
				t.Fatal(err)
			}
			values := map[Field]string{}
			for _, v := range c.Observed {
				values[v.Field] = v.Value
			}
			if got := Match(&c.Predicate, values); got != c.Match {
				t.Fatalf("match=%v, want %v", got, c.Match)
			}
			actual, err := json.Marshal(c.Predicate)
			if err != nil {
				t.Fatal(err)
			}
			var a, b any
			json.Unmarshal(actual, &a)
			json.Unmarshal(c.Canonical, &b)
			if !reflect.DeepEqual(a, b) {
				t.Fatalf("canonical protocol changed: %s", actual)
			}
		})
	}
}
