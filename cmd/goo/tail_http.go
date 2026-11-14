package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/internal/httpjson"
	"github.com/casselc/goo/internal/sourcefailure"
	"github.com/casselc/goo/internal/tailprotocol"
	"net/http"
	"time"
)

func registerTail(mux *http.ServeMux, source exploration.Source, bank *exploration.TailMembershipSnapshots) {
	slots := make(chan struct{}, 4)
	mux.HandleFunc("POST /tail", func(w http.ResponseWriter, r *http.Request) {
		fail := func(code int, failure tailprotocol.Response) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(code)
			json.NewEncoder(w).Encode(failure)
		}
		if r.URL.RawQuery != "" {
			fail(400, tailprotocol.InvalidSelector())
			return
		}
		var selector exploration.TailSelector
		if e := httpjson.Decode(w, r, &selector, 16384, "expected one JSON object"); e != nil {
			fail(400, tailprotocol.InvalidSelector())
			return
		}
		if e := selector.Query(time.Now().UnixMilli()).Validate(); e != nil {
			fail(400, tailprotocol.InvalidSelector())
			return
		}
		select {
		case slots <- struct{}{}:
			defer func() { <-slots }()
		default:
			fail(429, tailprotocol.Capacity())
			return
		}
		var connectionBytes [32]byte
		if _, err := rand.Read(connectionBytes[:]); err != nil {
			fail(500, tailprotocol.Unavailable())
			return
		}
		connection := hex.EncodeToString(connectionBytes[:])
		flush, ok := w.(http.Flusher)
		if !ok {
			fail(500, tailprotocol.Unavailable())
			return
		}
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("X-Accel-Buffering", "no")
		seq := 0
		send := func(kind string, value any) bool {
			b, e := json.Marshal(value)
			if e != nil {
				return false
			}
			seq++
			_ = http.NewResponseController(w).SetWriteDeadline(time.Now().Add(5 * time.Second))
			_, e = fmt.Fprintf(w, "id: %d\nevent: %s\ndata: %s\n\n", seq, kind, b)
			flush.Flush()
			return e == nil
		}
		if !send("reset", map[string]any{"connection": connection, "windowMs": exploration.TailWindowMS, "selector": selector, "continuity": "unknown", "replay": false}) {
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 60*time.Second)
		defer cancel()
		tracker := exploration.TailTracker{}
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()
		for {
			now := time.Now().UnixMilli()
			rows, limited, e := source.PollTail(ctx, selector, now)
			if e != nil {
				if r.Context().Err() != nil {
					return
				}
				if ended, ok := tailprotocol.End(r.Context(), ctx); ok {
					send("gap", ended)
				} else {
					send("gap", tailprotocol.Source(e))
				}
				return
			}
			observed, e := tracker.Observe(rows)
			if e != nil {
				if errors.Is(e, exploration.ErrTailIdentityBudget) {
					send("gap", tailprotocol.IdentityBudget())
				} else {
					send("gap", tailprotocol.Source(sourcefailure.Protocol(e)))
				}
				return
			}
			observedAt := time.Now().UTC()
			batch := map[string]any{"rows": observed, "limited": limited, "fromMs": now - exploration.TailWindowMS, "toMs": now, "observedAt": observedAt.Format(time.RFC3339Nano), "membershipState": "unavailable"}
			if len(observed) == 0 {
				batch["membershipState"] = "empty"
			} else if bank != nil {
				capability, err := bank.Freeze(exploration.TailMembershipBatch{Connection: connection, Selector: selector, FromMS: now - exploration.TailWindowMS, ToMS: now, ObservedAt: observedAt, Limited: limited, Rows: observed})
				if err == nil {
					batch["membershipState"] = "available"
					batch["membership"] = capability
				}
			}
			if !send("batch", batch) {
				return
			}
			select {
			case <-ctx.Done():
				if ended, ok := tailprotocol.End(r.Context(), ctx); ok {
					send("gap", ended)
				}
				return
			case <-ticker.C:
			}
		}
	})
}
