package main

import (
	"context"
	"errors"
	"log"
	"net"
	"net/http"
	"time"
)

// serve stops accepting work, gives active requests a bounded drain,
// then cancels remaining source reads and streams. Publications may still have
// committed when their connection closes; callers retain their existing retry contract.
func serve(ctx context.Context, server *http.Server) error {
	requests, cancel := context.WithCancel(context.Background())
	defer cancel()
	server.BaseContext = func(net.Listener) context.Context { return requests }
	finished := make(chan error, 1)
	go func() { finished <- server.ListenAndServe() }()
	select {
	case err := <-finished:
		return err
	case <-ctx.Done():
	}
	drain, stop := context.WithTimeout(context.Background(), 5*time.Second)
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

// SDK Shutdown flushes queued telemetry. Each provider gets an independent budget,
// including when another provider has already failed or exhausted its deadline.
func shutdownProvider(shutdown func(context.Context) error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := shutdown(ctx); err != nil {
		log.Printf("telemetry shutdown: %v", err)
	}
}
