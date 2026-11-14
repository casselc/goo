package verification

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/casselc/goo/internal/sourcecap"
	"github.com/casselc/goo/internal/sourcehttp"
	"github.com/casselc/observability/quintgo/qobs"
	"github.com/casselc/observability/quintgo/validate"
	"hegel.dev/go/hegel"
)

var sourceKinds431 = []string{sourcecap.Logs, sourcecap.Spans, sourcecap.Gauge, sourcecap.Counter, sourcecap.Histogram, sourcecap.ExponentialHistogram}

func sourceSchema431(t *testing.T) map[string][]json.RawMessage {
	t.Helper()
	data, err := os.ReadFile("../internal/sourcecap/testdata/schema.json")
	if err != nil {
		t.Fatal(err)
	}
	var bank map[string][]json.RawMessage
	if err = json.Unmarshal(data, &bank); err != nil {
		t.Fatal(err)
	}
	return bank
}
func sourceKind431(sql string) string {
	tables := map[string]string{"otel_logs": sourcecap.Logs, "otel_traces": sourcecap.Spans, "otel_metrics_gauge": sourcecap.Gauge, "otel_metrics_sum": sourcecap.Counter, "otel_metrics_histogram": sourcecap.Histogram, "otel_metrics_exp_histogram": sourcecap.ExponentialHistogram}
	for table, kind := range tables {
		if strings.Contains(sql, " "+table+" ") {
			return kind
		}
	}
	return ""
}
func TestHegelSourceCapabilities431(t *testing.T) {
	bank := sourceSchema431(t)
	hegel.Test(t, func(tc *hegel.T) {
		outcomes := map[string]string{}
		expected := map[string]string{}
		access := map[string]int{}
		for _, kind := range sourceKinds431 {
			outcome := hegel.Draw(tc, hegel.SampledFrom([]string{"valid", "missing", "denied", "plain-error", "malformed", "wrong-type", "no-access"}))
			outcomes[kind] = outcome
			expected[kind] = map[string]string{"valid": "compatible", "missing": "absent", "denied": "unreadable", "plain-error": "unreadable", "malformed": "unreadable", "wrong-type": "incompatible", "no-access": "unreadable"}[outcome]
		}
		got, err := sourcecap.Inspect(context.Background(), func(_ context.Context, sql string, _ map[string]string) ([]json.RawMessage, error) {
			kind := sourceKind431(sql)
			if kind == "" {
				tc.Fatal("dynamic or unexpected source", sql)
			}
			outcome := outcomes[kind]
			if strings.HasPrefix(sql, "DESCRIBE TABLE ") {
				switch outcome {
				case "missing":
					return nil, &sourcehttp.HTTPError{Status: 404, Code: 60}
				case "denied":
					return nil, &sourcehttp.HTTPError{Status: 403, Code: 497}
				case "plain-error":
					return nil, errors.New("Code:60 password=secret SELECT raw credential")
				case "malformed":
					return []json.RawMessage{json.RawMessage(`{"name":"TimeUnix"}`)}, nil
				case "wrong-type":
					rows := append([]json.RawMessage(nil), bank[kind]...)
					var first map[string]any
					if e := json.Unmarshal(rows[0], &first); e != nil {
						tc.Fatal(e)
					}
					first["type"] = "Nullable(String)"
					rows[0], _ = json.Marshal(first)
					return rows, nil
				default:
					return bank[kind], nil
				}
			}
			if !strings.HasPrefix(sql, "SELECT ") || !strings.HasSuffix(sql, " LIMIT 0 FORMAT JSONEachRow") {
				tc.Fatal("observation read during readiness", sql)
			}
			access[kind]++
			if outcome == "no-access" {
				return nil, &sourcehttp.HTTPError{Status: 403, Code: 497}
			}
			return nil, nil
		})
		if err != nil || len(got) != len(sourceKinds431) {
			tc.Fatal("independent capability failure", err, got)
		}
		for i, family := range got {
			if family.Kind != sourceKinds431[i] || family.State != expected[family.Kind] || family.Compatible() != (outcomes[family.Kind] == "valid") {
				tc.Fatal("family outcome changed", family, outcomes)
			}
			wantAccess := 0
			if outcomes[family.Kind] == "valid" || outcomes[family.Kind] == "no-access" {
				wantAccess = 1
			}
			if access[family.Kind] != wantAccess {
				tc.Fatal("access before valid metadata", family, access)
			}
			for _, secret := range []string{"password=", "secret", "SELECT raw", "credential"} {
				if strings.Contains(family.Detail, secret) {
					tc.Fatal("raw error escaped", family)
				}
			}
		}
		// Cancellation never turns the completed prefix into a successful inspection.
		cancelAt := hegel.Draw(tc, hegel.Integers(1, 4))
		calls := 0
		ctx, cancel := context.WithCancel(context.Background())
		result, err := sourcecap.Inspect(ctx, func(_ context.Context, sql string, _ map[string]string) ([]json.RawMessage, error) {
			calls++
			if calls == cancelAt {
				cancel()
			}
			if strings.HasPrefix(sql, "DESCRIBE TABLE ") {
				return bank[sourceKind431(sql)], nil
			}
			return nil, nil
		})
		cancel()
		if !errors.Is(err, context.Canceled) || result != nil || calls != cancelAt {
			tc.Fatal("cancelled prefix admitted", err, result, calls, cancelAt)
		}
	})
}

func TestQuintObservedSourceCapabilities431(t *testing.T) {
	quint, err := exec.LookPath("quint")
	if err != nil {
		t.Fatal("native readiness trace validation requires quint on PATH")
	}
	spec, err := filepath.Abs("model/source_capabilities_431.qnt")
	if err != nil {
		t.Fatal(err)
	}
	manifest := "spec: " + spec + "\nmodule: source_capabilities_431\nconstants: {INITIAL: '0'}\ninvariants: [safety]\nactions:\n"
	for _, action := range []string{"metadata_ok", "metadata_missing", "metadata_refused", "access_ok", "access_missing", "access_refused"} {
		manifest += "  " + action + ":\n    expect: 'true'\n"
	}
	for _, action := range []string{"publish_compatible", "publish_absent", "publish_incompatible", "publish_unreadable"} {
		manifest += "  " + action + ":\n    expect: 'canUse == {{.obs.compatible}} and missing == {{.obs.absent}}'\n"
	}
	dir := t.TempDir()
	config := filepath.Join(dir, "binding.yaml")
	if err = os.WriteFile(config, []byte(manifest), 0600); err != nil {
		t.Fatal(err)
	}
	bank := sourceSchema431(t)
	for _, scenario := range []string{"valid", "missing", "denied", "incompatible", "access-denied", "access-missing", "malformed"} {
		t.Run(scenario, func(t *testing.T) {
			for _, corrupt := range []bool{false, true} {
				sink := &qobs.MemorySink{}
				rec := qobs.NewRecorder(sink)
				result, err := sourcecap.Inspect(context.Background(), func(_ context.Context, sql string, _ map[string]string) ([]json.RawMessage, error) {
					if strings.HasPrefix(sql, "DESCRIBE TABLE ") {
						if scenario == "missing" {
							rec.Record(context.Background(), "metadata_missing", nil)
							return nil, &sourcehttp.HTTPError{Status: 404, Code: 60}
						}
						if scenario == "denied" {
							rec.Record(context.Background(), "metadata_refused", nil)
							return nil, &sourcehttp.HTTPError{Status: 403, Code: 497}
						}
						rec.Record(context.Background(), "metadata_ok", nil)
						if scenario == "malformed" {
							return nil, nil
						}
						if scenario == "incompatible" {
							return bank[sourcecap.Logs][1:], nil
						}
						return bank[sourcecap.Logs], nil
					}
					if scenario == "access-denied" {
						rec.Record(context.Background(), "access_refused", nil)
						return nil, &sourcehttp.HTTPError{Status: 403, Code: 497}
					}
					if scenario == "access-missing" {
						rec.Record(context.Background(), "access_missing", nil)
						return nil, &sourcehttp.HTTPError{Status: 404, Code: 60}
					}
					rec.Record(context.Background(), "access_ok", nil)
					return nil, nil
				}, sourcecap.Logs)
				if err != nil || len(result) != 1 {
					t.Fatal(result, err)
				}
				actual := result[0]
				if corrupt {
					if actual.Compatible() {
						actual.State = "absent"
					} else {
						actual.State = "compatible"
					}
				}
				rec.Record(context.Background(), "publish_"+actual.State, nil, "obs.compatible", actual.Compatible(), "obs.absent", actual.State == "absent")
				checker, err := validate.NewChecker(config, validate.Options{Quint: quint, Backend: "typescript", Dir: t.TempDir()})
				if err != nil {
					t.Fatal(err)
				}
				checked, err := checker.Check(context.Background(), sink.Steps())
				if err != nil {
					t.Fatal(err)
				}
				if checked.OK() == corrupt {
					t.Fatal(fmt.Sprintf("actual inspection trace acceptance changed: scenario=%s corrupt=%v accepted=%v", scenario, corrupt, checked.OK()))
				}
			}
		})
	}
}
