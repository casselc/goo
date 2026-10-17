package verification

import (
	"fmt"
	"github.com/casselc/goo/annotations"
	"hegel.dev/go/hegel"
	"testing"
)

func TestHegelOperationObservationCopies(t *testing.T) {
	manifest, err := annotations.ScanDir("../examples/operation-app")
	if err != nil {
		t.Fatal(err)
	}
	hegel.Test(t, func(tc *hegel.T) {
		count := hegel.Draw(tc, hegel.Integers(0, 30))
		copies := hegel.Draw(tc, hegel.Integers(1, 5))
		offset := hegel.Draw(tc, hegel.Integers(0, 30))
		rows := []annotations.Observation{}
		errors, panics := uint64(0), uint64(0)
		for i := 0; i < count; i++ {
			outcome := []string{"ok", "error", "panic"}[(i+offset)%3]
			if outcome == "error" {
				errors++
			}
			if outcome == "panic" {
				panics++
			}
			row := annotations.Observation{Resource: "test", Scope: manifest.Scope, ScopeVersion: manifest.ScopeVersion, SchemaURL: manifest.SchemaURL, TraceID: fmt.Sprintf("%032x", i+1), SpanID: fmt.Sprintf("%016x", i+1), Name: "checkout.reserve", Role: "server", Outcome: outcome, DurationNS: uint64(i + 1)}
			for j := 0; j < copies; j++ {
				rows = append(rows, row)
			}
		}
		a, err := manifest.Observe("test", "generated", rows)
		if err != nil {
			tc.Fatalf("%v", err)
		}
		for i, j := 0, len(rows)-1; i < j; i, j = i+1, j-1 {
			rows[i], rows[j] = rows[j], rows[i]
		}
		b, err := manifest.Observe("test", "generated", rows)
		if err != nil {
			tc.Fatalf("%v", err)
		}
		if fmt.Sprint(a) != fmt.Sprint(b) {
			tc.Fatalf("order changed report")
		}
		row := a.Operations[1]
		if row.Calls != uint64(count) || row.Errors != errors || row.Panics != panics || len(row.DurationsNS) != count {
			tc.Fatalf("copies changed invocation counts: %+v", row)
		}
	}, options(200)...)
}
