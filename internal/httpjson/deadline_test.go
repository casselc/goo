package httpjson

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestDecodeClearsDeadlineBeforeStreaming(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var v struct{}
		if err := decodeWithin(w, r, &v, 1024, "one object", 20*time.Millisecond); err != nil {
			t.Error(err)
			return
		}
		w.Header().Set("Content-Type", "text/event-stream")
		w.Write([]byte("data: first\n\n"))
		w.(http.Flusher).Flush()
		select {
		case <-time.After(60 * time.Millisecond):
		case <-r.Context().Done():
			t.Error("receive deadline cancelled response")
		}
		w.Write([]byte("data: last\n\n"))
	}))
	defer srv.Close()
	req, _ := http.NewRequestWithContext(context.Background(), "POST", srv.URL, strings.NewReader("{}"))
	resp, err := srv.Client().Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil || !strings.Contains(string(b), "last") {
		t.Fatal(string(b), err)
	}
}

func TestReceiveDeadlineWithoutContextExpiry(t *testing.T) {
	finished := make(chan error, 1)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var v struct{}
		finished <- decodeWithin(w, r, &v, 1024, "one object", 30*time.Millisecond)
	}))
	defer srv.Close()
	conn, err := net.Dial("tcp", srv.Listener.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	fmt.Fprint(conn, "POST / HTTP/1.1\r\nHost: test\r\nContent-Length: 100\r\n\r\n{}")
	select {
	case err := <-finished:
		if err == nil {
			t.Fatal("accepted incomplete body")
		}
	case <-time.After(time.Second):
		t.Fatal("receive deadline did not interrupt read")
	}
}
