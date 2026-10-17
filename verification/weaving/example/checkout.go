package example

import (
	"context"
	"errors"
)

type Decision struct {
	ID      string
	Outcome string
}

//goo:operation name:checkout.reserve role:server
func Reserve(ctx context.Context, fail bool) (result *Decision, err error) {
	result, err = Accept(ctx, fail)
	if err != nil {
		return nil, err
	}
	err = Complete(ctx, result)
	return result, err
}

//goo:event name:reservation.accepted subject:result.ID when:ok
func Accept(ctx context.Context, fail bool) (result *Decision, err error) {
	if fail {
		return nil, errors.New("private failure detail")
	}
	return &Decision{ID: "reservation-1", Outcome: "confirmed"}, nil
}

//goo:event name:reservation.terminal subject:decision.ID outcome:decision.Outcome when:ok
func Complete(ctx context.Context, decision *Decision) (err error) { return nil }

//goo:operation name:checkout.panic role:internal
func Panic(ctx context.Context) { panic("original panic") }
