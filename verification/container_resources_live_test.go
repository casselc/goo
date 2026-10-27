package verification

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/measurements"
	"github.com/casselc/observability/quintgo/qobs"
	"github.com/casselc/observability/quintgo/validate"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"
)

func rsDecode(v any, out any) bool {
	b, e := json.Marshal(v)
	return e == nil && json.Unmarshal(b, out) == nil
}
func rsFamily(snapshot psMap, name string) psMap {
	for _, v := range psArr(snapshot["families"]) {
		f := psObj(v)
		if f["family"] == name {
			return f
		}
	}
	return nil
}
func rsOrigin(v any) (exploration.PodDiscoveryQuestion, bool) {
	var q exploration.PodDiscoveryQuestion
	ok := rsDecode(v, &q)
	return q, ok
}
func rsRaw(f psMap, r *measurements.Result) bool {
	if r == nil || r.Question.Definition.Source.ExactStream == nil {
		return false
	}
	if cmRole(*r.Question.Definition.Source.ExactStream) != "" {
		_, ok := cmRawGauge(f, r)
		return ok
	}
	_, ok := cpRawGauge(f, r)
	return ok
}
func rsPair(f psMap, v any, family string, origin exploration.PodDiscoveryQuestion) bool {
	if family == "memory" {
		var p cmPair
		return rsDecode(v, &p) && cmPairOK(f, p, origin)
	}
	var p cpPair
	return rsDecode(v, &p) && cpPairOK(f, p, origin)
}
func rsNetwork(responses []any, endpoint string, want any) bool {
	for _, v := range responses {
		r := psObj(v)
		if r["endpoint"] == endpoint && r["status"] == float64(200) && aeJSONEqual(r["body"], want) {
			return true
		}
	}
	return false
}
func rsSnapshot(family psMap, response psMap, name string) bool {
	if family == nil || family["completed"] != true || family["pending"] != false {
		return false
	}
	html, ok := family["html"].(string)
	if !ok {
		return false
	}
	signals := psArr(response["signals"])
	rows := psArr(family["rows"])
	if len(rows) != 3 || len(signals) != 3 {
		return false
	}
	for n, v := range signals {
		s := psObj(v)
		row := psObj(rows[n])
		if row["role"] != s["role"] || row["state"] != s["state"] {
			return false
		}
		value, has := s["value"].(string)
		if !has {
			value = "unavailable"
		}
		if !strings.Contains(html, `data-value="`+value+`"`) {
			return false
		}
		if clock, ok := s["readAt"].(string); ok && !strings.Contains(html, clock) {
			return false
		}
		if basis, ok := s["basis"].(string); ok && !strings.Contains(html, basis) {
			return false
		}
		unit := "By (exact)"
		if name == "cpu" {
			unit = "{cpu} (exact exported value)"
		}
		if has && !strings.Contains(html, value+" "+unit) {
			return false
		}
	}
	return true
}
func rsMainOK(d psMap) bool {
	f := psObj(d["facts"])
	initial := psObj(d["initial"])
	origin, ok := rsOrigin(initial["request"])
	var inventory exploration.PodDiscovery
	if !ok || !rsDecode(initial["response"], &inventory) || !pdOracle(pdObservedFacts(f, origin), &inventory) {
		return false
	}
	responses := psArr(d["responses"])
	reads := map[string]pnRead{}
	for _, key := range []string{"opened", "second"} {
		var r pnRead
		if !rsDecode(d[key], &r) {
			return false
		}
		reads[key] = r
	}
	var ar pnRead
	if !rsDecode(psObj(d["aligned"])["opened"], &ar) {
		return false
	}
	reads["aligned"] = ar
	for _, b := range []struct{ read, row string }{{"opened", "primary"}, {"second", "cluster"}} {
		var row exploration.PodResourceObservation
		if !rsDecode(d[b.row], &row) || !reflect.DeepEqual(reads[b.read].DiscoveryRequest.Resource, row.Resource) || !reflect.DeepEqual(reads[b.read].DiscoveryRequest.ResourceSchemaURL, row.ResourceSchemaURL) {
			return false
		}
		found := false
		for _, v := range inventory.Observations {
			if reflect.DeepEqual(v, row) {
				found = true
			}
		}
		if !found {
			return false
		}
	}
	for _, read := range reads {
		if !ppMetadata(f, read) || !rsNetwork(responses, "/api/pod-metric-discovery", read.Discovery) || !rsNetwork(responses, "/api/pod-dashboard/plan", read.Plan) {
			return false
		}
	}
	for _, key := range []string{"opened", "second"} {
		var read pnRead
		if !rsDecode(d[key], &read) || !ppMetadata(f, read) || len(read.Plan.Widgets) != 0 || read.DiscoveryRequest.Environment != origin.Environment || read.DiscoveryRequest.FromMS != origin.FromMS || read.DiscoveryRequest.ToMS != origin.ToMS {
			return false
		}
	}
	pairs := []struct {
		v      any
		family string
	}{{d["memory"], "memory"}, {d["cpu"], "cpu"}, {d["memoryReload"], "memory"}, {psObj(d["aligned"])["memory"], "memory"}, {psObj(d["aligned"])["cpu"], "cpu"}}
	for pairIndex, p := range pairs {
		allowed := []pnRead{reads["opened"], reads["second"]}
		parent := reads["opened"].DiscoveryRequest
		if pairIndex >= 3 {
			allowed = []pnRead{reads["aligned"]}
			parent = reads["aligned"].DiscoveryRequest
		}
		if !aeJSONEqual(psObj(psObj(p.v)["request"])["parent"], parent) {
			return false
		}
		if !rsPair(f, p.v, p.family, origin) {
			return false
		}
		pair := psObj(p.v)
		if !rsNetwork(responses, "/api/container-"+p.family+"/view", pair["response"]) {
			return false
		}
		req := psObj(pair["request"])
		for _, key := range []string{"workingSet", "usage", "request", "limit"} {
			if value, ok := req[key]; ok {
				if !rsNetwork(responses, "/api/measurements", value) {
					return false
				}
				var result measurements.Result
				if !rsDecode(value, &result) {
					return false
				}
				found := false
				for _, read := range allowed {
					for _, candidate := range read.Plan.Candidates {
						if aeJSONEqual(candidate.Question, result.Question) {
							found = true
						}
					}
				}
				if !found {
					return false
				}
			}
		}
	}
	intents := psArr(d["loadIntents"])
	if len(intents) != 10 {
		return false
	}
	families := []string{"memory", "cpu", "memory", "memory", "cpu"}
	for n, p := range pairs {
		intent, done := psObj(intents[2*n]), psObj(intents[2*n+1])
		if intent["family"] != families[n] || done["family"] != families[n] || intent["phase"] != "invoke" || done["phase"] != "complete" || intent["selected"] != float64(3) || !psEq(intent["selection"], done["selection"]) {
			return false
		}
		keys := map[string]bool{}
		for _, v := range psArr(intent["selection"]) {
			key, ok := psObj(v)["key"].(string)
			if !ok || keys[key] {
				return false
			}
			keys[key] = true
		}
		req := psObj(psObj(p.v)["request"])
		for _, key := range []string{"workingSet", "usage", "request", "limit"} {
			if v, ok := req[key]; ok {
				var r measurements.Result
				if !rsDecode(v, &r) || !keys[measurements.Hash(r.Question)] {
					return false
				}
				delete(keys, measurements.Hash(r.Question))
			}
		}
		if len(keys) != 0 {
			return false
		}
	}
	combined, aligned := psObj(d["combined"]), psObj(psObj(d["aligned"])["snapshot"])
	if !rsSnapshot(rsFamily(combined, "memory"), psObj(psObj(d["memoryReload"])["response"]), "memory") || !rsSnapshot(rsFamily(combined, "cpu"), psObj(psObj(d["cpu"])["response"]), "cpu") || rsFamily(combined, "cpu")["html"] != d["heldCPU"] || !rsSnapshot(rsFamily(aligned, "memory"), psObj(psObj(psObj(d["aligned"])["memory"])["response"]), "memory") || !rsSnapshot(rsFamily(aligned, "cpu"), psObj(psObj(psObj(d["aligned"])["cpu"])["response"]), "cpu") {
		return false
	}
	var individual measurements.Result
	if !rsDecode(d["individual"], &individual) || !rsRaw(f, &individual) || !rsNetwork(responses, "/api/measurements", d["individual"]) || !aeJSONEqual(individual.Question, psObj(psObj(psObj(d["cpu"])["request"])["usage"])["question"]) || !psEq(d["saved"], d["reopened"]) || len(psArr(d["blocked"])) != 0 {
		return false
	}
	actions := []string{"inventory", "shared-first-metadata", "shared-second-metadata", "load-memory-only", "load-cpu-retains-memory", "reload-memory-retains-cpu", "native-stage", "retained-native-back", "native-run", "individual-capture", "aligned-both-families", "frozen-reopen"}
	source := []float64{1, 2, 3, 6, 9, 12, 12, 12, 13, 14, 21, 21}
	plans := []float64{0, 1, 2, 2, 2, 2, 3, 3, 4, 4, 5, 5}
	composes := []float64{0, 0, 0, 1, 2, 3, 3, 3, 3, 3, 5, 5}
	events := psArr(d["events"])
	if len(events) != len(actions) {
		return false
	}
	for n, v := range events {
		e := psObj(v)
		if e["action"] != actions[n] || e["source"] != source[n] || e["plans"] != plans[n] || e["compositions"] != composes[n] {
			return false
		}
	}
	counts := map[string]int{}
	familyLoads := map[string]int{}
	for _, v := range psArr(d["requests"]) {
		r := psObj(v)
		endpoint, _ := r["endpoint"].(string)
		if r["effect"] == "source" {
			counts[endpoint]++
		}
		if r["effect"] == "composition" {
			familyLoads[endpoint]++
		}
	}
	return counts["/api/pod-discovery"] == 1 && counts["/api/pod-metric-discovery"] == 3 && counts["/api/measurements"] == 16 && counts["/api/captures"] == 1 && familyLoads["/api/container-memory/view"] == 3 && familyLoads["/api/container-cpu/view"] == 2
}
func rsOwnerOK(d psMap) bool {
	f := psObj(d["facts"])
	cases := psArr(d["cases"])
	if len(cases) != 5 {
		return false
	}
	baseSource, basePlans, baseCompose := float64(0), float64(0), float64(0)
	reasons := []string{"next-source-close", "metadata-auto-close", "memory-ambiguity", "memory-role", "global-close"}
	for n, v := range cases {
		c := psObj(v)
		if c["reason"] != reasons[n] {
			return false
		}
		origin, ok := rsOrigin(psObj(c["initial"])["request"])
		if !ok {
			return false
		}
		held := psArr(c["held"])
		wanted := 1
		if n >= 2 {
			wanted = 2
		}
		if len(held) != wanted {
			return false
		}
		heldFamilies := map[string]psMap{}
		for _, v := range held {
			h := psObj(v)
			var r measurements.Result
			if h["status"] != float64(200) || !rsDecode(h["body"], &r) || !rsRaw(f, &r) || !aeJSONEqual(h["request"], r.Question) || r.Question.FromMS != origin.FromMS || r.Question.ToMS != origin.ToMS {
				return false
			}
			family := "cpu"
			if cmRole(*r.Question.Definition.Source.ExactStream) != "" {
				family = "memory"
			}
			if heldFamilies[family] != nil {
				return false
			}
			heldFamilies[family] = psObj(h["body"])
		}
		if heldFamilies["cpu"] == nil || (wanted == 2 && heldFamilies["memory"] == nil) {
			return false
		}
		extra := float64(0)
		if c["metadata"] != nil {
			extra = 1
			var read pnRead
			if !rsDecode(c["metadata"], &read) || !ppMetadata(f, read) {
				return false
			}
		}
		before, after := psObj(c["before"]), psObj(c["after"])
		cb, ca := psObj(c["countsBefore"]), psObj(c["countsAfter"])
		survives := n != 4
		if cb["source"] != baseSource+2+float64(wanted)+extra || cb["plans"] != basePlans+1+extra || cb["compositions"] != baseCompose || ca["source"] != cb["source"] || ca["plans"] != cb["plans"] {
			return false
		}
		delta := float64(0)
		if survives {
			delta = 1
		}
		if ca["compositions"] != baseCompose+delta {
			return false
		}
		for _, name := range []string{"memory", "cpu"} {
			b, a := rsFamily(before, name), rsFamily(after, name)
			if b == nil || a == nil || b["completed"] != false || b["pending"] != (name == "cpu" && survives) || a["pending"] != false || a["completed"] != (name == "cpu" && survives) {
				return false
			}
		}
		if survives {
			found := false
			for _, v := range psArr(d["responses"]) {
				r := psObj(v)
				if r["endpoint"] != "/api/container-cpu/view" || r["status"] != float64(200) {
					continue
				}
				request := psObj(r["request"])
				if !aeJSONEqual(request["usage"], heldFamilies["cpu"]) {
					continue
				}
				pair := map[string]any{"request": request, "response": r["body"]}
				if !rsPair(f, pair, "cpu", origin) || !rsSnapshot(rsFamily(after, "cpu"), psObj(r["body"]), "cpu") {
					return false
				}
				found = true
			}
			if !found {
				return false
			}
		}
		baseSource, basePlans, baseCompose = ca["source"].(float64), ca["plans"].(float64), ca["compositions"].(float64)
	}
	return baseSource == 20 && basePlans == 7 && baseCompose == 4
}
func rsWithdrawalOK(d psMap) bool {
	cases := psArr(d["cases"])
	if len(cases) != 3 {
		return false
	}
	for n, v := range cases {
		c := psObj(v)
		phase := []string{"metadata", "composition", "tab"}[n]
		if c["phase"] != phase || !psEq(c["before"], c["after"]) {
			return false
		}
		cb := psObj(c["countsBefore"])
		if cb["source"] != []float64{2, 5, 8}[n] || cb["plans"] != []float64{0, 1, 2}[n] || cb["compositions"] != []float64{0, 1, 1}[n] {
			return false
		}
		held := psObj(c["held"])
		if held["status"] != float64(200) {
			return false
		}
		origin, ok := rsOrigin(psObj(c["initial"])["request"])
		if !ok {
			return false
		}
		switch phase {
		case "metadata":
			var q measurements.PodMetricDiscoveryQuestion
			var r measurements.PodMetricDiscovery
			if !rsDecode(held["request"], &q) || !rsDecode(held["body"], &r) || !cmHeldMetadata(psObj(d["facts"]), q, r) {
				return false
			}
		case "composition":
			if !rsPair(psObj(d["facts"]), map[string]any{"request": held["request"], "response": held["body"]}, "cpu", origin) {
				return false
			}
		case "tab":
			var r measurements.Result
			if !rsDecode(held["body"], &r) || !rsRaw(psObj(d["facts"]), &r) {
				return false
			}
		}
		for _, name := range []string{"memory", "cpu"} {
			f := rsFamily(psObj(c["after"]), name)
			if f == nil || f["completed"] != false || f["pending"] != false {
				return false
			}
		}
	}
	counts := map[string]int{}
	for _, v := range psArr(d["requests"]) {
		r := psObj(v)
		if effect, ok := r["effect"].(string); ok {
			counts[effect]++
		}
	}
	return counts["source"] == 8 && counts["plan"] == 2 && counts["composition"] == 1
}
func rsObs(rec *qobs.Recorder, action string, source, plans, compositions float64, cache int, ml, cl, mp, cp bool) {
	rec.Record(context.Background(), action, nil, "obs.source", source, "obs.plans", plans, "obs.compositions", compositions, "obs.cached", cache, "obs.memoryLoaded", ml, "obs.cpuLoaded", cl, "obs.pendingMemory", mp, "obs.pendingCPU", cp)
}
func rsMainTrace(d psMap) *qobs.MemorySink {
	sink := &qobs.MemorySink{}
	r := qobs.NewRecorder(sink)
	actions := []string{"inventory", "discover", "discover", "memory_load", "cpu_load", "memory_load", "native_stage", "back", "native_run", "capture", "aligned", "freeze"}
	cache := []int{0, 1, 2, 2, 2, 2, 2, 2, 2, 2, 1, 0}
	for n, v := range psArr(d["events"]) {
		e := psObj(v)
		rsObs(r, actions[n], e["source"].(float64), e["plans"].(float64), e["compositions"].(float64), cache[n], n >= 3 && n < 11, n >= 4 && n < 11, false, false)
	}
	return sink
}
func rsOwnerTrace(c psMap, baseSource, basePlans, baseCompose float64) *qobs.MemorySink {
	sink := &qobs.MemorySink{}
	r := qobs.NewRecorder(sink)
	rsObs(r, "inventory", 1, 0, 0, 0, false, false, false, false)
	rsObs(r, "discover", 2, 1, 0, 1, false, false, false, false)
	rsObs(r, "start_cpu", 3, 1, 0, 1, false, false, false, true)
	hasMemory := len(psArr(c["held"])) == 2
	if hasMemory {
		rsObs(r, "start_memory", 4, 1, 0, 1, false, false, true, true)
	}
	action := "source_edit"
	cache := 1
	switch c["reason"] {
	case "metadata-auto-close", "memory-ambiguity":
		action = "metadata_changes_memory"
		cache = 2
	case "memory-role":
		action = "edit_memory"
	case "global-close":
		action = "close"
	}
	b := psObj(c["countsBefore"])
	before := psObj(c["before"])
	rsObs(r, action, b["source"].(float64)-baseSource, b["plans"].(float64)-basePlans, b["compositions"].(float64)-baseCompose, cache, rsFamily(before, "memory")["completed"] == true, rsFamily(before, "cpu")["completed"] == true, rsFamily(before, "memory")["pending"] == true, rsFamily(before, "cpu")["pending"] == true)
	after := psObj(c["after"])
	a := psObj(c["countsAfter"])
	action = "finish_cpu"
	if c["reason"] == "global-close" {
		action = "late_cpu"
	}
	rsObs(r, action, a["source"].(float64)-baseSource, a["plans"].(float64)-basePlans, a["compositions"].(float64)-baseCompose, cache, false, rsFamily(after, "cpu")["completed"] == true, false, false)
	if hasMemory {
		rsObs(r, "late_memory", a["source"].(float64)-baseSource, a["plans"].(float64)-basePlans, a["compositions"].(float64)-baseCompose, cache, false, rsFamily(after, "cpu")["completed"] == true, false, false)
	}
	return sink
}
func rsWithdrawTrace(c psMap) *qobs.MemorySink {
	sink := &qobs.MemorySink{}
	r := qobs.NewRecorder(sink)
	rsObs(r, "inventory", 1, 0, 0, 0, false, false, false, false)
	if c["phase"] == "metadata" {
		rsObs(r, "start_metadata", 2, 0, 0, 0, false, false, false, false)
		rsObs(r, "source_edit", 2, 0, 0, 0, false, false, false, false)
		rsObs(r, "late_metadata", 2, 0, 0, 0, false, false, false, false)
		return sink
	}
	rsObs(r, "discover", 2, 1, 0, 1, false, false, false, false)
	rsObs(r, "start_cpu", 3, 1, 0, 1, false, false, false, true)
	comp := float64(0)
	action := "close"
	if c["phase"] == "composition" {
		comp = 1
		rsObs(r, "begin_compose_cpu", 3, 1, 1, 1, false, false, false, true)
		action = "edit_cpu"
	}
	rsObs(r, action, 3, 1, comp, 1, false, false, false, false)
	rsObs(r, "late_cpu", 3, 1, comp, 1, false, false, false, false)
	return sink
}
func rsRead(t *testing.T, key string) (psMap, []byte, string) {
	t.Helper()
	path := os.Getenv(key)
	if path == "" {
		t.Skip("set " + key + " to actual208 evidence")
	}
	raw, e := os.ReadFile(path)
	if e != nil {
		t.Fatal(e)
	}
	var d psMap
	if json.Unmarshal(raw, &d) != nil {
		t.Fatal("ledger decode")
	}
	return d, raw, path
}
func TestContainerResourcesObserved(t *testing.T) {
	d, raw, path := rsRead(t, "GOO_CONTAINER_RESOURCES_LEDGER")
	if !rsMainOK(d) {
		t.Fatal("shared metadata/family raw values/retention/effects oracle")
	}
	before, after := psObj(d["storedBefore"]), psObj(d["storedAfter"])
	a, e := os.ReadFile(filepath.Join(filepath.Dir(path), before["file"].(string)))
	if e != nil {
		t.Fatal(e)
	}
	z, e := os.ReadFile(filepath.Join(filepath.Dir(path), after["file"].(string)))
	sum := sha256.Sum256(a)
	var stored exploration.SavedCapture
	var saved exploration.CapturePage
	if e != nil || string(a) != string(z) || len(a) != int(before["bytes"].(float64)) || hex.EncodeToString(sum[:]) != before["sha256"] || json.Unmarshal(a, &stored) != nil || !rsDecode(d["saved"], &saved) || stored.ID != saved.Receipt.ID || !reflect.DeepEqual(stored.CaptureReceipt, saved.Receipt) || !peLiveCaptureHash(exploration.CapturePage{Receipt: stored.CaptureReceipt, Measurement: stored.Measurement}) || !rsRaw(psObj(d["facts"]), stored.Measurement) || !aeJSONEqual(stored.Measurement.Question, psObj(d["individual"])["question"]) {
		t.Fatal("full immutable individual capture")
	}
	changes := []func(psMap){func(x psMap) { psObj(psArr(x["events"])[2])["source"] = float64(4) }, func(x psMap) { psObj(psArr(x["events"])[4])["compositions"] = float64(3) }, func(x psMap) { psObj(x["memory"])["response"] = psObj(x["cpu"])["response"] }, func(x psMap) { psObj(psArr(psObj(psObj(x["cpu"])["response"])["signals"])[0])["value"] = "32768" }, func(x psMap) {
		psObj(psArr(psObj(psObj(x["memoryReload"])["response"])["signals"])[0])["readAt"] = "wrong"
	}, func(x psMap) { x["heldCPU"] = "replaced by memory" }, func(x psMap) { rsFamily(psObj(x["combined"]), "cpu")["completed"] = false }, func(x psMap) { psObj(psArr(x["loadIntents"])[0])["selected"] = float64(1) }, func(x psMap) { psObj(psArr(psObj(psArr(x["loadIntents"])[0])["selection"])[0])["key"] = "wrong" }, func(x psMap) { psObj(psObj(x["reopened"])["receipt"])["id"] = "wrong" }, func(x psMap) { psObj(psObj(psObj(x["aligned"])["cpu"])["response"])["percentages"] = psMap{} }, func(x psMap) { psObj(psObj(psObj(x["individual"])["question"])["definition"])["environment"] = "wrong" }}
	for n, change := range changes {
		var x psMap
		_ = json.Unmarshal(raw, &x)
		change(x)
		if psEq(x, d) || rsMainOK(x) {
			t.Fatalf("Resources changed main control %d accepted", n)
		}
	}
}
func TestContainerResourcesOwnershipObserved(t *testing.T) {
	d, raw, _ := rsRead(t, "GOO_CONTAINER_RESOURCES_OWNER_LEDGER")
	if !rsOwnerOK(d) {
		t.Fatal("selective family cancellation/sibling continuation")
	}
	for _, key := range []string{"cancel-sibling", "publish-memory", "extra-read", "missing-compose", "bad-raw", "wrong-scope"} {
		var x psMap
		_ = json.Unmarshal(raw, &x)
		c := psObj(psArr(x["cases"])[2])
		switch key {
		case "cancel-sibling":
			rsFamily(psObj(c["before"]), "cpu")["pending"] = false
		case "publish-memory":
			rsFamily(psObj(c["after"]), "memory")["completed"] = true
		case "extra-read":
			psObj(c["countsAfter"])["source"] = float64(999)
		case "missing-compose":
			psObj(c["countsAfter"])["compositions"] = float64(2)
		case "bad-raw":
			psObj(psObj(psArr(c["held"])[0])["body"])["rows"] = []any{}
		case "wrong-scope":
			psObj(psObj(psObj(psArr(c["held"])[0])["request"])["definition"])["environment"] = "wrong"
		}
		if psEq(x, d) || rsOwnerOK(x) {
			t.Fatalf("Resources owner control %s accepted", key)
		}
	}
}
func TestContainerResourcesWithdrawalObserved(t *testing.T) {
	d, raw, _ := rsRead(t, "GOO_CONTAINER_RESOURCES_WITHDRAWAL_LEDGER")
	if !rsWithdrawalOK(d) {
		t.Fatal("held metadata/compose/tab withdrawal")
	}
	for _, key := range []string{"publish", "pending", "raw", "extra-read"} {
		var x psMap
		_ = json.Unmarshal(raw, &x)
		c := psObj(psArr(x["cases"])[1])
		switch key {
		case "publish":
			rsFamily(psObj(c["after"]), "cpu")["completed"] = true
		case "pending":
			rsFamily(psObj(c["after"]), "cpu")["pending"] = true
		case "raw":
			psObj(psObj(psObj(c["held"])["request"])["usage"])["basis"] = "wrong"
		case "extra-read":
			x["requests"] = append(psArr(x["requests"]), psMap{"endpoint": "/api/measurements", "effect": "source"})
		}
		if psEq(x, d) || rsWithdrawalOK(x) {
			t.Fatalf("Resources withdrawal control %s accepted", key)
		}
	}
}
func TestQuintContainerResourcesObserved(t *testing.T) {
	main, _, path := rsRead(t, "GOO_CONTAINER_RESOURCES_LEDGER")
	owner, _, _ := rsRead(t, "GOO_CONTAINER_RESOURCES_OWNER_LEDGER")
	withdraw, _, _ := rsRead(t, "GOO_CONTAINER_RESOURCES_WITHDRAWAL_LEDGER")
	if !rsMainOK(main) || !rsOwnerOK(owner) || !rsWithdrawalOK(withdraw) {
		t.Fatal("raw semantic precondition before finite replay")
	}
	checker, e := validate.NewChecker("model/container_resources.yaml", validate.Options{Quint: quintPath(t), Backend: "typescript", Dir: t.TempDir()})
	if e != nil {
		t.Fatal(e)
	}
	index := 0
	check := func(name string, sink *qobs.MemorySink) bool {
		ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
		defer cancel()
		report, e := checker.Check(ctx, sink.Steps())
		if e != nil {
			t.Fatal(e)
		}
		data, _ := json.MarshalIndent(sink.Steps(), "", "  ")
		_ = os.WriteFile(filepath.Join(filepath.Dir(path), fmt.Sprintf("protocol-%02d-%s.json", index, name)), data, 0644)
		_ = os.WriteFile(filepath.Join(filepath.Dir(path), fmt.Sprintf("protocol-%02d-%s.txt", index, name)), []byte(report.Summary()), 0644)
		index++
		return report.OK()
	}
	if !check("main", rsMainTrace(main)) {
		t.Fatal("actual Resources main protocol")
	}
	bs, bp, bc := float64(0), float64(0), float64(0)
	for n, v := range psArr(owner["cases"]) {
		c := psObj(v)
		if !check(fmt.Sprintf("owner-%d", n), rsOwnerTrace(c, bs, bp, bc)) {
			t.Fatal("actual sibling owner protocol", n)
		}
		a := psObj(c["countsAfter"])
		bs, bp, bc = a["source"].(float64), a["plans"].(float64), a["compositions"].(float64)
	}
	for n, v := range psArr(withdraw["cases"]) {
		if !check(fmt.Sprintf("withdraw-%d", n), rsWithdrawTrace(psObj(v))) {
			t.Fatal("actual withdrawal protocol", n)
		}
	}
	// These model-only bad traces bypass the semantic oracle intentionally.
	bad := rsMainTrace(main)
	bad.Steps()[3].Obs["source"] = float64(7)
	if check("bad-extra-read", bad) {
		t.Fatal("model accepted extra read")
	}
	bad = rsMainTrace(main)
	steps := bad.Steps()
	steps[0].Action = "cpu_load"
	bad = &qobs.MemorySink{}
	for n := range steps {
		bad.Record(context.Background(), &steps[n])
	}
	if bad.Steps()[0].Action != "cpu_load" {
		t.Fatal("bad order mutation did not persist")
	}
	if check("bad-order", bad) {
		t.Fatal("model accepted load before metadata")
	}
	c := psObj(psArr(owner["cases"])[2])
	bad = rsOwnerTrace(c, 7, 3, 2)
	bad.Steps()[4].Obs["pendingCPU"] = false
	if check("bad-sibling-cancel", bad) {
		t.Fatal("model accepted canceled sibling")
	}
	bad = rsOwnerTrace(c, 7, 3, 2)
	bad.Steps()[4].Obs["pendingMemory"] = true
	if check("bad-memory-pending", bad) {
		t.Fatal("model accepted uncleared affected work")
	}
}
