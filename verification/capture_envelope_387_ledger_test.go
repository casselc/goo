package verification

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/observability/quintgo/qobs"
	"github.com/casselc/observability/quintgo/validate"
)

type captureCounters387 struct{ Source, SQL, Captures, Documents, StoreReads, Receipts int }
type captureFact387 struct {
	Name          string
	Before, After captureCounters387
}
type captureBank387 struct {
	Page                                              exploration.CapturePage
	StoredBytesBefore, StoredBytesAfter, SHA256Before string
}
type captureHeld387 struct {
	Family, ID, Mode, Notice    string
	Status, OriginalStatus      int
	Before, After               json.RawMessage
	OriginalPage, CorruptedPage json.RawMessage
}
type captureFile387 struct {
	Family, ForeignFamily, ForeignField, Mode, ID, SourceID, StoredBytes, CanonicalSHA256 string
	Status                                                                                int
	NullIgnoredDigestWouldMatch                                                           bool
}
type captureLedger387 struct {
	WrongPageTransport struct {
		RequestURL, Notice         string
		Status                     int
		OriginalPage, ReturnedPage json.RawMessage
		Before, After              json.RawMessage
	}
	UIPaging struct {
		Forward, Backward struct {
			RequestURL string
			Response   exploration.CapturePage
			Status     int
		}
	}
	Version string
	Errors  []string
	Fixture struct {
		API, Worker, Environment, Namespace, Absent string
		Selection                                   exploration.Query
	}
	RawSpans []struct {
		TraceID, SpanID, ServiceName, AtNS, DurationNS, StatusCode string
		ResourceAttributes, SpanAttributes                         map[string]string
	}
	RawMetrics               []gaugeRaw386
	Bank                     map[string]captureBank387
	EmptyExplore             captureBank387
	Pagination               exploration.CapturePage
	ObservedExplanation      json.RawMessage
	ExplanationPromotion     struct{ SQLBefore, SQLAfter int }
	SourceRefusalProbe       struct{ Status int }
	FileMutants              []captureFile387
	FileUI, TransportMutants []captureHeld387
	OfflineReopens           []struct {
		Family, ID string
		Snapshot   json.RawMessage
	}
	Document struct {
		Text         string
		ReceiptCalls []struct {
			Path     string
			Status   int
			Response json.RawMessage
		}
	}
	HeldParent struct{ Before, After json.RawMessage }
	Child      struct {
		Request  exploration.Query
		Response exploration.Result
		Status   int
	}
	Facts []captureFact387
	Calls []struct {
		Path, Method      string
		Body, Response    json.RawMessage
		Status            int
		NegativeTransport json.RawMessage
	}
	ProxyCalls []struct {
		SQL    string
		Status int
	}
}

func transplantCapture387(c *exploration.SavedCapture, donor exploration.SavedCapture, question bool) {
	switch donor.Question.Kind {
	case "explore":
		if question {
			c.Question.Explore = donor.Question.Explore
		} else {
			c.Explore = donor.Explore
		}
	case "inventory":
		if question {
			c.Question.Inventory = donor.Question.Inventory
		} else {
			c.Inventory = donor.Inventory
		}
	case "comparison":
		if question {
			c.Question.Comparison = donor.Question.Comparison
		} else {
			c.Comparison = donor.Comparison
		}
	case "measurement":
		if question {
			c.Question.Measurement = donor.Question.Measurement
		} else {
			c.Measurement = donor.Measurement
		}
	case "explanation":
		if question {
			c.Question.Explanation = donor.Question.Explanation
		} else {
			c.Explanation = donor.Explanation
		}
	case "trace-search":
		if question {
			c.Question.TraceSearch = donor.Question.TraceSearch
		} else {
			c.TraceSearch = donor.TraceSearch
		}
	}
}
func captureMutantDigest387(m captureFile387, bank map[string]captureBank387) bool {
	// Independently reconstruct the specified attack from the original typed
	// source/donor, rather than trusting a fixture's digest-valid boolean.
	var c, donor exploration.SavedCapture
	if json.Unmarshal([]byte(bank[m.Family].StoredBytesBefore), &c) != nil || json.Unmarshal([]byte(bank[m.ForeignFamily].StoredBytesBefore), &donor) != nil {
		return false
	}
	var raw map[string]json.RawMessage
	if json.Unmarshal([]byte(m.StoredBytes), &raw) != nil {
		return false
	}
	var rawID string
	if json.Unmarshal(raw["id"], &rawID) != nil || rawID != m.ID {
		return false
	}
	var title string
	if json.Unmarshal(raw["title"], &title) != nil {
		return false
	}
	c.Title = title
	switch m.Mode {
	case "payload-mixed":
		transplantCapture387(&c, donor, false)
	case "question-mixed":
		transplantCapture387(&c, donor, true)
	case "payload-null", "question-null":
		object := raw
		if m.Mode == "question-null" {
			object = map[string]json.RawMessage{} // A nested decode must not merge into the root map.
			if json.Unmarshal(raw["question"], &object) != nil {
				return false
			}
		}
		if value, ok := object[m.ForeignField]; !ok || string(value) != "null" {
			return false
		}
	case "scope-corrupt":
		switch c.Question.Kind {
		case "explore":
			c.Explore.Question.FromMS++
		case "inventory":
			c.Inventory.Question.FromMS++
		case "measurement":
			c.Measurement.Question.FromMS++
		case "trace-search":
			c.TraceSearch.Request.FromMS++
		case "comparison":
			c.Comparison.Request.Selection.Question.FromMS++
		case "explanation":
			c.Explanation.Request.Selection.FromMS++
		}
	default:
		return false
	}
	c.ID = ""
	canonical, _ := json.Marshal(c)
	sum := sha256.Sum256(canonical)
	if hex.EncodeToString(sum[:]) != m.ID || m.CanonicalSHA256 != m.ID || m.SourceID != bank[m.Family].Page.Receipt.ID {
		return false
	}
	c.ID = m.ID
	expected, _ := json.Marshal(c)
	if m.Mode == "payload-null" {
		delete(raw, m.ForeignField)
	} else if m.Mode == "question-null" {
		var q map[string]json.RawMessage
		if json.Unmarshal(raw["question"], &q) != nil {
			return false
		}
		delete(q, m.ForeignField)
		raw["question"] = envelopeRaw387(q)
	}
	return sameTransport384(envelopeRaw387(raw), json.RawMessage(expected))
}
func admitCaptureLedger387(t *testing.T, r captureLedger387) error {
	t.Helper()
	fail := func(s string) error { return fmt.Errorf("capture387: %s", s) }
	if r.Version != "goo.captureEnvelope387/1" || len(r.Errors) > 0 || len(r.Bank) != 6 || len(r.RawSpans) != 111 || len(r.RawMetrics) != 4 {
		return fail("complete actual family bank absent")
	}
	for _, family := range captureFamilies387 {
		b, ok := r.Bank[family.kind]
		if !ok || b.Page.Receipt.Question.Kind != family.kind || b.StoredBytesBefore == "" || b.StoredBytesBefore != b.StoredBytesAfter {
			return fail("family/immutable bytes changed")
		}
		physical := sha256.Sum256([]byte(b.StoredBytesBefore))
		if hex.EncodeToString(physical[:]) != b.SHA256Before {
			return fail("physical byte digest changed")
		}
		var stored exploration.SavedCapture
		if json.Unmarshal([]byte(b.StoredBytesBefore), &stored) != nil {
			return fail("published stored envelope refused")
		}
		id := stored.ID
		stored.ID = ""
		wire, _ := json.Marshal(stored)
		sum := sha256.Sum256(wire)
		if hex.EncodeToString(sum[:]) != id || id != b.Page.Receipt.ID {
			return fail("canonical artifact digest changed")
		}
		stored.ID = id
		projected, e := stored.Page("")
		if e != nil || !sameTransport384(projected, b.Page) {
			return fail("frozen page no longer belongs to stored bytes")
		}
	}
	search := r.Bank["trace-search"].Page.TraceSearch
	if search == nil || search.MatchedTraces != "2" || len(search.Items) != 2 {
		return fail("qualified trace witnesses absent")
	}
	nullWitnesses := 0
	for _, item := range search.Items {
		for _, w := range item.Witnesses {
			found := false
			for _, row := range r.RawSpans {
				if row.TraceID == item.TraceID && row.SpanID == w.SpanID {
					found = true
					if w.Service != row.ServiceName || w.TimestampNS != row.AtNS || w.DurationNS != row.DurationNS {
						return fail("qualified witness differs from source occurrence")
					}
					for _, v := range w.FieldValues {
						if v.Field == (exploration.Field{Kind: "attribute", Key: r.Fixture.Absent}) {
							if _, present := row.SpanAttributes[r.Fixture.Absent]; present || v.Value != nil {
								return fail("legitimate missing witness was fabricated")
							}
							nullWitnesses++
						}
					}
				}
			}
			if !found {
				return fail("witness source occurrence absent")
			}
		}
	}
	if nullWitnesses != 2 {
		return fail("nested nullable witnesses not retained")
	}
	measurement := r.Bank["measurement"].Page.Measurement
	if measurement == nil || measurement.Gauges == nil || len(measurement.Gauges.Series) != 1 || len(measurement.Gauges.Series[0].Samples) != 4 {
		return fail("native gauge evidence absent")
	}
	mq := measurement.Question
	hi, ok := new(big.Int).SetString(mq.ToNS, 10)
	if !ok {
		return fail("native fixed clock absent")
	}
	for i, row := range r.RawMetrics {
		sample := measurement.Gauges.Series[0].Samples[i]
		state := "observed"
		at, ok := new(big.Int).SetString(row.AtNS, 10)
		if !ok {
			return fail("raw native clock malformed")
		}
		if row.Flags&1 != 0 {
			state = "unknown"
		} else if new(big.Int).Sub(hi, at).Cmp(new(big.Int).Mul(big.NewInt(mq.Definition.MaxGapMS), big.NewInt(1000000))) > 0 {
			state = "stale"
		}
		if sample.AtNS != row.AtNS || sample.State != state {
			return fail("native observation/freshness changed")
		}
		if state == "unknown" {
			if sample.Value != nil {
				return fail("unknown native value invented zero")
			}
		} else {
			value, e := strconv.ParseFloat(row.Value, 64)
			if e != nil || sample.Value == nil || *sample.Value != value {
				return fail("recorded native zero/value lost")
			}
		}
	}
	explore := r.Bank["explore"].Page
	wantIDs := []string{}
	sum := new(big.Int)
	errors := 0
	for _, row := range r.RawSpans {
		if row.ServiceName == r.Fixture.API && row.SpanAttributes["cohort"] == "selected" && row.ResourceAttributes["service.namespace"] == r.Fixture.Namespace && row.ResourceAttributes["deployment.environment.name"] == r.Fixture.Environment {
			wantIDs = append(wantIDs, row.SpanID)
			n, ok := new(big.Int).SetString(row.DurationNS, 10)
			if !ok {
				return fail("raw duration invalid")
			}
			sum.Add(sum, n)
			if row.StatusCode == "Error" {
				errors++
			}
		}
	}
	slices.Sort(wantIDs)
	if len(wantIDs) != 105 || explore.Explore == nil || explore.Explore.Count != "105" || explore.RetainedRecords != 105 || len(explore.Explore.Rows) != 100 || explore.Explore.Duration == nil || explore.Explore.Duration.SumNS != sum.String() || explore.Explore.SpanStatus == nil || explore.Explore.SpanStatus.Error != strconv.Itoa(errors) {
		return fail("full counts/duration/status changed")
	}
	first, e := reportedResultIDs353(*explore.Explore)
	if e != nil {
		return e
	}
	if r.Pagination.Explore == nil {
		return fail("second page absent")
	}
	second, e := reportedResultIDs353(*r.Pagination.Explore)
	if e != nil {
		return e
	}
	got := append(first, second...)
	slices.Sort(got)
	if !slices.Equal(got, wantIDs) || r.Pagination.Receipt.ID != explore.Receipt.ID || r.Pagination.Explore.Question.Offset != 100 || r.Pagination.Explore.Count != "105" {
		return fail("page boundary redefined population")
	}
	empty := r.EmptyExplore
	if empty.Page.Explore == nil || empty.Page.Explore.Count != "0" || empty.Page.RetainedRecords != 0 || len(empty.Page.Explore.Rows) != 0 || empty.StoredBytesBefore == "" || empty.StoredBytesBefore != empty.StoredBytesAfter {
		return fail("empty capture invented data or changed bytes")
	}
	if r.SourceRefusalProbe.Status != 502 || r.ExplanationPromotion.SQLBefore != r.ExplanationPromotion.SQLAfter {
		return fail("offline promotion reread source")
	}
	var observed map[string]json.RawMessage
	if json.Unmarshal(r.ObservedExplanation, &observed) != nil {
		return fail("held explanation absent")
	}
	delete(observed, "resultToken")
	if !sameTransport384(observed, r.Bank["explanation"].Page.Explanation) {
		return fail("promotion changed held explanation")
	}
	if len(r.FileMutants) != 30 || len(r.FileUI) != 6 || len(r.TransportMutants) != 18 || len(r.OfflineReopens) != 6 {
		return fail("negative campaign missing")
	}
	seen := map[string]bool{}
	for _, m := range r.FileMutants {
		key := m.Family + "/" + m.Mode
		if seen[key] || m.Status < 400 || m.Status >= 500 || !captureMutantDigest387(m, r.Bank) {
			return fail(fmt.Sprintf("malformed attack %s/%s lacks valid canonical hash or controlled refusal (status %d)", m.Family, m.Mode, m.Status))
		}
		seen[key] = true
		dir := t.TempDir()
		if e := os.WriteFile(filepath.Join(dir, m.ID+".json"), []byte(m.StoredBytes), 0600); e != nil {
			return e
		}
		if _, e := (exploration.CaptureStore{Dir: dir}).Load(m.ID); e == nil {
			return fail("rehashed malformed artifact admitted offline")
		}
	}
	for _, held := range append(slices.Clone(r.FileUI), r.TransportMutants...) {
		if !sameTransport384(held.Before, held.After) || !strings.Contains(strings.ToLower(held.Notice), "unavailable") {
			return fail("invalid capture displaced held evidence")
		}
	}
	for _, m := range r.TransportMutants {
		if m.OriginalStatus != 200 || !sameTransport384(m.OriginalPage, r.Bank[m.Family].Page) {
			return fail("negative transport lacks successful original")
		}
		var page exploration.CapturePage
		if json.Unmarshal(m.CorruptedPage, &page) == nil {
			return fail("invalid transport became full evidence")
		}
	}
	wrong := r.WrongPageTransport
	requested, e := url.Parse(wrong.RequestURL)
	if e != nil {
		return fail("requested stored location invalid")
	}
	cursor := requested.Query().Get("cursor")
	decoded, e := base64.RawURLEncoding.DecodeString(cursor)
	if e != nil || requested.Path != "/api/captures/"+explore.Receipt.ID || string(decoded) != explore.Receipt.ID+":100" || wrong.Status != 200 || !sameTransport384(wrong.Before, wrong.After) || !strings.Contains(strings.ToLower(wrong.Notice), "unavailable") {
		return fail("wrong-page request lost accepted location")
	}
	var original, returned exploration.CapturePage
	if json.Unmarshal(wrong.OriginalPage, &original) != nil || json.Unmarshal(wrong.ReturnedPage, &returned) != nil || original.Explore == nil || returned.Explore == nil || original.Explore.Question.Offset != 100 || returned.Explore.Question.Offset != 0 || original.Receipt.ID != explore.Receipt.ID || returned.Receipt.ID != explore.Receipt.ID || !sameTransport384(returned, explore) {
		return fail("wrong-page control is not valid DTO under wrong request context")
	}
	if r.UIPaging.Forward.Status != 200 || r.UIPaging.Backward.Status != 200 || !sameTransport384(r.UIPaging.Forward.Response, r.Pagination) || !sameTransport384(r.UIPaging.Backward.Response, explore) {
		return fail("stored page retry/return changed retained evidence")
	}
	for index, entry := range []struct {
		raw    string
		offset string
	}{{r.UIPaging.Forward.RequestURL, "100"}, {r.UIPaging.Backward.RequestURL, "0"}} {
		u, e := url.Parse(entry.raw)
		if e != nil || u.Path != "/api/captures/"+explore.Receipt.ID {
			return fail("paging request names wrong capture")
		}
		token := u.Query().Get("cursor")
		if index == 1 && token == "" {
			continue
		}
		decoded, e := base64.RawURLEncoding.DecodeString(token)
		if e != nil || string(decoded) != explore.Receipt.ID+":"+entry.offset {
			return fail("paging request names wrong offset")
		}
	}
	if len(r.Document.ReceiptCalls) != 6 {
		return fail("receipt-only references absent")
	}
	ids := map[string]bool{}
	for _, call := range r.Document.ReceiptCalls {
		if call.Status != 200 || !strings.HasSuffix(call.Path, "/receipt") {
			return fail("wrong reference capability")
		}
		var meta map[string]json.RawMessage
		if json.Unmarshal(call.Response, &meta) != nil {
			return fail("metadata invalid")
		}
		for _, arm := range captureFamilies387 {
			if _, ok := meta[arm.arm]; ok {
				return fail("metadata claims payload authority")
			}
		}
		for _, key := range []string{"retainedRecords", "recordsLimited", "nextCursor", "previousCursor"} {
			if _, ok := meta[key]; ok {
				return fail("metadata claims page authority")
			}
		}
		var receipt exploration.CaptureReceipt
		if json.Unmarshal(meta["receipt"], &receipt) != nil || !sameTransport384(receipt, r.Bank[receipt.Question.Kind].Page.Receipt) {
			return fail("metadata question/receipt drift")
		}
		if ids[receipt.ID] {
			return fail("duplicate reference instead of six families")
		}
		ids[receipt.ID] = true
		var full exploration.CapturePage
		if json.Unmarshal(call.Response, &full) == nil {
			return fail("metadata masquerades as full page")
		}
	}
	if !sameTransport384(r.HeldParent.Before, r.HeldParent.After) || r.Child.Status != 200 {
		return fail("held search parent displaced")
	}
	child := r.Child.Request
	if child.Signal != "spans" || child.FromMS != r.Fixture.Selection.FromMS || child.ToMS != r.Fixture.Selection.ToMS || !reflect.DeepEqual(child.EnvironmentSelector, r.Fixture.Selection.EnvironmentSelector) || child.Where == nil || len(child.Where.All) != 1 || child.Where.All[0].Field == nil || *child.Where.All[0].Field != (exploration.Field{Kind: "column", Key: "trace_id"}) || child.Where.All[0].Op != "eq" {
		return fail("fresh trace changed scope or retained population condition")
	}
	want := []string{}
	for _, row := range r.RawSpans {
		if row.TraceID == child.Where.All[0].Value {
			want = append(want, row.SpanID)
		}
	}
	slices.Sort(want)
	actual, e := reportedResultIDs353(r.Child.Response)
	if e != nil || !slices.Equal(actual, want) || r.Child.Response.Count != strconv.Itoa(len(want)) || !reflect.DeepEqual(r.Child.Response.Question, child) {
		return fail("fresh trace members differ from raw bank")
	}
	if len(r.Facts) != 39 {
		return fail("measured journey incomplete")
	}
	readFacts, referenceFacts := 0, 0
	factNames := map[string]bool{}
	for i, f := range r.Facts {
		if factNames[f.Name] {
			return fail("duplicate measured gesture")
		}
		factNames[f.Name] = true
		n := 0
		stores, refs := 0, 0
		switch {
		case f.Name == "explicit-fresh-trace-read":
			n = 1
			readFacts++
		case f.Name == "metadata-only-six-references":
			refs = 6
			referenceFacts++
		case strings.HasPrefix(f.Name, "offline-open-"), strings.HasPrefix(f.Name, "file-refusal-preserves-"), strings.HasPrefix(f.Name, "invalid-transport-"), f.Name == "offline-parent-reopen", f.Name == "offline-empty-explore", f.Name == "wrong-page-transport-refusal", f.Name == "stored-next-page-retry", f.Name == "stored-previous-page":
			stores = 1
		case f.Name == "stage-fresh-trace", f.Name == "back-to-held-search":
		default:
			return fail("unknown gesture " + f.Name)
		}
		if f.After.Source-f.Before.Source != n || f.After.StoreReads-f.Before.StoreReads != stores || f.After.Receipts-f.Before.Receipts != refs || f.After.Captures != 0 || f.Before.Captures != 0 || f.After.Documents != f.Before.Documents {
			return fail("gesture acquired implicit source/store/write")
		}
		if i > 0 && f.Before != r.Facts[i-1].After {
			return fail("counter discontinuity")
		}
		if n == 0 && f.After.SQL != f.Before.SQL || n > 0 && f.After.SQL <= f.Before.SQL {
			return fail("physical telemetry read budget changed")
		}
	}
	if readFacts != 1 || referenceFacts != 1 {
		return fail("missing/duplicate measured read or reference operation")
	}
	return nil
}
func replayCapture387(t *testing.T, label string, facts []captureFact387, mutate bool) bool {
	t.Helper()
	sink := &qobs.MemorySink{}
	rec := qobs.NewRecorder(sink)
	base := facts[0].Before.Source
	for _, f := range facts {
		action := "history"
		if f.Name == "explicit-fresh-trace-read" {
			action = "read"
		}
		reads := f.After.Source - base
		if mutate && f.Name == "metadata-only-six-references" {
			reads++
		}
		rec.Record(context.Background(), action, nil, "obs.reads", reads, "obs.captures", f.After.Captures)
	}
	wire, _ := json.Marshal(sink.Steps())
	if e := artifact("capture-envelope387-"+label+"-steps.json", wire); e != nil {
		t.Fatal(e)
	}
	checker, e := validate.NewChecker("model/captured_trace_search_383.yaml", validate.Options{Quint: quintPath(t), Backend: "typescript", Dir: t.TempDir()})
	if e != nil {
		t.Fatal(e)
	}
	result, e := checker.Check(t.Context(), sink.Steps())
	if e != nil {
		t.Fatal(e)
	}
	if e = artifact("capture-envelope387-"+label+"-model.txt", []byte(result.Summary())); e != nil {
		t.Fatal(e)
	}
	return result.OK()
}
func TestQuintCaptureEnvelopeObserved387(t *testing.T) {
	path := os.Getenv("GOO_CAPTURE_ENVELOPE_387_REPORT")
	if path == "" {
		t.Skip("actual all-family capture journey required")
	}
	raw, e := os.ReadFile(path)
	if e != nil {
		t.Fatal(e)
	}
	var r captureLedger387
	if e = json.Unmarshal(raw, &r); e != nil {
		t.Fatal(e)
	}
	if e = admitCaptureLedger387(t, r); e != nil {
		t.Fatal(e)
	}
	if !replayCapture387(t, "actual", r.Facts, false) {
		t.Fatal("actual measured facts violate model")
	}

	mutations := []func(*captureLedger387){
		func(r *captureLedger387) {
			b := r.Bank["inventory"]
			b.StoredBytesAfter += " "
			r.Bank["inventory"] = b
		},
		func(r *captureLedger387) { r.FileMutants[0].CanonicalSHA256 = strings.Repeat("a", 64) },
		func(r *captureLedger387) { r.FileMutants[0].Status = 200 },
		func(r *captureLedger387) { r.FileUI[0].After = json.RawMessage(`{"changed":true}`) },
		func(r *captureLedger387) { r.TransportMutants[0].CorruptedPage = r.TransportMutants[0].OriginalPage },
		func(r *captureLedger387) { r.HeldParent.After = json.RawMessage(`{"changed":true}`) },
		func(r *captureLedger387) { r.Child.Request.FromMS++ },
		func(r *captureLedger387) { r.ExplanationPromotion.SQLAfter++ },
		func(r *captureLedger387) { r.Facts[0].After.Source++ },
		func(r *captureLedger387) { r.Bank["explore"].Page.Explore.Count = "100" },
		func(r *captureLedger387) {
			r.Document.ReceiptCalls[0].Response = envelopeRaw387(r.Bank["explore"].Page)
		},
	}

	mutations = append(mutations,
		func(r *captureLedger387) { r.WrongPageTransport.ReturnedPage = r.WrongPageTransport.OriginalPage },
		func(r *captureLedger387) { r.WrongPageTransport.After = json.RawMessage(`{"location":{"offset":100}}`) },
	)
	for i, mutate := range mutations {
		var bad captureLedger387
		json.Unmarshal(raw, &bad)
		mutate(&bad)
		if admitCaptureLedger387(t, bad) == nil {
			t.Fatalf("semantic mutant%d admitted", i)
		}
	}
	if replayCapture387(t, "reference-read-mutant", r.Facts, true) {
		t.Fatal("reference telemetry read escaped model")
	}
}
