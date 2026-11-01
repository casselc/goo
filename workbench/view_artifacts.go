package workbench

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

const ViewArtifactVersion = "goo.view-artifact/1"
const ViewArtifactBytes = 65536

// ViewArtifact labels a portable fixed question. Labels do not resolve a registry
// or change the content-addressed identity of the saved ExplorationView.
type ViewArtifact struct {
	Version   string          `json:"version"`
	Namespace string          `json:"namespace"`
	Name      string          `json:"name"`
	Revision  string          `json:"revision"`
	View      ExplorationView `json:"view"`
}
type ViewArtifactEncodeRequest struct {
	Artifact ViewArtifact `json:"artifact"`
}
type ViewArtifactDecodeRequest struct {
	YAML string `json:"yaml"`
}
type ViewArtifactResult struct {
	Artifact ViewArtifact `json:"artifact"`
	YAML     string       `json:"yaml"`
}

var artifactSlug = regexp.MustCompile(`^[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?$`)
var artifactIntegerLiteral = regexp.MustCompile(`^[+-]?(?:[0-9][0-9_]*|0[xX][0-9a-fA-F_]+|0[oO][0-7_]+|0[bB][01_]+)$`)
var artifactRevision = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]{0,63}$`)

func (a ViewArtifact) Validate() error {
	if a.Version != ViewArtifactVersion || !artifactSlug.MatchString(a.Namespace) || !artifactSlug.MatchString(a.Name) || !artifactRevision.MatchString(a.Revision) {
		return errors.New("artifact requires goo.view-artifact/1, namespace/name slugs (1..63), and revision label (1..64)")
	}
	if a.View.ID != "" {
		return errors.New("portable view must omit store-local ID")
	}
	if err := a.View.Validate(); err != nil {
		return err
	}
	if a.View.Tail != nil && a.View.Tail.Baseline != "" {
		return errors.New("portable view cannot reference a store-local tail baseline capture")
	}
	if a.View.Composition != nil {
		for _, child := range a.View.Composition.Children {
			if child.Comparison != nil {
				return errors.New("portable view cannot reference a store-local comparison capture; remove that child explicitly")
			}
		}
	}
	return nil
}

// EncodeViewArtifact validates with the existing view owner and produces detached,
// canonical YAML. It never saves, evaluates, or acquires telemetry.
func EncodeViewArtifact(a ViewArtifact) (*ViewArtifactResult, error) {
	if err := a.Validate(); err != nil {
		return nil, err
	}
	raw, err := json.Marshal(a)
	if err != nil {
		return nil, err
	}
	var detached ViewArtifact
	if err = json.Unmarshal(raw, &detached); err != nil {
		return nil, err
	}
	var node yaml.Node
	if err = yaml.Unmarshal(raw, &node); err != nil {
		return nil, err
	}
	var block func(*yaml.Node)
	block = func(n *yaml.Node) {
		n.Style = 0
		for _, c := range n.Content {
			block(c)
		}
	}
	block(&node)
	out, err := yaml.Marshal(&node)
	if err != nil {
		return nil, err
	}
	if len(out) > ViewArtifactBytes {
		return nil, errors.New("artifact YAML exceeds 64 KiB")
	}
	return &ViewArtifactResult{Artifact: detached, YAML: string(out)}, nil
}

// DecodeViewArtifact bridges a bounded YAML AST into the existing strict JSON
// types. YAML-specific constructs cannot bypass the JSON/view validation owner.
func DecodeViewArtifact(input string) (*ViewArtifactResult, error) {
	if len(input) > ViewArtifactBytes {
		return nil, errors.New("artifact YAML exceeds 64 KiB")
	}
	decoder := yaml.NewDecoder(strings.NewReader(input))
	var root yaml.Node
	if err := decoder.Decode(&root); err != nil {
		return nil, err
	}
	if err := decoder.Decode(new(yaml.Node)); err != io.EOF {
		return nil, errors.New("require exactly one YAML document")
	}
	nodes := 0
	var convert func(*yaml.Node, int) (any, error)
	convert = func(n *yaml.Node, depth int) (any, error) {
		nodes++
		if depth > 32 || nodes > 8192 {
			return nil, errors.New("artifact exceeds YAML structure budget")
		}
		if n.Anchor != "" || n.Kind == yaml.AliasNode || n.Style&yaml.TaggedStyle != 0 {
			return nil, errors.New("YAML anchors, aliases and explicit tags unsupported")
		}
		switch n.Kind {
		case yaml.DocumentNode:
			if len(n.Content) != 1 {
				return nil, errors.New("require one YAML mapping")
			}
			return convert(n.Content[0], depth+1)
		case yaml.MappingNode:
			result := map[string]any{}
			for i := 0; i < len(n.Content); i += 2 {
				key := n.Content[i]
				if key.Kind != yaml.ScalarNode || key.Tag != "!!str" || key.Style&yaml.TaggedStyle != 0 || key.Anchor != "" {
					return nil, errors.New("YAML mapping keys must be strings")
				}
				if _, ok := result[key.Value]; ok {
					return nil, fmt.Errorf("duplicate YAML key %q", key.Value)
				}
				value, err := convert(n.Content[i+1], depth+1)
				if err != nil {
					return nil, err
				}
				result[key.Value] = value
			}
			return result, nil
		case yaml.SequenceNode:
			result := make([]any, 0, len(n.Content))
			for _, c := range n.Content {
				v, err := convert(c, depth+1)
				if err != nil {
					return nil, err
				}
				result = append(result, v)
			}
			return result, nil
		case yaml.ScalarNode:
			switch n.Tag {
			case "!!str", "!!bool", "!!int", "!!float", "!!null":
			default:
				return nil, errors.New("unsupported YAML scalar tag")
			}
			var value any
			if err := n.Decode(&value); err != nil {
				return nil, err
			}
			// yaml.v3 falls back to float64 for integer literals outside its
			// exact integer range; never round such an authored value.
			if n.Tag == "!!float" && artifactIntegerLiteral.MatchString(n.Value) {
				return nil, errors.New("YAML integer exceeds exact 64-bit range")
			}
			if n.Tag == "!!int" {
				switch value.(type) {
				case int, int64, uint64:
				default:
					return nil, errors.New("YAML integer exceeds exact 64-bit range")
				}
			}
			return value, nil
		default:
			return nil, errors.New("unsupported YAML node")
		}
	}
	value, err := convert(&root, 0)
	if err != nil {
		return nil, err
	}
	raw, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	var artifact ViewArtifact
	jsonDecoder := json.NewDecoder(bytes.NewReader(raw))
	jsonDecoder.DisallowUnknownFields()
	if err = jsonDecoder.Decode(&artifact); err != nil {
		return nil, err
	}
	return EncodeViewArtifact(artifact)
}
