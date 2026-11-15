package verification

import (
	"bytes"
	"encoding/json"
	"math"
	"math/big"
	"os"
	"reflect"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/casselc/goo/measurements"
	"github.com/casselc/observability/quintgo/qobs"
	"github.com/casselc/observability/quintgo/validate"
)

type histogram447Step struct {
	Name                                           string
	Local                                          bool
	SourceBefore, SourceAfter, SQLBefore, SQLAfter *int
	Posts                                          []string
}
type histogram447Bank struct {
	Name, ReadAt, Find string
	Keys               []string
	Detail             *string
}
type histogram447Report struct {
	Version, Status          string
	Cleanup, DatabaseCleanup bool
	Errors                   []string
	Fixture                  struct {
		Token, Environment, Service, Scope, Database string
		Clock                                        int64
	}
	Raw    []json.RawMessage
	Calls  []pair428Call
	Source []struct {
		SQL    string
		Status *int
	}
	Acquisitions []struct {
		Name              string
		Request, Response json.RawMessage
		Status            int
	}
	Steps    []histogram447Step
	Banks    []histogram447Bank
	Rejected []struct {
		Name, Path string
		Request    json.RawMessage
		Status     int
		SQL        *int
	}
	ResponseHold struct {
		Status         int
		ReadAt         string
		FocusPreserved bool
	}
	Bombadil struct {
		Snapshots int
		Completed struct {
			Version                                       string
			Phase, Source, Forbidden, BucketRows          int
			ReadFreeLocal, HeldRestored, Detail, Overflow bool
		}
	}
}
type histogram447Oracle struct {
	Resource, Attributes       map[string]string
	ResourceSchemaURL          string `json:"resource_schema_url"`
	Scope, Version, Name, Unit string
	SchemaURL                  string `json:"schema_url"`
	StartNS                    string `json:"start_ns"`
	EndNS                      string `json:"end_ns"`
	Flags                      uint32
	Temporality                int32
	Count                      string
	Bounds                     []string
	BucketCounts               []string `json:"bucket_counts"`
	Sum, Min, Max              json.RawMessage
}

func histogram447JSONEqual(a, b json.RawMessage) bool {
	var x, y any
	return json.Unmarshal(a, &x) == nil && json.Unmarshal(b, &y) == nil && reflect.DeepEqual(x, y)
}
func histogram447BucketState(raw histogram447Oracle) string {
	if raw.Flags&1 != 0 {
		return "absent"
	}
	count, ok := new(big.Int).SetString(raw.Count, 10)
	if !ok || count.Sign() < 0 || count.BitLen() > 64 || count.String() != raw.Count {
		return "invalid"
	}
	previous := math.Inf(-1)
	for _, text := range raw.Bounds {
		n, err := strconv.ParseFloat(text, 64)
		if err != nil || math.IsNaN(n) || math.IsInf(n, 0) || n <= previous {
			return "invalid"
		}
		previous = n
	}
	sum := new(big.Int)
	for _, text := range raw.BucketCounts {
		n, ok := new(big.Int).SetString(text, 10)
		if !ok || n.Sign() < 0 || n.BitLen() > 64 || n.String() != text {
			return "invalid"
		}
		sum.Add(sum, n)
	}
	if len(raw.Bounds) == 0 && len(raw.BucketCounts) == 0 {
		return "absent"
	}
	if len(raw.BucketCounts) != len(raw.Bounds)+1 || sum.Cmp(count) != 0 {
		return "invalid"
	}
	return "valid"
}

var histogram447Cancel = regexp.MustCompile(`^/api/queries/[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}/cancel$`)

func histogram447Admission(r histogram447Report) string {
	if r.Version != "goo.rawHistogram447/1" || r.Status != "passed" || !r.Cleanup || !r.DatabaseCleanup || len(r.Errors) != 0 {
		return "unfinished run"
	}
	if r.Fixture.Token == "" || r.Fixture.Environment == "" || r.Fixture.Service == "" || len(r.Raw) == 0 || len(r.Steps) == 0 || len(r.Acquisitions) == 0 {
		return "missing source witness"
	}
	var oracle []histogram447Oracle
	for _, encoded := range r.Raw {
		var row histogram447Oracle
		if json.Unmarshal(encoded, &row) != nil || row.Resource["fixture.bank"] != r.Fixture.Token || row.Resource["service.name"] != r.Fixture.Service {
			return "source oracle differs from fixture"
		}
		oracle = append(oracle, row)
	}
	sourcePaths := map[string]bool{"/api/metric-discovery": true, "/api/histogram-observations": true, "/api/metric-scopes": true, "/api/measurements": true}
	required := map[string]bool{"read": false, "page": false, "find": false, "inspect": false, "leave": false, "return": false}
	lastSource, lastSQL, retirements := 0, 0, 0
	var expectedPosts []string
	for _, step := range r.Steps {
		if step.SourceBefore == nil || step.SourceAfter == nil || step.SQLBefore == nil || step.SQLAfter == nil || *step.SourceBefore != lastSource || *step.SQLBefore != lastSQL || *step.SourceAfter < lastSource || *step.SQLAfter < lastSQL {
			return "missing or discontinuous authority counters"
		}
		posts, canceled := 0, 0
		expectedPosts = append(expectedPosts, step.Posts...)
		for _, path := range step.Posts {
			if sourcePaths[path] {
				posts++
			} else if histogram447Cancel.MatchString(path) && step.Name == "pending-departure" && !step.Local {
				canceled++
				retirements++
			} else if path != "/api/cel/context" && path != "/api/cel/plan" {
				return "raw interaction acquired store or unrelated authority"
			}
		}
		if step.Name == "pending-departure" && (canceled != 1 || len(step.Posts) != 2 || posts != 1 || step.Posts[0] != "/api/histogram-observations" || *step.SQLAfter-*step.SQLBefore != 3) {
			return "retirement is not the completed owned read"
		}
		if posts != *step.SourceAfter-*step.SourceBefore {
			return "source posts disagree with observed counters"
		}
		if step.Local && (*step.SourceAfter != *step.SourceBefore || *step.SQLAfter != *step.SQLBefore || len(step.Posts) != 0) {
			return "held action performed IO"
		}
		if _, ok := required[step.Name]; ok {
			required[step.Name] = true
			if step.Name != "read" && !step.Local {
				return "local scenario not admitted as local"
			}
		}
		lastSource, lastSQL = *step.SourceAfter, *step.SQLAfter
	}
	for name, seen := range required {
		if !seen {
			return "required interaction absent: " + name
		}
	}
	calls, cancels := 0, 0
	var actualPosts []string
	for _, call := range r.Calls {
		if call.Method == "POST" {
			actualPosts = append(actualPosts, call.Path)
			if sourcePaths[call.Path] {
				calls++
			} else if histogram447Cancel.MatchString(call.Path) {
				cancels++
			} else if call.Path != "/api/cel/context" && call.Path != "/api/cel/plan" {
				return "unexpected source/store POST"
			}
		}
	}
	if !slices.Equal(actualPosts, expectedPosts) || cancels != retirements || retirements != 1 {
		return "retirement or HTTP ledger lost ownership"
	}
	if calls != lastSource || len(r.Source) < lastSQL {
		return "authority ledger lost work"
	}
	for _, call := range r.Source {
		if call.Status == nil || !(strings.HasPrefix(call.SQL, "SELECT ") || strings.HasPrefix(call.SQL, "DESCRIBE TABLE ")) {
			return "source SQL is missing or not readonly"
		}
	}
	witnessed := 0
	readsByName := map[string]string{}
	for _, acq := range r.Acquisitions {
		var q measurements.RawHistogramQuestion
		if json.Unmarshal(acq.Request, &q) != nil || q.Validate() != nil {
			return "invalid acquisition request"
		}
		seen := false
		for _, call := range r.Calls {
			if call.Method == "POST" && call.Path == "/api/histogram-observations" && histogram447JSONEqual(call.Body, acq.Request) {
				seen = true
				break
			}
		}
		if !seen {
			return "acquisition missing from HTTP ledger"
		}
		if acq.Status != 200 {
			var fail struct {
				Rows    json.RawMessage
				Refusal struct{ Kind string }
			}
			if acq.Status < 400 || json.Unmarshal(acq.Response, &fail) != nil || len(fail.Rows) != 0 || fail.Refusal.Kind == "" {
				return "failed read exposed partial observations"
			}
			continue
		}
		var result measurements.RawHistogramResult
		decoder := json.NewDecoder(bytes.NewReader(acq.Response))
		decoder.DisallowUnknownFields()
		if decoder.Decode(&result) != nil {
			return "invalid result profile"
		}
		if result.Version != "goo/raw-histogram-observations/v1" || result.Limit != 10000 || !reflect.DeepEqual(result.Question, q.Question) || result.Instrument != q.Instrument {
			return "profile echo changed"
		}
		at, err := time.Parse(time.RFC3339Nano, result.ReadAt)
		if err != nil || !strings.HasSuffix(result.ReadAt, "Z") || at.UnixMilli() < q.Question.ToMS {
			return "read and event clocks confused"
		}
		readsByName[acq.Name] = result.ReadAt
		selection, _ := q.Question.EnvironmentSelection()
		expected := []int{}
		for index, raw := range oracle {
			ns, err := strconv.ParseInt(raw.EndNS, 10, 64)
			if err == nil && ns >= q.Question.FromMS*1000000 && ns < q.Question.ToMS*1000000 && selection.Matches(raw.Resource) && q.Question.Namespace.Matches(raw.Resource) && raw.Resource["service.name"] == q.Question.Service && raw.Scope == q.Instrument.Scope && raw.Version == q.Instrument.Version && raw.SchemaURL == q.Instrument.SchemaURL && raw.Name == q.Instrument.Metric && raw.Unit == q.Instrument.Unit && raw.Temporality == q.Instrument.Temporality {
				expected = append(expected, index)
			}
		}
		if len(result.Rows) != len(expected) {
			return "source multiplicity changed"
		}
		keys := map[string]bool{}
		for _, row := range result.Rows {
			match := slices.IndexFunc(expected, func(i int) bool { return histogram447JSONEqual(row.RawRow, r.Raw[i]) })
			if match < 0 {
				return "raw occurrence differs from independent source"
			}
			raw := oracle[expected[match]]
			expected = append(expected[:match], expected[match+1:]...)
			if row.OccurrenceKey == "" || keys[row.OccurrenceKey] || row.EndNS != raw.EndNS || row.StartNS == nil || *row.StartNS != raw.StartNS || row.Temporality != raw.Temporality || row.Flags == nil || *row.Flags != raw.Flags || !reflect.DeepEqual(row.Stream.Resource, raw.Resource) || !reflect.DeepEqual(row.Stream.Attributes, raw.Attributes) {
				return "point identity or occurrence key changed"
			}
			keys[row.OccurrenceKey] = true
			state := "recorded"
			if raw.Flags&1 != 0 {
				state = "no-recorded-value"
			} else if raw.Flags != 0 {
				state = "unknown"
			}
			if row.RecordingState != state || row.BucketsState != histogram447BucketState(raw) {
				return "recording or bucket semantics changed"
			}
			if state == "no-recorded-value" {
				if row.Count != nil || row.Bounds != nil || row.BucketCounts != nil {
					return "not-recorded population promoted"
				}
			} else if row.Count == nil || *row.Count != raw.Count || !reflect.DeepEqual(row.BucketCounts, raw.BucketCounts) {
				return "exact counts changed"
			}
			for _, scalar := range []measurements.RawHistogramScalar{row.Sum, row.Min, row.Max} {
				if scalar.PresenceState != "unavailable" {
					return "optional source presence invented"
				}
				if state == "no-recorded-value" && scalar.Value != nil {
					return "not-recorded statistic promoted"
				}
			}
		}
		witnessed++
	}
	if witnessed < 3 {
		return "missing separate instrument acquisitions"
	}
	var before, after *histogram447Bank
	for index := range r.Banks {
		bank := &r.Banks[index]
		if bank.Name == "before-return" {
			before = bank
		}
		if bank.Name == "after-return" {
			after = bank
		}
	}
	if before == nil || after == nil || before.ReadAt == "" || before.ReadAt != readsByName["main"] || before.Detail == nil || *before.Detail == "" || len(before.Keys) == 0 {
		return "missing held bank"
	}
	a, b := *before, *after
	a.Name = ""
	b.Name = ""
	if !reflect.DeepEqual(a, b) {
		return "return changed held observations"
	}
	for _, bank := range r.Banks {
		if bank.Name == "after-refusal" || bank.Name == "after-late" {
			held := bank
			held.Name = ""
			if !reflect.DeepEqual(a, held) {
				return "retired or refused read replaced held bank"
			}
		}
	}
	if r.ResponseHold.Status != 200 || !r.ResponseHold.FocusPreserved || r.ResponseHold.ReadAt == "" || r.ResponseHold.ReadAt == before.ReadAt {
		return "late transport completion not independently observed"
	}
	if len(r.Rejected) < 3 {
		return "missing typed refusal controls"
	}
	for _, refused := range r.Rejected {
		var q measurements.RawHistogramQuestion
		if refused.Path != "/api/histogram-observations" || refused.Status != 400 || refused.SQL == nil || *refused.SQL != 0 || json.Unmarshal(refused.Request, &q) == nil && q.Validate() == nil {
			return "unsupported request gained source authority"
		}
	}
	actor := r.Bombadil.Completed
	if r.Bombadil.Snapshots == 0 || actor.Version != "goo.rawHistogram447Actor/1" || actor.Phase != 8 || actor.Source != 1 || actor.Forbidden != 0 || !actor.ReadFreeLocal || !actor.HeldRestored || !actor.Detail || actor.BucketRows == 0 || actor.Overflow {
		return "actor did not finish held histogram journey"
	}
	return ""
}
func TestQuintRawHistogram447Observed(t *testing.T) {
	file := os.Getenv("GOO_RAW_HISTOGRAM447_REPORT")
	if file == "" {
		t.Skip("requires completed actual raw-histogram browser report")
	}
	raw, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	var report histogram447Report
	if err = json.Unmarshal(raw, &report); err != nil {
		t.Fatal(err)
	}
	if reason := histogram447Admission(report); reason != "" {
		t.Fatal("actual histogram admission:", reason)
	}
	checker, err := validate.NewChecker("model/population_drilldown.yaml", validate.Options{Quint: quintPath(t), Backend: "typescript", Dir: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	check := func(r histogram447Report) bool {
		coherent := histogram447Admission(r) == ""
		sink := &qobs.MemorySink{}
		rec := qobs.NewRecorder(sink)
		reads := 0
		for _, step := range r.Steps {
			count := 0
			if step.SQLBefore != nil && step.SQLAfter != nil {
				count = *step.SQLAfter - *step.SQLBefore
			}
			if count == 0 {
				rec.Record(t.Context(), "history", nil, "obs.reads", reads, "obs.depth", 0, "obs.coherent", coherent)
			}
			for range max(0, count) {
				reads++
				rec.Record(t.Context(), "read", nil, "obs.reads", reads, "obs.depth", 0, "obs.coherent", coherent)
			}
		}
		for reads < len(r.Source) {
			reads++
			rec.Record(t.Context(), "read", nil, "obs.reads", reads, "obs.depth", 0, "obs.coherent", coherent)
		}
		verdict, err := checker.Check(t.Context(), sink.Steps())
		if err != nil {
			t.Fatal(err)
		}
		return verdict.OK()
	}
	if !check(report) {
		t.Fatal("actual held acquisition trace rejected")
	}
	for _, mode := range []string{"cleanup", "source-bank", "scope", "exact-count", "presence", "return-keys", "return-read", "missing-counter", "actor", "cancel-wrong-phase", "late-bank"} {
		t.Run(mode, func(t *testing.T) {
			var bad histogram447Report
			json.Unmarshal(raw, &bad)
			switch mode {
			case "cleanup":
				bad.Cleanup = false
			case "source-bank":
				bad.Raw = bad.Raw[1:]
			case "scope":
				var result map[string]any
				json.Unmarshal(bad.Acquisitions[0].Response, &result)
				result["instrument"].(map[string]any)["unit"] = "changed"
				bad.Acquisitions[0].Response, _ = json.Marshal(result)
			case "exact-count":
				var result map[string]any
				json.Unmarshal(bad.Acquisitions[0].Response, &result)
				result["rows"].([]any)[0].(map[string]any)["count"] = "9"
				bad.Acquisitions[0].Response, _ = json.Marshal(result)
			case "presence":
				var result map[string]any
				json.Unmarshal(bad.Acquisitions[0].Response, &result)
				result["rows"].([]any)[0].(map[string]any)["sum"].(map[string]any)["presenceState"] = "present"
				bad.Acquisitions[0].Response, _ = json.Marshal(result)
			case "return-keys":
				for i := range bad.Banks {
					if bad.Banks[i].Name == "after-return" {
						bad.Banks[i].Keys[0] = "invented"
					}
				}
			case "return-read":
				for i := range bad.Steps {
					if bad.Steps[i].Name == "return" {
						n := *bad.Steps[i].SourceAfter + 1
						bad.Steps[i].SourceAfter = &n
					}
				}
			case "missing-counter":
				bad.Steps[0].SQLBefore = nil
			case "cancel-wrong-phase":
				for i := range bad.Steps {
					if bad.Steps[i].Name == "find" {
						bad.Steps[i].Posts = append(bad.Steps[i].Posts, "/api/queries/9b0fb668-203f-4958-9a8e-f479092fbddf/cancel")
					}
				}
			case "late-bank":
				for i := range bad.Banks {
					if bad.Banks[i].Name == "after-late" {
						bad.Banks[i].Keys[0] = "obsolete"
					}
				}
			case "actor":
				bad.Bombadil.Completed.Source++
			}
			if check(bad) {
				t.Fatal("histogram negative control accepted", mode)
			}
		})
	}
	t.Logf("admitted %d interactions, %d source rows, %d acquisitions and eleven rejecting controls", len(report.Steps), len(report.Raw), len(report.Acquisitions))
}
