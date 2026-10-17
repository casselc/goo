package verification

import (
	"encoding/json"
	"strconv"
	"strings"
	"testing"

	"github.com/casselc/goo/prebuilt"
	"hegel.dev/go/hegel"
)

func TestHegelPrebuiltEvidenceCannotBeBorrowedOrGuessed(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		profiles := prebuilt.Catalog().Profiles
		profile := profiles[hegel.Draw(tc, hegel.Integers(0, len(profiles)-1))]
		resource := map[string]string{}
		for _, key := range profile.Resources {
			resource[key] = "bound-value"
		}
		signals := []prebuilt.Contract{}
		for _, req := range profile.Requirements {
			signals = append(signals, req.AnyOf[0])
		}
		mode := hegel.Draw(tc, hegel.Integers(0, 3))
		switch mode {
		case 0:
			signals[0].SchemaURL = ""
		case 1:
			signals[0].Unit = "unknown"
		case 2:
			signals = signals[1:]
		case 3:
			delete(resource, profile.Resources[0])
		}
		// A separate perfectly instrumented scope cannot fill this one's holes.
		complete := []prebuilt.Contract{}
		for _, req := range profile.Requirements {
			complete = append(complete, req.AnyOf[0])
		}
		in := prebuilt.Inventory{APIVersion: "goo/inventory/v1alpha1", Basis: "property", Scopes: []prebuilt.Scope{{ID: "incomplete", Resource: resource, Signals: signals}, {ID: "other", Resource: map[string]string{"service.name": "other", "service.instance.id": "b", "host.id": "host"}, Signals: complete}}}
		out, err := prebuilt.Discover(in)
		if err != nil {
			tc.Fatal(err)
		}
		for _, p := range out {
			if p.Scope == "incomplete" && p.Profile == profile.ID && p.Status == "compatible-shape" {
				tc.Fatal("invented compatible capability", p)
			}
		}
		before, _ := json.Marshal(out)
		in.Scopes[0], in.Scopes[1] = in.Scopes[1], in.Scopes[0]
		in.Scopes[0].Signals = append(in.Scopes[0].Signals, complete...)
		again, err := prebuilt.Discover(in)
		if err != nil {
			tc.Fatal(err)
		}
		after, _ := json.Marshal(again)
		if string(before) != string(after) {
			tc.Fatal("ordering or duplicate signatures changed proposals")
		}
	}, options(200)...)
}

func TestHegelExpandedHTTPPolicyClassification(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		boundary := hegel.Draw(tc, hegel.Integers(400, 599))
		status := hegel.Draw(tc, hegel.Integers(100, 650))
		var contract prebuilt.Contract
		for _, p := range prebuilt.Catalog().Profiles {
			if p.ID == "otel.http.server.spans" {
				contract = p.Requirements[0].AnyOf[0]
			}
		}
		resource := map[string]string{"service.name": "checkout"}
		in := prebuilt.Inventory{APIVersion: "goo/inventory/v1alpha1", Basis: "property", Scopes: []prebuilt.Scope{{ID: "a", Resource: resource, Signals: []prebuilt.Contract{contract}}}}
		b := prebuilt.Binding{APIVersion: "goo/binding/v1alpha1", Name: "http_requests", Profile: "otel.http.server.spans", Scope: "a", PreparedTables: map[string]string{"clickhouse": "prepared_http"}, Policy: prebuilt.HTTPPolicy{FailureStatusRanges: []prebuilt.StatusRange{{Min: boundary, Max: 599}}, MissingStatus: "unknown", MissingRoute: "unknown-bucket"}}
		expanded, err := prebuilt.Expand(in, b)
		if err != nil {
			tc.Fatal(err)
		}
		prepared, err := expanded.PrepareHTTP([]prebuilt.HTTPRequest{{Kind: "server", Resource: resource, TraceID: "t", SpanID: "s", Route: "/checkout", Method: "POST", Status: &status}}, "e")
		if err != nil {
			tc.Fatal(err)
		}
		want := "unknown"
		if status >= 200 && status <= 599 {
			want = "success"
			if status >= boundary {
				want = "failure"
			}
		}
		if prepared.Facts[0]["classification"] != want {
			tc.Fatalf("status %d boundary %d: %v", status, boundary, prepared.Facts)
		}
		changed := b
		changed.Policy.FailureStatusRanges = []prebuilt.StatusRange{{Min: 200, Max: 599}}
		different, err := prebuilt.Expand(in, changed)
		if err != nil {
			tc.Fatal(err)
		}
		if expanded.Revision == different.Revision {
			tc.Fatal("policy change reused expansion identity")
		}
		missing, err := expanded.PrepareHTTP([]prebuilt.HTTPRequest{{Kind: "server", Resource: resource, TraceID: "t", SpanID: "s", Method: "POST"}}, "e")
		if err != nil {
			tc.Fatal(err)
		}
		if missing.Facts[0]["classification"] != "unknown" {
			tc.Fatal("missing status became healthy")
		}
	}, options(250)...)
}

func TestHegelLiveHTTPSourceDeduplication(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		copies := hegel.Draw(tc, hegel.Integers(1, 30))
		status := hegel.Draw(tc, hegel.Integers(200, 599))
		var contract prebuilt.Contract
		for _, p := range prebuilt.Catalog().Profiles {
			if p.ID == "otel.http.server.spans" {
				contract = p.Requirements[0].AnyOf[0]
			}
		}
		resource := map[string]string{"service.name": "checkout"}
		in := prebuilt.Inventory{APIVersion: "goo/inventory/v1alpha1", Basis: "source-property", Scopes: []prebuilt.Scope{{ID: "a", Resource: resource, Signals: []prebuilt.Contract{contract}}}}
		b := prebuilt.Binding{APIVersion: "goo/binding/v1alpha1", Name: "http_requests", Profile: "otel.http.server.spans", Scope: "a", PreparedTables: map[string]string{"clickhouse": "prepared_http"}, Policy: prebuilt.HTTPPolicy{FailureStatusRanges: []prebuilt.StatusRange{{Min: 500, Max: 599}}, MissingStatus: "unknown", MissingRoute: "unknown-bucket"}}
		e, err := prebuilt.Expand(in, b)
		if err != nil {
			tc.Fatal(err)
		}
		row := prebuilt.HTTPSpanRow{TimestampNS: "0", TraceID: strings.Repeat("1", 32), SpanID: strings.Repeat("2", 16), Resource: resource, Attributes: map[string]string{"http.request.method": "POST", "http.route": "/checkout", "http.response.status_code": strconv.Itoa(status)}}
		rows := make([]prebuilt.HTTPSpanRow, copies)
		for i := range rows {
			rows[i] = row
		}
		_, result, duplicates, err := e.EvaluateHTTPRows(rows, "e")
		if err != nil {
			tc.Fatal(err)
		}
		if duplicates != copies-1 || *result[0].Measures["requests"].Count != 1 {
			tc.Fatal("duplicate records multiplied denominator")
		}
		conflicting := row
		conflicting.TimestampNS = "1"
		if _, _, _, err = e.EvaluateHTTPRows(append(rows, conflicting), "e"); err == nil {
			tc.Fatal("conflicting identity accepted")
		}
	}, options(150)...)
}
