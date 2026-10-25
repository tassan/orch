---
name: orch
description: Use the `orch` CLI to find out which command-line tools are installed on this Mac, whether they are native OS tools or user-installed, where they came from, and to uninstall user-installed ones through their package manager. Use when the user asks what is installed, where a command comes from, whether a tool is safe to remove, or wants to uninstall or clean up CLI tools.
---

# orch

`orch` inventories executables on `PATH`, marks each **native** (ships with macOS) or **user** (installed by the user), and uninstalls user tools via the package manager that owns them (currently Homebrew). It is not a package manager and never deletes files itself. macOS only.

Check it is installed with `command -v orch`. If missing: `go install github.com/tassan/orch/cmd/orch@latest`.

## Commands

```
orch list [--native|--user] [--json]
orch info <name>
orch uninstall <name> [--dry-run] [--yes]
```

- `list --json` returns an array of `{name, path, real, kind, source}`. Prefer it over the table when you need to filter or count. `kind` is `native` or `user`; `source` (user tools only) is a path-based guess: `brew`, `cargo`, `go`, `npm`, `manual` or `unknown`.
- `info <name>` prints path, resolved path, kind, source, owner, version and a one-line description. It runs `<name> --version` (2s timeout), which executes the tool.
- `real` is the symlink-resolved path. Ownership and kind come from `real`, not `path`.

## Uninstalling

Uninstalling changes the user's machine, so follow this order:

1. `orch info <name>` to confirm it is `kind: user` and has an `owner`.
2. `orch uninstall <name> --dry-run` and show the user the exact command it prints.
3. Only after the user agrees to that command, run `orch uninstall <name> --yes`. Without `--yes` it waits for a prompt on stdin and will hang or abort in a non-interactive shell.

Do not run `--yes` on your own initiative, and do not batch-uninstall from a list without the user approving each tool.

## Outcomes to expect

- `is a native OS tool; refusing to uninstall`: stop. Do not try to remove it another way.
- `no package manager owns <name>; remove manually: <path>`: report the path to the user. Do not `rm` it unless they explicitly ask.
- `not found on PATH`: the name is wrong or the tool is not on `PATH`.
- Some brew formulas provide several commands. Uninstalling one removes the whole formula, so the other commands go too. Mention this when `real` points into `Cellar/<formula>/`.

## Caveats

- `source` is a heuristic. A few user tools show `unknown` (for example symlinks outside the native directories); treat `unknown` as "ask the user", not as safe to remove.
- Only the first match on `PATH` is listed. A shadowed copy in a later directory will not appear.
