# orch

[![CI](https://github.com/tassan/orch/actions/workflows/ci.yml/badge.svg)](https://github.com/tassan/orch/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/tassan/orch)](https://github.com/tassan/orch/releases/latest)
[![Go Reference](https://pkg.go.dev/badge/github.com/tassan/orch.svg)](https://pkg.go.dev/github.com/tassan/orch)
![Go version](https://img.shields.io/github/go-mod/go-version/tassan/orch)
![Platform](https://img.shields.io/badge/platform-macOS-lightgrey)
[![License](https://img.shields.io/badge/license-PolyForm%20Noncommercial-blue)](LICENSE)

A CLI for the CLI tools you installed and forgot about.

`orch` scans your `PATH`, marks every tool as **native** (ships with macOS) or **user** (you installed it), tells you which package manager owns it, and uninstalls it through that package manager. It is not a package manager: it never deletes files itself.

```
$ orch list --user
NAME          KIND  SOURCE  PATH
SvtAv1EncApp  user  brew    /opt/homebrew/bin/SvtAv1EncApp
code          user  brew    /opt/homebrew/bin/code
git           user  brew    /opt/homebrew/bin/git
homebridge    user  npm     /Users/you/.nvm/versions/node/v24.15.0/bin/homebridge
...

$ orch info git --no-run
name:     git
path:     /opt/homebrew/bin/git
real:     /opt/homebrew/Cellar/git/2.54.0/bin/git
kind:     user
source:   brew
owner:    brew formula git
shadows:  /usr/bin/git
about:    git(1)                   - the stupid content tracker

$ orch uninstall nss --dry-run
package:  brew formula nss
removes:  62 command(s) on PATH: addbuiltin, atob, baddbdir, bltest, btoa, certutil, chktest, cmsutil, crlutil, crmftest, +52 more
would run: brew uninstall --formula nss
```

## Why this exists

I install CLI tools constantly, across projects and studies, and forget most of them. A year later I have hundreds of commands and no idea which came with macOS, which I installed, or whether I used brew, cargo, npm or something else. `orch` answers that, and removes things without making me remember how I installed them.

## Vibecoded into oblivion, lol

This whole project was vibecoded: written with an AI coding assistant, steered by me, built to help my chaotic mind deal with all the CLI tools I pile up. In its defense, it has unit tests, CI with linting, a dry-run for every uninstall, and it refuses to touch macOS's own tools. Still, treat it like what it is: read the code, run `--dry-run` first, and don't point `--yes` at something you care about without looking.

## Install

Homebrew (installs the binary, the man page and shell completions):

```
brew install tassan/tap/orch
```

Or with Go 1.22+ (CI tests with the latest stable Go):

```
go install github.com/tassan/orch/cmd/orch@latest
```

Tagged releases publish macOS binaries (amd64, arm64). They are not signed or notarized, so a browser-downloaded copy may be blocked by Gatekeeper; clear it with `xattr -d com.apple.quarantine orch`. The Homebrew and `go install` routes are not affected.

## Usage

```
orch list [--native|--user] [--all] [--json]       # tools on PATH, with kind and source
orch info <name> [--json] [--no-run]               # path, kind, owner, shadows, version, description
orch uninstall <command|package> [--dry-run] [-y]  # via the owning package manager
orch --version
```

**Uninstall**
- Takes a command on `PATH` or an installed package name: `orch uninstall svt-av1` and `orch uninstall SvtAv1EncApp` do the same.
- Removes the whole package, and first prints how many commands on `PATH` go with it.
- Refuses native macOS tools. If no package manager owns a tool, it prints the path and leaves it alone.
- Asks for confirmation unless you pass `-y`; `--dry-run` only prints the command.

**Owners.** Homebrew (formulae and casks, matched to their `.app` bundle), cargo, global npm (including nvm) and pipx. Other tools still show a source guess (`nvm`, `asdf`, `mise`, `app`, `local`, `go`...), but `uninstall` only prints their path. `go install` binaries have no uninstall command, so they are never owned.

**What it sees**
- `list --all` also scans well-known dirs that are not on `PATH` (`~/.cargo/bin`, `~/go/bin`, `~/.local/bin`, `~/.bun/bin`, `~/.deno/bin`, `~/.dotnet/tools`).
- When a command exists in several `PATH` dirs, only the first is listed; the others show as `shadows`.
- A leading `~` in `PATH` is expanded; other relative entries are ignored.
- Shell aliases, functions and builtins are invisible: only the shell knows them.

**Safety.** `info` runs `<tool> --version` (2s timeout, no shell) to read the version, which executes the tool; pass `--no-run` to skip that. See [SECURITY.md](SECURITY.md).

## Man page and completions

`brew install tassan/tap/orch` installs the man page and the bash, zsh and fish completions, so `man orch` works right away.

Building from source instead? `make install-man` copies `man/orch.1` under `$(brew --prefix)/share/man`, and `orch completion zsh` (also `bash`, `fish`) prints a completion script. Don't run `make install-man` if you installed with Homebrew: the copied file blocks `brew link`.

## Agent skill

`skills/orch/SKILL.md` teaches coding agents how to use `orch` safely (dry-run first, ask before uninstalling). For Claude Code:

```
mkdir -p ~/.claude/skills && cp -r skills/orch ~/.claude/skills/
```

## Limitations

macOS only. Only brew, cargo, npm and pipx can uninstall. The `source` column is a path-based guess when no package manager owns the tool.

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) and [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md). Architecture in one page: [CLAUDE.md](CLAUDE.md).

## License

[PolyForm Noncommercial 1.0.0](LICENSE). Free for personal, educational, research and other noncommercial use. This is source-available, not open source: commercial use needs a separate paid agreement with the author. [Open an issue](https://github.com/tassan/orch/issues/new?title=Commercial+license&labels=licensing) to ask.
