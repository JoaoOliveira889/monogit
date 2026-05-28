# Features In-Depth

Monogit goes beyond just viewing statuses. Here is a deep dive into its core features.

## 🧙 Commit Wizard (`c`)

The Commit Wizard is a guided flow that ensures you commit precisely what you want.
1.  **Staging**: Select files using `space`. You can see a live diff of each file before staging.
2.  **Message**: Enter your commit message in an interactive prompt.
3.  **Confirm**: Monogit shows a confirmation modal before staging, committing, and optional push actions run. Fetch stays direct.
4.  **Sync**: After committing, Monogit will ask if you want to **Push**, completing your workflow in seconds.

## 🏷️ Deploy Tags (`t`)

Perfect for release management.
1.  Enter the **Version** (e.g., `v1.2.0`).
2.  Enter an **Annotation Message**.
3.  Monogit creates the annotated tag and automatically pushes it to the remote origin.

## 🔎 Repository Search & Local Tags

Use `/` to open the repository search field in the left panel. The repository list filters live while you type, `enter` keeps the current search active, and `esc` restores the previously applied filter. The search box is hidden when inactive. Use `ctrl+t` to edit local tags in the right panel for the selected repository, `d` to remove a tag with confirmation, and keep local repo tags capped at 4 per repository to keep the panel readable.

## 🛠️ Editor & Browser Integration (`e` | `w`)

Monogit intelligently detects your environment.
- **Editor**: It checks for `$MONOGIT_EDITOR`, then `$EDITOR`, and then looks for popular GUI editors like VS Code, Cursor, Zed, or IntelliJ.
- **Browser**: It parses your remote configuration and opens the repository page on GitHub, GitLab, Bitbucket, or Azure DevOps.

These integrations are explicit user actions and do not collect telemetry or background analytics.

## 📜 Command Log (`o`)

Transparency is key. Press `o` to see a temporary in-memory log of the most recent Git commands executed by Monogit. This is extremely useful for debugging failed pulls, pushes, or complex merge conflicts, and the buffer is cleared when the panel opens or closes.

## 🖼️ Startup Splash & Footer

On startup, Monogit renders its bundled splash artwork before dropping into the main workspace, giving the terminal enough time to show the branding instead of flashing past it. Once the app is running, the footer keeps `? help` visible and shows the current `MonoGit` version on the right.

## 📊 Visual Commit Graph (`g`)

Toggle the right panel to show a visual tree of your git history. This helps in understanding branches, merges, and where your local branch sits relative to the remote.

## 📦 Interactive Stash Panel (`S`)

No more blind popping. Press `S` to open a dedicated Stash panel on the right side:
1.  **List**: Instantly view all stashes in the current repository, displaying their indexes (`stash@{n}`) and descriptions.
2.  **Apply (`a`)**: Restore the stashes of your choice and keep them in the list.
3.  **Pop (`p` | `enter`)**: Apply a stash and delete it immediately.
4.  **Drop (`d`)**: Delete the selected stash permanently.
*Mutating actions that alter repository state prompt for user confirmation to prevent accidental data loss. Fetch stays direct.*

## 🔄 Advanced Batch Operations (`B` | `Z`)

Beyond the standard `f`/`p`/`u` operations, Monogit supports bulk actions that respect the current tag and search filters:

### Bulk Checkout (`B`)
1. Press `B` to open the branch-name input prompt.
2. Type the target branch (e.g., `main`) and press `enter`.
3. Confirm the operation — Monogit will `git checkout <branch>` in every visible repository concurrently (concurrency limit of 5).

### Bulk Stash (`Z`)
1. Press `Z` to stash all local changes at once.
2. Monogit shows how many dirty repositories will be affected.
3. After confirmation, `git stash` runs on every dirty repo within the current filter. Clean repos are skipped automatically.

Both operations log individual results to the Command Log (`o`) and refresh repository statuses when complete.

---

## 🚀 Smart Push with Auto-Upstream

Pushing local branches has never been easier:
- When you run a Push on a new local branch that doesn't have an upstream tracking branch configured, Monogit automatically detects this.
- It queries the configured git remotes for the repository and seamlessly pushes the branch using `--set-upstream <remote> <branch>`.
- No more command failures or manually configuring tracking branches from the terminal!

---

## 🔒 Security & Safety

- **Atomic Commands**: Monogit never uses shell strings. All commands are executed as discrete arguments, making it immune to shell injection.
- **No Telemetry**: Monogit does not ship analytics, tracking, or hidden collection of user data.
- **Confirmed Mutations**: Any action that mutates repository state requires a confirmation modal before execution. Fetch is the explicit exception.
- **Restricted Local State**: User config is stored with restrictive permissions on disk.
