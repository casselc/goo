package verification

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/casselc/goo/measurements"
	"os/exec"
	"strconv"
	"testing"
	"time"

	"github.com/casselc/goo/exploration"
	"hegel.dev/go/hegel"
)

// chScope checks the entire AST, not merely presence of required predicates.
func chScope(q exploration.Query, resource map[string]string, lo, hi int64, signal string) bool {
	if q.Signal != signal || q.Environment != "history-property" || q.FromMS != lo/1000000 || q.ToMS != (hi+999999)/1000000 || q.Limit != 100 || q.Offset != 0 || len(q.GroupBy) != 0 || q.Numeric != nil || q.NumericHistogram != nil || q.ReportedEdge != nil || q.Summary != "" || q.BucketMS != 0 || q.SpanStatus || q.GroupTimeSeries || q.Where == nil {
		return false
	}
	p := q.Where
	if p.Field != nil || p.Numeric != nil || len(p.Any) != 0 || p.Op != "" || p.Value != "" || len(p.All) != 8 {
		return false
	}
	want := map[string]bool{}
	for _, k := range []string{"container.id", "k8s.pod.uid", "k8s.container.name", "k8s.cluster.uid", "k8s.cluster.name", "k8s.namespace.name"} {
		want[k] = true
	}
	timeSeen := map[string]bool{}
	for _, leaf := range p.All {
		if leaf.Field == nil || leaf.Numeric != nil || len(leaf.All) != 0 || len(leaf.Any) != 0 {
			return false
		}
		if leaf.Field.Kind == "column" {
			if leaf.Field.Key != "timestamp_ns" || timeSeen[leaf.Op] {
				return false
			}
			timeSeen[leaf.Op] = true
			if leaf.Op == "gte" && leaf.Value == strconv.FormatInt(lo, 10) {
				continue
			}
			if leaf.Op == "lte" && leaf.Value == strconv.FormatInt(hi-1, 10) {
				continue
			}
			return false
		}
		k := leaf.Field.Key
		if leaf.Field.Kind != "resource" || !want[k] {
			return false
		}
		delete(want, k)
		value, present := resource[k]
		if present {
			if leaf.Op != "eq" || leaf.Value != value {
				return false
			}
		} else if leaf.Op != "missing" || leaf.Value != "" {
			return false
		}
	}
	return len(want) == 0 && timeSeen["gte"] && timeSeen["lte"]
}

func TestHegelContainerHistoryExactContext(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		seed := hegel.Draw(tc, hegel.Integers(0, 242))
		offset := hegel.Draw(tc, hegel.Integers(1, 999999))
		lo := int64(1800000000000000000) + int64(offset)
		hi := lo + 1000001 + int64(seed)
		// Mandatory IDs are always reported; each optional qualifier independently
		// traverses missing, empty and present states. Unrelated fields must be absent.
		resource := map[string]string{"container.id": "container", "k8s.pod.uid": "pod", "k8s.container.name": "worker", "service.name": "must-not-filter", "service.instance.id": "must-not-filter", "metric.only": "must-not-filter"}
		n := seed
		for _, k := range []string{"k8s.cluster.uid", "k8s.cluster.name", "k8s.namespace.name"} {
			s := n % 3
			n /= 3
			if s == 1 {
				resource[k] = ""
			}
			if s == 2 {
				resource[k] = "literal-" + k
			}
		}
		input, _ := json.Marshal(map[string]any{"parent": map[string]any{"environment": "history-property", "fromMs": 1800000000000, "toMs": 1800000060000, "resource": resource, "resourceSchemaURL": ""}, "range": map[string]string{"fromNS": fmt.Sprint(lo), "toNS": fmt.Sprint(hi)}})
		script := `const h=require('../ui/container-resource-history'),a=require('assert/strict'),fs=require('fs'),p=JSON.parse(fs.readFileSync(0,'utf8')),before=JSON.stringify(p);global.fetch=()=>{throw Error('unexpected source read')};const out=['logs','spans'].map(s=>h.contextQuestion(p.parent,p.range,s));for(const r of [{fromNS:p.range.toNS,toNS:p.range.fromNS},{fromNS:p.range.fromNS,toNS:p.range.fromNS},{fromNS:'0',toNS:p.range.toNS}])a.throws(()=>h.contextQuestion(p.parent,r,'logs'));a.equal(JSON.stringify(p),before);console.log(JSON.stringify(out))`
		cmd := exec.Command("node", "-e", script)
		cmd.Stdin = bytes.NewReader(input)
		raw, e := cmd.CombinedOutput()
		if e != nil {
			tc.Errorf("Node helper: %v %s", e, raw)
			return
		}
		var qs []exploration.Query
		if e = json.Unmarshal(raw, &qs); e != nil || len(qs) != 2 {
			tc.Errorf("decode: %v %s", e, raw)
			return
		}
		for i, q := range qs {
			signal := []string{"logs", "spans"}[i]
			if !chScope(q, resource, lo, hi, signal) {
				tc.Errorf("scope drift: %s", raw)
				return
			}
			if _, e = exploration.Compile(q); e != nil {
				tc.Errorf("Go compiler: %v", e)
				return
			}
			// Independent physical timestamps show why enclosing milliseconds alone
			// would incorrectly admit both endpoint controls.
			for _, at := range []int64{lo - 1, lo, lo + 1, hi - 1, hi, hi + 1} {
				admitted := true
				for _, p := range q.Where.All {
					if p.Field.Kind == "column" {
						v, _ := strconv.ParseInt(p.Value, 10, 64)
						if p.Op == "gte" {
							admitted = admitted && at >= v
						} else {
							admitted = admitted && at <= v
						}
					}
				}
				if admitted != (at >= lo && at < hi) {
					tc.Errorf("boundary membership %d", at)
					return
				}
			}
			bad := q
			bad.Environment = "wrong"
			if chScope(bad, resource, lo, hi, signal) {
				tc.Errorf("environment mutant admitted")
				return
			}
			b, _ := json.Marshal(q)
			var changed exploration.Query
			_ = json.Unmarshal(b, &changed)
			for j, p := range changed.Where.All {
				if p.Field.Kind == "resource" && p.Field.Key == "k8s.cluster.uid" {
					changed.Where.All[j].Op = "eq"
					changed.Where.All[j].Value = "wrong"
				}
			}
			if chScope(changed, resource, lo, hi, signal) {
				tc.Errorf("qualifier mutant admitted")
				return
			}
		}
	}, options(8)...)
}

// The expected timestamp groups below come from raw reports, not gauge samples
// or either composer. Evaluate constructs the held evidence under test.
func TestHegelContainerHistoryRawProjection(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		seed := hegel.Draw(tc, hegel.Integers(1, 1000))
		resource := map[string]string{"deployment.environment.name": "prod", "container.id": "c", "k8s.pod.uid": "p", "k8s.container.name": "worker"}
		parent := pnQuestion("p")
		parent.Resource = resource
		schema := "reported-schema"
		parent.ResourceSchemaURL = &schema
		for _, family := range []string{"memory", "cpu"} {
			q := cmQuestion("working-set", resource)
			if family == "cpu" {
				q = cpQuestion("usage", resource)
			}
			q.Definition.MaxGapMS = 5000
			valid := strconv.Itoa(seed)
			if family == "cpu" {
				valid = "0." + strconv.Itoa(seed)
			}
			// Intentionally shuffled, with duplicate equal reports, conflicting reports,
			// no-recorded-value flags, a domain-invalid value and a fresh true zero.
			ats := []int64{159000000000, 110000000000, 130000000000, 120000000000, 110000000000, 140000000000, 130000000000, 155000000000, 154999999999}
			vals := []string{"0", valid, "1", "2", valid, "-0.5", "3", "4", "5"}
			flags := []int{0, 0, 0, 1, 0, 0, 0, 0, 0}
			rows := []json.RawMessage{}
			for i, at := range ats {
				r := hrRow(q, vals[i])
				r["end_ns"] = fmt.Sprint(at)
				r["flags"] = flags[i]
				r["temporality"] = 0
				r["monotonic"] = false
				b, _ := json.Marshal(r)
				rows = append(rows, b)
			}
			result, e := measurements.Evaluate(q, rows, time.Unix(200, 0), "", nil)
			if e != nil {
				tc.Errorf("held result: %v", e)
				return
			}
			in := measurements.ContainerResourceHistoryRequest{Parent: parent}
			if family == "memory" {
				in.Memory = &measurements.ContainerMemoryHistoryResults{WorkingSet: result}
			} else {
				in.CPU = &measurements.ContainerCPUHistoryResults{Usage: result}
			}
			before, _ := json.Marshal(in)
			got, e := measurements.ComposeContainerResourceHistory(in)
			if e != nil {
				tc.Errorf("compose: %v", e)
				return
			}
			if len(got.Series) != 3 {
				tc.Errorf("roles omitted")
				return
			}
			s := got.Series[0]
			expectedTimes := []int64{110000000000, 120000000000, 130000000000, 140000000000, 154999999999, 155000000000, 159000000000}
			states := []string{"stale", "unknown", "unknown", "invalid", "stale", "observed", "observed"}
			expectedValue := valid
			if family == "cpu" {
				_, expectedValue, _ = cpDecimal(valid)
			}
			values := []string{expectedValue, "", "", "", "5", "4", "0"}
			if s.Basis != result.Basis || s.ReadAt != result.ReadAt || s.MaxGapNS != "5000000000" || len(s.Points) != 7 {
				tc.Errorf("provenance/points drift")
				return
			}
			for j, p := range s.Points {
				indices := []int{}
				for k, at := range ats {
					if at == expectedTimes[j] {
						indices = append(indices, k)
					}
				}
				x, _ := json.Marshal(indices)
				y, _ := json.Marshal(p.RawRowIndices)
				quality := states[j]
				if quality == "invalid" {
					quality = "stale"
				}
				if p.AtNS != fmt.Sprint(expectedTimes[j]) || p.State != states[j] || p.ObservationState != quality || p.Value != values[j] || !bytes.Equal(x, y) {
					tc.Errorf("%s point%d: %+v", family, j, p)
					return
				}
			}
			if got.Series[1].State != "unavailable" || len(got.Series[1].Points) != 0 {
				tc.Errorf("missing role fabricated")
				return
			}
			after, _ := json.Marshal(in)
			if !bytes.Equal(before, after) {
				tc.Errorf("held input mutated")
				return
			}
			got.Parent.Resource["container.id"] = "changed"
			got.Series[0].Question.Definition.Source.ExactStream.Resource["container.id"] = "changed"
			after, _ = json.Marshal(in)
			if !bytes.Equal(before, after) {
				tc.Errorf("output aliases input")
				return
			}
		}
	}, options(8)...)
}
