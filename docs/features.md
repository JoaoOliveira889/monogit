# Features In-Depth

Monogit goes beyond just viewing statuses. Here is a deep dive into its core features.

## 🧙 Commit Wizard (`c`)

The Commit Wizard is a guided flow that ensures you commit precisely what you want.
1.  **Staging**: Select files using `space`. You can see a live diff of each file before staging.
2.  **Message**: Enter your commit message in an interactive prompt.
3.  **Sync**: After committing, Monogit will ask if you want to **Pull** and then **Push**, completing your workflow in seconds.

## 🏷️ Deploy Tags (`t`)

Perfect for release management.
1.  Enter the **Version** (e.g., `v1.2.0`).
2.  Enter an **Annotation Message**.
3.  Monogit creates the annotated tag and automatically pushes it to the remote origin.

## 🛠️ Editor & Browser Integration (`e` / `w`)

Monogit intelligently detects your environment.
- **Editor**: It checks for `$MONOGIT_EDITOR`, then `$EDITOR`, and then looks for popular GUI editors like VS Code, Cursor, Zed, or IntelliJ.
- **Browser**: It parses your remote configuration and opens the repository page on GitHub, GitLab, Bitbucket, or Azure DevOps.

## 📜 Command Log (`o`)

Transparency is key. Press `o` to see a running log of every Git command executed by Monogit. This is extremely useful for debugging failed pulls, pushes, or complex merge conflicts.

## 📊 Visual Commit Graph (`g`)

Toggle the right panel to show a visual tree of your git history. This helps in understanding branches, merges, and where your local branch sits relative to the remote.

---

## 🔒 Security & Safety

- **Atomic Commands**: Monogit never uses shell strings. All commands are executed as discrete arguments, making it immune to shell injection.
- **Read-Only by Default**: Background fetches are non-destructive. Destructive actions (like discarding changes or deleting branches) always require user confirmation or a dedicated keypress.
