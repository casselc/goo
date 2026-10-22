// Package sourcehttp shares source HTTP mechanics without query, status, row or
// evidence semantics. Callers own URL policy, client timeouts and response closing.
package sourcehttp

import (
	"context"
	"io"
	"net/http"
	"strings"
)

// Post sends SQL verbatim with the caller's context, non-nil client and Basic Auth.
// The caller must close a successful response body and choose when to check status.
func Post(ctx context.Context, client *http.Client, address, user, password, sql string) (*http.Response, error) {
	request, e := http.NewRequestWithContext(ctx, "POST", address, strings.NewReader(sql))
	if e != nil {
		return nil, e
	}
	request.SetBasicAuth(user, password)
	return client.Do(request)
}

// ReadBounded consumes at most limit+1 bytes. Oversize and read-error outcomes
// return no partial payload; callers map them to family-specific errors. limit is
// a caller-owned bound satisfying 0 <= limit < MaxInt64. This buffering primitive
// is not an SSE reader.
func ReadBounded(reader io.Reader, limit int64) (data []byte, oversize bool, err error) {
	data, err = io.ReadAll(io.LimitReader(reader, limit+1))
	if err != nil {
		return nil, false, err
	}
	if int64(len(data)) > limit {
		return nil, true, nil
	}
	return data, false, nil
}
