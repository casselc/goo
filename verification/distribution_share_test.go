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

// The group denominator is emitted by the actual Go source DTO producer. The
// bounded fake physical reads make extreme/null/empty cases reproducible; the
// executable browser journey separately checks real collector/ClickHouse data.
func distributionSourceFixture(seed, mode int) (*exploration.ViewDataset, exploration.Query, error) {
	q := exploration.Query{Signal: "logs", Environment: "prod", FromMS: 1000, ToMS: 61000, Limit: 100, GroupBy: []exploration.Field{{Kind: "attribute", Key: "tenant"}}}
	counts := []string{strconv.Itoa(seed + 1), strconv.Itoa(seed%3 + 1), strconv.Itoa(seed + 3)}
	if mode == 1 {
		counts = []string{"18446744073709551613", "1", "1"}
	}
	total := new(big.Int)
	for _, s := range counts {
		n, _ := new(big.Int).SetString(s, 10)
		total.Add(total, n)
	}
	groups := []map[string]any{{"g0": nil, "count": counts[0]}, {"g0": "", "count": counts[1]}, {"g0": "worker", "count": counts[2]}}
	if mode == 2 {
		groups = nil
		total.SetInt64(0)
	}
	countRead := new(big.Int).Set(total)
	if mode == 3 {
		countRead.Add(countRead, big.NewInt(1))
	}
	reads := 0
	source := exploration.Source{Endpoint: "http://source.invalid", Client: &http.Client{Transport: snapshotTransport(func(req *http.Request) (*http.Response, error) {
		reads++
		body, _ := io.ReadAll(req.Body)
		sql := string(body)
		text := ""
		switch {
		case strings.Contains(sql, "SELECT DISTINCT"):
		case strings.Contains(sql, "GROUP BY"):
			for _, group := range groups {
				raw, _ := json.Marshal(group)
				text += string(raw) + "\n"
			}
		case strings.Contains(sql, "count()"):
			text = `{"count":"` + countRead.String() + `"}`
		}
		return &http.Response{StatusCode: 200, Header: http.Header{}, Body: io.NopCloser(strings.NewReader(text))}, nil
	})}}
	result, err := source.Fetch(context.Background(), q)
	if err != nil {
		return nil, q, err
	}
	if reads != 4 || result.ViewDataset == nil {
		return nil, q, fmt.Errorf("missing actual group DTO: reads%d", reads)
	}
	if mode == 3 && result.ViewDataset.Partition != nil {
		return nil, q, fmt.Errorf("separate count disagreement supplied a full partition")
	}
	return result.ViewDataset, q, nil
}

const distributionShareScript = `
import assert from 'node:assert/strict';import fs from 'node:fs';
import {createSourceAdapter,importSnapshot,selectRows,hasGuarantee} from '../ui/snapshots.mjs';import {populationShare,displayRows,validateDisplay} from '../ui/view-model.mjs';
const {snapshot:raw,question,seed=0,mode=0}=JSON.parse(fs.readFileSync(0,'utf8'));let reads=0;globalThis.fetch=()=>{reads++;throw Error('Display operations cannot read sources')};
const source=createSourceAdapter({id:'shares-'+seed}).adoptSnapshot(raw,{question}),before=JSON.stringify(source),rows=source.rows;
const options={kind:'bar',category:'category',value:'count',scale:'population',order:'value-desc'};
const total=rows.reduce((sum,row)=>sum+BigInt(row.values.count),0n);
if(mode===3){assert.equal(source.partition,undefined);assert(!validateDisplay(source,options).ok);for(const row of rows)assert(!populationShare(source,row.key).ok)}else{
 assert.deepEqual(source.partition,{kind:'disjoint-record-groups',measure:'count',total:String(total)});assert(hasGuarantee(source,'groups'));
 assert.equal(validateDisplay(source,options).ok,total>0n);
 if(total===0n){assert.equal(rows.length,0);assert(!populationShare(source,'missing').ok)}else{
  for(const row of rows){const share=populationShare(source,row.key),n=BigInt(row.values.count);assert(share.ok);assert.equal(share.count,String(n));assert.equal(share.total,String(total));assert.equal(share.fraction,String(n)+'/'+String(total));assert(Number.isFinite(share.coordinate));if(n>0n&&n<total){assert.notEqual(share.percent,'0%');assert.notEqual(share.percent,'100%');}}
  assert.equal(rows.find(row=>row.values.category===null).values.category,null);assert.equal(rows.find(row=>row.values.category==='').values.category,'');assert.notEqual(rows[0].key,rows[1].key);
  const sorted=displayRows(source,options);assert.equal(sorted.length,rows.length);assert(sorted.every(row=>rows.includes(row)));for(let i=1;i<sorted.length;i++)assert(BigInt(sorted[i-1].values.count)>=BigInt(sorted[i].values.count));
  assert.equal(displayRows(source,{...options,query:'missing'})[0],rows.find(row=>row.values.category===null));assert.equal(displayRows(source,{...options,query:'present empty string'})[0],rows.find(row=>row.values.category===''));
  const filtered=displayRows(source,{...options,query:'worker'});assert.equal(filtered.length,1);assert.equal(filtered[0],rows.find(row=>row.values.category==='worker'));assert.deepEqual(populationShare(source,filtered[0].key),populationShare(source,rows.find(row=>row.values.category==='worker').key));
  const imported=importSnapshot(source);assert.deepEqual(imported.partition,source.partition);assert(!hasGuarantee(imported,'groups'));assert(!validateDisplay(imported,options).ok);assert(!populationShare(imported,imported.rows[0].key).ok);
  const subset=selectRows(source,[rows[0].key]);assert.deepEqual(subset.partition,source.partition);assert(!validateDisplay(subset,options).ok);assert(!populationShare(subset,subset.rows[0].key).ok);
  const partialRaw=structuredClone(source);partialRaw.quality.extent='partial';delete partialRaw.completeness;const partial=createSourceAdapter({id:'partial-'+seed}).adoptSnapshot(partialRaw,{question});assert(!populationShare(partial,partial.rows[0].key).ok);
 }
}
assert.equal(JSON.stringify(source),before);assert.equal(reads,0);process.stdout.write('ok');
`

func TestHegelDistributionPartitionShares(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		seed := hegel.Draw(tc, hegel.Integers(0, 10000))
		mode := hegel.Draw(tc, hegel.Integers(0, 3))
		raw, q, err := distributionSourceFixture(seed, mode)
		if err != nil {
			tc.Fatal(err)
		}
		if _, err = runSnapshotJS(raw, q, map[string]any{"seed": seed, "mode": mode}, distributionShareScript); err != nil {
			tc.Fatal(err)
		}
	}, options(60)...)
}

const distributionMeasureConservationScript = `
import assert from 'node:assert/strict';import fs from 'node:fs';import {createSourceAdapter} from '../ui/snapshots.mjs';import {displayRows,validateDisplay,populationShare} from '../ui/view-model.mjs';
const {snapshot:raw,question,seed=0}=JSON.parse(fs.readFileSync(0,'utf8'));const source=createSourceAdapter({id:'measure-shares-'+seed}).adoptSnapshot(raw,{question}),before=JSON.stringify(source);
for(const value of ['count','numeric.mean','numeric.sum','numeric.min','numeric.max'])for(const order of ['source','label','value-desc','value-asc']){
 const display={kind:'table',category:'category',value,order,query:'service-b'};assert(validateDisplay(source,display).ok);const rows=displayRows(source,display);assert.equal(rows.length,1);assert.equal(rows[0],source.rows[1]);assert.equal(rows[0].values,source.rows[1].values);assert.equal(rows[0].cells,source.rows[1].cells);const share=populationShare(source,rows[0].key);assert.equal(BigInt(share.count)*2n,BigInt(share.total));assert.equal(share.fraction,share.count+'/'+share.total);
}
assert.equal(JSON.stringify(source),before);process.stdout.write('ok');
`

func TestHegelDistributionMeasureConservation(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		seed := hegel.Draw(tc, hegel.Integers(0, 10000))
		raw, q, err := snapshotSourceFixture(seed)
		if err != nil {
			tc.Fatal(err)
		}
		if _, err = runSnapshotJS(raw, q, map[string]any{"seed": seed}, distributionMeasureConservationScript); err != nil {
			tc.Fatal(err)
		}
	}, options(40)...)
}
