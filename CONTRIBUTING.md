# Contributing

Thanks for looking. orch is small on purpose.

## Setup

```
go build ./... && go vet ./... && go test -race ./...
golangci-lint run          # same config as CI
```

`make test` and `make lint` run the same things. See `CLAUDE.md` for the architecture in one page.

## Pull requests

- Open an issue first for anything bigger than a bug fix.
- Add or update tests. Providers are tested with fakes (see `internal/provider/others_test.go`); CLI behavior with a temp `PATH` (see `internal/cli/cli_test.go`).
- Use [Conventional Commits](https://www.conventionalcommits.org/) (`feat:`, `fix:`, `docs:`, `ci:`...).
- Add a line to `CHANGELOG.md` under `[Unreleased]`.
- Keep the safety rules: `uninstall` refuses native tools, never deletes files itself, and stays dry-run friendly.

## Adding a package manager

Implement `provider.Provider` in `internal/provider` (`Owner`, `Lookup`, `UninstallCmd`, `Available`) and add it to `All()`. `cargo.go` is the smallest example. Detect ownership from the resolved path or the manager's own data, and load anything expensive lazily, once.

## License of contributions

orch is licensed under [PolyForm Noncommercial 1.0.0](LICENSE) and the author offers separate commercial licenses. By submitting a contribution you agree that it is licensed under the same terms and that you grant the author a perpetual, worldwide, royalty-free right to also license it under other terms, including commercial ones. If that does not work for you, please do not submit code.

## Conduct

See [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md).
