package verification

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/casselc/goo/internal/httpjson"
	"hegel.dev/go/hegel"
)

func TestHegelPlanningJSONBoundaries(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		type input struct {
			Label string `json:"label"`
			Value int64  `json:"value"`
		}
		alphabet := []string{"", "plain", "é", "雪", "<>&", "\"\n"}
		label := strings.Repeat(alphabet[hegel.Draw(tc, hegel.Integers(0, len(alphabet)-1))], hegel.Draw(tc, hegel.Integers(0, 100)))
		want := input{Label: label, Value: 9007199254740993 + int64(hegel.Draw(tc, hegel.Integers(0, 100000)))}
		raw, e := json.Marshal(want)
		if e != nil {
			tc.Fatal(e)
		}
		whitespace := []string{" ", "\t", "\r", "\n"}[hegel.Draw(tc, hegel.Integers(0, 3))]
		body := string(raw) + strings.Repeat(whitespace, hegel.Draw(tc, hegel.Integers(0, 64)))
		message := []string{"require one JSON object", "expected one JSON object"}[hegel.Draw(tc, hegel.Integers(0, 1))]
		decode := func(text string, limit int64) (input, error) {
			var got input
			w := httptest.NewRecorder()
			r := httptest.NewRequest("POST", "/plan", strings.NewReader(text))
			e := httpjson.Decode(w, r, &got, limit, message)
			return got, e
		}
		got, e := decode(body, int64(len(body)))
		if e != nil || got != want {
			tc.Fatal("exact byte limit changed typed payload", e)
		}
		if _, e = decode(body+" ", int64(len(body))); e == nil || e.Error() != message {
			tc.Fatal("over-limit trailing policy changed", e)
		}
		if _, e = decode(body+" {}", int64(len(body)+3)); e == nil || e.Error() != message {
			tc.Fatal("second value accepted", e)
		}
		if _, e = decode(string(raw), int64(len(raw)-1)); e == nil {
			tc.Fatal("truncated byte budget accepted")
		}
		unknown := string(raw[:len(raw)-1]) + `,"foreign":1}`
		if _, e = decode(unknown, int64(len(unknown))); e == nil || !strings.Contains(e.Error(), `unknown field "foreign"`) {
			tc.Fatal("unknown field accepted", e)
		}
		w := httptest.NewRecorder()
		httpjson.Write(w, 200, want)
		if w.Code != 200 || !bytes.Equal(w.Body.Bytes(), append(raw, '\n')) || w.Header().Get("Content-Type") != "application/json" || w.Header().Get("Cache-Control") != "no-store" {
			tc.Fatal("JSON response changed")
		}
	}, options(100)...)
}
