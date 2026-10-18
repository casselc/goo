package exploration

import (
	"reflect"
	"testing"
)

func TestSurroundingQuestionClipsAndOmitsOriginalSelectors(t *testing.T) {
	q := Query{Signal: "logs", Environment: "prod", FromMS: 1000, ToMS: 100000, Limit: 1, Where: &Predicate{Field: &Field{"column", "body"}, Op: "contains", Value: "error"}}
	got, e := SurroundingQuestion(q, "1500123456", "worker", 30000, 100)
	if e != nil {
		t.Fatal(e)
	}
	want := Query{Signal: "logs", Environment: "prod", FromMS: 1000, ToMS: 31501, Limit: 100, Offset: 100, Where: &Predicate{Field: &Field{"column", "service"}, Op: "eq", Value: "worker"}}
	if !reflect.DeepEqual(got, want) || q.Where.Field.Key != "body" {
		t.Fatal(got)
	}
	for _, ns := range []string{"bad", "999999999", "100000000000"} {
		if _, e := SurroundingQuestion(q, ns, "worker", 30000, 0); e == nil {
			t.Fatal("out-of-scope anchor accepted", ns)
		}
	}
	if _, e := SurroundingQuestion(q, "1500123456", "worker", 300001, 0); e == nil {
		t.Fatal("unbounded radius")
	}
}
