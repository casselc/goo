package verification

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/casselc/goo/internal/immutablefile"
	"hegel.dev/go/hegel"
)

func TestHegelImmutablePublicationSequence(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		dir := t.TempDir()
		size := hegel.Draw(tc, hegel.Integers(0, 4096))
		value := hegel.Draw(tc, hegel.Integers(0, 255))
		attempts := hegel.Draw(tc, hegel.Integers(2, 12))
		winner := bytes.Repeat([]byte{byte(value)}, size)
		for i := 0; i < attempts; i++ {
			input := winner
			if i > 0 {
				input = bytes.Repeat([]byte{byte(value + i)}, size+i)
			}
			created, e := immutablefile.Publish(dir, "winner.json", ".draft-", input)
			if e != nil || created != (i == 0) {
				tc.Fatal("publication outcome", i, created, e)
			}
			if e := immutablefile.PublishNew(dir, "winner.json", ".capture-", input); !os.IsExist(e) {
				tc.Fatal("exclusive duplicate accepted", e)
			}
			got, e := os.ReadFile(filepath.Join(dir, "winner.json"))
			if e != nil || !bytes.Equal(got, winner) {
				tc.Fatal("winner changed", i, e)
			}
		}
		entries, e := os.ReadDir(dir)
		if e != nil || len(entries) != 1 {
			tc.Fatal("scratch leaked", e)
		}
	}, options(100)...)
}
