package exploration

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"
	"unicode/utf8"
)

const CaptureCatalogEntryBytes int64 = 64 << 10
const CaptureCatalogRebuildBytes int64 = 256 << 20

var ErrCaptureCatalogUnavailable = errors.New("capture metadata catalog unavailable; run goo rebuild-capture-catalog with the capture directory, or open by ID")

// Reserve the two known derived directories separately so a v1 -> v2 rebuild
// cannot make an admitted store exceed its own enumeration budget. All other
// physical entries count, including unrelated directories and temporary files.
func captureDirectoryEntries(dir *os.File) ([]os.DirEntry, error) {
	entries, err := dir.ReadDir(CaptureCatalogLimit + 3)
	if err != nil && err != io.EOF {
		return nil, err
	}
	counted := 0
	for _, entry := range entries {
		if entry.IsDir() && (entry.Name() == ".catalog-v1" || entry.Name() == ".catalog-v2") {
			continue
		}
		counted++
	}
	if counted > CaptureCatalogLimit {
		return nil, ErrCaptureSearchBudget
	}
	return entries, nil
}

type captureCatalogEntry struct {
	Version    int            `json:"version"`
	Summary    CaptureSummary `json:"summary"`
	Size       int64          `json:"size"`
	ModifiedNS int64          `json:"modifiedNS"`
	Digest     string         `json:"digest"`
}

func (s CaptureStore) catalogDir() string { return filepath.Join(s.Dir, ".catalog-v2") }
func (s CaptureStore) indexCapture(c *SavedCapture) error {
	info, e := os.Stat(filepath.Join(s.Dir, c.ID+".json"))
	if e != nil {
		return e
	}
	entry := captureCatalogEntry{Version: 2, Summary: summarizeCapture(c.CaptureReceipt), Size: info.Size(), ModifiedNS: info.ModTime().UnixNano()}
	entry.Digest = searchHash(entry)
	b, e := json.Marshal(entry)
	if e != nil {
		return e
	}
	if int64(len(b)) > CaptureCatalogEntryBytes {
		return ErrCaptureSearchBudget
	}
	dir := s.catalogDir()
	if e = os.MkdirAll(dir, 0700); e != nil {
		return e
	}
	f, e := os.CreateTemp(dir, ".metadata-")
	if e != nil {
		return e
	}
	defer os.Remove(f.Name())
	if _, e = f.Write(b); e != nil {
		f.Close()
		return e
	}
	if e = f.Sync(); e != nil {
		f.Close()
		return e
	}
	if e = f.Close(); e != nil {
		return e
	}
	if e = os.Rename(f.Name(), filepath.Join(dir, c.ID+".json")); e != nil {
		return e
	}
	d, e := os.Open(dir)
	if e != nil {
		return e
	}
	defer d.Close()
	return d.Sync()
}
func (s CaptureStore) catalogEntry(id string, budget int64) (CaptureSummary, int64, string, error) {
	var entry captureCatalogEntry
	indexInfo, indexErr := os.Lstat(filepath.Join(s.catalogDir(), id+".json"))
	if indexErr != nil {
		return entry.Summary, 0, "", ErrCaptureCatalogUnavailable
	}
	if !indexInfo.Mode().IsRegular() {
		return entry.Summary, 0, "", fmt.Errorf("capture metadata unsupported file for %s", id)
	}
	f, e := os.Open(filepath.Join(s.catalogDir(), id+".json"))
	if e != nil {
		return entry.Summary, 0, "", ErrCaptureCatalogUnavailable
	}
	defer f.Close()
	b, e := io.ReadAll(io.LimitReader(f, min(CaptureCatalogEntryBytes, budget)+1))
	if e != nil {
		return entry.Summary, 0, "", e
	}
	if int64(len(b)) > min(CaptureCatalogEntryBytes, budget) {
		return entry.Summary, 0, "", ErrCaptureSearchBudget
	}
	d := json.NewDecoder(bytes.NewReader(b))
	d.DisallowUnknownFields()
	if e = d.Decode(&entry); e != nil || d.Decode(new(any)) != io.EOF {
		return entry.Summary, 0, "", fmt.Errorf("capture metadata integrity failure for %s", id)
	}
	digest := entry.Digest
	entry.Digest = ""
	_, atError := time.Parse(time.RFC3339Nano, entry.Summary.CapturedAt)
	r := entry.Summary
	validEnvironment := r.Environment == nil || len(*r.Environment) <= 128 && utf8.ValidString(*r.Environment) && !strings.ContainsRune(*r.Environment, 0)
	validKind := slices.Contains([]string{"explore", "inventory", "comparison", "explanation", "measurement"}, r.Kind)
	validSignal := slices.Contains([]string{"", "logs", "spans", "metrics"}, r.Signal)
	validMeasurement := r.Kind != "measurement" && r.MeasurementKind == "" || r.Kind == "measurement" && slices.Contains([]string{"gauge", "counter", "histogram", "lifecycle", "stock", "state"}, r.MeasurementKind)
	if entry.Version != 2 || (r.Parent != "" && !regexpRevision.MatchString(r.Parent)) || r.ID != id || digest != searchHash(entry) || !validCaptureTitle(r.Title) || atError != nil || r.CapturedAt != canonicalCaptureTime(r.CapturedAt) || !validEnvironment || !validKind || !validSignal || !validMeasurement {
		return r, 0, "", fmt.Errorf("capture metadata integrity failure for %s", id)
	}

	info, e := os.Lstat(filepath.Join(s.Dir, id+".json"))
	if e != nil {
		return entry.Summary, 0, "", e
	}
	if !info.Mode().IsRegular() || info.Size() != entry.Size || info.ModTime().UnixNano() != entry.ModifiedNS {
		return entry.Summary, 0, "", fmt.Errorf("capture file changed since metadata verification for %s; rebuild catalog", id)
	}
	return entry.Summary, int64(len(b)), digest, nil
}

type CaptureCatalogRebuild struct {
	Indexed int   `json:"indexed"`
	Bytes   int64 `json:"bytes"`
}

// RebuildCatalog is explicit bounded administrative I/O; it never alters capture payloads.
func (s CaptureStore) RebuildCatalog(ctx context.Context) (CaptureCatalogRebuild, error) {
	var result CaptureCatalogRebuild
	dir, e := os.Open(s.Dir)
	if errors.Is(e, os.ErrNotExist) {
		return result, nil
	}
	if e != nil {
		return result, e
	}
	defer dir.Close()
	entries, e := captureDirectoryEntries(dir)
	if e != nil {
		return result, e
	}
	var ids []string
	var total int64
	for _, entry := range entries {
		if ctx.Err() != nil {
			return result, ctx.Err()
		}
		id := strings.TrimSuffix(entry.Name(), ".json")
		if !strings.HasSuffix(entry.Name(), ".json") || !regexpRevision.MatchString(id) {
			continue
		}
		info, e := entry.Info()
		if e != nil {
			return result, e
		}
		if !info.Mode().IsRegular() {
			return result, fmt.Errorf("unsupported capture file %s", id)
		}
		total += info.Size()
		if total > CaptureCatalogRebuildBytes {
			return result, errors.New("capture rebuild exceeds 256 MiB; direct IDs remain reopenable")
		}
		ids = append(ids, id)
	}
	for _, id := range ids {
		if ctx.Err() != nil {
			return result, ctx.Err()
		}
		c, n, e := s.loadBounded(id, CaptureCatalogRebuildBytes-result.Bytes)
		if e != nil {
			return result, e
		}
		if e = s.indexCapture(c); e != nil {
			return result, e
		}
		result.Bytes += n
		result.Indexed++
	}
	return result, nil
}
