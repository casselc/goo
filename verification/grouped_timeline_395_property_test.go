package verification

import (
	"bytes"
	"cmp"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/identity"
	"github.com/casselc/goo/workbench"
	"hegel.dev/go/hegel"
)

type groupedVector395 struct {
	Data        *exploration.Result               `json:"data"`
	Destination exploration.Query                 `json:"destination"`
	Declaration workbench.TimelineViewDeclaration `json:"declaration"`
}

// These generated HTTP replies exercise public source decoding/projection. They
// are synthetic contract evidence; the observed journey uses actual OTel data.
func groupedFixture395(t testing.TB, seed int, signal string, changed bool) *exploration.Result {
	t.Helper()
	q := exploration.Query{Signal: signal, Environment: "fixture", FromMS: int64(1000 + seed), ToMS: int64(4500 + seed), Limit: 2, BucketMS: 1000, GroupTimeSeries: true, GroupBy: []exploration.Field{{Kind: "attribute", Key: "region"}}, EnvironmentSelector: &identity.EnvironmentSelector{Version: identity.EnvironmentSelectorVersion, Policy: "effective", Scope: identity.EnvironmentScope{Mode: "named", Value: "fixture"}}}
	q.Environment = ""
	plan, err := exploration.Compile(q)
	if err != nil {
		t.Fatal(err)
	}
	empty, unicode := "", "雪<>&\u2028"
	categories := []*string{nil, &empty, &unicode}
	groups, split, rows := []string{}, []string{}, []string{}
	totals := make([]int, 4)
	total := 0
	for group, category := range categories {
		n := 1 + (seed+group)%3
		g, _ := json.Marshal(map[string]any{"g0": category, "count": strconv.Itoa(n + 1)})
		groups = append(groups, string(g))
		for _, cell := range []struct{ bucket, count int }{{group, n}, {3, 1}} {
			from := q.FromMS + int64(cell.bucket)*q.BucketMS
			totals[cell.bucket] += cell.count
			total += cell.count
			b, _ := json.Marshal(map[string]any{"g0": category, "from_ms": strconv.FormatInt(from, 10), "count": strconv.Itoa(cell.count)})
			split = append(split, string(b))
			for index := 0; index < cell.count; index++ {
				attributes := map[string]string{}
				if category != nil {
					attributes["region"] = *category
				}
				raw, _ := json.Marshal(map[string]any{"timestamp_ns": strconv.FormatInt((from+1)*1000000+int64(index), 10), "duration_ns": "0", "ServiceName": "fixture-worker", "SpanId": fmt.Sprintf("%016x", len(rows)+1), "TraceId": fmt.Sprintf("%032x", len(rows)+1), "SpanName": "operation", "SpanKind": "Internal", "Body": "record", "SeverityText": "INFO", "ResourceAttributes": map[string]string{"service.name": "fixture-worker", "deployment.environment.name": "fixture"}, "Attributes": attributes})
				rows = append(rows, string(raw))
			}
		}
	}
	slices.SortFunc(rows, func(a, b string) int {
		var left, right struct {
			At string `json:"timestamp_ns"`
			ID string `json:"SpanId"`
		}
		json.Unmarshal([]byte(a), &left)
		json.Unmarshal([]byte(b), &right)
		x, _ := strconv.ParseInt(left.At, 10, 64)
		y, _ := strconv.ParseInt(right.At, 10, 64)
		if order := cmp.Compare(x, y); order != 0 {
			return order
		}
		return strings.Compare(left.ID, right.ID)
	})
	timeRows := []string{}
	for bucket, n := range totals {
		if n > 0 {
			b, _ := json.Marshal(map[string]string{"from_ms": strconv.FormatInt(q.FromMS+int64(bucket)*q.BucketMS, 10), "count": strconv.Itoa(n)})
			timeRows = append(timeRows, string(b))
		}
	}
	count := total
	if changed {
		count++
	}
	transport := snapshotTransport(func(req *http.Request) (*http.Response, error) {
		wire, _ := io.ReadAll(req.Body)
		body := ""
		switch string(wire) {
		case plan.Count:
			body = fmt.Sprintf(`{"count":"%d"}`, count)
		case plan.Rows:
			body = strings.Join(rows[:min(len(rows), q.Limit+1)], "\n")
		case plan.Groups:
			body = strings.Join(groups, "\n")
		case plan.TimeSeries:
			body = strings.Join(timeRows, "\n")
		case plan.GroupedTimeSeries:
			body = strings.Join(split, "\n")
		case plan.Fields:
		default:
			return nil, fmt.Errorf("unexpected generated source query")
		}
		return &http.Response{StatusCode: 200, Header: http.Header{}, Body: io.NopCloser(strings.NewReader(body))}, nil
	})
	result, err := (exploration.Source{Endpoint: "http://fixture.invalid", Client: &http.Client{Transport: transport}}).Fetch(t.Context(), q)
	if err != nil {
		t.Fatal(err)
	}
	d := result.GroupedTimeDataset
	if d == nil || len(d.Rows) != 12 {
		t.Fatal("grouped projection missing", result)
	}
	// Reconstruct every count independently of the supplied projection and its
	// decoder. Null and empty groups get different typed tuple keys.
	for _, row := range d.Rows {
		category := row.Values["category"]
		group := -1
		for i, candidate := range categories {
			if reflect.DeepEqual(candidate, category) {
				group = i
			}
		}
		if group < 0 {
			t.Fatal("invented category")
		}
		from, _ := strconv.ParseInt(*row.Values["fromMs"], 10, 64)
		to, _ := strconv.ParseInt(*row.Values["toMs"], 10, 64)
		bucket := int((from - q.FromMS) / q.BucketMS)
		if to != min(from+q.BucketMS, q.ToMS) {
			t.Fatal("final interval changed")
		}
		want := 0
		if bucket == group {
			want = 1 + (seed+group)%3
		}
		if bucket == 3 {
			want = 1
		}
		if want == 0 && changed {
			if row.State != "unknown" || row.Values["count"] != nil {
				t.Fatal("disagreement invented zero")
			}
		} else if row.Values["count"] == nil || *row.Values["count"] != strconv.Itoa(want) || row.State != map[bool]string{true: "observed", false: "no_data"}[want > 0] {
			t.Fatal("cell differs from raw generator")
		}
		identity, _ := json.Marshal([]any{category, *row.Values["fromMs"], *row.Values["toMs"]})
		digest := sha256.Sum256(identity)
		if row.Key != "group-interval:"+hex.EncodeToString(digest[:]) {
			t.Fatal("typed category/interval identity changed")
		}
	}
	if changed && (d.Completeness != nil || d.Quality.Extent != "unknown") || !changed && (d.Completeness == nil || d.Quality.Extent != "complete") {
		t.Fatal("coverage disagrees with independent reads")
	}
	return result
}

func TestHegelGroupedTimelineContract395(t *testing.T) {
	vectors := []groupedVector395{}
	hegel.Test(t, func(tc *hegel.T) {
		seed := hegel.Draw(tc, hegel.Integers(1, 100000))
		for _, signal := range []string{"logs", "spans"} {
			for _, changed := range []bool{false, true} {
				data := groupedFixture395(t, seed, signal, changed)
				raw, _ := json.Marshal(data.Question)
				var destination exploration.Query
				json.Unmarshal(raw, &destination)
				destination.FromMS += 60000
				destination.ToMS += 61000
				destination.Limit = 1 + seed%5
				destination.EnvironmentSelector.Scope.Value = "destination"
				destination.Where = &exploration.Predicate{Field: &exploration.Field{Kind: "column", Key: "service"}, Op: "eq", Value: "destination-worker"}
				declaration := workbench.TimelineViewDeclaration{Version: "goo.view/1", Scope: workbench.ViewDeclarationScope{Kind: "record-query", Signal: signal}, Kind: workbench.TimelineViewKind{Name: []string{"small-multiples", "table", "heatmap"}[seed%3], Version: 1}, Display: workbench.TimelineViewDisplay{Profile: "grouped-count-v1", Value: "count", Order: []string{"source", "label"}[seed%2], Query: []string{"", "雪"}[seed%2]}}
				if err := declaration.Validate(data.Question); err != nil {
					tc.Fatal(err)
				}
				reused, err := workbench.ReuseTimelineViewDeclaration(declaration, data.Question, destination)
				if err != nil || reused != declaration {
					tc.Fatal("valid reuse refused", err)
				}
				vectors = append(vectors, groupedVector395{data, destination, declaration})
			}
		}
	}, options(8)...)
	input, _ := json.Marshal(vectors)
	cmd := exec.CommandContext(t.Context(), "node", "--input-type=module", "-e", groupedJS395)
	cmd.Stdin = bytes.NewReader(input)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("shared grouped contract: %v %s", err, stderr.String())
	}
	var views []workbench.ExplorationView
	if err = json.Unmarshal(out, &views); err != nil || len(views) != len(vectors) {
		t.Fatal("invalid shared view batch", err)
	}
	store := workbench.Store{Dir: t.TempDir()}
	for index, view := range views {
		if !reflect.DeepEqual(view.Queries, []exploration.Query{vectors[index].Data.Question}) {
			t.Fatal("composition replaced query")
		}
		saved, err := store.CreateExplorationView(view)
		if err != nil {
			t.Fatal(err)
		}
		loaded, err := store.LoadExplorationView(saved.ID)
		if err != nil || !reflect.DeepEqual(saved, loaded) {
			t.Fatal("immutable view changed", err)
		}
	}
	t.Logf("%d grouped Go/JS source projections, exact identities, reuse and store roundtrips", len(vectors))
}

const groupedJS395 = `import fs from 'node:fs';import assert from 'node:assert/strict';import * as T from '../ui/timeline-declaration.mjs';import * as C from '../ui/view-composition.mjs';import {validateSnapshot} from '../ui/view-model.mjs';import {adoptExploration} from '../ui/source-snapshots.mjs';import {hasGuarantee} from '../ui/snapshots.mjs';
globalThis.fetch=()=>{throw Error('local grouped configuration read telemetry')};const out=[];for(const v of JSON.parse(fs.readFileSync(0,'utf8'))){const before=JSON.stringify(v),q=v.data.question;const data=adoptExploration(v.data,q);const d=data.groupedTimeDataset;assert.equal(validateSnapshot(d).ok,true);assert.equal(hasGuarantee(d,'grouped-intervals'),v.data.state!=='changed_between_reads');assert.equal(d.partition,undefined);if(v.data.state==='observed')assert.throws(()=>adoptExploration({...v.data,state:'changed_between_reads'},q));assert.deepEqual(T.reuse(v.declaration,q,v.destination),v.declaration);for(const edit of [x=>x.bucketMs=10000,x=>x.groupBy[0].kind='resource',x=>x.groupBy[0].key='other',x=>x.groupTimeSeries=false,x=>x.summary='span-duration']){const target=structuredClone(v.destination);edit(target);assert.throws(()=>T.reuse(v.declaration,q,target));}for(const edit of [x=>x.display.profile='',x=>x.display.value='duration.mean',x=>x.kind.name='time-bar',x=>x.display.order='value-desc']){const declaration=structuredClone(v.declaration);edit(declaration);assert.throws(()=>T.validate(declaration,q));}const tooWide={...q,toMs:q.fromMs+33000};assert.throws(()=>T.validate({...v.declaration,kind:{name:'heatmap',version:1}},tooWide));const bad=structuredClone(v.data.groupedTimeDataset);bad.rows[0].values.category='invented';assert.equal(validateSnapshot(bad).ok,false);const view={version:1,title:'Grouped counts',queries:[q],composition:{version:'goo.composition/1',kind:{name:'stack',version:1},children:[{timeline:v.declaration},{records:{version:'goo.view/1',scope:{kind:'record-query',signal:q.signal},kind:{name:'record-table',version:1},display:{columns:['utc','summary']}}}]}};C.view(view);assert.equal(JSON.stringify(v),before);out.push(view)}process.stdout.write(JSON.stringify(out));`

func TestGroupedTimelineWireParity395(t *testing.T) {
	q := exploration.Query{Signal: "logs", Environment: "prod", FromMS: 1000, ToMS: 4500, Limit: 3, BucketMS: 1000, GroupTimeSeries: true, GroupBy: []exploration.Field{{Kind: "attribute", Key: "region"}}}
	type vector struct {
		Name        string            `json:"name"`
		Question    exploration.Query `json:"question"`
		Declaration json.RawMessage   `json:"declaration"`
		Accepted    bool              `json:"accepted"`
	}
	base := `{"version":"goo.view/1","scope":{"kind":"record-query","signal":"logs"},"kind":{"name":"table","version":1},"display":{"profile":"grouped-count-v1","value":"count"}}`
	vectors := []vector{}
	for _, change := range []struct {
		name, key string
		value     any
	}{{"omitted", "", nil}, {"query-null", "query", nil}, {"query-empty", "query", ""}, {"order-null", "order", nil}, {"order-empty", "order", ""}, {"profile-null", "profile", nil}, {"profile-empty", "profile", ""}, {"wrong-profile", "profile", "grouped-count-v2"}, {"query-number", "query", 3}, {"order-number", "order", 3}, {"query-control", "query", "a\nb"}, {"query-overbound", "query", strings.Repeat("雪", 86)}, {"unknown-order", "order", "value-desc"}} {
		var wire map[string]any
		json.Unmarshal([]byte(base), &wire)
		if change.key != "" {
			wire["display"].(map[string]any)[change.key] = change.value
		}
		raw, _ := json.Marshal(wire)
		var d workbench.TimelineViewDeclaration
		err := json.Unmarshal(raw, &d)
		accepted := err == nil && d.Validate(q) == nil
		vectors = append(vectors, vector{change.name, q, raw, accepted})
	}
	// Same label in a different qualified field is incompatible, not an alias.
	input, _ := json.Marshal(vectors)
	cmd := exec.CommandContext(t.Context(), "node", "--input-type=module", "-e", `import fs from 'node:fs';import assert from 'node:assert/strict';import {validate} from '../ui/timeline-declaration.mjs';let checked=0;for(const v of JSON.parse(fs.readFileSync(0,'utf8'))){let accepted=true;try{validate(v.declaration,v.question)}catch{accepted=false}assert.equal(accepted,v.accepted,v.name+' Go/JS admission differs');checked++}process.stdout.write(String(checked));`)
	cmd.Stdin = bytes.NewReader(input)
	if wire, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("grouped declaration wire parity: %v %s", err, wire)
	}
}
