# Keybindings Reference

Monogit is designed to be fully keyboard-driven. Below is a comprehensive list of all available shortcuts.

## 🌍 Global Navigation

| Key | Description |
|-----|-------------|
| `tab` | Cycle focus between visible panels |
| `1` | Jump to Repositories Panel |
| `2` | Jump to Branch/Commit List Panel |
| `3` | Jump to Diff/Output Panel |
| `h | ←` | Focus Left Panel |
| `l | →` | Focus Right Panel |
| `v` | Start/stop selection range |
| `y` | Copy selection to clipboard |
| `< | >` | Resize left panel (smaller / larger) |
| `ctrl+p | ?` | Toggle interactive Help Menu |
| `esc` | Back / Cancel / Close Modal |
| `q` | Quit Monogit |

The footer always keeps `? help` visible and shows the current `MonoGit` version in the bottom-right corner.

The repository list also shows compact health badges:
- `DET`: detached `HEAD`
- `UP`: no upstream tracking branch
- `CF`: merge conflicts present
- `ST`: stale branch
- `TG`: local tag on `HEAD` still not pushed to `origin`

---

## 📂 Repository Panel Actions

When the focus is on the left list:

| Key | Description |
|-----|-------------|
| `f` | **Fetch** only the selected repository |
| `F` | **Fetch All** repositories concurrently |
| `p` | **Pull** selected repository (confirmation required) |
| `P` | **Pull All** repositories (confirmation required) |
| `u` | **Push** selected repository (confirmation required) |
| `U` | **Push All** repositories (confirmation required) |
| `c` | Start the **Commit Wizard** |
| `t` | Start the **Tag/Deploy Wizard** |
| `/ | ctrl+f` | Open the repository search field in the left panel |
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
| `ctrl+y` | **Cherry-pick** a commit by hash |
| `ctrl+r` | **Revert** a commit by hash |
| `,` | Open the **Configuration Panel** |
| `g` | Toggle between **Graph** and **Simple** log views |
| `o` | Open the temporary **Command Log** to see raw output |
| `E` | **Export** command log (only visible inside the Command Log panel) |

---

## 🌿 Branch Manager Shortcuts

Inside the branch list (`b`):

| Key | Description |
|-----|-------------|
| `enter` | **Checkout** selected branch |
| `M` | **Merge** selected branch into current HEAD |
| `n` | Create a **New** branch |
| `d` | **Delete** selected branch and choose local/remote in the confirmation modal |
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

## Confirmation Modal

When a mutating action is triggered, Monogit shows a centered confirmation modal. `y` or `enter` accepts the action, `n` or `esc` cancels it, and branch deletion also supports `l` for local and `r` for remote. Fetch is direct and does not prompt. Commit wizard file selection stays local until the final commit confirmation.

## ⚡ Conflict Resolution

When a repository has merge conflicts, press `m` to show the list of conflicting files.

| Key | Description |
|-----|-------------|
| `enter` | Open the configured mergetool for the selected file (confirmation required) |
| `esc` | Return to repository list |

The mergetool takes over the terminal. On exit, Monogit restores and refreshes the repository status.
