package verification

import (
	"encoding/json"
	"hegel.dev/go/hegel"
	"os/exec"
	"testing"
)

func TestHegelOTelFieldHints(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		kind := hegel.Draw(tc, hegel.Integers(0, 2))
		key := hegel.Draw(tc, hegel.Integers(0, 8))
		value := hegel.Draw(tc, hegel.Integers(0, 3))
		args, _ := json.Marshal([]int{kind, key, value})
		script := `const a=require('node:assert/strict'),h=require('../ui/field-presentation'),{suggestionData}=require('../ui/attributes'),{normalize,validate}=require('../ui/filter-editor');const[k,i,v]=JSON.parse(process.argv[1]),kind=['attribute','resource','column'][k],key=['http.method','http.request.method','http.route','service.name','db.system','db.system.name','messaging.destination.name','rpc.method','custom.http.route'][i],value=['','GET','503','literal <&> 雪'][v],field={kind,key},p={field,op:'eq',value},before=JSON.stringify(p),metadata=h.hint(field);h.searchText(field);if(metadata){metadata.label='changed';a.notEqual(h.hint(field).label,'changed')}a.equal(JSON.stringify(p),before);a.deepEqual(normalize(p),{field,op:'eq',...(value?{value}:{})});if(kind==='column'){a.throws(()=>validate(p,'spans'));a.equal(metadata,null)}else{validate(p,'spans');const record=kind==='resource'?{ResourceAttributes:{[key]:value}}:{Attributes:{[key]:value}},data=suggestionData([record]);a.deepEqual(data.fields,[{field,values:[value]}]);a.equal(data.fields.length,1)}if(kind==='resource'&&key!=='service.name'||kind==='attribute'&&['service.name','custom.http.route'].includes(key))a.equal(metadata,null);`
		if out, e := exec.Command("node", "-e", script, string(args)).CombinedOutput(); e != nil {
			tc.Fatalf("field hint projection: %s %v", out, e)
		}
	}, options(80)...)
}
