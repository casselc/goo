package verification

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/identity"
	"hegel.dev/go/hegel"
)

func tsSelector340(policy, mode, name string) *identity.EnvironmentSelector {
	return &identity.EnvironmentSelector{Version: identity.EnvironmentSelectorVersion, Policy: policy, Scope: identity.EnvironmentScope{Mode: mode, Value: name}}
}

func tsTypedRecords340(request exploration.TraceSearchRequest, name string) []tsRecord {
	// Current-key presence wins even when empty. These are independent physical
	// resource shapes, not labels normalized by a production helper.
	type labels struct{ current, legacy *string }
	text := func(v string) *string { return &v }
	variants := []labels{
		{}, {current: text(name)}, {legacy: text(name)},
		{current: text(name), legacy: text("other")},
		{current: text("other"), legacy: text(name)},
		{current: text(""), legacy: text(name)}, {legacy: text("")},
		{current: text(name), legacy: text("")}, {current: text(""), legacy: text("")},
		{current: text("other")}, {legacy: text("other")},
	}
	var records []tsRecord
	for index, labels := range variants {
		id := fmt.Sprintf("%032x", index+100)
		for part, flags := range []int{1, 2} {
			record := tsGeneratedRecord(request, id, index*2+part, flags)
			record.Environment = "deliberately-wrong-fixture-label"
			delete(record.Values, exploration.Field{Kind: "resource", Key: "deployment.environment.name"})
			delete(record.Values, exploration.Field{Kind: "resource", Key: "deployment.environment"})
			if labels.current != nil {
				record.Values[exploration.Field{Kind: "resource", Key: "deployment.environment.name"}] = *labels.current
			}
			if labels.legacy != nil {
				record.Values[exploration.Field{Kind: "resource", Key: "deployment.environment"}] = *labels.legacy
			}
			records = append(records, record)
			if part == 0 {
				records = append(records, record)
			}
		}
	}
	return records
}

func TestHegelTypedTraceSearchEnvironment340(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		seed := hegel.Draw(tc, hegel.Integers(1, 100000))
		limit := hegel.Draw(tc, hegel.Integers(1, 4))
		name := fmt.Sprintf("team-雪-%d');--", seed)
		scalar := tsGeneratedRequest(seed, 2, limit)
		scalar.Environment = name
		requests := []exploration.TraceSearchRequest{scalar}
		for _, selector := range []*identity.EnvironmentSelector{
			tsSelector340("current-only", "named", name),
			tsSelector340("effective", "named", name),
			tsSelector340("effective", "empty", ""),
			tsSelector340("effective", "unreported", ""),
		} {
			request := scalar
			request.Environment = ""
			request.EnvironmentSelector = selector
			requests = append(requests, request)
		}
		// The exact resources are kept constant across policies. Policy must change
		// membership, not the selected fixture label or a display alias.
		expectedTotals := []string{"3", "3", "4", "3", "1"}
		parents := make([]exploration.Query, len(requests))
		for i, request := range requests {
			records := tsTypedRecords340(request, name)
			if !tsCheckReference(request, records) {
				tc.Errorf("independent typed reference mismatch profile%d", i)
				return
			}
			_, total := tsOracle(request, records)
			if total != expectedTotals[i] {
				tc.Errorf("presence/precedence profile%d total%s expected%s", i, total, expectedTotals[i])
				return
			}
			plan, err := exploration.CompileTraceSearch(request)
			if err != nil || strings.Contains(plan.SQL, name) {
				tc.Errorf("typed compiler interpolation/admission profile%d: %v", i, err)
				return
			}
			hasName := false
			for _, value := range plan.Parameters {
				hasName = hasName || value == name
			}
			if hasName != (i < 3) {
				tc.Errorf("wrong named binding profile%d", i)
				return
			}
			reversed := slices.Clone(records)
			slices.Reverse(reversed)
			if !tsCheckReference(request, reversed) {
				tc.Errorf("record order changed typed witnesses profile%d", i)
				return
			}
			before, totalBefore, _ := exploration.ReferenceTraceSearch(request, tsLogical(records))
			all := request
			all.Limit = 100
			full, fullTotal, _ := exploration.ReferenceTraceSearch(all, tsLogical(records))
			if totalBefore != fullTotal || !aeJSONEqual(before, full[:min(request.Limit, len(full))]) || len(full) != mustInt340(totalBefore) {
				tc.Errorf("limit changed full population profile%d", i)
				return
			}
			parents[i] = exploration.Query{Signal: "spans", Environment: request.Environment, EnvironmentSelector: request.EnvironmentSelector, FromMS: request.FromMS, ToMS: request.ToMS, Limit: 100, Offset: 100}
		}
		input, _ := json.Marshal(map[string]any{"parents": parents, "requests": requests})
		command := exec.Command("node", "-e", `const fs=require('fs'),m=require('../ui/trace-search.mjs'),x=JSON.parse(fs.readFileSync(0,'utf8')),before=JSON.stringify(x),requests=x.parents.map((p,i)=>m.buildRequest(p,x.requests[i].conditions,x.requests[i].limit));process.stdout.write(JSON.stringify({requests,unchanged:before===JSON.stringify(x)}))`)
		command.Stdin = bytes.NewReader(input)
		raw, err := command.CombinedOutput()
		if err != nil {
			tc.Errorf("actual typed UI builder: %v %s", err, raw)
			return
		}
		var output struct {
			Requests  []exploration.TraceSearchRequest
			Unchanged bool
		}
		if json.Unmarshal(raw, &output) != nil || !output.Unchanged || !aeJSONEqual(output.Requests, requests) {
			tc.Errorf("typed UI request lost exact policy/representation: %s", raw)
		}
	}, options(40)...)
}

func mustInt340(value string) int { n, _ := strconv.Atoi(value); return n }

func TestTypedTraceSearchPolicyContrasts340(t *testing.T) {
	request := tsGeneratedRequest(1, 2, 100)
	request.Environment = ""
	request.EnvironmentSelector = tsSelector340("effective", "named", "prod")
	records := tsTypedRecords340(request, "prod")
	// A resource condition can observe a shadowed legacy value without changing
	// which key selects the record population.
	request.Conditions = []exploration.Predicate{tsLeaf("resource", "deployment.environment", "eq", "other")}
	if !tsCheckReference(request, records) {
		t.Fatal("legacy field predicate and winning environment were conflated")
	}
	_, total := tsOracle(request, records)
	if total != "1" {
		t.Fatalf("wanted one current-prod/legacy-other trace, got%s", total)
	}
	request.Conditions = []exploration.Predicate{tsLeaf("resource", "deployment.environment.name", "missing", "")}
	_, total = tsOracle(request, records)
	if total != "1" || !tsCheckReference(request, records) {
		t.Fatal("legacy-only effective membership lost")
	}
	request.EnvironmentSelector = tsSelector340("effective", "unreported", "")
	if !tsCheckReference(request, records) {
		t.Fatal("unreported environment witness absence lost")
	}
	bad := request
	bad.Environment = "prod"
	if bad.Validate() == nil {
		t.Fatal("ambiguous scalar plus typed selector accepted")
	}
	bad = request
	bad.EnvironmentSelector = tsSelector340("current-only", "empty", "")
	if bad.Validate() == nil {
		t.Fatal("empty current-only selector accepted")
	}
}
