package status419

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/casselc/goo/exploration"
	"hegel.dev/go/hegel"
)

type statusExample struct {
	Name       string                      `json:"name"`
	Comparison *exploration.ViewComparison `json:"comparison"`
	Expected   [2]map[string]*string       `json:"expected"`
	Changes    map[string]*string          `json:"changes"`
}

func pointer(s string) *string { return &s }
func integer(s string) *big.Int {
	n, ok := new(big.Int).SetString(s, 10)
	if !ok {
		panic(s)
	}
	return n
}
func statusValues(s exploration.SpanStatusSummary) map[string]*string {
	out := map[string]*string{"count": pointer(s.Count), "status.error": pointer(s.Error), "status.ok": pointer(s.OK), "status.unset": pointer(s.Unset), "status.unknown": pointer(s.Unknown), "status.error_fraction": nil}
	if s.Count != "0" {
		out["status.error_fraction"] = pointer(new(big.Rat).SetFrac(integer(s.Error), integer(s.Count)).RatString())
	}
	return out
}
func statusCohort(s exploration.SpanStatusSummary, start int64) exploration.Cohort {
	name := "rpc/雪"
	state := "observed"
	groups := []exploration.ValueCount{{Value: &name, Count: s.Count, SpanStatus: &s}}
	if s.Count == "0" {
		state = "no_data"
		groups = []exploration.ValueCount{}
	}
	return exploration.Cohort{Question: exploration.Query{Signal: "spans", Environment: "prod-'雪", FromMS: start, ToMS: start + 60000, Limit: 100, SpanStatus: true, GroupBy: []exploration.Field{{Kind: "attribute", Key: "rpc.method"}}}, Basis: fmt.Sprintf("held-%d-%s", start, s.Count), CapturedAt: time.UnixMilli(start + 61000).UTC().Format(time.RFC3339Nano), Count: s.Count, State: state, Groups: groups, SpanStatus: &s}
}
func statusSummary(parts ...*big.Int) exploration.SpanStatusSummary {
	total := new(big.Int)
	for _, n := range parts {
		total.Add(total, n)
	}
	return exploration.SpanStatusSummary{Count: total.String(), Error: parts[0].String(), OK: parts[1].String(), Unset: parts[2].String(), Unknown: parts[3].String()}
}
func statusCase(tc interface{ Fatalf(string, ...any) }, name string, a, b exploration.Cohort) statusExample {
	input := exploration.ComparisonRequest{Baseline: a, Selection: b}
	before, _ := json.Marshal(input)
	result, err := exploration.CompareCohorts(input)
	after, _ := json.Marshal(input)
	if err != nil || result.ViewComparison == nil {
		tc.Fatalf("%s compare: %v %+v", name, err, result)
	}
	if !bytes.Equal(before, after) {
		tc.Fatalf("%s mutated held input", name)
	}
	c := result.ViewComparison
	expected := [2]map[string]*string{statusValues(*a.SpanStatus), statusValues(*b.SpanStatus)}
	changes := map[string]*string{}
	for i, side := range []*exploration.ViewDataset{c.Baseline, c.Selection} {
		if side.Origin != "browser-held" || side.Completeness != nil || side.Quality.Extent != "unknown" {
			tc.Fatalf("%s gained completeness", name)
		}
		for key, want := range expected[i] {
			got := side.Rows[0].Values[key]
			if got == nil != (want == nil) || got != nil && *got != *want {
				tc.Fatalf("%s side%d %s=%v wanted%v", name, i, key, got, want)
			}
		}
	}
	blocked := a.State == "changed_between_reads" || b.State == "changed_between_reads"
	for key := range expected[0] {
		change := c.Changes[key]
		av, bv := expected[0][key], expected[1][key]
		if blocked || av == nil || bv == nil {
			if change.State != "unavailable" || change.Value != nil || change.Reason == "" {
				tc.Fatalf("%s %s unavailable change fabricated: %+v", name, key, change)
			}
			changes[key] = nil
			continue
		}
		x, _ := new(big.Rat).SetString(*av)
		y, _ := new(big.Rat).SetString(*bv)
		want := new(big.Rat).Sub(y, x).RatString()
		if change.State != "derived" || change.Value == nil || *change.Value != want {
			tc.Fatalf("%s %s change %+v want%s", name, key, change, want)
		}
		changes[key] = pointer(want)
	}
	return statusExample{Name: name, Comparison: c, Expected: expected, Changes: changes}
}
func TestGeneratedRecordedStatusPreservesExactAccounting(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		seed := int64(hegel.Draw(tc, hegel.Integers(1, 1000000)))
		n := new(big.Int).Add(new(big.Int).Lsh(big.NewInt(1), 53), big.NewInt(seed))
		parts := []*big.Int{n, big.NewInt(seed%97 + 1), big.NewInt(seed%29 + 1), big.NewInt(seed%17 + 1)}
		a := statusCohort(statusSummary(parts...), 1800000000000+seed)
		twice := make([]*big.Int, 4)
		for i, v := range parts {
			twice[i] = new(big.Int).Mul(v, big.NewInt(2))
		}
		b := statusCohort(statusSummary(twice...), a.Question.FromMS)
		cases := []statusExample{statusCase(tc, "equal fractions unequal populations", a, b)}
		twice[0].Add(twice[0], big.NewInt(seed%71+1))
		b = statusCohort(statusSummary(twice...), a.Question.FromMS)
		cases = append(cases, statusCase(tc, "independent fraction delta", a, b), statusCase(tc, "reversed signed delta", b, a))
		zero := statusCohort(exploration.SpanStatusSummary{"0", "0", "0", "0", "0"}, a.Question.FromMS)
		cases = append(cases, statusCase(tc, "empty denominator", zero, b), statusCase(tc, "both empty", zero, zero))
		noError := statusCohort(statusSummary(big.NewInt(0), big.NewInt(0), n, big.NewInt(seed%17+1)), a.Question.FromMS)
		allError := statusCohort(statusSummary(n, big.NewInt(0), big.NewInt(0), big.NewInt(0)), a.Question.FromMS)
		cases = append(cases, statusCase(tc, "known zero Error with unset population", noError, allError), statusCase(tc, "negative unit fraction change", allError, noError))
		diagnostic := a
		diagnostic.State = "changed_between_reads"
		cases = append(cases, statusCase(tc, "diagnostic withholding", diagnostic, b))
		raw, err := json.Marshal(cases)
		if err != nil {
			tc.Fatalf("marshal: %v", err)
		}
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		cmd := exec.CommandContext(ctx, "node", "bridge.mjs")
		cmd.Stdin = bytes.NewReader(raw)
		var stderr bytes.Buffer
		cmd.Stderr = &stderr
		out, err := cmd.Output()
		if err != nil {
			tc.Fatalf("JS boundary: %v %s", err, stderr.String())
		}
		var names []string
		if err = json.Unmarshal(out, &names); err != nil || len(names) != len(cases) {
			tc.Fatalf("bridge dropped examples %s %v", out, err)
		}
	}, hegel.WithTestCases(32), hegel.WithSeed(419), hegel.WithDatabase(filepath.Join(t.TempDir(), "hegel")))
}
