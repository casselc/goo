package exploration

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

const TailMembershipVersion = "goo/tail-membership/v1alpha1"
const TailMembershipTTL = 2 * time.Minute
const TailMembershipMaxBatches = 64
const TailMembershipMaxBytes = 8 << 20
const TailMembershipBatchBytes = 1 << 20

var ErrTailMembershipUnavailable = errors.New("tail membership expired or unavailable")
var ErrTailMembershipCapacity = errors.New("tail membership capacity reached")

// TailMembershipBatch is trusted server read provenance.
// Client row payloads must never be passed to Freeze.
type TailMembershipBatch struct {
	Connection string       `json:"connection"`
	Selector   TailSelector `json:"selector"`
	FromMS     int64        `json:"fromMs"`
	ToMS       int64        `json:"toMs"`
	ObservedAt time.Time    `json:"observedAt"`
	Limited    bool         `json:"limited"`
	Rows       []TailRow    `json:"rows"`
}
type TailMembershipCapability struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expiresAt"`
	Basis     string    `json:"basis"`
}
type TailMembershipRef struct {
	Token string `json:"token"`
	Key   string `json:"key"`
}
type TailSelectionMember struct {
	Ref        TailMembershipRef `json:"ref"`
	FromMS     int64             `json:"fromMs"`
	ToMS       int64             `json:"toMs"`
	ObservedAt time.Time         `json:"observedAt"`
	Basis      string            `json:"basis"`
	Limited    bool              `json:"limited"`
	Late       bool              `json:"late"`
	Record     json.RawMessage   `json:"record"`
}

// TailSelectionDraft proves bounded server observation membership, not custody,
// physical event identity, completeness, or durable retry after a process restart.
type TailSelectionDraft struct {
	Version    string                `json:"version"`
	Digest     string                `json:"digest"`
	Connection string                `json:"connection"`
	Selector   TailSelector          `json:"selector"`
	Members    []TailSelectionMember `json:"members"`
}
type tailMembershipEntry struct {
	raw        []byte
	capability TailMembershipCapability
	scope      string
	connection string
}
type TailMembershipSnapshots struct {
	mu      sync.Mutex
	now     func() time.Time
	entries map[string]tailMembershipEntry
	bytes   int
}

func NewTailMembershipSnapshots(now func() time.Time) *TailMembershipSnapshots {
	if now == nil {
		now = time.Now
	}
	return &TailMembershipSnapshots{now: now, entries: map[string]tailMembershipEntry{}}
}
func tailScope(selector TailSelector, at int64) (TailSelector, string, error) {
	if selector.Signal == "" {
		selector.Signal = "logs"
	}
	if err := selector.Query(at).Validate(); err != nil {
		return selector, "", err
	}
	b, err := json.Marshal(selector)
	if err != nil {
		return selector, "", err
	}
	var copied TailSelector
	if err = json.Unmarshal(b, &copied); err != nil {
		return selector, "", err
	}
	return copied, tailMembershipHash(copied), nil
}
func tailMembershipHash(value any) string {
	raw, _ := json.Marshal(value)
	h := sha256.Sum256(raw)
	return hex.EncodeToString(h[:])
}
func validTailConnection(id string) bool {
	if len(id) != 64 {
		return false
	}
	b, e := hex.DecodeString(id)
	return e == nil && len(b) == 32 && hex.EncodeToString(b) == id
}
func (s *TailMembershipSnapshots) prune(now time.Time) {
	for key, e := range s.entries {
		if !now.Before(e.capability.ExpiresAt) {
			delete(s.entries, key)
			s.bytes -= len(e.raw)
		}
	}
}
func canonicalTailMember(row TailRow, from, to int64) (TailRow, error) {
	var value map[string]any
	decoder := json.NewDecoder(bytes.NewReader(row.Record))
	decoder.UseNumber()
	if err := decoder.Decode(&value); err != nil {
		return row, err
	}
	if decoder.Decode(new(any)) != io.EOF {
		return row, errors.New("require one tail record")
	}
	stamp, ok := value["timestamp_ns"].(string)
	n, err := strconv.ParseInt(stamp, 10, 64)
	if !ok || err != nil || strconv.FormatInt(n, 10) != stamp || n < from*1000000 || n >= to*1000000 {
		return row, errors.New("tail member outside its read window")
	}
	canonical, err := json.Marshal(value)
	if err != nil {
		return row, err
	}
	h := sha256.Sum256(canonical)
	parts := strings.Split(row.Key, ":")
	ordinal := 0
	if len(parts) == 2 {
		ordinal, _ = strconv.Atoi(parts[1])
	}
	if len(parts) != 2 || parts[0] != hex.EncodeToString(h[:]) || ordinal < 1 || ordinal > TailReadLimit || parts[1] != strconv.Itoa(ordinal) {
		return row, errors.New("tail handle does not match canonical record occurrence")
	}
	row.Record = canonical
	return row, nil
}
func (s *TailMembershipSnapshots) Freeze(batch TailMembershipBatch) (TailMembershipCapability, error) {
	now := s.now()
	selector, scope, err := tailScope(batch.Selector, batch.ToMS)
	if err != nil {
		return TailMembershipCapability{}, err
	}
	if !validTailConnection(batch.Connection) || batch.FromMS != batch.ToMS-TailWindowMS || len(batch.Rows) < 1 || len(batch.Rows) > TailReadLimit || batch.ObservedAt.IsZero() || batch.ObservedAt.Before(time.UnixMilli(batch.ToMS)) || batch.ObservedAt.After(now) {
		return TailMembershipCapability{}, errors.New("invalid server tail batch provenance")
	}
	batch.Selector = selector
	batch.ObservedAt = batch.ObservedAt.UTC()
	// Expiry is anchored to the original read, not extended by delayed issuance.
	expiry := batch.ObservedAt.Add(TailMembershipTTL)
	if !now.Before(expiry) {
		return TailMembershipCapability{}, ErrTailMembershipUnavailable
	}
	inputBytes := 0
	for _, row := range batch.Rows {
		inputBytes += len(row.Record)
		if inputBytes > TailMembershipBatchBytes {
			return TailMembershipCapability{}, errors.New("tail membership batch exceeds byte budget")
		}
	}
	rows := make([]TailRow, len(batch.Rows))
	seen := map[string]bool{}
	for i, row := range batch.Rows {
		if seen[row.Key] {
			return TailMembershipCapability{}, errors.New("duplicate tail observation handle")
		}
		seen[row.Key] = true
		rows[i], err = canonicalTailMember(row, batch.FromMS, batch.ToMS)
		if err != nil {
			return TailMembershipCapability{}, err
		}
	}
	batch.Rows = rows
	raw, err := json.Marshal(batch)
	if err != nil || len(raw) > TailMembershipBatchBytes {
		return TailMembershipCapability{}, errors.New("tail membership batch exceeds byte budget")
	}
	var random [32]byte
	if _, err = rand.Read(random[:]); err != nil {
		return TailMembershipCapability{}, err
	}
	cap := TailMembershipCapability{Token: hex.EncodeToString(random[:]), ExpiresAt: expiry, Basis: tailMembershipHash(batch)}
	s.mu.Lock()
	defer s.mu.Unlock()
	now = s.now()
	s.prune(now)
	if !now.Before(expiry) {
		return TailMembershipCapability{}, ErrTailMembershipUnavailable
	}
	for _, e := range s.entries {
		if e.connection == batch.Connection && e.scope != scope {
			return TailMembershipCapability{}, errors.New("tail connection selector changed")
		}
	}
	if len(s.entries) >= TailMembershipMaxBatches || s.bytes+len(raw) > TailMembershipMaxBytes {
		return TailMembershipCapability{}, ErrTailMembershipCapacity
	}
	s.entries[cap.Token] = tailMembershipEntry{raw: raw, capability: cap, scope: scope, connection: batch.Connection}
	s.bytes += len(raw)
	return cap, nil
}
func (s *TailMembershipSnapshots) Resolve(connection string, selector TailSelector, refs []TailMembershipRef) (*TailSelectionDraft, error) {
	now := s.now()
	selector, scope, err := tailScope(selector, now.UnixMilli())
	if err != nil {
		return nil, err
	}
	if !validTailConnection(connection) || len(refs) < 1 || len(refs) > 20 {
		return nil, errors.New("select 1..20 observations from one tail connection")
	}
	refs = append([]TailMembershipRef{}, refs...)
	sort.Slice(refs, func(i, j int) bool {
		if refs[i].Token != refs[j].Token {
			return refs[i].Token < refs[j].Token
		}
		return refs[i].Key < refs[j].Key
	})
	s.mu.Lock()
	defer s.mu.Unlock()
	s.prune(s.now())
	out := &TailSelectionDraft{Version: TailMembershipVersion, Connection: connection, Selector: selector, Members: []TailSelectionMember{}}
	seen := map[string]bool{}
	for _, ref := range refs {
		if len(ref.Token) != 64 || len(ref.Key) > 72 {
			return nil, ErrTailMembershipUnavailable
		}
		e, ok := s.entries[ref.Token]
		if !ok || e.connection != connection || e.scope != scope {
			return nil, ErrTailMembershipUnavailable
		}
		if seen[ref.Key] {
			return nil, errors.New("same tail occurrence selected twice")
		}
		seen[ref.Key] = true
		var batch TailMembershipBatch
		if err = json.Unmarshal(e.raw, &batch); err != nil {
			return nil, err
		}
		found := false
		for _, row := range batch.Rows {
			if row.Key == ref.Key {
				found = true
				out.Members = append(out.Members, TailSelectionMember{Ref: ref, FromMS: batch.FromMS, ToMS: batch.ToMS, ObservedAt: batch.ObservedAt, Basis: e.capability.Basis, Limited: batch.Limited, Late: row.Late, Record: row.Record})
				break
			}
		}
		if !found {
			return nil, fmt.Errorf("tail occurrence is not a member of the referenced batch")
		}
	}
	raw, err := json.Marshal(out)
	if err != nil || len(raw) > 2<<20 {
		return nil, errors.New("tail selection exceeds byte budget")
	}
	out.Digest = tailMembershipHash(out)
	return out, nil
}

// CanonicalRecords owns a normalized representation of the observation records.
// JSON whitespace/key order/string escapes are transport details; numeric tokens
// remain exact. This does not issue membership or change provenance.
func (d TailSelectionDraft) CanonicalRecords() (TailSelectionDraft, error) {
	if len(d.Members) < 1 || len(d.Members) > 20 {
		return d, errors.New("invalid tail member count")
	}
	d.Members = append([]TailSelectionMember(nil), d.Members...)
	inputBytes := 0
	for _, m := range d.Members {
		inputBytes += len(m.Record)
		if inputBytes > 2<<20 {
			return d, errors.New("tail selection exceeds byte budget")
		}
	}
	for i, m := range d.Members {
		if m.FromMS < 0 || m.ToMS > 9223372036854 || m.FromMS != m.ToMS-TailWindowMS {
			return d, errors.New("invalid tail member read window")
		}

		row, e := canonicalTailMember(TailRow{Key: m.Ref.Key, Record: m.Record, Late: m.Late}, m.FromMS, m.ToMS)
		if e != nil {
			return d, e
		}
		d.Members[i].Record = row.Record
	}
	return d, nil
}

// Validate checks the durable shape of a resolved draft without requiring live
// capabilities. It cannot establish membership; only Resolve may issue a draft.
func (d TailSelectionDraft) Validate() error {
	var err error
	d, err = d.CanonicalRecords()
	if err != nil {
		return err
	}
	digest := d.Digest
	d.Digest = ""
	raw, err := json.Marshal(d)
	if err != nil || len(raw) > 2<<20 || digest != tailMembershipHash(d) || d.Version != TailMembershipVersion || !validTailConnection(d.Connection) || len(d.Members) < 1 || len(d.Members) > 20 {
		return errors.New("invalid tail selection draft")
	}
	selector, _, err := tailScope(d.Selector, d.Members[0].ToMS)
	if err != nil || tailMembershipHash(selector) != tailMembershipHash(d.Selector) {
		return errors.New("invalid tail selector")
	}
	seen := map[string]bool{}
	previous := ""
	for _, m := range d.Members {
		order := m.Ref.Token + "/" + m.Ref.Key
		if !validTailConnection(m.Ref.Token) || !validTailConnection(m.Basis) || order <= previous || seen[m.Ref.Key] || m.FromMS < 0 || m.ToMS > 9223372036854 || m.FromMS != m.ToMS-TailWindowMS || m.ObservedAt.IsZero() || m.ObservedAt.Before(time.UnixMilli(m.ToMS)) {
			return errors.New("invalid tail member provenance")
		}

		seen[m.Ref.Key] = true
		previous = order
	}
	return nil
}
