package verification

import (
	"encoding/json"
	"fmt"
	"strconv"
	"testing"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/identity"
	"hegel.dev/go/hegel"
)

// The JavaScript side exercises the actual UI compatibility helper and the real
// canonical comparison validator; neither side computes a substitute delta.
const authoredComparisonParityJS = `import assert from 'node:assert/strict';import fs from 'node:fs';import * as C from '../ui/captured-count-comparison.mjs';import {validateComparison} from '../ui/view-comparison.mjs';const {a,b,accepted,comparison}=JSON.parse(fs.readFileSync(0,'utf8'));const before=JSON.stringify([a,b]);let ok=true;try{C.compatible(a,b)}catch{ok=false}assert.equal(ok,accepted);assert.equal(JSON.stringify([a,b]),before);assert.equal(validateComparison(comparison).ok,true);`

func TestHegelAuthoredComparisonCompatibilityAndCanonicalCounts(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		seed := hegel.Draw(tc, hegel.Integers(0, 1000))
		variant := hegel.Draw(tc, hegel.Integers(0, 10))
		countA := hegel.Draw(tc, hegel.Integers(0, 100000))
		countB := hegel.Draw(tc, hegel.Integers(0, 100000))
		if seed%7 == 0 {
			countA = 0
		}
		if seed%11 == 0 {
			countB = 0
		}
		a := exploration.Query{Signal: []string{"logs", "spans"}[seed%2], Environment: "prod", FromMS: 1000, ToMS: 61000, Limit: 1, Offset: 0, GroupBy: []exploration.Field{{Kind: "column", Key: "service"}}, BucketMS: 10000}
		if seed%4 != 0 {
			a.Environment = ""
			a.EnvironmentSelector = &identity.EnvironmentSelector{Version: identity.EnvironmentSelectorVersion, Policy: "effective", Scope: identity.EnvironmentScope{Mode: []string{"named", "empty", "unreported"}[seed%3]}}
			if a.EnvironmentSelector.Scope.Mode == "named" {
				a.EnvironmentSelector.Scope.Value = "prod"
			}
		}
		if seed%5 == 0 {
			a.Environment = ""
			a.EnvironmentSelector = &identity.EnvironmentSelector{Version: identity.EnvironmentSelectorVersion, Policy: "current-only", Scope: identity.EnvironmentScope{Mode: "named", Value: "prod"}}
		}
		raw, _ := json.Marshal(a)
		var b exploration.Query
		json.Unmarshal(raw, &b)
		b.FromMS = 121000
		b.ToMS = 241000
		b.Limit = 20
		b.Offset = seed % 1001
		accepted := true
		switch variant {
		case 1:
			b.Where = &exploration.Predicate{Field: &exploration.Field{Kind: "attribute", Key: "empty"}, Op: "eq", Value: ""}
			accepted = false
		case 2:
			b.Where = &exploration.Predicate{Field: &exploration.Field{Kind: "attribute", Key: "empty"}, Op: "missing"}
			accepted = false
		case 3:
			b.GroupBy = []exploration.Field{{Kind: "attribute", Key: "team"}}
			accepted = false
		case 4:
			b.BucketMS = 1000
			accepted = false
		case 5:
			b.Signal = map[string]string{"logs": "spans", "spans": "logs"}[a.Signal]
			accepted = false
		case 6:
			b.SpanStatus = true
			accepted = false
		case 7:
			b.GroupBy = append(b.GroupBy, exploration.Field{Kind: "column", Key: "scope"})
			accepted = false
		case 8:
			b.Environment = "other"
			b.EnvironmentSelector = nil
			accepted = false
		case 9:
			// Equivalent legacy/current-only is accepted only for a current-only named source.
			if a.Environment != "" || a.EnvironmentSelector.Policy == "current-only" {
				b.Environment = "prod"
				b.EnvironmentSelector = nil
			} else {
				b.Environment = "prod"
				b.EnvironmentSelector = nil
				accepted = false
			}
		case 10:
			if a.Signal == "spans" {
				b.Summary = "span-duration"
			} else {
				b.GroupBy = nil
			}
			accepted = false
		}
		before, _ := json.Marshal([]exploration.Query{a, b})
		got := exploration.CompatibleCapturedCountQueries(a, b) == nil
		after, _ := json.Marshal([]exploration.Query{a, b})
		if string(before) != string(after) {
			tc.Fatal("compatibility rewrote source question")
		}
		if got != accepted {
			tc.Fatalf("compatibility variant%d got%v want%v: %#v %#v", variant, got, accepted, a, b)
		}
		state := func(n int) string {
			if n == 0 {
				return "no_data"
			}
			return "observed"
		}
		cohort := func(q exploration.Query, n int, basis string) exploration.Cohort {
			groups := []exploration.ValueCount{}
			if n > 0 {
				key := "api"
				groups = append(groups, exploration.ValueCount{Value: &key, Count: strconv.Itoa(n)})
			}
			return exploration.Cohort{Question: q, Count: strconv.Itoa(n), State: state(n), Basis: basis, CapturedAt: "2026-10-31T00:00:00Z", Groups: groups}
		}
		selected := a
		selected.FromMS = 121000
		selected.ToMS = 241000
		baseline, selection := cohort(a, countA, "baseline"), cohort(selected, countB, "selection")
		diagnostic := seed%13 == 0
		if diagnostic {
			selection.State = "changed_between_reads"
		}
		r, err := exploration.CompareCohorts(exploration.ComparisonRequest{Baseline: baseline, Selection: selection})
		if err != nil || r.ViewComparison == nil {
			tc.Fatalf("canonical comparison unavailable: %v", err)
		}
		change := r.ViewComparison.Changes["count"]
		if diagnostic {
			if change.State != "unavailable" || change.Value != nil || r.ViewComparison.State != "diagnostic" {
				tc.Fatalf("disagreement became exactchange: %#v", change)
			}
		} else if change.State != "derived" || change.Value == nil || *change.Value != fmt.Sprint(countB-countA) {
			tc.Fatalf("exact count change: %#v", change)
		}
		if err := runNativeHistogramJS(map[string]any{"a": a, "b": b, "accepted": accepted, "comparison": r.ViewComparison}, authoredComparisonParityJS); err != nil {
			tc.Fatal(err)
		}
	}, options(100)...)
}
