package exploration

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"
)

// The fixture models independent source responses, not the browser-held pair.
func pairedSource428(t *testing.T, mode string) Source {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		sql := string(raw)
		selection := r.URL.Query().Get("param_from") == "61000"
		if selection && mode == "refused" {
			http.Error(w, "source unavailable", http.StatusServiceUnavailable)
			return
		}
		n := 2
		if selection {
			n = 3
			if mode == "empty" {
				n = 0
			}
		}
		switch {
		case strings.Contains(sql, "SELECT DISTINCT"):
		case strings.Contains(sql, "GROUP BY"):
			groups := n
			if selection && mode == "diagnostic" {
				groups++
			}
			if groups > 0 {
				fmt.Fprintf(w, "{\"g0\":\"worker\",\"count\":\"%d\"}\n", groups)
			}
		case strings.Contains(sql, "toString(count()) AS count"):
			fmt.Fprintf(w, "{\"count\":\"%d\"}\n", n)
		default:
			if selection && mode == "limit" {
				n = 1002
			}
			for i := 0; i < n; i++ {
				fmt.Fprintf(w, "{\"timestamp_ns\":\"%d\",\"duration_ns\":\"0\",\"SpanId\":\"%016x\"}\n", i+1, i+1)
			}
		}
	}))
	t.Cleanup(server.Close)
	return Source{Endpoint: server.URL}
}
func pairedQuestion428() CaptureQuestion {
	a := Query{Signal: "spans", Environment: "prod", FromMS: 1000, ToMS: 61000, Limit: 2, Offset: 100, GroupBy: []Field{{Kind: "column", Key: "service"}}}
	b := a
	b.FromMS = 61000
	b.ToMS = 121000
	return CaptureQuestion{Kind: "comparison", Comparison: &ComparisonQuestion{Baseline: a, Selection: b}}
}
func pairedObjects428(t *testing.T, s CaptureStore) []*SavedCapture {
	t.Helper()
	entries, err := os.ReadDir(s.Dir)
	if err != nil {
		t.Fatal(err)
	}
	var objects []*SavedCapture
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		c, err := s.Load(strings.TrimSuffix(entry.Name(), ".json"))
		if err != nil {
			t.Fatal(err)
		}
		objects = append(objects, c)
	}
	return objects
}
func TestPairedCapture428SecondSourceFailureRetainsBaseline(t *testing.T) {
	for _, mode := range []string{"refused", "limit"} {
		t.Run(mode, func(t *testing.T) {
			store := CaptureStore{Dir: t.TempDir()}
			question := pairedQuestion428()
			pair, err := store.Create(t.Context(), pairedSource428(t, mode), "Pair evidence", question, "")
			if err == nil || pair != nil {
				t.Fatal("source refusal became a published pair", pair, err)
			}
			objects := pairedObjects428(t, store)
			if len(objects) != 1 || objects[0].Question.Kind != "explore" || !reflect.DeepEqual(*objects[0].Question.Explore, question.Normalized().Comparison.Baseline) || objects[0].Explore.Count != "2" {
				t.Fatal("baseline child was lost or changed", objects)
			}
			before, _ := json.Marshal(objects[0])
			again, err := store.Load(objects[0].ID)
			after, _ := json.Marshal(again)
			if err != nil || string(before) != string(after) {
				t.Fatal("retained child did not reopen independently", err)
			}
		})
	}
}
func TestPairedCapture428RetainsEmptyAndDiagnosticSides(t *testing.T) {
	for _, mode := range []string{"empty", "diagnostic"} {
		t.Run(mode, func(t *testing.T) {
			store := CaptureStore{Dir: t.TempDir()}
			q := pairedQuestion428()
			pair, err := store.Create(t.Context(), pairedSource428(t, mode), "Independent pair", q, "")
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(pair.Question, q.Normalized()) || len(pairedObjects428(t, store)) != 3 {
				t.Fatal("pair changed original fixed operands")
			}
			child, err := store.Load(pair.Comparison.SelectionID)
			if err != nil {
				t.Fatal(err)
			}
			wantState, wantCount, wantPair := "no_data", "0", "no_data"
			if mode == "diagnostic" {
				wantState, wantCount, wantPair = "changed_between_reads", "3", "unknown"
			}
			if child.Explore.State != wantState || child.Explore.Count != wantCount || pair.Comparison.Request.Selection.State != wantState || pair.Comparison.Result.State != wantPair {
				t.Fatal("source uncertainty was repaired by storage", child.Explore.State, child.Explore.Count, pair.Comparison.Result.State)
			}
			if pair.Comparison.Result.ViewComparison == nil || pair.Comparison.Result.ViewComparison.Selection.Completeness != nil || pair.Comparison.Result.ViewComparison.Baseline.Completeness != nil {
				t.Fatal("stored pair invented coverage")
			}
			if mode == "diagnostic" && pair.Comparison.Result.ViewComparison.Changes["count"].Value != nil {
				t.Fatal("disagreement acquired a numeric difference")
			}
		})
	}
}

// Exercise the actual immutable publisher after real child acquisition. A
// deliberate incompatible occupant blocks only the pair's content-address path;
// this is not an injected filesystem fault inside createComparison's call stack.
func TestPairedCapture428PairPublishFailureDoesNotRollbackChildren(t *testing.T) {
	store := CaptureStore{Dir: t.TempDir()}
	source := pairedSource428(t, "ok")
	q := pairedQuestion428().Normalized()
	children := make([]*SavedCapture, 0, 2)
	for _, question := range []Query{q.Comparison.Baseline, q.Comparison.Selection} {
		c, err := store.Create(t.Context(), source, "Pair evidence", CaptureQuestion{Kind: "explore", Explore: &question}, "")
		if err != nil {
			t.Fatal(err)
		}
		children = append(children, c)
	}
	a, err := CohortFromResult(children[0].Explore)
	if err != nil {
		t.Fatal(err)
	}
	b, err := CohortFromResult(children[1].Explore)
	if err != nil {
		t.Fatal(err)
	}
	request := ComparisonRequest{Baseline: a, Selection: b}
	result, err := CompareCohorts(request)
	if err != nil {
		t.Fatal(err)
	}
	candidate := SavedCapture{CaptureReceipt: CaptureReceipt{Version: 1, Title: "Pair evidence", CapturedAt: time.Now().UTC().Format(time.RFC3339Nano), Question: q}, Comparison: &SavedComparison{BaselineID: children[0].ID, SelectionID: children[1].ID, Request: request, Result: result}}
	id, _, err := captureDigest(candidate)
	if err != nil {
		t.Fatal(err)
	}
	if err = os.Mkdir(filepath.Join(store.Dir, id+".json"), 0700); err != nil {
		t.Fatal(err)
	}
	pair, err := store.publish(candidate)
	if err == nil || pair != nil {
		t.Fatal("pair publication collision was hidden", pair, err)
	}
	for _, child := range children {
		before, _ := json.Marshal(child)
		retained, err := store.Load(child.ID)
		after, _ := json.Marshal(retained)
		if err != nil || string(before) != string(after) {
			t.Fatal("pair failure rolled back useful child", err)
		}
	}
	if len(pairedObjects428(t, store)) != 2 {
		t.Fatal("failed pair became a catalog artifact")
	}
}
