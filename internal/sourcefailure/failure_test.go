package sourcefailure

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"strings"
	"testing"
)

func TestClosedRefusalClassification344(t *testing.T) {
	secret := errors.New("http://user:password@private-source/?query=SELECT+secret")
	cases := []struct {
		name string
		err  error
		kind Kind
	}{
		{"unknown", secret, Unavailable},
		{"typed budget", Wrap(Budget, secret), Budget},
		{"typed malformed", Wrap(Malformed, secret), Malformed},
		{"unknown typed kind", Wrap(Kind("invented"), secret), Unavailable},
		{"wrapped cancellation", Wrap(Unavailable, fmt.Errorf("wrapped: %w", context.Canceled)), Cancelled},
		{"wrapped deadline", Wrap(Unavailable, fmt.Errorf("wrapped: %w", context.DeadlineExceeded)), Timeout},
		{"network timeout", Wrap(Unavailable, &net.DNSError{Err: "timeout secret", IsTimeout: true}), Timeout},
		{"text is not authority", errors.New("context deadline exceeded source budget malformed"), Unavailable},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := Describe(tc.err)
			if got.Refusal.Kind != tc.kind || len(got.Refusal.Detail) == 0 || len(got.Refusal.Detail) > MaxDetailBytes {
				t.Fatal(got)
			}
			wire, err := json.Marshal(got)
			if err != nil {
				t.Fatal(err)
			}
			for _, word := range []string{"private-source", "password", "SELECT", "secret", "retry"} {
				if strings.Contains(string(wire), word) {
					t.Fatal("unsafe or unsupported wire field", string(wire))
				}
			}
		})
	}
	if !errors.Is(Wrap(Budget, secret), secret) {
		t.Fatal("lost cause")
	}
}
func TestProtocolBoundaryClassification344(t *testing.T) {
	cause := errors.New("invalid sum/count")
	if Describe(Protocol(cause)).Refusal.Kind != Malformed || !errors.Is(Protocol(cause), cause) {
		t.Fatal("protocol error not retained")
	}
	for _, err := range []error{Wrap(Budget, cause), Wrap(Unavailable, cause), context.Canceled, context.DeadlineExceeded} {
		if Protocol(err) != err {
			t.Fatal("known classification changed")
		}
	}
	if Wrap(Budget, nil) != nil || Protocol(nil) != nil {
		t.Fatal("success became failure")
	}
}
