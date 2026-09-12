# Keybindings Reference

Monogit is designed to be fully keyboard-driven. Below is a comprehensive list of all available shortcuts.

## 🌍 Global Navigation

| Key | Description |
|-----|-------------|
| `tab \| ctrl+w w` | Cycle focus between visible panels |
| `ctrl+w 1` | Focus Repositories Panel |
| `ctrl+w 2` | Focus Branch/Commit List Panel |
| `ctrl+w 3` | Focus Diff/Output Panel |
| `h \| ←` | Focus Left Panel |
| `l \| →` | Focus Right Panel |
| `j \| ↓` | Move cursor down |
| `k \| ↑` | Move cursor up |
| `{count}` + motion | Repeat motion, e.g. `5j` moves down five rows |
| `gg \| home` | Jump to top / first repository |
| `G \| end` | Jump to bottom / last repository |
| `ctrl+d \| pgdown` | Scroll half-page down in lists / diffs |
| `ctrl+u \| pgup` | Scroll half-page up in lists / diffs |
| `}` | Next repository needing attention (dirty, ahead, behind, conflicted) |
| `{` | Previous repository needing attention |
| `ctrl+o` | Retrace cursor back in jump list |
| `ctrl+i` | Retrace cursor forward in jump list |
| `.` | Repeat last action on selected repository |
| `:` | Open **Command Palette** |
| `v` | Start/stop selection range |
| `y` | Copy selection to clipboard |
| `< \| >` | Resize left panel (smaller / larger) |
| `ctrl+p \| ?` | Toggle interactive Help Menu |
| `esc` | Back / Cancel / Close Modal |
| `q \| ctrl+c` | Quit Monogit |

The footer always keeps `? help` and the current `MonoGit` version in the bottom-right corner, including modal and panel modes. It shows only the most relevant contextual hints on the left; open help for the complete key map.

The repository list also shows compact health badges:
- `DET`: detached `HEAD`
- `UP`: no upstream tracking branch
- `CF`: merge conflicts present
- `ST`: stale branch
- `TG`: local tag on `HEAD` still not pushed to `origin`

---

## 🔍 Diff Viewer Shortcuts

Inside the full-screen diff viewer (`d`):

| Key | Description |
|-----|-------------|
| `j \| k` / `↓ \| ↑` | Navigate between changed files |
| `J \| K` | Jump to next / previous diff hunk |
| `ctrl+d \| ctrl+u` | Scroll diff viewport |
| `gg \| G` | Jump to top / bottom of diff |
| `space` | Stage or unstage selected file |
| `x` | Discard file changes (confirmation required) |
| `y` | Copy diff to clipboard |
| `d \| esc` | Close diff viewer |

---

## 📂 Repository Panel Actions

When the focus is on the left list:

| Key | Description |
|-----|-------------|
| `enter \| l` | Focus the repository **Details & Commits** panel |
| `f` | **Fetch** only the selected repository |
| `F` | **Fetch All** repositories concurrently |
| `p` | **Pull** selected repository (confirmation required) |
| `P` | **Pull All** repositories (confirmation required) |
| `u` | **Push** selected repository (confirmation required) |
| `U` | **Push All** repositories (confirmation required) |
| `c` | Start the **Commit Wizard** |
| `d` | Open the full-screen **Diff Viewer** |
| `t` | Start the **Tag/Deploy Wizard** |
| `/` | Open the repository search field in the left panel |
| `ctrl+f` | Open the **Status Filter** modal (All, Dirty, Behind, Ahead, Conflicts, Tagged) |
| typing | Filter repositories live while the search field is open |
| `enter` | Keep the current repo search active |
| `esc` | Close search and restore the previously applied filter |
| `ctrl+g` | Filter repositories by tags |
| `ctrl+t` | Edit repo tags in the right panel |
| `d` | Remove the selected tag in the tag editor with confirmation |
| `b` | Open **Branch Manager** |
| `e` | Open current repo in your default **Editor** |
| `w` | Open current repo in your **Web Browser** |
| `s` | **Stash** changes (confirmation required) |
| `S` | Open **Stash Panel** |
| `Z` | **Stash All** filtered repositories — stashes changes in every dirty visible repo (confirmation required) |
| `B` | **Checkout All** filtered repositories — prompts for a branch name, then checks it out in every visible repo (confirmation required) |
| `m` | **Resolve merge conflicts** — lists conflicting files and opens the configured mergetool |
| `z` | **Undo** (Soft reset the last commit) |
| `R` | **Interactive Rebase** — edit, reorder, squash, or drop commits |
| `ctrl+y` | **Cherry-pick** a commit by hash |
| `ctrl+r` | **Revert** a commit by hash |
| `,` | Open the **Configuration Panel** |
| `gl` | Toggle between **Graph** and **Simple** log views |
| `o` | Open the temporary **Command Log** to see raw output |
| `E` | **Export** command log after confirmation (only inside the Command Log panel) |

---

## 🌿 Branch Manager Shortcuts

Inside the branch list (`b`):

| Key | Description |
|-----|-------------|
| `enter` | **Checkout** selected branch — or **Open terminal** at the worktree path if the branch is active in a linked worktree (confirmation required) |
| `M` | **Merge** selected branch into current HEAD |
| `n` | Create a **New** branch |
| `d` | **Delete** selected branch. For worktree branches, allows removing the linked worktree and deleting the branch |
| `esc` | Return to repository list |

---

## 📝 Commit Wizard Shortcuts

Commit wizard entry:

| Key | Description |
|-----|-------------|
| `a` | **Add All** files and go straight to the commit message |
| `v` | Open manual file selection |
| `esc` | Cancel the commit wizard |

Inside the staging screen:

| Key | Description |
|-----|-------------|
| `space` | **Toggle** file selection |
| `a` | **Select All** files |
| `n` | **Deselect All** files |
| `x` | **Discard** changes in file (confirmation required) |
| `tab` | View **Diff** for the selected file |
| `C` | Toggle **compact diff** mode (shows only changed functions/classes) |
| `enter` | Confirm the local selection and move to **Commit Message** |
| `v` | Start a selection range in list-based panels |
| `y` | Copy the selected content to the clipboard |
| `ctrl+v` | Paste clipboard content into text inputs |

---

## 📦 Stash Panel Shortcuts

Inside the stash list (`S`):

| Key | Description |
|-----|-------------|
| `p | enter` | **Pop** selected stash (confirmation required) |
| `a` | **Apply** selected stash (confirmation required) |
| `d` | **Drop** selected stash (confirmation required) |
| `esc` | Return to repository list |

---

## 🔀 Interactive Rebase Shortcuts

Inside the interactive rebase panel (`R`):

| Key | Description |
|-----|-------------|
| `p | P` | Mark selected commit as **Pick** |
| `s | S` | Mark selected commit as **Squash** |
| `f | F` | Mark selected commit as **Fixup** |
| `r` | Mark selected commit as **Reword** |
| `d | D` | Mark selected commit as **Drop** |
| `J | K` / `Shift+↓/↑` | **Reorder** commit down / up |
| `enter` | **Execute** interactive rebase |
| `esc` | Cancel rebase |

---

## 🔍 Shortcuts Help Modal

Press `?` or `ctrl+p` anywhere to open the interactive shortcuts help modal.
- **Search & Filter**: Type immediately to filter shortcuts by key name, action description, or category in real time.
- **Clear & Close**: Press `esc` to clear the search query, or press `esc` (or `?`/`ctrl+p`) again to close the modal.
- **Scroll**: Use `↑/↓`, `pgup/pgdown`, or mouse wheel to scroll through the cheat sheet.

---

## Confirmation Modal

When a mutating action is triggered, Monogit shows a centered confirmation modal. `y` or `enter` accepts the action, `n` or `esc` cancels it, and branch deletion also supports `l` for local and `r` for remote. Fetch is direct and does not prompt. Commit wizard file selection stays local until the final commit confirmation.

## ⚡ Conflict Resolution

When a repository has merge conflicts, press `m` to show the list of conflicting files.

| Key | Description |
|-----|-------------|
| `enter` | Open the configured mergetool for the selected file (confirmation required) |
| `esc` | Return to repository list |

The mergetool takes over the terminal. On exit, Monogit restores and refreshes the repository status.

## Responsive Layout

At widths below 80 columns, Monogit renders one focused panel at a time. `tab` or `ctrl+w w` cycles visible panels; `ctrl+w 1`, `ctrl+w 2`, and `ctrl+w 3` jump directly. This avoids clipped or overlapping panels while preserving `? help` and version in the footer.
