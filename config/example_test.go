package config_test

import (
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/kusold/glib/config"
)

type serverConfig struct {
	Addr            string        `env:"ADDR" envDefault:":8080"`
	ShutdownTimeout time.Duration `env:"SHUTDOWN_TIMEOUT" envDefault:"5s"`
	DatabaseURL     string        `env:"DATABASE_URL,required,notEmpty"`
}

func (c serverConfig) Validate() error {
	if c.ShutdownTimeout <= 0 {
		return errors.New("SHUTDOWN_TIMEOUT must be positive")
	}
	return nil
}

func ExampleLoadWithOptions() {
	cfg, err := config.LoadWithOptions[serverConfig](config.Options[serverConfig]{
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
