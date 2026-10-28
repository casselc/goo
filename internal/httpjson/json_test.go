package httpjson_test

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/casselc/goo/internal/httpjson"
	"github.com/casselc/goo/internal/queryhttp"
)

func TestIncompleteBodyReleasesAdmission(t *testing.T) {
	for index, body := range []string{"{", "{}", "x", `{"` + strings.Repeat("unknown", 750) + `":1}`} {
		for _, cause := range []string{"cancel", "deadline"} {
			t.Run(fmt.Sprintf("%s/%d", cause, index), func(t *testing.T) {
				entered := make(chan struct{}, 2)
				finished := make(chan struct{}, 2)
				var effects atomic.Int32
				handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					entered <- struct{}{}
					var v struct{}

					if cause == "deadline" {
						ctx, cancel := context.WithTimeout(r.Context(), 100*time.Millisecond)
						defer cancel()
						r = r.WithContext(ctx)
					}
					if err := httpjson.Decode(w, r, &v, 16384, "one object"); err != nil {
						http.Error(w, err.Error(), 400)
						return
					}
					effects.Add(1)
					w.WriteHeader(204)
				})
				budget := queryhttp.NewBudget(handler, 1, func(r *http.Request) bool { return r.URL.Path == "/read" })
				srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					budget.ServeHTTP(w, r)
					if r.URL.Path == "/read" {
						finished <- struct{}{}
					}
				}))
				defer srv.Close()
				conn, err := net.Dial("tcp", srv.Listener.Addr().String())
				if err != nil {
					t.Fatal(err)
				}
				defer conn.Close()
				id := "12345678-1234-1234-1234-123456789abc"
				fmt.Fprintf(conn, "POST /read HTTP/1.1\r\nHost: test\r\nContent-Length: %d\r\nX-Goo-Query-ID: %s\r\n\r\n%s", len(body)+100, id, body)
				select {
				case <-entered:
				case <-time.After(time.Second):
					t.Fatal("handler not entered")
				}
				client := srv.Client()
				client.Timeout = time.Second
				if cause == "cancel" {
					resp, err := client.Post(srv.URL+"/queries/"+id+"/cancel", "application/json", nil)
					if err != nil {
						t.Fatal(err)
					}
					resp.Body.Close()
					if resp.StatusCode != 202 {
						t.Fatal(resp.StatusCode)
					}
				}
				select {
				case <-finished:
				case <-time.After(time.Second):
					t.Fatal("incomplete body retained admission")
				}
				if effects.Load() != 0 {
					t.Fatal("effects before complete valid body")
				}
				// Keep original TCP peer open throughout the next admission.
				resp, err := client.Post(srv.URL+"/read", "application/json", strings.NewReader("{}"))
				if err != nil {
					t.Fatal(err)
				}
				resp.Body.Close()
				if resp.StatusCode != 204 || effects.Load() != 1 {
					t.Fatal(resp.StatusCode, effects.Load())
				}
			})
		}
	}
}
