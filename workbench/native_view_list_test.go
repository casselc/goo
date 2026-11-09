package workbench

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func TestNativeViewList373BoundedCandidates(t *testing.T) {
	store := Store{Dir: t.TempDir()}
	in := nativeViewInput371(t)
	ids := []string{}
	for i := range 25 {
		in.Title = fmt.Sprintf("Display %02d", i)
		v, e := store.CreateNativeView(in)
		if e != nil {
			t.Fatal(e)
		}
		ids = append(ids, v.ID)
	}
	slices.Sort(ids)
	for _, i := range []int{0, 19, 20} {
		raw := []byte(`{"private":"must not escape"}`)
		if i == 20 {
			raw = []byte(strings.Repeat("x", MaxBytes+1))
		}
		if e := os.WriteFile(filepath.Join(store.Dir, "native-histogram-views", ids[i]+".json"), raw, 0600); e != nil {
			t.Fatal(e)
		}
	}
	first, e := store.NativeViews("")
	if e != nil {
		t.Fatal(e)
	}
	if len(first.Items) != 18 || len(first.Diagnostics) != 2 || !first.Partial || first.NextCursor != ids[19] {
		t.Fatal("cursor must follow scanned IDs including corrupt last candidate", first)
	}
	for _, d := range first.Diagnostics {
		if d.Reason != "unavailable-artifact" {
			t.Fatal("raw failure leaked")
		}
	}
	second, e := store.NativeViews(first.NextCursor)
	if e != nil {
		t.Fatal(e)
	}
	if len(second.Items) != 4 || len(second.Diagnostics) != 1 || !second.Partial || second.NextCursor != "" || second.Diagnostics[0].ID != ids[20] {
		t.Fatal("unloaded next-page corruption or continuation lost", second)
	}
	seen := map[string]bool{}
	for _, page := range []*NativeViewList{first, second} {
		raw, _ := json.Marshal(page)
		if len(raw) > NativeViewPageMaxBytes || strings.Contains(string(raw), "private") {
			t.Fatal("metadata bound or sanitation")
		}
		for _, item := range page.Items {
			if seen[item.ID] || item.Kind != "histogram" || item.Unit != "s" {
				t.Fatal("invalid/duplicate summary", item)
			}
			seen[item.ID] = true
		}
		for _, item := range page.Diagnostics {
			if seen[item.ID] {
				t.Fatal("overlap")
			}
			seen[item.ID] = true
		}
	}
	if len(seen) != 25 {
		t.Fatal("scanned IDs omitted", len(seen))
	}
	first.Items[0].Title = "mutated"
	again, e := store.NativeViews("")
	if e != nil || again.Items[0].Title == "mutated" {
		t.Fatal("caller mutation affected verified artifacts", e)
	}
}
func TestNativeViewList373EmptyPartialAndCap(t *testing.T) {
	store := Store{Dir: t.TempDir()}
	empty, e := store.NativeViews("")
	if e != nil || empty.Partial || empty.NextCursor != "" || empty.Items == nil || empty.Diagnostics == nil || len(empty.Items) != 0 {
		t.Fatal("false/nullable empty page", empty, e)
	}
	dir := filepath.Join(store.Dir, "native-histogram-views")
	if e = os.MkdirAll(dir, 0700); e != nil {
		t.Fatal(e)
	}
	for i := range 21 {
		if e = os.WriteFile(filepath.Join(dir, fmt.Sprintf("%064x.json", i)), []byte("null"), 0600); e != nil {
			t.Fatal(e)
		}
	}
	partial, e := store.NativeViews("")
	if e != nil || !partial.Partial || len(partial.Items) != 0 || len(partial.Diagnostics) != 20 || partial.NextCursor == "" {
		t.Fatal("corruption masqueraded as complete empty catalog", partial, e)
	}
	for i := 21; i <= CatalogLimit; i++ {
		if e = os.WriteFile(filepath.Join(dir, fmt.Sprintf("%064x.json", i)), nil, 0600); e != nil {
			t.Fatal(e)
		}
	}
	if _, e = store.NativeViews(""); e == nil {
		t.Fatal("catalog cap silently truncated")
	}
	if _, e = store.NativeViews("../bad"); e == nil {
		t.Fatal("bad cursor accepted")
	}
}
