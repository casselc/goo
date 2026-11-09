package workbench

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"maps"
	"slices"

	"github.com/casselc/goo/identity"
	"github.com/casselc/goo/measurements"
)

// Native display configuration is separate from measurement calculation identity.
type NativeViewScope struct {
	Kind   string `json:"kind"`
	Signal string `json:"signal"`
}
type NativeViewSelection struct {
	Kind string `json:"kind"`
}
type NativeViewDisplay struct {
	Category         string `json:"category"`
	Value            string `json:"value"`
	RequiresComplete bool   `json:"requiresComplete"`
}
type NativeViewDeclaration struct {
	Version   string              `json:"version"`
	Scope     NativeViewScope     `json:"scope"`
	Selection NativeViewSelection `json:"selection"`
	Kind      ViewKind            `json:"kind"`
	Display   NativeViewDisplay   `json:"display"`
}
type NativeViewInput struct {
	Version     int                   `json:"version"`
	Title       string                `json:"title"`
	Question    measurements.Question `json:"question"`
	Declaration NativeViewDeclaration `json:"declaration"`
}
type NativeView struct {
	Version     int                   `json:"version"`
	ID          string                `json:"id,omitempty"`
	Title       string                `json:"title"`
	Question    measurements.Question `json:"question"`
	Declaration NativeViewDeclaration `json:"declaration"`
}
type NativeViewReuseRequest struct {
	ID       string                `json:"id"`
	Question measurements.Question `json:"question"`
}

func nativeViewQuestion(q measurements.Question) error {
	d := q.Definition
	if d.Grain != "per-stream" || d.Source.Resource != nil || d.HistogramBucketMS != nil || d.CounterDomain != "" || d.StateProfile != "" || d.StockDomain != "" || d.Lifecycle != nil {
		return errors.New("native views require an unbucketed per-stream service gauge or histogram")
	}
	if d.Kind != "gauge" && (d.Kind != "histogram" || (d.Source.Unit != "s" && d.Source.Unit != "ms")) {
		return errors.New("native views require a service gauge or a histogram in s or ms")
	}
	_, err := measurements.Compile(q)
	return err
}
func (v NativeViewInput) Validate() error {
	if v.Version != 1 || !text(v.Title, 120, true, false) {
		return errors.New("require a versioned native view and bounded title")
	}
	if err := nativeViewQuestion(v.Question); err != nil {
		return err
	}
	d := v.Declaration
	if d.Version != "goo.view/1" || d.Scope != (NativeViewScope{"measurement-question", "metrics"}) || d.Kind.Version != 1 {
		return errors.New("require a versioned native measurement view")
	}
	if v.Question.Definition.Kind == "gauge" {
		if d.Selection.Kind != "retained-observations" || (d.Kind.Name != "timeline" && d.Kind.Name != "table") || d.Display != (NativeViewDisplay{"time", "recorded-value", false}) {
			return errors.New("require a retained-observation gauge timeline or table")
		}
	} else if d.Selection.Kind != "export-interval" || (d.Kind.Name != "histogram" && d.Kind.Name != "table") || d.Display != (NativeViewDisplay{"bucket", "observation-count", true}) {
		return errors.New("require a complete-bin histogram or table over an explicitly selected native export interval")
	}
	return nil
}
func (v NativeView) input() NativeViewInput {
	return NativeViewInput{v.Version, v.Title, v.Question, v.Declaration}
}
func (v NativeView) Validate() error {
	if err := v.input().Validate(); err != nil {
		return err
	}
	if v.ID != "" {
		if !ID.MatchString(v.ID) {
			return errors.New("invalid native view ID")
		}
		want, err := digest(v.input())
		if err != nil {
			return err
		}
		if want != v.ID {
			return errors.New("native view content differs from its ID")
		}
	}
	return nil
}
func strictNativeViewJSON(raw []byte, out any) error {
	if !identity.WellFormedJSON(raw) {
		return errors.New("invalid native view JSON or Unicode")
	}
	// Canonicalize objects before typed decoding so duplicate object properties
	// replace, rather than merge into, an earlier object's Go fields.
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	var value any
	if err := decoder.Decode(&value); err != nil {
		return err
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		return errors.New("require one native view JSON value")
	}
	canonical, err := json.Marshal(value)
	if err != nil {
		return err
	}
	decoder = json.NewDecoder(bytes.NewReader(canonical))
	decoder.DisallowUnknownFields()
	return decoder.Decode(out)
}
func nativeObject(raw json.RawMessage, allowed ...string) (map[string]json.RawMessage, error) {
	var fields map[string]json.RawMessage
	if json.Unmarshal(raw, &fields) != nil || fields == nil {
		return nil, errors.New("explicit native view object required")
	}
	for key, value := range fields {
		if !slices.Contains(allowed, key) || (key != "resourceSchemaURL" && key != "attributes" && key != "resource" && key != "namespace" && key != "exactStream" && bytes.Equal(bytes.TrimSpace(value), []byte("null"))) {
			return nil, errors.New("unknown or null native view field: " + key)
		}
	}
	return fields, nil
}
func nativeStringMap(raw json.RawMessage) error {
	if len(raw) == 0 || bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
		return nil
	}
	var fields map[string]json.RawMessage
	if json.Unmarshal(raw, &fields) != nil || fields == nil {
		return errors.New("native attributes require a string map")
	}
	for _, raw := range fields {
		var value string
		if bytes.Equal(bytes.TrimSpace(raw), []byte("null")) || json.Unmarshal(raw, &value) != nil {
			return errors.New("native attribute values must be strings")
		}
	}
	return nil
}
func nativeQuestionWire(raw json.RawMessage) error {
	fields, objectErr := nativeObject(raw, "definition", "fromMs", "toMs", "fromNS", "toNS")
	if objectErr != nil {
		return errors.New("explicit measurement question required")
	}
	for _, key := range []string{"definition", "fromMs", "toMs"} {
		v, ok := fields[key]
		if !ok || bytes.Equal(bytes.TrimSpace(v), []byte("null")) {
			return errors.New("explicit measurement definition and fixed bounds required")
		}
	}
	definition, definitionErr := nativeObject(fields["definition"], "apiVersion", "name", "title", "kind", "environment", "environmentSelector", "lookbackMinutes", "maxGapMs", "grain", "source")
	if definitionErr != nil {
		return errors.New("explicit measurement definition required")
	}
	source, sourceErr := nativeObject(definition["source"], "resource", "exactStream", "namespace", "services", "scope", "version", "schemaURL", "metric", "unit", "attributes")
	if sourceErr != nil {
		return errors.New("explicit native service source required")
	}
	if raw, present := source["namespace"]; present && !bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
		if _, err := nativeObject(raw, "mode", "value"); err != nil {
			return err
		}
	}
	if err := nativeStringMap(source["attributes"]); err != nil {
		return err
	}
	for _, key := range []string{"version", "schemaURL"} {
		var value string
		raw, present := source[key]
		if !present || bytes.Equal(bytes.TrimSpace(raw), []byte("null")) || json.Unmarshal(raw, &value) != nil {
			return errors.New("native source version and schemaURL require explicit strings; empty is valid")
		}
	}
	if raw, present := source["exactStream"]; present && !bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
		stream, streamErr := nativeObject(raw, "resource", "resourceSchemaURL", "scope", "version", "schemaURL", "name", "unit", "attributes")
		if streamErr != nil {
			return errors.New("invalid exact stream")
		}
		for _, key := range []string{"resource", "attributes"} {
			if err := nativeStringMap(stream[key]); err != nil {
				return err
			}
		}
		for _, key := range []string{"version", "schemaURL"} {
			var value string
			raw, present := stream[key]
			if !present || bytes.Equal(bytes.TrimSpace(raw), []byte("null")) || json.Unmarshal(raw, &value) != nil {
				return errors.New("exact stream version and schemaURL require explicit strings; empty is valid")
			}
		}
	}
	if _, present := source["resource"]; present {
		return errors.New("native view does not accept resource authority")
	}
	for _, key := range []string{"histogramBucketMs", "counterDomain", "stockDomain", "stateProfile", "lifecycle"} {
		if _, present := definition[key]; present {
			return errors.New("native view does not accept unrelated measurement configuration")
		}
	}
	for _, key := range []string{"fromNS", "toNS"} {
		if raw, ok := fields[key]; ok {
			var value string
			if json.Unmarshal(raw, &value) != nil || value == "" {
				return errors.New("omit unused native bounds rather than empty or null")
			}
		}
	}
	return nil
}

// The authored declaration is a closed wire contract. In particular false is an
// explicit gauge policy, not the zero value of a missing property.
func (d *NativeViewDeclaration) UnmarshalJSON(raw []byte) error {
	type wire NativeViewDeclaration
	var value wire
	if err := strictNativeViewJSON(raw, &value); err != nil {
		return err
	}
	fields, err := nativeObject(raw, "version", "scope", "selection", "kind", "display")
	if err != nil {
		return err
	}
	for _, spec := range []struct {
		key    string
		fields []string
	}{{"scope", []string{"kind", "signal"}}, {"selection", []string{"kind"}}, {"kind", []string{"name", "version"}}, {"display", []string{"category", "value", "requiresComplete"}}} {
		if _, err := nativeObject(fields[spec.key], spec.fields...); err != nil {
			return err
		}
	}
	for _, key := range []string{"version", "scope", "selection", "kind", "display"} {
		if _, ok := fields[key]; !ok {
			return errors.New("explicit native view declaration fields required")
		}
	}
	var display map[string]json.RawMessage
	if err := json.Unmarshal(fields["display"], &display); err != nil {
		return err
	}
	if value, ok := display["requiresComplete"]; !ok || bytes.Equal(bytes.TrimSpace(value), []byte("null")) {
		return errors.New("requiresComplete must be explicitly true or false")
	}
	*d = NativeViewDeclaration(value)
	return nil
}
func (v *NativeViewInput) UnmarshalJSON(raw []byte) error {
	type wire NativeViewInput
	var value wire
	if err := strictNativeViewJSON(raw, &value); err != nil {
		return err
	}
	fields, err := nativeObject(raw, "version", "id", "title", "question", "declaration")
	if err != nil {
		return err
	}
	if err := nativeQuestionWire(fields["question"]); err != nil {
		return err
	}
	*v = NativeViewInput(value)
	return nil
}
func (v *NativeView) UnmarshalJSON(raw []byte) error {
	type wire NativeView
	var value wire
	if err := strictNativeViewJSON(raw, &value); err != nil {
		return err
	}
	fields, err := nativeObject(raw, "version", "id", "title", "question", "declaration")
	if err != nil {
		return err
	}
	if err := nativeQuestionWire(fields["question"]); err != nil {
		return err
	}
	*v = NativeView(value)
	return nil
}
func (v *NativeViewReuseRequest) UnmarshalJSON(raw []byte) error {
	type wire NativeViewReuseRequest
	var value wire
	if err := strictNativeViewJSON(raw, &value); err != nil {
		return err
	}
	fields, err := nativeObject(raw, "id", "question")
	if err != nil {
		return err
	}
	if err := nativeQuestionWire(fields["question"]); err != nil {
		return err
	}
	*v = NativeViewReuseRequest(value)
	return nil
}
func detachNativeView(v NativeView) (*NativeView, error) {
	q := v.Question
	q.Definition.Source.Services = slices.Clone(q.Definition.Source.Services)
	q.Definition.Source.Attributes = maps.Clone(q.Definition.Source.Attributes)
	q.Definition.Source.Namespace = q.Definition.Source.Namespace.Clone()
	if selector := q.Definition.EnvironmentSelector; selector != nil {
		copy := *selector
		q.Definition.EnvironmentSelector = &copy
	}
	if stream := q.Definition.Source.ExactStream; stream != nil {
		copy := *stream
		copy.Resource = maps.Clone(stream.Resource)
		copy.Attributes = maps.Clone(stream.Attributes)
		q.Definition.Source.ExactStream = &copy
	}
	v.Question = q
	return &v, nil
}

func (s Store) CreateNativeView(in NativeViewInput) (*NativeView, error) {
	if err := in.Validate(); err != nil {
		return nil, err
	}
	id, err := digest(in)
	if err != nil {
		return nil, err
	}
	v, err := detachNativeView(NativeView{in.Version, id, in.Title, in.Question, in.Declaration})
	if err != nil {
		return nil, err
	}
	if err = s.publish("native-histogram-views", id, v); err != nil {
		return nil, err
	}
	return s.LoadNativeView(id)
}
func (s Store) LoadNativeView(id string) (*NativeView, error) {
	var v NativeView
	if err := s.read("native-histogram-views", id, &v); err != nil {
		return nil, err
	}
	if v.ID != id {
		return nil, errors.New("native view reference differs")
	}
	if err := v.Validate(); err != nil {
		return nil, err
	}
	return &v, nil
}

// Reuse transfers display intent only; the complete destination question wins.
func ReuseNativeView(source NativeView, destination measurements.Question) (*NativeView, error) {
	if err := source.Validate(); err != nil {
		return nil, err
	}
	if err := nativeViewQuestion(destination); err != nil {
		return nil, err
	}
	if source.Question.Definition.Kind != destination.Definition.Kind || source.Question.Definition.Grain != destination.Definition.Grain {
		return nil, errors.New("native view reuse requires the same measurement family and grain")
	}
	if source.Question.Definition.Source.Unit != destination.Definition.Source.Unit {
		return nil, errors.New("native view reuse requires the exact instrument unit; no conversion")
	}
	return detachNativeView(NativeView{Version: 1, Title: source.Title, Question: destination, Declaration: source.Declaration})
}

const NativeViewPageSize = 20
const NativeViewPageMaxBytes = 64 << 10

type NativeViewSummary struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	Kind      string `json:"kind"`
	Unit      string `json:"unit"`
	Family    string `json:"family"`
	Selection string `json:"selection"`
}
type NativeViewListDiagnostic struct {
	ID     string `json:"id"`
	Reason string `json:"reason"`
}
type NativeViewList struct {
	Items       []NativeViewSummary        `json:"items"`
	Partial     bool                       `json:"partial"`
	Diagnostics []NativeViewListDiagnostic `json:"diagnostics"`
	NextCursor  string                     `json:"nextCursor,omitempty"`
}

// NativeViews verifies at most one page of candidate artifacts. Cursor
// order is lexical content ID order, not time or a transactional catalog snapshot.
func (s Store) NativeViews(after string) (*NativeViewList, error) {
	if after != "" && !ID.MatchString(after) {
		return nil, errors.New("invalid native view cursor")
	}
	ids, err := s.ids("native-histogram-views")
	if err != nil {
		return nil, err
	}
	page := &NativeViewList{Items: []NativeViewSummary{}, Diagnostics: []NativeViewListDiagnostic{}}
	scanned, last := 0, ""
	for _, id := range ids {
		if id <= after {
			continue
		}
		if scanned == NativeViewPageSize {
			page.NextCursor = last
			break
		}
		scanned++
		last = id
		view, err := s.LoadNativeView(id)
		if err != nil {
			page.Partial = true
			page.Diagnostics = append(page.Diagnostics, NativeViewListDiagnostic{id, "unavailable-artifact"})
			continue
		}
		page.Items = append(page.Items, NativeViewSummary{id, view.Title, view.Declaration.Kind.Name, view.Question.Definition.Source.Unit, view.Question.Definition.Kind, view.Declaration.Selection.Kind})
	}
	raw, err := json.Marshal(page)
	if err != nil {
		return nil, err
	}
	if len(raw) > NativeViewPageMaxBytes {
		return nil, errors.New("native view metadata page exceeds 64 KiB")
	}
	return page, nil
}
