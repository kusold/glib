package env

import (
	"net/url"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestParseWithOptionsRequiredMissing(t *testing.T) {
	type config struct {
		DatabaseURL string `env:"DATABASE_URL,required"`
	}

	_, err := ParseWithOptions[config](Options{Environment: map[string]string{}})
	if err == nil {
		t.Fatal("expected missing required value to fail")
	}
	if !strings.Contains(err.Error(), "DATABASE_URL") {
		t.Fatalf("expected error to name missing variable, got %v", err)
	}
}

func TestParseWithOptionsDefaultsAndScalars(t *testing.T) {
	type config struct {
		Debug   bool          `env:"DEBUG" envDefault:"true"`
		Port    int           `env:"PORT" envDefault:"8080"`
		Timeout time.Duration `env:"TIMEOUT" envDefault:"10s"`
		BaseURL url.URL       `env:"BASE_URL" envDefault:"https://example.com"`
	}

	cfg, err := ParseWithOptions[config](Options{Environment: map[string]string{}})
	if err != nil {
		t.Fatalf("parse config: %v", err)
	}

	if !cfg.Debug {
		t.Fatal("expected default debug to be true")
	}
	if cfg.Port != 8080 {
		t.Fatalf("expected default port 8080, got %d", cfg.Port)
	}
	if cfg.Timeout != 10*time.Second {
		t.Fatalf("expected default timeout 10s, got %s", cfg.Timeout)
	}
	if cfg.BaseURL.String() != "https://example.com" {
		t.Fatalf("expected default URL, got %s", cfg.BaseURL.String())
	}
}

func TestParseWithOptionsEmptyNotEmpty(t *testing.T) {
	type config struct {
		Token string `env:"TOKEN,notEmpty"`
	}

	_, err := ParseWithOptions[config](Options{Environment: map[string]string{"TOKEN": ""}})
	if err == nil {
		t.Fatal("expected empty notEmpty value to fail")
	}
	if !strings.Contains(err.Error(), "TOKEN") {
		t.Fatalf("expected error to name empty variable, got %v", err)
	}
}

func TestParseWithOptionsParseFailure(t *testing.T) {
	type config struct {
		Port int `env:"PORT"`
	}

	_, err := ParseWithOptions[config](Options{Environment: map[string]string{"PORT": "abc"}})
	if err == nil {
		t.Fatal("expected invalid int to fail")
	}
	if !strings.Contains(err.Error(), "Port") {
		t.Fatalf("expected error to name parsed field, got %v", err)
	}
}

func TestParseWithOptionsPrefix(t *testing.T) {
	type config struct {
		Port int `env:"PORT"`
	}

	cfg, err := ParseWithOptions[config](Options{
		Environment: map[string]string{"APP_PORT": "9090"},
		Prefix:      "APP_",
	})
	if err != nil {
		t.Fatalf("parse config: %v", err)
	}
	if cfg.Port != 9090 {
		t.Fatalf("expected prefixed port 9090, got %d", cfg.Port)
	}
}

func TestParseIntoWithOptions(t *testing.T) {
	type config struct {
		Addr string `env:"ADDR" envDefault:":8080"`
	}

	var cfg config
	if err := ParseIntoWithOptions(&cfg, Options{Environment: map[string]string{}}); err != nil {
		t.Fatalf("parse into config: %v", err)
	}
	if cfg.Addr != ":8080" {
		t.Fatalf("expected default addr, got %q", cfg.Addr)
	}
}

func TestParseWithOptionsRequiredIfNoDefault(t *testing.T) {
	type config struct {
		Addr string `env:"ADDR" envDefault:":8080"`
		DSN  string `env:"DATABASE_URL"`
	}

	_, err := ParseWithOptions[config](Options{
		Environment:         map[string]string{},
		RequiredIfNoDefault: true,
	})
	if err == nil {
		t.Fatal("expected missing value without default to fail")
	}
	if !strings.Contains(err.Error(), "DATABASE_URL") {
		t.Fatalf("expected error to name missing variable, got %v", err)
	}
}

func TestParseWithOptionsUseFieldNameByDefault(t *testing.T) {
	type config struct {
		Port int
	}

	cfg, err := ParseWithOptions[config](Options{
		Environment:           map[string]string{"PORT": "7000"},
		UseFieldNameByDefault: true,
	})
	if err != nil {
		t.Fatalf("parse config: %v", err)
	}
	if cfg.Port != 7000 {
		t.Fatalf("expected field-name port 7000, got %d", cfg.Port)
	}
}

func TestParseWithOptionsCustomParser(t *testing.T) {
	type region string
	type config struct {
		Region region `env:"REGION"`
	}

	cfg, err := ParseWithOptions[config](Options{
		Environment: map[string]string{"REGION": "us-west"},
		FuncMap: map[reflect.Type]ParserFunc{
			reflect.TypeOf(region("")): func(value string) (any, error) {
				return region(strings.ToUpper(value)), nil
			},
		},
	})
	if err != nil {
		t.Fatalf("parse config: %v", err)
	}
	if cfg.Region != "US-WEST" {
		t.Fatalf("expected custom parsed region, got %q", cfg.Region)
	}
}
