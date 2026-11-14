package sourcehttp

import (
	"fmt"
	"net/http"
	"strconv"
)

// HTTPError retains only structured status metadata. Neither database prose nor
// response bodies enter the error chain or a browser-facing diagnostic.
type HTTPError struct{ Status, Code int }

func (e *HTTPError) Error() string {
	return fmt.Sprintf("source HTTP %d (exception code %d)", e.Status, e.Code)
}

// Failure reads the ClickHouse exception header, never error text. Missing or
// invalid codes remain unknown; callers must not infer table absence from them.
func Failure(response *http.Response) *HTTPError {
	code, err := strconv.Atoi(response.Header.Get("X-ClickHouse-Exception-Code"))
	if err != nil || code < 1 || code > 65535 {
		code = 0
	}
	return &HTTPError{Status: response.StatusCode, Code: code}
}
