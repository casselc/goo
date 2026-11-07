package workbench

import (
	"bytes"
	"encoding/json"
	"errors"
	"maps"
	"slices"

	"github.com/casselc/goo/identity"
	"github.com/casselc/goo/measurements"
)

// Native display configuration is separate from measurement calculation identity.
type NativeHistogramViewScope struct {
	Kind   string `json:"kind"`
	Signal string `json:"signal"`
}
type NativeHistogramViewSelection struct {
	Kind string `json:"kind"`
}
type NativeHistogramViewDisplay struct {
	Category         string `json:"category"`
	Value            string `json:"value"`
	RequiresComplete bool   `json:"requiresComplete"`
}
type NativeHistogramViewDeclaration struct {
	Version   string                       `json:"version"`
	Scope     NativeHistogramViewScope     `json:"scope"`
	Selection NativeHistogramViewSelection `json:"selection"`
	Kind      ViewKind                     `json:"kind"`
	Display   NativeHistogramViewDisplay   `json:"display"`
}
type NativeHistogramViewInput struct {
	Version     int                            `json:"version"`
	Title       string                         `json:"title"`
	Question    measurements.Question          `json:"question"`
	Declaration NativeHistogramViewDeclaration `json:"declaration"`
}
type NativeHistogramView struct {
	Version     int                            `json:"version"`
	ID          string                         `json:"id,omitempty"`
	Title       string                         `json:"title"`
	Question    measurements.Question          `json:"question"`
	Declaration NativeHistogramViewDeclaration `json:"declaration"`
}
type NativeHistogramViewReuseRequest struct {
	ID       string                `json:"id"`
	Question measurements.Question `json:"question"`
}

func nativeHistogramQuestion(q measurements.Question) error {
	d := q.Definition
	if d.Kind != "histogram" || d.Grain != "per-stream" || d.Source.Resource != nil || d.HistogramBucketMS != nil || (d.Source.Unit != "s" && d.Source.Unit != "ms") {
		return errors.New("native histogram views require a per-stream service histogram in s or ms without time buckets")
	}
	_, err := measurements.Compile(q)
	return err
}
func (v NativeHistogramViewInput) Validate() error {
	if v.Version != 1 || !text(v.Title, 120, true, false) {
		return errors.New("require a versioned native histogram view and bounded title")
	}
	if err := nativeHistogramQuestion(v.Question); err != nil {
		return err
	}
	d := v.Declaration
	if d.Version != "goo.view/1" || d.Scope != (NativeHistogramViewScope{"measurement-question", "metrics"}) || d.Selection.Kind != "export-interval" || d.Kind.Version != 1 || (d.Kind.Name != "histogram" && d.Kind.Name != "table") || d.Display != (NativeHistogramViewDisplay{"bucket", "observation-count", true}) {
		return errors.New("require a complete-bin histogram or table over an explicitly selected native export interval")
	}
	return nil
}
func (v NativeHistogramView) input() NativeHistogramViewInput {
	return NativeHistogramViewInput{v.Version, v.Title, v.Question, v.Declaration}
}
func (v NativeHistogramView) Validate() error {
	if err := v.input().Validate(); err != nil {
		return err
	}
	if v.ID != "" {
		if !ID.MatchString(v.ID) {
			return errors.New("invalid native histogram view ID")
		}
		want, err := digest(v.input())
		if err != nil {
			return err
		}
		if want != v.ID {
			return errors.New("native histogram view content differs from its ID")
		}
	}
	return nil
}
func strictNativeViewJSON(raw []byte, out any) error {
	if !identity.WellFormedJSON(raw) {
		return errors.New("invalid native histogram view JSON or Unicode")
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	return decoder.Decode(out)
}
func nativeQuestionWire(raw json.RawMessage) error {
	var fields map[string]json.RawMessage
	if json.Unmarshal(raw, &fields) != nil || fields == nil {
		return errors.New("explicit measurement question required")
	}
	for _, key := range []string{"definition", "fromMs", "toMs"} {
		v, ok := fields[key]
		if !ok || bytes.Equal(bytes.TrimSpace(v), []byte("null")) {
			return errors.New("explicit measurement definition and fixed bounds required")
		}
	}
	var definition map[string]json.RawMessage
	if json.Unmarshal(fields["definition"], &definition) != nil {
		return errors.New("explicit measurement definition required")
	}
	var source map[string]json.RawMessage
	if json.Unmarshal(definition["source"], &source) != nil || source == nil {
		return errors.New("explicit native histogram service source required")
	}
	for _, key := range []string{"version", "schemaURL"} {
		var value string
		raw, present := source[key]
		if !present || bytes.Equal(bytes.TrimSpace(raw), []byte("null")) || json.Unmarshal(raw, &value) != nil {
			return errors.New("native source version and schemaURL require explicit strings; empty is valid")
		}
	}
	if raw, present := source["exactStream"]; present && !bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
		var stream map[string]json.RawMessage
		if json.Unmarshal(raw, &stream) != nil || stream == nil {
			return errors.New("invalid exact stream")
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
		return errors.New("native histogram view does not accept resource authority")
	}
	if _, present := definition["histogramBucketMs"]; present {
		return errors.New("native histogram view does not accept time bucket configuration")
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
func (v *NativeHistogramViewInput) UnmarshalJSON(raw []byte) error {
	type wire NativeHistogramViewInput
	var value wire
	if err := strictNativeViewJSON(raw, &value); err != nil {
		return err
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(raw, &fields); err != nil {
		return err
	}
	if err := nativeQuestionWire(fields["question"]); err != nil {
		return err
	}
	*v = NativeHistogramViewInput(value)
	return nil
}
func (v *NativeHistogramView) UnmarshalJSON(raw []byte) error {
	type wire NativeHistogramView
	var value wire
	if err := strictNativeViewJSON(raw, &value); err != nil {
		return err
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(raw, &fields); err != nil {
		return err
	}
	if err := nativeQuestionWire(fields["question"]); err != nil {
		return err
	}
	*v = NativeHistogramView(value)
	return nil
}
func (v *NativeHistogramViewReuseRequest) UnmarshalJSON(raw []byte) error {
	type wire NativeHistogramViewReuseRequest
	var value wire
	if err := strictNativeViewJSON(raw, &value); err != nil {
		return err
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(raw, &fields); err != nil {
		return err
	}
	if err := nativeQuestionWire(fields["question"]); err != nil {
		return err
	}
	*v = NativeHistogramViewReuseRequest(value)
	return nil
}
func detachNativeHistogramView(v NativeHistogramView) (*NativeHistogramView, error) {
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

func (s Store) CreateNativeHistogramView(in NativeHistogramViewInput) (*NativeHistogramView, error) {
	if err := in.Validate(); err != nil {
		return nil, err
	}
	id, err := digest(in)
	if err != nil {
		return nil, err
	}
	v, err := detachNativeHistogramView(NativeHistogramView{in.Version, id, in.Title, in.Question, in.Declaration})
	if err != nil {
		return nil, err
	}
	if err = s.publish("native-histogram-views", id, v); err != nil {
		return nil, err
	}
	return s.LoadNativeHistogramView(id)
}
func (s Store) LoadNativeHistogramView(id string) (*NativeHistogramView, error) {
	var v NativeHistogramView
	if err := s.read("native-histogram-views", id, &v); err != nil {
		return nil, err
	}
	if v.ID != id {
		return nil, errors.New("native histogram view reference differs")
	}
	if err := v.Validate(); err != nil {
		return nil, err
	}
	return &v, nil
}

// Reuse transfers display intent only; the complete destination question wins.
func ReuseNativeHistogramView(source NativeHistogramView, destination measurements.Question) (*NativeHistogramView, error) {
	if err := source.Validate(); err != nil {
		return nil, err
	}
	if err := nativeHistogramQuestion(destination); err != nil {
		return nil, err
	}
	if source.Question.Definition.Source.Unit != destination.Definition.Source.Unit {
		return nil, errors.New("native histogram view reuse requires the exact instrument unit; no conversion")
	}
	return detachNativeHistogramView(NativeHistogramView{Version: 1, Title: source.Title, Question: destination, Declaration: source.Declaration})
}

const NativeHistogramViewPageSize = 20
const NativeHistogramViewPageMaxBytes = 64 << 10

type NativeHistogramViewSummary struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	Kind  string `json:"kind"`
	Unit  string `json:"unit"`
}
type NativeHistogramViewListDiagnostic struct {
	ID     string `json:"id"`
	Reason string `json:"reason"`
}
type NativeHistogramViewList struct {
	Items       []NativeHistogramViewSummary        `json:"items"`
	Partial     bool                                `json:"partial"`
	Diagnostics []NativeHistogramViewListDiagnostic `json:"diagnostics"`
	NextCursor  string                              `json:"nextCursor,omitempty"`
}

// NativeHistogramViews verifies at most one page of candidate artifacts. Cursor
// order is lexical content ID order, not time or a transactional catalog snapshot.
func (s Store) NativeHistogramViews(after string) (*NativeHistogramViewList, error) {
	if after != "" && !ID.MatchString(after) {
		return nil, errors.New("invalid native histogram view cursor")
	}
	ids, err := s.ids("native-histogram-views")
	if err != nil {
		return nil, err
	}
	page := &NativeHistogramViewList{Items: []NativeHistogramViewSummary{}, Diagnostics: []NativeHistogramViewListDiagnostic{}}
	scanned, last := 0, ""
	for _, id := range ids {
		if id <= after {
			continue
		}
		if scanned == NativeHistogramViewPageSize {
			page.NextCursor = last
			break
		}
		scanned++
		last = id
		view, err := s.LoadNativeHistogramView(id)
		if err != nil {
			page.Partial = true
			page.Diagnostics = append(page.Diagnostics, NativeHistogramViewListDiagnostic{id, "unavailable-artifact"})
			continue
		}
		page.Items = append(page.Items, NativeHistogramViewSummary{id, view.Title, view.Declaration.Kind.Name, view.Question.Definition.Source.Unit})
	}
	raw, err := json.Marshal(page)
	if err != nil {
		return nil, err
	}
	if len(raw) > NativeHistogramViewPageMaxBytes {
		return nil, errors.New("native histogram metadata page exceeds 64 KiB")
	}
	return page, nil
}
