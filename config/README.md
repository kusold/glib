# config

Package `config` loads typed application configuration from environment
variables and then runs validation hooks.

Struct field tags use the same `env` and `envDefault` format as
`github.com/kusold/glib/env`:

```go
type Config struct {
	Addr        string        `env:"ADDR" envDefault:":8080"`
	DatabaseURL string        `env:"DATABASE_URL,required,notEmpty"`
	Timeout     time.Duration `env:"TIMEOUT" envDefault:"5s"`
}

func (c Config) Validate() error {
	if c.Timeout <= 0 {
		return errors.New("TIMEOUT must be positive")
	}
	return nil
}

cfg, err := config.Load[Config]()
```

## Using go-playground/validator

Applications that prefer validation tags can plug in
`github.com/go-playground/validator/v10` without adding that dependency to
`glib/config` itself:

```go
import (
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/kusold/glib/config"
)

type Config struct {
	Addr        string        `env:"ADDR" envDefault:":8080" validate:"required"`
	DatabaseURL string        `env:"DATABASE_URL,required,notEmpty" validate:"required,url"`
	Timeout     time.Duration `env:"TIMEOUT" envDefault:"5s" validate:"gt=0"`
}

func LoadConfig() (Config, error) {
	validate := validator.New()

	return config.LoadWithOptions[Config](config.Options[Config]{
		Validators: []config.ValidateFunc[Config]{
			validate.Struct,
		},
	})
}
```

Use `Validate() error` or `ValidateFunc` for cross-field rules that are clearer
in Go code than in struct tags.
