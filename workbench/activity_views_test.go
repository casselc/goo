package workbench

import (
	"encoding/json"
	"github.com/casselc/goo/exploration"
	"os"
	"path/filepath"
	"testing"
)

func activityFixture() ActivityView {
	return ActivityView{Version: 1, Title: "API activity", Selector: exploration.TailSelector{Signal: "logs", Environment: "prod"}, Columns: []string{"summary", "utc"}, FieldColumns: []exploration.Field{{Kind: "resource", Key: "service.namespace"}}}
}
func TestActivityViewConfigurationAndIntegrity(t *testing.T) {
	s := Store{Dir: t.TempDir()}
	v := activityFixture()
	a, e := s.CreateActivityView(v)
	if e != nil {
		t.Fatal(e)
	}
	b, e := s.CreateActivityView(v)
	if e != nil || a.ID != b.ID {
		t.Fatal(e)
	}
	v.Columns[0] = "service"
	loaded, e := s.LoadActivityView(a.ID)
	if e != nil || loaded.Columns[0] != "summary" {
		t.Fatal(loaded, e)
	}
	raw, _ := json.Marshal(loaded)
	var keys map[string]any
	json.Unmarshal(raw, &keys)
	for _, k := range []string{"queries", "fromMs", "toMs", "rows", "token", "connection", "resultToken"} {
		if _, ok := keys[k]; ok {
			t.Fatal("Activity view persisted execution/custody", k)
		}
	}
	loaded.Title = "tampered"
	raw, _ = json.Marshal(loaded)
	os.WriteFile(filepath.Join(s.Dir, "activity-views", a.ID+".json"), raw, 0600)
	if _, e = s.LoadActivityView(a.ID); e == nil {
		t.Fatal("accepted corruption")
	}
}
func TestActivityViewValidationAndEnvironmentPaging(t *testing.T) {
	for _, mutate := range []func(*ActivityView){func(v *ActivityView) { v.Selector.Signal = "metrics" }, func(v *ActivityView) { v.Selector.Environment = "" }, func(v *ActivityView) { v.Title = "" }, func(v *ActivityView) { v.Columns = []string{"utc", "context", "summary"} }, func(v *ActivityView) { v.FieldColumns = []exploration.Field{{Kind: "column", Key: "SpanName"}} }} {
		v := activityFixture()
		mutate(&v)
		if v.Validate() == nil {
			t.Fatal("accepted invalid configuration", v)
		}
	}
	s := Store{Dir: t.TempDir()}
	for i := 0; i < 23; i++ {
		v := activityFixture()
		v.Title = string(rune('A' + i))
		if _, e := s.CreateActivityView(v); e != nil {
			t.Fatal(e)
		}
	}
	v := activityFixture()
	v.Selector.Environment = "other"
	s.CreateActivityView(v)
	a, e := s.ActivityViews("", "prod")
	if e != nil || len(a.Items) != 20 || a.NextCursor == "" {
		t.Fatal(a, e)
	}
	b, e := s.ActivityViews(a.NextCursor, "prod")
	if e != nil || len(b.Items) != 3 {
		t.Fatal(b, e)
	}
	for _, v := range append(a.Items, b.Items...) {
		if v.Selector.Environment != "prod" {
			t.Fatal("mixed environment")
		}
	}
}

func TestActivityBaselineReferenceIdentityAndLegacyBytes(t *testing.T) {
	s := Store{Dir: t.TempDir()}
	v := activityFixture()
	// The old wire projection pins pre-extension content IDs and permits no new default field.
	legacy := struct {
		Version      int                      `json:"version"`
		ID           string                   `json:"id,omitempty"`
		Title        string                   `json:"title"`
		Selector     exploration.TailSelector `json:"selector"`
		Columns      []string                 `json:"columns"`
		FieldColumns []exploration.Field      `json:"fieldColumns,omitempty"`
	}{v.Version, "", v.Title, v.Selector, v.Columns, v.FieldColumns}
	want, e := digest(legacy)
	if e != nil {
		t.Fatal(e)
	}
	old, e := s.CreateActivityView(v)
	if e != nil || old.ID != want {
		t.Fatal("legacy identity drift", e)
	}
	v.Selector.Signal = "spans"
	v.BaselineCapture = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	a, e := s.CreateActivityView(v)
	if e != nil {
		t.Fatal(e)
	}
	v.BaselineCapture = "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
	b, e := s.CreateActivityView(v)
	if e != nil || a.ID == b.ID {
		t.Fatal("baseline not bound to view identity", e)
	}
	loaded, e := s.LoadActivityView(a.ID)
	if e != nil || loaded.BaselineCapture != a.BaselineCapture {
		t.Fatal("reference changed", e)
	}
	loaded.BaselineCapture = b.BaselineCapture
	raw, _ := json.Marshal(loaded)
	os.WriteFile(filepath.Join(s.Dir, "activity-views", a.ID+".json"), raw, 0600)
	if _, e = s.LoadActivityView(a.ID); e == nil {
		t.Fatal("accepted reference tampering")
	}
	for _, bad := range []string{"x", "https://other/#capture=" + a.BaselineCapture} {
		v.BaselineCapture = bad
		if v.Validate() == nil {
			t.Fatal("accepted invalid reference")
		}
	}
	v.BaselineCapture = a.BaselineCapture
	v.Selector.Signal = "logs"
	if v.Validate() == nil {
		t.Fatal("logs accepted latency baseline")
	}
}
