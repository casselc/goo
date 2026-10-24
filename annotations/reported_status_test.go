package annotations

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
)

func TestReportedStatusIndependentFromDeclaredOutcome(t *testing.T) {
	m, source, seed := operationTransportFixture(t)
	for _, outcome := range []string{"ok", "error", "panic"} {
		var baseline *Conformance
		bases := map[string]bool{}
		for _, value := range []string{"<absent>", "", "Ok", "Error", "Unset", "VendorStatus"} {
			t.Run(outcome+"/"+value, func(t *testing.T) {
				row := seed
				row.Attributes = map[string]string{"goo.operation.name": "checkout.reserve", "goo.operation.role": "server", "goo.operation.outcome": outcome}
				if value != "<absent>" {
					v := value
					row.ReportedStatus = &v
				}
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					raw, _ := io.ReadAll(r.Body)
					if !strings.Contains(string(raw), "StatusCode AS reported_status") {
						t.Error("reported status not bound to the physical source")
					}
					json.NewEncoder(w).Encode(row)
				}))
				defer server.Close()
				live, err := FetchOperations(context.Background(), server.Client(), server.URL, "", "", *m, source, 1, 2)
				if err != nil {
					t.Fatal(err)
				}
				if len(live.Spans) != 1 || live.Spans[0].Outcome != outcome || !reflect.DeepEqual(live.Spans[0].ReportedStatus, row.ReportedStatus) {
					t.Fatalf("lost independent facts: %+v", live.Spans)
				}
				if baseline == nil {
					copy := live.Conformance
					baseline = &copy
				} else if !reflect.DeepEqual(baseline.Operations, live.Conformance.Operations) {
					t.Fatal("reported status changed declared accounting")
				}
				if bases[live.Conformance.Basis] {
					t.Fatal("source status did not bind evidence basis")
				}
				bases[live.Conformance.Basis] = true
				raw, _ := json.Marshal(live.Spans[0])
				var projected map[string]any
				json.Unmarshal(raw, &projected)
				if value == "<absent>" {
					if _, ok := projected["reportedStatus"]; ok {
						t.Fatal("legacy missing status was filled")
					}
				} else if projected["reportedStatus"] != value {
					t.Fatal("empty or unknown status lost")
				}
			})
		}
	}
	row := seed
	value := "Error"
	row.ReportedStatus = &value
	if _, _, err := m.ObserveRows(source, 1, 2, []OperationRow{seed, row}); err == nil {
		t.Fatal("conflicting reported statuses collapsed into one identity")
	}
}

func TestLegacyOperationSpanProjectionUnchanged(t *testing.T) {
	raw, err := json.Marshal(OperationSpan{Outcome: "error"})
	if err != nil {
		t.Fatal(err)
	}
	const old = `{"events":null,"traceID":"","spanID":"","parentSpanID":"","startNS":"","durationNS":"","operation":"","role":"","outcome":"error","attributes":null}`
	if string(raw) != old {
		t.Fatalf("legacy projection changed: %s", raw)
	}
	var roundtrip OperationSpan
	if err := json.Unmarshal([]byte(old), &roundtrip); err != nil || roundtrip.ReportedStatus != nil {
		t.Fatal("legacy evidence gained reported status", err)
	}
}
