package verification

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/measurements"
	"hegel.dev/go/hegel"
)

var captureFamilies387 = []struct{ kind, arm string }{{"explore", "explore"}, {"inventory", "inventory"}, {"comparison", "comparison"}, {"measurement", "measurement"}, {"explanation", "explanation"}, {"trace-search", "traceSearch"}}

type envelopeCase387 struct {
	Page            json.RawMessage `json:"page"`
	Full, Reference bool
	Label           string
}

// Canonical typed hashing precedes optional raw-null insertion. Under the old
// pointer decoder those nulls disappear without changing the digest: rejection
// must therefore come from presence admission, not an integrity mismatch.
func captureWrite387(dir string, c exploration.SavedCapture, rawMutation func(map[string]json.RawMessage)) (string, []byte, error) {
	c.ID = ""
	wire, e := json.Marshal(c)
	if e != nil {
		return "", nil, e
	}
	sum := sha256.Sum256(wire)
	c.ID = hex.EncodeToString(sum[:])
	wire, e = json.Marshal(c)
	if e != nil {
		return "", nil, e
	}
	if rawMutation != nil {
		var fields map[string]json.RawMessage
		json.Unmarshal(wire, &fields)
		rawMutation(fields)
		wire, e = json.Marshal(fields)
		if e != nil {
			return "", nil, e
		}
	}
	return c.ID, wire, os.WriteFile(filepath.Join(dir, c.ID+".json"), wire, 0600)
}
func captureBases387(t *testing.T) []exploration.SavedCapture {
	t.Helper()
	clock := "2026-11-01T00:00:01Z"
	basis := strings.Repeat("b", 64)
	var record exploration.SavedCapture
	wire, e := os.ReadFile("../testdata/storage/exploration-capture.json")
	if e != nil || json.Unmarshal(wire, &record) != nil {
		t.Fatal("record fixture", e)
	}
	record.Explore.TraceExamples = nil
	original := record.Explore.Rows[0]
	record.Explore.Rows = nil
	for i := range 101 {
		var fields map[string]json.RawMessage
		json.Unmarshal(original, &fields)
		fields["SpanId"] = json.RawMessage(strconv.Quote(fmt.Sprintf("%016x", i+1)))
		fields["inventory"] = json.RawMessage("null")
		fields["measurement"] = json.RawMessage("null")
		r, _ := json.Marshal(fields)
		record.Explore.Rows = append(record.Explore.Rows, r)
	}
	record.Explore.Count = "101"
	record.Explore.HasMore = false
	record.Explore.Groups = []json.RawMessage{json.RawMessage(`{"g0":"worker","count":"101"}`)}
	// Frozen historical source plans are data, not an instruction to recompile.
	record.Explore.Plan.Count += " /* historical adapter retained */"
	iq := exploration.InventoryQuestion{Environment: "prod", FromMS: 1000, ToMS: 61000}
	ip, e := exploration.CompileInventory(iq)
	if e != nil {
		t.Fatal(e)
	}
	inventory := exploration.SavedCapture{CaptureReceipt: exploration.CaptureReceipt{Version: 1, Title: "Generated empty inventory", CapturedAt: clock, Question: exploration.CaptureQuestion{Kind: "inventory", Inventory: &iq}}, Inventory: &exploration.InventoryResult{Question: iq, CapturedAt: clock, Basis: basis, State: "no_data", Count: "0", Rows: []exploration.InventoryRow{}, BucketLabels: []string{}, Warnings: []string{}, Plan: ip}}
	q := exploration.Query{Signal: "spans", Environment: "prod", FromMS: 1000, ToMS: 61000, Limit: 100, GroupBy: []exploration.Field{{Kind: "resource", Key: "team"}}}
	other := q
	other.FromMS, other.ToMS = 61000, 121000
	empty := ""
	pair := exploration.ComparisonRequest{Baseline: exploration.Cohort{Question: q, CapturedAt: clock, Basis: basis, Count: "3", State: "observed", Groups: []exploration.ValueCount{{Value: nil, Count: "2"}, {Value: &empty, Count: "1"}}}, Selection: exploration.Cohort{Question: other, CapturedAt: clock, Basis: strings.Repeat("c", 64), Count: "3", State: "observed", Groups: []exploration.ValueCount{{Value: nil, Count: "2"}, {Value: &empty, Count: "1"}}}}
	compared, e := exploration.CompareCohorts(pair)
	if e != nil {
		t.Fatal(e)
	}
	cq := exploration.ComparisonQuestion{Baseline: q, Selection: other}
	comparison := exploration.SavedCapture{CaptureReceipt: exploration.CaptureReceipt{Version: 1, Title: "Generated comparison nullable groups", CapturedAt: clock, Question: exploration.CaptureQuestion{Kind: "comparison", Comparison: &cq}}, Comparison: &exploration.SavedComparison{BaselineID: strings.Repeat("a", 64), SelectionID: strings.Repeat("b", 64), Request: pair, Result: compared}}
	mq, stream := nrcInputs(nil, "gauge")
	rows := []json.RawMessage{}
	for i := range 2 {
		raw, _ := json.Marshal(map[string]any{"resource_schema_url": "", "resource": stream.Resource, "scope": stream.Scope, "version": stream.Version, "schema_url": stream.SchemaURL, "name": stream.Name, "unit": stream.Unit, "attributes": stream.Attributes, "start_ns": "0", "end_ns": strconv.FormatInt(mq.FromMS*1000000+int64(i+1)*1000000000, 10), "value": "0", "flags": i})
		rows = append(rows, raw)
	}
	evaluated, e := measurements.Evaluate(mq, rows, time.UnixMilli(mq.ToMS+1))
	if e != nil {
		t.Fatal(e)
	}
	measurement := exploration.SavedCapture{CaptureReceipt: exploration.CaptureReceipt{Version: 1, Title: "Generated independent gauge observations", CapturedAt: time.UnixMilli(mq.ToMS + 2).UTC().Format(time.RFC3339Nano), Question: exploration.CaptureQuestion{Kind: "measurement", Measurement: &mq}}, Measurement: evaluated}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		if strings.Contains(string(raw), "GROUP BY") {
			fmt.Fprintln(w, `{"g0":null,"count":"2"}`)
			fmt.Fprintln(w, `{"g0":"","count":"1"}`)
		} else {
			fmt.Fprintln(w, `{"count":"3"}`)
		}
	}))
	defer server.Close()
	q.GroupBy = nil
	other.GroupBy = nil
	er := exploration.ExplanationRequest{Version: exploration.ExplanationVersion, Baseline: q, Selection: other, Fields: []exploration.Field{{Kind: "resource", Key: "team"}}}
	explained, e := (exploration.Source{Endpoint: server.URL}).Explain(context.Background(), er)
	if e != nil {
		t.Fatal(e)
	}
	explained.ResultToken = ""
	explanation := exploration.SavedCapture{CaptureReceipt: exploration.CaptureReceipt{Version: 1, Title: "Generated nullable explanation", CapturedAt: explained.ReadAt, Question: exploration.CaptureQuestion{Kind: "explanation", Explanation: &er}}, Explanation: explained}
	search := generatedSearchCapture383(387, 3, 2, false)
	bases := []exploration.SavedCapture{record, inventory, comparison, measurement, explanation, search}
	for i := range bases {
		bases[i].ID = ""
		wire, _ := json.Marshal(bases[i])
		sum := sha256.Sum256(wire)
		bases[i].ID = hex.EncodeToString(sum[:])
	}
	return bases
}
func cloneCapture387(t *testing.T, c exploration.SavedCapture) exploration.SavedCapture {
	t.Helper()
	wire, _ := json.Marshal(c)
	var out exploration.SavedCapture
	if e := json.Unmarshal(wire, &out); e != nil {
		t.Fatal(e)
	}
	return out
}
func envelopeRaw387(v any) json.RawMessage { b, _ := json.Marshal(v); return b }
func clearCaptureArm387(c *exploration.SavedCapture, question bool) {
	switch c.Question.Kind {
	case "explore":
		if question {
			c.Question.Explore = nil
		} else {
			c.Explore = nil
		}
	case "inventory":
		if question {
			c.Question.Inventory = nil
		} else {
			c.Inventory = nil
		}
	case "comparison":
		if question {
			c.Question.Comparison = nil
		} else {
			c.Comparison = nil
		}
	case "measurement":
		if question {
			c.Question.Measurement = nil
		} else {
			c.Measurement = nil
		}
	case "explanation":
		if question {
			c.Question.Explanation = nil
		} else {
			c.Explanation = nil
		}
	case "trace-search":
		if question {
			c.Question.TraceSearch = nil
		} else {
			c.TraceSearch = nil
		}
	}
}
func driftCaptureResult387(c *exploration.SavedCapture) {
	switch c.Question.Kind {
	case "explore":
		c.Explore.Question.FromMS++
	case "inventory":
		c.Inventory.Question.FromMS++
	case "comparison":
		c.Comparison.Request.Baseline.Question.FromMS++
	case "measurement":
		c.Measurement.Question.FromMS++
	case "explanation":
		c.Explanation.Request.Baseline.FromMS++
	case "trace-search":
		c.TraceSearch.Request.FromMS++
	}
}
func pageDrift387(raw json.RawMessage, path []string) json.RawMessage {
	var object map[string]json.RawMessage
	json.Unmarshal(raw, &object)
	if len(path) == 1 {
		var n int64
		json.Unmarshal(object[path[0]], &n)
		object[path[0]] = envelopeRaw387(n + 1)
	} else {
		object[path[0]] = pageDrift387(object[path[0]], path[1:])
	}
	return envelopeRaw387(object)
}
func TestHegelCaptureEnvelope387(t *testing.T) {
	bases := captureBases387(t)
	cases := []envelopeCase387{}
	seedBank := []map[string]any{}
	hegel.Test(t, func(tc *hegel.T) {
		seed := hegel.Draw(tc, hegel.Integers(1, 100000))
		for index := range 6 {
			c := cloneCapture387(t, bases[index])
			c.Title = fmt.Sprintf("Capture %d 雪", seed)
			dir := t.TempDir()
			id, wire, e := captureWrite387(dir, c, nil)
			if e != nil {
				tc.Fatal(e)
			}
			store := exploration.CaptureStore{Dir: dir}
			loaded, e := store.Load(id)
			if e != nil {
				tc.Fatal("valid family refused", c.Question.Kind, e)
			}
			page, e := loaded.Page("")
			if e != nil {
				tc.Fatal(e)
			}
			before, _ := json.Marshal(loaded)
			if len(seedBank) < 6 {
				seedBank = append(seedBank, map[string]any{"family": c.Question.Kind, "saved": json.RawMessage(wire), "page": page})
			}
			var roundTrip exploration.CapturePage
			if err := json.Unmarshal(envelopeRaw387(page), &roundTrip); err != nil || !sameTransport384(page, roundTrip) {
				tc.Fatal("full page output did not round trip", err)
			}
			cases = append(cases, envelopeCase387{Page: envelopeRaw387(page), Full: true, Label: c.Question.Kind + "-valid"})
			if index == 0 {
				second, e := loaded.Page(page.NextCursor)
				if e != nil || second.Explore.Question.Offset != 100 || len(second.Explore.Rows) != 1 || second.Explore.Count != "101" {
					tc.Fatal("stable page changed full count", e)
				}
				cases = append(cases, envelopeCase387{Page: envelopeRaw387(second), Full: true, Label: "historical-second-page"})
			}
			if index == 0 {
				empty := cloneCapture387(t, c)
				empty.Title = "Generated empty retained records"
				empty.Explore.Rows = nil
				empty.Explore.Groups = nil
				empty.Explore.Count = "0"
				empty.Explore.HasMore = false
				empty.Explore.State = "no_data"
				emptyDir := t.TempDir()
				emptyID, emptyWire, err := captureWrite387(emptyDir, empty, nil)
				if err != nil {
					tc.Fatal(err)
				}
				kept, err := (exploration.CaptureStore{Dir: emptyDir}).Load(emptyID)
				if err != nil {
					tc.Fatal("valid empty record capture refused", err)
				}
				ep, err := kept.Page("")
				if err != nil {
					tc.Fatal(err)
				}
				var round exploration.CapturePage
				if err = json.Unmarshal(envelopeRaw387(ep), &round); err != nil {
					tc.Fatal(err)
				}
				disk, err := os.ReadFile(filepath.Join(emptyDir, emptyID+".json"))
				if err != nil || !bytes.Equal(emptyWire, disk) {
					tc.Fatal("empty capture rewritten", err)
				}
				cases = append(cases, envelopeCase387{Page: envelopeRaw387(ep), Full: true, Label: "empty-null-row-capture"})
			}
			after, _ := json.Marshal(loaded)
			disk, _ := os.ReadFile(filepath.Join(dir, id+".json"))
			if !bytes.Equal(before, after) || !bytes.Equal(wire, disk) {
				tc.Fatal("admission/page rewrote immutable capture")
			}
			// A truthful metadata reference cannot masquerade as a full evidence page.
			reference := map[string]any{"receipt": page.Receipt, "definitionRevision": "", "readAt": c.CapturedAt, "summary": map[string]any{"kind": c.Question.Kind}}
			cases = append(cases, envelopeCase387{Page: envelopeRaw387(reference), Reference: true, Label: "metadata-reference"})
			reference["readAt"] = ""
			cases = append(cases, envelopeCase387{Page: envelopeRaw387(reference), Reference: true, Label: "metadata-reference-empty-read-clock"})
			selected := captureFamilies387[index].arm
			for _, foreign := range captureFamilies387 {
				if foreign.arm == selected {
					continue
				}
				for _, location := range []string{"result", "question"} {
					mutation := func(fields map[string]json.RawMessage) {
						if location == "result" {
							fields[foreign.arm] = json.RawMessage("null")
						} else {
							var question map[string]json.RawMessage
							json.Unmarshal(fields["question"], &question)
							question[foreign.arm] = json.RawMessage("null")
							fields["question"] = envelopeRaw387(question)
						}
					}
					mutantDir := t.TempDir()
					badID, _, err := captureWrite387(mutantDir, c, mutation)
					if err != nil {
						tc.Fatal(err)
					}
					if _, err = (exploration.CaptureStore{Dir: mutantDir}).Load(badID); err == nil {
						tc.Fatal("rehashed foreign null admitted", location, foreign.arm)
					}
					var bad map[string]json.RawMessage
					json.Unmarshal(envelopeRaw387(page), &bad)
					if location == "result" {
						bad[foreign.arm] = json.RawMessage("null")
					} else {
						var receipt map[string]json.RawMessage
						json.Unmarshal(bad["receipt"], &receipt)
						var question map[string]json.RawMessage
						json.Unmarshal(receipt["question"], &question)
						question[foreign.arm] = json.RawMessage("null")
						receipt["question"] = envelopeRaw387(question)
						bad["receipt"] = envelopeRaw387(receipt)
					}
					raw := envelopeRaw387(bad)
					var decoded exploration.CapturePage
					if json.Unmarshal(raw, &decoded) == nil {
						tc.Fatal("Go page foreign null admitted")
					}
					cases = append(cases, envelopeCase387{Page: raw, Label: location + "-foreign-null-" + foreign.arm})
				}
				badRef := map[string]any{}
				refWire := envelopeRaw387(reference)
				json.Unmarshal(refWire, &badRef)
				badRef[foreign.arm] = nil
				cases = append(cases, envelopeCase387{Page: envelopeRaw387(badRef), Label: "reference-foreign-null"})
			}

			for _, question := range []bool{false, true} {
				for _, explicitNull := range []bool{false, true} {
					mutant := cloneCapture387(t, c)
					clearCaptureArm387(&mutant, question)
					var rawNull func(map[string]json.RawMessage)
					if explicitNull {
						rawNull = func(fields map[string]json.RawMessage) {
							if question {
								var q map[string]json.RawMessage
								json.Unmarshal(fields["question"], &q)
								q[selected] = json.RawMessage("null")
								fields["question"] = envelopeRaw387(q)
							} else {
								fields[selected] = json.RawMessage("null")
							}
						}
					}
					mutantDir := t.TempDir()
					badID, _, err := captureWrite387(mutantDir, mutant, rawNull)
					if err != nil {
						tc.Fatal(err)
					}
					if _, err = (exploration.CaptureStore{Dir: mutantDir}).Load(badID); err == nil {
						tc.Fatal("rehashed selected missing/null admitted")
					}
				}
			}
			for _, mutate := range []func(*exploration.SavedCapture){func(c *exploration.SavedCapture) { c.Question.Kind = "unknown" }, driftCaptureResult387} {
				mutant := cloneCapture387(t, c)
				mutate(&mutant)
				mutantDir := t.TempDir()
				badID, _, err := captureWrite387(mutantDir, mutant, nil)
				if err != nil {
					tc.Fatal(err)
				}
				if _, err = (exploration.CaptureStore{Dir: mutantDir}).Load(badID); err == nil {
					tc.Fatal("rehashed discriminator/scope mismatch admitted")
				}
			}
			driftPath := []string{selected, "question", "fromMs"}
			switch index {
			case 2:
				driftPath = []string{selected, "request", "baseline", "question", "fromMs"}
			case 4:
				driftPath = []string{selected, "request", "baseline", "fromMs"}
			case 5:
				driftPath = []string{selected, "request", "fromMs"}
			}
			drifted := pageDrift387(envelopeRaw387(page), driftPath)
			var badScope exploration.CapturePage
			if json.Unmarshal(drifted, &badScope) == nil {
				tc.Fatal("Go page receipt/result scope drift admitted")
			}
			cases = append(cases, envelopeCase387{Page: drifted, Label: "receipt-result-scope-drift"})
			for _, mutate := range []func(*exploration.SavedCapture){func(c *exploration.SavedCapture) { c.CapturedAt = "2026-02-30T00:00:00Z" }, func(c *exploration.SavedCapture) { c.Version = 2 }} {
				mutant := cloneCapture387(t, c)
				mutate(&mutant)
				mutantDir := t.TempDir()
				badID, _, err := captureWrite387(mutantDir, mutant, nil)
				if err != nil {
					tc.Fatal(err)
				}
				if _, err = (exploration.CaptureStore{Dir: mutantDir}).Load(badID); err == nil {
					tc.Fatal("rehashed invalid receipt admitted")
				}
			}
			for _, mutate := range []func(map[string]json.RawMessage){
				func(p map[string]json.RawMessage) {
					var receipt map[string]json.RawMessage
					json.Unmarshal(p["receipt"], &receipt)
					receipt["capturedAt"] = json.RawMessage(`"2026-02-30T00:00:00Z"`)
					p["receipt"] = envelopeRaw387(receipt)
				},
				func(p map[string]json.RawMessage) { delete(p, "recordsLimited") },
				func(p map[string]json.RawMessage) { p["recordsLimited"] = json.RawMessage("null") },
				func(p map[string]json.RawMessage) { p["retainedRecords"] = json.RawMessage("null") },
				func(p map[string]json.RawMessage) { p["nextCursor"] = json.RawMessage(`""`) },
			} {
				var bad map[string]json.RawMessage
				json.Unmarshal(envelopeRaw387(page), &bad)
				mutate(bad)
				raw := envelopeRaw387(bad)
				var decoded exploration.CapturePage
				if json.Unmarshal(raw, &decoded) == nil {
					tc.Fatal("Go invalid page delivery/clock admitted")
				}
				cases = append(cases, envelopeCase387{Page: raw, Label: "clock-page-delivery"})
			}
			for _, key := range []string{selected, "retainedRecords", "recordsLimited", "nextCursor"} {
				var ref map[string]json.RawMessage
				json.Unmarshal(envelopeRaw387(reference), &ref)
				ref[key] = json.RawMessage("null")
				cases = append(cases, envelopeCase387{Page: envelopeRaw387(ref), Label: "reference-has-data-authority"})
			}
			if index == 2 {
				var bad map[string]json.RawMessage
				json.Unmarshal(envelopeRaw387(page), &bad)
				var pair map[string]json.RawMessage
				json.Unmarshal(bad[selected], &pair)
				var request map[string]json.RawMessage
				json.Unmarshal(pair["request"], &request)
				request["distribution"] = json.RawMessage("true")
				pair["request"] = envelopeRaw387(request)
				bad[selected] = envelopeRaw387(pair)
				raw := envelopeRaw387(bad)
				var decoded exploration.CapturePage
				if json.Unmarshal(raw, &decoded) == nil {
					tc.Fatal("Go comparison intent drift admitted")
				}
				cases = append(cases, envelopeCase387{Page: raw, Label: "comparison-distribution-drift"})
			}
			if index == 5 {
				var bad exploration.CapturePage
				if err := json.Unmarshal(envelopeRaw387(page), &bad); err != nil {
					tc.Fatal(err)
				}
				bad.Receipt.Question.TraceSearch.Limit = 1
				bad.TraceSearch.Request.Limit = 1
				raw := envelopeRaw387(bad)
				var decoded exploration.CapturePage
				if json.Unmarshal(raw, &decoded) == nil {
					tc.Fatal("Go page representative witnesses exceed request limit")
				}
				cases = append(cases, envelopeCase387{Page: raw, Label: "trace-search-retained-over-limit"})
			}
			// Null/missing selected arms and a discriminator swap are independent mutations.
			for _, mutation := range []func(map[string]json.RawMessage){func(p map[string]json.RawMessage) { p[selected] = json.RawMessage("null") }, func(p map[string]json.RawMessage) { delete(p, selected) }, func(p map[string]json.RawMessage) {
				var r map[string]json.RawMessage
				json.Unmarshal(p["receipt"], &r)
				var q map[string]json.RawMessage
				json.Unmarshal(r["question"], &q)
				q[selected] = json.RawMessage("null")
				r["question"] = envelopeRaw387(q)
				p["receipt"] = envelopeRaw387(r)
			}, func(p map[string]json.RawMessage) {
				var r map[string]json.RawMessage
				json.Unmarshal(p["receipt"], &r)
				var q map[string]json.RawMessage
				json.Unmarshal(r["question"], &q)
				q["kind"] = json.RawMessage(`"unknown"`)
				r["question"] = envelopeRaw387(q)
				p["receipt"] = envelopeRaw387(r)
			}} {
				var bad map[string]json.RawMessage
				json.Unmarshal(envelopeRaw387(page), &bad)
				mutation(bad)
				raw := envelopeRaw387(bad)
				var decoded exploration.CapturePage
				if json.Unmarshal(raw, &decoded) == nil {
					tc.Fatal("Go malformed selected arm admitted")
				}
				cases = append(cases, envelopeCase387{Page: raw, Label: "selected-arm"})
			}
		}
	}, options(10)...)
	if path := os.Getenv("GOO_CAPTURE_387_SEEDS"); path != "" {
		wire, _ := json.Marshal(seedBank)
		if e := os.WriteFile(path, wire, 0600); e != nil {
			t.Fatal(e)
		}
	}
	encoded, _ := json.Marshal(cases)
	cmd := exec.CommandContext(t.Context(), "node", "capture_envelope_387_fixture.mjs")
	cmd.Stdin = bytes.NewReader(encoded)
	out, e := cmd.CombinedOutput()
	if e != nil {
		t.Fatalf("capture boundary cross-language: %v\n%s", e, out)
	}
	var result struct{ Cases int }
	if json.Unmarshal(out, &result) != nil || result.Cases != len(cases) {
		t.Fatalf("incomplete envelope batch %s", out)
	}
}
