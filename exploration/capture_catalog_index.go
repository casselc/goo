package exploration

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"time"
)

const CaptureCatalogEntryBytes int64 = 64 << 10
const CaptureCatalogRebuildBytes int64 = 256 << 20

var ErrCaptureCatalogUnavailable = errors.New("capture metadata catalog unavailable; run goo rebuild-capture-catalog with the capture directory, or open by ID")

type captureCatalogEntry struct {
	Version    int            `json:"version"`
	Receipt    CaptureReceipt `json:"receipt"`
	Size       int64          `json:"size"`
	ModifiedNS int64          `json:"modifiedNS"`
	Digest     string         `json:"digest"`
}

func (s CaptureStore) catalogDir() string { return filepath.Join(s.Dir, ".catalog-v1") }
func (s CaptureStore) indexCapture(c *SavedCapture) error {
	info, e := os.Stat(filepath.Join(s.Dir, c.ID+".json"))
	if e != nil {
		return e
	}
	entry := captureCatalogEntry{Version: 1, Receipt: c.CaptureReceipt, Size: info.Size(), ModifiedNS: info.ModTime().UnixNano()}
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
func (s CaptureStore) catalogEntry(id string, budget int64) (CaptureReceipt, int64, string, error) {
	var entry captureCatalogEntry
	indexInfo, indexErr := os.Lstat(filepath.Join(s.catalogDir(), id+".json"))
	if indexErr != nil {
		return entry.Receipt, 0, "", ErrCaptureCatalogUnavailable
	}
	if !indexInfo.Mode().IsRegular() {
		return entry.Receipt, 0, "", fmt.Errorf("capture metadata unsupported file for %s", id)
	}
	f, e := os.Open(filepath.Join(s.catalogDir(), id+".json"))
	if e != nil {
		return entry.Receipt, 0, "", ErrCaptureCatalogUnavailable
	}
	defer f.Close()
	b, e := io.ReadAll(io.LimitReader(f, min(CaptureCatalogEntryBytes, budget)+1))
	if e != nil {
		return entry.Receipt, 0, "", e
	}
	if int64(len(b)) > min(CaptureCatalogEntryBytes, budget) {
		return entry.Receipt, 0, "", ErrCaptureSearchBudget
	}
	d := json.NewDecoder(strings.NewReader(string(b)))
	d.DisallowUnknownFields()
	if e = d.Decode(&entry); e != nil || d.Decode(new(any)) != io.EOF {
		return entry.Receipt, 0, "", fmt.Errorf("capture metadata integrity failure for %s", id)
	}
	digest := entry.Digest
	entry.Digest = ""
	_, atError := time.Parse(time.RFC3339Nano, entry.Receipt.CapturedAt)
	if entry.Version != 1 || entry.Receipt.Version != 1 || (entry.Receipt.Parent != "" && !regexpRevision.MatchString(entry.Receipt.Parent)) || entry.Receipt.ID != id || digest != searchHash(entry) || !validCaptureTitle(entry.Receipt.Title) || atError != nil || entry.Receipt.Question.Validate() != nil || !reflect.DeepEqual(entry.Receipt.Question, entry.Receipt.Question.Normalized()) {
		return entry.Receipt, 0, "", fmt.Errorf("capture metadata integrity failure for %s", id)
	}
	info, e := os.Lstat(filepath.Join(s.Dir, id+".json"))
	if e != nil {
		return entry.Receipt, 0, "", e
	}
	if !info.Mode().IsRegular() || info.Size() != entry.Size || info.ModTime().UnixNano() != entry.ModifiedNS {
		return entry.Receipt, 0, "", fmt.Errorf("capture file changed since metadata verification for %s; rebuild catalog", id)
	}
	return entry.Receipt, int64(len(b)), digest, nil
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
	entries, e := dir.ReadDir(CaptureCatalogLimit + 1)
	if e != nil && e != io.EOF {
		return result, e
	}
	if len(entries) > CaptureCatalogLimit {
		return result, ErrCaptureSearchBudget
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
