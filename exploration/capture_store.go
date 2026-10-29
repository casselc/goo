package exploration

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/casselc/goo/internal/immutablefile"
	"github.com/casselc/goo/measurements"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

const CaptureRowLimit = 1000
const CapturePageSize = 100
const CaptureMaxBytes = 24 << 20
const CaptureListLimit = 20
const CaptureCatalogLimit = 10000

var ErrCapturePlanChanged = errors.New("captured source plan differs from the current compiler; original evidence remains reopenable")

type ComparisonQuestion struct {
	Distribution bool  `json:"distribution,omitempty"`
	Baseline     Query `json:"baseline"`
	Selection    Query `json:"selection"`
}
type SavedComparison struct {
	BaselineID  string            `json:"baselineId"`
	SelectionID string            `json:"selectionId"`
	Request     ComparisonRequest `json:"request"`
	Result      *CohortComparison `json:"result"`
}
type CaptureQuestion struct {
	Explanation *ExplanationRequest    `json:"explanation,omitempty"`
	Measurement *measurements.Question `json:"measurement,omitempty"`
	Comparison  *ComparisonQuestion    `json:"comparison,omitempty"`
	Kind        string                 `json:"kind"`
	Explore     *Query                 `json:"explore,omitempty"`
	Inventory   *InventoryQuestion     `json:"inventory,omitempty"`
}

func (q CaptureQuestion) Validate() error {
	if q.Kind != "explanation" && q.Explanation != nil {
		return errors.New("unexpected explanation question")
	}
	if q.Kind != "measurement" && q.Measurement != nil {
		return errors.New("unexpected measurement question")
	}
	switch q.Kind {
	case "explanation":
		if q.Explanation == nil || q.Measurement != nil || q.Explore != nil || q.Inventory != nil || q.Comparison != nil {
			return errors.New("exclusive explanation question required")
		}
		return q.Explanation.Validate()
	case "measurement":
		if q.Measurement == nil || q.Explore != nil || q.Inventory != nil || q.Comparison != nil {
			return errors.New("exclusive measurement question required")
		}
		return q.Measurement.Validate()
	case "explore":
		if q.Explore == nil || q.Inventory != nil || q.Comparison != nil {
			return errors.New("explore question required")
		}
		return q.Explore.Validate()
	case "inventory":
		if q.Inventory == nil || q.Explore != nil || q.Comparison != nil {
			return errors.New("inventory question required")
		}
		return q.Inventory.Validate()
	case "comparison":
		if q.Comparison == nil || q.Explore != nil || q.Inventory != nil {
			return errors.New("comparison questions required")
		}
		a, b := q.Comparison.Baseline, q.Comparison.Selection
		if q.Comparison.Distribution && (!compatibleNumericBindings(a.Numeric, b.Numeric) || a.NumericHistogram == nil || b.NumericHistogram == nil) {
			return errors.New("distribution capture requires matching numeric bindings and histogram questions")
		}
		aggregateOnly := q.Comparison.Distribution && len(a.GroupBy) == 0 && len(b.GroupBy) == 0
		grouped := len(a.GroupBy) == 1 && len(b.GroupBy) == 1 && a.GroupBy[0] == b.GroupBy[0]
		if a.Validate() != nil || b.Validate() != nil || (!grouped && !aggregateOnly) || a.Signal != b.Signal || !a.SameEnvironment(b) {
			return errors.New("comparison requires valid questions with identical signal, environment and one breakdown; explicit distributions also support aggregate-only questions")
		}
		return nil
	default:
		return errors.New("kind must be explore, inventory, comparison or measurement")
	}
}

// Normalized removes inspection-page position, never semantic filters or bounds.
func (q CaptureQuestion) Normalized() CaptureQuestion {
	q.Explanation = detached(q.Explanation)
	if q.Inventory != nil {
		v := *q.Inventory
		v.Namespace = v.Namespace.Clone()
		q.Inventory = &v
	}
	if q.Measurement != nil {
		// Use the wire representation for immutable typed questions, including
		// equivalent empty/omitted selector maps. This also detaches caller maps.
		encoded, _ := json.Marshal(q.Measurement)
		var detached measurements.Question
		json.Unmarshal(encoded, &detached)
		q.Measurement = &detached
	}
	if q.Explore != nil {
		v := *q.Explore
		v.EnvironmentSelector = copyEnvironmentSelector(v.EnvironmentSelector)
		v.ReportedEdge = detached(v.ReportedEdge)
		v.Offset = 0
		v.Limit = CapturePageSize
		if len(v.GroupBy) == 0 {
			v.GroupBy = nil
		}
		q.Explore = &v
	}
	if q.Comparison != nil {
		pair := *q.Comparison
		pair.Baseline = *(CaptureQuestion{Explore: &pair.Baseline}).Normalized().Explore
		pair.Selection = *(CaptureQuestion{Explore: &pair.Selection}).Normalized().Explore
		q.Comparison = &pair
	}
	return q
}

type CaptureReceipt struct {
	Version    int             `json:"version"`
	ID         string          `json:"id"`
	Parent     string          `json:"parent,omitempty"`
	Title      string          `json:"title"`
	CapturedAt string          `json:"capturedAt"`
	Question   CaptureQuestion `json:"question"`
}
type SavedCapture struct {
	Explanation *ExplanationResult   `json:"explanation,omitempty"`
	Measurement *measurements.Result `json:"measurement,omitempty"`
	Comparison  *SavedComparison     `json:"comparison,omitempty"`
	CaptureReceipt
	Explore   *Result          `json:"explore,omitempty"`
	Inventory *InventoryResult `json:"inventory,omitempty"`
}
type CapturePage struct {
	Explanation     *ExplanationResult   `json:"explanation,omitempty"`
	Measurement     *measurements.Result `json:"measurement,omitempty"`
	Comparison      *SavedComparison     `json:"comparison,omitempty"`
	Receipt         CaptureReceipt       `json:"receipt"`
	Explore         *Result              `json:"explore,omitempty"`
	Inventory       *InventoryResult     `json:"inventory,omitempty"`
	RetainedRecords int                  `json:"retainedRecords"`
	RecordsLimited  bool                 `json:"recordsLimited"`
	NextCursor      string               `json:"nextCursor,omitempty"`
	PreviousCursor  string               `json:"previousCursor,omitempty"`
}
type CaptureList struct {
	Items      []CaptureReceipt `json:"items"`
	HasMore    bool             `json:"hasMore"`
	NextCursor string           `json:"nextCursor,omitempty"`
}
type CaptureStore struct{ Dir string }

func validCaptureTitle(title string) bool {
	return utf8.ValidString(title) && len(title) <= 120 && strings.TrimSpace(title) != "" && strings.IndexFunc(title, unicode.IsControl) < 0
}
func ValidateCaptureTitle(title string) error {
	if !validCaptureTitle(title) {
		return errors.New("title must be 1..120 UTF-8 bytes without control characters")
	}
	return nil
}
func captureDigest(c SavedCapture) (string, []byte, error) {
	c.ID = ""
	b, e := json.Marshal(c)
	if e != nil {
		return "", nil, e
	}
	h := sha256.Sum256(b)
	return hex.EncodeToString(h[:]), b, nil
}
func (s CaptureStore) Create(ctx context.Context, source Source, title string, q CaptureQuestion, parent string) (*SavedCapture, error) {
	if !validCaptureTitle(title) {
		return nil, errors.New("title must be 1..120 UTF-8 bytes without control characters")
	}
	if e := q.Validate(); e != nil {
		return nil, e
	}
	q = q.Normalized()
	if q.Kind == "explanation" {
		return nil, errors.New("explanation capture requires a held observed result; saving never reruns the source")
	}
	if parent != "" {
		p, e := s.Load(parent)
		if e != nil {
			return nil, e
		}
		if !reflect.DeepEqual(q, p.Question) {
			return nil, errors.New("refresh must retain the original question")
		}
		if q.Kind == "comparison" {
			for _, id := range []string{p.Comparison.BaselineID, p.Comparison.SelectionID} {
				child, e := s.Load(id)
				if e != nil {
					return nil, e
				}
				plan, e := compile(*child.Question.Explore, CaptureRowLimit)
				if e != nil {
					return nil, e
				}
				if !reflect.DeepEqual(plan, child.Explore.Plan) {
					return nil, ErrCapturePlanChanged
				}
			}
			return s.createComparison(ctx, source, title, q, parent, p.Comparison)
		}
		if q.Kind == "measurement" {
			plan, e := measurements.Compile(*q.Measurement)
			if e != nil {
				return nil, e
			}
			if !reflect.DeepEqual(plan, p.Measurement.Plan) {
				return nil, ErrCapturePlanChanged
			}
			return s.createMeasurement(ctx, source, title, q, parent, p.Measurement)
		}
		var current, previous Plan
		if q.Kind == "explore" {
			current, e = compile(*q.Explore, CaptureRowLimit)
			previous = p.Explore.Plan
		} else {
			current, e = CompileInventory(*q.Inventory)
			previous = p.Inventory.Plan
		}
		if e != nil {
			return nil, e
		}
		if !reflect.DeepEqual(current, previous) {
			return nil, ErrCapturePlanChanged
		}
	}
	if q.Kind == "comparison" {
		return s.createComparison(ctx, source, title, q, parent, nil)
	}
	if q.Kind == "measurement" {
		return s.createMeasurement(ctx, source, title, q, parent, nil)
	}
	c := SavedCapture{CaptureReceipt: CaptureReceipt{Version: 1, Parent: parent, Title: title, Question: q}}
	var e error
	if q.Kind == "explore" {
		c.Explore, e = source.captureRows(ctx, *q.Explore)
	} else {
		c.Inventory, e = source.Inventory(ctx, *q.Inventory)
	}
	if e != nil {
		return nil, e
	}
	c.CapturedAt = time.Now().UTC().Format(time.RFC3339Nano)
	return s.publish(c)
}
func validateCapture(c *SavedCapture) error {
	if c.Version != 1 || !validCaptureTitle(c.Title) || c.Question.Validate() != nil || !reflect.DeepEqual(c.Question, c.Question.Normalized()) {
		return errors.New("invalid capture receipt")
	}
	if _, e := time.Parse(time.RFC3339Nano, c.CapturedAt); e != nil {
		return errors.New("invalid capture time")
	}
	if c.Parent != "" && !regexpRevision.MatchString(c.Parent) {
		return errors.New("invalid capture parent")
	}
	if c.Question.Kind != "explanation" && c.Explanation != nil {
		return errors.New("unexpected explanation evidence")
	}
	if c.Question.Kind == "explanation" {
		if c.Explanation == nil || c.Measurement != nil || c.Explore != nil || c.Inventory != nil || c.Comparison != nil || c.Explanation.ResultToken != "" || !reflect.DeepEqual(c.Question.Explanation, &c.Explanation.Request) {
			return errors.New("invalid explanation capture envelope")
		}
		return ValidateExplanationResult(c.Explanation)
	}
	if c.Question.Kind == "measurement" {
		if c.Explore != nil || c.Inventory != nil || c.Comparison != nil || c.Measurement == nil || !reflect.DeepEqual(c.Measurement.Question, *c.Question.Measurement) {
			return errors.New("invalid measurement capture envelope")
		}
		return measurements.ValidateResult(c.Measurement)
	}
	if c.Measurement != nil {
		return errors.New("unexpected measurement evidence")
	}
	if c.Question.Kind == "comparison" {
		if c.Comparison == nil || c.Explore != nil || c.Inventory != nil || !regexpRevision.MatchString(c.Comparison.BaselineID) || !regexpRevision.MatchString(c.Comparison.SelectionID) {
			return errors.New("invalid saved comparison")
		}
		pair := c.Comparison
		if pair.Request.Distribution != c.Question.Comparison.Distribution {
			return errors.New("distribution comparison intent differs from question")
		}
		if !reflect.DeepEqual(pair.Request.Baseline.Question, c.Question.Comparison.Baseline) || !reflect.DeepEqual(pair.Request.Selection.Question, c.Question.Comparison.Selection) {
			return errors.New("comparison question drift")
		}
		result, e := CompareCohorts(pair.Request)
		if e != nil || !reflect.DeepEqual(result, pair.Result) {
			return errors.New("invalid comparison evidence")
		}
		return nil
	}
	if c.Comparison != nil {
		return errors.New("unexpected comparison evidence")
	}
	if c.Question.Kind == "explore" {
		if c.Explore != nil {
			for _, raw := range c.Explore.Groups {
				if _, e := groupValues(c.Explore.Question, raw); e != nil {
					return e
				}
			}
			if e := ValidateNumericResult(c.Explore); e != nil {
				return e
			}
			if c.Explore.State != "changed_between_reads" && !numericReadsAgree(c.Explore) {
				return errors.New("saved numeric measures disagree with aggregate reads")
			}
			if e := ValidateGroupedTimeSeries(*c.Question.Explore, c.Explore.GroupedTimeSeries); e != nil {
				return e
			}
			if e := ValidateTimeSeries(*c.Question.Explore, c.Explore.TimeSeries); e != nil {
				return e
			}
			if c.Question.Explore.GroupTimeSeries && c.Explore.State != "changed_between_reads" && !groupedSeriesAgree(c.Explore) {
				return errors.New("saved split chart disagrees with aggregate reads")
			}
			if c.Question.Explore.SpanStatus {
				if c.Explore.SpanStatus == nil || c.Explore.SpanStatus.Count != c.Explore.Count || c.Explore.SpanStatus.Validate() != nil {
					return errors.New("invalid saved span status")
				}
			} else if c.Explore.SpanStatus != nil {
				return errors.New("unexpected saved span status")
			}
			if c.Question.Explore.Summary == "span-duration" {
				if c.Explore.Duration == nil || c.Explore.Duration.Count != c.Explore.Count || c.Explore.Duration.Validate() != nil {
					return errors.New("invalid saved duration summary")
				}
			} else if c.Explore.Duration != nil {
				return errors.New("unexpected saved duration summary")
			}
		}
		if c.Explore == nil || c.Inventory != nil || !reflect.DeepEqual(c.Explore.Question, *c.Question.Explore) || len(c.Explore.Rows) > CaptureRowLimit || !regexpRevision.MatchString(c.Explore.Basis) {
			return errors.New("invalid saved exploration")
		}
		if err := validateSelectedRows(*c.Question.Explore, c.Explore.Rows); err != nil {
			return err
		}
	} else if c.Inventory == nil || c.Explore != nil || !reflect.DeepEqual(c.Inventory.Question, *c.Question.Inventory) || len(c.Inventory.Rows) > 1000 || !regexpRevision.MatchString(c.Inventory.Basis) {
		return errors.New("invalid saved inventory")
	}
	return nil
}
func (s CaptureStore) publish(c SavedCapture) (*SavedCapture, error) {
	if e := validateCapture(&c); e != nil {
		return nil, e
	}
	id, _, e := captureDigest(c)
	if e != nil {
		return nil, e
	}
	c.ID = id
	data, e := json.Marshal(c)
	if e != nil {
		return nil, e
	}
	if len(data) > CaptureMaxBytes {
		return nil, errors.New("capture exceeds 24 MiB; narrow question")
	}
	if _, e = immutablefile.Publish(s.Dir, id+".json", ".capture-", data); e != nil {
		return nil, e
	}
	verified, err := s.Load(id)
	if err != nil {
		return nil, err
	}
	// The immutable capture is already committed. Derived metadata failure must not
	// turn that success into an apparent failed publication; search reports absence.
	_ = s.indexCapture(verified)
	return verified, nil
}
func (s CaptureStore) Load(id string) (*SavedCapture, error) {
	c, _, err := s.loadBounded(id, CaptureMaxBytes)
	return c, err
}
func (s CaptureStore) loadBounded(id string, budget int64) (*SavedCapture, int64, error) {
	if !regexpRevision.MatchString(id) {
		return nil, 0, errors.New("invalid capture ID")
	}
	f, e := os.Open(filepath.Join(s.Dir, id+".json"))
	if e != nil {
		return nil, 0, e
	}
	defer f.Close()
	b, e := io.ReadAll(io.LimitReader(f, min(int64(CaptureMaxBytes), budget)+1))
	if e != nil {
		return nil, 0, e
	}
	if len(b) > CaptureMaxBytes {
		return nil, 0, errors.New("capture exceeds size limit")
	}
	if int64(len(b)) > budget {
		return nil, int64(len(b)), ErrCaptureSearchBudget
	}
	var c SavedCapture
	d := json.NewDecoder(strings.NewReader(string(b)))
	d.DisallowUnknownFields()
	if e = d.Decode(&c); e != nil {
		return nil, 0, e
	}
	if d.Decode(new(any)) != io.EOF {
		return nil, 0, errors.New("invalid capture document")
	}
	actual, _, e := captureDigest(c)
	if e != nil {
		return nil, 0, e
	}
	if c.ID != id || actual != id {
		return nil, 0, errors.New("capture integrity check failed")
	}
	if e = validateCapture(&c); e != nil {
		return nil, 0, e
	}
	return &c, int64(len(b)), nil
}

// Cursors bind a local position to an immutable receipt, not an authorization token.
func captureCursor(id string, offset int) string {
	return base64.RawURLEncoding.EncodeToString([]byte(id + ":" + strconv.Itoa(offset)))
}
func cursorOffset(id, cursor string) (int, error) {
	if cursor == "" {
		return 0, nil
	}
	if len(cursor) > 128 {
		return 0, errors.New("invalid capture cursor")
	}
	b, e := base64.RawURLEncoding.DecodeString(cursor)
	if e != nil {
		return 0, errors.New("invalid capture cursor")
	}
	parts := strings.Split(string(b), ":")
	if len(parts) != 2 || parts[0] != id {
		return 0, errors.New("cursor belongs to a different capture")
	}
	n, e := strconv.Atoi(parts[1])
	if e != nil || n < 0 || n >= CaptureRowLimit || n%CapturePageSize != 0 || strconv.Itoa(n) != parts[1] {
		return 0, errors.New("invalid capture position")
	}
	return n, nil
}
func (c *SavedCapture) Page(cursor string) (*CapturePage, error) {
	offset, e := cursorOffset(c.ID, cursor)
	if e != nil {
		return nil, e
	}
	p := &CapturePage{Explanation: c.Explanation, Measurement: c.Measurement, Receipt: c.CaptureReceipt, Inventory: c.Inventory, Comparison: c.Comparison}
	if c.Measurement != nil {
		p.RetainedRecords = len(c.Measurement.Rows)
	}
	if c.Explore == nil {
		if c.Explanation != nil && offset != 0 {
			return nil, errors.New("explanation has no record pages")
		}
		if offset != 0 {
			return nil, errors.New("inventory has no record pages")
		}
		return p, nil
	}
	n := len(c.Explore.Rows)
	if offset >= n && offset != 0 {
		return nil, errors.New("capture page outside retained records")
	}
	end := offset + CapturePageSize
	if end > n {
		end = n
	}
	result := *c.Explore
	result.Question.Offset = offset
	result.Question.Limit = CapturePageSize
	result.Rows = result.Rows[offset:end]
	if result.TraceExamples != nil {
		result.TraceExamples = SelectTraceExamples(result)
	}
	result.HasMore = end < n
	p.Explore = &result
	p.RetainedRecords = n
	p.RecordsLimited = c.Explore.HasMore
	if result.HasMore {
		p.NextCursor = captureCursor(c.ID, end)
	}
	if offset > 0 {
		p.PreviousCursor = captureCursor(c.ID, offset-CapturePageSize)
	}
	return p, nil
}

// Listing uses immutable ID ordering and an exclusive seek cursor. It is not a
// snapshot of a concurrently growing directory; direct IDs always remain reopenable.
func (s CaptureStore) List(after string) (*CaptureList, error) {
	if after != "" && !regexpRevision.MatchString(after) {
		return nil, errors.New("invalid list cursor")
	}
	list := &CaptureList{Items: []CaptureReceipt{}}
	directory, e := os.Open(s.Dir)
	if errors.Is(e, os.ErrNotExist) {
		return list, nil
	}
	if e != nil {
		return nil, e
	}
	defer directory.Close()
	entries, e := captureDirectoryEntries(directory)
	if errors.Is(e, ErrCaptureSearchBudget) {
		return nil, errors.New("capture catalog scan limit exceeded; direct capture IDs remain reopenable")
	}
	if errors.Is(e, os.ErrNotExist) {
		return list, nil
	}
	if e != nil && e != io.EOF {
		return nil, e
	}
	ids := []string{}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		id := strings.TrimSuffix(entry.Name(), ".json")
		if strings.HasSuffix(entry.Name(), ".json") && regexpRevision.MatchString(id) && id > after {
			ids = append(ids, id)
		}
	}
	sort.Strings(ids)
	if len(ids) > CaptureListLimit {
		list.HasMore = true
		ids = ids[:CaptureListLimit]
	}
	for _, id := range ids {
		c, e := s.Load(id)
		if e != nil {
			return nil, fmt.Errorf("capture catalog integrity failure for %s", id)
		}
		list.Items = append(list.Items, c.CaptureReceipt)
	}
	if list.HasMore {
		list.NextCursor = ids[len(ids)-1]
	}
	return list, nil
}

// A comparison is published only after both server-read child captures succeed.
// Children remain individually useful if the second read or pair publication fails.
func (s CaptureStore) createComparison(ctx context.Context, source Source, title string, q CaptureQuestion, parent string, previous *SavedComparison) (*SavedCapture, error) {
	numeric := q.Comparison.Baseline.Numeric != nil || q.Comparison.Selection.Numeric != nil
	if previous != nil {
		numeric = previous.Request.Numeric
	}
	if numeric && !compatibleNumericBindings(q.Comparison.Baseline.Numeric, q.Comparison.Selection.Numeric) {
		return nil, errors.New("numeric comparison requires identical declared bindings")
	}
	parents := [2]string{}
	if previous != nil {
		parents = [2]string{previous.BaselineID, previous.SelectionID}
	}
	captures := [2]*SavedCapture{}
	for i, question := range []Query{q.Comparison.Baseline, q.Comparison.Selection} {
		c, e := s.Create(ctx, source, title, CaptureQuestion{Kind: "explore", Explore: &question}, parents[i])
		if e != nil {
			return nil, e
		}
		captures[i] = c
	}
	snapshot := CohortFromResult
	if q.Comparison.Distribution {
		snapshot = CohortFromResultDistribution
	}
	a, e := snapshot(captures[0].Explore)
	if e != nil {
		return nil, e
	}
	b, e := snapshot(captures[1].Explore)
	if e != nil {
		return nil, e
	}
	if !numeric {
		a.Numeric = nil
		b.Numeric = nil
		for i := range a.Groups {
			a.Groups[i].Numeric = nil
		}
		for i := range b.Groups {
			b.Groups[i].Numeric = nil
		}
	}
	req := ComparisonRequest{Distribution: q.Comparison.Distribution, Numeric: numeric, Baseline: a, Selection: b}
	result, e := CompareCohorts(req)
	if e != nil {
		return nil, e
	}
	return s.publish(SavedCapture{CaptureReceipt: CaptureReceipt{Version: 1, Parent: parent, Title: title, Question: q, CapturedAt: time.Now().UTC().Format(time.RFC3339Nano)}, Comparison: &SavedComparison{BaselineID: captures[0].ID, SelectionID: captures[1].ID, Request: req, Result: result}})
}

func (s CaptureStore) createMeasurement(ctx context.Context, source Source, title string, q CaptureQuestion, parent string, previous *measurements.Result) (*SavedCapture, error) {
	adapter := measurements.Source{Endpoint: source.Endpoint, User: source.User, Password: source.Password, Client: source.Client, CustodyDir: filepath.Join(s.Dir, "custody")}
	result, e := adapter.Run(ctx, *q.Measurement, previous)
	if e != nil {
		return nil, e
	}
	c := SavedCapture{CaptureReceipt: CaptureReceipt{Version: 1, Parent: parent, Title: title, Question: q, CapturedAt: time.Now().UTC().Format(time.RFC3339Nano)}, Measurement: result}
	return s.publish(c)
}
