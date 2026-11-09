package verification

import (
	"encoding/json"
	"fmt"
	"io"
	"maps"
	"net/http"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/workbench"
	"hegel.dev/go/hegel"
)

type patternDisplayVector391 struct {
	Name        string            `json:"name"`
	Question    exploration.Query `json:"question"`
	Declaration json.RawMessage   `json:"declaration"`
	Accepted    bool              `json:"accepted"`
}

// Expected admission is the closed public profile, rather than the production
// validator's verdict. Each generated draw checks every boundary mutation.
func TestHegelLogPatternDisplayContract391(t *testing.T) {
	vectors := []patternDisplayVector391{}
	hegel.Test(t, func(tc *hegel.T) {
		seed := hegel.Draw(tc, hegel.Integers(1, 100000))
		for mutation := -1; mutation < 17; mutation++ {
			q := exploration.Query{Signal: "logs", Environment: "prod", FromMS: 1000, ToMS: 2000, Limit: 1, GroupBy: []exploration.Field{{Kind: "column", Key: "body_pattern_v1"}}}
			display := map[string]any{"profile": "log-pattern-count-v1", "category": "category", "value": "count", "scale": []string{"max", "population"}[seed%2], "order": []string{"source", "label", "value-desc", "value-asc"}[seed%4], "query": []string{"", "order", "unclassified", "雪"}[seed%4], "requiresComplete": seed%2 == 0}
			declaration := map[string]any{"version": "goo.view/1", "scope": map[string]any{"kind": "record-query", "signal": "logs"}, "kind": map[string]any{"name": []string{"bar", "table"}[seed%2], "version": 1}, "display": display}
			switch mutation {
			case 0:
				delete(display, "profile")
			case 1:
				display["profile"] = ""
			case 2:
				display["profile"] = nil
			case 3:
				display["profile"] = 1
			case 4:
				display["profile"] = "log-pattern-count-v2"
			case 5:
				q.Signal = "spans"
				declaration["scope"].(map[string]any)["signal"] = "spans"
			case 6:
				q.GroupBy[0].Key = "service"
			case 7:
				q.GroupBy = append(q.GroupBy, exploration.Field{Kind: "column", Key: "service"})
			case 8:
				q.GroupTimeSeries = true
			case 9:
				q.Numeric = &exploration.NumericBinding{Field: exploration.Field{Kind: "attribute", Key: "work.delta"}, Type: "int64", Unit: "items", Parse: "canonical-decimal"}
			case 10:
				display["value"] = "numeric.mean"
			case 11:
				declaration["kind"].(map[string]any)["name"] = "line"
			case 12:
				display["profile"] = "ordinary"
			case 13, 14, 15, 16:
				q.GroupBy[0].Key = "service"
				delete(display, "profile")
				if mutation == 14 {
					display["profile"] = ""
				}
				if mutation == 15 {
					display["profile"] = nil
				}
				if mutation == 16 {
					display["profile"] = 1
				}
			}
			wire, _ := json.Marshal(declaration)
			var d workbench.ViewDeclaration
			err := json.Unmarshal(wire, &d)
			accepted := err == nil && d.Validate(q) == nil
			expected := mutation < 0 || mutation >= 13 && mutation <= 15
			if accepted != expected {
				tc.Fatalf("profile mutation %d: accepted=%v decode=%v", mutation, accepted, err)
			}
			vectors = append(vectors, patternDisplayVector391{fmt.Sprintf("%d/%d", seed, mutation), q, wire, expected})
			if !accepted {
				continue
			}
			destination := q
			destination.Environment = "destination"
			destination.FromMS = 3000
			destination.ToMS = 4000
			before, _ := json.Marshal(q)
			reused, e := workbench.ReuseViewDeclaration(d, q, destination)
			if e != nil || !reflect.DeepEqual(reused, d) {
				tc.Fatal("compatible destination refused", e)
			}
			after, _ := json.Marshal(q)
			if string(before) != string(after) {
				tc.Fatal("reuse changed donor query")
			}
		}
	}, options(24)...)
	if len(vectors) == 0 {
		t.Fatal("generated no shared contract vectors")
	}
	if err := runNativeHistogramJS(map[string]any{"cases": vectors}, patternDisplayContractJS391); err != nil {
		t.Fatal(err)
	}
	t.Logf("%d shared Go/JavaScript profile vectors", len(vectors))
}

const patternDisplayContractJS391 = `import fs from 'node:fs';import assert from 'node:assert/strict';import * as D from '../ui/view-declaration.mjs';
for(const v of JSON.parse(fs.readFileSync(0,'utf8')).cases){let out;try{out=D.validate(v.declaration,v.question)}catch{}assert.equal(!!out,v.accepted,v.name);if(!out)continue;assert.deepEqual(out,v.declaration);const before=JSON.stringify(v);const destination={...structuredClone(v.question),environment:'destination',fromMs:3000,toMs:4000};assert.deepEqual(D.reuse(out,v.question,destination),out);const p=D.presentation(out);const edited=D.update(out,{...p,kind:p.kind==='bar'?'table':'bar',query:'unclassified',order:'label'},v.question);if(D.logPattern(v.question))assert.equal(edited.display.profile,'log-pattern-count-v1');assert.equal(edited.display.value,'count');assert.equal(JSON.stringify(v),before);const ordinary={...structuredClone(v.question),groupBy:[{kind:'column',key:'service'}]};if(D.logPattern(v.question))assert.throws(()=>D.reuse(out,v.question,ordinary));}
`

// Generated source responses are deliberately synthetic. The observed journey
// separately proves real OTel/ClickHouse data; this property covers projection
// refusal and local display laws over bounded adversarial populations.
func TestHegelLogPatternDisplayProjection391(t *testing.T) {
	fixtures := []map[string]any{}
	hegel.Test(t, func(tc *hegel.T) {
		seed := hegel.Draw(tc, hegel.Integers(1, 100000))
		prefix := fmt.Sprintf("team-%c ", 'a'+rune(seed%5))
		bodies := []string{prefix + fmt.Sprintf("order %d ready", seed), prefix + "order 0 ready", "order ⟪digits⟫ ready", "order <num> ready", "", "line\n\x00\\1", strings.Repeat("x", 257), string([]byte{0xff})}
		for repeat := 0; repeat < seed%4; repeat++ {
			bodies = append(bodies, prefix+"order 999 ready")
		}
		counts := map[string]int{}
		labels := map[string]string{}
		for _, body := range bodies {
			p := lpOracle(body)
			counts[p.Identity]++
			labels[p.Identity] = p.Label
		}
		ids := slices.Sorted(maps.Keys(counts))
		groups := []string{}
		for _, id := range ids {
			b, _ := json.Marshal(map[string]string{"g0": id, "count": strconv.Itoa(counts[id])})
			groups = append(groups, string(b))
		}
		q := exploration.Query{Signal: "logs", Environment: "prod", FromMS: 1000, ToMS: 2000, Limit: 1, GroupBy: []exploration.Field{{Kind: "column", Key: "body_pattern_v1"}}}
		for _, changed := range []bool{false, true} {
			reads := 0
			total := len(bodies)
			if changed {
				total++
			}
			plan, e := exploration.Compile(q)
			if e != nil {
				tc.Fatal(e)
			}
			source := exploration.Source{Endpoint: "http://source.invalid", Client: &http.Client{Transport: snapshotTransport(func(req *http.Request) (*http.Response, error) {
				reads++
				wire, _ := io.ReadAll(req.Body)
				body := ""
				switch string(wire) {
				case plan.Count:
					body = fmt.Sprintf(`{"count":"%d"}`, total)
				case plan.Groups:
					body = strings.Join(groups, "\n")
				case plan.Rows:
					first, _ := json.Marshal(map[string]string{"timestamp_ns": "1000000000", "Body": bodies[0]})
					second, _ := json.Marshal(map[string]string{"timestamp_ns": "1000000001", "Body": bodies[1]})
					body = string(first) + "\n" + string(second)
				case plan.Fields:
				default:
					return nil, fmt.Errorf("unexpected generated read")
				}
				return &http.Response{StatusCode: 200, Header: http.Header{}, Body: io.NopCloser(strings.NewReader(body))}, nil
			})}}
			result, e := source.Fetch(t.Context(), q)
			if e != nil || result.ViewDataset == nil || reads != 4 {
				tc.Fatalf("generated projection: reads=%d error=%v result=%+v", reads, e, result)
			}
			d := result.ViewDataset
			if d.Fields[0].Format != "log-pattern-v1" || d.Fields[0].Nullable || len(d.Rows) != len(counts) {
				tc.Fatal("pattern category lost canonical profile")
			}
			for _, row := range d.Rows {
				id := *row.Values["category"]
				if !lpCanonical(id) || *row.Values["count"] != strconv.Itoa(counts[id]) {
					tc.Fatal("source projection changed full count partition")
				}
			}
			if changed {
				if d.Partition != nil || d.Completeness != nil {
					tc.Fatal("changed read gained population authority")
				}
			} else if d.Partition == nil || d.Partition.Total != strconv.Itoa(len(bodies)) || d.Completeness == nil {
				tc.Fatal("complete partition lost full denominator")
			}
			fixtures = append(fixtures, map[string]any{"snapshot": d, "labels": labels, "counts": counts, "total": len(bodies), "changed": changed})
		}
	}, options(16)...)
	if e := runNativeHistogramJS(map[string]any{"cases": fixtures}, patternDisplayProjectionJS391); e != nil {
		t.Fatal(e)
	}
	t.Logf("%d shared source projection cases", len(fixtures))
}

const patternDisplayProjectionJS391 = `import fs from 'node:fs';import assert from 'node:assert/strict';import {displayRows,populationShare,groupIntent,validateSnapshot} from '../ui/view-model.mjs';import {categoryLabel} from '../ui/view-values.mjs';import {createSourceAdapter} from '../ui/snapshots.mjs';import {createActionContext} from '../ui/scope-actions.mjs';
for(const f of JSON.parse(fs.readFileSync(0,'utf8')).cases){const s=createSourceAdapter({id:'generated391'}).adoptSnapshot(f.snapshot,{question:f.snapshot.scope.selector});let epoch=1;const owner=createActionContext({epoch:()=>epoch,active:()=>true,canStage:()=>true});const binding=owner.bind({role:'current',slot:'groups',epoch,snapshot:s,question:s.scope.selector,coherent:!f.changed});const checked=validateSnapshot(s);assert.equal(checked.ok,true,JSON.stringify(checked));const before=JSON.stringify(s);const field=s.fields.find(x=>x.key==='category');const base={kind:'bar',category:'category',value:'count',scale:'population',order:'source',query:''};for(const row of s.rows){const id=row.values.category;assert.equal(categoryLabel(field,id),f.labels[id]);const share=populationShare(s,row.key);assert.equal(share.ok,!f.changed);if(share.ok)assert.equal(share.fraction,f.counts[id]+'/'+f.total);const intent=groupIntent(s,row.key);assert.equal(intent.values[0].value,id);assert.equal(owner.resolve(binding,intent).ok,!f.changed);}
for(const query of ['order','unclassified','\\n','⟪digits⟫']){const rows=displayRows(s,{...base,query,order:'value-desc'});const wanted=s.rows.filter(r=>f.labels[r.values.category].toLowerCase().includes(query.toLowerCase())).map((r,i)=>({r,i})).sort((a,b)=>f.counts[b.r.values.category]-f.counts[a.r.values.category]||a.i-b.i).map(x=>x.r.key);assert.deepEqual(rows.map(r=>r.key),wanted);for(const row of rows){const share=populationShare(s,row.key);if(share.ok)assert.equal(share.fraction,f.counts[row.values.category]+'/'+f.total);}}assert.equal(JSON.stringify(s),before,'local Find/order changed supplied population');epoch++;assert.equal(owner.resolve(binding,groupIntent(s,s.rows[0].key)).ok,false,'stale owner gained action authority');}
`
