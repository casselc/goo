package verification

import (
	"bytes"
	"encoding/json"
	"os/exec"
	"reflect"
	"strconv"
	"testing"

	"github.com/casselc/goo/identity"
	"github.com/casselc/goo/measurements"
	"github.com/casselc/goo/workbench"
	"hegel.dev/go/hegel"
)

// Configuration identity is independent from the calculation question: changing
// display changes the saved artifact, not a measurement revision or destination.
// Cross-runtime projection is compared to complete independent destination data.
func TestHegelNativeHistogramReuse371(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		seed := hegel.Draw(tc, hegel.Integers(1, 1000000))
		mode := hegel.Draw(tc, hegel.Integers(0, 3))
		nsMode := hegel.Draw(tc, hegel.Integers(0, 2))
		var ns *identity.Namespace
		if nsMode > 0 {
			ns = &identity.Namespace{Mode: "value"}
			if nsMode == 2 {
				ns.Value = "team-雪-" + strconv.Itoa(seed)
			}
		}
		origin, _ := nrcInputs(nil, "histogram")
		origin.Definition.Source.Unit = "s"
		destination, stream := nrcInputs(ns, "histogram")
		d := &destination.Definition
		d.Environment = ""
		d.EnvironmentSelector = &identity.EnvironmentSelector{Version: identity.EnvironmentSelectorVersion, Policy: "effective", Scope: identity.EnvironmentScope{Mode: "named", Value: "destination-" + strconv.Itoa(seed)}}
		if mode == 0 {
			d.EnvironmentSelector.Policy = "current-only"
		} else if mode == 2 {
			d.EnvironmentSelector.Scope = identity.EnvironmentScope{Mode: "empty"}
		} else if mode == 3 {
			d.EnvironmentSelector.Scope = identity.EnvironmentScope{Mode: "unreported"}
		}
		d.Name = "destination-histogram"
		d.Title = "Destination observations"
		d.Source.Services = []string{"destination-worker"}
		d.Source.Scope = "otel.destination"
		d.Source.Version = "2"
		d.Source.SchemaURL = "https://opentelemetry.io/schemas/1.37.0"
		d.Source.Metric = "destination.duration"
		d.Source.Unit = "s"
		d.Source.Attributes = map[string]string{"metric.only": "literal-Ω-" + strconv.Itoa(seed)}
		stream.Scope, stream.Version, stream.SchemaURL, stream.Name, stream.Unit = d.Source.Scope, d.Source.Version, d.Source.SchemaURL, d.Source.Metric, d.Source.Unit
		stream.Resource = map[string]string{"service.name": "destination-worker", "service.instance.id": "destination-instance"}
		stream.Attributes = map[string]string{"metric.only": d.Source.Attributes["metric.only"]}
		if ns != nil && ns.Mode == "value" {
			stream.Resource["service.namespace"] = ns.Value
		}
		if mode != 3 {
			stream.Resource["deployment.environment.name"] = d.EnvironmentSelector.Scope.Value
		}
		d.Source.ExactStream = &stream
		destination.FromNS = strconv.FormatInt(destination.FromMS*1000000+int64(hegel.Draw(tc, hegel.Integers(0, 999999))), 10)
		destination.ToNS = strconv.FormatInt(destination.ToMS*1000000-int64(hegel.Draw(tc, hegel.Integers(0, 999999))), 10)
		decl := workbench.NativeViewDeclaration{Version: "goo.view/1", Scope: workbench.NativeViewScope{Kind: "measurement-question", Signal: "metrics"}, Selection: workbench.NativeViewSelection{Kind: "export-interval"}, Kind: workbench.ViewKind{Name: "histogram", Version: 1}, Display: workbench.NativeViewDisplay{Category: "bucket", Value: "observation-count", RequiresComplete: true}}
		store := workbench.Store{Dir: t.TempDir()}
		source, err := store.CreateNativeView(workbench.NativeViewInput{Version: 1, Title: "Reusable native bins", Question: origin, Declaration: decl})
		if err != nil {
			tc.Fatal(err)
		}
		decl.Kind.Name = "table"
		table, err := store.CreateNativeView(workbench.NativeViewInput{Version: 1, Title: source.Title, Question: origin, Declaration: decl})
		if err != nil || source.ID == table.ID || !reflect.DeepEqual(source.Question, table.Question) {
			tc.Fatal("display artifact changed measurement identity", err)
		}
		if hegel.Draw(tc, hegel.Integers(0, 1)) == 1 {
			source = table
		}
		before, err := measurements.Compile(destination)
		if err != nil {
			tc.Fatal(err)
		}
		reused, err := workbench.ReuseNativeView(*source, destination)
		if err != nil || reused.ID != "" || !reflect.DeepEqual(reused.Question, destination) || !reflect.DeepEqual(reused.Declaration, source.Declaration) {
			tc.Fatal("native Go reuse changed destination or display", err)
		}
		after, err := measurements.Compile(reused.Question)
		if err != nil || before.Revision != after.Revision || before.SQL != after.SQL || !reflect.DeepEqual(before.Parameters, after.Parameters) {
			tc.Fatal("native display altered calculation revision or plan", err)
		}
		encoded, err := json.Marshal(map[string]any{"source": source, "destination": destination})
		if err != nil {
			tc.Fatal(err)
		}
		command := exec.CommandContext(t.Context(), "node", "native_histogram_reuse_371_fixture.mjs")
		command.Stdin = bytes.NewReader(encoded)
		out, err := command.CombinedOutput()
		if err != nil {
			tc.Fatalf("native display cross-runtime fixture: %v\n%s", err, out)
		}
		var actual struct {
			Reused    workbench.NativeView
			Refusals  int
			Unchanged bool
		}
		if err = json.Unmarshal(out, &actual); err != nil || !actual.Unchanged || actual.Refusals != 11 || !reflect.DeepEqual(actual.Reused, *reused) {
			tc.Fatalf("cross-runtime display/destination authority changed: %v\n%s", err, out)
		}
	}, options(40)...)
}
