package verification

import (
	"encoding/json"
	"github.com/casselc/goo/exploration"
	"hegel.dev/go/hegel"
	"os/exec"
	"strconv"
	"testing"
)

func TestHegelLatencyHistogramPartitions(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		a := hegel.Draw(tc, hegel.Integers(0, 7))
		b := hegel.Draw(tc, hegel.Integers(a, 7))
		script := `const {selection}=require('../ui/latency-histogram.mjs'),{partition}=require('../ui/population-selection.mjs');const [a,b]=process.argv.slice(1);const data={state:'observed',count:'1',groups:[{g0:'[5,10) ms',count:'1'}],question:{signal:'spans',environment:'test',fromMs:1700000000000,toMs:1700000030000,limit:1,offset:100,groupBy:[{kind:'column',key:'duration_bucket'}],where:{any:[{field:{kind:'attribute',key:'owner'},op:'missing'},{field:{kind:'attribute',key:'owner'},op:'eq',value:''}]}}};process.stdout.write(JSON.stringify({base:data.question,...partition(data.question,selection(data,b,a))}));`
		raw, e := exec.Command("node", "-e", script, strconv.Itoa(a), strconv.Itoa(b)).CombinedOutput()
		if e != nil {
			tc.Fatal(e, string(raw))
		}
		var p struct{ Base, Selected, Rest exploration.Query }
		if e = json.Unmarshal(raw, &p); e != nil {
			tc.Fatal(e)
		}
		for _, q := range []exploration.Query{p.Selected, p.Rest} {
			if _, e = exploration.Compile(q); e != nil {
				tc.Fatal(e)
			}
			if q.FromMS != p.Base.FromMS || q.ToMS != p.Base.ToMS || q.Offset != 0 {
				tc.Fatal("scope drift")
			}
		}
		edges := []uint64{0, 1000000, 5000000, 10000000, 50000000, 100000000, 500000000, 1000000000}
		ds := []uint64{0, 9007199254740993, 18446744073709551615}
		for _, v := range edges {
			ds = append(ds, v)
			if v > 0 {
				ds = append(ds, v-1)
			}
		}
		for _, d := range ds {
			for owner := 0; owner < 3; owner++ {
				vals := map[exploration.Field]string{{Kind: "column", Key: "duration_ns"}: strconv.FormatUint(d, 10), {Kind: "column", Key: "timestamp_ns"}: "1700000001000000000"}
				if owner > 0 {
					vals[exploration.Field{Kind: "attribute", Key: "owner"}] = []string{"", "", "outside"}[owner]
				}
				parent := exploration.Match(p.Base.Where, vals)
				s, r := exploration.Match(p.Selected.Where, vals), exploration.Match(p.Rest.Where, vals)
				expected := parent && d >= edges[a] && (b == 7 || d < edges[b+1])
				if s != expected || s && r || (s || r) != parent {
					tc.Fatal("partition mismatch", a, b, d, owner)
				}
			}
		}
	}, options(40)...)
}
