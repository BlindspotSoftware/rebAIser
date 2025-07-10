# AI Rebaser

A production-grade Golang application that performs AI-assisted rebasing of an internal Git repository tree on top of an upstream open-source repository.

## Overview

AI Rebaser automates the process of keeping your internal repository up-to-date with upstream changes by:

- 🔄 **Automated Rebasing**: Rebases your internal repo up to 3× per day
- 🤖 **AI Conflict Resolution**: Uses OpenAI API to intelligently resolve merge conflicts
- 🧪 **Test Validation**: Runs configurable tests to ensure changes don't break functionality
- 📋 **PR Creation**: Automatically creates GitHub pull requests with AI-generated descriptions
- 💬 **Slack Notifications**: Sends status updates to your team channels
- ⏰ **Auto-merge**: Merges PRs after 24 workday hours of inactivity

## Architecture

The application follows a clean architecture pattern with clear separation of concerns:

```
rebaiser/
├── cmd/ai-rebaser/           # CLI application entry point
│   ├── main.go              # Main orchestration logic
│   ├── main_test.go         # Unit tests for core workflow
│   ├── integration_test.go  # Integration tests
│   └── testdata/            # Test configuration files
├── internal/
│   ├── interfaces/          # Service interface definitions
│   │   ├── ai.go           # AI service interface
│   │   ├── git.go          # Git service interface
│   │   ├── github.go       # GitHub service interface
│   │   ├── notify.go       # Notification service interface
│   │   └── test.go         # Test service interface
│   ├── config/             # Configuration management
│   │   ├── config.go       # YAML configuration loading
│   │   └── config_test.go  # Configuration tests
│   ├── git/                # Git operations
│   │   └── service.go      # Git service implementation
│   ├── ai/                 # OpenAI integration
│   │   └── service.go      # AI service implementation
│   ├── github/             # GitHub API integration
│   │   └── service.go      # GitHub service implementation
│   ├── notify/             # Slack notifications
│   │   └── service.go      # Notification service implementation
│   ├── test/               # Test execution
│   │   └── service.go      # Test service implementation
│   └── mocks/              # Mock implementations for testing
│       ├── ai_service.go
│       ├── git_service.go
│       ├── github_service.go
│       ├── notify_service.go
│       └── test_service.go
├── config.yaml             # Sample configuration
├── go.mod                  # Go module definition
├── go.sum                  # Go module checksums
├── CLAUDE.md              # Development guidance
└── README.md              # This file
```

## Rebase Workflow

The AI Rebaser follows a six-phase workflow:

1. **🔧 Setup Phase**: Initialize services and prepare working directory
2. **🔄 Git Operations**: Clone repositories, fetch updates, and attempt rebase
3. **🤖 Conflict Resolution**: Use AI to resolve any merge conflicts
4. **🧪 Testing Phase**: Run configured tests to validate changes
5. **📋 PR Creation**: Create GitHub pull request with AI-generated content
6. **📢 Notifications**: Send Slack notifications about the operation status

## Installation

### Prerequisites

- Go 1.23.1 or later
- Git installed and configured
- OpenAI API key
- GitHub personal access token
- Slack webhook URL (optional)

### Build from Source

```bash
# Clone the repository
git clone https://github.com/BlindspotSoftware/rebAIser.git
cd rebaiser

# Build the application
go build -o ai-rebaser ./cmd/ai-rebaser

# Or install directly
go install ./cmd/ai-rebaser
```

## Configuration

Create a `config.yaml` file with your settings:

```yaml
# How often to run the rebase process (8h = 3 times per day)
interval: 8h

# Dry run mode - don't make actual changes
dry_run: false

# Git configuration
git:
  # Path to your internal repository
  internal_repo: "https://github.com/your-org/internal-repo.git"
  # Upstream repository to rebase against
  upstream_repo: "https://github.com/upstream/open-source-repo.git"
  # Working directory for git operations
  working_dir: "/tmp/ai-rebaser-work"
  # Branch to rebase onto
  branch: "main"

# OpenAI configuration
ai:
  # OpenAI API key - PREFER using OPENAI_API_KEY environment variable
  openai_api_key: ""  # Leave empty to use environment variable
  # Model to use for conflict resolution
  model: "gpt-4"
  # Maximum tokens for AI responses
  max_tokens: 2000

# GitHub configuration
github:
  # GitHub personal access token - PREFER using GITHUB_TOKEN environment variable
  token: ""  # Leave empty to use environment variable
  # Repository owner
  owner: "your-org"
  # Repository name
  repo: "internal-repo"
  # How long to wait before auto-merging PRs (24h = 1 workday)
  auto_merge_delay: 24h
  # Team to request reviews from
  reviewers_team: "core-team"

# Slack notification configuration
slack:
  # Slack webhook URL - PREFER using SLACK_WEBHOOK_URL environment variable
  webhook_url: ""  # Leave empty to use environment variable
  # Channel to send notifications to
  channel: "#ai-rebaser"
  # Username for the bot
  username: "AI Rebaser"

# Test configuration
tests:
  # Maximum time to wait for all tests to complete
  timeout: 30m
  # List of test commands to run
  commands:
    - name: "build"
      command: "go"
      args: ["build", "./..."]
      working_dir: ""
      environment:
        CGO_ENABLED: "0"
    
    - name: "test"
      command: "go"
      args: ["test", "./..."]
      working_dir: ""
      environment:
        CGO_ENABLED: "0"
    
    - name: "lint"
      command: "golangci-lint"
      args: ["run"]
      working_dir: ""
      environment: {}
```

## Environment Variables

AI Rebaser supports configuration through environment variables, which take precedence over config file values. This is especially useful for CI/CD environments and containerized deployments.

### Required Environment Variables

| Variable | Description | Example |
|----------|-------------|---------|
| `OPENAI_API_KEY` | OpenAI API key for AI conflict resolution | `sk-abc123def456...` |

### Optional Environment Variables

| Variable | Description | Default | Example |
|----------|-------------|---------|---------|
| `GITHUB_TOKEN` | GitHub personal access token | _(from config)_ | `ghp_abc123def456...` |
| `SLACK_WEBHOOK_URL` | Slack webhook URL for notifications | _(none)_ | `https://hooks.slack.com/services/...` |

### Setting Environment Variables

#### Linux/macOS
```bash
# Set environment variables
export OPENAI_API_KEY="sk-your-api-key-here"
export GITHUB_TOKEN="ghp_your-github-token-here"
export SLACK_WEBHOOK_URL="https://hooks.slack.com/services/your/webhook/url"

# Run the application
./ai-rebaser --config config.yaml
```

#### Windows (PowerShell)
```powershell
# Set environment variables
$env:OPENAI_API_KEY="sk-your-api-key-here"
$env:GITHUB_TOKEN="ghp_your-github-token-here"
$env:SLACK_WEBHOOK_URL="https://hooks.slack.com/services/your/webhook/url"

# Run the application
.\ai-rebaser.exe --config config.yaml
```

#### Docker
```bash
# Run with environment variables
docker run -e OPENAI_API_KEY="sk-your-key" \
           -e GITHUB_TOKEN="ghp_your-token" \
           -e SLACK_WEBHOOK_URL="https://hooks.slack.com/..." \
           -v $(pwd)/config.yaml:/app/config.yaml \
           ai-rebaser --config /app/config.yaml
```

#### GitHub Actions
```yaml
# .github/workflows/ai-rebaser.yml
name: AI Rebaser
on:
  schedule:
    - cron: '0 */8 * * *'  # Every 8 hours

jobs:
  rebase:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v4
        with:
          go-version: '1.23'
      
      - name: Run AI Rebaser
        env:
          OPENAI_API_KEY: ${{ secrets.OPENAI_API_KEY }}
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
          SLACK_WEBHOOK_URL: ${{ secrets.SLACK_WEBHOOK_URL }}
        run: |
          go build -o ai-rebaser ./cmd/ai-rebaser
          ./ai-rebaser --config config.yaml --run-once
```

### API Key Security

⚠️ **Security Best Practices**:

1. **Never commit API keys** to your repository
2. **Use environment variables** or secure secret management
3. **Rotate keys regularly** as part of security hygiene
4. **Limit API key permissions** to minimum required scope
5. **Monitor API usage** for unexpected activity

#### Getting Your API Keys

**OpenAI API Key**:
1. Visit [OpenAI API Keys](https://platform.openai.com/api-keys)
2. Create a new secret key
3. Copy the key (starts with `sk-`)
4. Set billing limits to control costs

**GitHub Token**:
1. Go to [GitHub Personal Access Tokens](https://github.com/settings/tokens)
2. Generate new token (classic)
3. Required scopes: `repo`, `workflow`, `write:org`
4. Copy the token (starts with `ghp_`)

**Slack Webhook URL**:
1. Go to [Slack Apps](https://api.slack.com/apps)
2. Create new app or use existing
3. Enable Incoming Webhooks
4. Add webhook to workspace
5. Copy the webhook URL

### Environment Variable Validation

AI Rebaser validates environment variables on startup:

```bash
# Missing required variables
$ ./ai-rebaser
ERROR: OPENAI_API_KEY environment variable is required

# Invalid API key format
$ OPENAI_API_KEY="invalid" ./ai-rebaser
ERROR: Invalid OpenAI API key format

# Valid configuration
$ OPENAI_API_KEY="sk-..." ./ai-rebaser
INFO: Environment variables validated successfully
```

## Usage

### Command Line Options

```bash
# Run continuously with default config
./ai-rebaser

# Run with custom config file
./ai-rebaser --config /path/to/config.yaml

# Run once and exit (don't run periodically)
./ai-rebaser --run-once

# Enable dry run mode
./ai-rebaser --dry-run

# Set log level
./ai-rebaser --log-level debug

# Show version
./ai-rebaser --version

# Show help
./ai-rebaser --help
```

### Example Commands

```bash
# Test configuration with dry run
./ai-rebaser --config config.yaml --dry-run --run-once

# Run in debug mode
./ai-rebaser --log-level debug

# Run with custom interval (via config file)
./ai-rebaser --config production-config.yaml
```

## Contributing

For development information, testing, and contribution guidelines, please see [DEVELOP.md](DEVELOP.md).

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Support

For questions, issues, or contributions:

- 📧 **Email**: [maintainers@blindspot.software](mailto:maintainers@blindspot.software)
- 🐛 **Issues**: [GitHub Issues](https://github.com/BlindspotSoftware/rebAIser/issues)
- 💬 **Discussions**: [GitHub Discussions](https://github.com/BlindspotSoftware/rebAIser/discussions)

---

**AI Rebaser** - Keeping your internal repositories in sync with upstream, intelligently automated.
