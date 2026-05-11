# monogit

<p align="center">
  <a href="https://github.com/JoaoOliveira889/monogit/releases/latest"><img src="https://img.shields.io/github/v/release/JoaoOliveira889/monogit?color=7aa2f7&label=release&logo=github&style=flat-square" alt="Latest Release"></a>
  <a href="https://github.com/JoaoOliveira889/monogit/actions/workflows/goreleaser.yml"><img src="https://img.shields.io/github/actions/workflow/status/JoaoOliveira889/monogit/goreleaser.yml?label=CI&logo=github-actions&style=flat-square" alt="CI Status"></a>
  <a href="https://go.dev/"><img src="https://img.shields.io/badge/go-1.23+-00ADD8?logo=go&style=flat-square" alt="Go Version"></a>
  <a href="LICENSE"><img src="https://img.shields.io/badge/license-MIT-green?style=flat-square" alt="License"></a>
</p>

**Multi-repo Git dashboard for your terminal.** A TUI tool that scans a root directory for Git repositories and gives you a panoramic view of branches, ahead/behind status, and dirty state — with one-key actions for fetch, pull, push, stash, and commit.

![Monogit Dashboard](https://joaooliveirablog.s3.us-east-1.amazonaws.com/SCR-20260511-hqmu.png)

Built with [Bubble Tea](https://github.com/charmbracelet/bubbletea), [Lip Gloss](https://github.com/charmbracelet/lipgloss), and [Bubbles](https://github.com/charmbracelet/bubbles).

---

## Features

- **Auto-scan**: Automatically discovers all Git repositories under the target directory.
- **Status dashboard**: Shows active branch, ahead/behind counters, and dirty-state indicators at a glance.
- **Auto-fetch**: Configurable background `git fetch --all` with customizable interval.
- **Concurrent operations**: All network operations run concurrently via goroutines for maximum performance.
- **Tokyo Night TUI**: Dark theme crafted with Lip Gloss.
- **Commit Wizard**: Interactive multi-step flow — add files → write message → pull → push.
- **Interactive Staging**: Pattern-based and toggle file staging for faster workflows.
- **Quick Undo**: One-key soft reset (`git reset --soft HEAD~1`) for the last commit.
- **Branch Manager**: List, create, and checkout local and remote branches.
- **Graph Log**: Switch between simple and `--graph` decorated commit history.
- **Command Log**: Inspect the raw output of every Git operation with `o`.

---

## Installation

### Option 1 — Pre-built binary (recommended)

Download the latest release for your platform from the [Releases page](https://github.com/JoaoOliveira889/monogit/releases/latest).

```bash
# macOS (Apple Silicon)
curl -LO https://github.com/JoaoOliveira889/monogit/releases/latest/download/monogit_Darwin_arm64.tar.gz
tar -xzf monogit_Darwin_arm64.tar.gz
sudo mv monogit /usr/local/bin/

# macOS (Intel)
curl -LO https://github.com/JoaoOliveira889/monogit/releases/latest/download/monogit_Darwin_x86_64.tar.gz
tar -xzf monogit_Darwin_x86_64.tar.gz
sudo mv monogit /usr/local/bin/

# Linux (amd64)
curl -LO https://github.com/JoaoOliveira889/monogit/releases/latest/download/monogit_Linux_x86_64.tar.gz
tar -xzf monogit_Linux_x86_64.tar.gz
sudo mv monogit /usr/local/bin/
```

### Option 2 — Install with `go install`

```bash
go install github.com/JoaoOliveira889/monogit/cmd/monogit@latest
```

> Requires Go 1.23 or later.

### Option 3 — Build from source

```bash
git clone https://github.com/JoaoOliveira889/monogit.git
cd monogit
go build -o monogit ./cmd/monogit
```

---

## Usage

```bash
# Scan current directory
monogit

# Scan a specific directory
monogit --path ~/projects

# Set auto-fetch interval to 10 minutes
monogit --interval 10m
```

### Flags

| Flag         | Default | Description                                |
|--------------|---------|--------------------------------------------|
| `--path`     | `.`     | Root directory to scan for Git repositories |
| `--interval` | `5m`    | Auto-fetch interval (e.g. `1m`, `10m`, `1h`) |

---

## Keybindings

### Global

| Key | Action |
|-----|--------|
| `↑` / `k` | Move cursor up |
| `↓` / `j` | Move cursor down |
| `tab` | Switch between panels |
| `ctrl+p` | Toggle Help Menu |
| `esc` | Back / Cancel / Close |
| `q` | Quit |

### Repository Panel

| Key | Action |
|-----|--------|
| `f` | Fetch selected repository |
| `F` | Fetch **all** repositories |
| `p` | Pull selected repository |
| `A` | Pull **all** clean repositories |
| `u` / `U` | Push selected / Push **all** |
| `c` | **Commit Wizard** (add → message → push) |
| `b` | List local & remote branches |
| `n` | Create new branch |
| `a` | Stage all modified files |
| `i` | **Interactive Stage** (toggle files by pattern) |
| `z` / `Z` | Stash / Stash Pop |
| `x` | **Quick Undo** (soft reset last commit) |
| `g` | Toggle Graph / Simple log view |
| `o` | Open Command Log |

---

## Layout

```
┌─────────────────────────┬──────────────────────────────┐
│  Repositories           │  ◈ api-gateway               │
│                         │                              │
│  ▸ api-gateway  main ↑2 │  Branch:  main               │
│    auth-svc     dev  ↓1 │  Ahead:   ↑ 2 commits       │
│    payment      main ✓  │  Behind:  0                  │
│    user-svc     feat ●  │  Status:  Modified ●         │
│                         │                              │
│                         │  Recent Commits:             │
│                         │  ─────────────────           │
│                         │  a1b2c3d Fix auth            │
│                         │  d4e5f6a Add rate limit      │
│                         │  g7h8i9j Update deps         │
└─────────────────────────┴──────────────────────────────┘
 ↑↓/jk nav │ f fetch │ F fetch all │ p pull │ A pull all │ c commit │ q quit
```

---

## Architecture

The project follows **Clean Architecture** principles, keeping business logic decoupled from implementation details:

```
monogit/
├── cmd/monogit/        # Entry point
├── internal/
│   ├── domain/         # Core entities: Repository, FileStatus, interfaces
│   ├── usecase/        # Business logic: Git operations, repo scanning
│   ├── adapters/
│   │   ├── git/        # CLI Git provider (exec-based, no shell injection)
│   │   └── tui/        # Bubble Tea UI: model, update, view, keys
│   └── pkg/
│       └── ui/         # Shared styles (Lip Gloss tokens)
└── .goreleaser.yaml    # Multi-platform release config
```

**Security note:** All Git commands are built using `exec.Command` with individual string arguments — no shell interpolation, no injection vectors.

---

## Contributing

Contributions are welcome! Please read [CONTRIBUTING.md](CONTRIBUTING.md) before opening a pull request.

1. Fork the repository
2. Create a feature branch: `git checkout -b feat/my-feature`
3. Commit your changes following [Conventional Commits](https://www.conventionalcommits.org/)
4. Push and open a Pull Request

---

## Support

If monogit helps you manage repositories more efficiently, consider supporting its development.

<a href="https://www.buymeacoffee.com/JoaoOliveira889" target="_blank"><img src="https://cdn.buymeacoffee.com/buttons/v2/default-yellow.png" alt="Buy Me a Coffee" style="height: 60px !important;width: 217px !important;" ></a>

---

## License

[MIT](LICENSE) © João Oliveira
