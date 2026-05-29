package errx

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"testing"
)

func TestWrapKeepsInternalCauseOutOfPublicError(t *testing.T) {
	cause := errors.New("sql: password=secret failed")
	err := Wrap(cause, CodeNotFound, "user not found")

	if !errors.Is(err, cause) {
		t.Fatal("expected wrapped cause to match with errors.Is")
	}
	if strings.Contains(err.Error(), "secret") {
		t.Fatalf("expected Error to hide internal details, got %q", err.Error())
	}
	if got := PublicMessage(err); got != "user not found" {
		t.Fatalf("expected public message, got %q", got)
	}
}

func TestErrorsIsMatchesCodesThroughWrapping(t *testing.T) {
	err := fmt.Errorf("handler failed: %w", Wrap(sql.ErrNoRows, CodeNotFound, "user not found"))

	if !errors.Is(err, New(CodeNotFound, "ignored")) {
		t.Fatal("expected errors.Is to match coded error")
	}
	if !Is(err, CodeNotFound) {
		t.Fatal("expected Is helper to match code")
	}
	if Is(err, CodeConflict) {
		t.Fatal("did not expect conflict code to match")
	}
	if !errors.Is(err, sql.ErrNoRows) {
		t.Fatal("expected internal cause to remain matchable")
	}
}

func TestErrorsAsExtractsCodedError(t *testing.T) {
	err := fmt.Errorf("outer: %w", New(CodeBadRequest, "name is required"))

	var coded *Error
	if !errors.As(err, &coded) {
		t.Fatal("expected errors.As to find coded error")
	}
	if coded.Code() != CodeBadRequest {
		t.Fatalf("expected bad request code, got %q", coded.Code())
	}
	if coded.PublicMessage() != "name is required" {
		t.Fatalf("expected public message, got %q", coded.PublicMessage())
	}
}

func TestCodeOf(t *testing.T) {
	err := fmt.Errorf("outer: %w", New(CodeConflict, "already exists"))

	code, ok := CodeOf(err)
	if !ok {
		t.Fatal("expected code")
	}
	if code != CodeConflict {
		t.Fatalf("expected conflict code, got %q", code)
	}

	if _, ok := CodeOf(errors.New("plain")); ok {
		t.Fatal("did not expect code for uncoded error")
	}
}

func TestPublicMessageDefaultsToSafeMessage(t *testing.T) {
	if got := PublicMessage(errors.New("private detail")); got != DefaultPublicMessage {
		t.Fatalf("expected default public message, got %q", got)
	}

	err := New(CodeInternalServerError, "")
	if got := PublicMessage(err); got != DefaultPublicMessage {
		t.Fatalf("expected default public message for empty message, got %q", got)
	}
}

func TestErrorfUsesInternalFormattedCause(t *testing.T) {
	cause := errors.New("constraint users_email_key")
	err := Errorf(CodeConflict, "email already exists", "insert user: %w", cause)

	if !errors.Is(err, cause) {
		t.Fatal("expected formatted cause to be matchable")
	}
	if strings.Contains(err.Error(), "users_email_key") {
		t.Fatalf("expected public error to hide formatted cause, got %q", err.Error())
	}
}

func TestHTTPStatusMapping(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want int
	}{
		{name: "nil", err: nil, want: http.StatusOK},
		{name: "uncoded", err: errors.New("plain"), want: http.StatusInternalServerError},
		{name: "bad request", err: New(CodeBadRequest, "bad request"), want: http.StatusBadRequest},
		{name: "not found", err: New(CodeNotFound, "missing"), want: http.StatusNotFound},
		{name: "internal server error", err: New(CodeInternalServerError, "internal error"), want: http.StatusInternalServerError},
		{name: "gateway timeout", err: New(CodeGatewayTimeout, "timeout"), want: http.StatusGatewayTimeout},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := HTTPStatus(tt.err); got != tt.want {
				t.Fatalf("expected status %d, got %d", tt.want, got)
			}
		})
	}
}

func TestHTTPStatusMapperAllowsApplicationCodes(t *testing.T) {
	const codePaymentRequired Code = "payment_required"

	mapper := HTTPStatusMapper{
		codePaymentRequired: http.StatusPaymentRequired,
	}
	err := New(codePaymentRequired, "payment required")

	if got := mapper.Status(err); got != http.StatusPaymentRequired {
		t.Fatalf("expected payment required, got %d", got)
	}
	if got := HTTPStatus(err); got != http.StatusInternalServerError {
		t.Fatalf("expected default mapper to hide unknown application code as 500, got %d", got)
	}
}

func TestNewRequiresCode(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("expected missing code to panic")
		}
	}()

	_ = New("", "missing code")
}
