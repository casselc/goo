package annotations

import "testing"

func TestObservedOperations(t *testing.T) {
	m, err := ScanDir("../verification/weaving/example")
	if err != nil {
		t.Fatal(err)
	}
	s := Observation{Resource: "service/a", Scope: m.Scope, ScopeVersion: m.ScopeVersion, SchemaURL: m.SchemaURL, TraceID: "11111111111111111111111111111111", SpanID: "1111111111111111", Name: "checkout.reserve", Role: "server", Outcome: "error", DurationNS: 9007199254740993}
	r, err := m.Observe("service/a", "fixture", []Observation{s, s})
	if err != nil {
		t.Fatal(err)
	}
	if len(r.Operations) != 2 || r.Operations[0].Status != "unobserved" || r.Operations[1].Calls != 1 || r.Operations[1].Errors != 1 || r.Operations[1].DurationsNS[0] != s.DurationNS {
		t.Fatalf("incorrect overview: %+v", r)
	}
	for _, change := range []func(*Observation){
		func(v *Observation) { v.Resource = "other" }, func(v *Observation) { v.ScopeVersion = "unknown" }, func(v *Observation) { v.SchemaURL = "" }, func(v *Observation) { v.Role = "client" }, func(v *Observation) { v.Outcome = "missing" }, func(v *Observation) { v.TraceID = "00000000000000000000000000000000" },
	} {
		bad := s
		change(&bad)
		if _, err = m.Observe("service/a", "fixture", []Observation{bad}); err == nil {
			t.Fatal("accepted incompatible evidence")
		}
	}
	bad := s
	bad.DurationNS++
	if _, err = m.Observe("service/a", "fixture", []Observation{s, bad}); err == nil {
		t.Fatal("accepted conflicting duplicate")
	}
}
