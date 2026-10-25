# Security Policy

## Supported versions

Only the latest release gets fixes.

## Reporting a vulnerability

Please do not open a public issue. Use GitHub's private reporting:
<https://github.com/tassan/orch/security/advisories/new>

Include what you ran, what you expected and what happened. Expect a first reply within a week; this is a one-person project.

## What to know about orch's attack surface

- `orch info <name>` executes `<name> --version` (2 second timeout, no shell). That runs whatever binary is first on your `PATH`. Use `--no-run` to skip it.
- `orch uninstall` runs a package manager (`brew`, `cargo`, `npm`, `pipx`) with arguments built from package names it read from the package manager's own data. It never invokes a shell and never deletes files itself.
- Release binaries are not signed or notarized.
