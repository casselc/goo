package exploration

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

const CaptureSearchMaxBytes int64 = 16 << 20
const CaptureSearchSummaryBytes = 8 << 20

var ErrCaptureSearchBudget = errors.New("capture search scan budget exceeded; open a capture by ID instead")
var ErrCaptureSearchCursor = errors.New("capture search cursor does not match the query or catalog; restart search")

type CaptureSearchQuery struct {
	Title       string  `json:"title"`
	Kind        string  `json:"kind"`
	Signal      string  `json:"signal"`
	Environment *string `json:"environment"`
	Order       string  `json:"order"`
}

// CaptureSummary is bounded discovery metadata, not a partial evidence receipt.
// Full questions and verified evidence are available only through exact-ID reads.
type CaptureSummary struct {
	ID              string  `json:"id"`
	Parent          string  `json:"parent,omitempty"`
	Title           string  `json:"title"`
	CapturedAt      string  `json:"capturedAt"`
	Kind            string  `json:"kind"`
	MeasurementKind string  `json:"measurementKind,omitempty"`
	Signal          string  `json:"signal"`
	Environment     *string `json:"environment"`
}

// canonicalCaptureTime bounds even legacy timestamps with long fractional tails.
// Preserve a boundary-year offset when converting it to UTC would leave RFC3339.
func canonicalCaptureTime(value string) string {
	at, _ := time.Parse(time.RFC3339Nano, value)
	utc := at.UTC()
	if utc.Year() >= 0 && utc.Year() <= 9999 {
		return utc.Format(time.RFC3339Nano)
	}
	return at.Format(time.RFC3339Nano)
}

func summarizeCapture(r CaptureReceipt) CaptureSummary {
	signal, environment, known := captureScope(r.Question)
	summary := CaptureSummary{ID: r.ID, Parent: r.Parent, Title: r.Title, CapturedAt: canonicalCaptureTime(r.CapturedAt), Kind: r.Question.Kind, Signal: signal}
	if known {
		summary.Environment = &environment
	}
	if r.Question.Measurement != nil {
		summary.MeasurementKind = r.Question.Measurement.Definition.Kind
	}
	return summary
}

type CaptureSearchPage struct {
	Version    string           `json:"version"`
	Items      []CaptureSummary `json:"items"`
	HasMore    bool             `json:"hasMore"`
	NextCursor string           `json:"nextCursor,omitempty"`
	Matched    int              `json:"matched"`
	Total      int              `json:"total"`
	Order      string           `json:"order"`
}

func (q CaptureSearchQuery) normalized() (CaptureSearchQuery, error) {
	if q.Order == "" {
		q.Order = "newest"
	}
	valid := func(s string, n int) bool {
		return utf8.ValidString(s) && len(s) <= n && strings.IndexFunc(s, unicode.IsControl) < 0
	}
	if !valid(q.Title, 120) || q.Environment != nil && !valid(*q.Environment, 256) {
		return q, errors.New("invalid capture title or environment search")
	}
	if !slices.Contains([]string{"", "explore", "inventory", "comparison", "explanation", "measurement"}, q.Kind) || !slices.Contains([]string{"", "logs", "spans", "metrics"}, q.Signal) || !slices.Contains([]string{"newest", "oldest"}, q.Order) {
		return q, errors.New("invalid capture kind, signal or capture-time order")
	}
	return q, nil
}

// Scope is reported only when the receipt question supports an unambiguous value.
func captureScope(q CaptureQuestion) (signal, environment string, knownEnvironment bool) {
	scope, err := q.Scope()
	if err != nil {
		return "", "", false
	}
	signal, environment, knownEnvironment = scope.Signal, scope.Environment, true
	for _, member := range scope.Members {
		if member.Signal != scope.Signal {
			signal = ""
		}
		if member.Environment != scope.Environment {
			knownEnvironment = false
		}
	}
	return signal, environment, knownEnvironment
}
func (q CaptureSearchQuery) matches(r CaptureSummary) bool {
	if q.Kind != "" && r.Kind != q.Kind || !strings.Contains(strings.ToLower(r.Title), strings.ToLower(q.Title)) {
		return false
	}
	if q.Signal == "" && q.Environment == nil {
		return true
	}
	return (q.Signal == "" || q.Signal == r.Signal) && (q.Environment == nil || r.Environment != nil && *q.Environment == *r.Environment)
}

type captureSearchCursor struct {
	Query   string `json:"query"`
	Catalog string `json:"catalog"`
	After   string `json:"after"`
}

func searchToken(c captureSearchCursor) string {
	b, _ := json.Marshal(c)
	return base64.RawURLEncoding.EncodeToString(b)
}
func searchHash(v any) string {
	b, _ := json.Marshal(v)
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}

// Search verifies metadata for a whole bounded catalog before filtering/paging. It performs no telemetry reads.
func (s CaptureStore) Search(ctx context.Context, q CaptureSearchQuery, cursor string) (*CaptureSearchPage, error) {
	q, err := q.normalized()
	if err != nil {
		return nil, err
	}
	if len(cursor) > 1024 {
		return nil, ErrCaptureSearchCursor
	}
	var position captureSearchCursor
	if cursor != "" {
		b, e := base64.RawURLEncoding.DecodeString(cursor)
		if e != nil || json.Unmarshal(b, &position) != nil || searchToken(position) != cursor || position.Query != searchHash(q) || !regexpRevision.MatchString(position.After) {
			return nil, ErrCaptureSearchCursor
		}
	}
	result := &CaptureSearchPage{Version: "goo/capture-search/v2", Items: []CaptureSummary{}, Order: q.Order}
	if err = ctx.Err(); err != nil {
		return nil, err
	}
	directory, err := os.Open(s.Dir)
	if errors.Is(err, os.ErrNotExist) {
		if cursor != "" {
			return nil, ErrCaptureSearchCursor
		}
		return result, nil
	}
	if err != nil {
		return nil, err
	}
	defer directory.Close()
	entries, err := captureDirectoryEntries(directory)
	if err != nil {
		return nil, err
	}
	ids := []string{}
	var size int64
	for _, entry := range entries {
		if err = ctx.Err(); err != nil {
			return nil, err
		}
		name := entry.Name()
		id := strings.TrimSuffix(name, ".json")
		if !strings.HasSuffix(name, ".json") || !regexpRevision.MatchString(id) {
			continue
		}
		info, e := entry.Info()
		if e != nil {
			return nil, e
		}
		if !info.Mode().IsRegular() {
			return nil, fmt.Errorf("capture catalog unsupported file for %s", id)
		}
		metadata, e := os.Lstat(filepath.Join(s.catalogDir(), id+".json"))
		if e != nil {
			return nil, ErrCaptureCatalogUnavailable
		}
		if !metadata.Mode().IsRegular() {
			return nil, fmt.Errorf("capture metadata unsupported file for %s", id)
		}
		size += metadata.Size()
		if size > CaptureSearchMaxBytes {
			return nil, ErrCaptureSearchBudget
		}
		ids = append(ids, id)
	}
	slices.Sort(ids)
	catalogParts := []string{}
	result.Total = len(ids)
	remaining := CaptureSearchMaxBytes
	metadata := 0
	type item struct {
		receipt CaptureSummary
		at      time.Time
	}
	matches := []item{}
	for _, id := range ids {
		if err = ctx.Err(); err != nil {
			return nil, err
		}
		// Lstat refuses links even when a directory entry was replaced after preflight.
		info, e := os.Lstat(filepath.Join(s.Dir, id+".json"))
		if e != nil {
			return nil, e
		}
		if !info.Mode().IsRegular() {
			return nil, fmt.Errorf("capture catalog unsupported file for %s", id)
		}
		receipt, n, digest, e := s.catalogEntry(id, remaining)
		if e != nil {
			return nil, e
		}
		remaining -= n
		catalogParts = append(catalogParts, id+":"+digest)
		if q.matches(receipt) {
			b, _ := json.Marshal(receipt)
			metadata += len(b)
			if metadata > CaptureSearchSummaryBytes {
				return nil, ErrCaptureSearchBudget
			}
			at, _ := time.Parse(time.RFC3339Nano, receipt.CapturedAt)
			matches = append(matches, item{receipt, at})
		}
	}
	catalog := searchHash(catalogParts)
	if cursor != "" && position.Catalog != catalog {
		return nil, ErrCaptureSearchCursor
	}

	if err = ctx.Err(); err != nil {
		return nil, err
	}
	slices.SortFunc(matches, func(a, b item) int {
		n := a.at.Compare(b.at)
		if q.Order == "newest" {
			n = -n
		}
		if n == 0 {
			return strings.Compare(a.receipt.ID, b.receipt.ID)
		}
		return n
	})
	result.Matched = len(matches)
	start := 0
	if cursor != "" {
		found := false
		for i, m := range matches {
			if m.receipt.ID == position.After {
				start = i + 1
				found = true
				break
			}
		}
		if !found {
			return nil, ErrCaptureSearchCursor
		}
	}
	end := min(start+CaptureListLimit, len(matches))
	for _, m := range matches[start:end] {
		result.Items = append(result.Items, m.receipt)
	}
	result.HasMore = end < len(matches)
	if result.HasMore {
		result.NextCursor = searchToken(captureSearchCursor{searchHash(q), catalog, matches[end-1].receipt.ID})
	}
	return result, nil
}

func (q CaptureSearchQuery) Validate() error { _, err := q.normalized(); return err }
