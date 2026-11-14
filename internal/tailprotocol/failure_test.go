package tailprotocol

import (
	"context"
	"encoding/json"
	"errors"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/casselc/goo/internal/sourcefailure"
	"github.com/casselc/goo/internal/sourcehttp"
)

func TestTailFailureClosed433(t *testing.T) {
	for _, tc := range []struct {
		err    error
		reason string
		retry  Retry
		kind   sourcefailure.Kind
	}{
		{&sourcehttp.HTTPError{Status: 404, Code: 60}, "source_missing", Explicit, sourcefailure.Unavailable},
		{&sourcehttp.HTTPError{Status: 500, Code: 81}, "source_missing", Explicit, sourcefailure.Unavailable},
		{&sourcehttp.HTTPError{Status: 403}, "source_access", Explicit, sourcefailure.Unavailable},
		{&sourcehttp.HTTPError{Status: 500, Code: 497}, "source_access", Explicit, sourcefailure.Unavailable},
		{&sourcehttp.HTTPError{Status: 500, Code: 47}, "source_schema", Explicit, sourcefailure.Unavailable},
		{&sourcehttp.HTTPError{Status: 503}, "source_unavailable", Explicit, sourcefailure.Unavailable},
		{errors.New("password SELECT privatehost"), "source_unavailable", Explicit, sourcefailure.Unavailable},
		{&net.DNSError{Err: "password SELECT privatehost"}, "source_unavailable", Automatic, sourcefailure.Unavailable},
		{&net.DNSError{Err: "password SELECT privatehost", IsNotFound: true}, "source_unavailable", Explicit, sourcefailure.Unavailable},
		{&net.DNSError{Err: "password SELECT privatehost", IsNotFound: true, IsTimeout: true}, "source_timeout", Automatic, sourcefailure.Timeout},
		{context.DeadlineExceeded, "source_timeout", Automatic, sourcefailure.Timeout},
		{context.Canceled, "source_cancelled", Explicit, sourcefailure.Cancelled},
		{sourcefailure.Wrap(sourcefailure.Malformed, errors.New("password SELECT privatehost")), "source_malformed", Explicit, sourcefailure.Malformed},
		{sourcefailure.Wrap(sourcefailure.Budget, errors.New("password SELECT privatehost")), "source_budget", Explicit, sourcefailure.Budget},
	} {
		got := Source(tc.err)
		if got.Reason != tc.reason || got.Retry != tc.retry || got.Refusal == nil || got.Refusal.Kind != tc.kind {
			t.Fatalf("%+v", got)
		}
		data, _ := json.Marshal(got)
		for _, secret := range []string{"password", "SELECT", "privatehost"} {
			if strings.Contains(string(data), secret) {
				t.Fatal(string(data))
			}
		}
	}
	if got := IdentityBudget(); got.Refusal == nil || got.Refusal.Kind != sourcefailure.Budget || got.Retry != Explicit {
		t.Fatal(got)
	}
	parent, cancel := context.WithCancel(context.Background())
	connection, stop := context.WithCancel(parent)
	stop()
	if _, ok := End(parent, connection); ok {
		t.Fatal("cancel interpreted as rollover")
	}
	cancel()
	if _, ok := End(parent, connection); ok {
		t.Fatal("caller cancellation emitted rollover")
	}
	expired, stop := context.WithDeadline(context.Background(), time.Unix(0, 0))
	defer stop()
	if got, ok := End(context.Background(), expired); !ok || got != Ended() {
		t.Fatal(got, ok)
	}
}
