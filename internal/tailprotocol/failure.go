// Package tailprotocol defines safe, closed diagnostics for live record streams.
package tailprotocol

import (
	"context"
	"errors"
	"net"

	"github.com/casselc/goo/internal/sourcefailure"
	"github.com/casselc/goo/internal/sourcehttp"
)

type Retry string

const (
	Automatic Retry = "automatic"
	Explicit  Retry = "explicit"
)

type Gap struct {
	Reason  string                 `json:"reason"`
	Retry   Retry                  `json:"retry"`
	Refusal *sourcefailure.Refusal `json:"refusal,omitempty"`
}

// Source classifies structured transport metadata only; error prose is never
// interpreted or transmitted. Stable refusals require an operator's new read.
func Source(err error) Gap {
	refusal := sourcefailure.Describe(err).Refusal
	gap := Gap{Reason: "source_unavailable", Retry: Explicit, Refusal: &refusal}
	if response, ok := errors.AsType[*sourcehttp.HTTPError](err); ok {
		switch {
		case response.Status == 401 || response.Status == 403 || response.Code == 497 || response.Code == 516:
			gap.Reason = "source_access"
			refusal.Detail = "The source refused access to the requested records."
		case response.Code == 60 || response.Code == 81:
			gap.Reason = "source_missing"
			refusal.Detail = "The requested record source is missing."
		case response.Code == 16 || response.Code == 47 || response.Code == 53:
			gap.Reason = "source_schema"
			refusal.Detail = "The record source does not match the supported exporter schema."
		}
		return gap
	}
	switch refusal.Kind {
	case sourcefailure.Timeout:
		gap.Reason, gap.Retry = "source_timeout", Automatic
	case sourcefailure.Cancelled:
		gap.Reason = "source_cancelled"
	case sourcefailure.Budget:
		gap.Reason = "source_budget"
	case sourcefailure.Malformed:
		gap.Reason = "source_malformed"
	default:
		// A structured NXDOMAIN is a configuration refusal, not a transient
		// disconnect. Timeout handling above deliberately takes precedence.
		if dns, ok := errors.AsType[*net.DNSError](err); ok && dns.IsNotFound {
			return gap
		}
		if _, ok := errors.AsType[net.Error](err); ok {
			gap.Retry = Automatic
		}
	}
	return gap
}

func Ended() Gap { return Gap{Reason: "connection_ended", Retry: Automatic} }
func IdentityBudget() Gap {
	refusal := sourcefailure.Describe(sourcefailure.Wrap(sourcefailure.Budget, errors.New("tail identity budget"))).Refusal
	return Gap{Reason: "identity_budget", Retry: Explicit, Refusal: &refusal}
}

// End distinguishes server rollover from caller cancellation, including when a
// source read was pending at the connection deadline.
func End(parent, connection context.Context) (Gap, bool) {
	if parent.Err() != nil {
		return Gap{}, false
	}
	if errors.Is(connection.Err(), context.DeadlineExceeded) {
		return Ended(), true
	}
	return Gap{}, false
}

type Response struct {
	Error  string `json:"error"`
	Reason string `json:"reason"`
	Retry  Retry  `json:"retry"`
}

func InvalidSelector() Response {
	return Response{"Invalid tail selector.", "invalid_selector", Explicit}
}
func Capacity() Response { return Response{"Live stream capacity reached.", "capacity", Automatic} }
func Unavailable() Response {
	return Response{"Live streaming unavailable.", "server_unavailable", Automatic}
}
