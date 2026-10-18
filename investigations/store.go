// Package investigations stores immutable, content-addressed investigation captures.
// Capture time is receipt time, not proof of immutable telemetry ingestion custody.
package investigations

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"time"
)

const MaxBytes = 24 << 20

var validID = regexp.MustCompile(`^[a-f0-9]{64}$`)

type Question struct {
	Package  string `json:"package"`
	Revision string `json:"packageRevision"`
	FromMS   int64  `json:"fromMs"`
	ToMS     int64  `json:"toMs"`
}
type Evidence struct {
	Operations json.RawMessage   `json:"operations"`
	Rates      json.RawMessage   `json:"rates,omitempty"`
	Histograms json.RawMessage   `json:"histograms,omitempty"`
	Errors     map[string]string `json:"errors,omitempty"`
}
type Capture struct {
	Version    int             `json:"version"`
	ID         string          `json:"id"`
	Parent     string          `json:"parent,omitempty"`
	CapturedAt string          `json:"capturedAt"`
	Question   Question        `json:"question"`
	Selection  json.RawMessage `json:"selection"`
	Evidence   Evidence        `json:"evidence"`
}
type Store struct{ Dir string }

func (q Question) Validate() error {
	if q.Package == "" || len(q.Package) > 128 || q.Revision == "" || q.FromMS < 60000 || q.ToMS <= q.FromMS || q.ToMS-q.FromMS > 120*60000 || q.ToMS > 4102444800000 {
		return errors.New("require package revision and fixed window of at most 120 minutes before 2100")
	}
	return nil
}
func digest(c Capture) (string, []byte, error) {
	c.ID = ""
	b, e := json.Marshal(c)
	if e != nil {
		return "", nil, e
	}
	// Normalize embedded JSON with Number, preserving exact decimal integer values.
	var v any
	d := json.NewDecoder(bytes.NewReader(b))
	d.UseNumber()
	if e = d.Decode(&v); e != nil {
		return "", nil, e
	}
	b, e = json.Marshal(v)
	if e != nil {
		return "", nil, e
	}
	h := sha256.Sum256(b)
	return hex.EncodeToString(h[:]), b, nil
}
func (s Store) Save(q Question, selection json.RawMessage, evidence Evidence, parent string) (*Capture, error) {
	if e := q.Validate(); e != nil {
		return nil, e
	}
	if len(selection) == 0 {
		selection = json.RawMessage(`{}`)
	}
	if e := ValidateSelection(selection); e != nil {
		return nil, e
	}
	if len(evidence.Operations) == 0 {
		return nil, errors.New("operation evidence required")
	}
	if parent != "" {
		p, e := s.Load(parent)
		if e != nil {
			return nil, e
		}
		if p.Question != q {
			return nil, errors.New("refresh must retain the original question and package revision")
		}
	}
	c := Capture{Version: 1, Parent: parent, CapturedAt: time.Now().UTC().Format(time.RFC3339Nano), Question: q, Selection: selection, Evidence: evidence}
	id, _, e := digest(c)
	if e != nil {
		return nil, e
	}
	c.ID = id
	data, e := json.Marshal(c)
	if e != nil {
		return nil, e
	}
	if len(data) > MaxBytes {
		return nil, errors.New("capture exceeds 24 MiB")
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
	// Link publishes a complete file atomically and never overwrites an existing ID.
	if e = os.Link(f.Name(), filepath.Join(s.Dir, id+".json")); e != nil {
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
func (s Store) Load(id string) (*Capture, error) {
	if !validID.MatchString(id) {
		return nil, errors.New("invalid capture ID")
	}
	f, e := os.Open(filepath.Join(s.Dir, id+".json"))
	if e != nil {
		return nil, e
	}
	defer f.Close()
	b, e := io.ReadAll(io.LimitReader(f, MaxBytes+1))
	if e != nil {
		return nil, e
	}
	if len(b) > MaxBytes {
		return nil, errors.New("capture exceeds size limit")
	}
	var c Capture
	if e = json.Unmarshal(b, &c); e != nil {
		return nil, e
	}
	actual, _, e := digest(c)
	if e != nil {
		return nil, e
	}
	if c.ID != id || actual != id || c.Version != 1 {
		return nil, fmt.Errorf("capture integrity check failed")
	}
	if e = c.Question.Validate(); e != nil {
		return nil, e
	}
	return &c, nil
}
