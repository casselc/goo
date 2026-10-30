// Package scopedview contains source-neutral supplied-display wire types.
package scopedview

import "encoding/json"

// Dataset is a supplied immutable display snapshot, not a source request.
// Completeness concerns its named result aspect, never upstream collection,
// pagination or a transactional count/page pair.
type Dataset struct {
	Version      string        `json:"version"`
	ID           string        `json:"id"`
	Revision     string        `json:"revision"`
	Scope        Scope         `json:"scope"`
	Shape        string        `json:"shape"`
	Fields       []Field       `json:"fields"`
	Rows         []Row         `json:"rows"`
	Quality      Quality       `json:"quality"`
	Completeness *Completeness `json:"completeness,omitempty"`
	CapturedAt   string        `json:"capturedAt,omitempty"`
	Origin       string        `json:"origin"`
	SourceBasis  string        `json:"sourceBasis,omitempty"`
	Interval     *Interval     `json:"interval,omitempty"`
	Dimensions   []string      `json:"dimensions,omitempty"`
	Partition    *Partition    `json:"partition,omitempty"`
	Histogram    *Histogram    `json:"histogram,omitempty"`
}

// Partition describes the disjoint captured-record groups in the same
// exhaustive GROUP BY read. It is independent of the separately read count and
// row page. Consumers still need source-admission authority to claim full shares.
type Partition struct {
	Kind    string `json:"kind"`
	Measure string `json:"measure"`
	Total   string `json:"total"`
}

type Interval struct {
	FromMS   string `json:"fromMs"`
	ToMS     string `json:"toMs"`
	BucketMS string `json:"bucketMs"`
}
type Scope struct {
	ID         string          `json:"id"`
	Generation string          `json:"generation"`
	Signal     string          `json:"signal"`
	Selector   json.RawMessage `json:"selector"`
}
type Field struct {
	Key         string       `json:"key"`
	Label       string       `json:"label"`
	Type        string       `json:"type"`
	Unit        string       `json:"unit,omitempty"`
	Nullable    bool         `json:"nullable,omitempty"`
	Aggregation *Aggregation `json:"aggregation,omitempty"`
}
type Row struct {
	Key    string          `json:"key"`
	Values Values          `json:"values"`
	State  string          `json:"state,omitempty"`
	Cells  map[string]Cell `json:"cells,omitempty"`
}

// Values retains exact source-derived text and explicit missing values.
// The declared fields define meaning; there is no renderer-specific payload.
type Values map[string]*string
type Quality struct {
	Extent    string `json:"extent"`
	Basis     string `json:"basis"`
	Value     string `json:"value"`
	Sampling  string `json:"sampling"`
	Freshness string `json:"freshness"`
}
type Completeness struct {
	Basis    string   `json:"basis"`
	Revision string   `json:"revision"`
	ScopeID  string   `json:"scopeId"`
	Aspects  []string `json:"aspects"`
}

// Aggregation describes the exact operation behind a supplied measure. It
// is not a client-side reducer; quantiles cannot be combined by averaging.
type Aggregation struct {
	Family      string `json:"family"`
	Kind        string `json:"kind"`
	Method      string `json:"method"`
	Quantile    string `json:"quantile,omitempty"`
	Numerator   string `json:"numerator,omitempty"`
	Denominator string `json:"denominator,omitempty"`
}
type Cell struct {
	State   string   `json:"state"`
	Support *Support `json:"support,omitempty"`
}
type Support struct {
	Total   string `json:"total"`
	Valid   string `json:"valid"`
	Missing string `json:"missing"`
	Empty   string `json:"empty"`
	Invalid string `json:"invalid"`
}

// Histogram discriminates numeric values, raw-span durations and native
// instrument observations. Numeric Support retains parsing coverage; the other
// families use their own captured population totals without parsing metadata.
type Histogram struct {
	Kind        string         `json:"kind"`
	Unit        string         `json:"unit,omitempty"`
	Total       string         `json:"total,omitempty"`
	Binning     string         `json:"binning"`
	Binding     NumericBinding `json:"binding,omitzero"`
	Support     Support        `json:"support,omitzero"`
	Closure     string         `json:"closure"`
	Grain       string         `json:"grain,omitempty"`
	Temporality string         `json:"temporality,omitempty"`
}

// NumericBinding is display metadata, not a source query or parser.
type NumericBinding struct {
	Scale int          `json:"scale,omitempty"`
	Field NumericField `json:"field"`
	Type  string       `json:"type"`
	Unit  string       `json:"unit"`
	Parse string       `json:"parse"`
}
type NumericField struct {
	Kind string `json:"kind"`
	Key  string `json:"key"`
}
