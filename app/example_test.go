package app_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/kusold/glib/app"
)

func ExampleRun() {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	worker := app.NewComponent("worker", func(ctx context.Context) error {
		<-ctx.Done()
		return ctx.Err()
	}, nil)

	if err := app.Run(ctx, worker); err != nil {
		fmt.Println(err)
	}

	// Output:
}

func ExampleHTTPServer() {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	ctx, stop := app.SignalContext(context.Background())
	defer stop()

	err := app.Run(ctx, app.HTTPServer("api", server))
	if err != nil && !errors.Is(err, context.Canceled) {
		fmt.Println(err)
	}
}
