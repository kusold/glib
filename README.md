# glib

Common Libraries for Go.

`glib` is a single Go module for small, composable packages that solve recurring
application concerns. Packages are added only when there is a concrete use case,
with reusable libraries living in top-level package directories and reference
application code living under `cmd` and `internal`.

See [docs/design.md](docs/design.md) for the repository layout, package
guidelines, and build order.

## Development

This repository uses Nix flakes with flake-parts for the development shell,
direnv for automatic shell activation, just as the command runner, and prek for
commit hooks.

Enable the development shell:

```sh
direnv allow
```

Install commit hooks:

```sh
just install-hooks
```

Common commands:

```sh
just fmt
just test
just lint
just hooks
```

The core Go test command is:

```sh
go test ./...
```
