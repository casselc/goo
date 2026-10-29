// Package capturehttp owns saved captures, observed promotion and live-comparison drafts.
package capturehttp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"path/filepath"
	"reflect"
	"strconv"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/internal/httpjson"
)

// Register binds durable storage, a source and the owning server's result bank.
func Register(mux *http.ServeMux, source exploration.Source, dir string, resultSnapshots *exploration.ResultSnapshots) {
	store := exploration.CaptureStore{Dir: filepath.Join(dir, "exploration")}
	registerLiveComparison(mux, source, store)
	reply := func(w http.ResponseWriter, code int, value any) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "no-store")
		w.WriteHeader(code)
		json.NewEncoder(w).Encode(value)
	}
	fail := func(w http.ResponseWriter, code int, e error) { reply(w, code, map[string]string{"error": e.Error()}) }
	cursor := func(r *http.Request) (string, error) {
		v := r.URL.Query()
		for key, values := range v {
			if key != "cursor" || len(values) != 1 {
				return "", fmt.Errorf("only one cursor parameter is supported")
			}
		}
		return v.Get("cursor"), nil
	}
	mux.HandleFunc("GET /captures/search", func(w http.ResponseWriter, r *http.Request) {
		values, parseErr := url.ParseQuery(r.URL.RawQuery)
		if parseErr != nil {
			fail(w, 400, errors.New("invalid capture search query encoding"))
			return
		}
		for key, v := range values {
			if len(v) != 1 || (key != "title" && key != "kind" && key != "signal" && key != "environment" && key != "order" && key != "cursor") {
				fail(w, 400, errors.New("one value per supported capture search parameter required"))
				return
			}
		}
		q := exploration.CaptureSearchQuery{Title: values.Get("title"), Kind: values.Get("kind"), Signal: values.Get("signal"), Order: values.Get("order")}
		if values.Has("environment") {
			env := values.Get("environment")
			q.Environment = &env
		}
		if err := q.Validate(); err != nil {
			fail(w, 400, err)
			return
		}
		page, err := store.Search(r.Context(), q, values.Get("cursor"))
		if err != nil {
			code := 500
			if errors.Is(err, exploration.ErrCaptureSearchCursor) {
				code = 400
			}
			if errors.Is(err, exploration.ErrCaptureSearchBudget) || errors.Is(err, exploration.ErrCaptureCatalogUnavailable) {
				code = 503
			}
			fail(w, code, err)
			return
		}
		reply(w, 200, page)
	})
	mux.HandleFunc("GET /captures", func(w http.ResponseWriter, r *http.Request) {
		after, e := cursor(r)
		if e != nil {
			fail(w, 400, e)
			return
		}
		result, e := store.List(after)
		if e != nil {
			fail(w, 500, e)
			return
		}
		reply(w, 200, result)
	})
	mux.HandleFunc("GET /captures/{id}/followups", func(w http.ResponseWriter, r *http.Request) {
		row := 0
		for key, values := range r.URL.Query() {
			if key != "row" || len(values) != 1 {
				fail(w, 400, fmt.Errorf("only one retained row index supported"))
				return
			}
			n, e := strconv.Atoi(values[0])
			if e != nil || strconv.Itoa(n) != values[0] {
				fail(w, 400, fmt.Errorf("canonical row index required"))
				return
			}
			row = n
		}
		c, e := store.Load(r.PathValue("id"))
		if e != nil {
			fail(w, 404, e)
			return
		}
		result, e := c.Followups(row)
		if e != nil {
			fail(w, 400, e)
			return
		}
		reply(w, 200, result)
	})
	mux.HandleFunc("GET /captures/{id}/compare", func(w http.ResponseWriter, r *http.Request) {
		values := r.URL.Query()
		if (len(values) != 1 && len(values) != 2) || len(values["selection"]) != 1 || (len(values) == 2 && (len(values["distribution"]) != 1 || values.Get("distribution") != "true")) {
			fail(w, 400, fmt.Errorf("one selection capture ID required"))
			return
		}
		a, e := store.Load(r.PathValue("id"))
		if e != nil {
			fail(w, 404, e)
			return
		}
		b, e := store.Load(values.Get("selection"))
		if e != nil {
			fail(w, 404, e)
			return
		}
		var result *exploration.EvidenceComparison
		if values.Get("distribution") == "true" {
			result, e = exploration.CompareDistributionEvidence(a, b)
		} else {
			result, e = exploration.CompareEvidence(a, b)
		}
		if e != nil {
			fail(w, 400, e)
			return
		}
		reply(w, 200, result)
	})
	mux.HandleFunc("GET /captures/{id}/windows", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.RawQuery != "" {
			fail(w, 400, fmt.Errorf("window discovery has no query parameters"))
			return
		}
		c, e := store.Load(r.PathValue("id"))
		if e != nil {
			fail(w, 404, e)
			return
		}
		result, e := exploration.DiscoverCapturedWindows(c)
		if e != nil {
			fail(w, 400, e)
			return
		}
		reply(w, 200, result)
	})
	mux.HandleFunc("GET /captures/{id}/receipt", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.RawQuery != "" {
			fail(w, 400, fmt.Errorf("receipt metadata has no query parameters"))
			return
		}
		c, e := store.Load(r.PathValue("id"))
		if e != nil {
			fail(w, 404, e)
			return
		}
		readAt := ""
		if c.Explore != nil {
			readAt = c.Explore.CapturedAt
		}
		revision := ""
		if c.Measurement != nil {
			revision = c.Measurement.Plan.Revision
			readAt = c.Measurement.ReadAt
		}
		summary := map[string]string{}
		if c.Explore != nil {
			summary["state"] = c.Explore.State
			summary["count"] = c.Explore.Count
			summary["grain"] = c.Explore.Question.Signal + " records"
		}
		if c.Measurement != nil {
			summary["state"] = c.Measurement.State
			summary["count"] = c.Measurement.Count
			summary["grain"] = "source records; " + c.Measurement.Question.Definition.Grain + " measurement"
		}
		reply(w, 200, map[string]any{"receipt": c.CaptureReceipt, "definitionRevision": revision, "readAt": readAt, "summary": summary})
	})
	mux.HandleFunc("GET /captures/{id}", func(w http.ResponseWriter, r *http.Request) {
		cur, e := cursor(r)
		if e != nil {
			fail(w, 400, e)
			return
		}
		c, e := store.Load(r.PathValue("id"))
		if e != nil {
			fail(w, 404, e)
			return
		}
		p, e := c.Page(cur)
		if e != nil {
			fail(w, 400, e)
			return
		}
		reply(w, 200, p)
	})
	mux.HandleFunc("POST /captures", func(w http.ResponseWriter, r *http.Request) {
		var in struct {
			Title       string                      `json:"title"`
			ResultToken string                      `json:"resultToken,omitempty"`
			Question    exploration.CaptureQuestion `json:"question"`
			Parent      string                      `json:"parent"`
		}
		policy, _ := PolicyFor(r.Method, r.URL.Path)
		if e := httpjson.Decode(w, r, &in, policy.BodyLimit, policy.TrailingMessage); e != nil {
			fail(w, 400, e)
			return
		}
		if in.ResultToken != "" {
			if e := exploration.ValidateCaptureTitle(in.Title); e != nil {
				fail(w, 400, e)
				return
			}
			if in.Parent != "" {
				fail(w, 400, fmt.Errorf("observed explanation promotion cannot refresh a parent"))
				return
			}
			observed, e := resultSnapshots.Load(in.ResultToken)
			if e != nil {
				fail(w, 409, e)
				return
			}
			if observed.Question.Kind != "explanation" || observed.Explanation == nil {
				fail(w, 400, fmt.Errorf("result token does not identify an observed explanation"))
				return
			}
			if in.Question.Kind != "" {
				if e = in.Question.Validate(); e != nil {
					fail(w, 400, e)
					return
				}
				a, _ := json.Marshal(in.Question.Normalized())
				b, _ := json.Marshal(observed.Question.Normalized())
				if string(a) != string(b) {
					fail(w, 409, fmt.Errorf("observed explanation question differs from supplied question"))
					return
				}
			} else if in.Question.Explanation != nil || in.Question.Explore != nil || in.Question.Measurement != nil || in.Question.Inventory != nil || in.Question.Comparison != nil {
				fail(w, 400, fmt.Errorf("partial promotion question is invalid"))
				return
			}
			saved, e := resultSnapshots.PromoteExplanation(in.ResultToken, in.Title, store)
			if e != nil {
				fail(w, 409, e)
				return
			}
			page, e := saved.Page("")
			if e != nil {
				fail(w, 500, e)
				return
			}
			reply(w, 201, page)
			return
		}
		if in.Parent != "" {
			p, e := store.Load(in.Parent)
			if e != nil {
				fail(w, 404, e)
				return
			}
			if in.Question.Kind != "" {
				if e := in.Question.Validate(); e != nil {
					fail(w, 400, e)
					return
				}
			}
			if in.Question.Kind != "" && !reflect.DeepEqual(in.Question.Normalized(), p.Question) {
				fail(w, 409, fmt.Errorf("refresh cannot change the captured question"))
				return
			}
			if in.Question.Kind == "" && (in.Question.Explanation != nil || in.Question.Explore != nil || in.Question.Inventory != nil || in.Question.Comparison != nil || in.Question.Measurement != nil) {
				fail(w, 400, fmt.Errorf("partial refresh question is invalid"))
				return
			}
			in.Question = p.Question
			if in.Title == "" {
				in.Title = p.Title
			}
		}
		if e := exploration.ValidateCaptureTitle(in.Title); e != nil {
			fail(w, 400, e)
			return
		}
		if e := in.Question.Validate(); e != nil {
			fail(w, 400, e)
			return
		}
		if in.Question.Kind == "explanation" {
			fail(w, 400, fmt.Errorf("save an observed explanation with its result token; run again explicitly to refresh"))
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), policy.Timeout)
		defer cancel()
		c, e := store.Create(ctx, source, in.Title, in.Question, in.Parent)
		if e != nil {
			if errors.Is(e, exploration.ErrCapturePlanChanged) {
				fail(w, 409, e)
			} else {
				fail(w, 502, e)
			}
			return
		}
		p, e := c.Page("")
		if e != nil {
			fail(w, 500, e)
			return
		}
		reply(w, 201, p)
	})
}
