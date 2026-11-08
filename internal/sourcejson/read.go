// Package sourcejson admits bounded JSONEachRow source objects. It owns framing
// only: callers retain SQL, HTTP status, source failure and evidence semantics.
package sourcejson

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"math"

	"github.com/casselc/goo/internal/sourcehttp"
)

type BlankLines uint8

const (
	RejectBlankLines BlankLines = iota
	SkipBlankLines
)

// Policy is explicit at each source owner. MaxRows counts non-skipped physical
// lines. TrimRows changes retained bytes; otherwise only a final LF is removed.
// Both framing modes require one complete JSON object per admitted line.
type Policy struct {
	MaxBytes   int64
	MaxRows    int
	BlankLines BlankLines
	TrimRows   bool
}

var (
	ErrByteLimit = errors.New("JSONEachRow byte limit exceeded")
	ErrRowLimit  = errors.New("JSONEachRow row limit exceeded")
	ErrObject    = errors.New("invalid JSONEachRow object")
)

// ReadObjects returns independently owned rows, or no partial rows on any
// failure. Read errors are returned unchanged. A byte overrun precedes parsing;
// after skipped blanks, a row overrun precedes validation of the overflow line.
// Duplicate keys and string encoding use ordinary encoding/json syntax rules.
func ReadObjects(reader io.Reader, policy Policy) ([]json.RawMessage, error) {
	if policy.MaxBytes < 0 || policy.MaxBytes == math.MaxInt64 || policy.MaxRows < 0 || policy.BlankLines > SkipBlankLines {
		return nil, errors.New("invalid JSONEachRow policy")
	}
	data, oversize, err := sourcehttp.ReadBounded(reader, policy.MaxBytes)
	if err != nil {
		return nil, err
	}
	if oversize {
		return nil, ErrByteLimit
	}
	rows := []json.RawMessage{}
	for line := range bytes.Lines(data) {
		trimmed := bytes.TrimSpace(line)
		if policy.BlankLines == SkipBlankLines && len(trimmed) == 0 {
			continue
		}
		if len(rows) >= policy.MaxRows {
			return nil, ErrRowLimit
		}
		if policy.TrimRows {
			line = trimmed
		}
		// Valid checks the entire line, including trailing bytes; the first byte
		// establishes object shape without allocating a map or decoding values.
		if len(trimmed) == 0 || trimmed[0] != '{' || !json.Valid(line) {
			return nil, ErrObject
		}
		rows = append(rows, json.RawMessage(bytes.Clone(bytes.TrimSuffix(line, []byte("\n")))))
	}
	return rows, nil
}
