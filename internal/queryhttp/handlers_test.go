package queryhttp

import (
	"context"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/synctest"
	"time"
)

type delayedJSON struct {
	input   *strings.Reader
	delayed bool
}

func (b *delayedJSON) Read(p []byte) (int, error) {
	if !b.delayed {
		b.delayed = true
		time.Sleep(20 * time.Second)
	}
	return b.input.Read(p)
}

type input struct{ Value int }

func TestReadDeadlineStartsAfterDecodeAndValidation(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		var owned context.Context
		handler := ReadHandler(Policy{Effect: Read, BodyLimit: 1024, Timeout: 10 * time.Second, TrailingMessage: "one object"}, func(input) error {
			time.Sleep(20 * time.Second)
			return nil
		}, func(ctx context.Context, in input) (input, error) {
			owned = ctx
			deadline, ok := ctx.Deadline()
			if !ok || time.Until(deadline) != 10*time.Second || ctx.Err() != nil {
				t.Fatal("decoding or validation consumed source deadline")
			}
			return in, nil
		})
		w := httptest.NewRecorder()
		handler(w, httptest.NewRequest("POST", "/read", &delayedJSON{input: strings.NewReader(`{"Value":1}`)}))
		if w.Code != 200 || owned == nil || owned.Err() != context.Canceled {
			t.Fatal("read did not finish and release its context", w.Code)
		}
	})
}

func TestPlanDeadlineIncludesDecodeAndComputation(t *testing.T) {
	for _, phase := range []string{"decode", "compute"} {
		t.Run(phase, func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				called := false
				handler := PlanHandler(Policy{Effect: Pure, BodyLimit: 1024, Timeout: 10 * time.Second, TrailingMessage: "one object"}, func(in input) (input, error) { called = true; time.Sleep(20 * time.Second); return in, nil })
				r := httptest.NewRequest("POST", "/plan", strings.NewReader(`{"Value":1}`))
				if phase == "decode" {
					r = httptest.NewRequest("POST", "/plan", &delayedJSON{input: strings.NewReader(`{"Value":1}`)})
				}
				w := httptest.NewRecorder()
				handler(w, r)
				if w.Code != 400 || !strings.Contains(w.Body.String(), context.DeadlineExceeded.Error()) {
					t.Fatal("expired plan was returned", w.Code, w.Body.String())
				}
				if called != (phase == "compute") {
					t.Fatal("planner ran after decode deadline")
				}
			})
		})
	}
}

func TestReadDeadlineCannotExtendParent(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		parent, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		handler := ReadHandler(Policy{Effect: Read, BodyLimit: 1024, Timeout: 10 * time.Second, TrailingMessage: "one object"}, func(input) error { return nil }, func(ctx context.Context, in input) (input, error) {
			deadline, _ := ctx.Deadline()
			if time.Until(deadline) != 5*time.Second {
				t.Fatal("extended parent deadline")
			}
			time.Sleep(6 * time.Second)
			return in, ctx.Err()
		})
		w := httptest.NewRecorder()
		handler(w, httptest.NewRequest("POST", "/read", strings.NewReader(`{"Value":1}`)).WithContext(parent))
		if w.Code != 502 || !strings.Contains(w.Body.String(), context.DeadlineExceeded.Error()) {
			t.Fatal("source deadline failure changed", w.Code, w.Body.String())
		}
	})
}
