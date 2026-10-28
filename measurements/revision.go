package measurements

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
)

// Revision validates an authored definition and returns its existing JSON-based
// identity. It preserves nil/empty selectors and declared names without adding
// defaults or reading telemetry. Invalid definitions have no revision.
func (d Definition) Revision() (string, error) {
	if err := d.Validate(); err != nil {
		return "", fmt.Errorf("definition revision: %w", err)
	}
	wire, err := json.Marshal(d)
	if err != nil {
		return "", fmt.Errorf("encode definition revision: %w", err)
	}
	sum := sha256.Sum256(wire)
	return hex.EncodeToString(sum[:]), nil
}
