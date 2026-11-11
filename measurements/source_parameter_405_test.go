package measurements

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNativeSourceParameterHTTP405(t *testing.T) {
	const value = "literal\\n\x00\t\n\r\b\f\x1b\x1f\x7f雪😀 +%&=#?'"
	const want = `literal\\n\x00\x09\x0a\x0d\x08\x0c\x1b\x1f\x7f雪😀 +%&=#?'`
	const sql = "SELECT {literal:String} FORMAT JSONEachRow"
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		body, err := io.ReadAll(r.Body)
		if err != nil || string(body) != sql || r.URL.Query().Get("param_literal") != want {
			t.Errorf("literal or SQL wire changed: body=%q literal=%q error=%v", body, r.URL.Query().Get("param_literal"), err)
		}
		if r.URL.Query().Get("max_result_rows") != "10001" || r.URL.Query().Has("max_memory_usage") {
			t.Error("native source policy changed")
		}
		_, _ = io.WriteString(w, "{}\n")
	}))
	defer server.Close()
	parameters := map[string]string{"literal": value}
	rows, err := (Source{Endpoint: server.URL}).read(context.Background(), Plan{SQL: sql, Parameters: parameters})
	if err != nil || len(rows) != 1 || calls != 1 || parameters["literal"] != value {
		t.Fatal("native roundtrip changed", err, calls, parameters)
	}
}
