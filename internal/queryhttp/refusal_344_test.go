package queryhttp

import (
	"context"
	"encoding/json"
	"errors"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/casselc/goo/internal/sourcefailure"
)

func TestReadHandlerTypedRefusal344(t *testing.T) {
	secret := errors.New("http://user:password@private.example/?query=SELECT+secret")
	for _, kind := range []sourcefailure.Kind{sourcefailure.Timeout, sourcefailure.Cancelled, sourcefailure.Budget, sourcefailure.Malformed, sourcefailure.Unavailable} {
		t.Run(string(kind), func(t *testing.T) {
			effects := 0
			h := ReadHandler(Policy{Effect: Read, BodyLimit: 1024, Timeout: time.Second}, func(input) error { return nil }, func(context.Context, input) (input, error) {
				effects++
				return input{}, sourcefailure.Wrap(kind, secret)
			})
			w := httptest.NewRecorder()
			h(w, httptest.NewRequest("POST", "/read", strings.NewReader(`{"Value":1}`)))
			var response sourcefailure.Response
			if w.Code != 502 || json.Unmarshal(w.Body.Bytes(), &response) != nil || response.Refusal.Kind != kind || effects != 1 || strings.Contains(w.Body.String(), "private.example") {
				t.Fatal(w.Code, w.Body.String(), effects)
			}
			if w.Header().Get("Cache-Control") != "no-store" {
				t.Fatal("cache policy changed")
			}
			w = httptest.NewRecorder()
			h(w, httptest.NewRequest("POST", "/read", strings.NewReader(`{"unknown":1}`)))
			if w.Code != 400 || effects != 1 || strings.Contains(w.Body.String(), `"refusal"`) {
				t.Fatal("strict admission changed", w.Code, w.Body.String(), effects)
			}
		})
	}
}
