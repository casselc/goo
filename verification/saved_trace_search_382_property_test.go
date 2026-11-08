package verification

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/identity"
	"github.com/casselc/goo/workbench"
	"hegel.dev/go/hegel"
)

type searchStepWire382 struct {
	Wire     json.RawMessage `json:"wire"`
	Accepted bool            `json:"accepted"`
}

func searchStep382(seed, count, limit, mode int) workbench.Step {
	selector := identity.EnvironmentSelector{Version: identity.EnvironmentSelectorVersion, Policy: "effective", Scope: identity.EnvironmentScope{Mode: "named", Value: fmt.Sprintf("prod-'雪-%d", seed)}}
	switch mode {
	case 0:
		selector.Policy = "current-only"
	case 2:
		selector.Scope = identity.EnvironmentScope{Mode: "empty"}
	case 3:
		selector.Scope = identity.EnvironmentScope{Mode: "unreported"}
	}
	copied := selector
	request := exploration.TraceSearchRequest{Version: exploration.TraceSearchVersion, EnvironmentSelector: &selector, FromMS: int64(100000 + seed), ToMS: int64(160000 + seed), Limit: limit}
	for i := 0; i < count; i++ {
		request.Conditions = append(request.Conditions, exploration.Predicate{All: []exploration.Predicate{
			{Field: &exploration.Field{Kind: "column", Key: "service"}, Op: "eq", Value: fmt.Sprintf("svc-'雪-%d-%d", seed, i)},
			{Numeric: &exploration.NumericBinding{Field: exploration.Field{Kind: "attribute", Key: fmt.Sprintf("work.bytes.%d", i)}, Type: "uint64", Unit: "By", Parse: "canonical-decimal"}, Op: "gte", Value: strconv.FormatUint(uint64(9007199254740993)+uint64(seed*4+i), 10)},
		}})
	}
	namespace := exploration.Predicate{Field: &exploration.Field{Kind: "resource", Key: "service.namespace"}, Op: "missing"}
	if seed%3 != 0 {
		namespace.Op = "eq"
		if seed%3 == 2 {
			namespace.Value = "team-'雪"
		}
	}
	request.Conditions[0].All = append(request.Conditions[0].All, namespace)
	return workbench.Step{Version: 1, Kind: "trace-search", Scope: workbench.StepScope{EnvironmentSelector: &copied, FromMS: request.FromMS, ToMS: request.ToMS}, View: "Traces", TraceSearch: &request}
}
func searchStepWireAdmission382(raw json.RawMessage) bool {
	var step workbench.Step
	return json.Unmarshal(raw, &step) == nil && step.Validate() == nil
}

const searchStepBatch382 = `import assert from 'node:assert/strict';import fs from 'node:fs';import * as P from '../ui/question-protocol.mjs';const cases=JSON.parse(fs.readFileSync(0,'utf8'));for(const [index,c]of cases.entries()){const before=structuredClone(c.wire);let accepted=false;try{P.validateStep(c.wire);accepted=true}catch{}assert.equal(accepted,c.accepted,'case '+index);assert.deepEqual(c.wire,before,'admission mutated case '+index);if(accepted){const actual=P.readStepLink(P.stepLink(c.wire));assert.deepEqual(actual,c.wire,'link roundtrip '+index);assert.deepEqual(P.captureQuestion(actual),{kind:'trace-search',traceSearch:actual.traceSearch},'typed capture question '+index)}}process.stdout.write(JSON.stringify({cases:cases.length}));`

func TestHegelSavedTraceSearchStep382(t *testing.T) {
	var cases []searchStepWire382
	hegel.Test(t, func(tc *hegel.T) {
		seed := hegel.Draw(tc, hegel.Integers(1, 100000))
		count := hegel.Draw(tc, hegel.Integers(1, 4))
		limit := hegel.Draw(tc, hegel.Integers(1, 100))
		mode := hegel.Draw(tc, hegel.Integers(0, 3))
		step := searchStep382(seed, count, limit, mode)
		before, _ := json.Marshal(step)
		if step.Validate() != nil || !searchStepWireAdmission382(before) {
			tc.Fatal("valid generated search intent refused")
		}
		var decoded workbench.Step
		if json.Unmarshal(before, &decoded) != nil || !reflect.DeepEqual(decoded, step) {
			tc.Fatal("typed search intent JSON roundtrip changed scope/order/numeric tokens")
		}
		store := workbench.Store{Dir: t.TempDir()}
		input := workbench.DocumentInput{Title: "Reusable cross-service trace question", Cards: []workbench.Card{{Kind: "question", Key: strings.Repeat("a", 64), Step: &step, Caption: "Exact fixed search"}}}
		saved, err := store.CreateDocument(input, "", nil)
		if err != nil {
			tc.Fatal(err)
		}
		reopened, err := store.LoadDocument(saved.ID)
		if err != nil || !reflect.DeepEqual(reopened, saved) || !reflect.DeepEqual(reopened.Cards[0].Step, &step) {
			tc.Fatal("exact immutable investigation revision changed search intent", err)
		}
		after, _ := json.Marshal(step)
		if !bytes.Equal(before, after) {
			tc.Fatal("persistence mutated caller intent")
		}
		refused := input
		refused.Cards = append([]workbench.Card(nil), input.Cards...)
		refused.Cards[0].Evidence = []string{strings.Repeat("b", 64)}
		called := false
		retainedCalled := false
		if _, err = store.CreateDocument(refused, "", func(string) error { called = true; return nil }, func(card workbench.Card) error {
			retainedCalled = true
			if !reflect.DeepEqual(card.Step, &step) {
				return fmt.Errorf("search question changed")
			}
			return nil
		}); err != nil || !called || !retainedCalled {
			tc.Fatal("typed search evidence did not use existing capture resolver", err)
		}
		cases = append(cases, searchStepWire382{Wire: before, Accepted: true})
		variants := []func(map[string]any){
			func(x map[string]any) { x["view"] = "Records" },
			func(x map[string]any) { x["scope"].(map[string]any)["toMs"] = float64(step.Scope.ToMS + 1) },
			func(x map[string]any) { x["scope"].(map[string]any)["fromNS"] = "1000000000" },
			func(x map[string]any) { x["record"] = map[string]any{} },
			func(x map[string]any) { x["measurement"] = map[string]any{} },
			func(x map[string]any) { x["traceSearch"] = nil },
			func(x map[string]any) { x["foreign"] = true },
			func(x map[string]any) {
				x["scope"].(map[string]any)["environmentSelector"] = map[string]any{"version": "goo/environment-selector/v1", "policy": "current-only", "scope": map[string]any{"mode": "named", "value": "foreign"}}
			},
			func(x map[string]any) {
				x["traceSearch"].(map[string]any)["conditions"].([]any)[0].(map[string]any)["all"].([]any)[1].(map[string]any)["value"] = float64(9007199254740992)
			},
			func(x map[string]any) {
				x["traceSearch"].(map[string]any)["conditions"].([]any)[0].(map[string]any)["all"].([]any)[0].(map[string]any)["field"].(map[string]any)["key"] = "unsupported-column"
			},
			func(x map[string]any) { x["traceSearch"].(map[string]any)["limit"] = 0 },
			func(x map[string]any) { x["traceSearch"].(map[string]any)["foreign"] = true },
			func(x map[string]any) { x["traceSearch"].(map[string]any)["conditions"] = []any{} },
		}
		for _, mutate := range variants {
			var value map[string]any
			json.Unmarshal(before, &value)
			mutate(value)
			raw, _ := json.Marshal(value)
			if searchStepWireAdmission382(raw) {
				tc.Fatal("changed/mixed search intent admitted", string(raw))
			}
			cases = append(cases, searchStepWire382{Wire: raw})
		}
	}, options(100)...)
	// Each individual condition stays below its 32-node allowance while the
	// request crosses its independent total 64-node boundary.
	for _, extra := range []int{1, 2} {
		step := searchStep382(1, 1, 100, 1)
		step.TraceSearch.Conditions = nil
		for group := 0; group < 3; group++ {
			n := 20
			if group == 0 {
				n += extra
			}
			children := make([]exploration.Predicate, n)
			for i := range children {
				children[i] = exploration.Predicate{Field: &exploration.Field{Kind: "column", Key: "service"}, Op: "eq", Value: fmt.Sprintf("%d-%d", group, i)}
			}
			step.TraceSearch.Conditions = append(step.TraceSearch.Conditions, exploration.Predicate{All: children})
		}
		raw, _ := json.Marshal(step)
		expected := extra == 1
		if searchStepWireAdmission382(raw) != expected {
			t.Fatal("64-node admission boundary changed", extra)
		}
		cases = append(cases, searchStepWire382{Wire: raw, Accepted: expected})
	}
	// Boundary is encoded Go request bytes, not UTF-16 string length or raw
	// visible HTML characters. Structural predicate limits stay well below cap.
	budgetStep := func(n int) workbench.Step {
		step := searchStep382(1, 1, 7, 1)
		leaves := make([]exploration.Predicate, 12)
		for index := range leaves {
			leaves[index] = exploration.Predicate{Field: &exploration.Field{Kind: "attribute", Key: fmt.Sprintf("http.target.%d", index)}, Op: "contains", Value: strings.Repeat("<", n)}
		}
		step.TraceSearch.Conditions = []exploration.Predicate{{All: leaves}}
		return step
	}
	lo, hi := 0, 1024
	for lo < hi {
		mid := (lo + hi + 1) / 2
		step := budgetStep(mid)
		encoded, _ := json.Marshal(step.TraceSearch)
		if len(encoded) <= 65536 {
			lo = mid
		} else {
			hi = mid - 1
		}
	}
	for _, n := range []int{lo, lo + 1} {
		step := budgetStep(n)
		encoded, _ := json.Marshal(step.TraceSearch)
		accepted := len(encoded) <= 65536
		raw, _ := json.Marshal(step)
		if searchStepWireAdmission382(raw) != accepted {
			t.Fatal("encoded request byte boundary diverged", len(encoded))
		}
		cases = append(cases, searchStepWire382{Wire: raw, Accepted: accepted})
	}
	wire, _ := json.Marshal(cases)
	cmd := exec.Command("node", "--input-type=module", "-e", searchStepBatch382)
	cmd.Stdin = bytes.NewReader(wire)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("batched mandatory Go/JS saved trace-search parity: %v\n%s", err, out)
	}
	t.Log(string(out))
}
