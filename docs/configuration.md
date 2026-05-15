# Configuration & Environment

Monogit works out of the box, but you can customize it to fit your workflow.

## 🚩 Command Line Flags

| Flag | Description | Default |
|------|-------------|---------|
| `--path` | Root directory to scan for Git repos. | `.` |
| `--interval` | Frequency of background auto-fetches. | `5m` |
| `--version` | Show the current version of Monogit. | - |

**Example:**
```bash
monogit --path ~/work --interval 1h
```

---

## 💻 Environment Variables

### `$MONOGIT_EDITOR`
If set, Monogit will always use this command to open repositories when you press `e`.

### `$EDITOR`
If `$MONOGIT_EDITOR` is not set, Monogit falls back to your system's default terminal editor (e.g., `nvim`, `vim`, `nano`).

### `$TERM_PROGRAM`
Monogit uses this to detect if you are running in **iTerm2** or **Ghostty** to provide better tab/window management when opening terminal editors.

---

## 🔌 Editor Auto-Detection

If no environment variables are set, Monogit attempts to find these editors in your system (macOS/Linux):
1.  **VS Code** (`code`)
2.  **Cursor** (`cursor`)
3.  **Zed** (`zed`)
4.  **Sublime Text** (`subl`)
5.  **IntelliJ IDEA**
6.  **WebStorm**

---

## 🕒 Auto-Fetch Logic

Background fetches only happen for the **active branch** of each repository to minimize network usage. If a repository has a clean state, Monogit will periodically run `git fetch --all` to update the ahead/behind counters.
