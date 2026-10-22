package exploration

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"sync"
	"time"
)

// ResultSnapshots holds exact server-produced pages, never client record payloads.
// Tokens are process-local capabilities, not authenticated user identities.
type ResultSnapshots struct {
	mu      sync.Mutex
	entries map[string]resultSnapshot
	bytes   int
	now     func() time.Time
}
type resultSnapshot struct {
	raw     []byte
	expires time.Time
}

const ResultSnapshotTTL = 2 * time.Minute
const ResultSnapshotMaxBytes = 2 << 20
const resultSnapshotBudget = 64 << 20

func (s *ResultSnapshots) clock() time.Time {
	if s.now != nil {
		return s.now()
	}
	return time.Now()
}
func (s *ResultSnapshots) prune(now time.Time) {
	for id, v := range s.entries {
		if !now.Before(v.expires) {
			delete(s.entries, id)
			s.bytes -= len(v.raw)
		}
	}
}
func (s *ResultSnapshots) Freeze(result *Result) (string, error) {
	if result == nil || result.Question.Validate() != nil {
		return "", errors.New("invalid result page")
	}
	// Serialize before publishing a token: later caller mutations cannot change it.
	copy := *result
	copy.ResultToken = ""
	raw, e := json.Marshal(copy)
	if e != nil || len(raw) > ResultSnapshotMaxBytes {
		return "", errors.New("result page exceeds snapshot bound")
	}
	var random [32]byte
	if _, e = rand.Read(random[:]); e != nil {
		return "", e
	}
	id := hex.EncodeToString(random[:])
	s.mu.Lock()
	defer s.mu.Unlock()
	now := s.clock()
	s.prune(now)
	if len(s.entries) >= 64 || s.bytes+len(raw) > resultSnapshotBudget {
		return "", errors.New("result snapshot capacity reached")
	}
	if s.entries == nil {
		s.entries = map[string]resultSnapshot{}
	}
	s.entries[id] = resultSnapshot{raw, now.Add(ResultSnapshotTTL)}
	s.bytes += len(raw)
	return id, nil
}
func (s *ResultSnapshots) Load(id string) (*SavedCapture, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.prune(s.clock())
	v, ok := s.entries[id]
	if !ok {
		return nil, errors.New("result page expired or unavailable; run the question and select again")
	}
	var result Result
	if e := json.Unmarshal(v.raw, &result); e != nil {
		return nil, e
	}
	return &SavedCapture{CaptureReceipt: CaptureReceipt{Version: 1, ID: id, Title: "Live result page", CapturedAt: result.CapturedAt, Question: CaptureQuestion{Kind: "explore", Explore: &result.Question}}, Explore: &result}, nil
}
