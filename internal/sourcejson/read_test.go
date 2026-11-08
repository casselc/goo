package sourcejson

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"math"
	"reflect"
	"strings"
	"testing"

	"github.com/casselc/goo/internal/sourcehttp"
)

func testPolicy() Policy { return Policy{MaxBytes: 1024, MaxRows: 4, BlankLines: RejectBlankLines} }
func TestReadObjectsFraming381(t *testing.T) {
	cases := []struct {
		name, body string
		policy     Policy
		want       []string
		failure    error
	}{
		{"empty", "", testPolicy(), []string{}, nil},
		{"physical rows", " \t{\"n\":9007199254740993}\r\n{\"n\":1}  ", testPolicy(), []string{" \t{\"n\":9007199254740993}\r", "{\"n\":1}  "}, nil},
		{"trim skip", "\r\n\t {\"n\":1} \r\n\u2003\n{}\n", Policy{MaxBytes: 1024, MaxRows: 2, BlankLines: SkipBlankLines, TrimRows: true}, []string{`{"n":1}`, `{}`}, nil},
		{"blank rejected", "{}\n\n", testPolicy(), nil, ErrObject},
		{"whitespace rejected", " \r\n", testPolicy(), nil, ErrObject},
		{"unicode envelope rejected", "\u2003{}\u2003", testPolicy(), nil, ErrObject},
		{"duplicate ordinary keys", `{"n":1,"n":2}`, testPolicy(), []string{`{"n":1,"n":2}`}, nil},
		{"no tail newline", "{}", testPolicy(), []string{`{}`}, nil},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			rows, err := ReadObjects(strings.NewReader(c.body), c.policy)
			if !errors.Is(err, c.failure) {
				t.Fatalf("error %v want %v", err, c.failure)
			}
			if err != nil {
				if rows != nil {
					t.Fatal("partial rows returned")
				}
				return
			}
			if rows == nil {
				t.Fatal("successful empty rows must be nonnil")
			}
			got := make([]string, len(rows))
			for i, r := range rows {
				got[i] = string(r)
			}
			if !reflect.DeepEqual(got, c.want) {
				t.Fatalf("rows %q want %q", got, c.want)
			}
		})
	}
}
func TestReadObjectsMalformedAndLimits381(t *testing.T) {
	for _, body := range []string{`null`, `[]`, `[{}]`, `true`, `0`, `"string"`, `{} {}`, `{}garbage`, `{"x":}`, "{\"x\":\"unescaped\nline\"}", "{}\r{}"} {
		rows, err := ReadObjects(strings.NewReader(body), testPolicy())
		if !errors.Is(err, ErrObject) || rows != nil {
			t.Fatalf("malformed %q: %q %v", body, rows, err)
		}
	}
	policy := Policy{MaxBytes: 2, MaxRows: 1}
	rows, err := ReadObjects(strings.NewReader(`{}`), policy)
	if err != nil || len(rows) != 1 {
		t.Fatal(rows, err)
	}
	rows, err = ReadObjects(strings.NewReader("{} "), policy)
	if !errors.Is(err, ErrByteLimit) || rows != nil {
		t.Fatal(rows, err)
	}
	policy.MaxBytes = 1024
	for _, suffix := range []string{"{}", "malformed", "\n"} {
		rows, err = ReadObjects(strings.NewReader("{}\n"+suffix), policy)
		if !errors.Is(err, ErrRowLimit) || rows != nil {
			t.Fatalf("overflow %q: %q %v", suffix, rows, err)
		}
	}
	policy.BlankLines = SkipBlankLines
	rows, err = ReadObjects(strings.NewReader("{}\n \r\n"), policy)
	if err != nil || len(rows) != 1 {
		t.Fatal(rows, err)
	}
	rows, err = ReadObjects(strings.NewReader("{}\n \r\n{"), policy)
	if !errors.Is(err, ErrRowLimit) || rows != nil {
		t.Fatal(rows, err)
	}
	for _, bad := range []Policy{{MaxBytes: -1}, {MaxBytes: math.MaxInt64}, {MaxRows: -1}, {BlankLines: 2}} {
		if rows, err := ReadObjects(strings.NewReader("{}"), bad); err == nil || rows != nil {
			t.Fatal("invalid policy admitted", bad, rows, err)
		}
	}
}

type failingReader381 struct {
	data []byte
	err  error
}

func (r *failingReader381) Read(p []byte) (int, error) {
	if len(r.data) > 0 {
		n := copy(p, r.data)
		r.data = r.data[n:]
		return n, r.err
	}
	return 0, r.err
}
func TestReadObjectsCauseAndOwnership381(t *testing.T) {
	for _, cause := range []error{io.ErrUnexpectedEOF, context.Canceled, context.DeadlineExceeded} {
		// A read failure takes precedence even if the same read exceeds the byte cap.
		rows, err := ReadObjects(&failingReader381{data: []byte("{}{}"), err: cause}, Policy{MaxBytes: 2, MaxRows: 1})
		if err != cause || rows != nil {
			t.Fatal("cause changed or partial rows", rows, err, cause)
		}
	}
	rows, err := ReadObjects(strings.NewReader("{\"a\":1}\n{\"b\":2}\n"), testPolicy())
	if err != nil {
		t.Fatal(err)
	}
	want := bytes.Clone(rows[1])
	rows[0] = append(rows[0], bytes.Repeat([]byte{'x'}, 64)...)
	if !bytes.Equal(rows[1], want) {
		t.Fatal("rows share appendable storage")
	}
}

// mapObject381 is the independent previous map decoder. These syntax cases cover
// invalid UTF-8 and surrogate escapes as well as ordinary nested/duplicate values.
func TestObjectAdmissionMatchesMapDecoder381(t *testing.T) {
	for _, body := range []string{`{}`, `{"x":null}`, `{"x":[1,2]}`, `{"x":1,"x":2}`, `{"x":"\ud800"}`, `{"\ud800":1}`, "{\"x\":\"\xff\"}", "{\"\xff\":1}", `{"n":1e100000}`, "\t{}\r\n", `null`, `[]`, `"x"`, `{} {}`} {
		var object map[string]json.RawMessage
		old := json.Unmarshal([]byte(body), &object) == nil && object != nil
		rows, err := ReadObjects(strings.NewReader(body), testPolicy())
		if (err == nil) != old {
			t.Fatalf("syntax disagreement for %q: map=%t shared=%v rows=%q", body, old, err, rows)
		}
	}
}

func benchmarkMapRows381(reader io.Reader, policy Policy) ([]json.RawMessage, error) {
	data, over, err := sourcehttp.ReadBounded(reader, policy.MaxBytes)
	if err != nil {
		return nil, err
	}
	if over {
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
		var object map[string]json.RawMessage
		if json.Unmarshal(line, &object) != nil || object == nil {
			return nil, ErrObject
		}
		rows = append(rows, json.RawMessage(bytes.Clone(bytes.TrimSuffix(line, []byte("\n")))))
	}
	return rows, nil
}
func BenchmarkReadObjects381(b *testing.B) {
	row := []byte("{\"name\":\"duration\",\"resource\":{\"service.name\":\"api\"},\"value\":\"9007199254740993\",\"attributes\":{\"route\":\"/work\"}}\n")
	data := bytes.Repeat(row, 1000)
	policy := Policy{MaxBytes: 8 << 20, MaxRows: 10000}
	for _, entry := range []struct {
		name string
		read func(io.Reader, Policy) ([]json.RawMessage, error)
	}{{"map", benchmarkMapRows381}, {"valid", ReadObjects}} {
		b.Run(entry.name, func(b *testing.B) {
			b.ReportAllocs()
			b.SetBytes(int64(len(data)))
			for b.Loop() {
				rows, err := entry.read(bytes.NewReader(data), policy)
				if err != nil || len(rows) != 1000 {
					b.Fatal(err)
				}
			}
		})
	}
}
