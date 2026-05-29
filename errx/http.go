package errx

import "net/http"

// HTTPStatusMapper maps error codes to HTTP status codes.
//
// Missing codes fall back to the package defaults and then to 500.
type HTTPStatusMapper map[Code]int

const statusClientClosedRequest = 499

var defaultHTTPStatusByCode = HTTPStatusMapper{
	CodeClientClosedRequest: statusClientClosedRequest,
	CodeBadRequest:          http.StatusBadRequest,
	CodeUnauthorized:        http.StatusUnauthorized,
	CodeForbidden:           http.StatusForbidden,
	CodeNotFound:            http.StatusNotFound,
	CodeConflict:            http.StatusConflict,
	CodePreconditionFailed:  http.StatusPreconditionFailed,
	CodeTooManyRequests:     http.StatusTooManyRequests,
	CodeInternalServerError: http.StatusInternalServerError,
	CodeServiceUnavailable:  http.StatusServiceUnavailable,
	CodeGatewayTimeout:      http.StatusGatewayTimeout,
}

// HTTPStatus returns the default HTTP status for err.
func HTTPStatus(err error) int {
	return defaultHTTPStatusByCode.Status(err)
}

// HTTPStatusCode returns the default HTTP status for code.
func HTTPStatusCode(code Code) int {
	return defaultHTTPStatusByCode.StatusCode(code)
}

// Status returns the HTTP status for err.
func (m HTTPStatusMapper) Status(err error) int {
	if err == nil {
		return http.StatusOK
	}

	code, ok := CodeOf(err)
	if !ok {
		return http.StatusInternalServerError
	}
	return m.StatusCode(code)
}

// StatusCode returns the HTTP status for code.
func (m HTTPStatusMapper) StatusCode(code Code) int {
	if status, ok := m[code]; ok {
		return normalizeHTTPStatus(status)
	}
	if status, ok := defaultHTTPStatusByCode[code]; ok {
		return status
	}
	return http.StatusInternalServerError
}

func normalizeHTTPStatus(status int) int {
	if status < 100 || status > 599 {
		return http.StatusInternalServerError
	}
	return status
}
