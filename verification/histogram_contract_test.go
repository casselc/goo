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

func histogramContractFixture(typ string, scale, seed, mode int) (*exploration.ViewDataset, exploration.Query, []map[string]string, map[string]string, error) {
	b := exploration.NumericBinding{Field: exploration.Field{Kind: "attribute", Key: "work.amount"}, Type: typ, Unit: "items", Parse: "canonical-decimal"}
	edgeType := typ
	if typ == "decimal64" {
		b.Scale = scale
		b.Unit = "USD"
		edgeType = "int64"
	}
	edges := histogramOracleEdges(edgeType)
	values := []*string{nil}
	empty := ""
	bad := "NaN"
	plus := "+1"
	values = append(values, &empty, &bad, &plus)
	add := func(text string) {
		value := text
		if typ == "decimal64" {
			value = decText(text, scale)
		}
		values = append(values, &value)
	}
	switch mode {
	case 1:
	case 2:
		values = nil
	default:
		for _, edge := range edges {
			n, _ := new(big.Int).SetString(edge, 10)
			for _, delta := range []int64{-1, 0, 1} {
				add(new(big.Int).Add(n, big.NewInt(delta)).String())
			}
		}
		for i := 0; i < seed%10; i++ {
			add(strconv.Itoa(seed + i))
		}
	}
	support := map[string]string{"total": strconv.Itoa(len(values)), "valid": "0", "missing": "0", "empty": "0", "invalid": "0"}
	counts := make([]int, len(edges)-1)
	for _, raw := range values {
		var n *big.Int
		state := "invalid"
		if typ == "decimal64" {
			n, state = decCoefficient(raw, scale)
		} else {
			state = predicateState(raw, typ)
			if state == "valid" {
				n, _ = new(big.Int).SetString(*raw, 10)
			}
		}
		current, _ := strconv.Atoi(support[state])
		support[state] = strconv.Itoa(current + 1)
		if state != "valid" {
			continue
		}
		found := false
		for i := range counts {
			lo, _ := new(big.Int).SetString(edges[i], 10)
			hi, _ := new(big.Int).SetString(edges[i+1], 10)
			if n.Cmp(lo) >= 0 && n.Cmp(hi) < 0 {
				counts[i]++
				found = true
				break
			}
		}
		if !found {
			return nil, exploration.Query{}, nil, nil, fmt.Errorf("oracle valid value outside coefficient domain")
		}
	}
	summary, err := exploration.EvaluateNumericValues(b, values)
	if err != nil {
		return nil, exploration.Query{}, nil, nil, err
	}
	hist, err := exploration.EvaluateNumericHistogram(b, values)
	if err != nil {
		return nil, exploration.Query{}, nil, nil, err
	}
	if summary.Valid != support["valid"] || summary.Missing != support["missing"] || summary.Empty != support["empty"] || summary.Invalid != support["invalid"] {
		return nil, exploration.Query{}, nil, nil, fmt.Errorf("producer coverage disagreed with independent oracle")
	}
	q := exploration.Query{Signal: "spans", Environment: "prod", FromMS: 1000, ToMS: 61000, Limit: 100, Numeric: &b, NumericHistogram: &exploration.NumericHistogramSpec{Binning: "powers-of-two"}}
	reads := 0
	source := exploration.Source{Endpoint: "http://source.invalid", Client: &http.Client{Transport: snapshotTransport(func(req *http.Request) (*http.Response, error) {
		reads++
		body, _ := io.ReadAll(req.Body)
		sql := string(body)
		text := ""
		if !strings.Contains(sql, "SELECT DISTINCT") && !strings.Contains(sql, " AS timestamp_ns") {
			if !strings.Contains(sql, "AS numeric_histogram") {
				return nil, fmt.Errorf("histogram was not in the count aggregate")
			}
			raw, _ := json.Marshal(map[string]any{"count": summary.Count, "numeric": summary, "numeric_histogram": hist.Bins})
			text = string(raw)
		}
		return &http.Response{StatusCode: 200, Header: http.Header{}, Body: io.NopCloser(strings.NewReader(text))}, nil
	})}}
	result, err := source.Fetch(context.Background(), q)
	if err != nil {
		return nil, q, nil, nil, err
	}
	if reads != 3 || result.HistogramDataset == nil {
		return nil, q, nil, nil, fmt.Errorf("missing same-read histogram DTO: reads%d", reads)
	}
	divisor := big.NewInt(1)
	if typ == "decimal64" {
		divisor.Exp(big.NewInt(10), big.NewInt(int64(scale)), nil)
	}
	rows := make([]map[string]string, len(counts))
	for i, count := range counts {
		lo, _ := new(big.Int).SetString(edges[i], 10)
		hi, _ := new(big.Int).SetString(edges[i+1], 10)
		rows[i] = map[string]string{"from": new(big.Rat).SetFrac(lo, divisor).RatString(), "to": new(big.Rat).SetFrac(hi, divisor).RatString(), "count": strconv.Itoa(count)}
	}
	return result.HistogramDataset, q, rows, support, nil
}

const histogramContractScript = `
import assert from 'node:assert/strict';import fs from 'node:fs';import {createSourceAdapter,importSnapshot,selectRows,hasGuarantee} from '../ui/snapshots.mjs';import {validateSnapshot,validateDisplay,eligibleKinds,histogramIntent} from '../ui/view-model.mjs';
const {snapshot:raw,question,rows,support,seed=0}=JSON.parse(fs.readFileSync(0,'utf8'));let reads=0;globalThis.fetch=()=>{reads++;throw Error('Histogram display cannot read sources')};const source=createSourceAdapter({id:'histogram-'+seed}).adoptSnapshot(raw,{question}),before=JSON.stringify(source);
assert.equal(source.shape,'histogram');assert.equal(source.histogram.binning,'powers-of-two');assert.equal(source.histogram.closure,'left-closed-right-open');assert.deepEqual(source.histogram.binding,question.numeric);assert.deepEqual(source.histogram.support,support);assert.equal(source.rows.length,rows.length);assert.equal(source.rows.length,question.numeric.type==='uint64'?65:128);assert.deepEqual(eligibleKinds(source),['table','histogram']);
let sum=0n;for(const [i,row]of source.rows.entries()){assert.equal(row.key,'bin:'+i);assert.deepEqual(row.values,rows[i]);assert.equal(row.state,row.values.count==='0'?'no_data':'observed');assert.equal(row.cells,undefined);sum+=BigInt(row.values.count)}assert.equal(sum,BigInt(support.valid));assert.equal(BigInt(support.total),BigInt(support.valid)+BigInt(support.missing)+BigInt(support.empty)+BigInt(support.invalid));
const intent=histogramIntent(source,source.rows[0].key,source.rows.at(-1).key);assert.deepEqual(intent,{type:'scope-histogram',datasetId:source.id,revision:source.revision,scopeId:source.scope.id,generation:source.scope.generation,fromKey:'bin:0',toKey:'bin:'+(source.rows.length-1)});assert.throws(()=>histogramIntent(source,source.rows.at(-1).key,source.rows[0].key));assert.throws(()=>histogramIntent(source,'bin:foreign'));
assert(hasGuarantee(source,'bins'));assert(validateDisplay(source,{kind:'histogram',requiresComplete:true}).ok);assert(validateDisplay(source,{kind:'table'}).ok);const held=importSnapshot(source);assert(validateSnapshot(held).ok);assert(validateDisplay(held,{kind:'histogram'}).ok);assert(!validateDisplay(held,{kind:'histogram',requiresComplete:true}).ok);assert(!hasGuarantee(held,'bins'));assert.deepEqual(held.rows,source.rows);assert.deepEqual(held.histogram,source.histogram);
assert.throws(()=>selectRows(source,[source.rows[0].key]),error=>error.code==='projection-shape');assert.throws(()=>selectRows(source,source.rows.map(r=>r.key)),error=>error.code==='projection-shape');const sliced=structuredClone(source);sliced.rows=sliced.rows.slice(1);assert(!validateSnapshot(sliced).ok);assert.throws(()=>importSnapshot(sliced));
assert.equal(JSON.stringify(source),before);assert.equal(reads,0);process.stdout.write('ok');
`

func TestHegelCanonicalHistogramGeometryAndCoverage(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		typ := []string{"int64", "uint64", "decimal64"}[hegel.Draw(tc, hegel.Integers(0, 2))]
		scale := hegel.Draw(tc, hegel.Integers(1, 9))
		seed := hegel.Draw(tc, hegel.Integers(0, 1000))
		mode := hegel.Draw(tc, hegel.Integers(0, 2))
		raw, q, rows, support, err := histogramContractFixture(typ, scale, seed, mode)
		if err != nil {
			tc.Fatal(err)
		}
		if _, err = runSnapshotJS(raw, q, map[string]any{"rows": rows, "support": support, "seed": seed}, histogramContractScript); err != nil {
			tc.Fatal(err)
		}
	}, options(60)...)
}

const histogramMalformedScript = `
import assert from 'node:assert/strict';import fs from 'node:fs';import {validateSnapshot} from '../ui/view-model.mjs';const {snapshot:raw,mode=0}=JSON.parse(fs.readFileSync(0,'utf8')),s=structuredClone(raw);
switch(mode){case 0:s.rows[0].values.from='1';break;case 1:s.rows.at(-1).values.to=s.rows.at(-1).values.from;break;case 2:s.rows[0].values.count='18446744073709551616';break;case 3:s.rows[0].values.count=9007199254740992;break;case 4:s.histogram.support.valid=String(BigInt(s.histogram.support.valid)+1n);break;case 5:s.histogram.support.total=String(BigInt(s.histogram.support.total)+1n);break;case 6:s.rows[0].state=s.rows[0].state==='observed'?'no_data':'observed';break;case 7:s.rows[0].key='bin:999';break;case 8:s.histogram.binding.scale=10;break;case 9:s.histogram.closure='closed';break}
assert(!validateSnapshot(s).ok,JSON.stringify({mode,s}));process.stdout.write('ok');
`

func TestHegelCanonicalHistogramMalformedContracts(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		typ := []string{"int64", "uint64", "decimal64"}[hegel.Draw(tc, hegel.Integers(0, 2))]
		seed := hegel.Draw(tc, hegel.Integers(0, 1000))
		mode := hegel.Draw(tc, hegel.Integers(0, 9))
		raw, q, _, _, err := histogramContractFixture(typ, seed%9+1, seed, 0)
		if err != nil {
			tc.Fatal(err)
		}
		if _, err = runSnapshotJS(raw, q, map[string]any{"mode": mode}, histogramMalformedScript); err != nil {
			tc.Fatal(err)
		}
	}, options(40)...)
}
