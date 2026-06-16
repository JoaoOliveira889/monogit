# Configuration & Environment

Monogit works out of the box, but you can customize it to fit your workflow.

## 🚩 Command Line Flags

| Flag | Description | Default |
|------|-------------|---------|
| `--path` | Root directory to scan for Git repos. | `.` |
| `--interval` | Frequency of background auto-fetches. | `5m` |
| `--version` | Show version, commit, and build date. | - |

**Example:**
```bash
monogit --path ~/work --interval 1h
```

---

## 💻 Environment Variables

### `$MONOGIT_EDITOR`
If set, Monogit will always use this command to open repositories when you press `e`. Command arguments are supported, so values like `code --reuse-window` work as expected.

### `$VISUAL`
If `$MONOGIT_EDITOR` is not set, Monogit falls back to `$VISUAL` before checking `$EDITOR`.

### `$EDITOR`
If neither `$MONOGIT_EDITOR` nor `$VISUAL` is set, Monogit falls back to your system's default terminal editor (e.g., `nvim`, `vim`, `nano`).

### `$TERM_PROGRAM`
Monogit uses this to detect if you are running in **iTerm2** or **Ghostty** to provide better tab/window management when opening terminal editors.

---

## 🗂️ Local Config File

Monogit stores its panel layout preference, mergetool setting, scan exclusions, and local repository tags in `~/.config/monogit/config.json` and writes it with restrictive file permissions. It also keeps a lightweight startup cache in `~/.config/monogit/startup_cache.json` so the repo list can appear faster on launch, and writes structured operational logs to `~/.config/monogit/monogit.log` (rotated at 2MB). Updating the binary does not remove these files Updating the binary does not remove these files, so local tags and the startup cache survive normal app upgrades unless they are deleted.

### `merge_tool`

The mergetool to use when resolving merge conflicts with `m`. If empty, Monogit delegates to the system's `git mergetool` configuration (`git config merge.tool`). Set this to any mergetool your system supports (e.g., `vimdiff`, `nvimdiff`, `meld`, `opendiff`, `smerge`).

### `scan_excludes`

Directory names that Monogit should skip during repository discovery. This is useful for large workspaces that contain generated output, package caches, or vendor trees that should never be scanned.

Example `config.json`:

```json
{
  "left_panel_ratio": 0.30,
  "repo_tags": {},
  "merge_tool": "nvimdiff",
  "scan_excludes": ["node_modules", "vendor", ".git", "dist", "build"]
}
```

---

## 🔌 Editor Auto-Detection

If no environment variables are set, Monogit attempts to find these editors in your system (macOS/Linux):
1.  **VS Code** (`code`)
2.  **Cursor** (`cursor`)
3.  **Zed** (`zed`)
4.  **Sublime Text** (`subl`)
5.  **Vim / Neovim / Nano / Emacs**
6.  **Other installed GUI editors detected from the local system**

---

## 🕒 Auto-Fetch Logic

Background fetches run on the configured interval and refresh repository state in the background so the TUI stays responsive while the local metadata is updated.

## 🪪 Footer Conventions

Monogit always shows the current application version in the bottom-right footer area and keeps `? help` visible as a global shortcut to the shortcuts modal.
