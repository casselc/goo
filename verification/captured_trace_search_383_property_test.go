package verification

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/casselc/goo/exploration"
	"hegel.dev/go/hegel"
)

// Rehash each planted semantic fault: integrity alone must not make a malformed
// union, bounded total or qualified witness appear to be validated.
func writeSearchCapture383(dir string, capture exploration.SavedCapture) (string, error) {
	capture.ID = ""
	raw, err := json.Marshal(capture)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(raw)
	capture.ID = hex.EncodeToString(sum[:])
	raw, err = json.Marshal(capture)
	if err != nil {
		return "", err
	}
	return capture.ID, os.WriteFile(filepath.Join(dir, capture.ID+".json"), raw, 0600)
}
func generatedSearchCapture383(seed, total, limit int, empty bool) exploration.SavedCapture {
	request := tsGeneratedRequest(seed, 4, limit)
	request.Conditions[1] = exploration.Predicate{All: []exploration.Predicate{tsLeaf("column", "status", "eq", "Error"), tsLeaf("column", "service", "eq", "worker")}}
	request.Conditions[3] = tsLeaf("resource", "team", "missing", "")
	if empty {
		request.Conditions[3] = tsLeaf("resource", "team", "eq", "")
	}
	request.Conditions[3] = exploration.Predicate{All: []exploration.Predicate{request.Conditions[3], tsLeaf("resource", "😀.qualified", "eq", "雪"), tsLeaf("resource", "\ue000.qualified", "eq", "✓")}}
	numericValue := "9007199254740993"
	switch seed % 3 {
	case 0:
		request.Conditions[2].Numeric.Type = "uint64"
	case 1:
		request.Conditions[2].Value = "-9007199254740993"
		numericValue = "-9007199254740993"
	case 2:
		request.Conditions[2].Numeric.Type = "decimal64"
		request.Conditions[2].Numeric.Scale = 9
		request.Conditions[2].Value = "-0.000000001"
		numericValue = "-0.000000001"
	}
	records := []tsRecord{}
	for index := 0; index < total; index++ {
		id := fmt.Sprintf("%032x", index+1)
		// Each trace has both a shared-span match and a distinct-service match.
		first := tsGeneratedRecord(request, id, index*2, 7)
		first.Values[exploration.Field{Kind: "attribute", Key: "work.metric"}] = numericValue
		first.Values[exploration.Field{Kind: "resource", Key: "😀.qualified"}] = "雪"
		first.Values[exploration.Field{Kind: "resource", Key: "\ue000.qualified"}] = "✓"
		if empty {
			first.Values[exploration.Field{Kind: "resource", Key: "team"}] = ""
		} else {
			delete(first.Values, exploration.Field{Kind: "resource", Key: "team"})
		}
		second := tsGeneratedRecord(request, id, index*2+1, 2)
		second.Values[exploration.Field{Kind: "column", Key: "service"}] = "worker"
		records = append(records, first, second)
	}
	items, count, err := exploration.ReferenceTraceSearch(request, tsLogical(records))
	if err != nil {
		panic(err)
	}
	expected, wantTotal := tsOracle(request, records)
	if count != wantTotal || !tsItemsEqual(items, expected) {
		panic("independent source oracle differs")
	}
	plan, err := exploration.CompileTraceSearch(request)
	if err != nil {
		panic(err)
	}
	state := "observed"
	if total == 0 {
		state = "no_data"
	}
	result := &exploration.TraceSearchResult{Request: request, ReadAt: time.UnixMilli(request.ToMS + 1).UTC().Format(time.RFC3339Nano), Basis: strings.Repeat("b", 64), State: state, MatchedTraces: count, HasMore: total > limit, Items: items, Plan: plan, Warnings: []string{"Bounded representative witnesses; not complete traces."}}
	return exploration.SavedCapture{CaptureReceipt: exploration.CaptureReceipt{Version: 1, Title: "Exact generated search evidence", CapturedAt: time.UnixMilli(request.ToMS + 2).UTC().Format(time.RFC3339Nano), Question: exploration.CaptureQuestion{Kind: "trace-search", TraceSearch: &request}}, TraceSearch: result}
}

type searchCaptureWire383 struct {
	Page     *exploration.CapturePage `json:"page"`
	Accepted bool                     `json:"accepted"`
	Label    string                   `json:"label"`
}

const searchCaptureBatch383 = `import assert from 'node:assert/strict';import fs from 'node:fs';import {validate} from '../ui/saved-capture-protocol.mjs';const cases=JSON.parse(fs.readFileSync(0,'utf8'));for(const [i,c]of cases.entries()){const before=structuredClone(c.page);let ok=false;try{validate(c.page,c.page.receipt.id);ok=true}catch{}assert.equal(ok,c.accepted,'capture wire case '+i+' '+c.label);assert.deepEqual(c.page,before,'mutated capture wire '+i)}process.stdout.write(JSON.stringify({cases:cases.length}));`

func TestHegelCapturedTraceSearch383(t *testing.T) {
	var cases []searchCaptureWire383
	hegel.Test(t, func(tc *hegel.T) {
		seed := hegel.Draw(tc, hegel.Integers(1, 10000))
		total := hegel.Draw(tc, hegel.Integers(0, 12))
		limit := hegel.Draw(tc, hegel.Integers(1, 8))
		empty := hegel.Draw(tc, hegel.Booleans())
		capture := generatedSearchCapture383(seed, total, limit, empty)
		dir := t.TempDir()
		store := exploration.CaptureStore{Dir: dir}
		id, err := writeSearchCapture383(dir, capture)
		if err != nil {
			tc.Fatal(err)
		}
		capture.ID = id // Clones represent a published capture, not an unpublished draft.
		loaded, err := store.Load(id)
		if err != nil {
			tc.Fatal("valid capture refused", err)
		}
		page, err := loaded.Page("")
		if err != nil || page.TraceSearch == nil || !reflect.DeepEqual(page.TraceSearch, capture.TraceSearch) || page.Receipt.Question.TraceSearch.Limit != limit {
			tc.Fatal("immutable search page changed evidence", err)
		}
		if page.RetainedRecords != len(page.TraceSearch.Items) || page.RecordsLimited != page.TraceSearch.HasMore || page.NextCursor != "" || page.PreviousCursor != "" {
			tc.Fatal("representative witnesses masquerade as record pages")
		}
		if page.TraceSearch.MatchedTraces != strconv.Itoa(total) || page.TraceSearch.HasMore != (total > limit) || len(page.TraceSearch.Items) != min(total, limit) {
			tc.Fatal("bounded prefix redefined full matching count")
		}
		cases = append(cases, searchCaptureWire383{Page: page, Accepted: true, Label: "valid"})
		if total > 0 {
			var found bool
			for _, value := range page.TraceSearch.Items[0].Witnesses[3].FieldValues {
				if value.Field.Kind == "resource" && value.Field.Key == "team" {
					found = true
					if (value.Value == nil) == empty {
						tc.Fatal("missing and empty qualified values collapsed")
					}
				}
			}
			if !found {
				tc.Fatal("qualified witness omitted")
			}
		}
		variants := []func(*exploration.SavedCapture){
			func(c *exploration.SavedCapture) { c.Question.Kind = "explore" },
			func(c *exploration.SavedCapture) { c.Question.Explore = &exploration.Query{} },
			func(c *exploration.SavedCapture) { c.TraceSearch.Request.Limit++ },
			func(c *exploration.SavedCapture) { c.TraceSearch.ReadAt = "not-a-clock" },
			func(c *exploration.SavedCapture) { c.TraceSearch.Basis = "" },
			func(c *exploration.SavedCapture) { c.TraceSearch.State = "healthy" },
			func(c *exploration.SavedCapture) { c.TraceSearch.MatchedTraces = "01" },
			func(c *exploration.SavedCapture) { c.TraceSearch.HasMore = !c.TraceSearch.HasMore },
		}
		if total > 0 {
			variants = append(variants, func(c *exploration.SavedCapture) { c.TraceSearch.Items[0].Witnesses[0].ConditionIndex = 99 }, func(c *exploration.SavedCapture) { c.TraceSearch.Items[0].Witnesses[0].MatchingRecords = "0" }, func(c *exploration.SavedCapture) { c.TraceSearch.Items[0].Witnesses[0].TimestampNS = "0" },
				func(c *exploration.SavedCapture) {
					for i := range c.TraceSearch.Items[0].Witnesses[2].FieldValues {
						v := &c.TraceSearch.Items[0].Witnesses[2].FieldValues[i]
						if v.Field.Kind == "attribute" && v.Field.Key == "work.metric" {
							bad := "+9007199254740993"
							v.Value = &bad
						}
					}
				},
				func(c *exploration.SavedCapture) {
					for i := range c.TraceSearch.Items[0].Witnesses[3].FieldValues {
						v := &c.TraceSearch.Items[0].Witnesses[3].FieldValues[i]
						if v.Field.Kind == "resource" && v.Field.Key == "team" {
							if v.Value == nil {
								empty := ""
								v.Value = &empty
							} else {
								v.Value = nil
							}
						}
					}
				})
		}
		original, _ := json.Marshal(capture)
		for index, mutate := range variants {
			var bad exploration.SavedCapture
			if json.Unmarshal(original, &bad) != nil {
				tc.Fatal("clone failed")
			}
			mutate(&bad)
			badID, e := writeSearchCapture383(dir, bad)
			if e != nil {
				tc.Fatal(e)
			}
			bad.ID = badID
			badPage := &exploration.CapturePage{Receipt: bad.CaptureReceipt, TraceSearch: bad.TraceSearch, Explore: bad.Explore, RetainedRecords: len(bad.TraceSearch.Items), RecordsLimited: bad.TraceSearch.HasMore}
			cases = append(cases, searchCaptureWire383{Page: badPage, Accepted: false, Label: fmt.Sprintf("semantic-mutant-%d", index)})
			if _, e = store.Load(badID); e == nil {
				tc.Fatal("semantically invalid rehashed capture admitted", index)
			}
		}
	}, options(100)...)
	wire, err := json.Marshal(cases)
	if err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("node", "--input-type=module", "-e", searchCaptureBatch383)
	cmd.Stdin = bytes.NewReader(wire)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("mandatory batched capture Go/JS admission: %v\n%s", err, out)
	}
	t.Log(string(out))
}
