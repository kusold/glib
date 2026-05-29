package errx

import (
	"errors"
	"fmt"
	"strings"
)

// Code identifies a stable application error class.
type Code string

const (
	// CodeClientClosedRequest reports that the client closed the request.
	CodeClientClosedRequest Code = "client_closed_request"

	// CodeBadRequest reports malformed or invalid caller input.
	CodeBadRequest Code = "bad_request"

	// CodeUnauthorized reports missing or invalid authentication.
	CodeUnauthorized Code = "unauthorized"

	// CodeForbidden reports that the caller is authenticated but not allowed.
	CodeForbidden Code = "forbidden"

	// CodeNotFound reports that the requested resource does not exist.
	CodeNotFound Code = "not_found"

	// CodeConflict reports that the operation conflicts with current state.
	CodeConflict Code = "conflict"

	// CodePreconditionFailed reports that a required precondition was not met.
	CodePreconditionFailed Code = "precondition_failed"

	// CodeTooManyRequests reports that the caller exceeded a limit.
	CodeTooManyRequests Code = "too_many_requests"

	// CodeInternalServerError reports an unexpected internal failure.
	CodeInternalServerError Code = "internal_server_error"

	// CodeServiceUnavailable reports that a dependency or service is temporarily unavailable.
	CodeServiceUnavailable Code = "service_unavailable"

	// CodeGatewayTimeout reports that the operation timed out.
	CodeGatewayTimeout Code = "gateway_timeout"
)

// DefaultPublicMessage is used when an error has no explicit safe message.
const DefaultPublicMessage = "internal error"

// Error is a coded application error.
type Error struct {
	code          Code
	publicMessage string
	cause         error
}

// New returns a coded error with a safe public message.
func New(code Code, publicMessage string) *Error {
	return &Error{
		code:          requireCode(code),
		publicMessage: cleanPublicMessage(publicMessage),
	}
}

// Wrap returns a coded error that wraps err.
//
// If err is nil, Wrap returns nil.
func Wrap(err error, code Code, publicMessage string) error {
	if err == nil {
		return nil
	}
	return &Error{
		code:          requireCode(code),
		publicMessage: cleanPublicMessage(publicMessage),
		cause:         err,
	}
}

// Errorf returns a coded error that wraps an internal fmt.Errorf error.
//
// The format string and arguments are only used for the internal cause. The
// coded error's Error method exposes the safe public message instead.
func Errorf(code Code, publicMessage, format string, args ...any) error {
	return Wrap(fmt.Errorf(format, args...), code, publicMessage)
}

// Error returns a safe error string containing only the code and public message.
func (e *Error) Error() string {
	if e == nil {
		return "<nil>"
	}
	return string(e.code) + ": " + e.PublicMessage()
}

// Code returns the stable application error code.
func (e *Error) Code() Code {
	if e == nil {
		return ""
	}
	return e.code
}

// PublicMessage returns the safe message for callers.
func (e *Error) PublicMessage() string {
	if e == nil {
		return ""
	}
	return publicMessageOrDefault(e.publicMessage)
}

// Unwrap returns the internal cause.
func (e *Error) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.cause
}

// Is reports whether target is a coded error with the same code.
func (e *Error) Is(target error) bool {
	if e == nil {
		return target == nil
	}

	targetErr, ok := target.(*Error)
	return ok && targetErr != nil && targetErr.code != "" && e.code == targetErr.code
}

// Is reports whether err contains a coded error with code.
func Is(err error, code Code) bool {
	if code == "" {
		return false
	}
	return errors.Is(err, &Error{code: code})
}

// CodeOf returns the first coded error code found in err.
func CodeOf(err error) (Code, bool) {
	var coded interface {
		Code() Code
	}
	if !errors.As(err, &coded) {
		return "", false
	}

	code := coded.Code()
	return code, code != ""
}

// PublicMessage returns a safe public message for err.
//
// Uncoded errors receive DefaultPublicMessage so private details are not
// accidentally exposed.
func PublicMessage(err error) string {
	if err == nil {
		return ""
	}

	var coded interface {
		PublicMessage() string
	}
	if !errors.As(err, &coded) {
		return DefaultPublicMessage
	}
	return publicMessageOrDefault(coded.PublicMessage())
}

func requireCode(code Code) Code {
	if strings.TrimSpace(string(code)) == "" {
		panic("errx: code is required")
	}
	return code
}

func cleanPublicMessage(message string) string {
	return strings.TrimSpace(message)
}

func publicMessageOrDefault(message string) string {
	if message == "" {
		return DefaultPublicMessage
	}
	return message
}
