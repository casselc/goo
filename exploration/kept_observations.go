package exploration

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

const KeepTailVersion = "goo/keep-tail/v1"
const KeptObservationsVersion = "goo/kept-observations/v1"
const KeptObservationsMaxBytes = 3 << 20

var ErrKeepConflict = errors.New("keep request ID belongs to another selection")
var keptPublishSlots = make(chan struct{}, 4)

// KeepTailIntent contains capabilities, never browser-provided record contents.
type KeepTailIntent struct {
	Version    string              `json:"version"`
	RequestID  string              `json:"requestID"`
	Connection string              `json:"connection"`
	Selector   TailSelector        `json:"selector"`
	Refs       []TailMembershipRef `json:"refs"`
}

func (in KeepTailIntent) normalized() (KeepTailIntent, error) {
	if in.Version != KeepTailVersion || !validTailConnection(in.RequestID) || !validTailConnection(in.Connection) || len(in.Refs) < 1 || len(in.Refs) > 20 {
		return in, errors.New("require a keep request ID and 1..20 observations from one connection")
	}
	selector, _, err := tailScope(in.Selector, 60000)
	if err != nil {
		return in, err
	}
	in.Selector = selector
	in.Refs = append([]TailMembershipRef(nil), in.Refs...)
	sort.Slice(in.Refs, func(i, j int) bool {
		if in.Refs[i].Token != in.Refs[j].Token {
			return in.Refs[i].Token < in.Refs[j].Token
		}
		return in.Refs[i].Key < in.Refs[j].Key
	})
	seen := map[string]bool{}
	for _, ref := range in.Refs {
		parts := strings.Split(ref.Key, ":")
		if !validTailConnection(ref.Token) || len(ref.Key) > 72 || len(parts) != 2 || !validTailConnection(parts[0]) || seen[ref.Key] {
			return in, errors.New("invalid or duplicate observation reference")
		}
		ordinal, e := strconv.Atoi(parts[1])
		if e != nil || ordinal < 1 || ordinal > TailReadLimit || strconv.Itoa(ordinal) != parts[1] {
			return in, errors.New("invalid observation occurrence")
		}
		seen[ref.Key] = true
	}
	return in, nil
}
func (in KeepTailIntent) Validate() error { _, err := in.normalized(); return err }

// KeptObservation preserves a server-observed occurrence, not a complete population
// or original OTLP/exporter bytes. Live capability tokens are never returned/stored.
type KeptObservation struct {
	Key        string          `json:"key"`
	FromMS     int64           `json:"fromMs"`
	ToMS       int64           `json:"toMs"`
	ObservedAt time.Time       `json:"observedAt"`
	Basis      string          `json:"basis"`
	Limited    bool            `json:"limited"`
	Late       bool            `json:"late"`
	Record     json.RawMessage `json:"record"`
}
type KeptObservations struct {
	Version      string            `json:"version"`
	ID           string            `json:"id"`
	IntentDigest string            `json:"intentDigest"`
	KeptAt       time.Time         `json:"keptAt"`
	Connection   string            `json:"connection"`
	Selector     TailSelector      `json:"selector"`
	Members      []KeptObservation `json:"members"`
	Digest       string            `json:"digest"`
}

func (v KeptObservations) Validate() error {
	if v.Version != KeptObservationsVersion || !validTailConnection(v.ID) || !validTailConnection(v.IntentDigest) || !validTailConnection(v.Connection) || v.KeptAt.IsZero() || len(v.Members) < 1 || len(v.Members) > 20 {
		return errors.New("invalid kept observation receipt")
	}
	if _, _, err := tailScope(v.Selector, 60000); err != nil {
		return err
	}
	seen := map[string]bool{}
	for _, m := range v.Members {
		if m.FromMS < 0 || m.ToMS > 4102444800000 || m.ToMS-m.FromMS != TailWindowMS || m.ObservedAt.IsZero() || m.ObservedAt.UnixMilli() < m.ToMS || !validTailConnection(m.Basis) || seen[m.Key] {
			return errors.New("invalid kept observation provenance")
		}
		seen[m.Key] = true
		if _, err := canonicalTailMember(TailRow{Key: m.Key, Record: m.Record, Late: m.Late}, m.FromMS, m.ToMS); err != nil {
			return err
		}
		q := v.Selector.Query(m.ToMS)
		selection, err := q.EnvironmentSelection()
		if err != nil {
			return err
		}
		if err = validateSelectedRecord(selection, q.Signal, q.FromMS, q.ToMS, m.Record); err != nil {
			return err
		}
	}
	digest := v.Digest
	v.Digest = ""
	if !validTailConnection(digest) || digest != tailMembershipHash(v) {
		return errors.New("kept observation digest mismatch")
	}
	return nil
}

// One immutable envelope per request ID makes a successful lost-ack retry
// recoverable after proof expiry or process restart. The content digest is separate.
type KeptObservationStore struct {
	Dir          string
	now          func() time.Time
	afterPublish func() error
}

type keptResource interface {
	Validate() error
	keepIdentity() (string, string)
}

func (v *KeptObservations) keepIdentity() (string, string) { return v.ID, v.IntentDigest }
func (s KeptObservationStore) Load(id string) (*KeptObservations, error) {
	v, e := s.load(id)
	if e != nil {
		return nil, e
	}
	tail, ok := v.(*KeptObservations)
	if !ok {
		return nil, ErrKeepConflict
	}
	return tail, nil
}

// Open verifies either immutable resource version before returning its wire value.
func (s KeptObservationStore) Open(id string) (json.RawMessage, error) {
	v, e := s.load(id)
	if e != nil {
		return nil, e
	}
	return json.Marshal(v)
}
func (s KeptObservationStore) load(id string) (keptResource, error) {
	if !validTailConnection(id) {
		return nil, errors.New("invalid kept observation ID")
	}
	path := filepath.Join(s.Dir, id+".json")
	info, err := os.Lstat(path)
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() || info.Size() > KeptObservationsMaxBytes {
		return nil, errors.New("invalid kept observation file")
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	raw, err := io.ReadAll(io.LimitReader(f, KeptObservationsMaxBytes+1))
	if err != nil {
		return nil, err
	}
	if len(raw) > KeptObservationsMaxBytes {
		return nil, errors.New("kept observations exceed byte bound")
	}
	var header struct {
		Version string `json:"version"`
	}
	if err = json.Unmarshal(raw, &header); err != nil {
		return nil, err
	}
	var v keptResource
	switch header.Version {
	case KeptObservationsVersion:
		v = &KeptObservations{}
	case KeptRecordsVersion:
		v = &KeptRecords{}
	default:
		return nil, errors.New("unsupported kept resource version")
	}
	d := json.NewDecoder(bytes.NewReader(raw))
	d.DisallowUnknownFields()
	if err = d.Decode(v); err != nil {
		return nil, err
	}
	if d.Decode(new(any)) != io.EOF {
		return nil, errors.New("require one kept observation envelope")
	}
	if storedID, _ := v.keepIdentity(); storedID != id {
		return nil, errors.New("kept observation ID mismatch")
	}
	if err = v.Validate(); err != nil {
		return nil, err
	}
	return v, nil
}
func (s KeptObservationStore) Keep(ctx context.Context, bank *TailMembershipSnapshots, input KeepTailIntent) (*KeptObservations, error) {
	in, err := input.normalized()
	if err != nil {
		return nil, err
	}
	result, err := s.publish(ctx, in.RequestID, tailMembershipHash(in), func() (keptResource, error) {
		if bank == nil {
			return nil, ErrTailMembershipUnavailable
		}
		draft, err := bank.Resolve(in.Connection, in.Selector, in.Refs)
		if err != nil {
			return nil, fmt.Errorf("%w: %v", ErrTailMembershipUnavailable, err)
		}
		now := time.Now()
		if s.now != nil {
			now = s.now()
		}
		v := &KeptObservations{Version: KeptObservationsVersion, ID: in.RequestID, IntentDigest: tailMembershipHash(in), KeptAt: now.UTC(), Connection: draft.Connection, Selector: draft.Selector, Members: make([]KeptObservation, 0, len(draft.Members))}
		for _, m := range draft.Members {
			v.Members = append(v.Members, KeptObservation{Key: m.Ref.Key, FromMS: m.FromMS, ToMS: m.ToMS, ObservedAt: m.ObservedAt, Basis: m.Basis, Limited: m.Limited, Late: m.Late, Record: append(json.RawMessage(nil), m.Record...)})
		}
		v.Digest = tailMembershipHash(v)
		if err = v.Validate(); err != nil {
			return nil, err
		}
		return v, nil
	})
	if err != nil {
		return nil, err
	}
	v, ok := result.(*KeptObservations)
	if !ok {
		return nil, ErrKeepConflict
	}
	return v, nil
}
func (s KeptObservationStore) publish(ctx context.Context, id, digest string, resolve func() (keptResource, error)) (keptResource, error) {
	winner := func() (keptResource, error) {
		v, e := s.load(id)
		if e == nil {
			_, intent := v.keepIdentity()
			if intent != digest {
				return nil, ErrKeepConflict
			}
		}
		return v, e
	}
	if v, e := winner(); e == nil {
		return v, nil
	} else if !errors.Is(e, os.ErrNotExist) {
		return nil, e
	}
	select {
	case keptPublishSlots <- struct{}{}:
		defer func() { <-keptPublishSlots }()
	case <-ctx.Done():
		return nil, ctx.Err()
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	// No source/bank locks are held during filesystem publication.
	v, err := resolve()
	if err != nil {
		if old, e := winner(); e == nil {
			return old, nil
		} else if !errors.Is(e, os.ErrNotExist) {
			return nil, e
		}
		return nil, err
	}
	storedID, storedIntent := v.keepIdentity()
	if storedID != id || storedIntent != digest {
		return nil, errors.New("kept candidate identity mismatch")
	}
	if err = v.Validate(); err != nil {
		return nil, err
	}
	raw, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	if len(raw) > KeptObservationsMaxBytes {
		return nil, errors.New("kept observations exceed byte bound")
	}
	if err = ctx.Err(); err != nil {
		return nil, err
	}
	if err = os.MkdirAll(s.Dir, 0700); err != nil {
		return nil, err
	}
	tmp, err := os.CreateTemp(s.Dir, ".keep-*")
	if err != nil {
		return nil, err
	}
	defer os.Remove(tmp.Name())
	defer tmp.Close()
	if _, err = tmp.Write(raw); err != nil {
		return nil, err
	}
	if err = tmp.Sync(); err != nil {
		return nil, err
	}
	if err = tmp.Close(); err != nil {
		return nil, err
	}
	if err = os.Link(tmp.Name(), filepath.Join(s.Dir, id+".json")); err != nil {
		if errors.Is(err, os.ErrExist) {
			return winner()
		}
		return nil, err
	}
	dir, err := os.Open(s.Dir)
	if err != nil {
		return nil, err
	}
	err = dir.Sync()
	dir.Close()
	if err != nil {
		return nil, fmt.Errorf("keep may have committed: %w", err)
	}
	if s.afterPublish != nil {
		if err = s.afterPublish(); err != nil {
			return nil, err
		}
	}
	return v, nil
}
