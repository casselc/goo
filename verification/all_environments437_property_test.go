package verification

import (
	"bytes"
	"encoding/json"
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

func all437Selector() (identity.EnvironmentSelector, error) {
	var selector identity.EnvironmentSelector
	err := json.Unmarshal([]byte(`{"version":"goo/environment-selector/v1","policy":"effective","scope":{"mode":"all"}}`), &selector)
	return selector, err
}

// Reference classification reads label presence directly; it does not call the
// production observation/selection helpers whose conservation it checks.
func all437Observed(resource map[string]string) identity.EnvironmentScope {
	value, present := resource[identity.EnvironmentKey]
	if !present {
		value, present = resource[identity.LegacyEnvironmentKey]
	}
	if !present {
		return identity.EnvironmentScope{Mode: "unreported"}
	}
	if value == "" {
		return identity.EnvironmentScope{Mode: "empty"}
	}
	return identity.EnvironmentScope{Mode: "named", Value: value}
}

func TestHegelAllEnvironments437Conservation(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		wire, err := all437Selector()
		if err != nil {
			tc.Fatal(err)
		}
		all, err := wire.Selection()
		if err != nil {
			tc.Fatal(err)
		}
		canonical, e := all.Selector()
		if e != nil || canonical != wire {
			tc.Fatal("All lost its versioned wire", e)
		}
		if _, e = json.Marshal(all); e == nil {
			tc.Fatal("private semantic selection silently serialized")
		}
		names := []string{"", "prod", "all", "unreported", "雪", "prod' OR 1=1 --"}
		size := hegel.Draw(tc, hegel.Integers(1, 40))
		partitions := map[identity.EnvironmentScope]int{}
		for range size {
			resource := map[string]string{"service.name": "worker"}
			if hegel.Draw(tc, hegel.Booleans()) {
				resource[identity.EnvironmentKey] = hegel.Draw(tc, hegel.SampledFrom(names))
			}
			if hegel.Draw(tc, hegel.Booleans()) {
				resource[identity.LegacyEnvironmentKey] = hegel.Draw(tc, hegel.SampledFrom(names))
			}
			expected := all437Observed(resource)
			if actual := identity.ObserveEnvironment(resource); actual.Scope != expected || actual.Scope.Mode == "all" {
				tc.Fatal("All became an observed environment", actual, expected)
			}
			partitions[expected]++
			if !all.Matches(resource) || (identity.EnvironmentSelection{}).Matches(resource) || (identity.EnvironmentScope{Mode: "all"}).Matches(resource) {
				tc.Fatal("selection wildcard leaked into absent/observed scope")
			}
			matches := 0
			for _, scope := range []identity.EnvironmentScope{{Mode: "empty"}, {Mode: "unreported"}, {Mode: "named", Value: "prod"}, {Mode: "named", Value: "all"}, {Mode: "named", Value: "unreported"}, {Mode: "named", Value: "雪"}, {Mode: "named", Value: "prod' OR 1=1 --"}} {
				selected, e := identity.EffectiveEnvironment(scope)
				if e != nil {
					tc.Fatal(e)
				}
				want := scope == expected
				if selected.Matches(resource) != want || selected.Equal(all) {
					tc.Fatal("partition leaked", scope, expected)
				}
				if selected.Matches(resource) {
					matches++
				}
			}
			if matches != 1 {
				tc.Fatal("effective partitions do not cover All exactly once", resource, matches)
			}
		}
		total := 0
		for _, count := range partitions {
			total += count
		}
		if total != size {
			tc.Fatal("partition total lost observations")
		}
		from := int64(hegel.Draw(tc, hegel.Integers(100000, 900000)))
		body := "bank-'雪-" + strconv.Itoa(size)
		q := exploration.Query{Signal: "logs", EnvironmentSelector: &wire, FromMS: from, ToMS: from + 60000, Limit: 100, Where: &exploration.Predicate{Field: &exploration.Field{Kind: "column", Key: "body"}, Op: "eq", Value: body}}
		if err := q.Validate(); err != nil {
			tc.Fatal(err)
		}
		plan, err := exploration.Compile(q)
		if err != nil {
			tc.Fatal(err)
		}
		if _, present := plan.Parameters["environment"]; present || plan.Parameters["from"] != strconv.FormatInt(from, 10) || plan.Parameters["to"] != strconv.FormatInt(from+60000, 10) {
			tc.Fatal("All changed bounds or acquired a named environment", plan.Parameters)
		}
		for _, sql := range []string{plan.Count, plan.Rows, plan.Fields} {
			if !strings.Contains(sql, "{from:Int64}") || !strings.Contains(sql, "{to:Int64}") || strings.Contains(sql, body) || strings.Contains(sql, "{environment:String}") || !strings.Contains(sql, "Body = {") {
				tc.Fatal("All erased the bounded predicate or interpolated data", sql)
			}
		}
		// Tail/fixed/capture/step carriers must keep the explicit wire, not erase it
		// into the historical empty environment string or a named literal "all".
		tail := exploration.TailSelector{Signal: "logs", EnvironmentSelector: &wire, Where: q.Where}.Query(q.ToMS)
		fixed := (exploration.Query{Signal: "logs", FromMS: q.FromMS, ToMS: q.ToMS, Limit: 10, Where: q.Where}).WithEnvironmentFrom(tail)
		captured := (exploration.CaptureQuestion{Kind: "explore", Explore: &fixed}).Normalized()
		for _, actual := range []exploration.Query{tail, fixed, *captured.Explore} {
			if actual.Validate() != nil || !actual.SameEnvironment(q) || actual.EnvironmentSelector == nil || actual.EnvironmentSelector.Scope.Mode != "all" {
				tc.Fatal("carrier narrowed/broadened All", actual)
			}
		}
		step := workbench.Step{Version: 1, Kind: "record", View: "Records", Scope: workbench.StepScope{EnvironmentSelector: fixed.EnvironmentSelector, FromMS: fixed.FromMS, ToMS: fixed.ToMS}, Record: &fixed}
		encoded, err := json.Marshal(step)
		if err != nil {
			tc.Fatal(err)
		}
		var restored workbench.Step
		if err = json.Unmarshal(encoded, &restored); err != nil || restored.Validate() != nil || !restored.Record.SameEnvironment(q) {
			tc.Fatal("step lost explicit All", err)
		}
		wire.Scope.Mode = "broken"
		if fixed.EnvironmentSelector.Scope.Mode != "all" || tail.EnvironmentSelector.Scope.Mode != "all" || captured.Explore.EnvironmentSelector.Scope.Mode != "all" {
			tc.Fatal("carriers retained caller alias")
		}
	}, options(150)...)
}

func TestAllEnvironments437GoJavaScriptParity(t *testing.T) {
	type parity struct {
		Selector identity.EnvironmentSelector `json:"selector"`
		Resource map[string]string            `json:"resource"`
		Matches  bool                         `json:"matches"`
		Observed identity.EnvironmentScope    `json:"observed"`
	}
	all, err := all437Selector()
	if err != nil {
		t.Fatal(err)
	}
	selectors := []identity.EnvironmentSelector{all}
	for _, scope := range []identity.EnvironmentScope{{Mode: "empty"}, {Mode: "unreported"}, {Mode: "named", Value: "all"}, {Mode: "named", Value: "prod"}} {
		selected, e := identity.EffectiveEnvironment(scope)
		if e != nil {
			t.Fatal(e)
		}
		wire, e := selected.Selector()
		if e != nil {
			t.Fatal(e)
		}
		selectors = append(selectors, wire)
	}
	current, _ := identity.CurrentEnvironment("all")
	wire, _ := current.Selector()
	selectors = append(selectors, wire)
	values := []string{"", "all", "prod", "unreported", "雪"}
	cases := []parity{}
	for _, cv := range values {
		for _, lv := range values {
			for bits := range 4 {
				resource := map[string]string{}
				if bits&1 != 0 {
					resource[identity.EnvironmentKey] = cv
				}
				if bits&2 != 0 {
					resource[identity.LegacyEnvironmentKey] = lv
				}
				for _, selector := range selectors {
					selection, e := selector.Selection()
					if e != nil {
						t.Fatal(e)
					}
					cases = append(cases, parity{selector, resource, selection.Matches(resource), all437Observed(resource)})
				}
			}
		}
	}
	input, _ := json.Marshal(cases)
	script := `import fs from 'node:fs';import * as E from '../ui/record-environment.mjs';const cases=JSON.parse(fs.readFileSync(0,'utf8'));let n=0;for(const c of cases){const normalized=E.normalize(c.selector);if(E.matches(normalized,c.resource)!==c.matches)throw Error('membership mismatch '+n);if(JSON.stringify(E.observe(c.resource).scope)!==JSON.stringify(c.observed))throw Error('observation mismatch '+n);if(c.selector.scope.mode==='all'){const q={signal:'logs',environmentSelector:c.selector,fromMs:100000,toMs:160000,limit:100,offset:0,where:{field:{kind:'column',key:'body'},op:'eq',value:'bank'}};const narrower=E.observedQuestion(q,c.resource);if(narrower.fromMs!==q.fromMs||narrower.toMs!==q.toMs||JSON.stringify(narrower.where)!==JSON.stringify(q.where)||!E.equal(narrower.environmentSelector,E.effective(c.observed))||!E.isAll(q))throw Error('observed narrowing changed parent '+n)}if(c.selector.scope.mode==='all'&&E.label(normalized)!=='All environments')throw Error('All label lost');if(c.selector.scope.mode==='named'&&c.selector.scope.value==='all'&&E.label(normalized)==='All environments')throw Error('literal all conflated');n++}process.stdout.write(JSON.stringify({cases:n}));`
	cmd := exec.CommandContext(t.Context(), "node", "--input-type=module", "-e", script)
	cmd.Stdin = bytes.NewReader(input)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("actual JS parity failed: %v %s", err, out)
	}
	var result struct{ Cases int }
	if err = json.Unmarshal(out, &result); err != nil || result.Cases != len(cases) {
		t.Fatal("parity omitted cases", err, string(out))
	}
	// Zero/blank/current-only All and malformed typed choices must never become broad.
	for _, raw := range []string{`{"version":"goo/environment-selector/v1","policy":"current-only","scope":{"mode":"all"}}`, `{"version":"goo/environment-selector/v1","policy":"effective","scope":{"mode":"all","value":"prod"}}`, `{"version":"goo/environment-selector/v1","policy":"effective","scope":{"mode":"all","value":null}}`} {
		var rejected identity.EnvironmentSelector
		if json.Unmarshal([]byte(raw), &rejected) == nil {
			t.Fatal("invalid All admitted", raw)
		}
	}
	if reflect.DeepEqual(selectors[0], selectors[3]) {
		t.Fatal("named all conflated")
	}
	t.Logf("Go/actual-JS parity passed %d resource/policy cases", len(cases))
}
