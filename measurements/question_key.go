package measurements

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
)

// Key validates a fixed native measurement question before identifying it. It
// retains the existing JSON digest, including declared selectors and exact bounds,
// without normalizing equivalent windows or reading telemetry. Invalid questions
// return no key. A question key is distinct from a definition revision or result basis.
func (q Question) Key() (string, error) {
	if err := q.Validate(); err != nil {
		return "", fmt.Errorf("question key: %w", err)
	}
	wire, err := json.Marshal(q)
	if err != nil {
		return "", fmt.Errorf("encode question key: %w", err)
	}
	sum := sha256.Sum256(wire)
	return hex.EncodeToString(sum[:]), nil
}
