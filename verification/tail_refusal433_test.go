package verification

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"
	"unicode/utf8"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/internal/sourcefailure"
	"github.com/casselc/goo/internal/sourcehttp"
	"github.com/casselc/goo/internal/tailprotocol"
	"github.com/casselc/observability/quintgo/qobs"
	"github.com/casselc/observability/quintgo/validate"
	"hegel.dev/go/hegel"
)

// Error text is deliberately arbitrary and sensitive-looking. Classification is
// derived from the chosen typed scenario, never from that text or output fields.
func TestHegelTailRefusal433(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		scenario := hegel.Draw(tc, hegel.Integers(0, 9))
		status := hegel.Draw(tc, hegel.SampledFrom([]int{400, 401, 403, 404, 429, 500, 502, 503}))
		code := hegel.Draw(tc, hegel.SampledFrom([]int{0, 16, 47, 53, 60, 81, 159, 241, 497, 516, 999}))
		secret := fmt.Sprintf("private-host-SELECT-password-%d", hegel.Draw(tc, hegel.Integers(0, 1000000)))
		cause := errors.New(secret)
		var input error
		wantReason, wantRetry, wantKind := "source_unavailable", tailprotocol.Explicit, sourcefailure.Unavailable
		switch scenario {
		case 0:
			input = &sourcehttp.HTTPError{Status: status, Code: code}
			// These independently chosen source metadata categories have priority over
			// unclassified HTTP failures. No status/body text grants automatic retry.
			if status == 401 || status == 403 || code == 497 || code == 516 {
				wantReason = "source_access"
			} else if code == 60 || code == 81 {
				wantReason = "source_missing"
			} else if code == 16 || code == 47 || code == 53 {
				wantReason = "source_schema"
			}
		case 1:
			input = cause
		case 2:
			notFound := hegel.Draw(tc, hegel.Booleans())
			input = &net.DNSError{Err: secret, IsNotFound: notFound}
			if !notFound {
				wantRetry = tailprotocol.Automatic
			}
		case 3:
			input = &net.DNSError{Err: secret, IsTimeout: true, IsNotFound: hegel.Draw(tc, hegel.Booleans())}
			wantReason, wantRetry, wantKind = "source_timeout", tailprotocol.Automatic, sourcefailure.Timeout
		case 4:
			input = context.DeadlineExceeded
			wantReason, wantRetry, wantKind = "source_timeout", tailprotocol.Automatic, sourcefailure.Timeout
		case 5:
			input = context.Canceled
			wantReason, wantKind = "source_cancelled", sourcefailure.Cancelled
		case 6:
			input = sourcefailure.Wrap(sourcefailure.Budget, cause)
			wantReason, wantKind = "source_budget", sourcefailure.Budget
		case 7:
			input = sourcefailure.Wrap(sourcefailure.Malformed, cause)
			wantReason, wantKind = "source_malformed", sourcefailure.Malformed
		case 8:
			input = sourcefailure.Wrap(sourcefailure.Cancelled, cause)
			wantReason, wantKind = "source_cancelled", sourcefailure.Cancelled
		case 9:
			input = sourcefailure.Wrap(sourcefailure.Timeout, cause)
			wantReason, wantRetry, wantKind = "source_timeout", tailprotocol.Automatic, sourcefailure.Timeout
		}
		input = fmt.Errorf("%s: %w", secret, input)
		got := tailprotocol.Source(input)
		if got.Reason != wantReason || got.Retry != wantRetry || got.Refusal == nil || got.Refusal.Kind != wantKind {
			tc.Fatal("typed error policy changed", scenario, status, code, got)
		}
		encoded, err := json.Marshal(got)
		if err != nil {
			tc.Fatal(err)
		}
		if strings.Contains(string(encoded), secret) || strings.Contains(string(encoded), "SELECT") || strings.Contains(string(encoded), "private-host") || len(got.Refusal.Detail) > sourcefailure.MaxDetailBytes {
			tc.Fatal("source prose reached wire", string(encoded))
		}
		// Wrapping and arbitrary prose must not change the safe wire diagnostic.
		again := tailprotocol.Source(fmt.Errorf("different untrusted error: %w", input))
		wire, _ := json.Marshal(again)
		if string(encoded) != string(wire) {
			tc.Fatal("prose or wrapping changed policy", got, again)
		}
		parent, cancel := context.WithCancel(context.Background())
		defer cancel()
		expired, stop := context.WithDeadline(parent, time.Unix(0, 0))
		defer stop()
		if hegel.Draw(tc, hegel.Booleans()) {
			cancel()
			if _, emit := tailprotocol.End(parent, expired); emit {
				tc.Fatal("caller cancellation emitted rollover")
			}
		} else {
			if end, emit := tailprotocol.End(parent, expired); !emit || end.Reason != "connection_ended" || end.Retry != tailprotocol.Automatic || end.Refusal != nil {
				tc.Fatal("connection deadline lost rollover semantics", end, emit)
			}
		}
	}, options(150)...)
}

type tailEvent433 struct {
	Event string          `json:"event"`
	ID    json.RawMessage `json:"id"`
	Data  json.RawMessage `json:"data"`
}
type tailStream433 struct {
	Request json.RawMessage `json:"request"`
	Status  int             `json:"status"`
	Events  []tailEvent433  `json:"events"`
}
type tailFact433 struct {
	Name            string            `json:"name"`
	SourceReads     *int              `json:"sourceReads"`
	Selector        json.RawMessage   `json:"selector"`
	Retained        *int              `json:"retained"`
	Phase           string            `json:"phase"`
	GapReason       string            `json:"gapReason"`
	RecoveryVisible *bool             `json:"recoveryVisible"`
	Gap             *tailprotocol.Gap `json:"gap"`
	HeldKey         string            `json:"heldKey"`
}
type tailCase433 struct {
	Signal       string                                    `json:"signal"`
	Selector     json.RawMessage                           `json:"selector"`
	Facts        []tailFact433                             `json:"facts"`
	MissingQuiet struct{ Before, After, Milliseconds int } `json:"missingQuiet"`
}
type tailProbe433 struct {
	Name        string            `json:"name"`
	Selector    json.RawMessage   `json:"selector"`
	FirstGap    tailprotocol.Gap  `json:"firstGap"`
	SourceReads int               `json:"sourceReads"`
	Retained    int               `json:"retained"`
	Recovered   bool              `json:"recovered"`
	Requests    []json.RawMessage `json:"requests"`
}
type tailReport433 struct {
	Version string          `json:"version"`
	Status  string          `json:"status"`
	Errors  []string        `json:"errors"`
	Cases   []tailCase433   `json:"cases"`
	Streams []tailStream433 `json:"streams"`
	Probes  []tailProbe433  `json:"probes"`
}
type tailStep433 struct {
	Action                                      string
	Reads                                       int
	PolicyOK, ScopeOK, HeldOK, OwnerOK, NoWiden bool
}

var tailActions433 = []struct{ name, action string }{
	{"missing", "missing"}, {"missing_quiet", "quiet"}, {"schema", "schema"}, {"recovered", "recover"},
	{"held_refusal", "fail_held"}, {"inspect_held", "inspect"}, {"dirty_retry_refused", "refuse_dirty"}, {"hidden_retry_refused", "refuse_hidden"},
}
var tailHandle433 = regexp.MustCompile(`^[a-f0-9]{64}:[1-9][0-9]*$`)

func tailDecode433(raw json.RawMessage, out any) error {
	if len(raw) == 0 || bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
		return fmt.Errorf("missing observed wire object")
	}
	d := json.NewDecoder(bytes.NewReader(raw))
	d.DisallowUnknownFields()
	if err := d.Decode(out); err != nil {
		return err
	}
	var extra any
	if err := d.Decode(&extra); err != io.EOF {
		return fmt.Errorf("trailing observed wire data")
	}
	return nil
}
func tailSelector433(raw json.RawMessage) (exploration.TailSelector, error) {
	var s exploration.TailSelector
	if err := tailDecode433(raw, &s); err != nil {
		return s, err
	}
	return s, s.Query(1800000).Validate()
}
func tailSameScope433(a, b exploration.TailSelector) bool {
	aq, bq := a.Query(1800000), b.Query(1800000)
	aw, _ := json.Marshal(aq.Where)
	bw, _ := json.Marshal(bq.Where)
	return aq.Signal == bq.Signal && aq.SameEnvironment(bq) && bytes.Equal(aw, bw)
}
func tailGap433(g tailprotocol.Gap) bool {
	if g.Reason == "connection_ended" {
		return g.Retry == tailprotocol.Automatic && g.Refusal == nil
	}
	if g.Refusal == nil || !utf8.ValidString(g.Refusal.Detail) || len(g.Refusal.Detail) == 0 || len(g.Refusal.Detail) > sourcefailure.MaxDetailBytes || strings.ContainsAny(g.Refusal.Detail, "\x00\r\n\t") {
		return false
	}
	switch g.Reason {
	case "source_missing", "source_access", "source_schema":
		return g.Retry == tailprotocol.Explicit && g.Refusal.Kind == sourcefailure.Unavailable
	case "source_unavailable":
		return (g.Retry == tailprotocol.Explicit || g.Retry == tailprotocol.Automatic) && g.Refusal.Kind == sourcefailure.Unavailable
	case "source_timeout":
		return g.Retry == tailprotocol.Automatic && g.Refusal.Kind == sourcefailure.Timeout
	case "source_cancelled":
		return g.Retry == tailprotocol.Explicit && g.Refusal.Kind == sourcefailure.Cancelled
	case "source_malformed":
		return g.Retry == tailprotocol.Explicit && g.Refusal.Kind == sourcefailure.Malformed
	case "source_budget", "identity_budget":
		return g.Retry == tailprotocol.Explicit && g.Refusal.Kind == sourcefailure.Budget
	}
	return false
}
func tailGapEqual433(a, b tailprotocol.Gap) bool {
	av, _ := json.Marshal(a)
	bv, _ := json.Marshal(b)
	return bytes.Equal(av, bv)
}

// The report records actual SSE bytes after connections complete. Facts can only
// refer to a gap or retained handle that was present on those wire observations.
func tailWire433(report tailReport433, selector exploration.TailSelector, minimumConnections, minimumGaps int) (map[string]bool, []tailprotocol.Gap, error) {
	keys := map[string]bool{}
	gaps := []tailprotocol.Gap{}
	connections := 0
	for _, stream := range report.Streams {
		s, err := tailSelector433(stream.Request)
		if err != nil {
			return nil, nil, err
		}
		if !tailSameScope433(s, selector) {
			continue
		}
		connections++
		if stream.Status != 200 {
			return nil, nil, fmt.Errorf("expected admitted SSE source connection")
		}
		for index, event := range stream.Events {
			var id string
			if json.Unmarshal(event.ID, &id) != nil {
				var number int
				if json.Unmarshal(event.ID, &number) != nil {
					return nil, nil, fmt.Errorf("invalid SSE sequence")
				}
				id = fmt.Sprint(number)
			}
			if id != fmt.Sprint(index+1) {
				return nil, nil, fmt.Errorf("nonconsecutive SSE sequence")
			}
			if index == 0 {
				var reset struct {
					Connection string          `json:"connection"`
					WindowMS   int64           `json:"windowMs"`
					Selector   json.RawMessage `json:"selector"`
					Continuity string          `json:"continuity"`
					Replay     bool            `json:"replay"`
				}
				if event.Event != "reset" || tailDecode433(event.Data, &reset) != nil || reset.WindowMS != exploration.TailWindowMS || reset.Replay || reset.Continuity != "unknown" || len(reset.Connection) != 64 {
					return nil, nil, fmt.Errorf("invalid observed tail reset")
				}
				rs, e := tailSelector433(reset.Selector)
				if e != nil || !tailSameScope433(rs, selector) {
					return nil, nil, fmt.Errorf("reset selector changed")
				}
				continue
			}
			switch event.Event {
			case "gap":
				var gap tailprotocol.Gap
				if tailDecode433(event.Data, &gap) != nil || !tailGap433(gap) || index != len(stream.Events)-1 {
					return nil, nil, fmt.Errorf("invalid or nonterminal observed gap")
				}
				gaps = append(gaps, gap)
			case "batch":
				var batch struct {
					Rows []struct {
						Key    string          `json:"key"`
						Record json.RawMessage `json:"record"`
						Late   bool            `json:"late"`
					} `json:"rows"`
					FromMS, ToMS int64
					ObservedAt   string
					Limited      bool
				}
				// Batch carries optional custody fields. The existing Go/browser contracts
				// own them; this replay checks source rows and their event/read clocks only.
				if json.Unmarshal(event.Data, &batch) != nil || batch.FromMS < 0 || batch.ToMS-batch.FromMS != exploration.TailWindowMS {
					return nil, nil, fmt.Errorf("invalid observed batch window")
				}
				clock, e := time.Parse(time.RFC3339Nano, batch.ObservedAt)
				if e != nil || clock.Before(time.UnixMilli(batch.ToMS)) {
					return nil, nil, fmt.Errorf("invalid observed batch clock")
				}
				for _, row := range batch.Rows {
					if !tailHandle433.MatchString(row.Key) {
						return nil, nil, fmt.Errorf("invalid observed row handle")
					}
					keys[row.Key] = true
				}
			default:
				return nil, nil, fmt.Errorf("unknown observed tail event")
			}
		}
	}
	if connections < minimumConnections || len(gaps) < minimumGaps {
		return nil, nil, fmt.Errorf("missing actual source refusal/retry connections")
	}
	return keys, gaps, nil
}
func tailObserved433(report tailReport433, c tailCase433) ([]tailStep433, error) {
	if len(c.Facts) != len(tailActions433) || c.MissingQuiet.Before != c.MissingQuiet.After || c.MissingQuiet.Milliseconds < 2300 {
		return nil, fmt.Errorf("missing complete quiet refusal observation")
	}
	origin, err := tailSelector433(c.Selector)
	if err != nil || origin.Query(1800000).Signal != c.Signal {
		return nil, fmt.Errorf("invalid original tail selector")
	}
	keys, gaps, err := tailWire433(report, origin, 3, 3)
	if err != nil {
		return nil, err
	}
	steps := make([]tailStep433, 0, len(c.Facts))
	reads := 0
	heldKey := ""
	for index, fact := range c.Facts {
		if fact.Name != tailActions433[index].name || fact.SourceReads == nil || *fact.SourceReads < 0 || fact.Retained == nil || fact.RecoveryVisible == nil {
			return nil, fmt.Errorf("missing observed fact %d", index)
		}
		s, e := tailSelector433(fact.Selector)
		if e != nil {
			return nil, e
		}
		scopeOK := tailSameScope433(s, origin)
		reads += *fact.SourceReads
		policyOK := true
		heldOK := *fact.Retained == 0
		if index == 3 {
			heldKey = fact.HeldKey
			heldOK = *fact.Retained > 0 && keys[heldKey]
			policyOK = fact.Gap == nil && fact.GapReason == "" && fact.Phase == "unknown"
		} else {
			want := "source_missing"
			if index == 2 {
				want = "source_schema"
			}
			policyOK = fact.Gap != nil && tailGap433(*fact.Gap) && fact.Gap.Reason == want && fact.Gap.Retry == tailprotocol.Explicit && fact.GapReason == want && fact.Phase == "unavailable"
			if fact.Gap != nil {
				seen := false
				for _, gap := range gaps {
					seen = seen || tailGapEqual433(gap, *fact.Gap)
				}
				policyOK = policyOK && seen
			}
			if index > 3 {
				heldOK = *fact.Retained > 0 && fact.HeldKey == heldKey && keys[fact.HeldKey]
			}
		}
		ownerOK := index < 5 || *fact.SourceReads == 0
		steps = append(steps, tailStep433{Action: tailActions433[index].action, Reads: reads, PolicyOK: policyOK, ScopeOK: scopeOK, HeldOK: heldOK, OwnerOK: ownerOK, NoWiden: !*fact.RecoveryVisible})
	}
	return steps, nil
}
func tailCheck433(t *testing.T, steps []tailStep433) (bool, error) {
	t.Helper()
	quint, err := exec.LookPath("quint")
	if err != nil {
		return false, fmt.Errorf("native tail refusal replay requires quint: %w", err)
	}
	spec, err := filepath.Abs("model/tail_refusal433.qnt")
	if err != nil {
		return false, err
	}
	manifest := "spec: " + spec + "\nmodule: tail_refusal433\nconstants: {INITIAL: '0'}\ninvariants: [safety]\nactions:\n"
	for _, a := range tailActions433 {
		manifest += "  " + a.action + ":\n    expect: 'reads == {{.obs.reads}} and policyOK == {{.obs.policyOK}} and scopeOK == {{.obs.scopeOK}} and heldOK == {{.obs.heldOK}} and ownerOK == {{.obs.ownerOK}} and noWiden == {{.obs.noWiden}}'\n"
	}
	config := filepath.Join(t.TempDir(), "binding.yaml")
	if err = os.WriteFile(config, []byte(manifest), 0600); err != nil {
		return false, err
	}
	sink := &qobs.MemorySink{}
	rec := qobs.NewRecorder(sink)
	for _, step := range steps {
		rec.Record(context.Background(), step.Action, nil, "obs.reads", step.Reads, "obs.policyOK", step.PolicyOK, "obs.scopeOK", step.ScopeOK, "obs.heldOK", step.HeldOK, "obs.ownerOK", step.OwnerOK, "obs.noWiden", step.NoWiden)
	}
	checker, err := validate.NewChecker(config, validate.Options{Quint: quint, Backend: "typescript", Dir: t.TempDir()})
	if err != nil {
		return false, err
	}
	result, err := checker.Check(context.Background(), sink.Steps())
	if err != nil {
		return false, err
	}
	return result.OK(), nil
}
func TestQuintObservedTailRefusal433(t *testing.T) {
	file := os.Getenv("GOO_TAIL433_REPORT")
	if file == "" {
		t.Skip("set GOO_TAIL433_REPORT to the completed real tail-refusal journey report for required acceptance")
	}
	raw, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	var report tailReport433
	if err = json.Unmarshal(raw, &report); err != nil {
		t.Fatal(err)
	}
	if report.Version != "goo.tailRefusal433/1" || report.Status != "passed" || len(report.Errors) > 0 || len(report.Cases) != 2 {
		t.Fatal("incomplete actual tail-refusal report")
	}
	seen := map[string]bool{}
	for _, c := range report.Cases {
		t.Run(c.Signal, func(t *testing.T) {
			if seen[c.Signal] || (c.Signal != "logs" && c.Signal != "spans") {
				t.Fatal("duplicate or unsupported source case")
			}
			seen[c.Signal] = true
			steps, err := tailObserved433(report, c)
			if err != nil {
				t.Fatal(err)
			}
			accepted, err := tailCheck433(t, steps)
			if err != nil || !accepted {
				t.Fatal("actual source refusal trace rejected", accepted, err)
			}
			for _, mutation := range []string{"retry-disposition", "scope", "implicit-inspection-read", "stale-owner", "lost-held-row", "false-empty-recovery"} {
				t.Run(mutation, func(t *testing.T) {
					bad := append([]tailStep433(nil), steps...)
					switch mutation {
					case "retry-disposition":
						bad[0].PolicyOK = false
					case "scope":
						bad[3].ScopeOK = false
					case "implicit-inspection-read":
						bad[5].Reads++
					case "stale-owner":
						bad[6].OwnerOK = false
					case "lost-held-row":
						bad[5].HeldOK = false
					case "false-empty-recovery":
						bad[0].NoWiden = false
					}
					accepted, err := tailCheck433(t, bad)
					if err != nil {
						t.Fatal(err)
					}
					if accepted {
						t.Fatal("unsafe counterfactual accepted", mutation)
					}
				})
			}
		})
	}
	if !seen["logs"] || !seen["spans"] {
		t.Fatal("missing source signal coverage")
	}

	t.Run("gap-before-reset", func(t *testing.T) {
		data, _ := json.Marshal(report)
		var bad tailReport433
		if err := json.Unmarshal(data, &bad); err != nil {
			t.Fatal(err)
		}
		changed := false
		var alteredSelector json.RawMessage
		for index, stream := range bad.Streams {
			if len(stream.Events) >= 2 && stream.Events[0].Event == "reset" && stream.Events[1].Event == "gap" {
				bad.Streams[index].Events = stream.Events[1:]
				for i := range bad.Streams[index].Events {
					bad.Streams[index].Events[i].ID, _ = json.Marshal(fmt.Sprint(i + 1))
				}
				changed = true
				alteredSelector = stream.Request
				break
			}
		}
		if !changed {
			t.Fatal("missing real initial refusal to mutate")
		}
		origin, err := tailSelector433(alteredSelector)
		if err != nil {
			t.Fatal(err)
		}
		if _, _, err = tailWire433(bad, origin, 1, 1); err == nil {
			t.Fatal("terminal gap before reset accepted")
		}
	})
	checkTailProbes433(t, report)
}

func tailProbeCheck433(t *testing.T, mode int, scopeOK, held bool, observedMode int) (bool, error) {
	t.Helper()
	quint, err := exec.LookPath("quint")
	if err != nil {
		return false, err
	}
	spec, err := filepath.Abs("model/tail_refusal433.qnt")
	if err != nil {
		return false, err
	}
	binding := "spec: " + spec + "\nmodule: tail_probe433\nconstants: {INITIAL: '0', MODE: '" + fmt.Sprint(mode) + "'}\ninvariants: [safety]\nactions:\n  refused:\n    expect: 'reads == {{.obs.reads}} and retryMode == {{.obs.retryMode}} and scopeOK == {{.obs.scopeOK}} and held == {{.obs.held}}'\n  recovered:\n    expect: 'reads == {{.obs.reads}} and retryMode == {{.obs.retryMode}} and scopeOK == {{.obs.scopeOK}} and held == {{.obs.held}}'\n"
	file := filepath.Join(t.TempDir(), "probe.yaml")
	if err = os.WriteFile(file, []byte(binding), 0600); err != nil {
		return false, err
	}
	sink := &qobs.MemorySink{}
	rec := qobs.NewRecorder(sink)
	rec.Record(context.Background(), "refused", nil, "obs.reads", 1, "obs.retryMode", observedMode, "obs.scopeOK", scopeOK, "obs.held", false)
	rec.Record(context.Background(), "recovered", nil, "obs.reads", 2, "obs.retryMode", observedMode, "obs.scopeOK", scopeOK, "obs.held", held)
	checker, err := validate.NewChecker(file, validate.Options{Quint: quint, Backend: "typescript", Dir: t.TempDir()})
	if err != nil {
		return false, err
	}
	result, err := checker.Check(context.Background(), sink.Steps())
	if err != nil {
		return false, err
	}
	return result.OK(), nil
}
func checkTailProbes433(t *testing.T, report tailReport433) {
	t.Helper()
	if len(report.Probes) != 2 {
		t.Fatal("missing access/transport source-boundary observations")
	}
	seen := map[string]bool{}
	for _, probe := range report.Probes {
		t.Run(probe.Name, func(t *testing.T) {
			if seen[probe.Name] || (probe.Name != "access" && probe.Name != "transport") {
				t.Fatal("duplicate or unknown recovery probe")
			}
			seen[probe.Name] = true
			selector, err := tailSelector433(probe.Selector)
			if err != nil {
				t.Fatal(err)
			}
			if len(probe.Requests) != 2 || probe.SourceReads != 2 || probe.Retained != 1 || !probe.Recovered {
				t.Fatal("incomplete actual recovery probe", probe)
			}
			scopeOK := true
			for _, request := range probe.Requests {
				s, err := tailSelector433(request)
				if err != nil {
					t.Fatal(err)
				}
				scopeOK = scopeOK && tailSameScope433(s, selector)
			}
			keys, gaps, err := tailWire433(report, selector, 2, 1)
			if err != nil {
				t.Fatal(err)
			}
			seenGap := false
			for _, gap := range gaps {
				seenGap = seenGap || tailGapEqual433(gap, probe.FirstGap)
			}
			mode := 0
			wantReason, wantRetry := "source_access", tailprotocol.Explicit
			if probe.Name == "transport" {
				mode = 1
				wantReason, wantRetry = "source_unavailable", tailprotocol.Automatic
			}
			if !tailGap433(probe.FirstGap) || !seenGap || probe.FirstGap.Reason != wantReason || probe.FirstGap.Retry != wantRetry {
				t.Fatal("actual source disposition missing", probe.FirstGap)
			}
			held := len(keys) > 0 && probe.Retained == 1
			ok, err := tailProbeCheck433(t, mode, scopeOK, held, mode)
			if err != nil || !ok {
				t.Fatal("actual source recovery policy rejected", ok, err)
			}
			ok, err = tailProbeCheck433(t, mode, scopeOK, held, 1-mode)
			if err != nil {
				t.Fatal(err)
			}
			if ok {
				t.Fatal("reversed source retry disposition accepted")
			}
		})
	}
}
