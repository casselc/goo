package exploration

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"path/filepath"
	"sync"
	"time"
)

// ResultSnapshots holds exact server-produced pages, never client record payloads.
// Tokens are process-local capabilities, not authenticated user identities.
type ResultSnapshots struct {
	mu                 sync.Mutex
	entries            map[string]resultSnapshot
	bytes              int
	now                func() time.Time
	publishExplanation func(CaptureStore, *ExplanationResult, string, string) (*SavedCapture, error)
}
type resultSnapshot struct {
	kind     string
	promoted map[string]explanationPromotion
	raw      []byte
	expires  time.Time
}

type explanationPromotion struct {
	CapturedAt string
	ID         string
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
	return s.freezeTyped(raw, "explore")
}

func (s *ResultSnapshots) freezeTyped(raw []byte, kind string) (string, error) {
	var random [32]byte
	if _, e := rand.Read(random[:]); e != nil {
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
	s.entries[id] = resultSnapshot{raw: raw, expires: now.Add(ResultSnapshotTTL), kind: kind}
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
	if v.kind != "" && v.kind != "explore" && v.kind != "explanation" {
		return nil, errors.New("unsupported result snapshot kind")
	}
	if v.kind == "explanation" {
		var result ExplanationResult
		if e := json.Unmarshal(v.raw, &result); e != nil {
			return nil, e
		}
		if e := ValidateExplanationResult(&result); e != nil {
			return nil, e
		}
		return &SavedCapture{Explanation: &result, CaptureReceipt: CaptureReceipt{Version: 1, ID: id, Title: "Observed population explanation", CapturedAt: result.ReadAt, Question: CaptureQuestion{Kind: "explanation", Explanation: detached(&result.Request)}}}, nil
	}
	var result Result
	if e := json.Unmarshal(v.raw, &result); e != nil {
		return nil, e
	}
	return &SavedCapture{CaptureReceipt: CaptureReceipt{Version: 1, ID: id, Title: "Live result page", CapturedAt: result.CapturedAt, Question: CaptureQuestion{Kind: "explore", Explore: &result.Question}}, Explore: &result}, nil
}

func (s *ResultSnapshots) FreezeExplanation(result *ExplanationResult) (string, error) {
	if e := ValidateExplanationResult(result); e != nil {
		return "", e
	}
	copy := *result
	copy.ResultToken = ""
	raw, e := json.Marshal(copy)
	if e != nil || len(raw) > ResultSnapshotMaxBytes {
		return "", errors.New("explanation exceeds snapshot bound")
	}
	return s.freezeTyped(raw, "explanation")
}

// Promotion is idempotent for one held observation/title/store while the token is
// alive. It is not a general durable retry key and never reloads telemetry.
func (s *ResultSnapshots) PromoteExplanation(token, title string, store CaptureStore) (*SavedCapture, error) {
	if e := ValidateCaptureTitle(title); e != nil {
		return nil, e
	}
	dir, e := filepath.Abs(store.Dir)
	if e != nil {
		return nil, e
	}
	key := explanationDigest(struct{ Dir, Title string }{dir, title})
	s.mu.Lock()
	defer s.mu.Unlock()
	s.prune(s.clock())
	entry, ok := s.entries[token]
	if !ok || entry.kind != "explanation" {
		return nil, errors.New("explanation expired or unavailable; run again explicitly")
	}
	promotion, reserved := entry.promoted[key]
	if promotion.ID != "" {
		return store.Load(promotion.ID)
	}
	if !reserved {
		if len(entry.promoted) >= 16 {
			return nil, errors.New("held explanation promotion capacity reached")
		}
		if entry.promoted == nil {
			entry.promoted = map[string]explanationPromotion{}
		}
		promotion = explanationPromotion{CapturedAt: s.clock().UTC().Format(time.RFC3339Nano)}
		entry.promoted[key] = promotion
		// Reserve custody before I/O. An error may follow a successful final link;
		// retrying must reproduce the same candidate, including its exact clock.
		s.entries[token] = entry
	}
	var result ExplanationResult
	if e := json.Unmarshal(entry.raw, &result); e != nil {
		return nil, e
	}
	publisher := s.publishExplanation
	if publisher == nil {
		publisher = func(store CaptureStore, result *ExplanationResult, title, at string) (*SavedCapture, error) {
			return store.saveExplanationAt(result, title, at)
		}
	}
	saved, e := publisher(store, &result, title, promotion.CapturedAt)
	if e != nil {
		return nil, e
	}
	if saved == nil {
		return nil, errors.New("explanation publication returned no receipt")
	}
	promotion.ID = saved.ID
	entry.promoted[key] = promotion
	s.entries[token] = entry
	return saved, nil
}
