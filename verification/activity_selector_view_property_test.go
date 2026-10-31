package verification

import (
	"encoding/json"
	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/identity"
	"github.com/casselc/goo/workbench"
	"hegel.dev/go/hegel"
	"strings"
	"testing"
)

const activitySelectorViewJS = `import assert from 'node:assert/strict';import fs from 'node:fs';import {model} from '../ui/activity-views.mjs';const {view,accepted}=JSON.parse(fs.readFileSync(0,'utf8'));let out;try{out=model({...view,id:'a'.repeat(64)})}catch{}assert.equal(!!out,accepted);if(out){assert.deepEqual(out.selector,view.selector);out.selector.environmentSelector.scope.mode='changed';assert.notEqual(view.selector.environmentSelector.scope.mode,'changed')}`

func TestHegelActivitySelectorViewParity(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		seed := hegel.Draw(tc, hegel.Integers(0, 1000))
		variant := hegel.Draw(tc, hegel.Integers(0, 5))
		scope := identity.EnvironmentScope{Mode: []string{"named", "empty", "unreported"}[seed%3]}
		if scope.Mode == "named" {
			scope.Value = "scope-" + strings.Repeat("x", seed%12)
		}
		selector := &identity.EnvironmentSelector{Version: identity.EnvironmentSelectorVersion, Policy: "effective", Scope: scope}
		v := workbench.ActivityView{Version: 1, Title: "Portable Activity", Selector: exploration.TailSelector{Signal: []string{"logs", "spans"}[seed%2], EnvironmentSelector: selector}, Columns: []string{"utc", "service", "summary"}}
		switch variant {
		case 1:
			selector.Policy = "invalid"
		case 2:
			selector.Policy = "current-only"
			selector.Scope.Mode = "unreported"
			selector.Scope.Value = ""
		case 3:
			selector.Version = "future"
		case 4:
			v.Selector.Environment = "ambiguous"
		case 5:
			selector.Scope.Mode = "named"
			selector.Scope.Value = ""
		}
		accepted := v.Validate() == nil
		if accepted != (variant == 0) {
			tc.Fatal("unexpected validation", variant)
		}
		before, _ := json.Marshal(v)
		if e := runNativeHistogramJS(map[string]any{"view": v, "accepted": accepted}, activitySelectorViewJS); e != nil {
			tc.Fatal(e)
		}
		after, _ := json.Marshal(v)
		if string(before) != string(after) {
			tc.Fatal("validation mutated view")
		}
	}, options(100)...)
}
