---
name: orch
description: Use the `orch` CLI to find out which command-line tools are installed on this Mac, whether they are native OS tools or user-installed, where they came from, and to uninstall user-installed ones through their package manager. Use when the user asks what is installed, where a command comes from, whether a tool is safe to remove, or wants to uninstall or clean up CLI tools.
---

# orch

`orch` inventories executables on `PATH`, marks each **native** (ships with macOS) or **user** (installed by the user), and uninstalls user tools via the package manager that owns them (currently Homebrew). It is not a package manager and never deletes files itself. macOS only.

Check it is installed with `command -v orch`. If missing: `brew install tassan/tap/orch` (installs the man page and shell completions too). Without Homebrew: `go install github.com/tassan/orch/cmd/orch@latest`. Check the version with `orch --version`.

## Commands

```
orch list [--native|--user] [--all] [--json]
orch info <name> [--json] [--no-run]
orch uninstall <command|package> [--dry-run] [--yes]
orch history [-n N] [--json] <command> [args...]
```

- `list --json` returns an array of `{name, path, real, kind, source, package}`. Prefer it over the table when you need to filter or count. `kind` is `native` or `user`; `source` (user tools only) is the owning package manager when one is found (`brew`, `cargo`, `npm`, `pipx`), otherwise a path-based guess (`nvm`, `asdf`, `mise`, `app`, `local`, `go`, `manual`, `unknown`). `package` (`{provider, name, cask}`) is set only when a package manager owns the tool.
- `info <name>` prints path, resolved path, kind, source, owner, shadows, version and a one-line description (`--json` for structured output). It runs `<name> --version` (2s timeout), which executes the tool; add `--no-run` for an untrusted binary.
- `list --all` also scans `~/.cargo/bin`, `~/go/bin`, `~/.local/bin` and similar dirs that are not on `PATH` (`off_path: true`).
- `real` is the symlink-resolved path. Ownership and kind come from `real`, not `path`.

- `history <command> [args...]` lists past calls of a command from the user's shell history (zsh, bash, fish), oldest first, last 20 by default. Use it to answer "when did I last use X" or "how did I call X" before suggesting an uninstall. Orch's flags go before the command (`orch history -n 5 git commit`). Shell history can contain secrets: show only what is needed, never copy the output into files, commits or external services.

## Uninstalling

Uninstalling changes the user's machine, so follow this order:

1. `orch info <name>` to confirm it is `kind: user` and has an `owner`.
2. `orch uninstall <name> --dry-run` and show the user the exact command and the `removes:` line. Uninstall takes the whole package, so those commands go too. The argument can be a command or a package name.
3. Only after the user agrees to that command, run `orch uninstall <name> --yes`. Without `--yes` it waits for a prompt on stdin and will hang or abort in a non-interactive shell.

Do not run `--yes` on your own initiative, and do not batch-uninstall from a list without the user approving each tool.

## Outcomes to expect

- `is a native OS tool; refusing to uninstall`: stop. Do not try to remove it another way.
- `no package manager owns <name>; remove manually: <path>`: report the path to the user. Do not `rm` it unless they explicitly ask.
- `neither a command on PATH nor an installed package`: the name is wrong.
- `no package manager owns` can also mean an app that is not a brew cask (for example Docker.app) or a version-manager shim (nvm, asdf). Report the path; do not guess a removal.

## Caveats

- `source` is a heuristic. A few user tools show `unknown` (for example symlinks outside the native directories); treat `unknown` as "ask the user", not as safe to remove.
- Only the first match on `PATH` is listed; later copies appear in `shadows`. Aliases and shell functions are invisible to `orch`.
- Only brew, cargo, npm and pipx can uninstall. Everything else (go, nvm, asdf, apps) is report-only.
