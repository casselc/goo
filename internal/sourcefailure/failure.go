// Package sourcefailure classifies refused source reads without interpreting
// database error prose or exposing source addresses, credentials or SQL.
package sourcefailure

import (
	"context"
	"errors"
	"net"
)

// MaxDetailBytes bounds the fixed technical explanation on the wire.
const MaxDetailBytes = 256

type Kind string

const (
	Timeout     Kind = "timeout"
	Cancelled   Kind = "cancelled"
	Budget      Kind = "budget"
	Malformed   Kind = "malformed"
	Unavailable Kind = "unavailable"
)

// Error retains the cause for errors.Is/As and internal diagnostics. Its wire
// representation is deliberately separate and never publishes Cause.Error().
type Error struct {
	Kind  Kind
	Cause error
}

func (e *Error) Error() string { return e.Cause.Error() }
func (e *Error) Unwrap() error { return e.Cause }
func Wrap(kind Kind, cause error) error {
	if cause == nil {
		return nil
	}
	return &Error{Kind: kind, Cause: cause}
}

// Protocol marks errors after successful query admission and source acquisition.
// Transport failures and cancellation retain their original classification.
func Protocol(err error) error {
	if err == nil {
		return nil
	}
	_, typed := errors.AsType[*Error](err)
	if typed || errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return err
	}
	return Wrap(Malformed, err)
}

type Refusal struct {
	Kind   Kind   `json:"kind"`
	Detail string `json:"detail"`
}
type Response struct {
	Error   string  `json:"error"`
	Refusal Refusal `json:"refusal"`
}

// Describe is a closed, bounded wire vocabulary. It makes no prediction about
// whether another explicit read of the same question will succeed.
func Describe(err error) Response {
	kind := Unavailable
	failure, typed := errors.AsType[*Error](err)
	network, networkError := errors.AsType[net.Error](err)
	switch {
	case errors.Is(err, context.Canceled):
		kind = Cancelled
	case errors.Is(err, context.DeadlineExceeded):
		kind = Timeout
	case networkError && network.Timeout():
		kind = Timeout
	case typed:
		switch failure.Kind {
		case Timeout, Cancelled, Budget, Malformed, Unavailable:
			kind = failure.Kind
		}
	}
	reason, detail := "Source unavailable.", "The source request could not be completed."
	switch kind {
	case Timeout:
		reason, detail = "Source read timed out.", "The source read exceeded its time limit."
	case Cancelled:
		reason, detail = "Source read cancelled.", "The source read was cancelled before completion."
	case Budget:
		reason, detail = "Source read exceeded a limit.", "The source response exceeded a bounded read or grouping limit. Try a narrower scope."
	case Malformed:
		reason, detail = "Source response rejected.", "The source returned invalid or unsupported record data. No partial result was admitted."
	}
	return Response{Error: reason, Refusal: Refusal{Kind: kind, Detail: detail}}
}
