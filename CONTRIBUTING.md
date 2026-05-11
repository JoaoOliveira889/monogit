# Contributing to MonoGit

First off, thank you for considering contributing to MonoGit! It's people like you that make MonoGit such a great tool for the community.

## How Can I Contribute?

### Reporting Bugs
* Check the existing issues to see if the bug has already been reported.
* If not, open a new issue. Include a clear title, a description, and steps to reproduce the bug.
* Mention your OS and terminal emulator.

### Suggesting Enhancements
* Open a new issue with the tag "enhancement".
* Describe the feature you'd like to see and why it would be useful.

### Pull Requests
1. Fork the repository.
2. Create a new branch for your feature or bug fix: `git checkout -b feature/your-feature-name` or `git checkout -b fix/your-bug-fix`.
3. Make your changes.
4. Ensure the code builds and passes all tests (see below).
5. Commit your changes with a clear and descriptive commit message.
6. Push to your fork and submit a Pull Request.

## Development Setup

### Prerequisites
* [Go](https://golang.org/dl/) (version 1.21 or later)
* A terminal emulator with ANSI color support.

### Building the Project
Clone the repository and run:
```bash
go build ./cmd/monogit
```

### Running Locally
```bash
go run ./cmd/monogit
```

### Running Tests
We take testing seriously. Please ensure all tests pass before submitting a PR:
```bash
go test ./...
```

## Architecture Overview
MonoGit follows a clean architecture pattern:
- `cmd/`: Entry point of the application.
- `internal/domain`: Core entities and logic.
- `internal/usecase`: Business logic and orchestration.
- `internal/adapters`: External interfaces (TUI using Bubble Tea, Git operations).
- `internal/pkg`: Shared utilities and UI styles.

## Coding Standards
* Follow standard Go formatting (`go fmt`).
* Use meaningful variable and function names.
* Keep functions small and focused.
* Add tests for new features or bug fixes.

## License
By contributing to MonoGit, you agree that your contributions will be licensed under the project's [MIT License](LICENSE).
