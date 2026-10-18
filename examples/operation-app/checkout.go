package main

import (
	"context"
	"crypto/rand"
	"errors"
	"github.com/casselc/goo/instrumentation"
	"log"
	"os"
	"time"
)

type subjectIDKey struct{}

type Decision struct {
	ID      string
	Outcome string
}

//goo:operation name:checkout.reserve role:server
func Reserve(ctx context.Context, fail bool) (result *Decision, err error) {
	err = CheckStock(ctx, fail)
	if err != nil {
		return nil, err
	}
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
	id, _ := ctx.Value(subjectIDKey{}).(string)
	if id == "" {
		id = rand.Text()
	}
	emitLog(ctx, "reservation.accepted", id, "")
	return &Decision{ID: id, Outcome: "confirmed"}, nil
}

//goo:event name:reservation.terminal subject:decision.ID outcome:decision.Outcome when:ok
func Complete(ctx context.Context, decision *Decision) (err error) {
	emitLog(ctx, "reservation.terminal", decision.ID, decision.Outcome)
	return nil
}
func emitLog(ctx context.Context, name, id, outcome string) {
	instance := os.Getenv("GOO_INSTANCE_ID")
	if instance == "" {
		instance = "woven-demo"
	}
	if e := instrumentation.EmitExampleLog(ctx, "lab-go-checkout", instance, name, id, outcome); e != nil {
		log.Printf("example log export failed: %v", e)
	}
}

//goo:operation name:checkout.panic role:internal
func Panic(ctx context.Context) { panic("original panic") }

//goo:operation name:checkout.stock role:internal
func CheckStock(ctx context.Context, fail bool) (err error) {
	// Bounded simulated dependency work gives the demo distinct raw duration bands.
	if fail {
		time.Sleep(25 * time.Millisecond)
		return errors.New("private failure detail")
	}
	time.Sleep(2 * time.Millisecond)
	return nil
}
