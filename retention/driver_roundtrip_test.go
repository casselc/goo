package retention

import (
	"context"
	"encoding/json"
	"os"
	"reflect"
	"testing"
	"time"
)

// Exercise the real driver's Simple Protocol text encoding, not only the pure
// manifest codec. A retained JSON number must never travel through float64.
func TestXTDBDriverStringParameterRoundTrip336(t *testing.T) {
	address := os.Getenv("GOO_XTDB_TEST_URL")
	if address == "" {
		t.Skip("optional XTDB instance required")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()

	service := "api's \\雪"
	body := "quoted 'value' \\path\nnext line 雪 <tag>&"
	note := "review 'quoted' \\path\nsecond line 雪"
	const clock = "1792629000000000001"
	const large = "18446744073709551615"
	row, err := json.Marshal(map[string]any{
		"ServiceName": service, "Body": body, "timestamp_ns": clock,
		"large": json.Number(large), "empty": "",
	})
	if err != nil {
		t.Fatal(err)
	}
	capture := sample()
	capture.ID = hash([]byte(t.Name() + time.Now().String()))
	capture.Explore.Rows = []json.RawMessage{row}
	manifest, err := Select(capture, []int{0})
	if err != nil {
		t.Fatal(err)
	}
	encoded, err := json.Marshal(manifest)
	if err != nil {
		t.Fatal(err)
	}
	store := Store{URL: address}
	receipt, err := store.Publish(ctx, manifest, []string{"review:driver"}, note, hash([]byte(capture.ID+"request")))
	if err != nil {
		t.Fatal(err)
	}
	if receipt.ManifestJSON != string(encoded) || receipt.Annotation.Note != note || receipt.AwaitToken == "" {
		t.Fatal("publication changed canonical JSON, note or committed basis")
	}
	store.AwaitToken = receipt.AwaitToken
	loaded, err := store.Load(ctx, manifest.ID)
	if err != nil || !reflect.DeepEqual(loaded, &manifest) {
		t.Fatal("manifest round trip", loaded, err)
	}
	annotations, err := store.Annotations(ctx, manifest.ID)
	if err != nil || len(annotations.Items) != 1 || annotations.Items[0].Revision.Note != note {
		t.Fatal("annotation text parameter round trip", annotations, err)
	}

	connection, err := store.connect(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer connection.Close(ctx)
	var stored, projectedService, projectedClock, projectedRow string
	if err = connection.QueryRow(ctx, "SELECT payload FROM goo_evidence WHERE _id=$1", manifest.ID).Scan(&stored); err != nil || stored != string(encoded) {
		t.Fatal("stored canonical JSON changed", err)
	}
	if err = connection.QueryRow(ctx, "SELECT service,event_time_ns,payload FROM goo_evidence_members WHERE manifest_id=$1", manifest.ID).Scan(&projectedService, &projectedClock, &projectedRow); err != nil || projectedService != service || projectedClock != clock || projectedRow != string(manifest.Members[0].Row) {
		t.Fatal("indexed text parameter round trip", err)
	}
	var tokens map[string]json.RawMessage
	if err = json.Unmarshal([]byte(projectedRow), &tokens); err != nil || string(tokens["large"]) != large || string(tokens["timestamp_ns"]) != `"`+clock+`"` || string(tokens["empty"]) != `""` {
		t.Fatal("driver changed exact numeric/time or empty-string tokens", err)
	}
}
