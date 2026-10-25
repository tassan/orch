# orch

CLI tool to manage CLI tools. `orch` scans your `PATH`, marks every tool as **native** (ships with the OS) or **user** (installed by you), shows details about a tool, and uninstalls it through the package manager that owns it.

It is not a package manager. It delegates uninstalls to one that is already installed (currently Homebrew), and never deletes files itself.

Status: early. macOS only.

## Install

Requires Go 1.22+.

```
go install github.com/tassan/orch/cmd/orch@latest
```

## Usage

```
orch list [--native|--user] [--json]   # tools on PATH, with kind and source
orch info <name>                       # path, kind, owner, version, description
orch uninstall <name> [--dry-run] [-y] # via the owning package manager
```

- `uninstall` refuses native OS tools.
- If no package manager owns a tool, `orch` prints its path and leaves it alone.
- `info` runs `<tool> --version` (2s timeout) to read the version, which executes the tool.

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
