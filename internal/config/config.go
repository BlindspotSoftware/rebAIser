package config

import (
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Interval time.Duration `yaml:"interval"`
	DryRun   bool          `yaml:"dry_run"`
	
	Git    GitConfig    `yaml:"git"`
	AI     AIConfig     `yaml:"ai"`
	GitHub GitHubConfig `yaml:"github"`
	Slack  SlackConfig  `yaml:"slack"`
	Tests  TestsConfig  `yaml:"tests"`
}

type GitConfig struct {
	InternalRepo string `yaml:"internal_repo"`
	UpstreamRepo string `yaml:"upstream_repo"`
	WorkingDir   string `yaml:"working_dir"`
	Branch       string `yaml:"branch"`
}

type AIConfig struct {
	OpenAIAPIKey string `yaml:"openai_api_key"`
	Model        string `yaml:"model"`
	MaxTokens    int    `yaml:"max_tokens"`
}

type GitHubConfig struct {
	Token            string        `yaml:"token"`
	Owner            string        `yaml:"owner"`
	Repo             string        `yaml:"repo"`
	AutoMergeDelay   time.Duration `yaml:"auto_merge_delay"`
	PRTemplate       string        `yaml:"pr_template"`
	ReviewersTeam    string        `yaml:"reviewers_team"`
}

type SlackConfig struct {
	WebhookURL string `yaml:"webhook_url"`
	Channel    string `yaml:"channel"`
	Username   string `yaml:"username"`
}

type TestsConfig struct {
	Commands []TestCommand `yaml:"commands"`
	Timeout  time.Duration `yaml:"timeout"`
}

type TestCommand struct {
	Name        string            `yaml:"name"`
	Command     string            `yaml:"command"`
	Args        []string          `yaml:"args"`
	WorkingDir  string            `yaml:"working_dir"`
	Environment map[string]string `yaml:"environment"`
}

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	// Set defaults
	if config.Interval == 0 {
		config.Interval = 8 * time.Hour // Default to 3 times per day
	}
	if config.AI.Model == "" {
		config.AI.Model = "gpt-4"
	}
	if config.AI.MaxTokens == 0 {
		config.AI.MaxTokens = 2000
	}
	if config.GitHub.AutoMergeDelay == 0 {
		config.GitHub.AutoMergeDelay = 24 * time.Hour
	}
	if config.Tests.Timeout == 0 {
		config.Tests.Timeout = 30 * time.Minute
	}

	return &config, nil
}