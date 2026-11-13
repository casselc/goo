package capturedrates418

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"os/exec"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/casselc/goo/exploration"
	"hegel.dev/go/hegel"
)

type example struct {
	Name          string                      `json:"name"`
	Comparison    *exploration.ViewComparison `json:"comparison"`
	BaselineRate  string                      `json:"baselineRate"`
	SelectionRate string                      `json:"selectionRate"`
	Difference    *string                     `json:"difference"`
}

func rate(count string, width int64) string {
	n, _ := new(big.Int).SetString(count, 10)
	return new(big.Rat).SetFrac(new(big.Int).Mul(n, big.NewInt(1000)), big.NewInt(width)).RatString()
}
func cohort(signal, count string, start, width int64) exploration.Cohort {
	service := "worker-雪"
	state := "observed"
	groups := []exploration.ValueCount{{Value: &service, Count: count}}
	if count == "0" {
		state = "no_data"
		groups = []exploration.ValueCount{}
	}
	return exploration.Cohort{Question: exploration.Query{Signal: signal, Environment: "ops-'雪", FromMS: start, ToMS: start + width, Limit: 100, GroupBy: []exploration.Field{{Kind: "column", Key: "service"}}}, Basis: fmt.Sprintf("basis-%d-%s", start, count), CapturedAt: time.UnixMilli(start + width + 1).UTC().Format(time.RFC3339Nano), Count: count, State: state, Groups: groups}
}
func makeExample(tc interface{ Fatalf(string, ...any) }, name string, a, b exploration.Cohort) example {
	before, _ := json.Marshal(exploration.ComparisonRequest{Baseline: a, Selection: b})
	result, err := exploration.CompareCohorts(exploration.ComparisonRequest{Baseline: a, Selection: b})
	if err != nil || result.ViewComparison == nil {
		tc.Fatalf("%s compare: %v / %+v", name, err, result)
	}
	after, _ := json.Marshal(exploration.ComparisonRequest{Baseline: a, Selection: b})
	if !bytes.Equal(before, after) {
		tc.Fatalf("%s mutated input", name)
	}
	c := result.ViewComparison
	ar, br := rate(a.Count, a.Question.ToMS-a.Question.FromMS), rate(b.Count, b.Question.ToMS-b.Question.FromMS)
	for i, side := range []*exploration.ViewDataset{c.Baseline, c.Selection} {
		wanted := []string{ar, br}[i]
		row := side.Rows[0]
		v := row.Values[exploration.RecordRateKey]
		if v == nil || *v != wanted {
			tc.Fatalf("%s side%d rate=%v want%s", name, i, v, wanted)
		}
		if row.Cells[exploration.RecordRateKey].State != "derived" {
			tc.Fatalf("%s known rate lost", name)
		}
		if side.Origin != "browser-held" || side.Completeness != nil || side.Quality.Extent != "unknown" {
			tc.Fatalf("%s held comparison gained source attestation", name)
		}
	}
	change := c.Changes[exploration.RecordRateKey]
	var delta *string
	if a.State == "changed_between_reads" || b.State == "changed_between_reads" {
		if c.State != "diagnostic" || change.State != "unavailable" || change.Value != nil || change.Reason == "" {
			tc.Fatalf("%s diagnostic manufactured change", name)
		}
	} else {
		x, _ := new(big.Rat).SetString(ar)
		y, _ := new(big.Rat).SetString(br)
		wanted := new(big.Rat).Sub(y, x).RatString()
		if change.State != "derived" || change.Value == nil || *change.Value != wanted {
			tc.Fatalf("%s change=%+v want%s", name, change, wanted)
		}
		delta = &wanted
	}
	return example{name, c, ar, br, delta}
}
func TestGeneratedCapturedRatesUseEachExactWindow(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		seed := hegel.Draw(tc, hegel.Integers(1, 1000000))
		n := new(big.Int).Add(new(big.Int).Lsh(big.NewInt(1), 53), big.NewInt(int64(seed)))
		twice := new(big.Int).Mul(new(big.Int).Set(n), big.NewInt(2))
		width := int64(1001 + seed%60000)
		start := int64(1800000000000 + seed*1000)
		var cases []example
		for _, signal := range []string{"logs", "spans"} {
			a, b := cohort(signal, n.String(), start, width), cohort(signal, twice.String(), start+width+31, width*2)
			cases = append(cases, makeExample(tc, signal+"/equal rate unequal count/window", a, b))
			b.Count = new(big.Int).Add(twice, big.NewInt(int64(seed%97+1))).String()
			b.Groups[0].Count = b.Count
			b.Question.ToMS += 17
			cases = append(cases, makeExample(tc, signal+"/independent rational difference", a, b))
			zero := cohort(signal, "0", start, width)
			cases = append(cases, makeExample(tc, signal+"/known zero", zero, b), makeExample(tc, signal+"/both zero", zero, cohort(signal, "0", start+width, width*2)))
			diagnostic := a
			diagnostic.State = "changed_between_reads"
			diagnostic.Groups = append([]exploration.ValueCount(nil), a.Groups...)
			diagnostic.Groups[0].Count = "1"
			cases = append(cases, makeExample(tc, signal+"/independent inconsistent read", diagnostic, b))
		}
		raw, err := json.Marshal(cases)
		if err != nil {
			tc.Fatalf("marshal %v", err)
		}
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		cmd := exec.CommandContext(ctx, "node", "bridge.mjs")
		cmd.Stdin = bytes.NewReader(raw)
		var stderr bytes.Buffer
		cmd.Stderr = &stderr
		out, err := cmd.Output()
		if err != nil {
			tc.Fatalf("actual JS comparison: %v %s", err, stderr.String())
		}
		var results []string
		if err = json.Unmarshal(out, &results); err != nil || len(results) != len(cases) {
			tc.Fatalf("bridge dropped cases %s %v", out, err)
		}
		for i, v := range results {
			if !reflect.DeepEqual(v, cases[i].Name) {
				tc.Fatalf("bridge changed ordering")
			}
		}
	}, hegel.WithTestCases(32), hegel.WithSeed(418), hegel.WithDatabase(filepath.Join(t.TempDir(), "hegel")))
}
