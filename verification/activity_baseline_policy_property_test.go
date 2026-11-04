package verification

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"testing"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/identity"
	"hegel.dev/go/hegel"
)

func TestHegelActivitySavedBaselinePolicy322(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		seed := hegel.Draw(tc, hegel.Integers(1, 100000))
		mode := hegel.Draw(tc, hegel.SampledFrom([]string{"current", "named", "empty", "unreported"}))
		scope := identity.EnvironmentScope{Mode: mode}
		if mode == "current" || mode == "named" {
			scope.Mode = "named"
			scope.Value = fmt.Sprintf("prod-%d'雪", seed%9)
		}
		selection, e := identity.EffectiveEnvironment(scope)
		if mode == "current" {
			selection, e = identity.CurrentEnvironment(scope.Value)
		}
		if e != nil {
			tc.Fatal(e)
			return
		}
		selector, _ := selection.Selector()
		ns := identity.Namespace{Mode: "missing"}
		if seed%3 != 0 {
			ns.Mode = "value"
			if seed%3 == 2 {
				ns.Value = "team'雪"
			}
		}
		baseline := exploration.Query{Signal: "spans", EnvironmentSelector: &selector, FromMS: 100000 + int64(seed)*1000, ToMS: 160000 + int64(seed)*1000, Limit: 100, GroupBy: []exploration.Field{{Kind: "column", Key: "duration_bucket"}}, Where: exploration.NamespacePredicate(&ns)}
		selected := baseline
		if mode == "current" {
			selected.Environment = scope.Value
			selected.EnvironmentSelector = nil
		}
		selected.FromMS += 60000
		selected.ToMS += 60000
		selected.Where = &exploration.Predicate{All: []exploration.Predicate{*baseline.Where, {Field: &exploration.Field{Kind: "attribute", Key: "tenant"}, Op: "eq", Value: strconv.Itoa(seed)}}}
		reads := 0
		fetch := func(q exploration.Query, n int) (*exploration.Result, error) {
			source := exploration.Source{Endpoint: "http://source.invalid", Client: &http.Client{Transport: snapshotTransport(func(req *http.Request) (*http.Response, error) {
				reads++
				body, _ := io.ReadAll(req.Body)
				sql := string(body)
				var text string
				switch {
				case strings.Contains(sql, "GROUP BY"):
					text = fmt.Sprintf("{\"g0\":\"[0,1) ms\",\"count\":\"%d\"}\n", n)
				case strings.Contains(sql, "SELECT DISTINCT"):
				case strings.Contains(sql, " AS timestamp_ns"):
					for i := 0; i < n; i++ {
						resource := map[string]string{}
						if scope.Mode == "named" {
							resource[identity.EnvironmentKey] = scope.Value
						}
						if scope.Mode == "empty" {
							resource[identity.EnvironmentKey] = ""
						}
						if ns.Mode == "value" {
							resource[identity.NamespaceKey] = ns.Value
						}
						row, _ := json.Marshal(map[string]any{"timestamp_ns": strconv.FormatInt(q.FromMS*1000000+int64(i), 10), "duration_ns": "1", "duration_bucket": "[0,1) ms", "SpanName": "work", "ResourceAttributes": resource})
						text += string(row) + "\n"
					}
				default:
					text = fmt.Sprintf("{\"count\":\"%d\"}\n", n)
				}
				return &http.Response{StatusCode: 200, Header: http.Header{}, Body: io.NopCloser(strings.NewReader(text))}, nil
			})}}
			return source.Fetch(context.Background(), q)
		}
		a, e := fetch(baseline, seed%3+1)
		if e != nil {
			tc.Fatal(e)
			return
		}
		b, e := fetch(selected, seed%4+1)
		if e != nil || reads != 8 || a.HistogramDataset == nil || b.HistogramDataset == nil {
			tc.Fatal("real source DTO failed", e, reads)
			return
		}
		id := strings.Repeat("a", 64)
		page := map[string]any{"receipt": map[string]any{"id": id, "title": "Saved baseline", "capturedAt": "2026-11-04T00:00:00Z", "question": map[string]any{"kind": "explore", "explore": baseline}}, "explore": a}
		input := map[string]any{"page": page, "current": b, "id": id, "mode": mode}
		script := `
 import assert from 'node:assert/strict';import fs from 'node:fs';import * as B from '../ui/activity-baseline.mjs';import * as A from '../ui/activity-explanation.mjs';import * as P from '../ui/question-protocol.mjs';import * as S from '../ui/question-session.mjs';
 const {page,current,id,mode}=JSON.parse(fs.readFileSync(0,'utf8')),before=JSON.stringify({page,current});const baseline=B.capture(page,id),pair=B.request(baseline,current);assert.equal(pair.baseline.count,page.explore.count);assert.equal(pair.baseline.basis,page.explore.basis);assert.equal(pair.selection.count,current.count);assert.equal(pair.selection.basis,current.basis);const q=A.savedScope(page,id,current.question);assert.equal(S.identity(q),S.identity(page.receipt.question.explore));
 const alternate=structuredClone(page);alternate.explore.question.limit=1;alternate.explore.question.offset=100;assert.equal(S.identity(A.savedScope(alternate,id,current.question)),S.identity(q));
 for(const mutate of [p=>p.receipt.id='b'.repeat(64),p=>p.explore.question.fromMs++,p=>p.explore.question.environmentSelector.scope={mode:'named',value:'foreign'}]){const bad=structuredClone(page);mutate(bad);assert.throws(()=>B.capture(bad,id));assert.throws(()=>A.savedScope(bad,id,current.question));}
 if(mode==='current'){const crossed=structuredClone(page);crossed.explore.question.environment=crossed.explore.question.environmentSelector.scope.value;delete crossed.explore.question.environmentSelector;assert.throws(()=>B.capture(crossed,id));assert.throws(()=>A.savedScope(crossed,id,current.question));}const invalid=structuredClone(page);invalid.receipt.question.explore.where=invalid.explore.question.where={field:{kind:'column',key:'not_a_column'},op:'eq',value:'x'};assert.throws(()=>A.savedScope(invalid,id,current.question));assert.throws(()=>B.capture(invalid,id));const badCurrent=structuredClone(current);delete badCurrent.question.environment;badCurrent.question.environmentSelector={version:'goo/environment-selector/v1',policy:'effective',scope:{mode:'named',value:'foreign'}};assert.throws(()=>B.request(baseline,badCurrent));assert.throws(()=>A.savedScope(page,id,badCurrent.question));const corrupt=structuredClone(page);corrupt.explore.count=String(BigInt(corrupt.explore.count)+1n);assert.throws(()=>B.capture(corrupt,id));assert.equal(S.identity(A.savedScope(corrupt,id,current.question)),S.identity(q),'scope-only reuse must not acquire stored counts');
 const returned=structuredClone(q);returned.where={all:[returned.where,{field:{kind:'attribute',key:'tenant'},op:'eq',value:'chosen'}]};const link=P.questionLink(returned);assert.equal(S.identity(P.readStepLink(link).record),S.identity(returned));q.where={field:{kind:'attribute',key:'changed'},op:'exists'};baseline.data.groups[0].count='999';assert.equal(JSON.stringify({page,current}),before);
 `
		if e = runNativeHistogramJS(input, script); e != nil {
			tc.Fatal(e)
			return
		}
		request := exploration.ExplanationRequest{Version: exploration.ExplanationVersion, Baseline: baseline, Selection: selected, Fields: []exploration.Field{{Kind: "attribute", Key: "tenant"}}}
		if _, e = exploration.PlanExplanation(request); e != nil {
			tc.Fatal(e)
			return
		}
		value := "chosen"
		for _, side := range []string{"baseline", "selection"} {
			q, err := exploration.PlanExplanationFilter(request, request.Fields[0], &value, side, "include")
			expected := baseline
			if side == "selection" {
				expected = selected
			}
			if err != nil || !q.SameEnvironment(expected) || q.FromMS != expected.FromMS || q.ToMS != expected.ToMS {
				tc.Fatal("side filter lost scope", err)
				return
			}
		}
	}, options(30)...)
}
