package config

import (
	"reflect"

	"github.com/kusold/glib/env"
)

// ValidateFunc validates a loaded configuration value.
type ValidateFunc[T any] func(T) error

// ParserFunc parses a string environment value into a custom type.
type ParserFunc = env.ParserFunc

// Validatable is implemented by configuration values that validate themselves.
type Validatable interface {
	Validate() error
}

// Options customizes configuration loading.
type Options[T any] struct {
	// Environment provides variables to parse instead of os.Environ.
	Environment map[string]string

	// Prefix is prepended to every environment variable name.
	Prefix string

	// RequiredIfNoDefault treats fields without envDefault as required.
	RequiredIfNoDefault bool

	// UseFieldNameByDefault uses the struct field name when the env tag has no name.
	UseFieldNameByDefault bool

	// FuncMap adds parsers for custom field types.
	FuncMap map[reflect.Type]ParserFunc

	// Validators run after environment parsing succeeds.
	Validators []ValidateFunc[T]
}

// Load loads configuration from os.Environ.
func Load[T any]() (T, error) {
	return LoadWithOptions[T](Options[T]{})
}

// LoadWithOptions loads configuration from environment variables and validates it.
//
// When parsing fails, LoadWithOptions returns the zero value for T. When
// validation fails, it returns the parsed configuration value alongside the
// validation error so callers may inspect the loaded values.
func LoadWithOptions[T any](opts Options[T]) (T, error) {
	cfg, err := env.ParseWithOptions[T](toEnvOptions(opts))
	if err != nil {
		var zero T
		return zero, newError(err)
	}

	var errs []error
	if validatable, ok := any(cfg).(Validatable); ok {
		errs = appendFlattened(errs, validatable.Validate())
	} else if validatable, ok := any(&cfg).(Validatable); ok {
		errs = appendFlattened(errs, validatable.Validate())
	}

	for _, validate := range opts.Validators {
		if validate == nil {
			continue
		}
		errs = appendFlattened(errs, validate(cfg))
	}

	if len(errs) > 0 {
		return cfg, &Error{errors: errs}
	}
	return cfg, nil
}

func toEnvOptions[T any](opts Options[T]) env.Options {
	return env.Options{
		Environment:           opts.Environment,
		Prefix:                opts.Prefix,
		RequiredIfNoDefault:   opts.RequiredIfNoDefault,
		UseFieldNameByDefault: opts.UseFieldNameByDefault,
		FuncMap:               opts.FuncMap,
	}
}
