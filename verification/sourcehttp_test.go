package verification

import (
	"bytes"
	"errors"
	"io"
	"testing"

	"github.com/casselc/goo/internal/sourcehttp"
	"hegel.dev/go/hegel"
)

type sourceCountReader struct {
	reader io.Reader
	read   int
}

func (r *sourceCountReader) Read(p []byte) (int, error) {
	n, e := r.reader.Read(p)
	r.read += n
	return n, e
}

type sourceFailReader struct{ err error }

func (r sourceFailReader) Read([]byte) (int, error) { return 0, r.err }
func TestHegelSourceHTTPByteBounds(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		size := hegel.Draw(tc, hegel.Integers(0, 32768))
		limit := hegel.Draw(tc, hegel.Integers(0, 32768))
		value := byte(hegel.Draw(tc, hegel.Integers(0, 255)))
		payload := bytes.Repeat([]byte{value}, size)
		reader := &sourceCountReader{reader: bytes.NewReader(payload)}
		got, oversize, e := sourcehttp.ReadBounded(reader, int64(limit))
		if e != nil || oversize != (size > limit) || reader.read > limit+1 {
			tc.Fatal("byte budget changed", e)
		}
		if oversize {
			if got != nil {
				tc.Fatal("partial oversized payload returned")
			}
		} else if !bytes.Equal(got, payload) {
			tc.Fatal("payload bytes changed")
		}
		failure := errors.New("generated body failure")
		prefix := payload
		if len(prefix) > limit {
			prefix = prefix[:limit]
		}
		reader = &sourceCountReader{reader: io.MultiReader(bytes.NewReader(prefix), sourceFailReader{failure})}
		got, oversize, e = sourcehttp.ReadBounded(reader, int64(limit))
		if !errors.Is(e, failure) || got != nil || oversize || reader.read > limit+1 {
			tc.Fatal("read error lost or partial payload returned", e)
		}
	}, options(100)...)
}
