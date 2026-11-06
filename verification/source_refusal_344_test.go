package verification

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"
	"unicode"
	"unicode/utf8"

	"github.com/casselc/goo/internal/sourcefailure"
	"hegel.dev/go/hegel"
)

type refusalTimeout344 struct{ error }

func (refusalTimeout344) Timeout() bool   { return true }
func (refusalTimeout344) Temporary() bool { return false }
func (e refusalTimeout344) Unwrap() error { return e.error }

func TestHegelSourceRefusal344(t *testing.T) {
	kinds := []sourcefailure.Kind{sourcefailure.Cancelled, sourcefailure.Timeout, sourcefailure.Budget, sourcefailure.Malformed, sourcefailure.Unavailable}
	hegel.Test(t, func(tc *hegel.T) {
		choice := hegel.Draw(tc, hegel.Integers(0, 8))
		depth := hegel.Draw(tc, hegel.Integers(0, 12))
		// Deliberately fake transport/SQL material must stay internal regardless
		// of wrapping depth, classification or cause precedence.
		leaf := errors.New("https://fixture-user:fixture-pass@source.invalid/?token=fake-secret SELECT private_fixture")
		want := sourcefailure.Unavailable
		var err error
		switch {
		case choice < len(kinds):
			want = kinds[choice]
			err = sourcefailure.Wrap(want, leaf)
		case choice == 5:
			want = sourcefailure.Cancelled
			err = errors.Join(sourcefailure.Wrap(sourcefailure.Budget, leaf), context.Canceled)
		case choice == 6:
			want = sourcefailure.Timeout
			err = errors.Join(sourcefailure.Wrap(sourcefailure.Malformed, leaf), context.DeadlineExceeded)
		case choice == 7:
			want = sourcefailure.Timeout
			err = sourcefailure.Wrap(sourcefailure.Unavailable, refusalTimeout344{leaf})
		default:
			err = sourcefailure.Wrap(sourcefailure.Kind("unknown"), leaf)
		}
		for i := range depth {
			err = fmt.Errorf("read stage %d: %w", i, err)
		}
		if !errors.Is(err, leaf) {
			tc.Fatal("internal cause identity lost")
		}
		response := sourcefailure.Describe(err)
		if response.Refusal.Kind != want || response.Error == "" || response.Refusal.Detail == "" || len(response.Refusal.Detail) > sourcefailure.MaxDetailBytes || !utf8.ValidString(response.Refusal.Detail) || strings.ContainsFunc(response.Refusal.Detail, unicode.IsControl) {
			tc.Fatal("invalid refusal vocabulary", response, want)
		}
		wire, e := json.Marshal(response)
		if e != nil {
			tc.Fatal(e)
		}
		for _, forbidden := range []string{"fixture-user", "fixture-pass", "source.invalid", "fake-secret", "SELECT", "private_fixture"} {
			if strings.Contains(string(wire), forbidden) {
				tc.Fatal("internal source detail escaped onto wire", forbidden)
			}
		}
		if !errors.Is(sourcefailure.Protocol(err), leaf) {
			tc.Fatal("protocol classification erased internal cause")
		}
	}, options(60)...)
}
