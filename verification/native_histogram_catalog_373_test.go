package verification

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"slices"
	"testing"

	"github.com/casselc/goo/workbench"
	"hegel.dev/go/hegel"
)

// The oracle owns candidate identity order independently from the catalog walk.
// Corrupt and vanished-target artifacts consume positions, never page capacity
// for successful views only. The final empty/partial states stay distinguishable.
func TestHegelNativeHistogramCatalog373(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		seed := hegel.Draw(tc, hegel.Integers(1, 1000000))
		store := workbench.Store{Dir: t.TempDir()}
		q, _ := nrcInputs(nil, "histogram")
		q.Definition.Source.Unit = "s"
		decl := workbench.NativeViewDeclaration{Version: "goo.view/1", Scope: workbench.NativeViewScope{Kind: "measurement-question", Signal: "metrics"}, Selection: workbench.NativeViewSelection{Kind: "export-interval"}, Kind: workbench.ViewKind{Name: "histogram", Version: 1}, Display: workbench.NativeViewDisplay{Category: "bucket", Value: "observation-count", RequiresComplete: true}}
		ids := []string{}
		views := map[string]*workbench.NativeView{}
		for i := 0; i < 43; i++ {
			decl.Kind.Name = "histogram"
			if i%2 == 1 {
				decl.Kind.Name = "table"
			}
			v, err := store.CreateNativeView(workbench.NativeViewInput{Version: 1, Title: fmt.Sprintf("Native display %d/%d 雪", seed, i), Question: q, Declaration: decl})
			if err != nil {
				tc.Fatal(err)
			}
			ids = append(ids, v.ID)
			views[v.ID] = v
		}
		slices.Sort(ids)
		bad := map[string]bool{ids[seed%20]: true, ids[20+seed%20]: true, ids[42]: true}
		dir := filepath.Join(store.Dir, "native-histogram-views")
		for id := range bad {
			if err := os.WriteFile(filepath.Join(dir, id+".json"), []byte(`{"broken":true}`), 0600); err != nil {
				tc.Fatal(err)
			}
		}
		// A listed entry whose target vanished is diagnosed, rather than skipped.
		brokenID := ids[41]
		bad[brokenID] = true
		if err := os.Remove(filepath.Join(dir, brokenID+".json")); err != nil {
			tc.Fatal(err)
		}
		if err := os.Symlink(filepath.Join(store.Dir, "missing-artifact"), filepath.Join(dir, brokenID+".json")); err != nil {
			tc.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "README.json"), []byte("ignored"), 0600); err != nil {
			tc.Fatal(err)
		}
		pages := []*workbench.NativeViewList{}
		after := ""
		for offset := 0; offset < len(ids); offset += 20 {
			page, err := store.NativeViews(after)
			if err != nil {
				tc.Fatal(err)
			}
			end := min(offset+20, len(ids))
			expectedItems := []workbench.NativeViewSummary{}
			expectedDiagnostics := []workbench.NativeViewListDiagnostic{}
			for _, id := range ids[offset:end] {
				if bad[id] {
					expectedDiagnostics = append(expectedDiagnostics, workbench.NativeViewListDiagnostic{ID: id, Reason: "unavailable-artifact"})
				} else {
					v := views[id]
					expectedItems = append(expectedItems, workbench.NativeViewSummary{ID: id, Title: v.Title, Family: v.Question.Definition.Kind, Selection: v.Declaration.Selection.Kind, Kind: v.Declaration.Kind.Name, Unit: v.Question.Definition.Source.Unit})
				}
			}
			cursor := ""
			if end < len(ids) {
				cursor = ids[end-1]
			}
			if !reflect.DeepEqual(page.Items, expectedItems) || !reflect.DeepEqual(page.Diagnostics, expectedDiagnostics) || page.Partial != (len(expectedDiagnostics) > 0) || page.NextCursor != cursor || len(page.Items)+len(page.Diagnostics) != end-offset {
				tc.Fatalf("page hid candidate loss or changed metadata: %+v", page)
			}
			pages = append(pages, page)
			after = page.NextCursor
		}
		var source *workbench.NativeView
		for _, id := range ids {
			if !bad[id] {
				var err error
				source, err = store.LoadNativeView(id)
				if err != nil {
					tc.Fatal(err)
				}
				break
			}
		}
		destination := q
		destination.Definition.Name = "destination-histogram"
		destination.Definition.Title = "Destination histogram"
		destination.Definition.Source.Services = []string{"destination-service"}
		destination.FromNS = fmt.Sprintf("%d", destination.FromMS*1000000+int64(seed%1000000))
		destination.ToNS = fmt.Sprintf("%d", destination.ToMS*1000000-int64(seed%1000000))
		reused, err := workbench.ReuseNativeView(*source, destination)
		if err != nil {
			tc.Fatal(err)
		}
		input, err := json.Marshal(map[string]any{"pages": pages, "source": source, "destination": destination, "reused": reused})
		if err != nil {
			tc.Fatal(err)
		}
		cmd := exec.CommandContext(t.Context(), "node", "native_histogram_catalog_373_fixture.mjs")
		cmd.Stdin = bytes.NewReader(input)
		out, err := cmd.CombinedOutput()
		if err != nil {
			tc.Fatalf("actual JS catalog admission: %v\n%s", err, out)
		}
		var actual struct {
			Reused workbench.NativeView
			Pages  int
		}
		if err = json.Unmarshal(out, &actual); err != nil || actual.Pages != 3 || !reflect.DeepEqual(actual.Reused, *reused) {
			tc.Fatalf("catalog/reuse authority changed: %v\n%s", err, out)
		}
	}, options(16)...)
}
