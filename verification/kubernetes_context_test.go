package verification

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/casselc/goo/exploration"
	"hegel.dev/go/hegel"
	"os/exec"
	"testing"
)

func workloadScopeOK(q exploration.Query, r map[string]string, kind string, from int64) bool {
	return workloadScope(q, r, kind, "workload-property", from, from+60000)
}
func workloadScope(q exploration.Query, r map[string]string, kind, environment string, from, to int64) bool {
	if q.Where == nil || q.Where.Field != nil || q.Where.Numeric != nil || len(q.Where.Any) != 0 || q.Where.Op != "" || q.Where.Value != "" || q.Numeric != nil || q.NumericHistogram != nil || q.ReportedEdge != nil || len(q.GroupBy) > 0 || q.BucketMS != 0 || q.Summary != "" || q.SpanStatus || q.GroupTimeSeries {
		return false
	}
	if q.Environment != environment || q.FromMS != from || q.ToMS != to || q.Limit != 100 || q.Offset != 0 || len(q.Where.All) != 6 {
		return false
	}
	keys := map[string]bool{}
	for _, k := range []string{"deployment.environment.name", "k8s.cluster.uid", "k8s.cluster.name", "k8s.namespace.name", "k8s." + kind + ".uid", "k8s." + kind + ".name"} {
		keys[k] = true
	}
	for _, p := range q.Where.All {
		if p.Numeric != nil || p.Field == nil || p.Field.Kind != "resource" || !keys[p.Field.Key] || len(p.All) > 0 || len(p.Any) > 0 {
			return false
		}
		delete(keys, p.Field.Key)
		v, ok := r[p.Field.Key]
		if ok {
			if p.Op != "eq" || p.Value != v {
				return false
			}
		} else if p.Op != "missing" {
			return false
		}
	}
	return len(keys) == 0
}
func TestHegelKubernetesWorkloadContextScope(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		seed := hegel.Draw(tc, hegel.Integers(0, 242))
		from := int64(1800000000000 + seed*1000)
		kinds := []string{"deployment", "replicaset", "statefulset", "daemonset", "job", "cronjob"}
		resources := []map[string]string{}
		for _, kind := range kinds {
			r := map[string]string{"deployment.environment.name": "workload-property", "service.name": "excluded", "service.instance.id": "excluded", "k8s.pod.uid": "excluded", "container.id": "excluded"}
			other := "job"
			if kind == "job" {
				other = "deployment"
			}
			r["k8s."+other+".name"] = "unrelated-workload"
			code := seed
			for _, key := range []string{"k8s.cluster.uid", "k8s.cluster.name", "k8s.namespace.name", "k8s." + kind + ".uid", "k8s." + kind + ".name"} {
				state := code % 3
				code /= 3
				if state == 1 {
					r[key] = ""
				} else if state == 2 {
					r[key] = fmt.Sprintf("literal-%s-%d", key, seed)
				}
			}
			if r["k8s."+kind+".uid"] == "" && r["k8s."+kind+".name"] == "" {
				r["k8s."+kind+".name"] = "reported-name"
			}
			resources = append(resources, r)
		}
		input, _ := json.Marshal(map[string]any{"resources": resources, "kinds": kinds, "question": map[string]any{"environment": "workload-property", "fromMs": from, "toMs": from + 60000}})
		script := `const fs=require('fs'),a=require('assert/strict'),h=require('../ui/kubernetes-context'),p=JSON.parse(fs.readFileSync(0,'utf8')),before=JSON.stringify(p);global.fetch=()=>{throw Error('unexpected read')};const out=p.resources.map((resource,i)=>{const row={resource},w=h.workloads(resource).find(w=>w.kind===p.kinds[i]);a.ok(w?.actionable);a.equal(h.filtered([row,{resource:{...resource,'k8s.pod.uid':'another','service.name':'another'}}],{workloadKey:w.key}).length,2);a.equal(h.filtered([row],{workloadKey:h.NOT_REPORTED}).length,0);a.throws(()=>h.workloadQuestion({...p.question,environment:'other'},row,w.key,'logs'));for(const field of ['fromNS','toNS'])a.throws(()=>h.workloadQuestion({...p.question,[field]:'1'},row,w.key,'logs'));return ['logs','spans'].map(signal=>h.workloadQuestion(p.question,row,w.key,signal))});a.equal(JSON.stringify(p),before);console.log(JSON.stringify(out));`
		cmd := exec.Command("node", "-e", script)
		cmd.Stdin = bytes.NewReader(input)
		raw, e := cmd.CombinedOutput()
		if e != nil {
			tc.Errorf("helper: %v %s", e, raw)
			return
		}
		var out [][]exploration.Query
		if e = json.Unmarshal(raw, &out); e != nil || len(out) != 6 {
			tc.Errorf("decode: %v", e)
			return
		}
		for n, pair := range out {
			if len(pair) != 2 {
				tc.Errorf("signals omitted")
				return
			}
			for j, q := range pair {
				if q.Signal != []string{"logs", "spans"}[j] || !workloadScopeOK(q, resources[n], kinds[n], from) {
					tc.Errorf("scope drift")
					return
				}
				if _, e := exploration.Compile(q); e != nil {
					tc.Errorf("compile: %v", e)
					return
				}
				bad := q
				bad.Environment = "other"
				if workloadScopeOK(bad, resources[n], kinds[n], from) {
					tc.Errorf("wrong scope accepted")
					return
				}
				bad = q
				bad.Where.All = append([]exploration.Predicate{}, q.Where.All...)
				for i, p := range bad.Where.All {
					if p.Field.Key == "k8s."+kinds[n]+".uid" {
						if p.Op == "missing" {
							bad.Where.All[i].Op = "eq"
							bad.Where.All[i].Value = ""
						} else {
							bad.Where.All[i].Op = "missing"
							bad.Where.All[i].Value = ""
						}
					}
				}
				if workloadScopeOK(bad, resources[n], kinds[n], from) {
					tc.Errorf("UID missing/empty mutation accepted")
					return
				}
			}
		}
	}, options(12)...)
}
