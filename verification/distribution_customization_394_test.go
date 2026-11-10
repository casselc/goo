package verification

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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

type customizeVector394 struct {
	Data         *exploration.Result `json:"data"`
	Presentation map[string]any      `json:"presentation"`
	Destination  exploration.Query   `json:"destination"`
}

// Synthetic source responses provide ordinary admitted snapshots, not evidence
// of actual ClickHouse membership. The observed journey separately supplies it.
// The generated outputs go through the real immutable public view store below.
func TestHegelDistributionCustomization394(t *testing.T) {
	vectors := []customizeVector394{}
	hegel.Test(t, func(tc *hegel.T) {
		seed := hegel.Draw(tc, hegel.Integers(1, 100000))
		for _, signal := range []string{"logs", "spans"} {
			a, b := seed%5+1, seed%4+2
			total := a + b
			q := exploration.Query{Signal: signal, FromMS: 1000 + int64(seed), ToMS: 61000 + int64(seed), Limit: seed%5 + 1, GroupBy: []exploration.Field{{Kind: "attribute", Key: "region"}}, EnvironmentSelector: &identity.EnvironmentSelector{Version: identity.EnvironmentSelectorVersion, Policy: []string{"current-only", "effective"}[seed%2], Scope: identity.EnvironmentScope{Mode: "named", Value: fmt.Sprintf("environment-%c", 'a'+seed%7)}}, Where: &exploration.Predicate{All: []exploration.Predicate{
				{Field: &exploration.Field{Kind: "resource", Key: "service.namespace"}, Op: "eq", Value: fmt.Sprintf("team-%c", 'a'+seed%5)},
				{Any: []exploration.Predicate{{Field: &exploration.Field{Kind: "attribute", Key: "tenant"}, Op: "eq", Value: "a'雪"}, {Field: &exploration.Field{Kind: "attribute", Key: "tenant"}, Op: "eq", Value: "b\\literal"}}},
				{Field: &exploration.Field{Kind: "attribute", Key: "optional"}, Op: "missing"},
			}}}
			plan, e := exploration.Compile(q)
			if e != nil {
				tc.Fatal(e)
			}
			groups := fmt.Sprintf("{\"g0\":\"north\",\"count\":\"%d\"}\n{\"g0\":\"south\",\"count\":\"%d\"}", a, b)
			rows := []string{}
			for i := 0; i < min(q.Limit+1, total); i++ {
				region := "north"
				if i >= a {
					region = "south"
				}
				wire, _ := json.Marshal(map[string]any{"timestamp_ns": strconv.FormatInt((q.FromMS+1+int64(i))*1000000, 10), "duration_ns": "0", "ServiceName": "fixture-worker", "SpanId": fmt.Sprintf("%016x", i+1), "TraceId": fmt.Sprintf("%032x", i+1), "SpanName": "fixture operation", "SpanKind": "Internal", "Body": "fixture record", "SeverityText": "INFO", "ResourceAttributes": map[string]string{"service.name": "fixture-worker", "deployment.environment.name": q.EnvironmentSelector.Scope.Value, "service.namespace": q.Where.All[0].Value}, "Attributes": map[string]string{"tenant": "a'雪", "region": region}})
				rows = append(rows, string(wire))
			}
			source := exploration.Source{Endpoint: "http://source.invalid", Client: &http.Client{Transport: snapshotTransport(func(req *http.Request) (*http.Response, error) {
				wire, _ := io.ReadAll(req.Body)
				body := ""
				switch string(wire) {
				case plan.Count:
					body = fmt.Sprintf(`{"count":"%d"}`, total)
				case plan.Groups:
					body = groups
				case plan.Rows:
					body = strings.Join(rows, "\n")
				case plan.Fields:
				default:
					return nil, fmt.Errorf("unexpected snapshot fixture read")
				}
				return &http.Response{StatusCode: 200, Header: http.Header{}, Body: io.NopCloser(strings.NewReader(body))}, nil
			})}}
			data, e := source.Fetch(t.Context(), q)
			if e != nil || data.ViewDataset == nil || data.ViewDataset.Completeness == nil || data.State != "observed" {
				tc.Fatal("ordinary complete snapshot fixture refused", e)
			}
			raw, _ := json.Marshal(q)
			var destination exploration.Query
			json.Unmarshal(raw, &destination)
			destination.FromMS += 120000
			destination.ToMS += 120000
			destination.Limit = (q.Limit % 5) + 1
			destination.EnvironmentSelector.Scope.Value = "destination"
			destination.Where = &exploration.Predicate{Field: &exploration.Field{Kind: "column", Key: "service"}, Op: "eq", Value: "destination-worker"}
			presentation := map[string]any{"kind": []string{"bar", "table"}[seed%2], "category": "category", "value": "count", "scale": []string{"max", "population"}[seed%2], "order": []string{"source", "label", "value-desc", "value-asc"}[seed%4], "query": []string{"", "north", "south"}[seed%3], "requiresComplete": seed%2 == 0}
			vectors = append(vectors, customizeVector394{data, presentation, destination})
		}
	}, options(16)...)
	input, _ := json.Marshal(vectors)
	cmd := exec.CommandContext(t.Context(), "node", "--input-type=module", "-e", customizeJS394)
	cmd.Stdin = bytes.NewReader(input)
	var errors bytes.Buffer
	cmd.Stderr = &errors
	wire, e := cmd.Output()
	if e != nil {
		t.Fatalf("shared customization: %v %s", e, errors.String())
	}
	var staged []workbench.ExplorationView
	if e = json.Unmarshal(wire, &staged); e != nil || len(staged) != len(vectors) {
		t.Fatal("invalid staged configuration batch", e)
	}
	store := workbench.Store{Dir: t.TempDir()}
	for index, v := range staged {
		q := vectors[index].Data.Question
		if !reflect.DeepEqual(v.Queries, []exploration.Query{q}) || v.Validate() != nil {
			t.Fatal("staged configuration changed full typed question")
		}
		saved, e := store.CreateExplorationView(v)
		if e != nil {
			t.Fatal(e)
		}
		held, e := store.LoadExplorationView(saved.ID)
		if e != nil || !reflect.DeepEqual(saved, held) {
			t.Fatal("immutable configured view changed", e)
		}
		d, ok := v.Displays[q.Signal]
		if !ok {
			t.Fatal("count declaration absent")
		}
		before, _ := json.Marshal(vectors[index].Destination)
		reused, e := workbench.ReuseViewDeclaration(d, q, vectors[index].Destination)
		if e != nil || !reflect.DeepEqual(reused, d) {
			t.Fatal("compatible display reuse refused", e)
		}
		after, _ := json.Marshal(vectors[index].Destination)
		if string(before) != string(after) {
			t.Fatal("reuse mutated destination question")
		}
	}
	t.Logf("%d shared ordinary logs/spans handoffs and immutable store round trips", len(vectors))
}

const customizeJS394 = `import fs from 'node:fs';import assert from 'node:assert/strict';import {distributionCustomization} from '../ui/distributions.mjs';import {adoptExploration} from '../ui/source-snapshots.mjs';import * as D from '../ui/view-declaration.mjs';
globalThis.fetch=()=>{throw Error('Customize attempted a telemetry read')};const out=[];for(const v of JSON.parse(fs.readFileSync(0,'utf8'))){const before=JSON.stringify(v);const data=adoptExploration(v.data,v.data.question);const staged=distributionCustomization(data,v.presentation);assert.deepEqual(staged.queries,[v.data.question]);assert.equal(staged.displays[v.data.question.signal].display.value,'count');assert.deepEqual(D.reuse(staged.displays[v.data.question.signal],v.data.question,v.destination),staged.displays[v.data.question.signal]);assert.equal(Object.hasOwn(staged,'basis'),false);assert.equal(Object.hasOwn(staged,'data'),false);assert.equal(Object.hasOwn(staged,'viewDataset'),false);for(const change of [x=>x.question.offset=1,x=>x.state='changed_between_reads',x=>x.state='no_data',x=>x.state='unavailable',x=>x.question.groupTimeSeries=true,x=>x.question.summary='span-duration',x=>x.question.numeric={field:{kind:'attribute',key:'size'},type:'int64',unit:'items',parse:'canonical-decimal'}]){const bad={...data,question:structuredClone(data.question)};change(bad);assert.throws(()=>distributionCustomization(bad,v.presentation));}assert.throws(()=>distributionCustomization(data,{...v.presentation,value:'numeric.mean'}));assert.throws(()=>distributionCustomization(data,{...v.presentation,kind:'line'}));const edited=D.update(staged.displays[v.data.question.signal],{...D.presentation(staged.displays[v.data.question.signal]),kind:v.presentation.kind==='bar'?'table':'bar'},v.data.question);assert.equal(edited.display.value,'count');assert.equal(JSON.stringify(v),before,'handoff/reuse mutated source or destination');out.push(staged);}process.stdout.write(JSON.stringify(out));`
