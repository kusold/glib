package log_test

import (
	"bytes"
	"fmt"
	"log/slog"
	"net/http"

	gliblog "github.com/kusold/glib/log"
)

func ExampleNew() {
	var out bytes.Buffer
	logger := gliblog.New(gliblog.Options{
		Writer:      &out,
		ReplaceAttr: dropTime,
	})

	logger.Info(
		"server started",
		gliblog.Component("api"),
		gliblog.Operation("listen"),
	)

	fmt.Print(out.String())

	// Output:
	// level=INFO msg="server started" component=api operation=listen
}

func ExampleNew_httpRequest() {
	var out bytes.Buffer
	logger := gliblog.New(gliblog.Options{
		Writer:      &out,
		ReplaceAttr: dropTime,
	})

	req, _ := http.NewRequest(http.MethodGet, "/users", nil)
	req.Header.Set("X-Request-ID", "req-123")

	logger.InfoContext(
		req.Context(),
		"request handled",
		gliblog.RequestID(req.Header.Get("X-Request-ID")),
		gliblog.Component("http"),
		gliblog.Operation(req.Method+" "+req.URL.Path),
	)

	fmt.Print(out.String())

	// Output:
	// level=INFO msg="request handled" request_id=req-123 component=http operation="GET /users"
}

func dropTime(_ []string, attr slog.Attr) slog.Attr {
	if attr.Key == slog.TimeKey {
		return slog.Attr{}
	}
	return attr
}
