package exploration

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/casselc/goo/identity"
	"github.com/casselc/goo/internal/directorylock"
)

const KeptCatalogLimit = 1000
const KeptCatalogEntryBytes = 4096
const KeptRepairBytes = 32 << 20
const KeptRepairBatch = 64
const keptCatalogVersion = "goo/kept-search/v1"

var ErrKeepCapacity = errors.New("keep storage capacity reached; existing exact IDs remain accessible")
var ErrKeptCatalog = errors.New("kept observation index unavailable; repair the index or open an exact ID")
var ErrKeptCursor = errors.New("kept observation cursor no longer matches; restart discovery or repair")
var keptCatalogSlots = make(chan struct{}, 2)

type KeptSummary struct {
	ID                  string                        `json:"id"`
	Version             string                        `json:"version"`
	Origin              string                        `json:"origin"`
	Signal              string                        `json:"signal"`
	Environment         *string                       `json:"environment,omitempty"`
	EnvironmentSelector *identity.EnvironmentSelector `json:"environmentSelector,omitempty"`
	Observations        int                           `json:"observations"`
	KeptAt              time.Time                     `json:"keptAt"`
	Digest              string                        `json:"digest"`
}

func keptSummary(v keptResource) (KeptSummary, error) {
	var r KeptSummary
	switch v := v.(type) {
	case *KeptObservations:
		r = KeptSummary{ID: v.ID, Version: v.Version, Origin: "tail-selection", Signal: v.Selector.Signal, EnvironmentSelector: detached(v.Selector.EnvironmentSelector), Observations: len(v.Members), KeptAt: v.KeptAt, Digest: v.Digest}
		if v.Selector.EnvironmentSelector == nil {
			value := v.Selector.Environment
			r.Environment = &value
		}
	case *KeptRecords:
		r = KeptSummary{ID: v.ID, Version: v.Version, Origin: v.Origin, Signal: v.Question.Signal, EnvironmentSelector: detached(v.Question.EnvironmentSelector), Observations: len(v.Members), KeptAt: v.KeptAt, Digest: v.Digest}
		if v.Question.EnvironmentSelector == nil {
			value := v.Question.Environment
			r.Environment = &value
		}
	default:
		return r, ErrKeptCatalog
	}
	if r.Signal == "" {
		r.Signal = "logs"
	}
	return r, r.Validate()
}
func (r KeptSummary) Validate() error {
	if !validTailConnection(r.ID) || !validTailConnection(r.Digest) || r.KeptAt.IsZero() || r.Observations < 1 || r.Observations > 20 || !slices.Contains([]string{"logs", "spans"}, r.Signal) {
		return ErrKeptCatalog
	}
	if r.Version == KeptObservationsVersion {
		if r.Origin != "tail-selection" {
			return ErrKeptCatalog
		}
	} else if r.Version != KeptRecordsVersion || !slices.Contains([]string{"result-page", "saved-capture"}, r.Origin) {
		return ErrKeptCatalog
	}
	selector := TailSelector{Signal: r.Signal, EnvironmentSelector: r.EnvironmentSelector}
	if r.Environment != nil {
		selector.Environment = *r.Environment
	}
	if (r.Environment == nil) == (r.EnvironmentSelector == nil) {
		return ErrKeptCatalog
	}
	if _, _, e := tailScope(selector, 60000); e != nil {
		return ErrKeptCatalog
	}
	return nil
}

type keptIndexEntry struct {
	Version    int         `json:"version"`
	Summary    KeptSummary `json:"summary"`
	Size       int64       `json:"size"`
	ModifiedNS int64       `json:"modifiedNS"`
	Checksum   string      `json:"checksum"`
}
type keptFile struct {
	ID               string
	Size, ModifiedNS int64
}

func (s KeptObservationStore) catalogDir() string { return filepath.Join(s.Dir, ".catalog-v1") }

// Every physical root entry consumes admission except the real derived directory.
// Call under a directory lock so publisher scratch files cannot transiently overflow reads.
func (s KeptObservationStore) keepFiles() (files []keptFile, physical int, err error) {
	dir, e := os.Open(s.Dir)
	if e != nil {
		return nil, 0, e
	}
	defer dir.Close()
	entries, e := dir.ReadDir(KeptCatalogLimit + 2)
	if e != nil && e != io.EOF {
		return nil, 0, e
	}
	for _, entry := range entries {
		if entry.Name() == ".catalog-v1" && entry.IsDir() {
			continue
		}
		physical++
		if physical > KeptCatalogLimit {
			return nil, physical, ErrKeepCapacity
		}
		id := strings.TrimSuffix(entry.Name(), ".json")
		if !strings.HasSuffix(entry.Name(), ".json") || !validTailConnection(id) {
			continue
		}
		info, e := entry.Info()
		if e != nil {
			return nil, physical, e
		}
		if !info.Mode().IsRegular() || info.Size() > KeptObservationsMaxBytes {
			return nil, physical, ErrKeptCatalog
		}
		files = append(files, keptFile{id, info.Size(), info.ModTime().UnixNano()})
	}
	slices.SortFunc(files, func(a, b keptFile) int { return strings.Compare(a.ID, b.ID) })
	return files, physical, nil
}
func (s KeptObservationStore) indexKeep(v keptResource) error {
	summary, e := keptSummary(v)
	if e != nil {
		return e
	}
	info, e := os.Lstat(filepath.Join(s.Dir, summary.ID+".json"))
	if e != nil || !info.Mode().IsRegular() {
		return ErrKeptCatalog
	}
	entry := keptIndexEntry{Version: 1, Summary: summary, Size: info.Size(), ModifiedNS: info.ModTime().UnixNano()}
	entry.Checksum = tailMembershipHash(entry)
	raw, e := json.Marshal(entry)
	if e != nil {
		return e
	}
	if len(raw) > KeptCatalogEntryBytes {
		return ErrKeptCatalog
	}
	if e = os.Mkdir(s.catalogDir(), 0700); e != nil && !errors.Is(e, os.ErrExist) {
		return e
	}
	catalogInfo, e := os.Lstat(s.catalogDir())
	if e != nil || !catalogInfo.IsDir() {
		return ErrKeptCatalog
	}
	f, e := os.CreateTemp(s.catalogDir(), ".metadata-")
	if e != nil {
		return e
	}
	defer os.Remove(f.Name())
	defer f.Close()
	if _, e = f.Write(raw); e != nil {
		return e
	}
	if e = f.Sync(); e != nil {
		return e
	}
	if e = f.Close(); e != nil {
		return e
	}
	if e = os.Rename(f.Name(), filepath.Join(s.catalogDir(), summary.ID+".json")); e != nil {
		return e
	}
	dir, e := os.Open(s.catalogDir())
	if e != nil {
		return e
	}
	defer dir.Close()
	return dir.Sync()
}
func (s KeptObservationStore) readKeepIndex(file keptFile) (KeptSummary, string, error) {
	var entry keptIndexEntry
	catalogInfo, e := os.Lstat(s.catalogDir())
	if e != nil || !catalogInfo.IsDir() {
		return entry.Summary, "", ErrKeptCatalog
	}
	path := filepath.Join(s.catalogDir(), file.ID+".json")
	info, e := os.Lstat(path)
	if e != nil || !info.Mode().IsRegular() || info.Size() > KeptCatalogEntryBytes {
		return entry.Summary, "", ErrKeptCatalog
	}
	f, e := os.Open(path)
	if e != nil {
		return entry.Summary, "", ErrKeptCatalog
	}
	defer f.Close()
	raw, e := io.ReadAll(io.LimitReader(f, KeptCatalogEntryBytes+1))
	if e != nil || len(raw) > KeptCatalogEntryBytes {
		return entry.Summary, "", ErrKeptCatalog
	}
	d := json.NewDecoder(bytes.NewReader(raw))
	d.DisallowUnknownFields()
	if d.Decode(&entry) != nil || d.Decode(new(any)) != io.EOF {
		return entry.Summary, "", ErrKeptCatalog
	}
	checksum := entry.Checksum
	entry.Checksum = ""
	if entry.Version != 1 || entry.Summary.ID != file.ID || entry.Summary.Validate() != nil || checksum != tailMembershipHash(entry) || entry.Size != file.Size || entry.ModifiedNS != file.ModifiedNS {
		return entry.Summary, "", ErrKeptCatalog
	}
	return entry.Summary, checksum, nil
}

type KeptSearchQuery struct {
	Origin string `json:"origin,omitempty"`
	Signal string `json:"signal,omitempty"`
}

func (q KeptSearchQuery) Validate() error {
	if q.Origin != "" && !slices.Contains([]string{"tail-selection", "result-page", "saved-capture"}, q.Origin) || q.Signal != "" && !slices.Contains([]string{"logs", "spans"}, q.Signal) {
		return errors.New("unsupported kept origin or signal")
	}
	return nil
}

type KeptSearchPage struct {
	Version    string          `json:"version"`
	Query      KeptSearchQuery `json:"query"`
	Items      []KeptSummary   `json:"items"`
	Total      int             `json:"total"`
	Matched    int             `json:"matched"`
	HasMore    bool            `json:"hasMore"`
	NextCursor string          `json:"nextCursor,omitempty"`
}
type keptCatalogCursor struct {
	Basis string `json:"basis"`
	Query string `json:"query"`
	After string `json:"after"`
}

func keptCursorToken(v keptCatalogCursor) string {
	b, _ := json.Marshal(v)
	return base64.RawURLEncoding.EncodeToString(b)
}
func readKeptCursor(token, query string) (keptCatalogCursor, error) {
	var c keptCatalogCursor
	if token == "" {
		return c, nil
	}
	if len(token) > 1024 {
		return c, ErrKeptCursor
	}
	b, e := base64.RawURLEncoding.DecodeString(token)
	if e != nil || json.Unmarshal(b, &c) != nil || keptCursorToken(c) != token || !validTailConnection(c.Basis) || !validTailConnection(c.After) || c.Query != query {
		return c, ErrKeptCursor
	}
	return c, nil
}
func catalogAdmission(ctx context.Context) (func(), error) {
	select {
	case keptCatalogSlots <- struct{}{}:
		return func() { <-keptCatalogSlots }, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}
func (s KeptObservationStore) Search(ctx context.Context, q KeptSearchQuery, cursor string) (*KeptSearchPage, error) {
	if e := q.Validate(); e != nil {
		return nil, e
	}
	position, e := readKeptCursor(cursor, tailMembershipHash(q))
	if e != nil {
		return nil, e
	}
	release, e := catalogAdmission(ctx)
	if e != nil {
		return nil, e
	}
	defer release()
	page := &KeptSearchPage{Version: keptCatalogVersion, Query: q, Items: []KeptSummary{}}
	unlock, e := directorylock.Acquire(ctx, s.Dir, false)
	if errors.Is(e, os.ErrNotExist) {
		if cursor != "" {
			return nil, ErrKeptCursor
		}
		return page, nil
	}
	if e != nil {
		return nil, e
	}
	defer unlock()
	files, _, e := s.keepFiles()
	if e != nil {
		return nil, e
	}
	sums := make([]KeptSummary, 0, len(files))
	fingerprints := make([]string, 0, len(files))
	for _, file := range files {
		if e = ctx.Err(); e != nil {
			return nil, e
		}
		summary, hash, e := s.readKeepIndex(file)
		if e != nil {
			return nil, e
		}
		fingerprints = append(fingerprints, hash)
		if (q.Origin == "" || q.Origin == summary.Origin) && (q.Signal == "" || q.Signal == summary.Signal) {
			sums = append(sums, summary)
		}
	}
	basis := tailMembershipHash(fingerprints)
	if cursor != "" && position.Basis != basis {
		return nil, ErrKeptCursor
	}
	slices.SortFunc(sums, func(a, b KeptSummary) int {
		if n := b.KeptAt.Compare(a.KeptAt); n != 0 {
			return n
		}
		return strings.Compare(a.ID, b.ID)
	})
	page.Total = len(files)
	page.Matched = len(sums)
	start := 0
	if cursor != "" {
		start = slices.IndexFunc(sums, func(v KeptSummary) bool { return v.ID == position.After }) + 1
		if start == 0 {
			return nil, ErrKeptCursor
		}
	}
	end := min(start+20, len(sums))
	page.Items = append(page.Items, sums[start:end]...)
	page.HasMore = end < len(sums)
	if page.HasMore {
		page.NextCursor = keptCursorToken(keptCatalogCursor{basis, tailMembershipHash(q), sums[end-1].ID})
	}
	return page, nil
}

type KeptRepairRequest struct {
	Cursor string `json:"cursor,omitempty"`
}
type KeptRepairPage struct {
	Indexed    int    `json:"indexed"`
	Bytes      int64  `json:"bytes"`
	Completed  int    `json:"completed"`
	Total      int    `json:"total"`
	HasMore    bool   `json:"hasMore"`
	NextCursor string `json:"nextCursor,omitempty"`
}

func (s KeptObservationStore) RepairCatalog(ctx context.Context, cursor string) (KeptRepairPage, error) {
	var result KeptRepairPage
	position, e := readKeptCursor(cursor, "repair-v1")
	if e != nil {
		return result, e
	}
	release, e := catalogAdmission(ctx)
	if e != nil {
		return result, e
	}
	defer release()
	unlock, e := directorylock.Acquire(ctx, s.Dir, true)
	if errors.Is(e, os.ErrNotExist) {
		if cursor != "" {
			return result, ErrKeptCursor
		}
		return result, nil
	}
	if e != nil {
		return result, e
	}
	defer unlock()
	files, _, e := s.keepFiles()
	if e != nil {
		return result, e
	}
	basis := tailMembershipHash(files)
	if cursor != "" && position.Basis != basis {
		return result, ErrKeptCursor
	}
	start := 0
	if cursor != "" {
		start = slices.IndexFunc(files, func(f keptFile) bool { return f.ID == position.After }) + 1
		if start == 0 {
			return result, ErrKeptCursor
		}
	}
	result.Total = len(files)
	result.Completed = start
	for i := start; i < len(files); i++ {
		file := files[i]
		if result.Indexed >= KeptRepairBatch || result.Bytes+file.Size > KeptRepairBytes {
			result.HasMore = true
			break
		}
		if e = ctx.Err(); e != nil {
			return result, e
		}
		v, n, e := s.loadBounded(file.ID, KeptRepairBytes-result.Bytes)
		if e != nil {
			return result, fmt.Errorf("keep %s: %w", file.ID, e)
		}
		if e = s.indexKeep(v); e != nil {
			return result, e
		}
		result.Bytes += n
		result.Indexed++
		result.Completed++
	}
	if result.HasMore {
		result.NextCursor = keptCursorToken(keptCatalogCursor{basis, "repair-v1", files[result.Completed-1].ID})
	}
	return result, nil
}
