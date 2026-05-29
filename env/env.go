package env

import (
	"fmt"
	"reflect"

	cenv "github.com/caarlos0/env/v11"
)

// ParserFunc parses a string environment value into a custom type.
type ParserFunc = cenv.ParserFunc

// Options customizes struct parsing.
type Options struct {
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
}

// Parse loads environment variables into a new value of type T.
func Parse[T any]() (T, error) {
	return ParseWithOptions[T](Options{})
}

// ParseWithOptions loads environment variables into a new value of type T.
func ParseWithOptions[T any](opts Options) (T, error) {
	cfg, err := cenv.ParseAsWithOptions[T](toUpstreamOptions(opts))
	if err != nil {
		var zero T
		return zero, fmt.Errorf("env: parse config: %w", err)
	}
	return cfg, nil
}

// ParseInto loads environment variables into target.
func ParseInto(target any) error {
	return ParseIntoWithOptions(target, Options{})
}

// ParseIntoWithOptions loads environment variables into target.
func ParseIntoWithOptions(target any, opts Options) error {
	if err := cenv.ParseWithOptions(target, toUpstreamOptions(opts)); err != nil {
		return fmt.Errorf("env: parse config: %w", err)
	}
	return nil
}

func toUpstreamOptions(opts Options) cenv.Options {
	return cenv.Options{
		Environment:           opts.Environment,
		Prefix:                opts.Prefix,
		RequiredIfNoDef:       opts.RequiredIfNoDefault,
		UseFieldNameByDefault: opts.UseFieldNameByDefault,
		FuncMap:               opts.FuncMap,
	}
}
