package config

import (
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"
)

type selfValidatingConfig struct {
	Timeout time.Duration `env:"TIMEOUT" envDefault:"0s"`
}

func (c selfValidatingConfig) Validate() error {
	if c.Timeout <= 0 {
		return errors.New("TIMEOUT must be positive")
	}
	return nil
}

func TestLoadWithOptionsDefaultsAndRequired(t *testing.T) {
	type appConfig struct {
		Addr        string        `env:"ADDR" envDefault:":8080"`
		Port        int           `env:"PORT" envDefault:"8080"`
		Debug       bool          `env:"DEBUG" envDefault:"true"`
		Timeout     time.Duration `env:"TIMEOUT" envDefault:"5s"`
		DatabaseURL string        `env:"DATABASE_URL,required,notEmpty"`
	}

	cfg, err := LoadWithOptions[appConfig](Options[appConfig]{
		Environment: map[string]string{"DATABASE_URL": "postgres://localhost/glib"},
	})
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	if cfg.Addr != ":8080" {
		t.Fatalf("expected default addr, got %q", cfg.Addr)
	}
	if cfg.Port != 8080 {
		t.Fatalf("expected default port 8080, got %d", cfg.Port)
	}
	if !cfg.Debug {
		t.Fatal("expected default debug true")
	}
	if cfg.Timeout != 5*time.Second {
		t.Fatalf("expected default timeout 5s, got %s", cfg.Timeout)
	}
	if cfg.DatabaseURL != "postgres://localhost/glib" {
		t.Fatalf("expected database URL from environment, got %q", cfg.DatabaseURL)
	}
}

func TestLoadWithOptionsAggregatesRequiredErrors(t *testing.T) {
	type appConfig struct {
		Username string `env:"USERNAME,required"`
		Password string `env:"PASSWORD,required"`
	}

	_, err := LoadWithOptions[appConfig](Options[appConfig]{
		Environment: map[string]string{},
	})
	if err == nil {
		t.Fatal("expected missing required values to fail")
	}

	cfgErr := requireConfigError(t, err)
	if len(cfgErr.Errors()) != 2 {
		t.Fatalf("expected 2 errors, got %d: %v", len(cfgErr.Errors()), err)
	}
	requireErrorContains(t, err, "USERNAME", "PASSWORD")
}

func TestLoadWithOptionsAggregatesParseErrors(t *testing.T) {
	type appConfig struct {
		Port    int           `env:"PORT"`
		Timeout time.Duration `env:"TIMEOUT"`
	}

	_, err := LoadWithOptions[appConfig](Options[appConfig]{
		Environment: map[string]string{
			"PORT":    "wat",
			"TIMEOUT": "later",
		},
	})
	if err == nil {
		t.Fatal("expected parse errors to fail")
	}

	cfgErr := requireConfigError(t, err)
	if len(cfgErr.Errors()) != 2 {
		t.Fatalf("expected 2 errors, got %d: %v", len(cfgErr.Errors()), err)
	}
	requireErrorContains(t, err, "Port", "Timeout")
}

func TestLoadWithOptionsRunsValidators(t *testing.T) {
	type appConfig struct {
		Port int `env:"PORT" envDefault:"0"`
	}

	cfg, err := LoadWithOptions[appConfig](Options[appConfig]{
		Environment: map[string]string{},
		Validators: []ValidateFunc[appConfig]{
			func(cfg appConfig) error {
				if cfg.Port <= 0 {
					return errors.New("PORT must be positive")
				}
				return nil
			},
			func(cfg appConfig) error {
				if cfg.Port < 1024 {
					return errors.New("PORT must be unprivileged")
				}
				return nil
			},
		},
	})
	if err == nil {
		t.Fatal("expected validation errors")
	}
	if cfg.Port != 0 {
		t.Fatalf("expected parsed config to be returned, got port %d", cfg.Port)
	}

	cfgErr := requireConfigError(t, err)
	if len(cfgErr.Errors()) != 2 {
		t.Fatalf("expected 2 errors, got %d: %v", len(cfgErr.Errors()), err)
	}
	requireErrorContains(t, err, "PORT must be positive", "PORT must be unprivileged")
}

func TestLoadWithOptionsRunsStructValidate(t *testing.T) {
	_, err := LoadWithOptions[selfValidatingConfig](Options[selfValidatingConfig]{
		Environment: map[string]string{},
	})
	if err == nil {
		t.Fatal("expected struct validation to fail")
	}
	requireErrorContains(t, err, "TIMEOUT must be positive")
}

func TestLoadWithOptionsCustomParser(t *testing.T) {
	type region string
	type appConfig struct {
		Region region `env:"REGION"`
	}

	cfg, err := LoadWithOptions[appConfig](Options[appConfig]{
		Environment: map[string]string{"REGION": "us-west"},
		FuncMap: map[reflect.Type]ParserFunc{
			reflect.TypeOf(region("")): func(value string) (any, error) {
				return region(strings.ToUpper(value)), nil
			},
		},
	})
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if cfg.Region != "US-WEST" {
		t.Fatalf("expected parsed region, got %q", cfg.Region)
	}
}

func requireConfigError(t *testing.T, err error) *Error {
	t.Helper()

	var cfgErr *Error
	if !errors.As(err, &cfgErr) {
		t.Fatalf("expected *Error, got %T: %v", err, err)
	}
	return cfgErr
}

func requireErrorContains(t *testing.T, err error, values ...string) {
	t.Helper()

	message := err.Error()
	for _, value := range values {
		if !strings.Contains(message, value) {
			t.Fatalf("expected %q to contain %q", message, value)
		}
	}
}
