package workbench

import (
	"errors"
	"fmt"
	"github.com/casselc/goo/identity"
	"sort"
	"strings"
	"time"
)

// MeaningTarget identifies metadata subjects, never implicit telemetry membership.
type MeaningTarget struct {
	Kind        string              `json:"kind"`
	ID          string              `json:"id,omitempty"`
	Source      string              `json:"source,omitempty"`
	Environment string              `json:"environment,omitempty"`
	Service     string              `json:"service,omitempty"`
	Namespace   *identity.Namespace `json:"namespace,omitempty"`
}

func (t MeaningTarget) Validate() error {
	switch t.Kind {
	case "service":
		if t.ID != "" || t.Source != "clickstack" || !text(t.Environment, 128, true, false) || !text(t.Service, 1024, false, false) || t.Namespace == nil || t.Namespace.Validate() != nil {
			return errors.New("service target requires clickstack source, environment, literal service and exact namespace")
		}
	case "view", "document", "package", "evidence":
		if !ID.MatchString(t.ID) || t.Source != "" || t.Environment != "" || t.Service != "" || t.Namespace != nil {
			return errors.New("artifact target requires only kind and exact revision ID")
		}
	default:
		return errors.New("unsupported annotation target")
	}
	return nil
}

type MeaningWindow struct {
	FromMS int64 `json:"fromMs"`
	ToMS   int64 `json:"toMs"`
}
type MeaningInput struct {
	Target  MeaningTarget  `json:"target"`
	Parent  string         `json:"parent,omitempty"`
	Tags    []string       `json:"tags"`
	Note    string         `json:"note"`
	State   string         `json:"state"`
	Applies *MeaningWindow `json:"applies,omitempty"`
}
type MeaningRevision struct {
	MeaningInput
	Version    int    `json:"version"`
	ID         string `json:"id"`
	RecordedAt string `json:"recordedAt"`
	Origin     string `json:"origin"`
}

func validTag(t string) bool {
	if len(t) < 1 || len(t) > 64 {
		return false
	}
	for i, c := range t {
		if c >= 'a' && c <= 'z' || c >= '0' && c <= '9' {
			continue
		}
		if i == 0 || c != '.' && c != '_' && c != ':' && c != '/' && c != '-' {
			return false
		}
	}
	return true
}
func (in MeaningInput) Validate() error {
	if e := in.Target.Validate(); e != nil {
		return e
	}
	if in.Parent != "" && !ID.MatchString(in.Parent) {
		return errors.New("invalid parent revision")
	}
	if len(in.Tags) > 16 || !text(in.Note, 4000, false, true) {
		return errors.New("at most 16 tags and 4000 UTF-8 note bytes")
	}
	seen := map[string]bool{}
	for _, tag := range in.Tags {
		if !validTag(tag) || seen[tag] {
			return errors.New("require unique lowercase tags of 1..64 letters, digits or ._:/-; start with letter/digit")
		}
		seen[tag] = true
	}
	if in.State != "active" && in.State != "withdrawn" || in.State == "active" && len(in.Tags) == 0 && strings.TrimSpace(in.Note) == "" || in.State == "withdrawn" && (in.Parent == "" || len(in.Tags) > 0 || in.Note != "") {
		return errors.New("active revision needs tags or note; withdrawal needs parent and empty content")
	}
	if in.Target.Kind == "service" {
		if in.Applies == nil || in.Applies.FromMS < 0 || in.Applies.ToMS <= in.Applies.FromMS || in.Applies.ToMS > 8640000000000000 {
			return errors.New("service annotation needs an explicit half-open applicability window")
		}
	} else if in.Applies != nil {
		return errors.New("artifact revision has no service applicability window")
	}
	return nil
}
func (s Store) CreateMeaning(in MeaningInput, checkPackage func(string) error) (*MeaningRevision, error) {
	if e := in.Validate(); e != nil {
		return nil, e
	}
	switch in.Target.Kind {
	case "evidence":
		return nil, errors.New("retained evidence targets require the optional evidence store")
	case "view":
		if _, e := s.LoadExplorationView(in.Target.ID); e != nil {
			return nil, fmt.Errorf("view target: %w", e)
		}
	case "document":
		if _, e := s.LoadDocument(in.Target.ID); e != nil {
			return nil, fmt.Errorf("document target: %w", e)
		}
	case "package":
		if checkPackage == nil {
			return nil, errors.New("package reference resolver required")
		}
		if e := checkPackage(in.Target.ID); e != nil {
			return nil, e
		}
	}
	if in.Parent != "" {
		p, e := s.LoadMeaning(in.Parent)
		if e != nil {
			return nil, e
		}
		a, _ := digest(in.Target)
		b, _ := digest(p.Target)
		if a != b {
			return nil, errors.New("parent belongs to a different target")
		}
	}
	r, e := PrepareMeaning(in)
	if e != nil {
		return nil, e
	}
	id := r.ID
	if e = s.publish("meanings", id, r); e != nil {
		return nil, e
	}
	return r, nil
}

// PrepareMeaning constructs a server-timed revision after target and parent resolution by its owning store.
func PrepareMeaning(in MeaningInput) (*MeaningRevision, error) {
	if e := in.Validate(); e != nil {
		return nil, e
	}
	in.Target.Namespace = in.Target.Namespace.Clone()
	in.Tags = append([]string{}, in.Tags...)
	sort.Strings(in.Tags)
	if in.Applies != nil {
		v := *in.Applies
		in.Applies = &v
	}
	r := &MeaningRevision{MeaningInput: in, Version: 1, RecordedAt: time.Now().UTC().Format(time.RFC3339Nano), Origin: "local-user-unverified"}
	id, e := digest(r)
	if e != nil {
		return nil, e
	}
	r.ID = id
	return r, nil
}
func (s Store) LoadMeaning(id string) (*MeaningRevision, error) {
	var r MeaningRevision
	if e := s.read("meanings", id, &r); e != nil {
		return nil, e
	}
	if r.ID != id {
		return nil, errors.New("revision ID mismatch")
	}
	if e := ValidateMeaningRevision(r); e != nil {
		return nil, e
	}
	return &r, nil
}

// ValidateMeaningRevision verifies immutable content independently of the storage adapter.
func ValidateMeaningRevision(r MeaningRevision) error {
	id := r.ID
	r.ID = ""
	want, e := digest(r)
	if e != nil || !ID.MatchString(id) || want != id || r.Version != 1 || !validTime(r.RecordedAt) || r.Origin != "local-user-unverified" || r.MeaningInput.Validate() != nil || !sort.StringsAreSorted(r.Tags) {
		return errors.New("invalid annotation revision integrity")
	}
	return nil
}

type MeaningItem struct {
	Revision *MeaningRevision `json:"revision"`
	Head     bool             `json:"head"`
}
type MeaningPage struct {
	Items      []MeaningItem `json:"items"`
	NextCursor string        `json:"nextCursor,omitempty"`
}

// Heads derive from parent edges over the bounded catalog, never wall-clock order.
func (s Store) Meanings(target *MeaningTarget, tag, cursor string, history bool) (*MeaningPage, error) {
	if target != nil {
		if e := target.Validate(); e != nil {
			return nil, e
		}
	}
	if tag != "" && !validTag(tag) || cursor != "" && !ID.MatchString(cursor) {
		return nil, errors.New("invalid tag/cursor")
	}
	ids, e := s.ids("meanings")
	if e != nil {
		return nil, e
	}
	rows := []*MeaningRevision{}
	parents := map[string]bool{}
	for _, id := range ids {
		r, e := s.LoadMeaning(id)
		if e != nil {
			return nil, e
		}
		rows = append(rows, r)
		if r.Parent != "" {
			parents[r.Parent] = true
		}
	}
	result := &MeaningPage{Items: []MeaningItem{}}
	matched := []MeaningItem{}
	key := ""
	if target != nil {
		key, _ = digest(target)
	}
	for _, r := range rows {
		if !history && (parents[r.ID] || r.State != "active") {
			continue
		}
		if target != nil {
			k, _ := digest(r.Target)
			if k != key {
				continue
			}
		}
		if tag != "" {
			found := false
			for _, t := range r.Tags {
				found = found || t == tag
			}
			if !found {
				continue
			}
		}
		matched = append(matched, MeaningItem{r, !parents[r.ID]})
	}
	start := 0
	if cursor != "" {
		found := false
		for i, r := range matched {
			if r.Revision.ID == cursor {
				start = i + 1
				found = true
				break
			}
		}
		if !found {
			return nil, errors.New("cursor no longer matches this catalog; restart first page")
		}
	}
	end := start + 20
	if end > len(matched) {
		end = len(matched)
	}
	result.Items = matched[start:end]
	if end < len(matched) {
		result.NextCursor = matched[end-1].Revision.ID
	}
	return result, nil
}
