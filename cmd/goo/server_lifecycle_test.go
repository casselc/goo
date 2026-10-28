package main

import (
	"context"
	"errors"
	"net"
	"net/http"
	"testing"
	"time"
)

func TestServerShutdownCancelsRemainingRequest(t *testing.T) {
	entered, cancelled := make(chan struct{}), make(chan struct{})
	server := &http.Server{Addr: "127.0.0.1:0", Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		close(entered)
		<-r.Context().Done()
		close(cancelled)
	})}
	// Supply an already-bound listener to avoid a port reservation race.
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	server.Addr = listener.Addr().String()
	ctx, stop := context.WithCancel(t.Context())
	defer stop()
	result := make(chan error, 1)
	go func() { result <- runHTTPListener(ctx, server, listener, 30*time.Millisecond) }()
	client := &http.Client{Timeout: time.Second}
	defer client.CloseIdleConnections()
	requestDone := make(chan struct{})
	go func() {
		defer close(requestDone)
		for {
			response, err := client.Get("http://" + server.Addr)
			if err == nil {
				response.Body.Close()
				return
			}
			select {
			case <-entered:
				return
			case <-ctx.Done():
				return
			case <-time.After(time.Millisecond):
			}
		}
	}()
	select {
	case <-entered:
	case <-time.After(3 * time.Second):
		t.Fatal("request did not start")
	}
	stop()
	select {
	case err := <-result:
		if !errors.Is(err, context.DeadlineExceeded) {
			t.Fatalf("shutdown: %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("shutdown exceeded bound")
	}
	select {
	case <-cancelled:
	case <-time.After(time.Second):
		t.Fatal("request context not cancelled")
	}
	<-requestDone
}

func TestServerListenFailureReturns(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	if err := runHTTPServer(t.Context(), &http.Server{Addr: listener.Addr().String()}, time.Second); err == nil {
		t.Fatal("lost listen error")
	}
}

func TestServerShutdownDrainsCompletedRequest(t *testing.T) {
	entered, release := make(chan struct{}), make(chan struct{})
	server := &http.Server{Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		close(entered)
		<-release
		if r.Context().Err() != nil {
			t.Error("request cancelled before draining")
		}
		w.Write([]byte("published"))
	})}
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	ctx, stop := context.WithCancel(t.Context())
	defer stop()
	result := make(chan error, 1)
	go func() { result <- runHTTPListener(ctx, server, listener, time.Second) }()
	done := make(chan error, 1)
	go func() {
		response, err := http.Get("http://" + listener.Addr().String())
		if err == nil {
			response.Body.Close()
		}
		done <- err
	}()
	<-entered
	stop()
	close(release)
	if err := <-result; err != nil {
		t.Fatal(err)
	}
	if err := <-done; err != nil {
		t.Fatal(err)
	}
}
