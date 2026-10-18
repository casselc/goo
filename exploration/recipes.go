package exploration

import (
	"encoding/json"
	"fmt"
)

// Recipe is a question transform over an explicit captured-record population.
// It neither discovers instrumentation coverage nor evaluates application health.
type Recipe struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Signal   string `json:"signal"`
	Requires string `json:"requires"`
	Limits   string `json:"limits"`
	Next     string `json:"next"`
}

func Recipes() []Recipe {
	return []Recipe{
		{"latency", "Which operations have long spans?", "spans", "Captured spans with valid durations and operation names.", "Span duration includes instrumented work only. Mixed span kinds and retries remain in scope; missing names remain visible. No request-latency or completeness guarantee.", "Rank operations by p95 or maximum, refine a group, then inspect a trace. Compare a previous window only when its population is meaningful."},
		{"errors", "Where are recorded errors concentrated?", "spans", "Captured spans with exporter status; unknown and Unset stay visible.", "Recorded Error spans / all matching captured spans. This is not a business failure rate; retries and mixed span kinds remain in the denominator. Unset is not confirmed success.", "Rank services by recorded error percentage, refine a service and inspect Error evidence. Pin a baseline to investigate a regression."},
		{"logs", "Which log severities are being recorded?", "logs", "Captured logs; exporter severity text is grouped exactly as recorded.", "Severity casing, empty values and missing telemetry are not normalized into health. Counts include duplicates; logs without trace IDs need time context.", "Refine a severity, read messages, then follow a reported trace ID or explicitly open surrounding logs."},
	}
}

type RecipePlan struct {
	Transform *QuestionTransform `json:"transform"`
	Recipe    Recipe             `json:"recipe"`
	Question  Query              `json:"question"`
	Plan      Plan               `json:"plan"`
}

// PlanRecipe preserves signal, environment, fixed bounds and the entire predicate
// tree. It replaces breakdown/measures and resets the bounded inspection page.
func PlanRecipe(id string, q Query) (*RecipePlan, error) {
	if e := q.Validate(); e != nil {
		return nil, e
	}
	var recipe Recipe
	for _, r := range Recipes() {
		if r.ID == id {
			recipe = r
			break
		}
	}
	if recipe.ID == "" {
		return nil, fmt.Errorf("unknown investigation recipe")
	}
	if q.Signal != recipe.Signal {
		return nil, fmt.Errorf("recipe requires an applied %s question; signal and filters are never translated implicitly", recipe.Signal)
	}
	data, e := json.Marshal(q)
	if e != nil {
		return nil, e
	}
	var next Query
	if e = json.Unmarshal(data, &next); e != nil {
		return nil, e
	}
	next.Limit = 100
	next.Offset = 0
	next.Summary = ""
	next.SpanStatus = false
	key := "severity"
	switch id {
	case "latency":
		key = "name"
		next.Summary = "span-duration"
	case "errors":
		key = "service"
		next.SpanStatus = true
	}
	next.GroupBy = []Field{{Kind: "column", Key: key}}
	plan, e := Compile(next)
	if e != nil {
		return nil, e
	}
	transform, e := DescribeTransform(CaptureQuestion{Kind: "explore", Explore: &q}, CaptureQuestion{Kind: "explore", Explore: &next}, "same-population")
	if e != nil {
		return nil, e
	}
	return &RecipePlan{Recipe: recipe, Question: next, Plan: plan, Transform: transform}, nil
}
