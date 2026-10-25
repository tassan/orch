# Changelog

All notable changes to this project are documented here.
Format based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/); versioning follows [Semantic Versioning](https://semver.org/).

## [Unreleased]

### Added
- `orch list` with `--native`, `--user` and `--json` filters: scans `PATH`, tags each tool as native or user-installed, and guesses the install source (brew, cargo, go, npm, manual).
- `orch info <name>`: path, resolved path, kind, owning package manager, version (`--version`, 2s timeout) and man-page description.
- `orch uninstall <name>` with `--dry-run` and `--yes`: delegates to the owning package manager; refuses native tools and never deletes files itself.
- Homebrew provider; the `Provider` interface allows adding more managers.
- Man page (`man/orch.1`) and `make install-man` / `make install` targets.
- PolyForm Noncommercial 1.0.0 license.

[Unreleased]: https://github.com/tassan/orch/commits/main
