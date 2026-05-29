package app

import (
	"context"
	"errors"
	"fmt"
	"time"
)

// DefaultShutdownTimeout is used when Options.ShutdownTimeout is zero.
const DefaultShutdownTimeout = 30 * time.Second

// Phase identifies the lifecycle phase that returned an error.
type Phase string

const (
	// PhaseRun reports errors returned by Component.Run.
	PhaseRun Phase = "run"

	// PhaseShutdown reports errors returned by Component.Shutdown.
	PhaseShutdown Phase = "shutdown"
)

// Component is a long-running application component.
//
// Run should block until the component exits, fails, or ctx is canceled.
// Shutdown should gracefully stop the component within ctx.
type Component interface {
	Name() string
	Run(ctx context.Context) error
	Shutdown(ctx context.Context) error
}

// ComponentError wraps an error returned by a component lifecycle phase.
type ComponentError struct {
	Component string
	Phase     Phase
	Err       error
}

// Error returns a message that names the failing component and phase.
func (e *ComponentError) Error() string {
	if e == nil {
		return "<nil>"
	}
	return fmt.Sprintf("app: %s %s: %v", e.Component, e.Phase, e.Err)
}

// Unwrap returns the underlying component error.
func (e *ComponentError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}

// Options customizes RunWithOptions.
type Options struct {
	// ShutdownTimeout bounds the graceful shutdown phase. The default is
	// DefaultShutdownTimeout. Use a negative value to disable the timeout.
	ShutdownTimeout time.Duration
}

// Run runs components until ctx is canceled, a component returns, or a component
// fails. Context cancellation and clean component exits are treated as graceful
// shutdown paths. Component run and shutdown failures are returned.
func Run(ctx context.Context, components ...Component) error {
	return RunWithOptions(ctx, Options{}, components...)
}

// RunWithOptions runs components with opts.
func RunWithOptions(ctx context.Context, opts Options, components ...Component) error {
	if ctx == nil {
		return errors.New("app: context is required")
	}
	if err := validateComponents(components); err != nil {
		return err
	}
	if len(components) == 0 {
		return nil
	}

	runCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	results := make(chan componentResult, len(components))
	for _, component := range components {
		go runComponent(runCtx, component, results)
	}

	var errs []error
	remaining := len(components)
	shutdownStarted := false
	var shutdownCtx context.Context
	var shutdownCancel context.CancelFunc

	for remaining > 0 {
		select {
		case result := <-results:
			remaining--
			if result.err != nil {
				errs = append(errs, newComponentError(result.name, PhaseRun, result.err))
			}
			if !shutdownStarted {
				shutdownStarted = true
				shutdownCtx, shutdownCancel = newShutdownContext(opts)
				cancel()
				errs = append(errs, shutdownComponents(shutdownCtx, components)...)
			}
		case <-ctx.Done():
			if !shutdownStarted {
				shutdownStarted = true
				shutdownCtx, shutdownCancel = newShutdownContext(opts)
				cancel()
				errs = append(errs, shutdownComponents(shutdownCtx, components)...)
			}
		case <-done(shutdownStarted, shutdownCtx):
			errs = append(errs, newComponentError("components", PhaseShutdown, shutdownCtx.Err()))
			if shutdownCancel != nil {
				shutdownCancel()
			}
			return errors.Join(errs...)
		}
	}

	if shutdownCancel != nil {
		shutdownCancel()
	}
	return errors.Join(errs...)
}

type componentResult struct {
	name string
	err  error
}

func runComponent(ctx context.Context, component Component, results chan<- componentResult) {
	err := component.Run(ctx)
	if ctx.Err() != nil && (errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded)) {
		err = nil
	}
	results <- componentResult{name: component.Name(), err: err}
}

func validateComponents(components []Component) error {
	seen := make(map[string]struct{}, len(components))
	for i, component := range components {
		if component == nil {
			return fmt.Errorf("app: component %d is nil", i)
		}
		name := component.Name()
		if name == "" {
			return fmt.Errorf("app: component %d name is required", i)
		}
		if _, ok := seen[name]; ok {
			return fmt.Errorf("app: duplicate component name %q", name)
		}
		seen[name] = struct{}{}
	}
	return nil
}

func newShutdownContext(opts Options) (context.Context, context.CancelFunc) {
	switch {
	case opts.ShutdownTimeout < 0:
		return context.WithCancel(context.Background())
	case opts.ShutdownTimeout == 0:
		return context.WithTimeout(context.Background(), DefaultShutdownTimeout)
	default:
		return context.WithTimeout(context.Background(), opts.ShutdownTimeout)
	}
}

func shutdownComponents(ctx context.Context, components []Component) []error {
	errs := make([]error, 0, len(components))
	for i := len(components) - 1; i >= 0; i-- {
		component := components[i]
		if err := component.Shutdown(ctx); err != nil {
			errs = append(errs, newComponentError(component.Name(), PhaseShutdown, err))
		}
	}
	return errs
}

func newComponentError(name string, phase Phase, err error) error {
	if err == nil {
		return nil
	}
	return &ComponentError{
		Component: name,
		Phase:     phase,
		Err:       err,
	}
}

func done(enabled bool, ctx context.Context) <-chan struct{} {
	if !enabled || ctx == nil {
		return nil
	}
	return ctx.Done()
}
