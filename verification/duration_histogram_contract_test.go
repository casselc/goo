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
	"hegel.dev/go/hegel"
)

var durationContractEdges = []string{"0", "1000000", "5000000", "10000000", "50000000", "100000000", "500000000", "1000000000", "+inf"}
var durationContractLabels = []string{"[0,1) ms", "[1,5) ms", "[5,10) ms", "[10,50) ms", "[50,100) ms", "[100,500) ms", "[500,1000) ms", "[1000,+inf) ms"}

func durationContractFixture(seed, mode int) (*exploration.ViewDataset, exploration.Query, []string, error) {
	q := exploration.Query{Signal: "spans", Environment: "prod", FromMS: 1000, ToMS: 61000, Limit: 100, GroupBy: []exploration.Field{{Kind: "column", Key: "duration_bucket"}}}
	counts := make([]uint64, 8)
	durations := []uint64{0, 1, 999999, 1000000, 1000001, 4999999, 5000000, 9999999, 10000000, 49999999, 50000000, 99999999, 100000000, 499999999, 500000000, 999999999, 1000000000, 18446744073709551615}
	for i := 0; i < seed%7; i++ {
		durations = append(durations, uint64(seed)*1000000)
	}
	if mode == 1 {
		durations = nil
	}
	if mode == 2 {
		durations = []uint64{1}
	}
	if mode == 3 {
		durations = nil
		counts[7] = 18446744073709551615
	}
	for _, n := range durations {
		index := 7
		for i := 1; i < 8; i++ {
			bound, _ := strconv.ParseUint(durationContractEdges[i], 10, 64)
			if n < bound {
				index = i - 1
				break
			}
		}
		counts[index]++
	}
	total := uint64(0)
	texts := make([]string, 8)
	groups := []string{}
	for i, n := range counts {
		total += n
		texts[i] = strconv.FormatUint(n, 10)
		if n > 0 {
			r, _ := json.Marshal(map[string]any{"g0": durationContractLabels[i], "count": texts[i]})
			groups = append(groups, string(r))
		}
	}
	switch mode {
	case 4:
		total = 1
		groups = []string{`{"g0":"unrecognized","count":"1"}`}
	case 5:
		total = 1
		groups = []string{`{"g0":null,"count":"1"}`}
	case 6:
		groups = []string{`{"g0":"[0,1) ms","count":"18446744073709551615"}`, `{"g0":"[1,5) ms","count":"1"}`}
	case 7:
		total += counts[0]
		groups = append(groups, groups[0])
	}
	reads := 0
	source := exploration.Source{Endpoint: "http://source.invalid", Client: &http.Client{Transport: snapshotTransport(func(req *http.Request) (*http.Response, error) {
		reads++
		body, _ := io.ReadAll(req.Body)
		sql := string(body)
		text := ""
		switch {
		case strings.Contains(sql, " AS timestamp_ns"):
		case strings.Contains(sql, "SELECT DISTINCT"):
		case strings.Contains(sql, "GROUP BY"):
			text = strings.Join(groups, "\n")
		default:
			text = `{"count":"` + strconv.FormatUint(total, 10) + `"}`
		}
		return &http.Response{StatusCode: 200, Header: http.Header{}, Body: io.NopCloser(strings.NewReader(text))}, nil
	})}}
	result, err := source.Fetch(context.Background(), q)
	if err != nil {
		return nil, q, nil, err
	}
	if reads != 4 || result.HistogramDataset == nil {
		return nil, q, nil, fmt.Errorf("missing bounded same-group duration snapshot (%d reads)", reads)
	}
	return result.HistogramDataset, q, texts, nil
}

const durationHistogramScript = `
import assert from 'node:assert/strict';import fs from 'node:fs';import {createSourceAdapter,importSnapshot,hasGuarantee,selectRows} from '../ui/snapshots.mjs';import {validateSnapshot,validateDisplay,histogramIntent} from '../ui/view-model.mjs';import {durationRangeFacts} from '../ui/histogram-values.mjs';
const {snapshot:raw,question,counts,mutation=-1}=JSON.parse(fs.readFileSync(0,'utf8')),edges=['0','1000000','5000000','10000000','50000000','100000000','500000000','1000000000','+inf'];
if(mutation>=0){const s=structuredClone(raw);switch(mutation){case 0:delete s.histogram.kind;break;case 1:s.histogram.kind='numeric';break;case 2:s.histogram.support={total:'0',valid:'0',missing:'0',empty:'0',invalid:'0'};break;case 3:s.rows[0].values.from='+inf';break;case 4:s.rows[0].values.to='+inf';break;case 5:s.rows.at(-1).values.to='18446744073709551616';break;case 6:s.rows[0].values.from='1/2';break;case 7:s.histogram.total=String(BigInt(s.histogram.total)+1n);break;case 8:s.rows[0].values.count=9007199254740992;break;case 9:s.histogram.binding=question.numeric||{};break}assert(!validateSnapshot(s).ok);assert.throws(()=>importSnapshot(s));process.stdout.write('ok');}
else{let reads=0;globalThis.fetch=()=>{reads++;throw Error('No display IO')};const s=createSourceAdapter({id:'duration'}).adoptSnapshot(raw,{question}),before=JSON.stringify(s);assert.equal(s.histogram.kind,'span-duration');assert.equal(s.histogram.binning,'goo-duration-v1');assert.equal(s.histogram.unit,'ns');assert.equal(s.histogram.binding,undefined);assert.equal(s.histogram.support,undefined);assert.equal(s.rows.length,8);assert(hasGuarantee(s,'bins'));let total=0n;for(const [i,r]of s.rows.entries()){assert.equal(r.key,'bin:'+i);assert.deepEqual(r.values,{from:edges[i],to:edges[i+1],count:counts[i]});assert.equal(r.state,counts[i]==='0'?'no_data':'observed');total+=BigInt(counts[i])}assert.equal(s.histogram.total,String(total));assert(validateDisplay(s,{kind:'histogram',requiresComplete:true}).ok);const held=importSnapshot(s);assert.deepEqual(held.rows,s.rows);assert(!hasGuarantee(held,'bins'));assert(!validateDisplay(held,{kind:'histogram',requiresComplete:true}).ok);assert.throws(()=>selectRows(s,['bin:0']));assert(Object.isFrozen(s.rows[0].values));
for(let from=0;from<8;from++)for(let to=from;to<8;to++){const f=durationRangeFacts(s,'bin:'+from,'bin:'+to),count=counts.slice(from,to+1).reduce((n,v)=>n+BigInt(v),0n);assert.equal(f.fromNS,edges[from]);assert.equal(f.toNS,edges[to+1]);assert.equal(f.count,String(count));assert.equal(f.total,String(total));assert.equal(f.fromBin,from);assert.equal(f.toBin,to+1);assert.equal(f.maxMS,to===7?'':String(BigInt(edges[to+1])/1000000n));assert.equal(histogramIntent(s,'bin:'+from,'bin:'+to).type,'scope-histogram')};assert.equal(JSON.stringify(s),before);assert.equal(reads,0);process.stdout.write('ok');}
`

func TestHegelCanonicalDurationHistogramGeometry(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		seed := hegel.Draw(tc, hegel.Integers(0, 1000))
		mode := hegel.Draw(tc, hegel.Integers(0, 3))
		s, q, counts, err := durationContractFixture(seed, mode)
		if err != nil {
			tc.Fatal(err)
		}
		if _, err = runSnapshotJS(s, q, map[string]any{"counts": counts}, durationHistogramScript); err != nil {
			tc.Fatal(err)
		}
	}, options(60)...)
}
func TestHegelCanonicalDurationHistogramMalformed(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		seed := hegel.Draw(tc, hegel.Integers(0, 1000))
		mutation := hegel.Draw(tc, hegel.Integers(0, 9))
		s, q, counts, err := durationContractFixture(seed, 0)
		if err != nil {
			tc.Fatal(err)
		}
		if _, err = runSnapshotJS(s, q, map[string]any{"counts": counts, "mutation": mutation}, durationHistogramScript); err != nil {
			tc.Fatal(err)
		}
	}, options(20)...)
}
func TestHegelTaggedNumericHistogramNonRegression(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		seed := hegel.Draw(tc, hegel.Integers(0, 1000))
		typ := []string{"int64", "uint64", "decimal64"}[seed%3]
		s, q, rows, support, err := histogramContractFixture(typ, seed%9+1, seed, seed%3)
		if err != nil {
			tc.Fatal(err)
		}
		if s.Histogram.Kind != "numeric" {
			tc.Fatal("missing explicit numeric family")
		}
		if _, err = runSnapshotJS(s, q, map[string]any{"rows": rows, "support": support}, histogramContractScript); err != nil {
			tc.Fatal(err)
		}
	}, options(20)...)
}

func TestCanonicalDurationHistogramRefusesUnknownSourceGroups(t *testing.T) {
	for mode := 4; mode <= 7; mode++ {
		if _, _, _, err := durationContractFixture(0, mode); err == nil {
			t.Fatalf("unrecognized/missing/overflow/duplicate source grouping admitted mode %d", mode)
		}
	}
}
