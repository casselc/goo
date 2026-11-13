package exploration

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
)

func exceptionGroupFixture416(fingerprint string) string {
	raw, _ := json.Marshal(map[string]any{"fingerprint": fingerprint, "count": "2", "first_ns": "1000001", "last_ns": "1000002", "type_min": "Error", "type_max": "Error", "stack_min": "frame", "stack_max": "frame", "affected": [][]any{{"api", 0, "", 0, ""}}, "affected_count": "1"})
	return string(raw) + "\n"
}
func TestExceptionGroups416ConservationProjectionAndRefusals(t *testing.T) {
	q := Query{Signal: "logs", Environment: "prod", FromMS: 1, ToMS: 60001, Limit: 100}
	identity := ReferenceLogExceptionFingerprint(map[string]string{"exception.type": "Error", "exception.stacktrace": "frame"}).Fingerprint
	for _, mode := range []string{"observed", "no_data", "changed", "mismatch-input", "late-failure", "bad-context", "out-of-window"} {
		t.Run(mode, func(t *testing.T) {
			calls := 0
			source := Source{Endpoint: "http://fixture.invalid", Client: &http.Client{Transport: recordWireTransport(func(r *http.Request) (*http.Response, error) {
				calls++
				body, _ := io.ReadAll(r.Body)
				if r.URL.Query().Get("readonly") != "1" {
					t.Fatal("writable source")
				}
				payload := "{\"count\":\"2\"}\n"
				status := 200
				if mode == "no_data" {
					payload = "{\"count\":\"0\"}\n"
				}
				if mode == "changed" {
					payload = "{\"count\":\"3\"}\n"
				}
				if calls == 2 {
					if !strings.Contains(string(body), "groupArraySorted(33)") || !strings.Contains(string(body), "isNotNull(namespace)") {
						t.Fatal("affected projection not bounded/canonical")
					}
					payload = exceptionGroupFixture416(identity)
					if mode == "no_data" {
						payload = ""
					}
					if mode == "mismatch-input" {
						payload = strings.ReplaceAll(payload, "frame", "other")
					}
					if mode == "late-failure" {
						status = 503
						payload = "source unavailable"
					}
					if mode == "bad-context" {
						payload = strings.Replace(payload, `"affected_count":"1"`, `"affected_count":"0"`, 1)
					}
					if mode == "out-of-window" {
						payload = strings.Replace(payload, "1000001", "0", 1)
					}
				}
				return &http.Response{StatusCode: status, Body: io.NopCloser(strings.NewReader(payload)), Header: make(http.Header)}, nil
			})}}
			result, e := source.FetchExceptionGroups(context.Background(), q)
			if mode == "mismatch-input" || mode == "late-failure" || mode == "bad-context" || mode == "out-of-window" {
				if e == nil || result != nil {
					t.Fatal("malformed/partial projection admitted", result, e)
				}
				return
			}
			if e != nil || calls != 2 || result.Version != ExceptionGroupsVersion {
				t.Fatal(result, e, calls)
			}
			want := map[string]string{"observed": "observed", "changed": "changed_between_reads", "no_data": "no_data"}[mode]
			if result.State != want {
				t.Fatal(result.State, want)
			}
			if mode == "observed" {
				g := result.Groups[0]
				if g.Inputs.Stacktrace != "frame" || g.Affected[0].Namespace != nil || g.Affected[0].Version != nil || g.AffectedLimited || g.AffectedCount != "1" {
					t.Fatal(g)
				}
			}
		})
	}
	for _, change := range []func(*Query){func(q *Query) { q.Offset = 1 }, func(q *Query) { q.GroupBy = []Field{{Kind: "column", Key: "service"}} }, func(q *Query) { q.BucketMS = 1000 }, func(q *Query) { q.SpanStatus = true }} {
		bad := q
		change(&bad)
		if _, e := CompileExceptionGroups(bad); e == nil {
			t.Fatal("display projection ignored")
		}
	}
}

func TestExceptionGroups416AffectedPrefixAdmission(t *testing.T) {
	q := Query{Signal: "spans", Environment: "prod", FromMS: 1, ToMS: 60001, Limit: 100}
	id := ReferenceLogExceptionFingerprint(map[string]string{"exception.type": "Error", "exception.stacktrace": "frame"}).Fingerprint
	affected := make([][]any, 33)
	for i := range affected {
		affected[i] = []any{fmt.Sprintf("service-%02d", i), 0, "", 0, ""}
	}
	affected[1] = []any{"service-01", 1, "", 1, ""}
	wire := map[string]any{"fingerprint": id, "count": "36", "first_ns": "1000001", "last_ns": "1000002", "type_min": "Error", "type_max": "Error", "stack_min": "frame", "stack_max": "frame", "affected": affected, "affected_count": "36"}
	raw, _ := json.Marshal(wire)
	g, e := decodeExceptionGroup(q, raw)
	if e != nil || len(g.Affected) != 32 || !g.AffectedLimited || g.AffectedCount != "36" || g.Affected[0].Namespace != nil || g.Affected[1].Namespace == nil || *g.Affected[1].Namespace != "" || g.Affected[31].Service != "service-31" {
		t.Fatal(g, e)
	}
	for _, mutate := range []func(){func() { affected[2], affected[3] = affected[3], affected[2] }, func() { affected[2] = affected[1] }, func() { wire["affected_count"] = "32" }, func() { wire["first_ns"] = "0" }, func() { wire["unknown"] = true }} {
		saved, _ := json.Marshal(wire)
		mutate()
		raw, _ = json.Marshal(wire)
		if _, e = decodeExceptionGroup(q, raw); e == nil {
			t.Fatal("invalid prefix/provenance admitted")
		}
		json.Unmarshal(saved, &wire)
		// Rebind the mutable tuple array after restoring the independent wire copy.
		affected = make([][]any, 33)
		for i := range affected {
			affected[i] = []any{fmt.Sprintf("service-%02d", i), 0, "", 0, ""}
		}
		affected[1] = []any{"service-01", 1, "", 1, ""}
		wire["affected"] = affected
	}
}
