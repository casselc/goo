package measurements

import (
	"encoding/json"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestHostAuthoringCanonicalServicesRoundtrip(t *testing.T) {
	definition := hostResourceQuestion(t, "gauge").Definition
	encoded, e := yaml.Marshal(definition)
	if e != nil {
		t.Fatal(e)
	}
	for _, form := range []string{"omitted", "null", "empty"} {
		t.Run(form, func(t *testing.T) {
			lines := strings.Split(string(encoded), "\n")
			for i, line := range lines {
				if strings.HasPrefix(strings.TrimSpace(line), "services:") {
					switch form {
					case "omitted":
						lines = append(lines[:i], lines[i+1:]...)
					case "null":
						lines[i] = strings.Replace(line, "[]", "null", 1)
					}
					break
				}
			}
			authored, e := Read(strings.Join(lines, "\n"))
			if e != nil {
				t.Fatal(e)
			}
			if authored.Source.Services == nil {
				t.Fatal("host authoring retained noncanonical nil services")
			}
			published, e := yaml.Marshal(authored)
			if e != nil {
				t.Fatal(e)
			}
			reopened, e := Read(string(published))
			if e != nil {
				t.Fatal(e)
			}
			if Hash(authored) != Hash(reopened) {
				t.Fatal("host publication changed definition identity on reopen")
			}
		})
	}
}
func TestHostJSONNilServicesRemainUnchanged(t *testing.T) {
	q := hostResourceQuestion(t, "gauge")
	q.Definition.Source.Services = nil
	before, e := json.Marshal(q)
	if e != nil {
		t.Fatal(e)
	}
	if e = q.Validate(); e != nil {
		t.Fatal(e)
	}
	after, e := json.Marshal(q)
	if e != nil {
		t.Fatal(e)
	}
	if string(before) != string(after) || q.Definition.Source.Services != nil || !strings.Contains(string(after), `"services":null`) {
		t.Fatal("direct JSON or retained question was normalized")
	}
}

func TestHostAuthoringCanonicalEmptyStreamAttributes(t *testing.T) {
	for _, servicesNil := range []bool{false, true} {
		for _, form := range []string{"omitted", "null", "empty"} {
			d := hostResourceQuestion(t, "gauge").Definition
			if servicesNil {
				d.Source.Services = nil
			}
			d.Source.ExactStream.Attributes = map[string]string{}
			encoded, e := yaml.Marshal(d)
			if e != nil {
				t.Fatal(e)
			}
			lines := strings.Split(string(encoded), "\n")
			if servicesNil {
				for index, line := range lines {
					if strings.HasPrefix(strings.TrimSpace(line), "services:") {
						lines[index] = strings.Replace(line, "[]", "null", 1)
						break
					}
				}
			}
			for index, line := range lines {
				if strings.TrimSpace(line) == "attributes: {}" {
					switch form {
					case "omitted":
						lines = append(lines[:index], lines[index+1:]...)
					case "null":
						lines[index] = strings.Replace(line, "{}", "null", 1)
					}
					break
				}
			}
			authored, e := Read(strings.Join(lines, "\n"))
			if e != nil {
				t.Fatal(e)
			}
			if authored.Source.ExactStream.Attributes == nil {
				t.Fatalf("nil attributes after host Read: servicesNil=%v form=%s", servicesNil, form)
			}
			published, e := yaml.Marshal(authored)
			if e != nil {
				t.Fatal(e)
			}
			reopened, e := Read(string(published))
			if e != nil || Hash(authored) != Hash(reopened) {
				t.Fatal("dimensionless host hash drift", e)
			}
		}
	}
}
func TestHostJSONNilStreamAttributesRemainUnchanged(t *testing.T) {
	q := hostResourceQuestion(t, "gauge")
	q.Definition.Source.ExactStream.Attributes = nil
	before, e := json.Marshal(q)
	if e != nil {
		t.Fatal(e)
	}
	if e = q.Validate(); e != nil {
		t.Fatal(e)
	}
	after, e := json.Marshal(q)
	if e != nil {
		t.Fatal(e)
	}
	if string(before) != string(after) || q.Definition.Source.ExactStream.Attributes != nil || !strings.Contains(string(after), `"attributes":null`) {
		t.Fatal("direct JSON nil attributes normalized")
	}
}
