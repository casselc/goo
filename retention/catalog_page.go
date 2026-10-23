package retention

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"sort"

	"github.com/casselc/goo/workbench"
)

// paginateCatalog preserves the original v1 fingerprint/cursor JSON. Origin
// adapters provide their own typed rows/filter, without converting tail custody
// metadata into a captured result page.
func paginateCatalog[T any](filter any, cursorText string, items []T, id func(T) string) ([]T, string, error) {
	sort.Slice(items, func(i, j int) bool { return id(items[i]) < id(items[j]) })
	raw, _ := json.Marshal(struct {
		Query any
		Items []T
	}{filter, items})
	fingerprint := hash(raw)
	cursor := struct {
		After       string
		Fingerprint string
	}{}
	if cursorText != "" {
		raw, e := base64.RawURLEncoding.DecodeString(cursorText)
		if e != nil || json.Unmarshal(raw, &cursor) != nil || !workbench.ID.MatchString(cursor.After) || cursor.Fingerprint != fingerprint {
			return nil, "", errors.New("catalog changed or filter differs; restart browsing")
		}
	}
	page := []T{}
	for _, item := range items {
		if id(item) <= cursor.After {
			continue
		}
		if len(page) == 20 {
			cursor.After = id(page[19])
			cursor.Fingerprint = fingerprint
			raw, _ = json.Marshal(cursor)
			return page, base64.RawURLEncoding.EncodeToString(raw), nil
		}
		page = append(page, item)
	}
	return page, "", nil
}

// Records here are already source-adapted immutable evidence. A missing service
// and a present empty service remain distinct in both catalog families.
func retainedServices(records []json.RawMessage) (map[string]bool, error) {
	services := map[string]bool{}
	for _, record := range records {
		var row struct {
			Service *string `json:"ServiceName"`
		}
		if e := json.Unmarshal(record, &row); e != nil {
			return nil, e
		}
		if row.Service != nil {
			services[*row.Service] = true
		}
	}
	return services, nil
}
