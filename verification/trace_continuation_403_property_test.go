package verification

import (
	"bytes"
	"encoding/json"
	"os/exec"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/identity"
	"hegel.dev/go/hegel"
)

type tracePageVector403 struct {
	Request exploration.TraceSearchRequest `json:"request"`
	Hash    string                         `json:"hash"`
	Result  *exploration.TraceSearchResult `json:"result"`
	Page    *exploration.CapturePage       `json:"page"`
}

// Each generated bank is a fixed reference population. It proves exact page
// composition and admission, not a stable snapshot across production reads.
func TestHegelTraceContinuation403(t *testing.T) {
	vectors := []tracePageVector403{}
	hegel.Test(t, func(tc *hegel.T) {
		seed := hegel.Draw(tc, hegel.Integers(1, 100000))
		total := hegel.Draw(tc, hegel.Integers(1, 9))
		limit := hegel.Draw(tc, hegel.Integers(1, 3))
		capture := generatedSearchCapture383(seed, total, limit, seed%2 == 0)
		request := *capture.Question.TraceSearch
		// Unicode keys outside the BMP, HTML characters and literal numeric tokens
		// occur in the existing generated qualified witness bank.
		// generatedSearchCapture383 already checks the complete reference bank
		// against the independent set-intersection and witness oracle.
		complete := generatedSearchCapture383(seed, total, 100, seed%2 == 0).TraceSearch.Items
		hash, err := exploration.TraceSearchRequestHash(request)
		if err != nil {
			tc.Fatal(err)
		}
		var position *exploration.TraceSearchPosition
		seen := []exploration.TraceSearchItem{}
		for turn := 0; turn < total+2; turn++ {
			items := []exploration.TraceSearchItem{}
			for _, item := range complete {
				if position == nil || item.TraceID > position.LastTraceID {
					items = append(items, item)
				}
			}
			remaining := len(items)
			more := remaining > limit
			if more {
				items = items[:limit]
			}
			result := *capture.TraceSearch
			if remaining == 0 {
				result.State = "no_data"
			}
			result.Items = items
			result.MatchedTraces = strconv.Itoa(total)
			result.RemainingTraces = strconv.Itoa(remaining)
			result.HasMore = more
			result.Position = position
			result.NextPosition = nil
			if more {
				result.NextPosition = &exploration.TraceSearchPosition{Version: exploration.TraceSearchPositionVersion, RequestHash: hash, LastTraceID: items[len(items)-1].TraceID}
			}
			if position == nil {
				result.Plan, err = exploration.CompileTraceSearch(request)
			} else {
				result.Plan, err = exploration.CompileTraceSearchPage(exploration.TraceSearchPageRequest{Request: request, Position: *position})
			}
			if err != nil {
				tc.Fatal(err)
			}
			if err = exploration.ValidateTraceSearchResult(&result); err != nil {
				tc.Fatal(err)
			}
			stored := exploration.SavedCapture{CaptureReceipt: exploration.CaptureReceipt{Version: 1, ID: strings.Repeat("a", 64), Title: "Generated trace page", CapturedAt: time.UnixMilli(request.ToMS + 3).UTC().Format(time.RFC3339Nano), Question: exploration.CaptureQuestion{Kind: "trace-search", TraceSearch: &request, TraceSearchPosition: position}}, TraceSearch: &result}
			page, err := stored.Page("")
			if err != nil {
				tc.Fatal(err)
			}
			before, _ := json.Marshal(page)
			vectors = append(vectors, tracePageVector403{Request: request, Hash: hash, Result: &result, Page: page})
			seen = append(seen, items...)
			after, _ := json.Marshal(page)
			if !bytes.Equal(before, after) {
				tc.Fatal("page carrier mutated")
			}
			if remaining == 0 {
				break
			}
			if !more {
				position = &exploration.TraceSearchPosition{Version: exploration.TraceSearchPositionVersion, RequestHash: hash, LastTraceID: items[len(items)-1].TraceID}
			} else {
				position = result.NextPosition
			}
		}
		if !reflect.DeepEqual(seen, complete) {
			tc.Fatal("page partition changes lexical IDs or qualified witnesses")
		}
	}, options(24)...)
	raw, _ := json.Marshal(vectors)
	if err := artifact("trace-continuation403-carriers.json", raw); err != nil {
		t.Fatal(err)
	}
	cmd := exec.CommandContext(t.Context(), "node", "trace_continuation_403_fixture.mjs")
	cmd.Stdin = bytes.NewReader(raw)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Go/JS page carriers: %v\n%s", err, out)
	}
	var got struct{ Cases, Controls int }
	if json.Unmarshal(out, &got) != nil || got.Cases != len(vectors) || got.Controls < len(vectors)*8 {
		t.Fatalf("incomplete page controls: %s", out)
	}
	t.Logf("%d generated complete page carriers / %d Go+JS position and retained-envelope controls", got.Cases, got.Controls)
}

func TestTracePositionHashGolden403(t *testing.T) {
	type vector struct {
		Request exploration.TraceSearchRequest `json:"request"`
		Hash    string                         `json:"hash"`
	}
	vectors := []vector{}
	for _, literal := range []string{"<tag>&'\\\"", "雪😀\ue000", "line\u2028paragraph\u2029", "9007199254740993", ""} {
		r := tsGeneratedRequest(7, 1, 2)
		r.Conditions = []exploration.Predicate{tsLeaf("attribute", "😀.qualified", "eq", literal)}
		hash, err := exploration.TraceSearchRequestHash(r)
		if err != nil {
			t.Fatal(err)
		}
		vectors = append(vectors, vector{r, hash})
	}
	for _, selector := range []identity.EnvironmentSelector{
		{Version: identity.EnvironmentSelectorVersion, Policy: "current-only", Scope: identity.EnvironmentScope{Mode: "named", Value: "prod-雪"}},
		{Version: identity.EnvironmentSelectorVersion, Policy: "effective", Scope: identity.EnvironmentScope{Mode: "named", Value: "prod-雪"}},
		{Version: identity.EnvironmentSelectorVersion, Policy: "effective", Scope: identity.EnvironmentScope{Mode: "empty"}},
		{Version: identity.EnvironmentSelectorVersion, Policy: "effective", Scope: identity.EnvironmentScope{Mode: "unreported"}},
	} {
		r := tsGeneratedRequest(7, 4, 2)
		r.Environment = ""
		r.EnvironmentSelector = &selector
		hash, err := exploration.TraceSearchRequestHash(r)
		if err != nil {
			t.Fatal(err)
		}
		vectors = append(vectors, vector{r, hash})
	}
	raw, _ := json.Marshal(vectors)
	cmd := exec.CommandContext(t.Context(), "node", "--input-type=module", "-e", `import fs from 'node:fs';import assert from 'node:assert/strict';import {requestHash} from '../ui/trace-search-position.mjs';const cases=JSON.parse(fs.readFileSync(0,'utf8'));for(const c of cases)assert.equal(requestHash(c.request),c.hash);console.log(cases.length+' canonical Unicode/escaping/numeric hash vectors');`)
	cmd.Stdin = bytes.NewReader(raw)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("canonical position hash: %v\n%s", err, out)
	}
	t.Log(strings.TrimSpace(string(out)))
}
