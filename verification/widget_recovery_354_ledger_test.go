package verification

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/casselc/observability/quintgo/qobs"
	"github.com/casselc/observability/quintgo/validate"
)

type recoveryObject354 = map[string]any

func recoveryObjectValue354(v any) recoveryObject354 { o, _ := v.(map[string]any); return o }
func recoveryEqual354(a, b any) bool                 { return continuationJSON352(a) == continuationJSON352(b) }
func recoveryRequire354(ok bool, label string) error {
	if !ok {
		return fmt.Errorf("%s", label)
	}
	return nil
}
func recoveryAdmission354(l recoveryObject354, c recoveryObject354) error {
	plan := recoveryObjectValue354(c["plan"])
	scope := recoveryObjectValue354(c["scope"])
	if !recoveryEqual354(plan["scope"], scope) || plan["service"] != recoveryObjectValue354(l["fixture"])["service"] {
		return fmt.Errorf("plan scope/service")
	}
	var descriptor any
	endpoint := ""
	widgets, _ := plan["widgets"].([]any)
	for _, raw := range widgets {
		w := recoveryObjectValue354(raw)
		if w["key"] == c["failedKey"] {
			if w["state"] != "supported" {
				return fmt.Errorf("unsupported retry")
			}
			descriptor = w["query"]
			endpoint = "/api/record-aggregates"
			if descriptor == nil {
				descriptor = w["measurement"]
				endpoint = "/api/measurements"
			}
		}
	}
	failed := recoveryObjectValue354(c["failedRequest"])
	retry := recoveryObjectValue354(c["retry"])
	if descriptor == nil || failed["path"] != endpoint || retry["path"] != endpoint || !recoveryEqual354(failed["body"], descriptor) || !recoveryEqual354(retry["request"], descriptor) || failed["status"] != float64(502) || retry["status"] != float64(200) {
		return fmt.Errorf("whole retry descriptor/status")
	}
	refusal := recoveryObjectValue354(recoveryObjectValue354(failed["response"])["refusal"])
	if refusal["kind"] != "unavailable" {
		return fmt.Errorf("source refusal")
	}
	response := recoveryObjectValue354(retry["response"])
	if !recoveryEqual354(response["question"], descriptor) {
		return fmt.Errorf("response owner")
	}
	if endpoint == "/api/record-aggregates" {
		duration := recoveryObjectValue354(response["duration"])
		if response["count"] != "3" || duration["count"] != "3" || duration["sumNs"] != "30000000" {
			return fmt.Errorf("independent fixture span population")
		}
	} else {
		gauges := recoveryObjectValue354(response["gauges"])
		series, _ := gauges["series"].([]any)
		if len(series) != 1 {
			return fmt.Errorf("native population")
		}
		samples, _ := recoveryObjectValue354(series[0])["samples"].([]any)
		found := false
		for _, s := range samples {
			if recoveryObjectValue354(s)["value"] == float64(42) {
				found = true
			}
		}
		if !found {
			return fmt.Errorf("native raw fixture value")
		}
	}
	if !recoveryEqual354(c["healthyBefore"], c["healthyAfter"]) {
		return fmt.Errorf("healthy DOM identity/read clocks changed")
	}
	healthy, _ := c["healthyBefore"].([]any)
	if len(healthy) < 2 {
		return fmt.Errorf("missing healthy witnesses")
	}
	for _, h := range healthy {
		o := recoveryObjectValue354(h)
		if o["state"] != "observed" || o["slotCount"] != float64(1) || o["evidence"] == nil || o["domID"] == nil || o["preID"] == nil {
			return fmt.Errorf("healthy witness")
		}
	}
	before := recoveryObjectValue354(c["failedWidgetBefore"])
	after := recoveryObjectValue354(c["failedWidgetAfter"])
	if before["state"] != "unavailable" || after["state"] != "observed" || before["domID"] != after["domID"] || after["slotCount"] != float64(1) || before["preID"] == after["preID"] {
		return fmt.Errorf("failed slot replacement")
	}
	evidence := recoveryObjectValue354(after["evidence"])
	if !recoveryEqual354(evidence["question"], descriptor) {
		return fmt.Errorf("receipt descriptor")
	}
	facts, _ := c["facts"].([]any)
	names := []string{"inspect-refusal", "explicit-retry", "retired-after-success"}
	if len(facts) != len(names) {
		return fmt.Errorf("missing gesture")
	}
	calls, _ := l["calls"].([]any)
	for i, raw := range facts {
		f := recoveryObjectValue354(raw)
		b := recoveryObjectValue354(f["before"])
		a := recoveryObjectValue354(f["after"])
		if f["name"] != names[i] || a["plans"] != b["plans"] {
			return fmt.Errorf("gesture/plan budget")
		}
		start, ok1 := b["calls"].(float64)
		end, ok2 := a["calls"].(float64)
		if !ok1 || !ok2 || start < 0 || end < start || end > float64(len(calls)) {
			return fmt.Errorf("call boundaries")
		}
		source := 0
		for _, rawCall := range calls[int(start):int(end)] {
			call := recoveryObjectValue354(rawCall)
			if call["method"] == "POST" {
				switch call["path"] {
				case "/api/record-aggregates", "/api/measurements", "/api/discovery", "/api/explore", "/api/inventory", "/api/metric-discovery":
					source++
				}
			}
		}
		want := 0
		if i == 1 {
			want = 1
		}
		if source != want || a["source"].(float64)-b["source"].(float64) != float64(want) {
			return fmt.Errorf("explicit-only source budget")
		}
		if i != 1 && a["sql"] != b["sql"] {
			return fmt.Errorf("passive SQL")
		}
	}
	if !recoveryEqual354(c["callbackBefore"], c["callbackAfter"]) {
		return fmt.Errorf("retired success callback")
	}
	return nil
}

func TestQuintWidgetRecoveryObserved354(t *testing.T) {
	path := os.Getenv("GOO_WIDGET_RECOVERY_354_REPORT")
	if path == "" {
		t.Skip("explicit native observed report required")
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var ledger recoveryObject354
	if err = json.Unmarshal(raw, &ledger); err != nil {
		t.Fatal(err)
	}
	cases, _ := ledger["cases"].([]any)
	errors, _ := ledger["errors"].([]any)
	if ledger["version"] != "goo.widgetRecovery354/1" || len(cases) != 2 || len(errors) != 0 {
		t.Fatal("incomplete native campaign")
	}
	check := func(label string, actions []string, admission error, want bool) {
		t.Helper()
		if (admission == nil) != want {
			t.Fatalf("%s admission: %v", label, admission)
		}
		sink := &qobs.MemorySink{}
		rec := qobs.NewRecorder(sink, qobs.WithActor(label))
		for i, a := range actions {
			rec.Record(context.Background(), a, nil, "obs.coherent", admission == nil || i != len(actions)-1)
		}
		checker, e := validate.NewChecker("model/widget_recovery_354.yaml", validate.Options{Quint: quintPath(t), Backend: "typescript", Dir: t.TempDir()})
		if e != nil {
			t.Fatal(e)
		}
		r, e := checker.Check(context.Background(), sink.Steps())
		if e != nil {
			t.Fatal(e)
		}
		artifact("widget-recovery354-"+label+".txt", []byte(r.Summary()))
		if r.OK() != want {
			t.Fatalf("%s: %s", label, r.Summary())
		}
	}
	actions := []string{"start", "source_failed", "retry", "succeed", "refused", "finish"}
	for _, c := range cases {
		o := recoveryObjectValue354(c)
		check(o["key"].(string), actions, recoveryAdmission354(ledger, o), true)
	}
	late := recoveryObjectValue354(ledger["late"])
	lateOK := late["status"] == float64(200) && recoveryEqual354(late["question"], late["request"]) && recoveryEqual354(recoveryObjectValue354(late["response"])["question"], late["question"]) && recoveryEqual354(late["pendingBefore"], late["pendingAfter"]) && recoveryEqual354(late["beforeRelease"], late["afterRelease"]) && recoveryEqual354(late["beforeRelease"], late["afterOldCallback"]) && recoveryEqual354(recoveryObjectValue354(late["abort"])["request"], late["question"])
	check("retired-owner", []string{"start", "source_failed", "retry", "refused", "retire", "late", "refused", "finish"}, recoveryRequire354(lateOK, "late response/callback withdrawal"), true)
	mutations := []func(recoveryObject354){func(c recoveryObject354) {
		recoveryObjectValue354(c["retry"])["request"] = map[string]any{"signal": "logs"}
	}, func(c recoveryObject354) { c["healthyAfter"] = []any{} }, func(c recoveryObject354) { recoveryObjectValue354(c["failedWidgetAfter"])["slotCount"] = float64(2) }, func(c recoveryObject354) { recoveryObjectValue354(c["failedWidgetAfter"])["state"] = "unavailable" }, func(c recoveryObject354) { c["facts"] = []any{} }, func(c recoveryObject354) { c["callbackAfter"] = nil }}
	for i, mutate := range mutations {
		var copy recoveryObject354
		json.Unmarshal(raw, &copy)
		c := recoveryObjectValue354(copy["cases"].([]any)[0])
		mutate(c)
		check(fmt.Sprintf("poison-%d", i), actions, recoveryAdmission354(copy, c), false)
	}
	// Protocol controls reject a duplicate acquisition while busy and a successful
	// response after retirement, independently of the admission expectation bit.
	for i, bad := range [][]string{{"start", "source_failed", "retry", "retry", "succeed", "finish"}, {"start", "source_failed", "retry", "retire", "succeed", "finish"}} {
		sink := &qobs.MemorySink{}
		rec := qobs.NewRecorder(sink)
		for _, a := range bad {
			rec.Record(context.Background(), a, nil, "obs.coherent", true)
		}
		checker, e := validate.NewChecker("model/widget_recovery_354.yaml", validate.Options{Quint: quintPath(t), Backend: "typescript", Dir: t.TempDir()})
		if e != nil {
			t.Fatal(e)
		}
		r, e := checker.Check(context.Background(), sink.Steps())
		if e != nil || r.OK() {
			t.Fatalf("protocol negative %d: %v", i, e)
		}
	}
}
