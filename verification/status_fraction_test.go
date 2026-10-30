package verification

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"strconv"
	"strings"
	"testing"

	"github.com/casselc/goo/exploration"
	"hegel.dev/go/hegel"
)

type statusFractionFixture struct {
	Groups      *exploration.ViewDataset `json:"groups"`
	Time        *exploration.ViewDataset `json:"time"`
	Summary     *exploration.ViewDataset `json:"summary"`
	Question    exploration.Query        `json:"question"`
	UnknownTime *exploration.ViewDataset `json:"unknownTime"`
}

func statusFractionSourceFixture(seed, mode int) (statusFractionFixture, error) {
	q := exploration.Query{Signal: "spans", Environment: "prod", FromMS: 1000, ToMS: 3500, BucketMS: 1000, Limit: 100, SpanStatus: true, GroupBy: []exploration.Field{{Kind: "attribute", Key: "route"}}}
	a := []string{strconv.Itoa(seed % 7), strconv.Itoa(seed + 10), strconv.Itoa(seed % 5), strconv.Itoa(seed % 3)}
	b := []string{strconv.Itoa(seed%2 + 1), "1", "2", "1"}
	if mode == 1 {
		a = []string{"1", "18446744073709551607", "1", "2"}
		b = []string{"0", "1", "2", "1"}
	}
	if mode == 4 {
		a = []string{"18446744073709551607", "1", "1", "2"}
		b = []string{"3", "1", "0", "0"}
	}
	if mode == 3 {
		a = []string{"0", "0", strconv.Itoa(seed + 1), "2"}
		b = []string{"0", "0", "2", "1"}
	}
	total := []string{"0", "0", "0", "0"}
	for i := range total {
		x, _ := new(big.Int).SetString(a[i], 10)
		y, _ := new(big.Int).SetString(b[i], 10)
		total[i] = new(big.Int).Add(x, y).String()
	}
	if mode == 2 {
		total = []string{"0", "0", "0", "0"}
	}
	row := func(parts []string) map[string]any {
		n := new(big.Int)
		for _, part := range parts {
			x, _ := new(big.Int).SetString(part, 10)
			n.Add(n, x)
		}
		return map[string]any{"count": n.String(), "status_error": parts[0], "status_ok": parts[1], "status_unset": parts[2], "status_unknown": parts[3]}
	}
	reads := 0
	source := exploration.Source{Endpoint: "http://source.invalid", Client: &http.Client{Transport: snapshotTransport(func(req *http.Request) (*http.Response, error) {
		reads++
		body, _ := io.ReadAll(req.Body)
		sql := string(body)
		text := ""
		emit := func(value any) { raw, _ := json.Marshal(value); text += string(raw) + "\n" }
		switch {
		case strings.Contains(sql, "SELECT DISTINCT"):
		case strings.Contains(sql, "GROUP BY from_ms"):
			if mode != 2 {
				r := row(a)
				r["from_ms"] = "1000"
				emit(r)
				r = row(b)
				r["from_ms"] = "2000"
				emit(r)
			}
		case strings.Contains(sql, "GROUP BY"):
			if mode != 2 {
				r := row(a)
				r["g0"] = nil
				emit(r)
				r = row(b)
				r["g0"] = ""
				emit(r)
			}
		case strings.Contains(sql, "count()"):
			emit(row(total))
		}
		return &http.Response{StatusCode: 200, Header: http.Header{}, Body: io.NopCloser(strings.NewReader(text))}, nil
	})}}
	result, err := source.Fetch(context.Background(), q)
	if err != nil {
		return statusFractionFixture{}, err
	}
	if reads != 5 || result.ViewDataset == nil || result.TimeDataset == nil || result.SummaryDataset == nil {
		return statusFractionFixture{}, fmt.Errorf("status source DTOs missing: reads%d groups%t time%t summary%t", reads, result.ViewDataset != nil, result.TimeDataset != nil, result.SummaryDataset != nil)
	}
	var unknown exploration.ViewDataset
	rawUnknown, _ := json.Marshal(result.TimeDataset)
	if err = json.Unmarshal(rawUnknown, &unknown); err != nil {
		return statusFractionFixture{}, err
	}
	unknown.Origin = "browser-held"
	unknown.Quality.Extent = "unknown"
	unknown.Completeness = nil
	rowUnknown := &unknown.Rows[len(unknown.Rows)-1]
	rowUnknown.State = "unknown"
	rowUnknown.Values["count"] = nil
	for _, key := range []string{"status.error", "status.ok", "status.unset", "status.unknown", "status.error_fraction"} {
		rowUnknown.Values[key] = nil
		rowUnknown.Cells[key] = exploration.ViewCell{State: "unknown"}
	}
	return statusFractionFixture{Groups: result.ViewDataset, Time: result.TimeDataset, Summary: result.SummaryDataset, Question: q, UnknownTime: &unknown}, nil
}

const statusFractionScript = `
import assert from 'node:assert/strict';import fs from 'node:fs';import {createSourceAdapter,importSnapshot,selectRows} from '../ui/snapshots.mjs';import {validateSnapshot,validateDisplay,displayRows,fractionFacts} from '../ui/view-model.mjs';import {formatCell} from '../ui/view-values.mjs';
const {snapshot,fixture,seed=0}=JSON.parse(fs.readFileSync(0,'utf8'));const adapter=createSourceAdapter({id:'status-'+seed});let reads=0;globalThis.fetch=()=>{reads++;throw Error('Status presentation cannot fetch')};
const key='status.error_fraction',classes=['error','ok','unset','unknown'],gcd=(a,b)=>b===0n?a:gcd(b,a%b),ratio=(n,d)=>{const g=gcd(n,d);return d/g===1n?String(n/g):String(n/g)+'/'+String(d/g)};
for(const raw of [fixture.groups,fixture.time,fixture.summary]){
 const source=adapter.adoptSnapshot(raw,{question:fixture.question}),before=JSON.stringify(source),field=source.fields.find(f=>f.key===key);
 assert.deepEqual(field.aggregation,{family:'ratio',kind:'fraction',method:'exact-rational',numerator:'status.error',denominator:'count'});assert.equal(field.type,'rational');assert.equal(field.unit,'1');assert(field.nullable);
 for(const row of source.rows){
  const n=BigInt(row.values['status.error']),d=BigInt(row.values.count);assert.equal(classes.reduce((sum,c)=>sum+BigInt(row.values['status.'+c]),0n),d);
  for(const c of classes){const f=source.fields.find(f=>f.key==='status.'+c);assert.equal(f.type,'number');assert.equal(f.unit,'records');assert.deepEqual(f.aggregation,{family:'span-status',kind:'count',method:'exact'});assert.equal(row.cells[f.key].state,'derived');assert.equal(row.cells[f.key].support,undefined)}
  assert.equal(row.cells[key].support,undefined);const facts=fractionFacts(source,row,key);assert.equal(facts.numerator.value,row.values['status.error']);assert.equal(facts.denominator.value,row.values.count);assert.equal(facts.partition.length,4);
  if(d===0n){assert.equal(row.values[key],null);assert.equal(row.cells[key].state,'no_data');assert(!formatCell(field,row.values[key],row.cells[key]).includes('0%'))}else{assert.equal(row.values[key],ratio(n,d));assert.equal(row.cells[key].state,'derived');const human=formatCell(field,row.values[key],row.cells[key]);assert(human.includes('%'));if(n>0n&&n<d){assert.notEqual(human,'0%');assert.notEqual(human,'100%')}}
 }
 assert(!formatCell({key:'plain',type:'rational',unit:'1'},'1/2').includes('%'));
 const generic=structuredClone(source),alias='observed.failure_fraction';generic.fields.find(f=>f.key===key).key=alias;for(const row of generic.rows){row.values[alias]=row.values[key];row.cells[alias]=row.cells[key];delete row.values[key];delete row.cells[key]}const genericHeld=importSnapshot(generic);assert(validateSnapshot(genericHeld).ok);for(const row of genericHeld.rows)if(row.values[alias]!==null)assert(formatCell(genericHeld.fields.find(f=>f.key===alias),row.values[alias],row.cells[alias]).includes('%'));
 const held=importSnapshot(source);assert(validateSnapshot(held).ok);for(const row of held.rows)assert.deepEqual(row.values,source.rows.find(r=>r.key===row.key).values);
 if(source.shape==='ranked'&&source.rows.length){const display={kind:'bar',category:'category',value:key,scale:'max',order:'value-desc'};assert(validateDisplay(source,display).ok);assert(validateDisplay(held,display).ok);const subset=selectRows(source,[source.rows[0].key]);assert(validateDisplay(subset,display).ok);assert.equal(subset.rows[0],source.rows[0]);for(const order of ['source','label','value-desc','value-asc']){const rows=displayRows(source,{...display,order});assert(rows.every(r=>source.rows.includes(r)));}assert.equal(JSON.stringify(source),before)}
 assert.equal(JSON.stringify(source),before);
}
const badEmpty=structuredClone(fixture.time);badEmpty.rows.at(-1).values[key]='0';badEmpty.rows.at(-1).cells[key]={state:'derived'};assert(!validateSnapshot(badEmpty).ok);const badUnknown=structuredClone(fixture.unknownTime);badUnknown.rows.at(-1).values[key]='0';assert(!validateSnapshot(badUnknown).ok);
const unknown=importSnapshot(fixture.unknownTime);assert(validateSnapshot(unknown).ok);const unknownRow=unknown.rows.at(-1);assert.equal(unknownRow.values.count,null);assert.equal(unknownRow.values[key],null);assert.equal(unknownRow.cells[key].state,'unknown');assert(!formatCell(unknown.fields.find(f=>f.key===key),null,unknownRow.cells[key]).includes('0%'));assert(validateDisplay(unknown,{kind:'time-bar',value:key}).ok);
assert.equal(reads,0);process.stdout.write('ok');
`

func TestHegelStatusFractionExactSourceRows(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		seed := hegel.Draw(tc, hegel.Integers(0, 10000))
		mode := hegel.Draw(tc, hegel.Integers(0, 4))
		fixture, err := statusFractionSourceFixture(seed, mode)
		if err != nil {
			tc.Fatal(err)
		}
		if _, err = runSnapshotJS(fixture.Groups, fixture.Question, map[string]any{"fixture": fixture, "seed": seed}, statusFractionScript); err != nil {
			tc.Fatal(err)
		}
	}, options(60)...)
}

const statusFractionMalformedScript = `
import assert from 'node:assert/strict';import fs from 'node:fs';import {validateSnapshot} from '../ui/view-model.mjs';
const {snapshot:raw,mode=0}=JSON.parse(fs.readFileSync(0,'utf8')),s=structuredClone(raw),r=s.rows[0],key='status.error_fraction',f=s.fields.find(f=>f.key===key);
switch(mode){
case 0:r.values['status.ok']=String(BigInt(r.values['status.ok'])+1n);break;
case 1:r.values[key]='2';break;
case 2:f.aggregation.numerator='not.declared';break;
case 3:f.aggregation.denominator='category';break;
case 4:r.values.count=9007199254740992;break;
case 5:r.cells[key].support={total:r.values.count,valid:r.values.count,missing:'0',empty:'0',invalid:'0'};break;
case 6:r.cells['status.error'].state='estimated';break;
case 7:{const extra=structuredClone(s.fields.find(f=>f.key==='status.ok'));extra.key='status.extra';s.fields.push(extra);r.values[extra.key]='0';r.cells[extra.key]={state:'derived'};break}
case 8:r.values['status.error']='18446744073709551616';break;
case 9:f.aggregation.method='exact';break;
}
assert.equal(validateSnapshot(s).ok,false,JSON.stringify({mode,s}));process.stdout.write('ok');
`

func TestHegelStatusFractionMalformedContracts(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		seed := hegel.Draw(tc, hegel.Integers(0, 10000))
		mode := hegel.Draw(tc, hegel.Integers(0, 9))
		fixture, err := statusFractionSourceFixture(seed, 0)
		if err != nil {
			tc.Fatal(err)
		}
		if _, err = runSnapshotJS(fixture.Groups, fixture.Question, map[string]any{"mode": mode}, statusFractionMalformedScript); err != nil {
			tc.Fatal(err)
		}
	}, options(40)...)
}
