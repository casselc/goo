package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/casselc/goo/exploration"
)

// Bounded, expiring, process-local snapshot capabilities. Restart loses drafts,
// never published captures. Repeated freezes of a token return the same receipt.
func registerLiveComparison(mux *http.ServeMux, client *http.Client, store exploration.CaptureStore) {
	type draft struct {
		snapshot *exploration.SavedCapture
		expires  time.Time
		size     int
		frozen   *exploration.SavedCapture
	}
	var mu sync.Mutex
	drafts := map[string]*draft{}
	reply := func(w http.ResponseWriter, code int, v any) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "no-store")
		w.WriteHeader(code)
		json.NewEncoder(w).Encode(v)
	}
	fail := func(w http.ResponseWriter, code int, e error) { reply(w, code, map[string]string{"error": e.Error()}) }
	mux.HandleFunc("POST /captures/live", func(w http.ResponseWriter, r *http.Request) {
		var in struct {
			Baseline string `json:"baseline"`
			Freeze   string `json:"freeze"`
			Replace  string `json:"replace"`
		}
		decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 16384))
		decoder.DisallowUnknownFields()
		if e := decoder.Decode(&in); e != nil {
			fail(w, 400, e)
			return
		}
		if decoder.Decode(new(any)) != io.EOF || (in.Baseline == "") == (in.Freeze == "") {
			fail(w, 400, fmt.Errorf("provide exactly one baseline or freeze capability"))
			return
		}
		if in.Freeze != "" {
			mu.Lock()
			defer mu.Unlock()
			d := drafts[in.Freeze]
			if d == nil || time.Now().After(d.expires) {
				fail(w, 410, fmt.Errorf("displayed snapshot expired; refresh explicitly"))
				return
			}
			if d.frozen == nil {
				var e error
				d.frozen, e = store.FreezeSnapshot(d.snapshot)
				if e != nil {
					fail(w, 500, e)
					return
				}
			}
			page, e := d.frozen.Page("")
			if e != nil {
				fail(w, 500, e)
				return
			}
			reply(w, 201, page)
			return
		}
		baseline, e := store.Load(in.Baseline)
		if e != nil {
			fail(w, 404, e)
			return
		}
		endpoint := os.Getenv("CLICKHOUSE_URL")
		if endpoint == "" {
			endpoint = "http://localhost:8123"
		}
		source := exploration.Source{Endpoint: endpoint, User: os.Getenv("CLICKHOUSE_USER"), Password: os.Getenv("CLICKHOUSE_PASSWORD"), Client: client}
		ctx, cancel := context.WithTimeout(r.Context(), 55*time.Second)
		defer cancel()
		var snapshot *exploration.SavedCapture
		var comparison *exploration.EvidenceComparison
		var native *exploration.LiveCounterInfo
		if baseline.Measurement != nil {
			snapshot, comparison, native, e = source.LiveCounterComparison(ctx, baseline, time.Now())
		} else {
			snapshot, comparison, e = source.LiveRecordComparison(ctx, baseline, time.Now())
		}
		if e != nil {
			fail(w, 502, e)
			return
		}
		raw, e := json.Marshal(snapshot)
		if e != nil {
			fail(w, 500, e)
			return
		}
		if len(raw) > 4<<20 {
			fail(w, 413, fmt.Errorf("snapshot exceeds 4 MiB budget"))
			return
		}
		var token [32]byte
		if _, e = rand.Read(token[:]); e != nil {
			fail(w, 500, e)
			return
		}
		key := hex.EncodeToString(token[:])
		expiry := time.Now().Add(2 * time.Minute)
		mu.Lock()
		defer mu.Unlock()
		total := len(raw)
		for k, d := range drafts {
			if time.Now().After(d.expires) {
				delete(drafts, k)
			} else {
				if k != in.Replace {
					total += d.size
				}
			}
		}
		entries := len(drafts)
		if drafts[in.Replace] != nil {
			entries--
		}
		if entries >= 16 || total > 16<<20 {
			fail(w, 429, fmt.Errorf("live snapshot budget full; pause and retry after expiry"))
			return
		}
		delete(drafts, in.Replace)
		drafts[key] = &draft{snapshot: snapshot, expires: expiry, size: len(raw)}
		reply(w, 200, map[string]any{"token": key, "expiresAt": expiry.UTC().Format(time.RFC3339Nano), "comparison": comparison, "basis": liveBasis(snapshot), "baselineBasis": liveBasis(baseline), "native": native})
	})
}

func liveBasis(c *exploration.SavedCapture) string {
	if c.Measurement != nil {
		return c.Measurement.Basis
	}
	if c.Explore != nil {
		return c.Explore.Basis
	}
	return ""
}
