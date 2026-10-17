// Package prebuilt recognizes telemetry contracts. Compatibility proposes
// capabilities; it never proves completeness or installs a measurement/SLO.
package prebuilt

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"sort"

	"gopkg.in/yaml.v3"
)

//go:embed catalog.yaml
var builtin []byte

type Contract struct {
	Kind       string   `yaml:"kind" json:"kind"`
	Name       string   `yaml:"name" json:"name"`
	Unit       string   `yaml:"unit" json:"unit"`
	SchemaURL  string   `yaml:"schemaURL" json:"schemaURL"`
	Attributes []string `yaml:"attributes" json:"attributes"`
}
type Requirement struct {
	ID    string     `yaml:"id" json:"id"`
	AnyOf []Contract `yaml:"anyOf" json:"anyOf"`
}
type Feature struct {
	ID             string `yaml:"id" json:"id"`
	Kind           string `yaml:"kind" json:"kind"`
	Implementation string `yaml:"implementation" json:"implementation"`
}
type Profile struct {
	ID           string        `yaml:"id" json:"id"`
	Description  string        `yaml:"description" json:"description"`
	Stability    string        `yaml:"stability" json:"stability"`
	Resources    []string      `yaml:"resources" json:"resources"`
	Requirements []Requirement `yaml:"requirements" json:"requirements"`
	Features     []Feature     `yaml:"features" json:"features"`
	Caveats      []string      `yaml:"caveats" json:"caveats"`
}
type Registry struct {
	APIVersion string    `yaml:"apiVersion" json:"apiVersion"`
	Profiles   []Profile `yaml:"profiles" json:"profiles"`
}
type Scope struct {
	ID       string            `json:"id"`
	Resource map[string]string `json:"resource"`
	Signals  []Contract        `json:"signals"`
}
type Inventory struct {
	APIVersion string  `json:"apiVersion"`
	Basis      string  `json:"basis"`
	Scopes     []Scope `json:"scopes"`
}
type Check struct {
	Requirement string     `json:"requirement"`
	Matched     bool       `json:"matched"`
	Expected    []Contract `json:"expected"`
}
type Proposal struct {
	Profile          string    `json:"profile"`
	Scope            string    `json:"scope"`
	Basis            string    `json:"basis"`
	Status           string    `json:"status"`
	MissingResources []string  `json:"missingResources"`
	Checks           []Check   `json:"checks"`
	Features         []Feature `json:"features"`
	Caveats          []string  `json:"caveats"`
}

func Catalog() Registry {
	var r Registry
	if err := yaml.Unmarshal(builtin, &r); err != nil {
		panic(err)
	}
	return r
}
func ReadInventory(r io.Reader) (Inventory, error) {
	var in Inventory
	body, err := io.ReadAll(io.LimitReader(r, (4<<20)+1))
	if err != nil {
		return in, err
	}
	if len(body) > 4<<20 {
		return in, fmt.Errorf("inventory exceeds 4 MiB")
	}
	dec := json.NewDecoder(bytes.NewReader(body))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&in); err != nil {
		return in, err
	}
	if err := dec.Decode(new(any)); err != io.EOF {
		return in, fmt.Errorf("inventory must contain one JSON document")
	}
	return in, validate(in)
}
func validate(in Inventory) error {
	if in.APIVersion != "goo/inventory/v1alpha1" || in.Basis == "" {
		return fmt.Errorf("inventory requires goo/inventory/v1alpha1 and an evidence basis")
	}
	seen := map[string]bool{}
	for _, s := range in.Scopes {
		if s.ID == "" || seen[s.ID] {
			return fmt.Errorf("scope IDs must be unique and nonempty")
		}
		seen[s.ID] = true
		for _, c := range s.Signals {
			if c.Kind == "" {
				return fmt.Errorf("signal kind is required")
			}
		}
	}
	return nil
}
func contains(haystack []string, needle string) bool {
	for _, s := range haystack {
		if s == needle {
			return true
		}
	}
	return false
}
func matches(actual, expected Contract) bool {
	if actual.Kind != expected.Kind || (expected.Name != "" && actual.Name != expected.Name) || actual.Unit != expected.Unit || actual.SchemaURL != expected.SchemaURL {
		return false
	}
	for _, key := range expected.Attributes {
		if !contains(actual.Attributes, key) {
			return false
		}
	}
	return true
}

// Discover joins requirements only within an explicitly supplied resource scope.
// Every contract must match one signal signature: attributes from unrelated
// series are never unioned. Unknown units or versions cannot activate a profile.
func Discover(in Inventory) ([]Proposal, error) {
	if err := validate(in); err != nil {
		return nil, err
	}
	out := []Proposal{}
	for _, scope := range in.Scopes {
		for _, profile := range Catalog().Profiles {
			p := Proposal{Profile: profile.ID, Scope: scope.ID, Basis: in.Basis, Status: "compatible-shape", MissingResources: []string{}, Checks: []Check{}, Features: profile.Features, Caveats: profile.Caveats}
			for _, key := range profile.Resources {
				if scope.Resource[key] == "" {
					p.MissingResources = append(p.MissingResources, key)
				}
			}
			if len(p.MissingResources) > 0 {
				p.Status = "needs-evidence-or-binding"
			}
			for _, req := range profile.Requirements {
				check := Check{Requirement: req.ID, Expected: req.AnyOf}
				for _, alternative := range req.AnyOf {
					for _, actual := range scope.Signals {
						if matches(actual, alternative) {
							check.Matched = true
						}
					}
				}
				if !check.Matched {
					p.Status = "needs-evidence-or-binding"
				}
				p.Checks = append(p.Checks, check)
			}
			out = append(out, p)
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Scope == out[j].Scope {
			return out[i].Profile < out[j].Profile
		}
		return out[i].Scope < out[j].Scope
	})
	return out, nil
}

// JSON is stable under inventory ordering and duplicate signatures.
func JSON(in Inventory) ([]byte, error) {
	p, e := Discover(in)
	if e != nil {
		return nil, e
	}
	var b bytes.Buffer
	enc := json.NewEncoder(&b)
	enc.SetIndent("", "  ")
	e = enc.Encode(p)
	return b.Bytes(), e
}
