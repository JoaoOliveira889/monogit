# Troubleshooting & FAQ

## ❓ Common Issues

### "Repository not showing up"
Ensure the directory contains a `.git` folder. Monogit only scans for immediate subdirectories of the provided `--path` that are valid Git repositories.

### "Editor fails to open"
If you are using a GUI editor, make sure its command-line tool is installed in your `$PATH` (e.g., `code` for VS Code). On macOS, you can usually do this via the editor's command palette (e.g., "Shell Command: Install 'code' command in PATH").

### "Ahead/Behind counters are not updating"
This usually means a background `git fetch` failed. Check the Command Log (`o`) to see if there are any network errors or authentication prompts (SSH/GPG) blocking the process.

---

## 💬 Frequently Asked Questions

**Does Monogit support SVN or Mercurial?**
No, Monogit is exclusively focused on providing the best experience for Git repositories.

**Can I run Monogit on Windows?**
Yes, Monogit supports Windows, though integration features like "Open in Editor" are currently optimized for macOS and Linux.

**How do I update Monogit?**
If installed via Homebrew: `brew upgrade monogit`.
If using pre-built binary: download the new version and replace the old binary.

---

## 🆘 Still having trouble?

If your issue isn't listed here, please feel free to:
1.  Check the [GitHub Issues](https://github.com/JoaoOliveira889/monogit/issues).
2.  Open a new issue with your logs and environment details.
