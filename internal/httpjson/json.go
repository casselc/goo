// Package httpjson shares bounded JSON mechanics for explicitly compatible routes.
// Callers own semantic validation, effects, status mapping and per-route policy.
package httpjson

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"
)

// Decode rejects unknown fields and any input after the first JSON value except
// whitespace. limit and trailingMessage are caller policy. It intentionally keeps
// encoding/json's typed null, duplicate-field, number and content-type behavior.
// A target may be partly populated on error; callers must stop before effects.
func Decode(w http.ResponseWriter, r *http.Request, target any, limit int64, trailingMessage string) error {
	return decodeWithin(w, r, target, limit, trailingMessage, 15*time.Second)
}

// Bound transport receive independently of execution and response streaming.
// ResponseController unwraps compatible middleware. In-memory test responders
// have no transport deadline; their bodies must already be bounded readers.
func decodeWithin(w http.ResponseWriter, r *http.Request, target any, limit int64, trailingMessage string, receiveTimeout time.Duration) (err error) {
	controller := http.NewResponseController(w)
	if err := controller.SetReadDeadline(time.Now().Add(receiveTimeout)); err != nil && !errors.Is(err, http.ErrNotSupported) {
		return err
	}
	done := make(chan struct{})
	stop := context.AfterFunc(r.Context(), func() {
		defer close(done)
		_ = controller.SetReadDeadline(time.Now())
	})
	defer func() {
		// Join a running callback before clearing, so it cannot install an expired
		// deadline after the handler has advanced to its response phase.
		if !stop() {
			<-done
		}
		if err != nil {
			// An early syntax error can leave bytes unread. Do not let net/http's
			// implicit drain block while writing the error response. This request
			// cannot safely reuse its HTTP/1 connection after an incomplete body.
			r.Close = true
			w.Header().Set("Connection", "close")
			_ = controller.SetReadDeadline(time.Now())
		} else {
			_ = controller.SetReadDeadline(time.Time{})
		}
	}()
	if err := r.Context().Err(); err != nil {
		return err
	}
	d := json.NewDecoder(http.MaxBytesReader(w, r.Body, limit))
	d.DisallowUnknownFields()
	if e := d.Decode(target); e != nil {
		return e
	}
	if d.Decode(new(any)) != io.EOF {
		return errors.New(trailingMessage)
	}
	return r.Context().Err()
}

// Write matches the adopted routes' JSON/no-store envelope. Like their original
// writers, it leaves encoding/write errors unreported. It is not an SSE responder.
func Write(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(value)
}
