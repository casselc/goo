package verification

import (
	"hegel.dev/go/hegel"
	"os/exec"
	"strconv"
	"testing"
)

func TestHegelInstrumentCapabilityHints(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		kind := []string{"gauge", "counter", "histogram"}[hegel.Draw(tc, hegel.Integers(0, 2))]
		unit := []string{"s", "ms", "By", ""}[hegel.Draw(tc, hegel.Integers(0, 3))]
		support := hegel.Draw(tc, hegel.Integers(0, 1))
		tempo := hegel.Draw(tc, hegel.Integers(0, 3))
		script := `const a=require('node:assert/strict'),{describe}=require('../ui/instrument-capabilities');const [kind,unit,support,tempo]=process.argv.slice(1),i={metric:'http.server.request.duration',kind,unit,scope:'ordinary',temporality:Number(tempo),supported:support==='1',definition:{source:{unit}},dimensions:[]},before=structuredClone(i),c=describe(i);a.equal(c.shapeMatches,kind==='histogram'&&unit==='s');a.equal(c.eligible,support==='1');a.equal(c.verification,'not-run');a.deepEqual(i,before);a.equal(describe({...i,definition:null}).eligible,false);a.equal(describe({...i,metric:'other.metric'}).profile,'generic');a(c.hints.some(x=>x.includes('no HTTP error-rate')));if(kind!=='histogram')a(!c.hints.some(x=>x.includes('bucket ranges')));`
		if out, e := exec.Command("node", "-e", script, kind, unit, strconv.Itoa(support), strconv.Itoa(tempo)).CombinedOutput(); e != nil {
			tc.Fatal(e, string(out))
		}
	}, options(40)...)
}
