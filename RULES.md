# Monogit Rules

These rules apply to Monogit specifically.

## Security
- No telemetry, analytics, or hidden data collection.
- Every mutating Git action must present a confirmation modal before execution.
- Use `exec.Command` with explicit arguments only.
- Validate paths, branch names, tags, and patterns before passing them to Git.
- Store local config with restrictive permissions.

## UX
- Keep the TUI dense, responsive, and terminal-first.
- Use centered modals for confirmations and destructive actions.
- Follow the `monostack` visual pattern for panels, headers, and status messages.
- Never let text overlap or force broken layouts on smaller terminals.

## Architecture
- Keep domain, use case, adapter, and package boundaries clean.
- Do not block `Update` with IO or Git work.
- Use background commands and messages for expensive work.
- Keep docs, tests, and keybindings in sync with behavior.
