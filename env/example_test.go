package env_test

import (
	"fmt"
	"log"
	"time"

	"github.com/kusold/glib/env"
)

func ExampleParseWithOptions() {
	type Config struct {
		// Address the HTTP server listens on.
		Addr string `env:"ADDR" envDefault:":8080"`

		// Graceful shutdown timeout.
		ShutdownTimeout time.Duration `env:"SHUTDOWN_TIMEOUT" envDefault:"5s"`

		// Database connection string.
		DatabaseURL string `env:"DATABASE_URL,required,notEmpty"`
	}

	cfg, err := env.ParseWithOptions[Config](env.Options{
		Environment: map[string]string{"DATABASE_URL": "postgres://localhost/glib"},
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(cfg.Addr)
	fmt.Println(cfg.ShutdownTimeout)
	fmt.Println(cfg.DatabaseURL)

	// Output:
	// :8080
	// 5s
	// postgres://localhost/glib
}
