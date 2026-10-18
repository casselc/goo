package exploration

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"testing"
)

func testPackage(t *testing.T) TelemetryPackage {
	t.Helper()
	q := basic()
	q.GroupBy = []Field{{Kind: "attribute", Key: "custom.region"}}
	q.Where = &Predicate{Field: &Field{Kind: "column", Key: "service"}, Op: "eq", Value: "python-indexer"}
	p, e := DraftTelemetryPackage(q, "index-jobs", "Index jobs", 5)
	if e != nil {
		t.Fatal(e)
	}
	return p
}
func TestTelemetryPackageRoundtripReusesSelectorsNotPagesOrTimes(t *testing.T) {
	p := testPackage(t)
	plan, e := PlanTelemetryPackage(p, 1000, 61000, &AnnotationTarget{"index.run", "consumer"})
	if e != nil {
		t.Fatal(e)
	}
	again, e := ReadTelemetryPackage(plan.YAML)
	if e != nil || !reflect.DeepEqual(p, again) {
		t.Fatal(e, plan.YAML)
	}
	other, e := PlanTelemetryPackage(again, 9999, 69999, nil)
	if e != nil || other.Revision != plan.Revision || other.Plan.Count != plan.Plan.Count {
		t.Fatal("preview changed definition", e)
	}
	if strings.Contains(plan.Plan.Rows, "python-indexer") || !strings.Contains(plan.Handoff.GoExample, "//goo:operation name:index.run role:consumer") {
		t.Fatal("unsafe parameter or annotation")
	}
	q := basic()
	q.Offset = 100
	q.GroupBy = p.GroupBy
	q.Where = p.Where
	draft, e := DraftTelemetryPackage(q, p.Name, p.Title, 5)
	if e != nil {
		t.Fatal(e)
	}
	q.Where.Value = "mutated"
	if draft.Where.Value == "mutated" || draft.Question(1000, 61000).Offset != 0 {
		t.Fatal("caller or page leaked into definition")
	}
}
func TestTelemetryPackageRejectsUnsafeAndUnsupportedContracts(t *testing.T) {
	p := testPackage(t)
	plan, _ := PlanTelemetryPackage(p, 1000, 61000, nil)
	for _, text := range []string{plan.YAML + "sql: SELECT 1\n", plan.YAML + "---\nname: two\n", strings.Replace(plan.YAML, "name: index-jobs", "name: &alias index-jobs", 1) + "title: *alias\n", strings.Replace(plan.YAML, "captured-record-count", "business-success-ratio", 1), strings.Replace(plan.YAML, "clickhouse.otel/v1", "arbitrary.sql/v1", 1), strings.Replace(plan.YAML, "groupBy:", "groupBy: !!merge", 1)} {
		if _, e := ReadTelemetryPackage(text); e == nil {
			t.Fatalf("accepted invalid YAML %s", text)
		}
	}
	if _, e := PlanTelemetryPackage(p, 1000, 61000, &AnnotationTarget{"index.run", "internal|server"}); e == nil {
		t.Fatal("invalid role accepted")
	}
	if _, e := ReadTelemetryPackage(strings.Repeat("x", PackageBytes+1)); e == nil {
		t.Fatal("oversized package accepted")
	}
}
func TestTelemetryPackageImmutableConcurrentStoreAndTampering(t *testing.T) {
	store := PackageStore{Dir: t.TempDir()}
	p := testPackage(t)
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, e := store.Save(p); e != nil {
				t.Error(e)
			}
		}()
	}
	wg.Wait()
	first, e := store.Save(p)
	if e != nil {
		t.Fatal(e)
	}
	p.Title = "Changed title"
	next, e := store.Save(p)
	if e != nil || first.Revision == next.Revision {
		t.Fatal(e)
	}
	old, e := store.Load(first.Revision)
	if e != nil || old.Definition.Title != "Index jobs" {
		t.Fatal("old revision mutated", e)
	}
	list, e := store.List()
	if e != nil || len(list) != 2 {
		t.Fatal(e, len(list))
	}
	if _, e := store.Load("../secret"); e == nil {
		t.Fatal("path traversal accepted")
	}
	if e = os.WriteFile(filepath.Join(store.Dir, first.Revision+".json"), []byte(`{}`), 0600); e != nil {
		t.Fatal(e)
	}
	if _, e := store.Load(first.Revision); e == nil {
		t.Fatal("corruption accepted")
	}
}
