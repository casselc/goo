// Generate wire-shape schemas from Go JSON types. This does not generate routes
// or enforce semantic validation. Run from the repository root.
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strings"

	"github.com/casselc/goo/annotations"
	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/investigations"
	"github.com/casselc/goo/measurements"
	"github.com/casselc/goo/prebuilt"
	"github.com/casselc/goo/retention"
	"github.com/casselc/goo/workbench"
)

type schema = map[string]any

var schemas = map[string]schema{}

func nullable(s schema) schema { return schema{"anyOf": []any{s, schema{"type": "null"}}} }
func shape(t reflect.Type, input bool) schema {
	if t == reflect.TypeOf(json.RawMessage{}) {
		return schema{}
	}
	if t.Kind() == reflect.Pointer {
		return nullable(shape(t.Elem(), input))
	}
	if t.PkgPath() == "time" && t.Name() == "Time" {
		return schema{"type": "string", "format": "date-time"}
	}
	if t.Kind() == reflect.Struct && t.Name() != "" {
		pkg := t.PkgPath()
		pkg = pkg[strings.LastIndex(pkg, "/")+1:]
		name := pkg + "." + t.Name()
		if input {
			name += "Input"
		}
		if _, ok := schemas[name]; !ok {
			schemas[name] = schema{}
			schemas[name] = object(t, input)
			schemas[name]["x-go-type"] = t.PkgPath() + "." + t.Name()
		}
		ref := schema{"$ref": "#/components/schemas/" + name}
		if input {
			return nullable(ref)
		}
		return ref
	}
	var s schema
	switch t.Kind() {
	case reflect.Struct:
		s = object(t, input)
	case reflect.String:
		s = schema{"type": "string"}
	case reflect.Bool:
		s = schema{"type": "boolean"}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		s = schema{"type": "integer"}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		s = schema{"type": "integer", "minimum": 0}
	case reflect.Float32, reflect.Float64:
		s = schema{"type": "number"}
	case reflect.Slice, reflect.Array:
		s = schema{"type": "array", "items": shape(t.Elem(), input)}
		if !input && t.Kind() == reflect.Slice {
			return nullable(s)
		}
	case reflect.Map:
		s = schema{"type": "object", "additionalProperties": shape(t.Elem(), input)}
		if !input {
			return nullable(s)
		}
	default:
		return schema{}
	}
	if input {
		return nullable(s)
	}
	return s
}
func object(t reflect.Type, input bool) schema {
	props := map[string]any{}
	required := []string{}
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if f.PkgPath != "" {
			continue
		}
		tag := f.Tag.Get("json")
		parts := strings.Split(tag, ",")
		name := parts[0]
		if name == "-" {
			continue
		}
		if f.Anonymous && name == "" {
			base := f.Type
			if base.Kind() == reflect.Pointer {
				base = base.Elem()
			}
			if base.Kind() == reflect.Struct {
				embedded := object(base, input)
				for k, v := range embedded["properties"].(map[string]any) {
					props[k] = v
				}
				if !input {
					required = append(required, embedded["required"].([]string)...)
				}
				continue
			}
		}
		if name == "" {
			name = f.Name
		}
		props[name] = shape(f.Type, input)
		if !input && !strings.Contains(tag, ",omitempty") {
			required = append(required, name)
		}
	}
	s := schema{"type": "object", "properties": props, "required": required}
	if input {
		delete(s, "required")
		s["additionalProperties"] = false
	}
	return s
}
func root(name string, v any, input bool) { s := shape(reflect.TypeOf(v), input); schemas[name] = s }
func main() {
	inputs := map[string]any{"TraceSearchRequest": exploration.TraceSearchRequest{}, "PoolSignalsRequest": measurements.PoolSignalsRequest{}, "PoolViewRequest": measurements.PoolViewRequest{}, "HostDiscoveryQuestion": measurements.HostDiscoveryQuestion{}, "HostDashboardRequest": measurements.HostDashboardRequest{}, "RecordQuery": exploration.Query{}, "ExplanationRequest": exploration.ExplanationRequest{}, "InventoryRequest": exploration.InventoryQuestion{}, "CompareRequest": exploration.ComparisonRequest{}, "TailSelector": exploration.TailSelector{}, "DashboardRequest": exploration.DashboardRequest{}, "PopulationRequest": exploration.PopulationRequest{}, "NativeQuestion": measurements.Question{}, "MetricDiscoveryRequest": measurements.DiscoveryQuestion{}, "MeaningInput": workbench.MeaningInput{}, "ExplorationView": workbench.ExplorationView{}, "DeploymentInput": workbench.DeploymentInput{}, "DocumentInput": workbench.DocumentInput{}}
	for name, v := range inputs {
		root(name, v, true)
	}
	outputs := map[string]any{"TraceSearchResult": exploration.TraceSearchResult{}, "PoolSignalsPlan": measurements.PoolSignalsPlan{}, "PoolView": measurements.PoolView{}, "HostDiscovery": measurements.HostDiscovery{}, "HostDashboardPlan": measurements.HostDashboardPlan{}, "RecordResult": exploration.Result{}, "CountedValuesResult": exploration.CountedValuesResult{}, "RecordAggregatesResult": exploration.RecordAggregatesResult{}, "ExplanationResult": exploration.ExplanationResult{}, "ExplanationPlan": exploration.ExplanationPlan{}, "InventoryResult": exploration.InventoryResult{}, "ServiceDiscovery": exploration.ServiceDiscovery{}, "CompareResult": exploration.CohortComparison{}, "PackagePlan": exploration.PackagePlan{}, "NumericPlan": exploration.NumericPlan{}, "SavedPackage": exploration.SavedPackage{}, "RecipePlan": exploration.RecipePlan{}, "DashboardPlan": exploration.DashboardPlan{}, "NativeResult": measurements.Result{}, "MetricDiscoveryResult": measurements.DiscoveryResult{}, "CapturePage": exploration.CapturePage{}, "CaptureList": exploration.CaptureList{}, "InvestigationCapture": investigations.Capture{}, "LiveOperations": annotations.LiveOperations{}, "LiveRates": annotations.LiveRates{}, "LiveHistograms": annotations.LiveHistograms{}, "LiveHTTPResult": prebuilt.LiveHTTPResult{}, "MeaningRevision": workbench.MeaningRevision{}, "Document": workbench.Document{}, "Deployment": workbench.Deployment{}, "RetainedReceipt": retention.Receipt{}, "CorrectionReceipt": retention.CorrectionReceipt{}, "RetainedCatalog": retention.CatalogPage{}, "PopulationPlan": exploration.PopulationPlan{}, "NumericHistogramSelection": exploration.NumericHistogramSelection{}, "TailRow": exploration.TailRow{}}
	for name, v := range outputs {
		root(name, v, false)
	}

	root("TailIntent", retention.TailIntent{}, true)
	root("TailReceipt", retention.TailReceipt{}, false)
	root("TailMembershipCapability", exploration.TailMembershipCapability{}, false)
	root("ActivityViewRequest", workbench.ActivityView{}, true)
	root("ActivityView", workbench.ActivityView{}, false)
	root("ActivityViewList", workbench.ActivityViewList{}, false)
	root("EvidenceContent", retention.EvidenceContent{}, false)
	root("EvidenceCatalog", retention.EvidenceCatalogPage{}, false)
	root("MeaningPage", workbench.MeaningPage{}, false)
	root("ExplorationViewResult", workbench.ExplorationView{}, false)
	root("ExplorationViewList", workbench.ExplorationViewList{}, false)
	root("DeploymentContext", workbench.DeploymentContext{}, false)
	root("DocumentList", workbench.DocumentList{}, false)
	root("ComparisonReceipt", workbench.ComparisonReceipt{}, false)
	root("EvidenceComparison", exploration.EvidenceComparison{}, false)
	root("CapturedWindowDiscovery", exploration.CapturedWindowDiscovery{}, false)
	root("FollowupSet", exploration.FollowupSet{}, false)
	root("RecipeList", []exploration.Recipe{}, false)
	root("WorkflowList", []measurements.Workflow{}, false)
	root("OperationPackageList", []*annotations.OperationPlan{}, false)
	root("TelemetryPackageList", []exploration.SavedPackage{}, false)
	root("PrebuiltRegistry", prebuilt.Registry{}, false)
	root("RetainedManifest", struct {
		*retention.Manifest
		ManifestJSON string `json:"manifestJSON"`
	}{}, false)
	root("CaptureReceiptMetadata", struct {
		Receipt            exploration.CaptureReceipt `json:"receipt"`
		DefinitionRevision string                     `json:"definitionRevision"`
		ReadAt             string                     `json:"readAt"`
		Summary            map[string]string          `json:"summary"`
	}{}, false)
	root("LiveComparisonDraft", struct {
		RecordSeries  *exploration.RecordSeriesPair   `json:"recordSeries,omitempty"`
		Token         string                          `json:"token"`
		ExpiresAt     string                          `json:"expiresAt"`
		Comparison    *exploration.EvidenceComparison `json:"comparison"`
		Basis         string                          `json:"basis"`
		BaselineBasis string                          `json:"baselineBasis"`
		Native        *exploration.LiveCounterInfo    `json:"native"`
	}{}, false)
	// Inline request envelopes mirror their handlers. Their nested exported types
	// above are reflected; fixture conformance and route coverage run in CI.
	root("DocumentRequest", struct {
		Parent   string                  `json:"parent"`
		Document workbench.DocumentInput `json:"document"`
	}{}, true)
	root("RecipeRequest", struct {
		ID       string            `json:"id"`
		Question exploration.Query `json:"question"`
	}{}, true)
	root("TelemetryDraftRequest", struct {
		Question exploration.Query `json:"question"`
		Name     string            `json:"name"`
		Title    string            `json:"title"`
		Lookback int               `json:"lookbackMinutes"`
	}{}, true)
	root("DistributionSelectionRequest", struct {
		Comparison exploration.ComparisonRequest `json:"comparison"`
		Side       string                        `json:"side"`
		FromBin    int                           `json:"fromBin"`
		ToBin      int                           `json:"toBin"`
	}{}, true)
	root("NumericHistogramSelectionRequest", struct {
		Parent  exploration.Query `json:"parent"`
		FromBin int               `json:"fromBin"`
		ToBin   int               `json:"toBin"`
	}{}, true)
	root("NumericPlanRequest", struct {
		Question exploration.Query          `json:"question"`
		Binding  exploration.NumericBinding `json:"binding"`
	}{}, true)
	root("TelemetryPlanRequest", struct {
		YAML       string                        `json:"yaml"`
		Definition *exploration.TelemetryPackage `json:"definition"`
		FromMS     int64                         `json:"fromMs"`
		ToMS       int64                         `json:"toMs"`
		Annotation *exploration.AnnotationTarget `json:"annotation"`
	}{}, true)
	root("YAMLRequest", struct {
		YAML string `json:"yaml"`
	}{}, true)
	root("NativePlanRequest", struct {
		YAML   string `json:"yaml"`
		FromMS int64  `json:"fromMs"`
		ToMS   int64  `json:"toMs"`
		FromNS string `json:"fromNS"`
		ToNS   string `json:"toNS"`
	}{}, true)
	root("ExplanationFilterRequest", struct {
		Request exploration.ExplanationRequest `json:"request"`
		Field   exploration.Field              `json:"field"`
		Value   *string                        `json:"value"`
		Side    string                         `json:"side"`
		Mode    string                         `json:"mode"`
	}{}, true)
	root("ExplanationStagedQuestion", struct {
		Question exploration.Query `json:"question"`
	}{}, false)
	root("CaptureRequest", struct {
		Title       string                       `json:"title"`
		Question    *exploration.CaptureQuestion `json:"question,omitempty"`
		Parent      string                       `json:"parent"`
		ResultToken string                       `json:"resultToken,omitempty"`
	}{}, true)
	root("InvestigationRequest", struct {
		Question  investigations.Question `json:"question"`
		Selection json.RawMessage         `json:"selection"`
		Parent    string                  `json:"parent"`
	}{}, true)
	root("LiveComparisonRequest", struct {
		Baseline string `json:"baseline"`
		Freeze   string `json:"freeze"`
		Replace  string `json:"replace"`
	}{}, true)
	root("SaveComparisonRequest", struct {
		Baseline  string                          `json:"baseline"`
		Selection string                          `json:"selection"`
		Expected  *exploration.EvidenceComparison `json:"expected"`
	}{}, true)
	root("RetainRequest", struct {
		RequestID   string   `json:"requestID"`
		CaptureID   string   `json:"captureID"`
		ResultToken string   `json:"resultToken"`
		Indexes     []int    `json:"indexes"`
		Tags        []string `json:"tags"`
		Note        string   `json:"note"`
	}{}, true)
	root("CorrectRequest", struct {
		workbench.MeaningInput
		RequestID  string `json:"requestID"`
		AwaitToken string `json:"awaitToken"`
	}{}, true)
	b, e := json.MarshalIndent(schemas, "", "  ")
	if e != nil {
		panic(e)
	}
	if len(os.Args) != 2 {
		panic("usage: go run ./scripts/openapi-schema output.json")
	}
	if e = os.WriteFile(os.Args[1], append(b, '\n'), 0644); e != nil {
		panic(e)
	}
	fmt.Println("Generated", len(schemas), "wire schemas")
}
