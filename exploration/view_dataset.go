package exploration

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"strconv"
)

// ViewDataset is a supplied, immutable display snapshot, not a source request.
// Its completeness claim concerns this bounded GROUP BY result, never collection
// coverage, raw records, a transactional count/page pair, or upstream sampling.
type ViewDataset struct {
	Version      string            `json:"version"`
	ID           string            `json:"id"`
	Revision     string            `json:"revision"`
	Scope        ViewScope         `json:"scope"`
	Shape        string            `json:"shape"`
	Fields       []ViewField       `json:"fields"`
	Rows         []ViewRow         `json:"rows"`
	Quality      ViewQuality       `json:"quality"`
	Completeness *ViewCompleteness `json:"completeness,omitempty"`
	CapturedAt   string            `json:"capturedAt"`
}
type ViewScope struct {
	ID         string          `json:"id"`
	Generation string          `json:"generation"`
	Signal     string          `json:"signal"`
	Selector   json.RawMessage `json:"selector"`
}
type ViewField struct {
	Key      string `json:"key"`
	Label    string `json:"label"`
	Type     string `json:"type"`
	Unit     string `json:"unit,omitempty"`
	Nullable bool   `json:"nullable,omitempty"`
}
type ViewRow struct {
	Key    string     `json:"key"`
	Values ViewValues `json:"values"`
}
type ViewValues struct {
	Category *string `json:"category"`
	Count    string  `json:"count"`
}
type ViewQuality struct {
	Extent    string `json:"extent"`
	Basis     string `json:"basis"`
	Value     string `json:"value"`
	Sampling  string `json:"sampling"`
	Freshness string `json:"freshness"`
}
type ViewCompleteness struct {
	Basis    string   `json:"basis"`
	Revision string   `json:"revision"`
	ScopeID  string   `json:"scopeId"`
	Aspects  []string `json:"aspects"`
}

// groupViewDataset is called only after the source read has completed and its
// LIMIT1001 sentinel has rejected more than1000 groups. Matching a separate count
// or a page's row count never establishes this attestation.
func groupViewDataset(r *Result, exhaustiveGroups bool) *ViewDataset {
	q := r.Question
	if len(q.GroupBy) != 1 || q.Numeric != nil || q.NumericHistogram != nil || q.Summary != "" || q.SpanStatus || q.GroupTimeSeries || q.BucketMS != 0 || len(r.Groups) > 1000 || !regexpRevision.MatchString(r.Basis) {
		return nil
	}
	selector, err := json.Marshal(q)
	if err != nil {
		return nil
	}
	scopeHash := sha256.Sum256(selector)
	scopeID := hex.EncodeToString(scopeHash[:])
	d := &ViewDataset{Version: "goo.dataset/1", ID: scopeID + ":groups", Revision: r.Basis, Scope: ViewScope{ID: scopeID, Generation: r.Basis, Signal: q.Signal, Selector: selector}, Shape: "ranked", Fields: []ViewField{{Key: "category", Label: q.GroupBy[0].Kind + ":" + q.GroupBy[0].Key, Type: "string", Nullable: true}, {Key: "count", Label: "Captured records", Type: "number", Unit: "records"}}, Rows: []ViewRow{}, Quality: ViewQuality{Extent: "unknown", Basis: "captured-query", Value: "derived", Sampling: "unknown", Freshness: "fixed"}, CapturedAt: r.CapturedAt}
	seen := map[string]bool{}
	for _, raw := range r.Groups {
		var g map[string]json.RawMessage
		if json.Unmarshal(raw, &g) != nil {
			return nil
		}
		value, ok := g["g0"]
		if !ok {
			return nil
		}
		var category *string
		var count string
		if json.Unmarshal(value, &category) != nil || json.Unmarshal(g["count"], &count) != nil || category != nil && len(*category) > 1024 {
			return nil
		}
		n, err := strconv.ParseUint(count, 10, 64)
		if err != nil || strconv.FormatUint(n, 10) != count {
			return nil
		}
		canonical, _ := json.Marshal(category)
		categoryHash := sha256.Sum256(canonical)
		key := "category:" + hex.EncodeToString(categoryHash[:])
		if seen[key] {
			return nil
		}
		seen[key] = true
		d.Rows = append(d.Rows, ViewRow{Key: key, Values: ViewValues{Category: category, Count: count}})
	}
	if exhaustiveGroups && r.State != "changed_between_reads" {
		d.Quality.Extent = "complete"
		d.Completeness = &ViewCompleteness{Basis: "captured-query", Revision: r.Basis, ScopeID: scopeID, Aspects: []string{"groups"}}
	}
	return d
}
