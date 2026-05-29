package config

import (
	"fmt"
	"strings"
)

// Error reports one or more configuration loading or validation errors.
type Error struct {
	errors []error
}

// Error returns a compact, human-readable description of all configuration errors.
func (e *Error) Error() string {
	if e == nil || len(e.errors) == 0 {
		return "config: no errors"
	}
	if len(e.errors) == 1 {
		return "config: " + e.errors[0].Error()
	}

	var builder strings.Builder
	fmt.Fprintf(&builder, "config: %d errors:", len(e.errors))
	for _, err := range e.errors {
		builder.WriteString(" ")
		builder.WriteString(err.Error())
		builder.WriteString(";")
	}
	return strings.TrimRight(builder.String(), ";")
}

// Errors returns the individual errors.
func (e *Error) Errors() []error {
	if e == nil {
		return nil
	}
	return append([]error(nil), e.errors...)
}

// Unwrap returns the individual errors for errors.Is and errors.As.
func (e *Error) Unwrap() []error {
	if e == nil {
		return nil
	}
	return e.errors
}

func newError(err error) *Error {
	return &Error{errors: appendFlattened(nil, err)}
}

func appendFlattened(errs []error, err error) []error {
	if err == nil {
		return errs
	}
	for _, err := range flatten(err) {
		errs = append(errs, err)
	}
	return errs
}

func flatten(err error) []error {
	if err == nil {
		return nil
	}

	if multi, ok := err.(interface{ Unwrap() []error }); ok {
		var errs []error
		for _, child := range multi.Unwrap() {
			errs = appendFlattened(errs, child)
		}
		return errs
	}

	if single, ok := err.(interface{ Unwrap() error }); ok {
		child := single.Unwrap()
		if _, ok := child.(interface{ Unwrap() []error }); ok {
			return flatten(child)
		}
	}

	return []error{err}
}
