# orch

[![CI](https://github.com/tassan/orch/actions/workflows/ci.yml/badge.svg)](https://github.com/tassan/orch/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/tassan/orch)](https://github.com/tassan/orch/releases/latest)
[![Go Reference](https://pkg.go.dev/badge/github.com/tassan/orch.svg)](https://pkg.go.dev/github.com/tassan/orch)
![Go version](https://img.shields.io/github/go-mod/go-version/tassan/orch)
![Platform](https://img.shields.io/badge/platform-macOS-lightgrey)
[![License](https://img.shields.io/badge/license-PolyForm%20Noncommercial-blue)](LICENSE)

A CLI for the CLI tools you installed and forgot about.

`orch` scans your `PATH`, marks every tool as **native** (ships with macOS) or **user** (you installed it), tells you which package manager owns it, uninstalls it through that package manager, and shows when you last used it from your shell history. It is not a package manager: it never deletes files itself.

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

$ orch history -n 3 orch
2026-10-25 15:16  orch info orch
2026-10-25 15:26  orch history terraform
2026-10-25 15:26  orch history -n 60 terraform
showing 3 of 17 calls of "orch" from .zsh_history, .bash_history
```

## Why this exists

I install CLI tools constantly, across projects and studies, and forget most of them. A year later I have hundreds of commands and no idea which came with macOS, which I installed, or whether I used brew, cargo, npm or something else. `orch` answers that, shows when I last actually ran a tool, and removes things without making me remember how I installed them.

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

Upgrade with `brew upgrade orch` (or rerun `go install`). `orch --version` shows what you have.

## Usage

```
orch list [--native|--user] [--all] [--json]       # tools on PATH, with kind and source
orch info <name> [--json] [--no-run]               # path, kind, owner, shadows, version, description
orch uninstall <command|package> [--dry-run] [-y]  # via the owning package manager
orch history [-n N] [--json] <command> [args...]   # past calls of a command, from your shell history
orch --version
```

**Uninstall**
- Takes a command on `PATH` or an installed package name: `orch uninstall svt-av1` and `orch uninstall SvtAv1EncApp` do the same.
- Removes the whole package, and first prints how many commands on `PATH` go with it.
- Refuses native macOS tools. If no package manager owns a tool, it prints the path and leaves it alone.
- Asks for confirmation unless you pass `-y`; `--dry-run` only prints the command.

**Owners.** Homebrew (formulae and casks, matched to their `.app` bundle), cargo, global npm (including nvm) and pipx. Other tools still show a source guess (`nvm`, `asdf`, `mise`, `app`, `local`, `go`...), but `uninstall` only prints their path. `go install` binaries have no uninstall command, so they are never owned.

**History**
- `orch history git commit` lists past calls that start with `git commit`, oldest first (last 20 by default; `-n 0` for all). With no args it matches every call of the command.
- Reads zsh, bash and fish history (`$HISTFILE`, `~/.zsh_history`, `~/.bash_history`, fish's `fish_history`), or files you pass with `--file`. Timestamps show when the shell recorded them.
- Matches the command wherever one can start: after `FOO=1`, `sudo`, `time`, `env`, in pipelines and after `&&` or `;`. `echo git status` is not a call of `git`.
- Put orch's own flags before the command: `orch history -n 5 ls -la` searches for `ls -la`.
- Shells often write history only when they exit, so the current session may be missing. History can contain secrets typed on the command line; `orch` only prints it to your terminal.

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

`skills/orch/SKILL.md` teaches coding agents how to use `orch` safely (dry-run first, ask before uninstalling, careful with history output). For Claude Code:

```
mkdir -p ~/.claude/skills && cp -r skills/orch ~/.claude/skills/
```

## Limitations

macOS only. Only brew, cargo, npm and pipx can uninstall. The `source` column is a path-based guess when no package manager owns the tool. `history` can only see what your shell wrote to its history files.

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) and [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md). Architecture in one page: [CLAUDE.md](CLAUDE.md).

## License

[PolyForm Noncommercial 1.0.0](LICENSE). Free for personal, educational, research and other noncommercial use. This is source-available, not open source: commercial use needs a separate paid agreement with the author. [Open an issue](https://github.com/tassan/orch/issues/new?title=Commercial+license&labels=licensing) to ask.
