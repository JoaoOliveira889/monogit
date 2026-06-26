# Monogit Rules

These are project-specific rules for Monogit. For shared workspace rules, see [../RULES.md](../RULES.md).

## Project-specific
- All Git operations must go through `internal/adapters/git/` using `exec.Command` with discrete arguments.
- The TUI model lives in `internal/adapters/tui/`.
- Repo scanning lives in `internal/pkg/scanner/`.
- Editor detection lives in `internal/pkg/editor/`.
- Shared UI styles live in `internal/pkg/ui/`.
- Build: `go build -o monogit ./cmd/monogit`. No Makefile — use `go build` directly.
- Release: `.goreleaser.yaml` handles multi-platform builds, Homebrew tap, and changelog generation.
