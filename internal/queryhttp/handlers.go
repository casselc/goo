package queryhttp

import (
	"context"
	"net/http"
	"time"

	"github.com/casselc/goo/internal/httpjson"
)

// Effect describes a family's request lifetime, not authentication or rollback.
type Effect string

const (
	Pure                    Effect = "pure"
	Read                    Effect = "source-read"
	ReadAndSnapshot         Effect = "source-read-and-snapshot"
	ReadWithOptionalCustody Effect = "source-read-with-optional-custody"
	ReadOrPromote           Effect = "source-read-and-publication-or-held-promotion"
	ReadOrFreeze            Effect = "source-read-and-draft-or-held-freeze"
)

// Policy is returned by family-owned method/path declarations. Source timeouts
// begin after validation; pure planner timeouts include decoding and computation.
// Mixed publication policies apply Timeout only to their source branches and
// retain family-owned publication lifetimes and error handling.
type Policy struct {
	Effect          Effect
	BodyLimit       int64
	Timeout         time.Duration
	TrailingMessage string
}

// ReadHandler shares typed wire mechanics, leaving validation and effects in the
// supplied family functions. Type parameters are inferred at registration sites.
func ReadHandler[Q, R any](policy Policy, validate func(Q) error, read func(context.Context, Q) (R, error)) http.HandlerFunc {
	switch policy.Effect {
	case Read, ReadAndSnapshot, ReadWithOptionalCustody:
	default:
		panic("queryhttp: source-read policy required; mixed publication routes own their handlers")
	}
	if policy.Timeout <= 0 {
		panic("queryhttp: source-read policy required")
	}
	return func(w http.ResponseWriter, r *http.Request) {
		fail := func(code int, err error) { httpjson.Write(w, code, map[string]string{"error": err.Error()}) }
		var question Q
		if err := httpjson.Decode(w, r, &question, policy.BodyLimit, policy.TrailingMessage); err != nil {
			fail(400, err)
			return
		}
		if err := validate(question); err != nil {
			fail(400, err)
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), policy.Timeout)
		defer cancel()
		result, err := read(ctx, question)
		if err != nil {
			fail(502, err)
			return
		}
		httpjson.Write(w, 200, result)
	}
}

// PlanHandler preserves pre-decode and pre/post-computation deadline checks.
// A synchronous pure planner is not preempted in the middle of its computation.
func PlanHandler[Q, R any](policy Policy, plan func(Q) (R, error)) http.HandlerFunc {
	if policy.Effect != Pure || policy.Timeout <= 0 {
		panic("queryhttp: bounded pure-plan policy required")
	}
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), policy.Timeout)
		defer cancel()
		fail := func(err error) { httpjson.Write(w, 400, map[string]string{"error": err.Error()}) }
		var input Q
		if err := httpjson.Decode(w, r, &input, policy.BodyLimit, policy.TrailingMessage); err != nil {
			fail(err)
			return
		}
		if err := ctx.Err(); err != nil {
			fail(err)
			return
		}
		result, err := plan(input)
		if err != nil {
			fail(err)
			return
		}
		if err := ctx.Err(); err != nil {
			fail(err)
			return
		}
		httpjson.Write(w, 200, result)
	}
}
