package measurements

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

// Workflow is a data-only ordered investigation. Opening a step previews a
// package; it never runs SQL or treats completion of a step as a finding.
type Workflow struct {
	APIVersion   string         `json:"apiVersion"`
	Name         string         `json:"name"`
	Title        string         `json:"title"`
	Description  string         `json:"description"`
	Requirements []string       `json:"requirements"`
	Steps        []WorkflowStep `json:"steps"`
}
type WorkflowStep struct {
	Title   string `json:"title"`
	Why     string `json:"why"`
	Family  string `json:"family"`
	Package string `json:"package"`
}

func ReadWorkflows(r io.Reader) ([]Workflow, error) {
	var workflows []Workflow
	data, e := io.ReadAll(io.LimitReader(r, 65537))
	if e != nil {
		return nil, e
	}
	if len(data) > 65536 {
		return nil, fmt.Errorf("workflow catalog exceeds 64 KiB")
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if e := decoder.Decode(&workflows); e != nil {
		return nil, e
	}
	if decoder.Decode(new(any)) != io.EOF {
		return nil, fmt.Errorf("one workflow document required")
	}
	if len(workflows) < 1 || len(workflows) > 32 {
		return nil, fmt.Errorf("workflow catalog size")
	}
	seen := map[string]bool{}
	for _, w := range workflows {
		if w.APIVersion != "goo/investigation-workflow/v1alpha1" || !stable.MatchString(w.Name) || seen[w.Name] || w.Title == "" || !text(w.Title, 160) || !text(w.Description, 1024) || len(w.Requirements) < 1 || len(w.Requirements) > 8 || len(w.Steps) < 1 || len(w.Steps) > 12 {
			return nil, fmt.Errorf("invalid workflow")
		}
		seen[w.Name] = true
		for _, s := range w.Requirements {
			if s == "" || !text(s, 512) {
				return nil, fmt.Errorf("invalid requirement")
			}
		}
		for _, s := range w.Steps {
			if s.Title == "" || !text(s.Title, 160) || s.Why == "" || !text(s.Why, 1024) || !stable.MatchString(s.Package) || strings.Contains(s.Package, "..") || (s.Family != "measurement" && s.Family != "telemetry") {
				return nil, fmt.Errorf("invalid workflow step")
			}
		}
	}
	return workflows, nil
}
