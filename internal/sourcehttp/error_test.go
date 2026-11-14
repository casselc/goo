package sourcehttp

import (
	"net/http"
	"strings"
	"testing"
)

func TestFailureRetainsOnlyStructuredMetadata(t *testing.T) {
	for _, tc := range []struct {
		header string
		code   int
	}{{"60", 60}, {"497", 497}, {"", 0}, {"bad password secret SQL", 0}, {"-60", 0}, {"999999", 0}} {
		r := &http.Response{StatusCode: 404, Header: http.Header{"X-Clickhouse-Exception-Code": []string{tc.header}}}
		e := Failure(r)
		if e.Status != 404 || e.Code != tc.code || strings.Contains(e.Error(), "secret") {
			t.Fatalf("unsafe or incorrect error: %#v", e)
		}
	}
}
