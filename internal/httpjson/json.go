// Package httpjson shares bounded JSON mechanics for explicitly compatible routes.
// Callers own semantic validation, effects, status mapping and per-route policy.
package httpjson

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

// Decode rejects unknown fields and any input after the first JSON value except
// whitespace. limit and trailingMessage are caller policy. It intentionally keeps
// encoding/json's typed null, duplicate-field, number and content-type behavior.
// A target may be partly populated on error; callers must stop before effects.
func Decode(w http.ResponseWriter, r *http.Request, target any, limit int64, trailingMessage string) error {
	d := json.NewDecoder(http.MaxBytesReader(w, r.Body, limit))
	d.DisallowUnknownFields()
	if e := d.Decode(target); e != nil {
		return e
	}
	if d.Decode(new(any)) != io.EOF {
		return errors.New(trailingMessage)
	}
	return nil
}

// Write matches the adopted routes' JSON/no-store envelope. Like their original
// writers, it leaves encoding/write errors unreported. It is not an SSE responder.
func Write(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(value)
}
