package workbench

import (
	"bytes"
	"encoding/json"
	"os/exec"
	"reflect"
	"testing"

	"github.com/casselc/goo/identity"
	"github.com/casselc/goo/measurements"
)

func TestTypedNativeGaugeStep319(t *testing.T) {
	for _, mode := range []string{"named", "empty", "unreported"} {
		t.Run(mode, func(t *testing.T) {
			wire := identity.EnvironmentSelector{Version: identity.EnvironmentSelectorVersion, Policy: "effective", Scope: identity.EnvironmentScope{Mode: mode}}
			if mode == "named" {
				wire.Scope.Value = "prod"
			}
			definition := measurements.Definition{APIVersion: measurements.Version, Name: "gauge", Title: "Gauge", Kind: "gauge", EnvironmentSelector: &wire, LookbackMinutes: 1, MaxGapMS: 60000, Grain: "per-stream", Source: measurements.Binding{Services: []string{"api"}, Scope: "meter", Metric: "queue.depth", Unit: "items"}}
			scopeWire := wire
			step := Step{Version: 1, Kind: "measurement", Scope: StepScope{EnvironmentSelector: &scopeWire, FromMS: 1000, ToMS: 2000}, View: "Overview", Measurement: &definition}
			if err := step.Validate(); err != nil {
				t.Fatal("typed gauge step refused", err)
			}
			raw, _ := json.Marshal(step)
			command := exec.Command("node", "--input-type=module", "-e", `import fs from 'node:fs';import assert from 'node:assert/strict';import * as P from '../ui/question-protocol.mjs';const s=JSON.parse(fs.readFileSync(0,'utf8'));P.validateStep(s);const round=P.readStepLink(P.stepLink(s));assert.deepEqual(round,s);const capture=P.captureQuestion(round);assert.deepEqual(capture.measurement.definition.environmentSelector,s.scope.environmentSelector);assert.deepEqual(P.evidenceDifference(s,capture),[]);process.stdout.write(JSON.stringify(round));`)
			command.Stdin = bytes.NewReader(raw)
			output, err := command.CombinedOutput()
			if err != nil {
				t.Fatalf("typed native Go/JS step: %v %s", err, output)
			}
			var round Step
			if err := json.Unmarshal(output, &round); err != nil || !reflect.DeepEqual(step, round) {
				t.Fatal("native link changed policy", err, string(output))
			}
			for _, change := range []func(*Step){
				func(s *Step) { s.Scope.EnvironmentSelector = nil; s.Scope.Environment = "prod" },
				func(s *Step) {
					s.Scope.EnvironmentSelector = &identity.EnvironmentSelector{Version: identity.EnvironmentSelectorVersion, Policy: "current-only", Scope: identity.EnvironmentScope{Mode: "named", Value: "prod"}}
				},
				func(s *Step) { s.Measurement.Kind = "stock" },
			} {
				var invalid Step
				if err := json.Unmarshal(raw, &invalid); err != nil {
					t.Fatal(err)
				}
				change(&invalid)
				if err := invalid.Validate(); err == nil {
					t.Fatal("native step admitted changed policy or unsupported family")
				}
			}
		})
	}
}
