package annotations

import "testing"

func TestScopedSubjectEvents(t *testing.T) {
	m, err := ScanDir("../examples/operation-app")
	if err != nil {
		t.Fatal(err)
	}
	source := OperationSource{Service: "checkout", Environment: "test", SubjectNamespace: "reservation"}
	attrs := map[string]string{"goo.event.name": "reservation.accepted", "goo.subject.namespace": "reservation", "goo.subject.tenant": "tenant-a", "goo.subject.id": "id"}
	row := OperationRow{TimestampNS: "100", DurationNS: "10", EventTimes: []string{"101"}, EventNames: []string{"reservation.accepted"}, EventAttributes: []map[string]string{attrs}}
	events, err := m.SubjectEvents(source, row)
	if err != nil || len(events) != 1 || events[0].Subject == nil || events[0].Subject.Tenant != "tenant-a" || events[0].Subject.Service != "checkout" {
		t.Fatalf("%+v %v", events, err)
	}
	for _, key := range []string{"goo.subject.namespace", "goo.subject.tenant", "goo.subject.id", "goo.event.name"} {
		saved := attrs[key]
		delete(attrs, key)
		events, err = m.SubjectEvents(source, row)
		if err != nil || events[0].Subject != nil || events[0].Diagnostic == "" {
			t.Fatalf("silently linked missing %s", key)
		}
		attrs[key] = saved
	}
	unbound := source
	unbound.SubjectNamespace = ""
	events, err = m.SubjectEvents(unbound, row)
	if err != nil || events[0].Subject != nil {
		t.Fatal("linked unbound subject")
	}
	row.EventTimes = []string{"111"}
	if _, err = m.SubjectEvents(source, row); err == nil {
		t.Fatal("accepted event outside span")
	}
	row.EventTimes = nil
	if _, err = m.SubjectEvents(source, row); err == nil {
		t.Fatal("accepted mismatched event arrays")
	}
}
