# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

You are building a production-grade Golang application that performs AI-assisted rebasing of an internal Git repository tree on top of an upstream open-source repository.

### Goals:
- Rebase the internal tree up to 3×/day.
- Resolve merge conflicts via OpenAI API.
- Run user-injectable tests (e.g., build scripts, FirmwareCI).
- Push changes to a new GitHub branch.
- Open a PR with a changelog/summary.
- Notify via Slack.
- Auto-merge after 24 workday hours of inactivity.

### Requirements:
- Written in Golang
- Use `logrus` for structured logging
- Pluggable test framework (scripts or Go functions)
- Use GitHub API for PR creation and merging
- Use OpenAI API for conflict resolution (gpt-4/gpt-4o)
- Use a clean, idiomatic package layout:
  - `cmd/ai-rebaser/main.go` (entry point)
  - `internal/git/` for Git commands and conflict handling
  - `internal/ai/` for OpenAI interactions
  - `internal/test/` for test execution interface
  - `internal/github/` for GitHub PR operations
  - `internal/notify/` for Slack integration
  - `internal/config/` for YAML-based config and cron settings
- Configurable via a single YAML file (intervals, repos, test commands, etc.)
- Unit-testable and modular

## Development Commands

Since this is a new Go project, standard Go commands will be used:

- `go build` - Build the project
- `go run .` - Run the main package
- `go test ./...` - Run all tests
- `go mod tidy` - Clean up dependencies
- `go fmt ./...` - Format code
- `go vet ./...` - Run static analysis
- `go mod download` - Download dependencies

## Architecture

The project structure is minimal at this stage with only:
- `go.mod` - Go module definition

As the project develops, typical Go project structure would likely include:
- `main.go` or `cmd/` directory for executables
- Internal packages organized by functionality
- `pkg/` for exportable packages if this becomes a library

## Commit Message Style

Use Conventional Commits format for all commit messages:
- `feat:` for new features
- `fix:` for bug fixes
- `docs:` for documentation changes
- `style:` for code style changes
- `refactor:` for code refactoring
- `test:` for adding tests
- `chore:` for maintenance tasks

Examples:
- `feat: add AI conflict resolution service`
- `fix: resolve git rebase merge conflicts`
- `docs: update installation instructions`

## Notes

- Standard Go project conventions should be followed
