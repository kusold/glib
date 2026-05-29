# env

Package `env` loads typed configuration structs from environment variables. It
uses `github.com/caarlos0/env/v11` at runtime, so struct tags follow the
caarlos0/env format:

```go
type Config struct {
	Addr    string        `env:"ADDR" envDefault:":8080"`
	Timeout time.Duration `env:"TIMEOUT" envDefault:"5s"`
	DSN     string        `env:"DATABASE_URL,required,notEmpty"`
}
```

## Documentation

The repository tracks `github.com/g4s8/envdoc` as a Go tool dependency. Use it
from application config packages to generate environment variable documentation
from the same struct tags and field comments used by `env`.

```go
//go:generate go tool envdoc -types Config -output environments.md
type Config struct {
	// Address the HTTP server listens on.
	Addr string `env:"ADDR" envDefault:":8080"`
}
```

Run generators with:

```sh
go generate ./...
```

`envdoc` expects `go generate` metadata, so keep the directive next to the
config type that owns the environment variable comments.
