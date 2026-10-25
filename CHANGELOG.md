# Changelog

All notable changes to this project are documented here.
Format based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/); versioning follows [Semantic Versioning](https://semver.org/).

## [Unreleased]

### Added
- `orch history <command> [args...]`: past calls of a command from zsh, bash and fish history, with `-n`, `--json` and `--file`.

## [0.1.1] - 2026-10-25

### Added
- Community files: `CONTRIBUTING.md`, `SECURITY.md`, `CODE_OF_CONDUCT.md`, issue and PR templates, Dependabot.
- golangci-lint (CI and `make lint`), race detector in CI, `make build|test|lint`.
- Shell completions (bash, zsh, fish) in release archives and the Homebrew formula; `orch --version` includes commit and date.
- README: sample output, origin story, limitations.
- Homebrew tap: `brew install tassan/tap/orch`, published by the release workflow.

## [0.1.0] - 2026-10-25

### Changed
- `source` names version managers and apps (`nvm`, `asdf`, `mise`, `pyenv`, `rbenv`, `pipx`, `app`, `local`) instead of `unknown`.
- `PATH` entries starting with `~` are expanded; other relative entries are skipped.
- `orch info` and `orch uninstall` also find tools in well-known dirs that are not on `PATH`.
- `orch uninstall` accepts a package name as well as a command name.
- `orch uninstall` prints the package and the commands on `PATH` that will go with it before asking to confirm.
- `Provider` interface now works on packages (`Owner`, `Lookup`, `UninstallCmd(Package)`) instead of single tools.
- `orch list` reports the owning package manager as `source` and adds a `package` field to `--json`.

### Added
- README badges: CI, Go Reference, Go version, platform, license.
- Providers for cargo, global npm (including nvm) and pipx.
- `shadows`: same-named commands hidden behind the first one on `PATH`, shown by `info` and `list --json`.
- `orch list --all` (scan well-known dirs outside `PATH`), `orch info --json`, `orch info --no-run`, and `orch --version`.
- CI (gofmt, vet, test) and a GoReleaser release workflow for macOS binaries.
- Tests for `info`, `scan` and the other providers.
- Homebrew cask support: casks are detected via `Caskroom` paths and `.app` bundles, and uninstalled with `brew uninstall --cask`.
- Tests for the CLI uninstall flow (native refusal, no owner, dry-run, abort) and the brew provider.
- `orch list` with `--native`, `--user` and `--json` filters: scans `PATH`, tags each tool as native or user-installed, and guesses the install source (brew, cargo, go, npm, manual).
- `orch info <name>`: path, resolved path, kind, owning package manager, version (`--version`, 2s timeout) and man-page description.
- `orch uninstall <name>` with `--dry-run` and `--yes`: delegates to the owning package manager; refuses native tools and never deletes files itself.
- Homebrew provider; the `Provider` interface allows adding more managers.
- Man page (`man/orch.1`) and `make install-man` / `make install` targets.
- Agent skill (`skills/orch/SKILL.md`).
- PolyForm Noncommercial 1.0.0 license.

[Unreleased]: https://github.com/tassan/orch/compare/v0.1.1...HEAD
[0.1.1]: https://github.com/tassan/orch/compare/v0.1.0...v0.1.1
[0.1.0]: https://github.com/tassan/orch/releases/tag/v0.1.0
