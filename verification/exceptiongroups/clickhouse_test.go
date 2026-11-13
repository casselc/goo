package exceptiongroups

import (
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/casselc/goo/exploration"
)

// Only the compiler's physical table reference is replaced. All production
// fingerprint/group SQL, source response limits and decoder logic still execute.
type inlineSource struct {
	next     http.RoundTripper
	relation string
	calls    *int
}

func (r inlineSource) RoundTrip(req *http.Request) (*http.Response, error) {
	body, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, err
	}
	req.Body.Close()
	sql := string(body)
	table := " FROM otel_logs WHERE "
	if strings.Contains(sql, " FROM otel_traces WHERE ") {
		table = " FROM otel_traces WHERE "
	}
	if strings.Count(sql, table) != 1 {
		return nil, fmt.Errorf("compiler source boundary changed; refuse rewrite")
	}
	sql = strings.Replace(sql, table, " FROM ("+r.relation+") WHERE ", 1)
	copy := req.Clone(req.Context())
	copy.Body = io.NopCloser(strings.NewReader(sql))
	copy.ContentLength = int64(len(sql))
	copy.GetBody = nil
	(*r.calls)++
	return r.next.RoundTrip(copy)
}
func hexSQL(value string) string { return "unhex('" + hex.EncodeToString([]byte(value)) + "')" }
func mapSQL(values map[string]string) string {
	parts := []string{}
	for _, key := range []string{"exception.type", "exception.stacktrace"} {
		if value, ok := values[key]; ok {
			parts = append(parts, hexSQL(key), hexSQL(value))
		}
	}
	if len(parts) == 0 {
		return "CAST(map() AS Map(String,String))"
	}
	return "map(" + strings.Join(parts, ",") + ")"
}
func inlineRecord(signal string, names []string, data []map[string]string) string {
	common := "SELECT fromUnixTimestamp64Milli(1700000000001) AS Timestamp, 'synthetic416' AS ServiceName, map('deployment.environment.name','verification416','service.version','') AS ResourceAttributes"
	if signal == "logs" {
		return common + ", " + mapSQL(data[0]) + " AS LogAttributes"
	}
	ns, as := []string{}, []string{}
	for _, name := range names {
		ns = append(ns, hexSQL(name))
	}
	for _, value := range data {
		as = append(as, mapSQL(value))
	}
	namesSQL := "CAST([] AS Array(String))"
	if len(ns) > 0 {
		namesSQL = "[" + strings.Join(ns, ",") + "]"
	}
	attrsSQL := "CAST([] AS Array(Map(String,String)))"
	if len(as) > 0 {
		attrsSQL = "[" + strings.Join(as, ",") + "]"
	}
	return common + ", " + namesSQL + " AS `Events.Name`, " + attrsSQL + " AS `Events.Attributes`"
}
func TestActualClickHouseFingerprintAndGroupDifferential(t *testing.T) {
	endpoint := os.Getenv("GOO_EXCEPTION_GROUPS_CLICKHOUSE")
	if endpoint == "" {
		t.Skip("actual readonly ClickHouse endpoint supplied by GOO_EXCEPTION_GROUPS_CLICKHOUSE")
	}
	user, password := os.Getenv("GOO_TEST_CLICKHOUSE_USER"), os.Getenv("GOO_TEST_CLICKHOUSE_PASSWORD")
	if user == "" {
		user = "lab"
	}
	if password == "" {
		password = "lab-local"
	}
	type sourceCase struct {
		name, signal string
		events       []tuple
		names        []string
		malformed    bool
	}
	full := tuple{true, "Failure雪\x00", true, "at main.go:12\n at handler\t literal \x00"}
	cases := []sourceCase{}
	logs := []tuple{full, {false, "", true, "stack"}, {true, "", true, "stack"}, {true, strings.Repeat("T", 256), true, strings.Repeat("s", 8192)}, {true, strings.Repeat("T", 257), true, "s"}, {true, "T", true, strings.Repeat("s", 8193)}, {true, string([]byte{0xff}), true, "s"}, {true, "T", true, string([]byte{0xff})}, {true, "A", true, "BC"}, {true, "AB", true, "C"}}
	for i, v := range logs {
		cases = append(cases, sourceCase{name: fmt.Sprint("logs-", i), signal: "logs", events: []tuple{v}})
	}
	spans := [][]tuple{{full, full}, {{true, "Type", false, ""}, {false, "", true, "stack"}}, {{true, "A", true, "S1"}, {true, "B", true, "S2"}}, nil, {{true, "T", true, strings.Repeat("é", 4096)}}}
	for _, n := range []int{32, 33} {
		values := make([]tuple, n)
		for i := range values {
			values[i] = tuple{true, "T", true, "S"}
		}
		spans = append(spans, values)
	}
	exact := make([]tuple, 8)
	for i := range exact {
		exact[i] = tuple{true, "Type", true, strings.Repeat("s", 8188)}
	}
	spans = append(spans, exact)
	over := append([]tuple(nil), exact...)
	over[0].Stack += "x"
	spans = append(spans, over)
	for i, events := range spans {
		cases = append(cases, sourceCase{name: fmt.Sprint("spans-", i), signal: "spans", events: events})
	}
	cases = append(cases, sourceCase{name: "spans-unaligned", signal: "spans", names: []string{"exception"}, malformed: true})
	if len(cases) != 20 {
		t.Fatalf("unexpected source campaign size: %d", len(cases))
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			names := append([]string(nil), c.names...)
			data := []map[string]string{}
			for _, event := range c.events {
				names = append(names, "exception")
				data = append(data, attrs(event))
			}
			if c.signal == "logs" {
				names = nil
			}
			q := exploration.Query{Signal: c.signal, Environment: "verification416", FromMS: 1700000000000, ToMS: 1700000060000, Limit: 100}
			calls := 0
			client := &http.Client{Transport: inlineSource{next: http.DefaultTransport, relation: inlineRecord(c.signal, names, data), calls: &calls}, Timeout: 20 * time.Second}
			source := exploration.Source{Endpoint: endpoint, User: user, Password: password, Client: client}
			result, err := source.FetchExceptionGroups(t.Context(), q)
			if calls != 2 {
				t.Fatalf("metadata SQL budget: got%d want2", calls)
			}
			if c.malformed {
				if err == nil {
					t.Fatal("unaligned source event arrays accepted")
				}
				return
			}
			if err != nil {
				t.Fatalf("production SQL/source decoder: %v", err)
			}
			if result.Count != "1" || result.State != "observed" || len(result.Groups) != 1 {
				t.Fatalf("unexpected complete record group")
			}
			group := result.Groups[0]
			want := oracle(c.events)
			checkFingerprint(t, exploration.ExceptionFingerprint{Fingerprint: group.Fingerprint, State: group.State, Reason: group.Reason, Inputs: group.Inputs}, want)
			if group.Count != "1" || group.FirstNS != "1700000000001000000" || group.LastNS != group.FirstNS || group.AffectedCount != "1" || group.AffectedLimited || len(group.Affected) != 1 {
				t.Fatalf("record-time or distinct-context metadata changed: %+v", group)
			}
			identity := group.Affected[0]
			if identity.Service != "synthetic416" || identity.Namespace != nil || identity.Version == nil || *identity.Version != "" {
				t.Fatalf("missing/empty context collapsed: %+v", identity)
			}
		})
	}
}
