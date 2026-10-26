package metrics

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

func serviceObservationIdentity(s Stream) bool {
	return s.Resource["service.name"] != "" && s.Resource["service.instance.id"] != ""
}
func validateHostObservationID(id string) error {
	if id == "" || len(id) > 512 || !utf8.ValidString(id) || strings.ContainsRune(id, 0) {
		return fmt.Errorf("require a nonempty bounded host.id")
	}
	return nil
}
