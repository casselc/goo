package verification

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/identity"
	"hegel.dev/go/hegel"
)

// Vary multiplicity, signal and effective environment while replaying the same
// publication after bank expiry. The store must retain the original clock/bytes.
func TestKeptObservationProperties(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		count := hegel.Draw(tc, hegel.Integers(1, 20))
		mode := hegel.Draw(tc, hegel.Integers(0, 2))
		spans := hegel.Draw(tc, hegel.Integers(0, 1)) == 1
		now := time.UnixMilli(200000).UTC()
		bank := exploration.NewTailMembershipSnapshots(func() time.Time { return now })
		scope := identity.EnvironmentScope{Mode: "unreported"}
		resource := map[string]string{}
		switch mode {
		case 1:
			scope.Mode = "empty"
			resource["deployment.environment.name"] = ""
			resource["deployment.environment"] = "ignored"
		case 2:
			scope.Mode = "named"
			scope.Value = "prod"
			resource["deployment.environment"] = "prod"
		}
		env, _ := identity.EffectiveEnvironment(scope)
		dto, _ := env.Selector()
		signal, field := "logs", "Body"
		if spans {
			signal, field = "spans", "SpanName"
		}
		wire, _ := json.Marshal(map[string]any{"timestamp_ns": "150000000000", field: "same", "ResourceAttributes": resource, "exact": json.Number("18446744073709551615")})
		raw := make([]json.RawMessage, count)
		for i := range raw {
			raw[i] = wire
		}
		tracker := exploration.TailTracker{}
		rows, e := tracker.Observe(raw)
		if e != nil {
			t.Fatal(e)
		}
		selector := exploration.TailSelector{Signal: signal, EnvironmentSelector: &dto}
		connection := strings.Repeat("a", 64)
		cap, e := bank.Freeze(exploration.TailMembershipBatch{Connection: connection, Selector: selector, FromMS: 100000, ToMS: 160000, ObservedAt: now, Rows: rows})
		if e != nil {
			t.Fatal(e)
		}
		in := exploration.KeepTailIntent{Version: exploration.KeepTailVersion, RequestID: fmt.Sprintf("%064x", count*10+mode+1), Connection: connection, Selector: selector}
		for _, row := range rows {
			in.Refs = append(in.Refs, exploration.TailMembershipRef{Token: cap.Token, Key: row.Key})
		}
		store := exploration.KeptObservationStore{Dir: t.TempDir()}
		v, e := store.Keep(context.Background(), bank, in)
		if e != nil || len(v.Members) != count {
			t.Fatal("multiplicity", e)
		}
		keys := map[string]bool{}
		for _, m := range v.Members {
			if keys[m.Key] || !strings.Contains(string(m.Record), "18446744073709551615") {
				t.Fatal("lost exact occurrence")
			}
			keys[m.Key] = true
		}
		before, _ := json.Marshal(v)
		now = now.Add(3 * time.Minute)
		again, e := store.Keep(context.Background(), nil, in)
		if e != nil {
			t.Fatal(e)
		}
		after, _ := json.Marshal(again)
		if string(before) != string(after) {
			t.Fatal("retry replaced original evidence")
		}
		foreign := in
		foreign.Connection = strings.Repeat("b", 64)
		if _, e = store.Keep(context.Background(), bank, foreign); e == nil {
			t.Fatal("changed intent reused receipt")
		}
	}, options(50)...)
}
