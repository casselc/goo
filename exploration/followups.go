package exploration

import (
	"encoding/json"
	"fmt"
	"github.com/casselc/goo/measurements"
	"regexp"
	"strings"
)

// Followups are source-free questions over a verified stored capture. They are
// navigation suggestions, never causal attribution or telemetry completeness.
type Followup struct {
	Transform   *QuestionTransform              `json:"transform,omitempty"`
	ID          string                          `json:"id"`
	Title       string                          `json:"title"`
	Why         string                          `json:"why"`
	Relation    string                          `json:"relation"`
	Preserves   []string                        `json:"preserves"`
	Changes     []string                        `json:"changes"`
	Query       *Query                          `json:"query,omitempty"`
	Measurement *measurements.Question          `json:"measurement,omitempty"`
	Discovery   *measurements.DiscoveryQuestion `json:"discovery,omitempty"`
}
type FollowupObservation struct {
	Index   int    `json:"index"`
	Service string `json:"service"`
	TraceID string `json:"traceId"`
	SpanID  string `json:"spanId"`
	Name    string `json:"name"`
}
type FollowupSet struct {
	Origin       string                `json:"origin"`
	Row          int                   `json:"row"`
	Observations []FollowupObservation `json:"observations"`
	Actions      []Followup            `json:"actions"`
	Warnings     []string              `json:"warnings"`
}

var followupTrace = regexp.MustCompile(`^[a-f0-9]{32}$`)
var followupSpan = regexp.MustCompile(`^[a-f0-9]{16}$`)

func followupEqual(key, value string) Predicate {
	return Predicate{Field: &Field{Kind: "column", Key: key}, Op: "eq", Value: value}
}
func (c SavedCapture) Followups(row int) (*FollowupSet, error) {
	if row < 0 || row >= 100 {
		return nil, fmt.Errorf("select an observation in the first 100 retained records")
	}
	if e := c.Question.Validate(); e != nil {
		return nil, e
	}
	// Detach all query/definition maps from stored evidence before constructing plans.
	raw, _ := json.Marshal(c.Question)
	var origin CaptureQuestion
	json.Unmarshal(raw, &origin)
	out := &FollowupSet{Origin: c.ID, Row: row, Observations: []FollowupObservation{}, Actions: []Followup{}, Warnings: []string{"Suggestions are questions, not findings. Source reads require explicit execution; old evidence stays fixed.", "Observation choices cover at most the first 100 retained records and never define aggregate totals. Missing records are not proof of absence."}}
	addQuery := func(id, title, why, relation string, q Query) {
		b, _ := json.Marshal(q)
		var detached Query
		json.Unmarshal(b, &detached)
		q = detached
		q.Limit = 100
		q.Offset = 0
		out.Actions = append(out.Actions, Followup{ID: id, Title: title, Why: why, Relation: relation, Query: &q})
	}
	addDiscovery := func(service string, from, to int64, environment string) {
		d := measurements.DiscoveryQuestion{Environment: environment, Service: service, FromMS: from, ToMS: to}
		if d.Validate() != nil {
			return
		}
		out.Actions = append(out.Actions, Followup{ID: "resources-" + service, Title: "Discover resource instruments for " + service, Why: "Look for observed instruments before choosing a resource-pressure measurement. Instrument presence alone does not establish saturation or a causal link.", Relation: "Service/time context only", Discovery: &d})
	}
	if origin.Measurement != nil {
		q := *origin.Measurement
		out.Actions = append(out.Actions, Followup{ID: "measurement", Title: "Rerun the same fixed measurement", Why: "Read later-arriving evidence for the exact same definition and window. Capture separately if useful.", Relation: "Same typed question", Measurement: &q})
		for _, service := range q.Definition.Source.Services {
			addDiscovery(service, q.FromMS, q.ToMS, q.Definition.Environment)
		}
	} else if origin.Explore != nil && c.Explore != nil {
		q := *origin.Explore
		if q.EnvironmentSelector != nil {
			out.Warnings = append(out.Warnings, "Native discovery, saved views and reusable packages do not support this explicit environment selector yet.")
		}
		if q.Signal == "spans" {
			derived := q
			derived.Summary = "span-duration"
			derived.SpanStatus = true
			derived.BucketMS = 0
			derived.GroupTimeSeries = false
			derived.GroupBy = []Field{{Kind: "column", Key: "service"}, {Kind: "column", Key: "status_class"}}
			addQuery("duration-status", "Compare duration and recorded status by service", "Test whether latency and recorded error status differ within the original matching span population. Recorded status is not a business-outcome policy.", "Same matching span population", derived)
		}
		for i, b := range c.Explore.Rows {
			if i >= 100 {
				break
			}
			var r struct {
				Service    string            `json:"ServiceName"`
				Trace      string            `json:"TraceId"`
				Span       string            `json:"SpanId"`
				Name       string            `json:"SpanName"`
				Attributes map[string]string `json:"Attributes"`
			}
			if e := json.Unmarshal(b, &r); e != nil {
				return nil, e
			}
			out.Observations = append(out.Observations, FollowupObservation{Index: i, Service: r.Service, TraceID: r.Trace, SpanID: r.Span, Name: r.Name})
			if i != row {
				continue
			}
			if q.Signal == "spans" {
				for _, key := range []string{"goo.operation.outcome", "http.response.status_code", "error.type"} {
					if _, ok := r.Attributes[key]; ok {
						derived := q
						derived.GroupBy = []Field{{Kind: "attribute", Key: key}}
						derived.Summary = ""
						derived.SpanStatus = false
						derived.BucketMS = 0
						derived.GroupTimeSeries = false
						addQuery("reported-"+key, "Break down reported "+key, "Inspect literal reported values, including missing and empty values in the matching population. A reported attribute does not establish an exhaustive outcome policy.", "Same matching span population", derived)
					}
				}
			}
			if followupTrace.MatchString(r.Trace) && r.Trace != strings.Repeat("0", 32) {
				trace := Query{Signal: "spans", Environment: q.Environment, FromMS: q.FromMS, ToMS: q.ToMS, Where: ptrPredicate(followupEqual("trace_id", r.Trace))}
				trace = trace.WithEnvironmentFrom(q)
				addQuery("trace", "Follow this trace across services", "Inspect spans reporting this exact trace ID within the original window. Sampling and out-of-window parents remain unknown.", "Exact reported trace ID", trace)
				logs := trace
				logs.Signal = "logs"
				addQuery("trace-logs", "Inspect logs for this trace", "Check logs carrying the same reported trace ID. Missing logs do not prove that no logging occurred.", "Exact reported trace ID", logs)
				if followupSpan.MatchString(r.Span) && r.Span != strings.Repeat("0", 16) {
					logs.Where = &Predicate{All: []Predicate{followupEqual("trace_id", r.Trace), followupEqual("span_id", r.Span)}}
					addQuery("span-logs", "Inspect logs for this span", "Narrow to logs reporting both selected IDs. This is exact ID correlation, not temporal proximity.", "Exact reported trace and span IDs", logs)
				}
			}
			if r.Service != "" && q.EnvironmentSelector == nil {
				addDiscovery(r.Service, q.FromMS, q.ToMS, q.Environment)
			}
		}
		if row >= len(out.Observations) && row != 0 {
			return nil, fmt.Errorf("retained observation unavailable")
		}
	}
	for i := range out.Actions {
		a := &out.Actions[i]
		var transform *QuestionTransform
		var err error
		if a.Query != nil {
			relation := "same-population"
			if a.ID == "trace" || a.ID == "trace-logs" {
				relation = "exact-trace"
			}
			if a.ID == "span-logs" {
				relation = "exact-span"
			}
			transform, err = DescribeTransform(c.Question, CaptureQuestion{Kind: "explore", Explore: a.Query}, relation)
		} else if a.Measurement != nil {
			transform, err = DescribeTransform(c.Question, CaptureQuestion{Kind: "measurement", Measurement: a.Measurement}, "same-question")
		} else if a.Discovery != nil {
			transform, err = DescribeDiscoveryTransform(c.Question, *a.Discovery)
		}
		if err != nil {
			return nil, err
		}
		a.Transform = transform
		a.Preserves = transform.Preserves
		a.Changes = transform.Replaces

		if a.Query != nil {
			if _, e := Compile(*a.Query); e != nil {
				return nil, e
			}
		}
		if a.Measurement != nil {
			if _, e := measurements.Compile(*a.Measurement); e != nil {
				return nil, e
			}
		}
	}
	return out, nil
}
func ptrPredicate(p Predicate) *Predicate { return &p }
