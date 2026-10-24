package verification

import (
	"encoding/json"
	"github.com/casselc/goo/exploration"
	"hegel.dev/go/hegel"
	"os/exec"
	"strconv"
	"testing"
)

// Cross-runtime conservation: JS validates exact grouped wire counts, while the
// independent native cohort checker verifies the missing/empty partition.
func TestHegelFilterValuePartitions(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		seed := hegel.Draw(tc, hegel.Integers(1, 1000000))
		script := `const {query,result}=require('../ui/filter-values');const n=9007199254740993n+BigInt(process.argv[1]),q=query({signal:'logs',environment:'prod',fromMs:1000,toMs:61000,where:{field:{kind:'column',key:'service'},op:'eq',value:'api'}},{kind:'attribute',key:'release'});const data={question:q,state:'observed',count:String(n+2n),basis:'property',capturedAt:'2026-10-24T00:00:00Z',groups:[{g0:null,count:'1'},{g0:'',count:'1'},{g0:'common',count:String(n)}]};const model=result(q,data);process.stdout.write(JSON.stringify({q,model}));`
		raw, e := exec.Command("node", "-e", script, strconv.Itoa(seed)).CombinedOutput()
		if e != nil {
			tc.Fatal(e, string(raw))
		}
		var wire struct {
			Q     exploration.Query
			Model struct {
				Total string
				Items []struct {
					Value *string
					Count string
				}
			}
		}
		if e = json.Unmarshal(raw, &wire); e != nil {
			tc.Fatal(e)
		}
		cohort := exploration.Cohort{Question: wire.Q, State: "observed", Basis: "property", Count: wire.Model.Total, Groups: []exploration.ValueCount{}}
		for _, item := range wire.Model.Items {
			cohort.Groups = append(cohort.Groups, exploration.ValueCount{Value: item.Value, Count: item.Count})
		}
		same, e := exploration.CompareCohorts(exploration.ComparisonRequest{Baseline: cohort, Selection: cohort})
		if e != nil || len(same.Rows) != 3 {
			tc.Fatal("partition changed", e)
		}
		for _, row := range same.Rows {
			if row.DifferencePP != "0.00" {
				tc.Fatal("exact counts changed")
			}
		}
		cohort.Groups[0].Count = "0"
		if _, e = exploration.CompareCohorts(exploration.ComparisonRequest{Baseline: cohort, Selection: cohort}); e == nil {
			tc.Fatal("broken counted partition accepted")
		}
	}, options(80)...)
}
