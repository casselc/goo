// Package annotations validates goo's bounded source annotation dialect.
// A manifest describes source declarations, not confirmed runtime coverage.
package annotations

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

type Declaration struct {
	Kind     string `json:"kind"`
	Function string `json:"function"`
	Name     string `json:"name"`
	Role     string `json:"role,omitempty"`
	Subject  string `json:"subject,omitempty"`
	Outcome  string `json:"outcome,omitempty"`
	When     string `json:"when,omitempty"`
	Unit     string `json:"unit"`
}
type Source struct {
	Path   string `json:"path"`
	Digest string `json:"digest"`
}
type Signal struct {
	Name       string   `json:"name"`
	Kind       string   `json:"kind"`
	Unit       string   `json:"unit"`
	Dimensions []string `json:"dimensions"`
}
type Manifest struct {
	APIVersion   string        `json:"apiVersion"`
	Revision     string        `json:"revision"`
	Package      string        `json:"package"`
	Scope        string        `json:"scope"`
	ScopeVersion string        `json:"scopeVersion"`
	SchemaURL    string        `json:"schemaURL"`
	Sources      []Source      `json:"sources"`
	Declarations []Declaration `json:"declarations"`
	Signals      []Signal      `json:"signals"`
	Requirements []string      `json:"requirements"`
}

var stableName = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9_.-]{0,127}$`)

func hash(v []byte) string { s := sha256.Sum256(v); return hex.EncodeToString(s[:]) }
func selectorRoot(value string) (string, error) {
	expression, err := parser.ParseExpr(value)
	if err != nil {
		return "", err
	}
	for {
		switch e := expression.(type) {
		case *ast.Ident:
			if e.Name == "_" {
				return "", fmt.Errorf("blank selector root")
			}
			return e.Name, nil
		case *ast.SelectorExpr:
			expression = e.X
		default:
			return "", fmt.Errorf("only identifiers and field selectors are supported")
		}
	}
}
func parse(fn *ast.FuncDecl, text string, contextNames map[string]bool) (Declaration, error) {
	var d Declaration
	parts := strings.Fields(strings.TrimSpace(strings.TrimPrefix(text, "//")))
	if len(parts) == 0 {
		return d, fmt.Errorf("empty directive")
	}
	d.Kind = strings.TrimPrefix(parts[0], "goo:")
	if d.Kind != "operation" && d.Kind != "event" {
		return d, fmt.Errorf("unsupported goo directive %s", d.Kind)
	}
	values := map[string]string{}
	for _, part := range parts[1:] {
		kv := strings.SplitN(part, ":", 2)
		if len(kv) != 2 || kv[1] == "" || values[kv[0]] != "" {
			return d, fmt.Errorf("invalid or duplicate annotation field %q", part)
		}
		values[kv[0]] = kv[1]
	}
	allowed := map[string]bool{"name": true}
	if d.Kind == "operation" {
		allowed["role"] = true
	} else {
		allowed["subject"] = true
		allowed["outcome"] = true
		allowed["when"] = true
	}
	for key := range values {
		if !allowed[key] {
			return d, fmt.Errorf("unsupported %s field %s", d.Kind, key)
		}
	}
	d.Name = values["name"]
	if !stableName.MatchString(d.Name) {
		return d, fmt.Errorf("require a stable bounded name")
	}
	names := map[string]bool{}
	contextCount := 0
	for _, field := range fn.Type.Params.List {
		for _, name := range field.Names {
			names[name.Name] = true
		}
		if selector, ok := field.Type.(*ast.SelectorExpr); ok {
			if root, ok := selector.X.(*ast.Ident); ok && contextNames[root.Name] && selector.Sel.Name == "Context" && len(field.Names) == 1 && field.Names[0].Name != "_" {
				contextCount++
			}
		}
	}
	if contextCount != 1 {
		return d, fmt.Errorf("require exactly one named context.Context parameter")
	}
	errorResults := 0
	if fn.Type.Results != nil {
		for _, field := range fn.Type.Results.List {
			for _, name := range field.Names {
				names[name.Name] = true
			}
			if typ, ok := field.Type.(*ast.Ident); ok && typ.Name == "error" {
				errorResults++
				if errorResults > 1 {
					return d, fmt.Errorf("only one error result is supported")
				}
				if len(field.Names) != 1 || field.Names[0].Name == "_" {
					return d, fmt.Errorf("error result must be named for return-time observation")
				}
			}
		}
	}
	if d.Kind == "operation" {
		d.Role = values["role"]
		switch d.Role {
		case "internal", "server", "client", "producer", "consumer":
		default:
			return d, fmt.Errorf("explicit supported operation role required")
		}
		d.Unit = "invocation"
	} else {
		d.Subject = values["subject"]
		d.Outcome = values["outcome"]
		d.When = values["when"]
		d.Unit = "subject-event"
		if d.When != "ok" {
			return d, fmt.Errorf("event requires explicit when:ok")
		}
		for _, expression := range []string{d.Subject, d.Outcome} {
			if expression == "" {
				continue
			}
			root, e := selectorRoot(expression)
			if e != nil || !names[root] {
				return d, fmt.Errorf("selector must name a parameter or named result: %s", expression)
			}
		}
		if d.Subject == "" {
			return d, fmt.Errorf("event subject is required")
		}
	}
	d.Function = fn.Name.Name
	if fn.Recv != nil {
		var receiver strings.Builder // receiver syntax is only used to distinguish methods in the manifest
		switch typ := fn.Recv.List[0].Type.(type) {
		case *ast.Ident:
			receiver.WriteString(typ.Name)
		case *ast.StarExpr:
			if id, ok := typ.X.(*ast.Ident); ok {
				receiver.WriteString("*" + id.Name)
			}
		default:
			return d, fmt.Errorf("generic receiver annotations are not supported in this slice")
		}
		if receiver.Len() == 0 {
			return d, fmt.Errorf("unsupported receiver")
		}
		d.Function = receiver.String() + "." + d.Function
	}
	return d, nil
}
func ScanDir(directory string) (*Manifest, error) {
	entries, err := os.ReadDir(directory)
	if err != nil {
		return nil, err
	}
	m := &Manifest{APIVersion: "goo/instrumentation-manifest/v1alpha1", Scope: "github.com/casselc/goo/instrumentation", ScopeVersion: "0.1.0", SchemaURL: "goo://instrumentation/v1", Sources: []Source{}, Declarations: []Declaration{}, Signals: []Signal{}, Requirements: []string{"Source declarations only; build constraints and dependency versions are not resolved.", "Run lint before weaving; Go compilation checks selector result types.", "Application configures OTel providers/exporters; counts describe invocations, not business subjects.", "Events require a recording parent span; missing subject or parent is a diagnosed drop.", "Return-time events do not prove transaction commit, durability or completeness.", "Subject IDs require explicit tenant/environment/business-key scope before joins or subject aggregation."}}
	functions := map[string]bool{}
	operationNames := map[string]bool{}
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") || name == "orchestrion.tool.go" {
			continue
		}
		data, e := os.ReadFile(filepath.Join(directory, name))
		if e != nil {
			return nil, e
		}
		file, e := parser.ParseFile(token.NewFileSet(), name, data, parser.ParseComments)
		if e != nil {
			return nil, e
		}
		contextNames := map[string]bool{}
		for _, imp := range file.Imports {
			if imp.Path.Value == `"context"` {
				alias := "context"
				if imp.Name != nil {
					alias = imp.Name.Name
				}
				contextNames[alias] = true
			}
		}
		found := false
		for _, declaration := range file.Decls {
			fn, ok := declaration.(*ast.FuncDecl)
			if !ok || fn.Doc == nil {
				continue
			}
			count := 0
			for _, comment := range fn.Doc.List {
				if !strings.HasPrefix(comment.Text, "//goo:") {
					continue
				}
				count++
				if count > 1 {
					return nil, fmt.Errorf("%s: one goo directive per function is supported", name)
				}
				d, e := parse(fn, comment.Text, contextNames)
				if e != nil {
					return nil, fmt.Errorf("%s.%s: %w", name, fn.Name.Name, e)
				}
				if functions[d.Function] {
					return nil, fmt.Errorf("duplicate annotated function %s", d.Function)
				}
				functions[d.Function] = true
				if d.Kind == "operation" {
					if operationNames[d.Name] {
						return nil, fmt.Errorf("duplicate operation name %s", d.Name)
					}
					operationNames[d.Name] = true
				}
				if m.Package != "" && m.Package != file.Name.Name {
					return nil, fmt.Errorf("mixed annotated packages")
				}
				m.Package = file.Name.Name
				m.Declarations = append(m.Declarations, d)
				found = true
			}
		}
		if found {
			m.Sources = append(m.Sources, Source{Path: name, Digest: hash(data)})
		}
	}
	sort.Slice(m.Declarations, func(i, j int) bool { return m.Declarations[i].Function < m.Declarations[j].Function })
	if len(m.Declarations) == 0 {
		return nil, fmt.Errorf("no supported goo annotations found")
	}
	if len(operationNames) > 0 {
		dims := []string{"goo.operation.name", "goo.operation.role", "goo.operation.outcome"}
		m.Signals = append(m.Signals, Signal{Name: "goo.operation.calls", Kind: "counter", Unit: "{invocation}", Dimensions: dims}, Signal{Name: "goo.operation.duration", Kind: "histogram", Unit: "s", Dimensions: dims}, Signal{Name: "goo.operation.active", Kind: "updowncounter", Unit: "{invocation}", Dimensions: dims[:2]})
	}
	for _, d := range m.Declarations {
		if d.Kind == "event" {
			m.Signals = append(m.Signals, Signal{Name: "goo.instrumentation.dropped_events", Kind: "counter", Unit: "{event}", Dimensions: []string{"goo.event.name", "goo.drop.reason"}})
			break
		}
	}
	b, _ := json.Marshal(m)
	m.Revision = hash(b)
	return m, nil
}
