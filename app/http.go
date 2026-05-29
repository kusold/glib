package app

import (
	"context"
	"errors"
	"net"
	"net/http"
)

// HTTPServerOption customizes an HTTP server component.
type HTTPServerOption func(*httpServer)

// WithListener serves HTTP on listener instead of calling ListenAndServe.
func WithListener(listener net.Listener) HTTPServerOption {
	return func(component *httpServer) {
		component.listener = listener
	}
}

// HTTPServer adapts server into a Component.
func HTTPServer(name string, server *http.Server, opts ...HTTPServerOption) Component {
	component := &httpServer{
		name:   name,
		server: server,
	}
	for _, opt := range opts {
		if opt != nil {
			opt(component)
		}
	}
	return component
}

type httpServer struct {
	name     string
	server   *http.Server
	listener net.Listener
}

func (s *httpServer) Name() string {
	return s.name
}

func (s *httpServer) Run(context.Context) error {
	if s.server == nil {
		return errors.New("http server is required")
	}

	var err error
	if s.listener != nil {
		err = s.server.Serve(s.listener)
	} else {
		err = s.server.ListenAndServe()
	}
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return err
}

func (s *httpServer) Shutdown(ctx context.Context) error {
	if s.server == nil {
		return nil
	}
	err := s.server.Shutdown(ctx)
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return err
}
