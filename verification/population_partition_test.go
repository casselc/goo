package verification

import (
	"encoding/json"
	"github.com/casselc/goo/exploration"
	"hegel.dev/go/hegel"
	"os/exec"
	"strconv"
	"testing"
)

func TestHegelPopulationPartition(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		seed := hegel.Draw(tc, hegel.Integers(0, 1000))
		script := `const {partitionQueries}=require('../ui/population-workspace'),assert=require('node:assert/strict');const n=Number(process.argv[1]),parents=['spans','logs'].map(signal=>({signal,environment:'test',fromMs:1000,toMs:2000,limit:1,offset:100,where:n%5===0?undefined:{any:[{field:{kind:'attribute',key:'owner'},op:'missing'},{field:{kind:'attribute',key:'owner'},op:'eq',value:'core'}]}})),before=JSON.stringify(parents),values=[null,'','west'],items=[{signal:'spans',field:{kind:'attribute',key:'region'},value:values[n%3],mode:n%2?'include':'exclude'},{signal:'spans',field:{kind:'resource',key:'region'},value:values[Math.floor(n/3)%3],mode:n%4<2?'include':'exclude'},{signal:'logs',field:{kind:'attribute',key:'region'},value:values[Math.floor(n/9)%3],mode:n%8<4?'include':'exclude'}];items.push({signal:'spans',field:{kind:'attribute',key:'region'},value:values[Math.floor(n/27)%3],mode:'include'});const pair=partitionQueries(parents,items);assert.equal(JSON.stringify(parents),before);process.stdout.write(JSON.stringify({parents,...pair}));`
		raw, err := exec.Command("node", "-e", script, strconv.Itoa(seed)).CombinedOutput()
		if err != nil {
			tc.Fatal(err, string(raw))
		}
		var p struct{ Parents, Selected, Rest []exploration.Query }
		if json.Unmarshal(raw, &p) != nil || len(p.Selected) != 2 || len(p.Rest) != 2 {
			tc.Fatal(string(raw))
		}
		for i, parent := range p.Parents {
			for _, q := range []exploration.Query{p.Selected[i], p.Rest[i]} {
				if _, err := exploration.Compile(q); err != nil {
					tc.Fatal(err)
				}
				if q.Environment != parent.Environment || q.FromMS != parent.FromMS || q.ToMS != parent.ToMS || q.Limit != 1 || q.Offset != 0 {
					tc.Fatal("scope changed")
				}
			}
			fields := []exploration.Field{{Kind: "attribute", Key: "owner"}, {Kind: "attribute", Key: "region"}, {Kind: "resource", Key: "region"}}
			values := []string{"", "", "west", "east", "core"}
			for a := 0; a < 5; a++ {
				for b := 0; b < 5; b++ {
					for c := 0; c < 5; c++ {
						record := map[exploration.Field]string{}
						for j, k := range []int{a, b, c} {
							if k > 0 {
								record[fields[j]] = values[k]
							}
						}
						parentMatch := exploration.Match(parent.Where, record)
						s := exploration.Match(p.Selected[i].Where, record)
						r := exploration.Match(p.Rest[i].Where, record)
						if s && r || (s || r) != parentMatch {
							tc.Fatal("not a partition", seed, a, b, c, s, r, parentMatch)
						}
					}
				}
			}
		}
	}, options(60)...)
}
