package exploration

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
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

type CaptureQuestion struct {
	Kind      string             `json:"kind"`
	Explore   *Query             `json:"explore,omitempty"`
	Inventory *InventoryQuestion `json:"inventory,omitempty"`
}

func (q CaptureQuestion) Validate() error {
	switch q.Kind {
	case "explore":
		if q.Explore == nil || q.Inventory != nil {
			return errors.New("explore question required")
		}
		return q.Explore.Validate()
	case "inventory":
		if q.Inventory == nil || q.Explore != nil {
			return errors.New("inventory question required")
		}
		return q.Inventory.Query().Validate()
	default:
		return errors.New("kind must be explore or inventory")
	}
}

// Normalized removes inspection-page position, never semantic filters or bounds.
func (q CaptureQuestion) Normalized() CaptureQuestion {
	if q.Explore != nil {
		v := *q.Explore
		v.Offset = 0
		v.Limit = CapturePageSize
		if len(v.GroupBy) == 0 {
			v.GroupBy = nil
		}
		q.Explore = &v
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
	CaptureReceipt
	Explore   *Result          `json:"explore,omitempty"`
	Inventory *InventoryResult `json:"inventory,omitempty"`
}
type CapturePage struct {
	Receipt         CaptureReceipt   `json:"receipt"`
	Explore         *Result          `json:"explore,omitempty"`
	Inventory       *InventoryResult `json:"inventory,omitempty"`
	RetainedRecords int              `json:"retainedRecords"`
	RecordsLimited  bool             `json:"recordsLimited"`
	NextCursor      string           `json:"nextCursor,omitempty"`
	PreviousCursor  string           `json:"previousCursor,omitempty"`
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
	if parent != "" {
		p, e := s.Load(parent)
		if e != nil {
			return nil, e
		}
		if !reflect.DeepEqual(q, p.Question) {
			return nil, errors.New("refresh must retain the original question")
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
	if c.Question.Kind == "explore" {
		if c.Explore != nil {
			if e := ValidateTimeSeries(*c.Question.Explore, c.Explore.TimeSeries); e != nil {
				return e
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
	} else if c.Inventory == nil || c.Explore != nil || c.Inventory.Question != *c.Question.Inventory || len(c.Inventory.Rows) > 1000 || !regexpRevision.MatchString(c.Inventory.Basis) {
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
	if e = os.MkdirAll(s.Dir, 0700); e != nil {
		return nil, e
	}
	f, e := os.CreateTemp(s.Dir, ".capture-")
	if e != nil {
		return nil, e
	}
	defer os.Remove(f.Name())
	if _, e = f.Write(data); e != nil {
		f.Close()
		return nil, e
	}
	if e = f.Sync(); e != nil {
		f.Close()
		return nil, e
	}
	if e = f.Close(); e != nil {
		return nil, e
	}
	if e = os.Link(f.Name(), filepath.Join(s.Dir, id+".json")); e != nil && !errors.Is(e, os.ErrExist) {
		return nil, e
	}
	dir, e := os.Open(s.Dir)
	if e != nil {
		return nil, e
	}
	e = dir.Sync()
	dir.Close()
	if e != nil {
		return nil, e
	}
	return s.Load(id)
}
func (s CaptureStore) Load(id string) (*SavedCapture, error) {
	if !regexpRevision.MatchString(id) {
		return nil, errors.New("invalid capture ID")
	}
	f, e := os.Open(filepath.Join(s.Dir, id+".json"))
	if e != nil {
		return nil, e
	}
	defer f.Close()
	b, e := io.ReadAll(io.LimitReader(f, CaptureMaxBytes+1))
	if e != nil {
		return nil, e
	}
	if len(b) > CaptureMaxBytes {
		return nil, errors.New("capture exceeds size limit")
	}
	var c SavedCapture
	d := json.NewDecoder(strings.NewReader(string(b)))
	d.DisallowUnknownFields()
	if e = d.Decode(&c); e != nil {
		return nil, e
	}
	if d.Decode(new(any)) != io.EOF {
		return nil, errors.New("invalid capture document")
	}
	actual, _, e := captureDigest(c)
	if e != nil {
		return nil, e
	}
	if c.ID != id || actual != id {
		return nil, errors.New("capture integrity check failed")
	}
	if e = validateCapture(&c); e != nil {
		return nil, e
	}
	return &c, nil
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
	p := &CapturePage{Receipt: c.CaptureReceipt, Inventory: c.Inventory}
	if c.Explore == nil {
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
	entries, e := directory.ReadDir(CaptureCatalogLimit + 1)
	if errors.Is(e, os.ErrNotExist) {
		return list, nil
	}
	if e != nil && e != io.EOF {
		return nil, e
	}
	if len(entries) > CaptureCatalogLimit {
		return nil, errors.New("capture catalog scan limit exceeded; direct capture IDs remain reopenable")
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
