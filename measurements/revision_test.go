package measurements

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/casselc/goo/identity"
)

func TestDefinitionRevisionPreservesExistingIdentitiesWithoutMutation(t *testing.T) {
	files, err := filepath.Glob("../config/measurements/*.yaml")
	if err != nil || len(files) == 0 {
		t.Fatal("missing built-in definitions", err)
	}
	for _, file := range files {
		t.Run(filepath.Base(file), func(t *testing.T) {
			data, err := os.ReadFile(file)
			if err != nil {
				t.Fatal(err)
			}
			d, err := Read(string(data))
			if err != nil {
				t.Fatal(err)
			}
			before, err := json.Marshal(d)
			if err != nil {
				t.Fatal(err)
			}
			got, err := d.Revision()
			if err != nil {
				t.Fatal(err)
			}
			if got != Hash(d) {
				t.Fatal("existing definition identity changed")
			}
			after, err := json.Marshal(d)
			if err != nil || !bytes.Equal(before, after) {
				t.Fatal("revision mutated definition", err)
			}
		})
	}
	data, err := os.ReadFile("../testdata/storage/native-measurement.yaml")
	if err != nil {
		t.Fatal(err)
	}
	d, err := Read(string(data))
	if err != nil {
		t.Fatal(err)
	}
	data, err = os.ReadFile("../testdata/storage/identities.json")
	if err != nil {
		t.Fatal(err)
	}
	var ids map[string]string
	if err = json.Unmarshal(data, &ids); err != nil {
		t.Fatal(err)
	}
	got, err := d.Revision()
	if err != nil || got != ids["nativeMeasurementRevision"] {
		t.Fatal("frozen stored revision changed", got, err)
	}
}
func TestDefinitionRevisionRetainsNamespaceAndMapRepresentations(t *testing.T) {
	d := definition(t, "worker-utilization")
	for _, namespace := range []*identity.Namespace{nil, {Mode: "missing"}, {Mode: "value", Value: ""}, {Mode: "value", Value: "worker-team"}} {
		for _, attributes := range []map[string]string{nil, {}, {"custom.attr": ""}} {
			d.Source.Namespace = namespace
			d.Source.Attributes = attributes
			got, err := d.Revision()
			if err != nil || got != Hash(d) {
				t.Fatal("definition representation rewritten", got, err)
			}
		}
	}
}
func TestDefinitionRevisionRejectsInvalidDefinitionsWithoutAnIdentifier(t *testing.T) {
	valid := definition(t, "worker-utilization")
	cases := []Definition{{}, valid, valid, valid, valid}
	cases[1].Name = ""
	cases[2].Environment = string([]byte{0xff})
	cases[3].Kind = "unsupported"
	cases[4].LookbackMinutes = 0
	for _, d := range cases {
		got, err := d.Revision()
		if err == nil || got != "" {
			t.Fatal("invalid definition acquired an identifier", got, err)
		}
	}
}
