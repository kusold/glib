package app

import (
	"context"
	"errors"
	"net"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"
)

func TestRunShutsDownWhenContextIsCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	started := make(chan struct{})
	shutdown := make(chan struct{})

	component := NewComponent("worker", func(ctx context.Context) error {
		close(started)
		<-ctx.Done()
		return ctx.Err()
	}, func(context.Context) error {
		close(shutdown)
		return nil
	})

	errs := make(chan error, 1)
	go func() {
		errs <- RunWithOptions(ctx, testOptions(), component)
	}()

	waitForClosed(t, started, "component start")
	cancel()

	if err := waitForError(t, errs); err != nil {
		t.Fatalf("expected graceful shutdown, got %v", err)
	}
	waitForClosed(t, shutdown, "component shutdown")
}

func TestRunReturnsComponentFailureAndShutsDownSiblings(t *testing.T) {
	boom := errors.New("boom")
	siblingShutdown := make(chan struct{})

	failing := NewComponent("failing", func(context.Context) error {
		return boom
	}, nil)
	sibling := NewComponent("sibling", func(ctx context.Context) error {
		<-ctx.Done()
		return ctx.Err()
	}, func(context.Context) error {
		close(siblingShutdown)
		return nil
	})

	err := RunWithOptions(context.Background(), testOptions(), failing, sibling)

	if !errors.Is(err, boom) {
		t.Fatalf("expected component failure, got %v", err)
	}
	var componentErr *ComponentError
	if !errors.As(err, &componentErr) {
		t.Fatalf("expected ComponentError, got %T", err)
	}
	if componentErr.Component != "failing" || componentErr.Phase != PhaseRun {
		t.Fatalf("expected failing run error, got %s %s", componentErr.Component, componentErr.Phase)
	}
	waitForClosed(t, siblingShutdown, "sibling shutdown")
}

func TestRunReturnsShutdownFailure(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	shutdownErr := errors.New("close database")

	component := NewComponent("db", func(ctx context.Context) error {
		<-ctx.Done()
		return ctx.Err()
	}, func(context.Context) error {
		return shutdownErr
	})

	errs := make(chan error, 1)
	go func() {
		errs <- RunWithOptions(ctx, testOptions(), component)
	}()

	cancel()
	err := waitForError(t, errs)
	if !errors.Is(err, shutdownErr) {
		t.Fatalf("expected shutdown error, got %v", err)
	}
	if !strings.Contains(err.Error(), "db shutdown") {
		t.Fatalf("expected component and phase in error, got %q", err.Error())
	}
}

func TestRunRejectsInvalidComponents(t *testing.T) {
	tests := []struct {
		name       string
		components []Component
		want       string
	}{
		{
			name:       "nil",
			components: []Component{nil},
			want:       "component 0 is nil",
		},
		{
			name:       "empty name",
			components: []Component{NewComponent("", func(context.Context) error { return nil }, nil)},
			want:       "name is required",
		},
		{
			name: "duplicate name",
			components: []Component{
				NewComponent("worker", func(context.Context) error { return nil }, nil),
				NewComponent("worker", func(context.Context) error { return nil }, nil),
			},
			want: `duplicate component name "worker"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := RunWithOptions(context.Background(), testOptions(), tt.components...)
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("expected %q error, got %v", tt.want, err)
			}
		})
	}
}

func TestHTTPServerComponent(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	requests := make(chan struct{}, 1)
	server := &http.Server{
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requests <- struct{}{}
			w.WriteHeader(http.StatusNoContent)
		}),
	}

	errs := make(chan error, 1)
	go func() {
		errs <- RunWithOptions(ctx, testOptions(), HTTPServer("http", server, WithListener(listener)))
	}()

	resp, err := http.Get("http://" + listener.Addr().String())
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if err := resp.Body.Close(); err != nil {
		t.Fatalf("close response body: %v", err)
	}
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", resp.StatusCode)
	}
	waitForClosedOrValue(t, requests, "http request")

	cancel()
	if err := waitForError(t, errs); err != nil {
		t.Fatalf("expected graceful HTTP shutdown, got %v", err)
	}
}

func TestSignalContextCancelsOnSignal(t *testing.T) {
	ctx, stop := SignalContext(context.Background(), os.Interrupt)
	defer stop()

	process, err := os.FindProcess(os.Getpid())
	if err != nil {
		t.Fatalf("find process: %v", err)
	}
	if err := process.Signal(os.Interrupt); err != nil {
		t.Fatalf("signal process: %v", err)
	}

	select {
	case <-ctx.Done():
	case <-time.After(time.Second):
		t.Fatal("expected signal to cancel context")
	}
}

func testOptions() Options {
	return Options{ShutdownTimeout: time.Second}
}

func waitForError(t *testing.T, errs <-chan error) error {
	t.Helper()

	select {
	case err := <-errs:
		return err
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for Run")
		return nil
	}
}

func waitForClosed(t *testing.T, ch <-chan struct{}, name string) {
	t.Helper()

	select {
	case <-ch:
	case <-time.After(time.Second):
		t.Fatalf("timed out waiting for %s", name)
	}
}

func waitForClosedOrValue(t *testing.T, ch <-chan struct{}, name string) {
	t.Helper()

	select {
	case <-ch:
	case <-time.After(time.Second):
		t.Fatalf("timed out waiting for %s", name)
	}
}
