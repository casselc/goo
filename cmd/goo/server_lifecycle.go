package main

import (
	"context"
	"errors"
	"net"
	"net/http"
	"time"
)

// runHTTPServer stops accepting work, gives active requests a bounded drain,
// then cancels remaining source reads and streams. Publications may still have
// committed when their connection closes; callers retain their existing retry contract.
func runHTTPServer(ctx context.Context, server *http.Server, grace time.Duration) error {
	listener, err := net.Listen("tcp", server.Addr)
	if err != nil {
		return err
	}
	return runHTTPListener(ctx, server, listener, grace)
}

func runHTTPListener(ctx context.Context, server *http.Server, listener net.Listener, grace time.Duration) error {
	requests, cancel := context.WithCancel(context.Background())
	defer cancel()
	server.BaseContext = func(net.Listener) context.Context { return requests }
	finished := make(chan error, 1)
	go func() { finished <- server.Serve(listener) }()
	select {
	case err := <-finished:
		return err
	case <-ctx.Done():
	}
	drain, stop := context.WithTimeout(context.Background(), grace)
	defer stop()
	err := server.Shutdown(drain)
	cancel()
	if err != nil {
		err = errors.Join(err, server.Close())
	}
	serveErr := <-finished
	if errors.Is(serveErr, http.ErrServerClosed) {
		serveErr = nil
	}
	return errors.Join(err, serveErr)
}
