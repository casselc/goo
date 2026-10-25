package main

import (
	"reflect"
	"testing"
)

func TestSourceSummaryCannotPublishSourceContent(t *testing.T) {
	got := sourceSummary(`{"read_rows":"123","read_bytes":"18446744073709551615","elapsed_ns":"0","query":"secret record","result_bytes":"record contents","written_rows":"01","written_bytes":"18446744073709551616","result_rows":-1,"total_rows_to_read":123}`)
	want := map[string]string{"read_rows": "123", "read_bytes": "18446744073709551615", "elapsed_ns": "0"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unsafe summary: %#v", got)
	}
	for _, raw := range []string{`null`, `"source content"`, `[]`, `{"read_rows":{}}`, `invalid`} {
		if len(sourceSummary(raw)) != 0 {
			t.Fatalf("accepted malformed summary %s", raw)
		}
	}
}
