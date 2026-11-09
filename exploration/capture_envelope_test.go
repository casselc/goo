package exploration

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCaptureEnvelope387FiniteFamilies(t *testing.T) {
	kinds := []string{"trace-search", "explanation", "measurement", "comparison", "explore", "inventory"}
	for selected, kind := range kinds {
		for bits := 0; bits < 64; bits++ {
			var present [6]bool
			for i := range present {
				present[i] = bits&(1<<i) != 0
			}
			if got := validateCaptureArms(kind, present) == nil; got != (bits == 1<<selected) {
				t.Fatalf("%s arms=%06b accepted=%v", kind, bits, got)
			}
		}
		for foreign, key := range captureArmKeys {
			body := `{"kind":"` + kind + `","` + captureArmKeys[selected] + `":{}}`
			var fields map[string]json.RawMessage
			if err := json.Unmarshal([]byte(body), &fields); err != nil {
				t.Fatal(err)
			}
			if err := captureWireArm(fields, kind); err != nil {
				t.Fatal("shape-only admission", kind, err)
			}
			if foreign == selected {
				continue
			}
			for _, value := range []string{"null", "{}"} {
				bad := strings.TrimSuffix(body, "}") + `,"` + key + `":` + value + `}`
				if json.Unmarshal([]byte(bad), new(CaptureQuestion)) == nil {
					t.Fatal("foreign arm admitted", bad)
				}
			}
		}
		for _, bad := range []string{`{"kind":"` + kind + `"}`, `{"kind":"` + kind + `","` + captureArmKeys[selected] + `":null}`, `{"kind":"` + kind + `","` + captureArmKeys[selected] + `":[]}`} {
			if json.Unmarshal([]byte(bad), new(CaptureQuestion)) == nil {
				t.Fatal("missing/null/nonobject selected arm", bad)
			}
		}
	}
}

func TestCaptureEnvelope387StoredNullPresenceAndPageCoherence(t *testing.T) {
	store := CaptureStore{Dir: t.TempDir()}
	saved, err := store.publish(searchFixture(t))
	if err != nil {
		t.Fatal(err)
	}
	raw, err := json.Marshal(saved)
	if err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"traceSearch", "explanation", "measurement", "comparison", "inventory"} {
		// The unchanged original ID would pass the old typed hash check: null erased
		// to an omitted nil pointer. Admission must reject its actual wire presence.
		bad := strings.TrimSuffix(string(raw), "}") + `,"` + key + `":null}`
		if err = os.WriteFile(filepath.Join(store.Dir, saved.ID+".json"), []byte(bad), 0600); err != nil {
			t.Fatal(err)
		}
		if _, err = store.Load(saved.ID); err == nil {
			t.Fatal("foreign null persisted arm admitted", key)
		}
	}
	page, err := saved.Page("")
	if err != nil {
		t.Fatal(err)
	}
	raw, err = json.Marshal(page)
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(raw, new(CapturePage)); err != nil {
		t.Fatal("valid page rejected", err)
	}
	for _, mutate := range []func(string) string{
		func(s string) string { return strings.TrimSuffix(s, "}") + `,"inventory":null}` },
		func(s string) string { return strings.Replace(s, `"retainedRecords":1`, `"retainedRecords":1001`, 1) },
		func(s string) string { return strings.Replace(s, `"recordsLimited":false`, `"recordsLimited":null`, 1) },
		func(s string) string { return strings.TrimSuffix(s, "}") + `,"nextCursor":"foreign"}` },
		func(s string) string { return strings.Replace(s, `"receipt":{`, `"receipt":{"Version":1,`, 1) },
	} {
		bad := mutate(string(raw))
		if bad == string(raw) {
			continue
		}
		if json.Unmarshal([]byte(bad), new(CapturePage)) == nil {
			t.Fatal("malformed page admitted", bad)
		}
	}
	polluted := *saved
	polluted.Inventory = &InventoryResult{}
	if _, err = polluted.Page(""); err == nil {
		t.Fatal("programmatic mixed page admitted")
	}
	receipt := saved.CaptureReceipt
	saved.Version = 0
	if _, err = saved.Page(""); err == nil {
		t.Fatal("incomplete receipt admitted")
	}
	saved.CaptureReceipt = receipt
}

func TestCaptureEnvelope387TracePageRespectsSelectedLimit(t *testing.T) {
	q := traceSearchTestRequest()
	q.Limit = 2
	source, _ := traceCapture383Source(t, q)
	saved, err := (CaptureStore{Dir: t.TempDir()}).Create(context.Background(), source, "Two retained traces", CaptureQuestion{Kind: "trace-search", TraceSearch: &q}, "")
	if err != nil {
		t.Fatal(err)
	}
	page, err := saved.Page("")
	if err != nil || page.RetainedRecords != 2 || len(page.TraceSearch.Items) != 2 {
		t.Fatal("valid two-record page", page, err)
	}
	raw, err := json.Marshal(page)
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(raw, new(CapturePage)); err != nil {
		t.Fatal("valid page rejected", err)
	}
	// Keep receipt and evidence requests in agreement. The only fault is retaining
	// two records for a selected one-record question, bypassing full leaf validation.
	page.Receipt.Question.TraceSearch.Limit = 1
	page.TraceSearch.Request.Limit = 1
	raw, err = json.Marshal(page)
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(raw, new(CapturePage)); err == nil {
		t.Fatal("trace page exceeded selected request limit")
	}
}
