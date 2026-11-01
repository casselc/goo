package verification

import (
	"encoding/json"
	"fmt"
	"github.com/casselc/goo/workbench"
	"hegel.dev/go/hegel"
	"testing"
)

func TestHegelPortableViewArtifactRoundTripAndIsolation(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		seed := hegel.Draw(tc, hegel.Integers(1, 100000))
		signal := "logs"
		if seed%2 == 0 {
			signal = "spans"
		}
		q := map[string]any{"signal": signal, "fromMs": int64(seed) * 1000, "toMs": int64(seed)*1000 + 60000, "limit": 100, "offset": 0, "where": map[string]any{"field": map[string]any{"kind": "attribute", "key": "tenant"}, "op": "eq", "value": fmt.Sprintf("null: #tenant-%d [9007199254740993] \"quoted\"", seed)}}
		if seed%3 == 0 {
			q["environment"] = "prod"
		} else {
			policy := "effective"
			if seed%2 == 0 {
				policy = "current-only"
				q["environment"] = ""
			}
			q["environmentSelector"] = map[string]any{"version": "goo/environment-selector/v1", "policy": policy, "scope": map[string]any{"mode": "named", "value": "prod"}}
		}
		recordKind := "record-list"
		if signal == "spans" && seed%3 == 0 {
			recordKind = "record-span-waterfall"
		}
		scope := map[string]any{"kind": "record-query", "signal": signal}
		records := map[string]any{"records": map[string]any{"version": "goo.view/1", "scope": scope, "kind": map[string]any{"name": recordKind, "version": 1}, "display": map[string]any{"columns": []string{"utc", "summary", "service"}, "fieldColumns": []any{map[string]any{"kind": "attribute", "key": "zone"}, map[string]any{"kind": "resource", "key": "zone"}}}}}
		children := []any{records}
		if seed%5 != 0 {
			q["groupBy"] = []any{map[string]any{"kind": "attribute", "key": "cohort"}}
			kind := "bar"
			if seed%2 == 0 {
				kind = "table"
			}
			children = []any{map[string]any{"aggregate": map[string]any{"version": "goo.view/1", "scope": scope, "kind": map[string]any{"name": kind, "version": 1}, "display": map[string]any{"category": "category", "value": "count"}}}, records}
		}
		viewSpec := map[string]any{"version": 1, "title": fmt.Sprintf("Team \"cohort\": #%d →", seed), "queries": []any{q}, "composition": map[string]any{"version": "goo.composition/1", "kind": map[string]any{"name": "stack", "version": 1}, "children": children}}
		if seed%7 == 0 {
			delete(viewSpec, "composition")
			viewSpec["layout"] = "records"
			viewSpec["columns"] = []string{"utc", "summary"}
			viewSpec["fieldColumns"] = []any{}
			viewSpec["displays"] = map[string]any{}
			viewSpec["breakdowns"] = map[string]any{}
			q["limit"] = 100
			delete(q, "offset")
		}
		raw, _ := json.Marshal(viewSpec)
		var view workbench.ExplorationView
		if e := json.Unmarshal(raw, &view); e != nil {
			tc.Fatal(e)
			return
		}
		artifact := workbench.ViewArtifact{Version: workbench.ViewArtifactVersion, Namespace: "null", Name: fmt.Sprintf("service-%d", seed), Revision: fmt.Sprint(seed), View: view}
		before, _ := json.Marshal(artifact)
		encoded, e := workbench.EncodeViewArtifact(artifact)
		if e != nil {
			tc.Fatal(e)
			return
		}
		decoded, e := workbench.DecodeViewArtifact(encoded.YAML)
		if e != nil {
			tc.Fatal(e)
			return
		}
		after, _ := json.Marshal(decoded.Artifact)
		if string(before) != string(after) || encoded.YAML != decoded.YAML {
			tc.Fatal("portable AST/canonical YAML changed", encoded.YAML)
			return
		}
		again, e := workbench.EncodeViewArtifact(decoded.Artifact)
		if e != nil || again.YAML != encoded.YAML {
			tc.Fatal("canonical encoding unstable", e)
			return
		}
		// Both encode and decode return detached data, never aliases into author state.
		encoded.Artifact.View.Queries[0].Where.Value = "edited encode"
		decoded.Artifact.View.Queries[0].Where.Value = "edited decode"
		after, _ = json.Marshal(artifact)
		if string(before) != string(after) {
			tc.Fatal("codec mutated author state")
			return
		}
		for _, bad := range []string{encoded.YAML + "\n---\n{}\n", encoded.YAML + "\nunknown: true\n", "version: goo.view-artifact/1\nversion: goo.view-artifact/1\n"} {
			if _, e = workbench.DecodeViewArtifact(bad); e == nil {
				tc.Fatal("malformed artifact accepted")
				return
			}
		}
	}, options(50)...)
}
