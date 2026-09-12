# Monogit Rules

These are project-specific rules for Monogit. For shared workspace rules, see [../RULES.md](../RULES.md).

## Project-specific
- All Git operations must go through `internal/adapters/git/` using `exec.Command` with discrete arguments.
- The TUI model lives in `internal/adapters/tui/`.
- Repo scanning lives in `internal/pkg/scanner/`.
- Editor detection lives in `internal/pkg/editor/`.
- Shared UI styles live in `internal/pkg/ui/`.
- Build: `go build -o monogit ./cmd/monogit`. No Makefile — use `go build` directly.
- The TUI targets Bubble Tea v2 (`charm.land/bubbletea/v2`), Lip Gloss v2 and Bubbles v2.
- Detail views are a single `Model.detailView` value and modals a single `Model.overlays` stack. Do not reintroduce parallel `showX` booleans.
- `refreshViewports` marks panels stale; only `View` renders them. Never render a panel from a message handler.
- `cmd/monogit` accepts an internal `-rebase-todo` flag used as `GIT_SEQUENCE_EDITOR`. It is not a user-facing flag.
- Release: `.goreleaser.yaml` handles multi-platform builds, Homebrew tap, and changelog generation.
