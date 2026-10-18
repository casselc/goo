package lifecycle

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/casselc/goo/measurement"
)

type fixture struct {
	Name           string           `json:"name"`
	Policy         map[string]any   `json:"policy"`
	Events         []map[string]any `json:"events"`
	Counts         map[string]int   `json:"counts"`
	State          string           `json:"state"`
	SubjectCount   int              `json:"subjectCount"`
	Ratio          *float64         `json:"ratio"`
	Subjects       map[string][]any `json:"subjects"`
	Duplicates     int              `json:"duplicates"`
	EventConflicts int              `json:"eventConflicts"`
	Invalid        int              `json:"invalid"`
	AfterCutoff    int              `json:"afterCutoff"`
	OutsideScope   int              `json:"outsideScope"`
}

func fixtures(t *testing.T) (map[string]any, map[string]any, []fixture) {
	t.Helper()
	b, err := os.ReadFile("testdata/cases.json")
	if err != nil {
		t.Fatal(err)
	}
	var all struct {
		Policy   map[string]any `json:"policy"`
		Defaults map[string]any `json:"eventDefaults"`
		Cases    []fixture      `json:"cases"`
	}
	if err := json.Unmarshal(b, &all); err != nil {
		t.Fatal(err)
	}
	return all.Policy, all.Defaults, all.Cases
}

func policyFor(t *testing.T, base, override map[string]any) Policy {
	t.Helper()
	m := map[string]any{}
	for k, v := range base {
		m[k] = v
	}
	for k, v := range override {
		m[k] = v
	}
	b, _ := json.Marshal(m)
	var p Policy
	if err := json.Unmarshal(b, &p); err != nil {
		t.Fatal(err)
	}
	return p
}

func eventsFor(defaults map[string]any, recipes []map[string]any) []json.RawMessage {
	out := []json.RawMessage{}
	for _, recipe := range recipes {
		repeat := 1
		if n, ok := recipe["repeat"].(float64); ok {
			repeat = int(n)
		}
		for i := 0; i < repeat; i++ {
			e := map[string]any{}
			for k, v := range defaults {
				e[k] = v
			}
			for k, v := range recipe {
				if k == "repeat" {
					continue
				}
				if s, ok := v.(string); ok {
					v = strings.ReplaceAll(s, "{i}", strconv.Itoa(i))
				}
				e[k] = v
			}
			b, _ := json.Marshal(e)
			out = append(out, b)
		}
	}
	return out
}

func definition(t *testing.T) *measurement.Definition {
	t.Helper()
	f, err := os.Open("../examples/lifecycle/measurement.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	d, err := measurement.Read(f)
	if err != nil {
		t.Fatal(err)
	}
	return d
}

func TestIndependentLifecycleFixtures(t *testing.T) {
	base, defaults, cases := fixtures(t)
	d := definition(t)
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			p := policyFor(t, base, c.Policy)
			events := eventsFor(defaults, c.Events)
			prepared, err := p.PrepareRaw(events)
			if err != nil {
				t.Fatal(err)
			}
			r := prepared.Receipt
			if r.CohortStartNS != "57" || r.CohortEndNS != "87" || r.State != c.State || !reflect.DeepEqual(r.Counts, c.Counts) {
				t.Fatalf("unexpected receipt: %+v; expected state %s, counts %v", r, c.State, c.Counts)
			}
			if r.DuplicateDeliveries != c.Duplicates || r.ConflictingEventIDs != c.EventConflicts || r.InvalidRecords != c.Invalid || r.AfterCutoff != c.AfterCutoff || r.OutsideScope != c.OutsideScope {
				t.Fatalf("unexpected delivery/quality diagnostics: %+v", r)
			}
			if len(prepared.Subjects) != c.SubjectCount {
				t.Fatalf("subjects %d, want %d", len(prepared.Subjects), c.SubjectCount)
			}
			if c.Subjects != nil {
				actual := map[string][]any{}
				for _, s := range prepared.Subjects {
					actual[s.TenantID+"/"+s.SubjectID] = []any{s.Classification, s.Reason, s.InCohort, s.IsMature}
					if s.Reason == "conflicting_acceptance" && (s.AcceptedTimeNS != "" || s.Release != "") {
						t.Fatal("ambiguous acceptance must not select an arbitrary timestamp/release")
					}
				}
				if !reflect.DeepEqual(actual, c.Subjects) {
					t.Fatalf("subjects: %v; want %v", actual, c.Subjects)
				}
			}
			result, err := prepared.Evaluate(d, nil)
			if err != nil {
				t.Fatal(err)
			}
			m := result.Results[0].Measures
			for _, name := range []string{"eligible", "good", "bad", "unknown"} {
				if *m[name].Count != uint64(c.Counts[name]) {
					t.Fatalf("aggregate %s does not match independently expected count", name)
				}
			}
			v := m["success"]
			if v.State != c.State || (v.Ratio == nil) != (c.Ratio == nil) || c.Ratio != nil && math.Abs(*v.Ratio-*c.Ratio) > 1e-12 {
				t.Fatalf("ratio/state %+v; want %s / %v", v, c.State, c.Ratio)
			}
			// Preparation is order independent, including its receipt digest.
			for i, j := 0, len(events)-1; i < j; i, j = i+1, j-1 {
				events[i], events[j] = events[j], events[i]
			}
			reordered, err := p.PrepareRaw(events)
			if err != nil || encoded(reordered) != encoded(prepared) {
				t.Fatal("event ordering changed the prepared snapshot")
			}
		})
	}
}

func TestExactEpochNanosecondsAndZero(t *testing.T) {
	base, defaults, _ := fixtures(t)
	for _, epoch := range []int64{0, 1700000000000000000, math.MaxInt64 - 100} {
		t.Run(strconv.FormatInt(epoch, 10), func(t *testing.T) {
			p := policyFor(t, base, map[string]any{"evaluationTimeNs": fmt.Sprint(epoch + 100), "evidenceCutoffNs": fmt.Sprint(epoch + 100), "windowNs": "100"})
			events := []json.RawMessage{}
			for _, item := range []struct {
				id               string
				accept, terminal int64
			}{{"at", 0, 10}, {"after", 1, 12}} {
				for _, kind := range []string{"accepted", "terminal"} {
					e := map[string]any{}
					for k, v := range defaults {
						e[k] = v
					}
					e["event_id"], e["kind"], e["subject_id"] = item.id+kind, kind, item.id
					e["received_time_ns"] = fmt.Sprint(epoch + 99)
					offset := item.accept
					if kind == "terminal" {
						offset = item.terminal
						e["outcome"] = "confirmed"
					}
					e["event_time_ns"] = fmt.Sprint(epoch + offset)
					b, _ := json.Marshal(e)
					events = append(events, b)
				}
			}
			prepared, err := p.PrepareRaw(events)
			if err != nil || prepared.Receipt.Counts["good"] != 1 || prepared.Receipt.Counts["bad"] != 1 {
				t.Fatalf("one-tick boundary or epoch zero lost: %v, %v", prepared, err)
			}
		})
	}
}

func TestInvalidPolicyAndRecords(t *testing.T) {
	base, defaults, _ := fixtures(t)
	for _, patch := range []map[string]any{
		{"deadlineNs": "0"}, {"allowanceNs": "-1"}, {"windowNs": "0"},
		{"evaluationTimeNs": "-9223372036854775808", "evidenceCutoffNs": "-9223372036854775808"},
		{"evidenceCutoffNs": "101"}, {"deadlineNs": "1.5"}, {"subjectField": "evaluation_id"},
		{"scopes": []Scope{{Environment: "test", TenantID: "a", ServiceName: "reservations", Coverage: "complete"}}},
		{"successOutcome": "not-declared"}, {"scopes": []Scope{}},
	} {
		p := policyFor(t, base, patch)
		if _, err := p.Prepare(nil); err == nil {
			t.Fatalf("invalid policy accepted: %v", patch)
		}
	}
	p := policyFor(t, base, nil)
	for _, field := range []string{"event_time_ns", "received_time_ns"} {
		events := eventsFor(defaults, []map[string]any{{"event_id": "a", "kind": "accepted", "subject_id": "x", "event_time_ns": "60", field: float64(60)}})
		prepared, err := p.PrepareRaw(events)
		if err != nil || prepared.Receipt.InvalidRecords != 1 || prepared.Receipt.State != "unknown" {
			t.Fatalf("numeric timestamp should be diagnosed as invalid, not rounded or zeroed: %v", err)
		}
	}
	for _, record := range []string{`null`, `{"event_id":"x"}`, `{"unknown":true}`} {
		prepared, err := p.PrepareRaw([]json.RawMessage{json.RawMessage(record)})
		if err != nil || prepared.Receipt.InvalidRecords != 1 || prepared.Receipt.State != "unknown" {
			t.Fatal("malformed unscoped record was silently discarded")
		}
	}
	if _, err := p.PrepareRaw([]json.RawMessage{json.RawMessage(`{} {}`)}); err == nil {
		t.Fatal("multiple JSON values accepted")
	}
	b, _ := os.ReadFile("../examples/lifecycle/policy.yaml")
	for _, suffix := range []string{"\nunknown: true\n", "\nevaluationId: duplicate\n", "\n---\n{}\n"} {
		if _, err := ReadPolicy(bytes.NewReader(append(append([]byte{}, b...), []byte(suffix)...))); err == nil {
			t.Fatal("invalid policy YAML accepted")
		}
	}
}

func TestCorrectionDoesNotRewriteOriginalAndMixedSnapshotsFail(t *testing.T) {
	base, defaults, cases := fixtures(t)
	var original, corrected *Prepared
	for _, c := range cases {
		if c.Name == "original_snapshot_before_late_delivery" || c.Name == "corrected_snapshot_after_late_delivery" {
			p := policyFor(t, base, c.Policy)
			prepared, err := p.PrepareRaw(eventsFor(defaults, c.Events))
			if err != nil {
				t.Fatal(err)
			}
			if c.Name == "original_snapshot_before_late_delivery" {
				original = prepared
			} else {
				corrected = prepared
			}
		}
	}
	if original == nil || corrected == nil || original.Receipt.SnapshotDigest == corrected.Receipt.SnapshotDigest || original.Subjects[0].Classification != "bad" || corrected.Subjects[0].Classification != "good" {
		t.Fatal("correction must retain independently replayable original and corrected snapshots")
	}
	if err := corrected.Receipt.Policy.ValidateCorrection(original.Receipt); err != nil {
		t.Fatal(err)
	}
	for _, kind := range []string{"clock", "question", "basis", "scope"} {
		p := corrected.Receipt.Policy
		switch kind {
		case "clock":
			p.CustodyClock = "another-clock"
		case "question":
			p.WindowNS = "31"
		case "basis":
			p.EvidenceBasis = original.Receipt.Policy.EvidenceBasis
		case "scope":
			p.Scopes = append([]Scope(nil), p.Scopes...)
			p.Scopes[0].TenantID = "another-tenant"
		}
		if err := p.ValidateCorrection(original.Receipt); err == nil {
			t.Fatalf("invalid %s correction accepted", kind)
		}
	}
	corrected.Facts = append(corrected.Facts, original.Facts...)
	if _, err := corrected.Evaluate(definition(t), nil); err == nil {
		t.Fatal("mixed snapshot facts accepted")
	}
}

func TestExclusiveBasisAndVisibleEvidenceIdentity(t *testing.T) {
	base, defaults, _ := fixtures(t)
	p := policyFor(t, base, map[string]any{"evidenceCutoffNs": "95"})
	events := eventsFor(defaults, []map[string]any{
		{"event_id": "a", "kind": "accepted", "subject_id": "x", "event_time_ns": "60", "received_time_ns": "60"},
		{"event_id": "t", "kind": "terminal", "subject_id": "x", "event_time_ns": "70", "received_time_ns": "95", "outcome": "confirmed"},
	})
	old, err := p.PrepareRaw(events)
	if err != nil {
		t.Fatal(err)
	}
	if old.Receipt.Counts["bad"] != 1 || old.Receipt.AfterCutoff != 1 {
		t.Fatal("receipt exactly at basis must be excluded")
	}
	p.EvidenceCutoffNS = "96"
	newer, err := p.PrepareRaw(events)
	if err != nil || newer.Receipt.Counts["good"] != 1 {
		t.Fatal("receipt must become visible above its custody time")
	}
	p.EvidenceCutoffNS = "95"
	extra := eventsFor(defaults, []map[string]any{{"event_id": "future", "kind": "retry", "subject_id": "x", "event_time_ns": "70", "received_time_ns": "100"}})
	appended, err := p.PrepareRaw(append(events, extra...))
	if err != nil {
		t.Fatal(err)
	}
	if old.Receipt.SnapshotDigest != appended.Receipt.SnapshotDigest || old.Receipt.EvidenceDigest != appended.Receipt.EvidenceDigest || old.Receipt.InputManifestDigest == appended.Receipt.InputManifestDigest {
		t.Fatal("historical visible identity must be independent of invisible source additions")
	}
}

func TestLaterKnowledgeRequiresExplicitPolicy(t *testing.T) {
	// Reuse the repository's policy fixture rather than bypassing validation.
	p := Policy{APIVersion: "goo/lifecycle/v1alpha1", EvaluationID: "e1", DefinitionRevision: "d1", SourceRevision: "s1", EvidenceBasis: "b1", CustodyClock: "local", EvaluationTimeNS: "100", EvidenceCutoffNS: "110", DeadlineNS: "10", AllowanceNS: "3", WindowNS: "50", SubjectField: "subject_id", SuccessOutcome: "ok", TerminalOutcomes: []string{"ok"}, Scopes: []Scope{{Environment: "local", TenantID: "demo", ServiceName: "service", Coverage: "unverified"}}}
	if _, e := p.Prepare(nil); e == nil {
		t.Fatal("implicit future knowledge accepted")
	}
	p.KnowledgeAfterEvaluation = true
	future, e := p.Prepare([]Event{{EventID: "future", Kind: "accepted", Environment: "local", TenantID: "demo", ServiceName: "service", SubjectID: "s1", Release: "r1", EventTimeNS: "105", ReceivedTimeNS: "109"}})
	if e != nil || future.Receipt.InvalidRecords != 1 || len(future.Subjects) != 0 {
		t.Fatal("future business event became evidence", e)
	}
	prior, e := p.Prepare(nil)
	if e != nil {
		t.Fatal(e)
	}
	p.EvaluationID = "e2"
	p.EvidenceBasis = "b2"
	p.EvidenceCutoffNS = "120"
	p.SupersedesEvaluationID = "e1"
	if e = p.ValidateCorrection(prior.Receipt); e != nil {
		t.Fatal(e)
	}
	p.KnowledgeAfterEvaluation = false
	if e = p.ValidateCorrection(prior.Receipt); e == nil {
		t.Fatal("knowledge mode changed during correction")
	}
}
