package verification

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"reflect"
	"strings"
	"testing"

	"github.com/casselc/goo/internal/sourcejson"
	"hegel.dev/go/hegel"
)

// referenceRows381 intentionally uses string splitting and the former object
// unmarshaler instead of the extracted iterator/Valid implementation. Numeric
// lexemes remain raw; object admission must not decode through float64.
func referenceRows381(payload string, p sourcejson.Policy) ([]json.RawMessage, error) {
	if int64(len(payload)) > p.MaxBytes {
		return nil, sourcejson.ErrByteLimit
	}
	rows := []json.RawMessage{}
	if payload == "" {
		return rows, nil
	}
	lines := strings.Split(payload, "\n")
	if strings.HasSuffix(payload, "\n") {
		lines = lines[:len(lines)-1]
	}
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if p.BlankLines == sourcejson.SkipBlankLines && trimmed == "" {
			continue
		}
		if len(rows) == p.MaxRows {
			return nil, sourcejson.ErrRowLimit
		}
		if p.TrimRows {
			line = trimmed
		}
		var object map[string]json.RawMessage
		if json.Unmarshal([]byte(line), &object) != nil || object == nil {
			return nil, sourcejson.ErrObject
		}
		rows = append(rows, json.RawMessage(line))
	}
	return rows, nil
}

type fragmentReader381 struct {
	payload []byte
	chunk   int
}

func (r *fragmentReader381) Read(p []byte) (int, error) {
	if len(r.payload) == 0 {
		return 0, io.EOF
	}
	n := min(len(p), r.chunk, len(r.payload))
	copy(p, r.payload[:n])
	r.payload = r.payload[n:]
	return n, nil
}

func TestHegelBoundedSourceObjects381(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		family := hegel.Draw(tc, hegel.Integers(0, 1))
		count := hegel.Draw(tc, hegel.Integers(0, 18))
		limit := hegel.Draw(tc, hegel.Integers(0, 12))
		variant := hegel.Draw(tc, hegel.Integers(0, 8))
		cut := hegel.Draw(tc, hegel.Integers(0, 3))
		chunk := hegel.Draw(tc, hegel.Integers(1, 23))
		seed := hegel.Draw(tc, hegel.Integers(0, 100000))
		p := sourcejson.Policy{MaxRows: limit, MaxBytes: 1 << 20}
		if family == 0 {
			p.BlankLines = sourcejson.SkipBlankLines
			p.TrimRows = true
		}
		// Independent lexical examples include high integers, exponent overflow,
		// duplicate keys, CRLF, missing final LF, Unicode whitespace and invalid rows.
		object := fmt.Sprintf(` {"integer":%d,"negativeZero":-0,"exp":1e400,"key":1,"key":2,"label":"雪"} `, uint64(18446744073709551615)-uint64(seed))
		lines := make([]string, count)
		for i := range lines {
			lines[i] = object
			if i%2 == 1 {
				lines[i] += "\r"
			}
		}
		if count > 0 {
			switch variant {
			case 1:
				lines[count/2] = " "
			case 2:
				lines[count/2] = "null"
			case 3:
				lines[count/2] = "[]"
			case 4:
				lines[count/2] = "{} {}"
			case 5:
				lines[count/2] = "{broken"
			case 6:
				lines[count/2] = "\u00a0{}\u00a0"
			case 7:
				lines[count/2] = `{"escaped":"a\\nb"}`
			case 8:
				lines[count/2] = "{\"invalidUtf8\":\"" + string([]byte{0xff}) + "\"}"
			}
		}
		payload := strings.Join(lines, "\n")
		if variant%2 == 0 && count > 0 {
			payload += "\n"
		}
		switch cut {
		case 0:
			p.MaxBytes = int64(len(payload))
		case 1:
			if len(payload) > 0 {
				p.MaxBytes = int64(len(payload) - 1)
			}
		case 2:
			p.MaxBytes = int64(len(payload) + 1)
		}
		want, wantErr := referenceRows381(payload, p)
		got, err := sourcejson.ReadObjects(&fragmentReader381{payload: []byte(payload), chunk: chunk}, p)
		if !errors.Is(err, wantErr) || !reflect.DeepEqual(got, want) {
			tc.Fatalf("independent admission mismatch family=%d rows=%d limit=%d variant=%d cut=%d: got %d/%v want %d/%v", family, count, limit, variant, cut, len(got), err, len(want), wantErr)
		}
		if err != nil && got != nil {
			tc.Fatal("failure admitted partial rows")
		}
		if len(got) > 1 {
			neighbor := bytes.Clone(got[1])
			got[0] = append(got[0], bytes.Repeat([]byte("x"), len(got[1])+1)...)
			if !bytes.Equal(got[1], neighbor) {
				tc.Fatal("appending a retained row overwrote another row")
			}
		}
	}, options(200)...)
}
