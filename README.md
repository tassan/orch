# orch

[![CI](https://github.com/tassan/orch/actions/workflows/ci.yml/badge.svg)](https://github.com/tassan/orch/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/tassan/orch.svg)](https://pkg.go.dev/github.com/tassan/orch)
[![Go Report Card](https://goreportcard.com/badge/github.com/tassan/orch)](https://goreportcard.com/report/github.com/tassan/orch)
![Go version](https://img.shields.io/github/go-mod/go-version/tassan/orch)
![Platform](https://img.shields.io/badge/platform-macOS-lightgrey)
[![License](https://img.shields.io/badge/license-PolyForm%20Noncommercial-blue)](LICENSE)

CLI tool to manage CLI tools. `orch` scans your `PATH`, marks every tool as **native** (ships with the OS) or **user** (installed by you), shows details about a tool, and uninstalls it through the package manager that owns it.

It is not a package manager. It delegates uninstalls to one that is already installed (currently Homebrew), and never deletes files itself.

Status: early. macOS only.

## Install

Requires Go 1.22+ (CI tests with the latest stable Go).

```
go install github.com/tassan/orch/cmd/orch@latest
```

Tagged releases (`v*`) publish macOS binaries (amd64, arm64) through GitHub Actions and GoReleaser.

## Usage

```
orch list [--native|--user] [--all] [--json]   # tools on PATH, with kind and source
orch info <name> [--json] [--no-run]         # path, kind, owner, shadows, version, description
orch uninstall <command|package> [--dry-run] [-y] # via the owning package manager
```

- `uninstall` takes a command on `PATH` or an installed package name (`orch uninstall svt-av1` and `orch uninstall SvtAv1EncApp` do the same). It removes the whole package, and first prints how many commands on `PATH` go with it.
- Homebrew formulae and casks are both detected. Casks are matched to their `.app` bundle, so tools inside apps like Cursor or VS Code resolve to their cask.
- `uninstall` refuses native OS tools.
- If no package manager owns a tool, `orch` prints its path and leaves it alone.
- `info` runs `<tool> --version` (2s timeout) to read the version, which executes the tool. Pass `--no-run` to skip that.
- Owners: Homebrew (formulae and casks), cargo, global npm (including nvm) and pipx. Other tools still show a source guess (`nvm`, `asdf`, `mise`, `app`, `local`, `go`...) but `uninstall` only prints their path. `go install` binaries have no uninstall command, so they are never owned.
- `list --all` also scans well-known dirs that are not on `PATH` (`~/.cargo/bin`, `~/go/bin`, `~/.local/bin`, `~/.bun/bin`, `~/.deno/bin`, `~/.dotnet/tools`).
- When a command exists in several `PATH` dirs, only the first is listed; `info` and `list --json` show the others as `shadows`.
- Not visible to `orch`: shell aliases, functions and builtins, which only the shell knows. A leading `~` in `PATH` is expanded; other relative entries are ignored.
- `orch --version` prints the version.

## Man page

```
make install-man   # installs man/orch.1 under $(brew --prefix)/share/man
man orch
```

## Agent skill

`skills/orch/SKILL.md` teaches coding agents how to use `orch` safely (dry-run first, ask before uninstalling). For Claude Code:

```
mkdir -p ~/.claude/skills && cp -r skills/orch ~/.claude/skills/
```

## Development

```
go build ./... && go vet ./... && go test ./...
```

## License

[PolyForm Noncommercial 1.0.0](LICENSE). Free for personal, educational, research and other noncommercial use. This is source-available, not open source: commercial use needs a separate paid agreement with the author. Open an issue at https://github.com/tassan/orch to ask.
