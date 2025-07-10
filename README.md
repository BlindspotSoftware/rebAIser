# RebAIser - AI-Assisted Git Rebaser

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

# AI configuration
ai:
  # OpenAI API key - PREFER using OPENAI_API_KEY environment variable
  openai_api_key: ""  # Leave empty to use environment variable
  # OpenRouter API key - PREFER using OPENROUTER_API_KEY environment variable
  # Provider is auto-detected: if openrouter_api_key is set, OpenRouter is used
  openrouter_api_key: ""  # Leave empty to use environment variable
  # Base URL for OpenRouter or custom endpoints (auto-configured for OpenRouter)
  base_url: ""
  # Model to use for conflict resolution
  # OpenAI models: gpt-4, gpt-4-turbo, gpt-3.5-turbo
  # OpenRouter models: anthropic/claude-3.5-sonnet, meta-llama/llama-3.1-8b-instruct, etc.
  model: "gpt-4"  # Auto-configured based on provider
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
| `OPENAI_API_KEY` | OpenAI API key (when using OpenAI provider) | `sk-abc123def456...` |
| `OPENROUTER_API_KEY` | OpenRouter API key (when using OpenRouter provider) | `sk-or-v1-abc123def456...` |

### Optional Environment Variables

| Variable | Description | Default | Example |
|----------|-------------|---------|---------|
| `AI_BASE_URL` | Custom base URL for AI API | _(auto-configured)_ | `https://openrouter.ai/api/v1` |
| `GITHUB_TOKEN` | GitHub personal access token | _(from config)_ | `ghp_abc123def456...` |
| `SLACK_WEBHOOK_URL` | Slack webhook URL for notifications | _(none)_ | `https://hooks.slack.com/services/...` |

**Note**: The AI provider is automatically detected based on which API key is provided. If `OPENROUTER_API_KEY` is set, OpenRouter is used. If only `OPENAI_API_KEY` is set, OpenAI is used. If both are set, OpenRouter is used by default with a warning message.

### Setting Environment Variables

#### Linux/macOS
```bash
# Set environment variables for OpenAI
export OPENAI_API_KEY="sk-your-api-key-here"
export GITHUB_TOKEN="ghp_your-github-token-here"
export SLACK_WEBHOOK_URL="https://hooks.slack.com/services/your/webhook/url"

# Or set environment variables for OpenRouter
export OPENROUTER_API_KEY="sk-or-v1-your-api-key-here"
export GITHUB_TOKEN="ghp_your-github-token-here"
export SLACK_WEBHOOK_URL="https://hooks.slack.com/services/your/webhook/url"

# Run the application
./ai-rebaser --config config.yaml
```

#### Windows (PowerShell)
```powershell
# Set environment variables for OpenAI
$env:OPENAI_API_KEY="sk-your-api-key-here"
$env:GITHUB_TOKEN="ghp_your-github-token-here"
$env:SLACK_WEBHOOK_URL="https://hooks.slack.com/services/your/webhook/url"

# Or set environment variables for OpenRouter
$env:OPENROUTER_API_KEY="sk-or-v1-your-api-key-here"
$env:GITHUB_TOKEN="ghp_your-github-token-here"
$env:SLACK_WEBHOOK_URL="https://hooks.slack.com/services/your/webhook/url"

# Run the application
.\ai-rebaser.exe --config config.yaml
```

#### Docker
```bash
# Run with OpenAI
docker run -e OPENAI_API_KEY="sk-your-key" \
           -e GITHUB_TOKEN="ghp_your-token" \
           -e SLACK_WEBHOOK_URL="https://hooks.slack.com/..." \
           -v $(pwd)/config.yaml:/app/config.yaml \
           ai-rebaser --config /app/config.yaml

# Run with OpenRouter
docker run -e OPENROUTER_API_KEY="sk-or-v1-your-key" \
           -e GITHUB_TOKEN="ghp_your-token" \
           -e SLACK_WEBHOOK_URL="https://hooks.slack.com/..." \
           -v $(pwd)/config.yaml:/app/config.yaml \
           ai-rebaser --config /app/config.yaml
```

#### GitHub Actions

You can use AI Rebaser as a GitHub Action in your workflows. The action is available in the GitHub Actions Marketplace.

##### Using the Action

```yaml
# .github/workflows/ai-rebaser.yml
name: AI Rebaser
on:
  schedule:
    - cron: '0 */8 * * *'  # Every 8 hours
  workflow_dispatch:  # Allow manual triggering

jobs:
  rebase:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout
        uses: actions/checkout@v4
      
      - name: Run AI Rebaser (OpenAI)
        uses: BlindspotSoftware/rebAIser@v1
        with:
          config_file: 'config.yaml'
          log_level: 'info'
          dry_run: 'false'
          run_once: 'true'
          openai_api_key: ${{ secrets.OPENAI_API_KEY }}
          github_token: ${{ secrets.GITHUB_TOKEN }}
          slack_webhook_url: ${{ secrets.SLACK_WEBHOOK_URL }}
      
      # Alternative: Using OpenRouter
      - name: Run AI Rebaser (OpenRouter)
        uses: BlindspotSoftware/rebAIser@v1
        with:
          config_file: 'config.yaml'
          log_level: 'info'
          dry_run: 'false'
          run_once: 'true'
          openrouter_api_key: ${{ secrets.OPENROUTER_API_KEY }}
          github_token: ${{ secrets.GITHUB_TOKEN }}
          slack_webhook_url: ${{ secrets.SLACK_WEBHOOK_URL }}
```

##### Action Inputs

| Input | Description | Required | Default |
|-------|-------------|----------|---------|
| `config_file` | Path to configuration file | No | `config.yaml` |
| `log_level` | Log level (debug, info, warn, error) | No | `info` |
| `dry_run` | Run in dry-run mode | No | `false` |
| `run_once` | Run once and exit | No | `true` |
| `keep_artifacts` | Keep temporary artifacts | No | `false` |
| `openai_api_key` | OpenAI API key for conflict resolution | No | - |
| `openrouter_api_key` | OpenRouter API key (auto-detects provider) | No | - |
| `ai_base_url` | Custom base URL for AI API | No | - |
| `github_token` | GitHub token | Yes | - |
| `slack_webhook_url` | Slack webhook URL | No | - |

##### Action Outputs

| Output | Description |
|--------|-------------|
| `pull_request_number` | Number of created pull request |
| `conflicts_resolved` | Number of conflicts resolved |
| `tests_passed` | Whether all tests passed |

##### Complete Example

```yaml
name: AI Rebaser Workflow
on:
  schedule:
    - cron: '0 8,16 * * 1-5'  # 8 AM and 4 PM on weekdays
  workflow_dispatch:
    inputs:
      dry_run:
        description: 'Run in dry-run mode'
        required: false
        default: 'false'
        type: boolean

jobs:
  rebase:
    runs-on: ubuntu-latest
    outputs:
      pr_number: ${{ steps.ai_rebaser.outputs.pull_request_number }}
      conflicts: ${{ steps.ai_rebaser.outputs.conflicts_resolved }}
      tests_passed: ${{ steps.ai_rebaser.outputs.tests_passed }}
    
    steps:
      - name: Checkout repository
        uses: actions/checkout@v4
        
      - name: Run AI Rebaser
        id: ai_rebaser
        uses: BlindspotSoftware/rebAIser@v1
        with:
          config_file: 'config.yaml'
          log_level: 'info'
          dry_run: ${{ inputs.dry_run || 'false' }}
          run_once: 'true'
          openai_api_key: ${{ secrets.OPENAI_API_KEY }}
          # openrouter_api_key: ${{ secrets.OPENROUTER_API_KEY }}  # Uncomment if using OpenRouter
          github_token: ${{ secrets.GITHUB_TOKEN }}
          slack_webhook_url: ${{ secrets.SLACK_WEBHOOK_URL }}
      
      - name: Report results
        if: always()
        run: |
          echo "PR Number: ${{ steps.ai_rebaser.outputs.pull_request_number }}"
          echo "Conflicts Resolved: ${{ steps.ai_rebaser.outputs.conflicts_resolved }}"
          echo "Tests Passed: ${{ steps.ai_rebaser.outputs.tests_passed }}"
```

##### Self-Hosted Alternative

If you prefer to build from source in your GitHub Actions:

```yaml
name: AI Rebaser (Self-Hosted)
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
      
      - name: Build and Run AI Rebaser
        env:
          # For OpenAI
          OPENAI_API_KEY: ${{ secrets.OPENAI_API_KEY }}
          # For OpenRouter (uncomment if using)
          # OPENROUTER_API_KEY: ${{ secrets.OPENROUTER_API_KEY }}
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
          SLACK_WEBHOOK_URL: ${{ secrets.SLACK_WEBHOOK_URL }}
        run: |
          go build -o ai-rebaser ./cmd/rebAIser
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

**OpenRouter API Key**:
1. Visit [OpenRouter API Keys](https://openrouter.ai/keys)
2. Create a new API key
3. Copy the key (starts with `sk-or-v1-`)
4. Set spending limits to control costs

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
# Missing required variables (OpenAI)
$ ./ai-rebaser
ERROR: OPENAI_API_KEY environment variable is required

# Missing required variables (OpenRouter)
$ OPENROUTER_API_KEY="" ./ai-rebaser
ERROR: no AI API key provided. Set either OPENAI_API_KEY or OPENROUTER_API_KEY environment variable

# Invalid API key format
$ OPENAI_API_KEY="invalid" ./ai-rebaser
ERROR: Invalid OpenAI API key format

# Valid configuration (OpenAI)
$ OPENAI_API_KEY="sk-..." ./ai-rebaser
INFO: Environment variables validated successfully

# Valid configuration (OpenRouter)
$ OPENROUTER_API_KEY="sk-or-v1-..." ./ai-rebaser
INFO: Environment variables validated successfully

# Both keys set (OpenRouter takes precedence)
$ OPENAI_API_KEY="sk-..." OPENROUTER_API_KEY="sk-or-v1-..." ./ai-rebaser
WARN: Both OpenAI and OpenRouter API keys are set. Using OpenRouter by default.
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
