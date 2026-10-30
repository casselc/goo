package verification

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"strconv"
	"strings"
	"testing"

	"github.com/casselc/goo/exploration"
	"hegel.dev/go/hegel"
)

type snapshotTransport func(*http.Request) (*http.Response, error)

func (f snapshotTransport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

// Use the actual Go source DTO producer, not a parallel hand-written JS schema.
// The fake physical source avoids external services in generated constructor
// checks; the SDK/ClickHouse browser journey validates the real source boundary.
func snapshotSourceFixture(seed int) (*exploration.ViewDataset, exploration.Query, error) {
	scale := seed%9 + 1
	binding := exploration.NumericBinding{Field: exploration.Field{Kind: "attribute", Key: "cost"}, Type: "decimal64", Scale: scale, Unit: "USD", Parse: "canonical-decimal"}
	q := exploration.Query{Signal: "spans", Environment: "prod", FromMS: 1000, ToMS: 61000, Limit: 100, GroupBy: []exploration.Field{{Kind: "column", Key: "service"}}, Numeric: &binding}
	n := uint64(9007199254740993 + seed)
	count := strconv.FormatUint(n, 10)
	total := strconv.FormatUint(2*n, 10)
	coefficient := strconv.Itoa(seed + 1)
	part := exploration.NumericSummary{Scale: scale, State: "observed", Unit: "USD", Count: count, Valid: "1", Missing: strconv.FormatUint(n-1, 10), Empty: "0", Invalid: "0", Sum: coefficient, Min: coefficient, Max: coefficient}
	all := part
	all.Count = total
	all.Valid = "2"
	all.Missing = strconv.FormatUint(2*(n-1), 10)
	all.Sum = strconv.Itoa(2 * (seed + 1))
	reads := 0
	source := exploration.Source{Endpoint: "http://source.invalid", Client: &http.Client{Transport: snapshotTransport(func(req *http.Request) (*http.Response, error) {
		reads++
		body, _ := io.ReadAll(req.Body)
		sql := string(body)
		text := ""
		switch {
		case strings.Contains(sql, " AS timestamp_ns"):
			text = `{"timestamp_ns":"1500000000","duration_ns":"1"}`
		case strings.Contains(sql, "SELECT DISTINCT"):
		case strings.Contains(sql, "GROUP BY"):
			a, _ := json.Marshal(map[string]any{"g0": "service-a", "count": count, "numeric": part})
			b, _ := json.Marshal(map[string]any{"g0": "service-b", "count": count, "numeric": part})
			text = string(a) + "\n" + string(b)
		default:
			raw, _ := json.Marshal(map[string]any{"count": total, "numeric": all})
			text = string(raw)
		}
		return &http.Response{StatusCode: 200, Header: http.Header{}, Body: io.NopCloser(strings.NewReader(text))}, nil
	})}}
	result, err := source.Fetch(context.Background(), q)
	if err != nil {
		return nil, q, err
	}
	if reads != 4 || result.ViewDataset == nil || result.ViewDataset.Completeness == nil || len(result.ViewDataset.Rows) != 2 {
		return nil, q, fmt.Errorf("source fixture lost its checked exhaustive two-group result: reads%d", reads)
	}
	return result.ViewDataset, q, nil
}

func runSnapshotJS(snapshot *exploration.ViewDataset, question exploration.Query, parameters map[string]any, script string) ([]byte, error) {
	payload := map[string]any{"snapshot": snapshot, "question": question}
	for key, value := range parameters {
		payload[key] = value
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	cmd := exec.Command("node", "--input-type=module", "-e", script)
	cmd.Stdin = bytes.NewReader(raw)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return out, fmt.Errorf("actual JS snapshot constructor: %w: %s", err, out)
	}
	return out, nil
}

const snapshotProjectionScript = `
import assert from 'node:assert/strict';import fs from 'node:fs';
import {createSourceAdapter,importSnapshot,asSnapshot,isSnapshot,getSnapshotInfo,hasGuarantee,selectRows} from '../ui/snapshots.mjs';import {validateDisplay} from '../ui/view-model.mjs';
const {snapshot:raw,question,mask=1,seed=0}=JSON.parse(fs.readFileSync(0,'utf8'));let unexpectedReads=0;globalThis.fetch=()=>{unexpectedReads++;throw Error('Unexpected constructor source IO')};
const adapter=createSourceAdapter({id:'source-'+seed});const source=adapter.adoptSnapshot(raw,{question});
assert(isSnapshot(source));assert(validateDisplay(source,{kind:'table',requiresComplete:true}).ok);assert.equal(getSnapshotInfo(source).entry,'source');assert.equal(getSnapshotInfo(source).sourceId,'source-'+seed);assert(hasGuarantee(source,'groups'));
assert.notEqual(source,raw);assert.equal(asSnapshot(source),source);assert.equal(adapter.adoptSnapshot(source,{question}),source);
const before=JSON.stringify(source);raw.rows[0].values.count='0';assert.equal(JSON.stringify(source),before);
function frozen(value){if(value&&typeof value==='object'){assert(Object.isFrozen(value));for(const child of Object.values(value))frozen(child)}}frozen(source);
assert(!Object.hasOwn(source,'sourceInstance'));assert(!Object.hasOwn(source,'guarantees'));const saved=structuredClone(source);assert.equal(validateDisplay(saved,{kind:'table',requiresComplete:true}).ok,false);assert.equal(isSnapshot(saved),false);assert.equal(getSnapshotInfo(saved),null);assert.equal(hasGuarantee(saved,'groups'),false);
const imported=importSnapshot(saved);assert.notEqual(imported.id,source.id);assert.notEqual(imported.revision,source.revision);const importedAgain=importSnapshot(saved);assert.notEqual(importedAgain.id,imported.id);assert.notEqual(importedAgain.revision,imported.revision);assert.equal(asSnapshot(imported),imported);assert.equal(getSnapshotInfo(imported).entry,'held');assert.equal(hasGuarantee(imported,'groups'),false);assert.equal(imported.completeness,undefined);assert.equal(imported.quality.extent,'unknown');assert.equal(imported.origin,'browser-held');assert.deepEqual(imported.rows,source.rows);frozen(imported);
const keys=source.rows.filter((_,i)=>mask&(1<<i)).map(row=>row.key).reverse();const selected=selectRows(source,keys);assert.equal(selected.rows.length,keys.length);
if(keys.length===source.rows.length){assert.equal(selected,source);assert(hasGuarantee(selected,'groups'))}else{assert.notEqual(selected,source);assert.equal(getSnapshotInfo(selected).entry,'projection');assert.equal(hasGuarantee(selected,'groups'),false);assert.equal(selected.completeness,undefined);assert.equal(selected.origin,'browser-held');assert.equal(selected.fields,source.fields);assert.deepEqual(selected.rows,source.rows.filter(row=>keys.includes(row.key)));for(const row of selected.rows){const original=source.rows.find(value=>value.key===row.key);assert.equal(row,original);assert.equal(row.values,original.values);assert.equal(row.cells,original.cells)}}
assert.equal(validateDisplay(selected,{kind:'table',requiresComplete:true}).ok,keys.length===source.rows.length);frozen(selected);assert.equal(JSON.stringify(source),before);assert(hasGuarantee(source,'groups'));
const other=createSourceAdapter({id:'source-'+seed});assert.throws(()=>other.adoptSnapshot(source,{question}),error=>error.code==='source-instance');const otherSource=other.adoptSnapshot(structuredClone(source),{question});assert.notEqual(getSnapshotInfo(otherSource).sourceInstance,getSnapshotInfo(source).sourceInstance);
assert.throws(()=>selectRows(source,['foreign']),error=>error.code==='projection-keys');assert.throws(()=>selectRows(source,[source.rows[0].key,source.rows[0].key]),error=>error.code==='projection-keys');assert.throws(()=>selectRows(saved,[source.rows[0].key]),error=>error.code==='snapshot-required');
assert.equal(unexpectedReads,0);process.stdout.write(JSON.stringify({rows:source.rows.length,selected:selected.rows.length,complete:hasGuarantee(selected,'groups')}));
`

func TestHegelSnapshotConstructionProjection(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		seed := hegel.Draw(tc, hegel.Integers(0, 10000))
		mask := hegel.Draw(tc, hegel.Integers(0, 3))
		raw, q, err := snapshotSourceFixture(seed)
		if err != nil {
			tc.Fatal(err)
		}
		if _, err = runSnapshotJS(raw, q, map[string]any{"seed": seed, "mask": mask}, snapshotProjectionScript); err != nil {
			tc.Fatal(err)
		}
	}, options(50)...)
}

const snapshotDefensiveScript = `
import assert from 'node:assert/strict';import fs from 'node:fs';
import {createSourceAdapter,importSnapshot,hasGuarantee} from '../ui/snapshots.mjs';
const {snapshot:raw,question,mode=0,seed=0}=JSON.parse(fs.readFileSync(0,'utf8'));const adapter=createSourceAdapter({id:'defensive-'+seed});
let calls=0;const expected=['accessor','cycle','prototype','non-json','budget-depth','budget-bytes','budget-nodes'];
const bad=structuredClone(raw);
switch(mode){case 0:Object.defineProperty(bad,'bad',{enumerable:true,get(){calls++;return 1}});break;case 1:bad.bad=bad;break;case 2:Object.setPrototypeOf(bad,{untrusted:true});break;case 3:bad.bad=()=>1;break;case 4:{let at=bad;for(let i=0;i<34;i++){at.bad={};at=at.bad}break}case 5:bad.bad='x'.repeat(8*1024*1024+seed+1);break;case 6:bad.bad=Array(250001).fill(null);break}
assert.throws(()=>adapter.adoptSnapshot(bad,{question}),error=>error.code===expected[mode]);assert.equal(calls,0);assert.equal(hasGuarantee(bad,'groups'),false);
assert.throws(()=>adapter.adoptSnapshot(raw,{question:{...question,toMs:question.toMs+1}}),error=>error.code==='source-question');
const imported=importSnapshot(raw);assert.equal(imported.completeness,undefined);assert.equal(hasGuarantee(imported,'groups'),false);
process.stdout.write('ok');
`

func TestHegelSnapshotDefensiveBoundary(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		seed := hegel.Draw(tc, hegel.Integers(0, 1000))
		mode := hegel.Draw(tc, hegel.Integers(0, 6))
		raw, q, err := snapshotSourceFixture(seed)
		if err != nil {
			tc.Fatal(err)
		}
		if _, err = runSnapshotJS(raw, q, map[string]any{"seed": seed, "mode": mode}, snapshotDefensiveScript); err != nil {
			tc.Fatal(err)
		}
	}, options(30)...)
}
