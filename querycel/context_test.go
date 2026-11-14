package querycel

import (
	"errors"
	"github.com/casselc/goo/exploration"
	"testing"
)

func TestRecordContextValidatedIdentity(t *testing.T) {
	base := planningBase()
	context, err := ContextFor(base)
	if err != nil {
		t.Fatal(err)
	}
	request, err := NewPlanRequest(base, `log.body == "x"`, "slot", "mount", "1")
	if err != nil {
		t.Fatal(err)
	}
	if context.BaseIdentity != request.Binding.BaseIdentity || context.Operation != RecordOperation || context.Library != LibraryVersion || context.Compiler != CompilerVersion || context.Target != RecordTarget || context.Profile != RecordProfile {
		t.Fatalf("context differs from planner: %+v", context)
	}
	base.Limit++
	changed, err := ContextFor(base)
	if err != nil || changed.BaseIdentity == context.BaseIdentity {
		t.Fatalf("base change lost: %v", err)
	}
}
func TestRecordContextRefusesWithoutPartialIdentity(t *testing.T) {
	for _, base := range []exploration.Query{{}, {Signal: "logs", Environment: "prod", FromMS: 1000, ToMS: 61000, Limit: 20, Where: &exploration.Predicate{All: []exploration.Predicate{}}}} {
		context, err := ContextFor(base)
		var diagnostic *Diagnostic
		if context != nil || !errors.As(err, &diagnostic) || diagnostic.Path != "base" {
			t.Fatalf("partial or untyped context: %+v %v", context, err)
		}
	}
}
