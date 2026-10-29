package verification

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/casselc/goo/exploration"
	"hegel.dev/go/hegel"
)

// Accepted public writer inputs must also be accepted by metadata discovery and
// explicit recovery. The oracle uses the submitted scope and immutable bytes.
func TestHegelAcceptedCaptureIndexability(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		if strings.Contains(string(b), "toString(count()) AS count") {
			fmt.Fprintln(w, `{"count":"0"}`)
		}
	}))
	defer server.Close()
	hegel.Test(t, func(tc *hegel.T) {
		alphabet := []string{"<", ">", "&", "雪", "λ", "\u2028", "\ufffd", "\\", "\""}
		text := strings.Repeat(hegel.Draw(tc, hegel.SampledFrom(alphabet)), hegel.Draw(tc, hegel.Integers(1, 340)))
		leaves := make([]exploration.Predicate, hegel.Draw(tc, hegel.Integers(1, 31)))
		for i := range leaves {
			leaves[i] = exploration.Predicate{Field: &exploration.Field{Kind: "column", Key: "service"}, Op: "eq", Value: text}
		}
		signal := hegel.Draw(tc, hegel.SampledFrom([]string{"logs", "spans"}))
		env := hegel.Draw(tc, hegel.SampledFrom([]string{"prod", "<雪>&", "λ"}))
		query := exploration.Query{Signal: signal, Environment: env, FromMS: 1000, ToMS: 61000, Limit: 100, Where: &exploration.Predicate{All: leaves}}
		q := exploration.CaptureQuestion{Kind: "explore", Explore: &query}
		store := exploration.CaptureStore{Dir: tc.TempDir()}
		saved, e := store.Create(tc.Context(), exploration.Source{Endpoint: server.URL}, "Saved <雪>", q, "")
		if e != nil {
			tc.Fatal(e)
		}
		file := filepath.Join(store.Dir, saved.ID+".json")
		before, e := os.ReadFile(file)
		if e != nil {
			tc.Fatal(e)
		}
		check := func() {
			p, e := store.Search(tc.Context(), exploration.CaptureSearchQuery{Signal: signal, Environment: &env}, "")
			if e != nil || p.Total != 1 || p.Matched != 1 || len(p.Items) != 1 {
				tc.Fatal("accepted capture not findable", p, e)
			}
			s := p.Items[0]
			if s.ID != saved.ID || s.Title != saved.Title || s.Kind != "explore" || s.Signal != signal || s.Environment == nil || *s.Environment != env {
				tc.Fatal("wrong projection", s)
			}
			wire, _ := json.Marshal(s)
			if len(wire) > 2048 || bytes.Contains(wire, []byte(`"question"`)) {
				tc.Fatal("unbounded or false receipt", len(wire))
			}
		}
		check()
		if e = os.RemoveAll(filepath.Join(store.Dir, ".catalog-v2")); e != nil {
			tc.Fatal(e)
		}
		r, e := store.RebuildCatalog(tc.Context())
		if e != nil || r.Indexed != 1 {
			tc.Fatal(r, e)
		}
		check()
		after, e := os.ReadFile(file)
		if e != nil || !bytes.Equal(before, after) {
			tc.Fatal("repair rewrote evidence", e)
		}
	}, options(60)...)
}
