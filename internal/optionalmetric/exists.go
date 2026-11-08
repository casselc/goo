// Package optionalmetric shares the fixed optional metric-table probe contract.
// Callers retain their own transport, scope, budgets and error classification.
package optionalmetric

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

const ExponentialHistogramProbeSQL = "EXISTS TABLE otel_metrics_exp_histogram FORMAT JSONEachRow"

// DecodeExists accepts exactly one result field with a canonical numeric 0/1.
func DecodeExists(rows []json.RawMessage) (bool, error) {
	invalid := func() (bool, error) { return false, fmt.Errorf("invalid optional metric table availability response") }
	if len(rows) != 1 {
		return invalid()
	}
	d := json.NewDecoder(bytes.NewReader(rows[0]))
	token, err := d.Token()
	if err != nil || token != json.Delim('{') {
		return invalid()
	}
	token, err = d.Token()
	if err != nil || token != "result" {
		return invalid()
	}
	var raw json.RawMessage
	if d.Decode(&raw) != nil {
		return invalid()
	}
	token, err = d.Token()
	if err != nil || token != json.Delim('}') {
		return invalid()
	}
	if _, err = d.Token(); err != io.EOF {
		return invalid()
	}
	switch strings.TrimSpace(string(raw)) {
	case "0":
		return false, nil
	case "1":
		return true, nil
	default:
		return invalid()
	}
}
