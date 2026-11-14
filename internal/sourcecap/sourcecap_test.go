package sourcecap

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/casselc/goo/internal/sourcefailure"
	"github.com/casselc/goo/internal/sourcehttp"
)

func schemaBank(t *testing.T) map[string][]json.RawMessage {
	t.Helper()
	data, err := os.ReadFile("testdata/schema.json")
	if err != nil {
		t.Fatal(err)
	}
	var bank map[string][]json.RawMessage
	if err = json.Unmarshal(data, &bank); err != nil {
		t.Fatal(err)
	}
	return bank
}
func familyKind(sql string) string {
	for _, s := range specs() {
		if strings.Contains(sql, " "+s.table+" ") {
			return s.kind
		}
	}
	return ""
}
func TestProfilesAreFixedIndependentAndDoNotReadObservations(t *testing.T) {
	bank := schemaBank(t)
	var calls []string
	result, err := Inspect(context.Background(), func(_ context.Context, sql string, params map[string]string) ([]json.RawMessage, error) {
		if params != nil {
			t.Fatal("unexpected dynamic parameters")
		}
		calls = append(calls, sql)
		kind := familyKind(sql)
		if kind == "" {
			t.Fatal("unregistered table", sql)
		}
		if strings.HasPrefix(sql, "DESCRIBE TABLE ") {
			return bank[kind], nil
		}
		if !strings.HasPrefix(sql, "SELECT ") || !strings.HasSuffix(sql, " LIMIT 0 FORMAT JSONEachRow") {
			t.Fatal("access probe can scan observations", sql)
		}
		return nil, nil
	})
	if err != nil || len(result) != 6 || len(calls) != 12 {
		t.Fatalf("families=%+v calls=%d err=%v", result, len(calls), err)
	}
	for i, f := range result {
		if !f.Compatible() || f.Kind != specs()[i].kind || f.Table != specs()[i].table || !strings.Contains(f.Detail, "compatible and readable") {
			t.Fatal("profile semantics", f)
		}
	}
}

func TestLogSeverityColumnReadiness435(t *testing.T) {
	for _, typ := range []string{"UInt8", "UInt64", "Nullable(UInt8)", "String", ""} {
		t.Run(typ, func(t *testing.T) {
			bank := schemaBank(t)
			columns := []json.RawMessage{}
			for _, raw := range bank[Logs] {
				var c struct{ Name, Type string }
				if json.Unmarshal(raw, &c) != nil {
					t.Fatal("invalid schema fixture")
				}
				if c.Name == "SeverityNumber" {
					if typ == "" {
						continue
					}
					c.Type = typ
					raw, _ = json.Marshal(map[string]string{"name": c.Name, "type": c.Type})
				}
				columns = append(columns, raw)
			}
			calls := 0
			got, err := Inspect(context.Background(), func(_ context.Context, sql string, _ map[string]string) ([]json.RawMessage, error) {
				calls++
				if strings.HasPrefix(sql, "DESCRIBE") {
					return columns, nil
				}
				if !strings.Contains(sql, "`SeverityNumber`") {
					t.Fatal("access probe omitted intrinsic severity number")
				}
				return nil, nil
			}, Logs)
			if err != nil || len(got) != 1 {
				t.Fatal(got, err)
			}
			if typ == "UInt8" {
				if !got[0].Compatible() || calls != 2 {
					t.Fatal(got, calls)
				}
			} else if got[0].State != "incompatible" || !strings.Contains(got[0].Detail, "SeverityNumber") || calls != 1 {
				t.Fatal(got, calls)
			}
		})
	}
}
func TestSubsetValidationPrecedesReadsAndDoesNotCache(t *testing.T) {
	bank := schemaBank(t)
	reads := 0
	read := func(_ context.Context, sql string, _ map[string]string) ([]json.RawMessage, error) {
		reads++
		if strings.HasPrefix(sql, "DESCRIBE") {
			return bank[Gauge], nil
		}
		return nil, nil
	}
	for _, kinds := range [][]string{{"not-a-family"}, {"logs; DROP TABLE otel_logs"}, {Gauge, Gauge}} {
		if _, err := Inspect(context.Background(), read, kinds...); err == nil || reads != 0 {
			t.Fatal("invalid selector reached source", kinds, reads)
		}
	}
	for range 2 {
		result, err := Inspect(context.Background(), read, Gauge)
		if err != nil || len(result) != 1 || !result[0].Compatible() {
			t.Fatal(result, err)
		}
	}
	if reads != 4 {
		t.Fatal("unexpected caching", reads)
	}
	if _, err := Inspect(context.Background(), nil); err == nil {
		t.Fatal("nil reader accepted")
	}
}
func TestMissingPermissionAndIncompatibleRemainDistinct(t *testing.T) {
	bank := schemaBank(t)
	result, err := Inspect(context.Background(), func(_ context.Context, sql string, _ map[string]string) ([]json.RawMessage, error) {
		kind := familyKind(sql)
		switch kind {
		case Logs:
			return nil, sourcefailure.Wrap(sourcefailure.Unavailable, &sourcehttp.HTTPError{Status: 404, Code: 60})
		case Spans:
			return nil, &sourcehttp.HTTPError{Status: 500, Code: 81}
		case Gauge:
			return nil, &sourcehttp.HTTPError{Status: 403, Code: 497}
		case Counter:
			return nil, fmt.Errorf("Code: 60, unknown table; password secret; SQL server")
		case Histogram:
			return []json.RawMessage{json.RawMessage(`{"name":"TimeUnix","type":"String"}`)}, nil
		default:
			if strings.HasPrefix(sql, "DESCRIBE") {
				return bank[kind], nil
			}
			return nil, nil
		}
	})
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"absent", "absent", "unreadable", "unreadable", "incompatible", "compatible"}
	for i, f := range result {
		if f.State != want[i] || strings.Contains(f.Detail, "secret") || strings.Contains(f.Detail, "SQL server") {
			t.Fatal(f, want[i])
		}
	}
}
func TestAccessProbeDoesNotOverclaimReadPermission(t *testing.T) {
	bank := schemaBank(t)
	for _, failure := range []error{&sourcehttp.HTTPError{Status: 403, Code: 497}, fmt.Errorf("denied SELECT")} {
		calls := 0
		got, err := Inspect(context.Background(), func(_ context.Context, sql string, _ map[string]string) ([]json.RawMessage, error) {
			calls++
			if strings.HasPrefix(sql, "DESCRIBE") {
				return bank[Logs], nil
			}
			return nil, failure
		}, Logs)
		if err != nil || calls != 2 || got[0].State != "unreadable" {
			t.Fatal(got, calls, err)
		}
	}
	// A table disappearing after DESCRIBE is explicitly absent, not observed empty.
	got, err := Inspect(context.Background(), func(_ context.Context, sql string, _ map[string]string) ([]json.RawMessage, error) {
		if strings.HasPrefix(sql, "DESCRIBE") {
			return bank[Logs], nil
		}
		return nil, &sourcehttp.HTTPError{Status: 404, Code: 60}
	}, Logs)
	if err != nil || got[0].State != "absent" {
		t.Fatal(got, err)
	}
}
func TestMalformedAndEmptySchemaAreNeverAbsence(t *testing.T) {
	bank := schemaBank(t)
	cases := [][]json.RawMessage{nil, {json.RawMessage(`{`)}, {json.RawMessage(`{"name":"Timestamp"}`)}, {json.RawMessage(`{"name":"Timestamp","type":null}`)}, {bank[Logs][0], bank[Logs][0]}, make([]json.RawMessage, 257)}
	for _, rows := range cases {
		got, err := Inspect(context.Background(), func(context.Context, string, map[string]string) ([]json.RawMessage, error) { return rows, nil }, Logs)
		if err != nil || got[0].State != "unreadable" {
			t.Fatal(got, err)
		}
	}
	got, err := Inspect(context.Background(), func(_ context.Context, sql string, _ map[string]string) ([]json.RawMessage, error) {
		if strings.HasPrefix(sql, "DESCRIBE") {
			return bank[Logs], nil
		}
		return []json.RawMessage{json.RawMessage(`{"unexpected":"row"}`)}, nil
	}, Logs)
	if err != nil || got[0].State != "unreadable" {
		t.Fatal(got, err)
	}
}
func TestMissingAndWrongTypeAreIncompatible(t *testing.T) {
	bank := schemaBank(t)
	for _, rows := range [][]json.RawMessage{bank[Logs][1:], append([]json.RawMessage{json.RawMessage(`{"name":"Timestamp","type":"DateTime"}`)}, bank[Logs][1:]...)} {
		calls := 0
		got, err := Inspect(context.Background(), func(context.Context, string, map[string]string) ([]json.RawMessage, error) { calls++; return rows, nil }, Logs)
		if err != nil || calls != 1 || got[0].State != "incompatible" || !strings.Contains(got[0].Detail, "Timestamp") {
			t.Fatal(got, calls, err)
		}
	}
}
func TestCancellationReturnsNoPartialInspection(t *testing.T) {
	bank := schemaBank(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	calls := 0
	got, err := Inspect(ctx, func(context.Context, string, map[string]string) ([]json.RawMessage, error) { calls++; return nil, nil })
	if !errors.Is(err, context.Canceled) || got != nil || calls != 0 {
		t.Fatal(got, calls, err)
	}
	ctx, cancel = context.WithCancel(context.Background())
	defer cancel()
	calls = 0
	got, err = Inspect(ctx, func(_ context.Context, sql string, _ map[string]string) ([]json.RawMessage, error) {
		calls++
		if calls == 3 {
			cancel()
		}
		if strings.HasPrefix(sql, "DESCRIBE") {
			return bank[familyKind(sql)], nil
		}
		return nil, nil
	})
	if !errors.Is(err, context.Canceled) || got != nil || calls != 3 {
		t.Fatal(got, calls, err)
	}
	got, err = Inspect(context.Background(), func(context.Context, string, map[string]string) ([]json.RawMessage, error) {
		return nil, fmt.Errorf("wrapped: %w", context.DeadlineExceeded)
	}, Logs)
	if !errors.Is(err, context.DeadlineExceeded) || got != nil {
		t.Fatal(got, err)
	}
}
func TestColumnTypeCompatibility(t *testing.T) {
	for _, c := range []struct {
		raw, shape string
		want       bool
	}{
		{"LowCardinality(String)", "string", true}, {"Map(LowCardinality(String), String)", "map", true}, {"Array(Map(LowCardinality(String), String))", "maps", true},
		{"DateTime64(9, 'UTC')", "time", true}, {"DateTime64(3)", "time", true}, {"Array(DateTime64(9, 'UTC'))", "times", true},
		{"Nullable(String)", "string", false}, {"Map(String,Nullable(String))", "map", false}, {"DateTime64(10)", "time", false}, {"DateTime", "time", false},
		{"Array(DateTime64(9)) trailing", "times", false}, {"LowCardinality(String,UInt8)", "string", false}, {"String); SELECT secret", "string", false},
		{"Bool", "bool", true}, {"Boolean", "bool", true}, {"UInt8", "bool", false}, {"UInt64", "bool", false},
	} {
		if got := matches(c.raw, c.shape); got != c.want {
			t.Errorf("matches(%q,%q)=%v want %v", c.raw, c.shape, got, c.want)
		}
	}
}
