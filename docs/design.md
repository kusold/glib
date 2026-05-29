# glib Design

`glib` is a collection of common Go libraries intended to accelerate development
across many different Go projects. The repository should provide small,
composable packages for recurring application concerns, plus a reference app that
shows how those packages fit together in a real service.

## Goals

- Provide reusable packages for common application behavior.
- Keep package APIs small, stable, and easy to adopt independently.
- Prefer the Go standard library unless a dependency clearly earns its place.
- Use a reference app as a living integration example for the libraries.
- Follow the Go module layout guidance from https://go.dev/doc/modules/layout.

## Non-Goals

- Do not create a full application framework.
- Do not hide standard Go concepts behind large abstractions.
- Do not place reusable libraries under `internal`, since downstream projects
  need to import them.
- Do not add packages until a real use case exists in the reference app or in
  another consuming project.

## Repository Layout

The repository is a single Go module:

```text
github.com/kusold/glib
```

Reusable libraries live as top-level package directories. Commands live under
`cmd`. App-specific implementation details for the reference app live under
`internal` so they can change freely without becoming part of the public API.

```text
/
  go.mod
  README.md
  docs/
    design.md

  app/
  config/
  env/
  errx/
  health/
  httputil/
  log/
  retry/
  worker/
  db/
  testkit/

  cmd/
    glib-reference/
      main.go

  internal/
    reference/
      app/
      handlers/
      storage/
      jobs/
      migrations/
```

## Package Guidelines

- Each public package should solve one clear problem.
- Package names should be short, concrete, and idiomatic.
- Public APIs should be designed around standard Go types such as `context.Context`,
  `error`, `http.Handler`, `log/slog`, `database/sql`, and `testing.T`.
- Packages should compose with each other but remain independently useful.
- Cross-package dependencies should stay shallow. Foundation packages should not
  import higher-level packages.
- Tests should demonstrate expected behavior and serve as usage examples.

## Initial Libraries

### `env`

Small helpers for reading environment variables with typed parsing, defaults,
and required-value checks.

### `config`

Typed configuration loading built on top of `env`, with predictable defaults and
validation hooks.

### `log`

Structured logging helpers around `log/slog`, including production-friendly
defaults and request/job correlation fields.

### `errx`

Error helpers for wrapping internal errors while exposing stable public error
codes, messages, and HTTP mappings.

### `app`

Application lifecycle orchestration for startup, shutdown, signal handling, and
coordinated background components.

### `health`

Liveness and readiness primitives that can be reused across HTTP servers,
workers, and CLIs.

### `httputil`

HTTP server helpers for JSON encoding/decoding, middleware, request IDs, panic
recovery, and error responses.

### `retry`

Retry, backoff, jitter, and timeout helpers for transient operations.

### `worker`

Background worker and task-group primitives with context cancellation and
graceful shutdown.

### `db`

Database connection, transaction, and migration helpers. The package should
start narrow and avoid becoming an ORM.

### `testkit`

Testing helpers for environment isolation, HTTP handlers, temporary databases,
golden files, and common assertions.

## Reference App

The reference app should be a small but realistic service under
`cmd/glib-reference` with app-specific code in `internal/reference`.

The app should exercise the public packages together:

- Load typed configuration.
- Configure structured logging.
- Start an HTTP API.
- Expose liveness and readiness endpoints.
- Decode and encode JSON.
- Map internal errors to public responses.
- Use a database-backed storage layer.
- Run a background worker.
- Shut down gracefully.
- Include tests that use `testkit`.

The reference app should remain intentionally modest. Its job is to demonstrate
library integration, not to become a product.

## Build Order

1. Initialize the Go module and core repository conventions.
2. Build foundation packages: `env`, `config`, `log`, and `errx`.
3. Build runtime packages: `app`, `health`, and `httputil`.
4. Create the first version of `cmd/glib-reference`.
5. Add operational packages: `retry`, `worker`, and `db`.
6. Add `testkit` once the testing patterns are visible in real package and
   reference app tests.
