package verification

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/casselc/observability/quintgo/connect"
	"github.com/casselc/observability/quintgo/itf"
)

type campaignCheckpoint struct {
	Action   string `json:"action"`
	Target   int64  `json:"target"`
	Issued   int64  `json:"issued"`
	Checked  int64  `json:"checked"`
	Revisits int64  `json:"revisits"`
	Done     bool   `json:"done"`
}
type browserCampaignPlan struct {
	Version     int                  `json:"version"`
	Seed        int64                `json:"seed"`
	Trace       string               `json:"trace"`
	Checkpoints []campaignCheckpoint `json:"checkpoints"`
}

func campaignProjection(action string, vars map[string]itf.Value) campaignCheckpoint {
	return campaignCheckpoint{Action: action, Target: vars["target"].Int64(),
		Issued: vars["issued"].Int64(), Checked: vars["checked"].Int64(),
		Revisits: vars["revisits"].Int64(), Done: vars["done"].B}
}

func replayBrowserCampaign(trace *itf.Trace, observations []campaignCheckpoint) error {
	if len(trace.States) != 10 || len(observations) != 9 || !observations[8].Done {
		return fmt.Errorf("require complete nine-step campaign")
	}
	driver := connect.Driver{Actions: map[string]connect.Handler{}}
	for _, action := range []string{"issue", "inspect", "revisit", "finish", "idle"} {
		driver.Actions[action] = func(s *connect.Step) error {
			actual := observations[s.Index-1]
			expected := campaignProjection(s.Action, s.After)
			if actual != expected {
				return fmt.Errorf("browser checkpoint %+v differs from model %+v", actual, expected)
			}
			return nil
		}
	}
	return driver.Replay(trace)
}

func TestQuintGenerateBrowserCampaign(t *testing.T) {
	seed := int64(42)
	if value := os.Getenv("GOO_CAMPAIGN_SEED"); value != "" {
		var err error
		seed, err = strconv.ParseInt(value, 10, 64)
		if err != nil || seed <= 0 {
			t.Fatal("GOO_CAMPAIGN_SEED must be a positive int64")
		}
	}
	spec, _ := filepath.Abs("model/browser_campaign.qnt")
	files, err := (connect.Generate{Quint: quintPath(t), Spec: spec, Main: "browserCampaignInstance",
		Traces: 1, MaxSteps: 9, Seed: seed, Backend: "typescript", Dir: t.TempDir(),
		Extra: []string{"--invariant=safety"}}).Run(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	trace, err := itf.ReadFile(files[0])
	if err != nil {
		t.Fatal(err)
	}
	plan := browserCampaignPlan{Version: 1, Seed: seed, Trace: "generated.itf.json"}
	for _, state := range trace.States[1:] {
		plan.Checkpoints = append(plan.Checkpoints, campaignProjection(state.Action, state.Vars))
	}
	if err = replayBrowserCampaign(trace, plan.Checkpoints); err != nil {
		t.Fatal(err)
	}
	// These are protocol projection fixtures, not application/browser evidence.
	for _, mutation := range []func([]campaignCheckpoint){
		func(c []campaignCheckpoint) { c[0].Issued++ },
		func(c []campaignCheckpoint) { c[1].Target++ },
		func(c []campaignCheckpoint) { c[0], c[1] = c[1], c[0] },
	} {
		bad := append([]campaignCheckpoint(nil), plan.Checkpoints...)
		mutation(bad)
		if replayBrowserCampaign(trace, bad) == nil {
			t.Fatal("mutated projection accepted")
		}
	}
	if replayBrowserCampaign(trace, plan.Checkpoints[:8]) == nil {
		t.Fatal("truncated campaign accepted")
	}
	if output := os.Getenv("GOO_CAMPAIGN_PLAN_OUTPUT"); output != "" {
		data, err := json.MarshalIndent(plan, "", "  ")
		if err != nil {
			t.Fatal(err)
		}
		if err = os.WriteFile(output, data, 0644); err != nil {
			t.Fatal(err)
		}
		data, err = os.ReadFile(files[0])
		if err != nil {
			t.Fatal(err)
		}
		if err = os.WriteFile(filepath.Join(filepath.Dir(output), plan.Trace), data, 0644); err != nil {
			t.Fatal(err)
		}
	}
	t.Logf("generated seed %d with checkpoints %+v", seed, plan.Checkpoints)
}

func TestQuintBrowserCampaignLiveObservations(t *testing.T) {
	file := os.Getenv("GOO_CAMPAIGN_OBSERVATIONS")
	if file == "" {
		t.Skip("optional: supply the actual generated browser campaign observations")
	}
	data, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	var observed struct {
		Version     int
		Checkpoints []campaignCheckpoint
	}
	if err = json.Unmarshal(data, &observed); err != nil {
		t.Fatal(err)
	}
	if observed.Version != 1 {
		t.Fatal("unsupported campaign observation version")
	}
	trace, err := itf.ReadFile(filepath.Join(filepath.Dir(file), "generated.itf.json"))
	if err != nil {
		t.Fatal(err)
	}
	if err = replayBrowserCampaign(trace, observed.Checkpoints); err != nil {
		t.Fatal(err)
	}
	t.Log("all nine actual browser checkpoints conform to generated Quint trace")
}

func TestQuintBrowserCampaignScheduleVariety(t *testing.T) {
	spec, _ := filepath.Abs("model/browser_campaign.qnt")
	files, err := (connect.Generate{Quint: quintPath(t), Spec: spec, Main: "browserCampaignInstance",
		Traces: 8, MaxSteps: 9, Seed: 2026, Backend: "typescript", Dir: t.TempDir(),
		Extra: []string{"--invariant=safety"}}).Run(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	sequences := map[string]bool{}
	for _, file := range files {
		trace, err := itf.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		var observations []campaignCheckpoint
		sequence := ""
		for _, state := range trace.States[1:] {
			observation := campaignProjection(state.Action, state.Vars)
			observations = append(observations, observation)
			sequence += fmt.Sprintf("%s:%d ", state.Action, observation.Target)
		}
		if err = replayBrowserCampaign(trace, observations); err != nil {
			t.Fatal(err)
		}
		sequences[sequence] = true
	}
	if len(sequences) < 2 {
		t.Fatal("generated campaign did not exercise varied schedules")
	}
	t.Logf("%d complete model traces cover %d distinct schedules", len(files), len(sequences))
}
