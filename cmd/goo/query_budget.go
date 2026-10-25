package main

import (
	"context"
	"encoding/json"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"
)

// Local execution controls, not authentication. IDs are short-lived cancellation
// capabilities; they carry no identity or right to inspect another query's data.
type queryEntry struct {
	cancel  context.CancelFunc
	expires time.Time
}
type queryBudget struct {
	mu            sync.Mutex
	entries       map[string]queryEntry
	active, limit int
	next          http.Handler
}

var queryID = regexp.MustCompile(`^[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}$`)

func newQueryBudget(next http.Handler, limit int) http.Handler {
	return &queryBudget{entries: map[string]queryEntry{}, limit: limit, next: next}
}
func (b *queryBudget) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	controlled := r.Method == "POST" && (r.URL.Path == "/metric-discovery" || r.URL.Path == "/measurements" || r.URL.Path == "/explore" || r.URL.Path == "/explore/counted-values" || r.URL.Path == "/trace-search" || r.URL.Path == "/population/explain" || r.URL.Path == "/inventory" || r.URL.Path == "/discovery" || r.URL.Path == "/captures" || r.URL.Path == "/captures/live")
	cancelPath := r.Method == "POST" && strings.HasPrefix(r.URL.Path, "/queries/") && strings.HasSuffix(r.URL.Path, "/cancel")
	if !controlled && !cancelPath {
		b.next.ServeHTTP(w, r)
		return
	}
	fail := func(code int, message string) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "no-store")
		if code == 429 {
			w.Header().Set("Retry-After", "1")
		}
		w.WriteHeader(code)
		json.NewEncoder(w).Encode(map[string]string{"error": message})
	}
	id := r.Header.Get("X-Goo-Query-ID")
	if cancelPath {
		id = strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/queries/"), "/cancel")
	}
	if id != "" && !queryID.MatchString(id) || cancelPath && id == "" {
		fail(400, "invalid query ID")
		return
	}
	b.mu.Lock()
	now := time.Now()
	for key, e := range b.entries {
		if e.cancel == nil && now.After(e.expires) {
			delete(b.entries, key)
		}
	}
	if cancelPath {
		entry, ok := b.entries[id]
		if ok && entry.cancel != nil {
			entry.cancel()
		} else if !ok && len(b.entries) >= 256 {
			b.mu.Unlock()
			fail(429, "cancellation registry full; retry after expiry")
			return
		}
		// Retain a tombstone so cancellation also wins a request/admission race.
		if !ok || entry.cancel == nil {
			b.entries[id] = queryEntry{expires: now.Add(time.Minute)}
		}
		b.mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "no-store")
		w.WriteHeader(202)
		json.NewEncoder(w).Encode(map[string]string{"state": "cancellation-requested"})
		return
	}
	if _, exists := b.entries[id]; id != "" && exists {
		b.mu.Unlock()
		fail(409, "query ID already used or cancelled; use a new ID")
		return
	}
	if b.active >= b.limit || id != "" && len(b.entries) >= 256 {
		b.mu.Unlock()
		fail(429, "query capacity reached; retry after another query completes")
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 60*time.Second)
	b.active++
	if id != "" {
		b.entries[id] = queryEntry{cancel: cancel}
	}
	b.mu.Unlock()
	defer func() {
		cancel()
		b.mu.Lock()
		b.active--
		if id != "" {
			b.entries[id] = queryEntry{expires: time.Now().Add(time.Minute)}
		}
		b.mu.Unlock()
	}()
	b.next.ServeHTTP(w, r.WithContext(ctx))
}
