package verification

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/identity"
	"github.com/casselc/goo/querycel"
	"github.com/casselc/goo/workbench"
	"hegel.dev/go/hegel"
)

// Exercise the actual definition/store/YAML owners while independently computing
// the predicate digest. Authored bytes and admitted meaning have different IDs.
func TestHegelCELPreservation436RoundTrip(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		signal := hegel.Draw(tc, hegel.SampledFrom([]string{"logs", "spans"}))
		literal := hegel.Draw(tc, hegel.SampledFrom([]string{"", "needle雪🦀", "x' OR 1=1 --", "quote\"\\\nline"}))
		threshold := hegel.Draw(tc, hegel.SampledFrom([]uint64{0, 17, 255, 9007199254740993, ^uint64(0)}))
		seed := hegel.Draw(tc, hegel.Integers(1000, 100000))
		base := exploration.Query{Signal: signal, Environment: "prod", FromMS: int64(seed), ToMS: int64(seed + 60000), Limit: hegel.Draw(tc, hegel.Integers(1, 100)), Where: &exploration.Predicate{Any: []exploration.Predicate{
			{Field: &exploration.Field{Kind: "attribute", Key: "tenant"}, Op: "missing"},
			{Field: &exploration.Field{Kind: "resource", Key: "region"}, Op: "eq", Value: "west"},
		}}}
		if hegel.Draw(tc, hegel.Booleans()) {
			base.Environment = ""
			base.EnvironmentSelector = &identity.EnvironmentSelector{Version: identity.EnvironmentSelectorVersion, Policy: "effective", Scope: hegel.Draw(tc, hegel.SampledFrom([]identity.EnvironmentScope{{Mode: "named", Value: "prod"}, {Mode: "empty"}, {Mode: "unreported"}}))}
		}
		if hegel.Draw(tc, hegel.Booleans()) {
			base.GroupBy = []exploration.Field{{Kind: "column", Key: "service"}}
		}
		textvar, numvar := "log.body", "log.severity_number"
		if signal == "spans" {
			textvar, numvar = "span.name", "span.duration_ns"
			base.Summary = "span-duration"
			base.SpanStatus = true
		} else {
			base.LogSeverity = true
		}
		source := textvar + ".contains(" + celLiteral(literal) + ") && " + numvar + " >= " + strconv.FormatUint(threshold, 10) + "u"
		request, err := querycel.NewPlanRequest(base, source, "condition-雪", "mount-1", "9007199254740993")
		if err != nil {
			tc.Fatal(err)
		}
		plan, err := querycel.PlanRecord(request)
		if err != nil {
			tc.Fatal(err)
		}
		authored, err := workbench.NewRecordAuthoring(request, "parent-18446744073709551615")
		if err != nil {
			tc.Fatal(err)
		}
		raw, err := json.Marshal(plan.Predicate)
		if err != nil {
			tc.Fatal(err)
		}
		sum := sha256.Sum256(raw)
		if authored.LoweringDigest != hex.EncodeToString(sum[:]) || authored.Source != source || authored.Binding != request.Binding || !reflect.DeepEqual(authored.Base, base) {
			tc.Fatal("source/binding/base/digest differs from independently checked input")
		}
		view := workbench.ExplorationView{Version: 1, Title: "Authored record condition", Queries: []exploration.Query{plan.Question}, Layout: "records", Columns: []string{"utc", "service", "summary"}, Authoring: map[string]workbench.RecordAuthoring{signal: authored}}
		if err = view.Validate(); err != nil {
			tc.Fatal(err)
		}
		store := workbench.Store{Dir: tc.TempDir()}
		saved, err := store.CreateExplorationView(view)
		if err != nil {
			tc.Fatal(err)
		}
		loaded, err := store.LoadExplorationView(saved.ID)
		if err != nil {
			tc.Fatal(err)
		}
		loaded.ID = ""
		if !reflect.DeepEqual(*loaded, view) {
			tc.Fatal("stored definition lost exact authoring or ordinary meaning")
		}
		artifact := workbench.ViewArtifact{Version: workbench.ViewArtifactVersion, Namespace: "team", Name: "checked-condition", Revision: "v1", View: *loaded}
		encoded, err := workbench.EncodeViewArtifact(artifact)
		if err != nil {
			tc.Fatal(err)
		}
		decoded, err := workbench.DecodeViewArtifact(encoded.YAML)
		if err != nil {
			tc.Fatal(err)
		}
		if !reflect.DeepEqual(decoded.Artifact, artifact) || decoded.YAML != encoded.YAML {
			tc.Fatal("portable YAML lost exact source or contract")
		}
		imported, err := store.CreateExplorationView(decoded.Artifact.View)
		if err != nil || imported.ID != saved.ID {
			tc.Fatal("portable import changed saved content identity", err)
		}
		formatted := request
		formatted.Expression = "// retained user comment 雪\n(\n  " + source + "\n) // end"
		second, err := workbench.NewRecordAuthoring(formatted, authored.ParentRevision)
		if err != nil {
			tc.Fatal(err)
		}
		if second.LoweringDigest != authored.LoweringDigest || second.Source == authored.Source {
			tc.Fatal("formatting and lowered meaning were conflated")
		}
		formattedView := preservation436Clone(view)
		formattedView.Authoring[signal] = second
		formattedSaved, err := store.CreateExplorationView(formattedView)
		if err != nil {
			tc.Fatal(err)
		}
		if formattedSaved.ID == saved.ID || !reflect.DeepEqual(formattedSaved.Queries, saved.Queries) {
			tc.Fatal("exact source must alter definition identity without changing the question")
		}
		// The retained original revision cannot be changed by detached read/codec edits.
		entry := loaded.Authoring[signal]
		entry.Base.Where.Any[0].Op = "exists"
		loaded.Authoring[signal] = entry
		again, err := store.LoadExplorationView(saved.ID)
		if err != nil || !reflect.DeepEqual(again.Authoring[signal], authored) {
			tc.Fatal("detached reopen altered stored original", err)
		}
	}, options(80)...)
}

func preservation436Clone(view workbench.ExplorationView) workbench.ExplorationView {
	raw, err := json.Marshal(view)
	if err != nil {
		panic(err)
	}
	var out workbench.ExplorationView
	if err = json.Unmarshal(raw, &out); err != nil {
		panic(err)
	}
	return out
}

// Each mutation is independently refused by all existing admission/publication
// boundaries, and does not create even an empty store directory.
func TestCELPreservation436RejectsTamperedContracts(t *testing.T) {
	base := exploration.Query{Signal: "logs", Environment: "prod", FromMS: 1000, ToMS: 61000, Limit: 20, Where: &exploration.Predicate{Field: &exploration.Field{Kind: "resource", Key: "region"}, Op: "eq", Value: "west"}}
	request, err := querycel.NewPlanRequest(base, `log.body.contains("needle")`, "condition", "mount", "1")
	if err != nil {
		t.Fatal(err)
	}
	plan, err := querycel.PlanRecord(request)
	if err != nil {
		t.Fatal(err)
	}
	authored, err := workbench.NewRecordAuthoring(request, "parent-1")
	if err != nil {
		t.Fatal(err)
	}
	good := workbench.ExplorationView{Version: 1, Title: "Checked definition", Queries: []exploration.Query{plan.Question}, Layout: "records", Columns: []string{"utc", "summary"}, Authoring: map[string]workbench.RecordAuthoring{"logs": authored}}
	cases := []struct {
		name string
		edit func(*workbench.ExplorationView)
	}{
		{"source", func(v *workbench.ExplorationView) {
			a := v.Authoring["logs"]
			a.Source = `log.body == "changed"`
			v.Authoring["logs"] = a
		}},
		{"base", func(v *workbench.ExplorationView) {
			a := v.Authoring["logs"]
			a.Base.Environment = "foreign"
			v.Authoring["logs"] = a
		}},
		{"query", func(v *workbench.ExplorationView) { v.Queries[0].ToMS++ }},
		{"old-library", func(v *workbench.ExplorationView) {
			a := v.Authoring["logs"]
			a.Binding.Library = "goo.cel.otel/v0"
			v.Authoring["logs"] = a
		}},
		{"compiler", func(v *workbench.ExplorationView) {
			a := v.Authoring["logs"]
			a.Binding.Compiler = "other"
			v.Authoring["logs"] = a
		}},
		{"base-identity", func(v *workbench.ExplorationView) {
			a := v.Authoring["logs"]
			a.Binding.BaseIdentity = strings.Repeat("0", 64)
			v.Authoring["logs"] = a
		}},
		{"revision", func(v *workbench.ExplorationView) {
			a := v.Authoring["logs"]
			a.Binding.Revision = ""
			v.Authoring["logs"] = a
		}},
		{"parent-revision", func(v *workbench.ExplorationView) {
			a := v.Authoring["logs"]
			a.ParentRevision = ""
			v.Authoring["logs"] = a
		}},
		{"digest", func(v *workbench.ExplorationView) {
			a := v.Authoring["logs"]
			a.LoweringDigest = strings.Repeat("0", 64)
			v.Authoring["logs"] = a
		}},
		{"representation", func(v *workbench.ExplorationView) {
			a := v.Authoring["logs"]
			a.Representation = "sql"
			v.Authoring["logs"] = a
		}},
		{"version", func(v *workbench.ExplorationView) {
			a := v.Authoring["logs"]
			a.Version = "goo.cel-authoring/0"
			v.Authoring["logs"] = a
		}},
		{"foreign-signal-key", func(v *workbench.ExplorationView) {
			v.Authoring["spans"] = v.Authoring["logs"]
			delete(v.Authoring, "logs")
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			bad := preservation436Clone(good)
			tc.edit(&bad)
			if err := bad.Validate(); err == nil {
				t.Fatal("tampered authoring admitted")
			}
			root := filepath.Join(t.TempDir(), "uncreated")
			if out, err := (workbench.Store{Dir: root}).CreateExplorationView(bad); err == nil || out != nil {
				t.Fatal("tampered definition published", err)
			}
			if _, err := os.Stat(root); !os.IsNotExist(err) {
				t.Fatal("rejected definition created store state", err)
			}
			artifact := workbench.ViewArtifact{Version: workbench.ViewArtifactVersion, Namespace: "team", Name: "checked-condition", Revision: "v1", View: bad}
			if out, err := workbench.EncodeViewArtifact(artifact); err == nil || out != nil {
				t.Fatal("tampered artifact encoded", err)
			}
			raw, err := json.Marshal(artifact)
			if err != nil {
				t.Fatal(err)
			}
			// JSON is a YAML subset; this exercises the actual strict import decoder.
			if out, err := workbench.DecodeViewArtifact(string(raw)); err == nil || out != nil {
				t.Fatal("tampered portable artifact decoded", err)
			}
			files := 0
			_ = filepath.WalkDir(root, func(_ string, d fs.DirEntry, err error) error {
				if err == nil && !d.IsDir() {
					files++
				}
				return nil
			})
			if files != 0 {
				t.Fatal("refusal wrote files")
			}
		})
	}
}
