package identity

import (
	"encoding/json"
	"gopkg.in/yaml.v3"
	"os"
	"reflect"
	"testing"
)

func TestEnvironmentSelectorProtocolBank(t *testing.T) {
	data, err := os.ReadFile("../testdata/protocol/environment-selectors.json")
	if err != nil {
		t.Fatal(err)
	}
	var bank struct {
		Resources []map[string]string `json:"resources"`
		Cases     []struct {
			Name      string
			Selector  json.RawMessage
			Canonical EnvironmentSelector
			Matches   []bool
		}
		Invalid []json.RawMessage
	}
	if err = json.Unmarshal(data, &bank); err != nil {
		t.Fatal(err)
	}
	for _, fixture := range bank.Cases {
		t.Run(fixture.Name, func(t *testing.T) {
			var wire EnvironmentSelector
			if err := json.Unmarshal(fixture.Selector, &wire); err != nil {
				t.Fatal(err)
			}
			selection, err := wire.Selection()
			if err != nil {
				t.Fatal(err)
			}
			canonical, err := selection.Selector()
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(canonical, fixture.Canonical) {
				t.Fatal("wire roundtrip changed", canonical)
			}
			for i, resource := range bank.Resources {
				if selection.Matches(resource) != fixture.Matches[i] {
					t.Fatal("membership changed", i)
				}
			}
			raw, err := json.Marshal(canonical)
			if err != nil {
				t.Fatal(err)
			}
			var again EnvironmentSelector
			if err = json.Unmarshal(raw, &again); err != nil {
				t.Fatal(err)
			}
			same, _ := again.Selection()
			if !same.Equal(selection) {
				t.Fatal("JSON lost policy")
			}
			raw, err = yaml.Marshal(canonical)
			if err != nil {
				t.Fatal(err)
			}
			if err = yaml.Unmarshal(raw, &again); err != nil {
				t.Fatal(err)
			}
			same, err = again.Selection()
			if err != nil || !same.Equal(selection) {
				t.Fatal("YAML lost policy", err)
			}
		})
	}
	for _, raw := range bank.Invalid {
		var wire EnvironmentSelector
		if err = json.Unmarshal(raw, &wire); err == nil {
			t.Fatalf("invalid wire accepted: %s", raw)
		}
	}
	if _, err := (EnvironmentSelection{}).Selector(); err == nil {
		t.Fatal("zero selector encoded")
	}
}

func TestEnvironmentSelectorFailedDecodeClearsSelection(t *testing.T) {
	selection, _ := CurrentEnvironment("prod")
	wire, _ := selection.Selector()
	if err := json.Unmarshal([]byte(`{"Version":"goo/environment-selector/v1"}`), &wire); err == nil || wire.Validate() == nil {
		t.Fatal("failed decode retained previous checked policy")
	}
}

func TestEnvironmentSelectorRawUnicodeBank249(t *testing.T) {
	data, err := os.ReadFile("../testdata/protocol/environment-selectors.json")
	if err != nil {
		t.Fatal(err)
	}
	var bank struct {
		InvalidJSON []string
		ValidJSON   []string
	}
	if err = json.Unmarshal(data, &bank); err != nil {
		t.Fatal(err)
	}
	for _, raw := range bank.InvalidJSON {
		var wire EnvironmentSelector
		if err = json.Unmarshal([]byte(raw), &wire); err == nil {
			t.Fatalf("lossy JSON accepted %s", raw)
		}
	}
	for _, raw := range bank.ValidJSON {
		var wire EnvironmentSelector
		if err = json.Unmarshal([]byte(raw), &wire); err != nil {
			t.Fatalf("valid Unicode refused %s: %v", raw, err)
		}
	}
	var wire EnvironmentSelector
	raw := []byte(`{"version":"goo/environment-selector/v1","policy":"effective","scope":{"mode":"named","value":"X"}}`)
	for i := range raw {
		if raw[i] == 'X' {
			raw[i] = 0xff
		}
	}
	if err = json.Unmarshal(raw, &wire); err == nil {
		t.Fatal("invalid raw UTF-8 accepted")
	}
}
