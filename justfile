set dotenv-load := false

default:
    @just --list

fmt:
    gofmt -w .
    nix fmt flake.nix

test:
    go test ./...

lint:
    golangci-lint run ./...

hooks:
    prek run --all-files

install-hooks:
    prek install
