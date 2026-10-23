package verification

import (
	"encoding/json"
	"github.com/casselc/goo/exploration"
	"hegel.dev/go/hegel"
	"os/exec"
	"strconv"
	"testing"
)

func TestHegelTimeDurationPartition(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		a := hegel.Draw(tc, hegel.Integers(0, 2))
		b := hegel.Draw(tc, hegel.Integers(a+1, 4))
		low := hegel.Draw(tc, hegel.Integers(0, 20))
		high := hegel.Draw(tc, hegel.Integers(low+1, 30))
		script := `const {partition}=require('../ui/population-selection'),assert=require('node:assert/strict');const [a,b,low,high]=process.argv.slice(1).map(Number),base={signal:'spans',environment:'test',fromMs:1700000000000,toMs:1700000004000,limit:1,offset:100,groupBy:[{kind:'column',key:'duration_bucket'}],where:{any:[{field:{kind:'attribute',key:'owner'},op:'missing'},{field:{kind:'attribute',key:'owner'},op:'eq',value:''}]}},before=JSON.stringify(base),p=partition(base,{fromMs:base.fromMs+a*1000,toMs:base.fromMs+b*1000,minMS:String(low),maxMS:String(high)});assert.equal(JSON.stringify(base),before);process.stdout.write(JSON.stringify({base,...p}));`
		raw, err := exec.Command("node", "-e", script, strconv.Itoa(a), strconv.Itoa(b), strconv.Itoa(low), strconv.Itoa(high)).CombinedOutput()
		if err != nil {
			tc.Fatal(err, string(raw))
		}
		var p struct{ Base, Selected, Rest exploration.Query }
		if json.Unmarshal(raw, &p) != nil {
			tc.Fatal(string(raw))
		}
		for _, q := range []exploration.Query{p.Selected, p.Rest} {
			if _, err := exploration.Compile(q); err != nil {
				tc.Fatal(err)
			}
			if q.FromMS != p.Base.FromMS || q.ToMS != p.Base.ToMS || q.Environment != p.Base.Environment || q.Offset != 0 || q.Limit != 1 {
				tc.Fatal("scope changed")
			}
		}
		times := []uint64{1700000000000000000 + uint64(a)*1000000000 - 1, 1700000000000000000 + uint64(a)*1000000000, 1700000000000000000 + uint64(b)*1000000000 - 1, 1700000000000000000 + uint64(b)*1000000000}
		durations := []uint64{0, uint64(low) * 1000000, uint64(high)*1000000 - 1, uint64(high) * 1000000, 9007199254740993, 18446744073709551615}
		for _, ts := range times {
			for _, d := range durations {
				for owner := 0; owner < 3; owner++ {
					values := map[exploration.Field]string{{Kind: "column", Key: "timestamp_ns"}: strconv.FormatUint(ts, 10), {Kind: "column", Key: "duration_ns"}: strconv.FormatUint(d, 10)}
					if owner > 0 {
						values[exploration.Field{Kind: "attribute", Key: "owner"}] = []string{"", "", "outside"}[owner]
					}
					parent := exploration.Match(p.Base.Where, values)
					s, r := exploration.Match(p.Selected.Where, values), exploration.Match(p.Rest.Where, values)
					expected := parent && ts >= 1700000000000000000+uint64(a)*1000000000 && ts < 1700000000000000000+uint64(b)*1000000000 && d >= uint64(low)*1000000 && d < uint64(high)*1000000
					if s != expected || s && r || (s || r) != parent {
						tc.Fatal("not exact partition", ts, d, owner, s, r, parent)
					}
				}
			}
		}
	}, options(40)...)
}
