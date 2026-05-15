# Getting Started

This guide will help you install and start using Monogit on your machine.

## 📦 Installation

### 1. Homebrew (macOS & Linux)

This is the recommended way to install and keep Monogit updated.

```bash
brew tap JoaoOliveira889/tap
brew install monogit
```

### 2. Pre-built Binaries

Download the latest version for your architecture from the [GitHub Releases](https://github.com/JoaoOliveira889/monogit/releases/latest) page.

**For macOS (Apple Silicon):**
```bash
curl -LO https://github.com/JoaoOliveira889/monogit/releases/latest/download/monogit_Darwin_arm64.tar.gz
tar -xzf monogit_Darwin_arm64.tar.gz
sudo mv monogit /usr/local/bin/
```

### 3. Using Go Install

If you have Go 1.23+ installed:

```bash
go install github.com/JoaoOliveira889/monogit/cmd/monogit@latest
```

---

## 🏁 Basic Usage

### Scanning a Directory

By default, Monogit scans the current directory:

```bash
monogit
```

To scan a specific path (e.g., your projects folder):

```bash
monogit --path ~/dev/projects
```

### Auto-Fetch Interval

Monogit automatically fetches updates in the background. You can change the frequency (default is 5 minutes):

```bash
monogit --interval 10m
```

### Navigation Basics

- Use **Tab** to switch between the Repository list and the Detail/Log panel.
- Use **j/k** or **arrows** to move the selection.
- Press **q** to quit at any time.
