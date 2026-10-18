package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/casselc/goo/exploration"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
)

func TestCaptureHTTPServerCustodyAndSourceFreePaging(t *testing.T) {
	var reads atomic.Int64
	source := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reads.Add(1)
		b, _ := io.ReadAll(r.Body)
		sql := string(b)
		switch {
		case strings.Contains(sql, "SELECT DISTINCT"):
		case strings.Contains(sql, "GROUP BY"):
			fmt.Fprintln(w, `{"g0":"worker","count":"105"}`)
		case strings.Contains(sql, "count()"):
			fmt.Fprintln(w, `{"count":"105"}`)
		default:
			for i := 0; i < 105; i++ {
				fmt.Fprintf(w, "{\"timestamp_ns\":\"%d\",\"duration_ns\":\"0\"}\n", i+1)
			}
		}
	}))
	defer source.Close()
	t.Setenv("CLICKHOUSE_URL", source.URL)
	dir := t.TempDir()
	mux := http.NewServeMux()
	registerCaptures(mux, source.Client(), dir)
	call := func(method, path, body string) *httptest.ResponseRecorder {
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, httptest.NewRequest(method, path, bytes.NewBufferString(body)))
		return w
	}
	valid := `{"title":"Worker investigation","question":{"kind":"explore","explore":{"signal":"spans","environment":"prod","fromMs":1000,"toMs":61000,"limit":10,"offset":100,"groupBy":[{"kind":"column","key":"service"}]}}}`
	for _, body := range []string{`{}`, strings.Replace(valid, `"title":"Worker investigation"`, `"title":""`, 1), valid + ` {}`, strings.Replace(valid, `"title":`, `"evidence":{},"title":`, 1), strings.Replace(valid, `"kind":"explore"`, `"kind":"unknown"`, 1)} {
		if w := call("POST", "/captures", body); w.Code != 400 {
			t.Fatal(w.Code, w.Body.String())
		}
	}
	if reads.Load() != 0 {
		t.Fatal("invalid request reached source")
	}
	w := call("POST", "/captures", valid)
	if w.Code != 201 {
		t.Fatal(w.Code, w.Body.String())
	}
	var first exploration.CapturePage
	json.Unmarshal(w.Body.Bytes(), &first)
	if len(first.Explore.Rows) != 100 || first.RetainedRecords != 105 || first.Explore.Count != "105" || first.NextCursor == "" || first.Receipt.Question.Explore.Offset != 0 {
		t.Fatal("invalid captured bank", first)
	}
	oldReads := reads.Load()
	source.Close()
	// New handler/store instance simulates process restart; source is now unavailable.
	mux = http.NewServeMux()
	registerCaptures(mux, source.Client(), dir)
	metadata := call("GET", "/captures/"+first.Receipt.ID+"/receipt", "")
	var brief struct {
		Receipt exploration.CaptureReceipt `json:"receipt"`
	}
	if metadata.Code != 200 || json.Unmarshal(metadata.Body.Bytes(), &brief) != nil || brief.Receipt.ID != first.Receipt.ID || strings.Contains(metadata.Body.String(), `"rows"`) {
		t.Fatal("invalid receipt-only response", metadata.Body.String())
	}
	if reads.Load() != oldReads {
		t.Fatal("metadata read queried telemetry")
	}
	followups := call("GET", "/captures/"+first.Receipt.ID+"/followups", "")
	var next exploration.FollowupSet
	if followups.Code != 200 || json.Unmarshal(followups.Body.Bytes(), &next) != nil || next.Origin != first.Receipt.ID || reads.Load() != oldReads {
		t.Fatal("follow-up planning queried source", followups.Body.String())
	}
	for _, suffix := range []string{"?row=100", "?row=-1", "?row=00", "?trace=forged", "?row=0&row=1"} {
		if call("GET", "/captures/"+first.Receipt.ID+"/followups"+suffix, "").Code != 400 {
			t.Fatal("unchecked follow-up selector", suffix)
		}
	}
	for _, path := range []string{"/captures", "/captures/" + first.Receipt.ID, "/captures/" + first.Receipt.ID + "?cursor=" + first.NextCursor} {
		w := call("GET", path, "")
		if w.Code != 200 || w.Header().Get("Cache-Control") != "no-store" {
			t.Fatal(w.Code, w.Body.String())
		}
	}
	if reads.Load() != oldReads {
		t.Fatal("reopen queried telemetry")
	}
	w = call("GET", "/captures/"+first.Receipt.ID+"?cursor="+first.NextCursor, "")
	var last exploration.CapturePage
	json.Unmarshal(w.Body.Bytes(), &last)
	if len(last.Explore.Rows) != 5 || last.Explore.Count != "105" || last.Explore.Basis != first.Explore.Basis || last.NextCursor != "" || last.PreviousCursor == "" {
		t.Fatal(last)
	}
	for _, suffix := range []string{"?offset=100", "?cursor=bad", "?cursor=a&cursor=b", "?cursor=" + first.NextCursor + "&limit=2"} {
		if w := call("GET", "/captures/"+first.Receipt.ID+suffix, ""); w.Code != 400 {
			t.Fatal(w.Code, w.Body.String())
		}
	}
	bad := strings.Replace(valid, `"question":`, `"parent":"`+first.Receipt.ID+`","question":`, 1)
	bad = strings.Replace(bad, `"environment":"prod"`, `"environment":"other"`, 1)
	if w = call("POST", "/captures", bad); w.Code != 409 {
		t.Fatal(w.Code, w.Body.String())
	}
	if w = call("POST", "/captures", `{"parent":"`+first.Receipt.ID+`"}`); w.Code != 502 || strings.Contains(w.Body.String(), `"count":"0"`) {
		t.Fatal("outage became new empty receipt", w.Code, w.Body.String())
	}
	if w = call("GET", "/captures/"+first.Receipt.ID, ""); w.Code != 200 {
		t.Fatal("refresh failure lost predecessor")
	}
}
