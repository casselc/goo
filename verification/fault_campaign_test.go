package verification

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/casselc/observability/quintgo/qobs"
	"github.com/casselc/observability/quintgo/validate"
)

type faultFacts struct {
	RetrySeen, Stalled, TerminalSeen bool
	Terminal, Signature, Meaning     string
	Attempts, Errors                 int
}
type faultLedger struct {
	Version           int
	Subject, Scenario string
	Source, Display   faultFacts
	Steps             int
	Passed, Complete  bool
}

func checkFaultLedger(t *testing.T, l faultLedger, dir string) bool {
	t.Helper()
	if l.Version != 1 || !l.Complete || l.Subject == "" {
		return false
	}
	sink := &qobs.MemorySink{}
	rec := qobs.NewRecorder(sink, qobs.WithActor("fault-observer"), qobs.WithProcess("campaign"))
	ctx := qobs.Thread(context.Background())
	rec.Record(ctx, "accepted", nil, "subject", l.Subject, "scenario", l.Scenario)
	rec.Record(ctx, "source_ready", nil, "subject", l.Subject, "signature", l.Source.Signature,
		"retrySeen", l.Source.RetrySeen, "stalled", l.Source.Stalled, "terminalSeen", l.Source.TerminalSeen, "terminal", l.Source.Terminal,
		"attempts", l.Source.Attempts, "errors", l.Source.Errors)
	rec.Record(ctx, "displayed", nil, "subject", l.Subject, "signature", l.Display.Signature, "meaning", l.Display.Meaning)
	rec.Record(ctx, "explored", nil, "steps", l.Steps, "passed", l.Passed)
	checker, err := validate.NewChecker("model/fault_observation.yaml", validate.Options{Quint: quintPath(t), Backend: "typescript", Dir: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	report, err := checker.Check(ctx, sink.Steps())
	if err != nil {
		t.Fatal(err)
	}
	t.Log(report.Summary())
	if dir != "" {
		if err = os.WriteFile(filepath.Join(dir, "fault-report.txt"), []byte(report.Summary()), 0644); err != nil {
			t.Fatal(err)
		}
		var lines []byte
		for _, step := range sink.Steps() {
			line, err := step.MarshalLine()
			if err != nil {
				t.Fatal(err)
			}
			lines = append(lines, line...)
			lines = append(lines, '\n')
		}
		if err = os.WriteFile(filepath.Join(dir, "fault-steps.jsonl"), lines, 0644); err != nil {
			t.Fatal(err)
		}
		for _, actor := range report.Actors {
			if actor.ITF != "" {
				data, err := os.ReadFile(actor.ITF)
				if err != nil {
					t.Fatal(err)
				}
				if err = os.WriteFile(filepath.Join(dir, "fault.itf.json"), data, 0644); err != nil {
					t.Fatal(err)
				}
			}
		}
	}
	return report.OK()
}

func TestQuintFaultEvidenceContract(t *testing.T) {
	// Labelled protocol fixtures; real records enter through the live-ledger test.
	for _, scenario := range []string{"retry", "reject", "missing"} {
		t.Run(scenario, func(t *testing.T) {
			facts := faultFacts{Signature: "captured-members-and-question", Attempts: 1}
			switch scenario {
			case "retry":
				facts.TerminalSeen = true
				facts.RetrySeen = true
				facts.Terminal = "confirmed"
				facts.Attempts = 2
				facts.Errors = 1
				facts.Meaning = "confirmed-after-retry"
			case "reject":
				facts.TerminalSeen = true
				facts.Terminal = "rejected"
				facts.Meaning = "rejected"
			case "missing":
				facts.Stalled = true
				facts.Meaning = "unknown-no-terminal"
			}
			base := faultLedger{Version: 1, Subject: "actual-subject", Scenario: scenario, Source: facts, Display: facts, Steps: 10, Passed: true, Complete: true}
			if !checkFaultLedger(t, base, "") {
				t.Fatal("valid fixture rejected")
			}
			for _, mutate := range []func(*faultLedger){
				func(l *faultLedger) { l.Display.Meaning = "healthy" },
				func(l *faultLedger) { l.Display.Signature = "other-window-or-members" },
				func(l *faultLedger) { l.Source.Terminal = "invented" },
				func(l *faultLedger) { l.Source.TerminalSeen = !l.Source.TerminalSeen },
			} {
				bad := base
				mutate(&bad)
				if checkFaultLedger(t, bad, "") {
					t.Fatal("mutated evidence accepted")
				}
			}
		})
	}
}

func TestQuintFaultCampaignLiveLedger(t *testing.T) {
	file := os.Getenv("GOO_FAULT_LEDGER")
	if file == "" {
		t.Skip("optional: supply actual fault campaign ledger")
	}
	data, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	var ledger faultLedger
	if err = json.Unmarshal(data, &ledger); err != nil {
		t.Fatal(err)
	}
	if !checkFaultLedger(t, ledger, filepath.Dir(file)) {
		t.Fatal("live fault ledger rejected")
	}
}
