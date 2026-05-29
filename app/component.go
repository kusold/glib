package app

import "context"

// NewComponent returns a component backed by functions.
//
// A nil shutdown function is treated as a no-op, which is useful for components
// that stop themselves when the run context is canceled.
func NewComponent(name string, run func(context.Context) error, shutdown func(context.Context) error) Component {
	return functionComponent{
		name:     name,
		run:      run,
		shutdown: shutdown,
	}
}

type functionComponent struct {
	name     string
	run      func(context.Context) error
	shutdown func(context.Context) error
}

func (c functionComponent) Name() string {
	return c.name
}

func (c functionComponent) Run(ctx context.Context) error {
	if c.run == nil {
		return errMissingRun
	}
	return c.run(ctx)
}

func (c functionComponent) Shutdown(ctx context.Context) error {
	if c.shutdown == nil {
		return nil
	}
	return c.shutdown(ctx)
}
