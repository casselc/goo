package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/casselc/goo/measurements"
	"github.com/casselc/goo/workbench"
)

func TestNativeHistogramViewListHTTP373(t *testing.T) {
	reads := 0
	previous := http.DefaultTransport
	http.DefaultTransport = templateNoSourceTransport{&reads}
	t.Cleanup(func() { http.DefaultTransport = previous })
	dir := t.TempDir()
	mux := http.NewServeMux()
	registerNativeHistogramViews(mux, dir)
	var cases []planningWireCase
	call := func(name, path string, want int) *httptest.ResponseRecorder {
		t.Helper()
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, httptest.NewRequest("GET", path, nil))
		if w.Code != want {
			t.Fatalf("%s: %d want%d: %s", name, w.Code, want, w.Body.String())
		}
		if w.Header().Get("Cache-Control") != "no-store" {
			t.Fatal("cacheable metadata")
		}
		if want < 400 {
			cases = append(cases, planningWireCase{Name: name, Method: "GET", Path: path, Status: w.Code, Headers: map[string]string{"Content-Type": w.Header().Get("Content-Type"), "Cache-Control": w.Header().Get("Cache-Control")}, Response: w.Body.String()})
		}
		return w
	}
	decode := func(w *httptest.ResponseRecorder) *workbench.NativeHistogramViewList {
		t.Helper()
		var p workbench.NativeHistogramViewList
		if e := json.Unmarshal(w.Body.Bytes(), &p); e != nil {
			t.Fatal(e)
		}
		return &p
	}
	empty := decode(call("empty", "/native-histogram-views", 200))
	if empty.Partial || len(empty.Items) != 0 || len(empty.Diagnostics) != 0 || empty.NextCursor != "" {
		t.Fatal(empty)
	}
	raw, e := os.ReadFile("../../config/measurements/checkout-duration.yaml")
	if e != nil {
		t.Fatal(e)
	}
	d, e := measurements.Read(string(raw))
	if e != nil {
		t.Fatal(e)
	}
	in := workbench.NativeHistogramViewInput{Version: 1, Title: "Display", Question: measurements.Question{Definition: d, FromMS: 1000, ToMS: 61000}, Declaration: workbench.NativeHistogramViewDeclaration{Version: "goo.view/1", Scope: workbench.NativeHistogramViewScope{Kind: "measurement-question", Signal: "metrics"}, Selection: workbench.NativeHistogramViewSelection{Kind: "export-interval"}, Kind: workbench.ViewKind{Name: "histogram", Version: 1}, Display: workbench.NativeHistogramViewDisplay{Category: "bucket", Value: "observation-count", RequiresComplete: true}}}
	store := workbench.Store{Dir: filepath.Join(dir, "workbench")}
	ids := []string{}
	for i := range 25 {
		in.Title = fmt.Sprintf("Display %02d", i)
		if i%2 == 0 {
			in.Declaration.Kind.Name = "table"
		} else {
			in.Declaration.Kind.Name = "histogram"
		}
		v, e := store.CreateNativeHistogramView(in)
		if e != nil {
			t.Fatal(e)
		}
		ids = append(ids, v.ID)
	}
	slices.Sort(ids)
	full := decode(call("full-first", "/native-histogram-views", 200))
	if len(full.Items) != 20 || full.Partial || full.NextCursor != ids[19] {
		t.Fatal(full)
	}
	family := filepath.Join(store.Dir, "native-histogram-views")
	for _, i := range []int{0, 19, 20} {
		if e = os.WriteFile(filepath.Join(family, ids[i]+".json"), []byte(`{"secret":"do not expose"}`), 0600); e != nil {
			t.Fatal(e)
		}
	}
	partial := decode(call("partial-first", "/native-histogram-views", 200))
	if len(partial.Items) != 18 || len(partial.Diagnostics) != 2 || !partial.Partial || partial.NextCursor != ids[19] {
		t.Fatal(partial)
	}
	last := decode(call("partial-next", "/native-histogram-views?after="+partial.NextCursor, 200))
	if len(last.Items) != 4 || len(last.Diagnostics) != 1 || !last.Partial || last.NextCursor != "" {
		t.Fatal(last)
	}
	decode(call("after-end", "/native-histogram-views?after="+strings.Repeat("f", 64), 200))
	for _, query := range []string{"?after=", "?after", "?after=bad", "?limit=20", "?offset=0", "?after=" + ids[0] + "&after=" + ids[1], "?after=" + ids[0] + "&unknown=1"} {
		call("bad-query", "/native-histogram-views"+query, 400)
	}
	if e = os.Rename(family, family+"-held"); e != nil {
		t.Fatal(e)
	}
	if e = os.WriteFile(family, nil, 0600); e != nil {
		t.Fatal(e)
	}
	unavailable := call("storage-failure", "/native-histogram-views", 500)
	if strings.Contains(unavailable.Body.String(), dir) {
		t.Fatal("filesystem path leaked")
	}
	if reads != 0 {
		t.Fatal("listing queried telemetry", reads)
	}
	if file := os.Getenv("GOO_NATIVE_HISTOGRAM_LIST_OBSERVATIONS"); file != "" {
		raw, e := json.MarshalIndent(planningWireFixtures{Version: 1, Cases: cases}, "", "  ")
		if e != nil {
			t.Fatal(e)
		}
		if e = os.WriteFile(file, raw, 0600); e != nil {
			t.Fatal(e)
		}
	}
}
