package verification

import (
	"encoding/json"
	"github.com/casselc/goo/exploration"
	"hegel.dev/go/hegel"
	"os/exec"
	"reflect"
	"strconv"
	"testing"
)

// Removing an entire conjunct must be monotone, including missing/empty values
// and nested ORs. The actual Go evaluator and SQL compiler validate the JS intent.
func TestHegelPopulationBroadening(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		seed := hegel.Draw(tc, hegel.Integers(0, 1000))
		script := `const {removeFilterQueries}=require('../ui/population-workspace.mjs'),assert=require('node:assert/strict');const n=Number(process.argv[1]),f={kind:'attribute',key:'region'},leaf=(op,value)=>({field:f,op,...(value===undefined?{}:{value})}),clauses=[{any:[leaf('missing'),leaf('eq','')]},{field:{kind:'resource',key:'region'},op:n%2?'missing':'eq',...(n%2?{}:{value:'west'})}];let where=n%3===0?clauses[0]:n%3===1?{all:[clauses[0]]}:{all:clauses};const parents=['spans','logs'].map(signal=>({signal,environment:'test',fromMs:1000,toMs:2000,limit:1,offset:100,where})),saved=JSON.stringify(parents),broadened=removeFilterQueries(parents,'spans',n%3===2?n%2:0);assert.equal(JSON.stringify(parents),saved);assert.deepEqual(broadened[1],parents[1]);process.stdout.write(JSON.stringify({parents,broadened}));`
		raw, err := exec.Command("node", "-e", script, strconv.Itoa(seed)).CombinedOutput()
		if err != nil {
			tc.Fatal(err, string(raw))
		}
		var p struct{ Parents, Broadened []exploration.Query }
		if json.Unmarshal(raw, &p) != nil || len(p.Broadened) != 2 {
			tc.Fatal(string(raw))
		}
		var expected *exploration.Predicate
		if seed%3 == 2 {
			expected = &p.Parents[0].Where.All[1-seed%2]
		}
		if !reflect.DeepEqual(p.Broadened[0].Where, expected) {
			tc.Fatal("incorrect retained condition", seed)
		}
		for _, q := range p.Broadened {
			if _, err := exploration.Compile(q); err != nil {
				tc.Fatal(err)
			}
		}
		if p.Broadened[0].Environment != p.Parents[0].Environment || p.Broadened[0].FromMS != 1000 || p.Broadened[0].ToMS != 2000 || p.Broadened[0].Offset != 0 {
			tc.Fatal("scope drift")
		}
		for a := 0; a < 4; a++ {
			for b := 0; b < 4; b++ {
				r := map[exploration.Field]string{}
				for j, v := range []int{a, b} {
					if v > 0 {
						kind := []string{"attribute", "resource"}[j]
						r[exploration.Field{Kind: kind, Key: "region"}] = []string{"", "", "west", "east"}[v]
					}
				}
				if exploration.Match(p.Parents[0].Where, r) && !exploration.Match(p.Broadened[0].Where, r) {
					tc.Fatal("removal narrowed membership", seed, a, b)
				}
			}
		}
	}, options(40)...)
}
