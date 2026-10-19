package workbench

import (
	"errors"
	"strings"
	"time"
)

type Card struct {
	Step      *Step  `json:"step,omitempty"`
	Kind      string `json:"kind"`
	Role      string `json:"role,omitempty"`
	Text      string `json:"text,omitempty"`
	Reference string `json:"reference,omitempty"`
	Caption   string `json:"caption,omitempty"`
}

func (c Card) Validate() error {
	if c.Kind != "question" && c.Step != nil {
		return errors.New("only question cards carry a step")
	}
	switch c.Kind {
	case "question":
		if c.Step == nil || c.Reference != "" || c.Role != "" || c.Text != "" || !text(c.Caption, 160, false, false) {
			return errors.New("question card requires a step and optional caption")
		}
		return c.Step.Validate()
	case "note":
		if c.Reference != "" || c.Caption != "" || !text(c.Text, 4000, true, true) || (c.Role != "hypothesis" && c.Role != "finding" && c.Role != "conclusion" && c.Role != "next-step") {
			return errors.New("note requires a supported role and 1..4000 UTF-8 bytes")
		}
	case "capture", "deployment":
		if !ID.MatchString(c.Reference) || !text(c.Caption, 160, false, false) || c.Role != "" || c.Text != "" {
			return errors.New("reference card requires an exact capture/deployment ID and optional caption")
		}
	default:
		return errors.New("card kind must be note, capture or deployment")
	}
	return nil
}

type DocumentInput struct {
	Title         string `json:"title"`
	Question      string `json:"question,omitempty"`
	Finding       string `json:"finding,omitempty"`
	OpenQuestions string `json:"openQuestions,omitempty"`
	Cards         []Card `json:"cards"`
}

func (d DocumentInput) Validate() error {
	if !text(d.Title, 120, true, false) || (len(d.Cards) < 1 && strings.TrimSpace(d.Question) == "") || len(d.Cards) > 32 || !text(d.Question, 4000, false, true) || !text(d.Finding, 4000, false, true) || !text(d.OpenQuestions, 4000, false, true) {
		return errors.New("document requires title, a question or 1..32 ordered cards, and bounded investigation summaries")
	}
	for _, c := range d.Cards {
		if e := c.Validate(); e != nil {
			return e
		}
	}
	return nil
}

type Document struct {
	ID         string `json:"id"`
	Version    int    `json:"version"`
	Parent     string `json:"parent,omitempty"`
	Root       string `json:"root,omitempty"`
	RecordedAt string `json:"recordedAt"`
	DocumentInput
}

// Revisions deliberately branch from an explicit parent; there is no mutable
// latest head and no silent last-write-wins update. References never requery a source.
func (s Store) CreateDocument(in DocumentInput, parent string, checkCapture func(string) error) (*Document, error) {
	if in.Cards == nil {
		in.Cards = []Card{}
	}
	if e := in.Validate(); e != nil {
		return nil, e
	}
	d := Document{Version: 1, Parent: parent, RecordedAt: time.Now().UTC().Format(time.RFC3339Nano), DocumentInput: in}
	if parent != "" {
		p, e := s.LoadDocument(parent)
		if e != nil {
			return nil, e
		}
		d.Root = p.Root
		if d.Root == "" {
			d.Root = p.ID
		}
	}
	for _, c := range in.Cards {
		switch c.Kind {
		case "capture":
			if checkCapture == nil {
				return nil, errors.New("capture reference resolver required")
			}
			if e := checkCapture(c.Reference); e != nil {
				return nil, errors.New("capture reference unavailable")
			}
		case "deployment":
			if _, e := s.LoadDeployment(c.Reference); e != nil {
				return nil, errors.New("deployment reference unavailable")
			}
		}
	}
	id, e := digest(d)
	if e != nil {
		return nil, e
	}
	d.ID = id
	if e = s.publish("documents", id, d); e != nil {
		return nil, e
	}
	return s.LoadDocument(id)
}
func (s Store) LoadDocument(id string) (*Document, error) {
	var d Document
	if e := s.read("documents", id, &d); e != nil {
		return nil, e
	}
	actual := d.ID
	d.ID = ""
	hash, e := digest(d)
	d.ID = actual
	if e != nil || actual != id || hash != id || d.Version != 1 || !validTime(d.RecordedAt) || d.DocumentInput.Validate() != nil || (d.Parent != "" && (!ID.MatchString(d.Parent) || !ID.MatchString(d.Root))) || (d.Parent == "" && d.Root != "") {
		return nil, errors.New("document integrity check failed")
	}
	return &d, nil
}

type DocumentList struct {
	Items      []Document `json:"items"`
	NextCursor string     `json:"nextCursor,omitempty"`
}

func (s Store) Documents(after string) (*DocumentList, error) {
	if after != "" && !ID.MatchString(after) {
		return nil, errors.New("invalid document cursor")
	}
	ids, e := s.ids("documents")
	if e != nil {
		return nil, e
	}
	r := &DocumentList{Items: []Document{}}
	for _, id := range ids {
		if id <= after {
			continue
		}
		if len(r.Items) == 20 {
			r.NextCursor = r.Items[19].ID
			break
		}
		d, e := s.LoadDocument(id)
		if e != nil {
			return nil, e
		}
		r.Items = append(r.Items, *d)
	}
	return r, nil
}
