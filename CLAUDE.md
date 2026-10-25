# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this is

`orch`: Go CLI that inventories CLI tools on PATH, tags each **native** (ships with OS) or **user** (installed), shows info, and uninstalls by delegating to the owning package manager (currently brew). It is NOT a package manager. v1 targets macOS only.

## Commands

```
go build ./... && go vet ./... && go test ./...
go test ./internal/classify -run TestClassify   # single test
go run ./cmd/orch list [--native|--user] [--json]
go run ./cmd/orch info <name>
go run ./cmd/orch uninstall <name> [--dry-run] [--yes]
```

## Architecture

Pipeline: `scan` (PATH walk, records `Shadows`) -> `classify` (native/user + source guess) -> `provider` (ownership + uninstall) ; `info` probes a tool; `internal/cli` wires cobra commands and builds the `entry` view model.

- `scan`: first PATH dir wins on duplicate names; records both `Path` and symlink-resolved `Real`. All classification and ownership use `Real`, never `Path` (brew bins in `/opt/homebrew/bin` are symlinks into `Cellar`).
- `classify`: native = resolved path under fixed macOS prefixes (`/usr/bin`, `/bin`, `/usr/sbin`, `/usr/libexec`, `/sbin`, `/System`, `/Library/Apple`). `Source` is a path-based heuristic for display only.
- `provider.Provider` interface (`Owner`, `Lookup`, `UninstallCmd`, `Available`) works on `Package` values: add new managers (cargo, npm...) in `internal/provider` and register in `All()`. Providers return an `*exec.Cmd`; the CLI owns dry-run/confirm/execution. `cli.providers` is a var so tests inject a fake.
- Providers: brew, cargo (`$CARGO_HOME/.crates2.json`), npm (`/lib/node_modules/<pkg>`), pipx (`/pipx/venvs/<pkg>`). Go has no uninstall command, so it has no provider. Owner order in `All()` matters: brew claims `/Cellar/` paths first.
- Brew ownership: formula from `.../Cellar/<formula>/...`; cask from `/Caskroom/<token>/` or from the `.app` bundle in the path, matched against `brew info --cask --json=v2 --installed` (loaded lazily, once, about 0.5s).
- `uninstall` resolves its argument as a PATH command first, then as an installed package name, and lists the other PATH commands the package provides before confirming.

## Safety rules (keep)

- `uninstall` refuses native tools; with no owning provider it only prints the path, never deletes files.
- `info` probes `--version` only, with a 2s timeout and no shell. Other flags (`-v`, `-V`, `version`) have side effects or garbage output on some tools.
