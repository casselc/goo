package workbench

import (
	"errors"
	"time"
)

var ErrCaptureReviewIntentConflict = errors.New("review requestID is already bound to another kind or input")

// A durable request intent fixes a server clock before publication. Repeating
// the same intent and input recovers the same immutable revision after lost ACKs.
// It grants neither latest-head nor multi-artifact transaction semantics.
type captureReviewIntent struct {
	Version    int    `json:"version"`
	RequestID  string `json:"requestID"`
	Kind       string `json:"kind"`
	InputHash  string `json:"inputHash"`
	RecordedAt string `json:"recordedAt"`
	Origin     string `json:"origin"`
	Seal       string `json:"seal"`
}

func (s Store) captureReviewIntentClock(requestID, kind string, input any) (string, error) {
	if !ID.MatchString(requestID) {
		return "", errors.New("review requestID requires 64 lowercase hexadecimal characters")
	}
	hash, err := digest(input)
	if err != nil {
		return "", err
	}
	intent := captureReviewIntent{Version: 1, RequestID: requestID, Kind: kind, InputHash: hash, RecordedAt: time.Now().UTC().Format(time.RFC3339Nano), Origin: captureReviewOrigin}
	intent.Seal, err = digest(intent)
	if err != nil {
		return "", err
	}
	if err = s.publish("capture-review-intents", requestID, intent); err != nil {
		return "", err
	}
	var winner captureReviewIntent
	if err = s.readReview("capture-review-intents", requestID, &winner); err != nil {
		return "", err
	}
	seal := winner.Seal
	winner.Seal = ""
	wantSeal, sealErr := digest(winner)
	if sealErr != nil || !ID.MatchString(seal) || seal != wantSeal {
		return "", errors.New("corrupt review request intent seal")
	}
	if winner.Version != 1 || winner.RequestID != requestID || !ID.MatchString(winner.InputHash) || !reviewClock(winner.RecordedAt, winner.Origin) {
		return "", errors.New("corrupt review request intent")
	}
	if winner.Kind != kind || winner.InputHash != hash {
		return "", ErrCaptureReviewIntentConflict
	}
	return winner.RecordedAt, nil
}
func (s Store) CreateCaptureReviewTaskOnce(requestID string, in CaptureReviewTaskInput) (*CaptureReviewTask, error) {
	if err := in.Validate(); err != nil {
		return nil, err
	}
	at, err := s.captureReviewIntentClock(requestID, "task", in)
	if err != nil {
		return nil, err
	}
	return s.createCaptureReviewTaskAt(in, at)
}
func (s Store) CreateCaptureReviewDecisionOnce(requestID string, in CaptureReviewDecisionInput, resolve CaptureReviewResolver) (*CaptureReviewDecision, error) {
	if err := in.Validate(); err != nil {
		return nil, err
	}
	at, err := s.captureReviewIntentClock(requestID, "decision", in)
	if err != nil {
		return nil, err
	}
	return s.createCaptureReviewDecisionAt(in, resolve, at)
}
func (s Store) CreateCaptureReviewManifestOnce(requestID string, in CaptureReviewManifestInput, resolve CaptureReviewResolver) (*CaptureReviewManifest, error) {
	if err := in.Validate(); err != nil {
		return nil, err
	}
	at, err := s.captureReviewIntentClock(requestID, "manifest", in)
	if err != nil {
		return nil, err
	}
	return s.createCaptureReviewManifestAt(in, resolve, at)
}
