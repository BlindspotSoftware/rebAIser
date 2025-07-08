package interfaces

import (
	"context"
	"time"
)

type TestService interface {
	RunTests(ctx context.Context, workingDir string) (*TestResult, error)
	RunCommand(ctx context.Context, cmd TestCommand) (*CommandResult, error)
}

type TestCommand struct {
	Name        string
	Command     string
	Args        []string
	WorkingDir  string
	Environment map[string]string
	Timeout     time.Duration
}

type TestResult struct {
	Success     bool
	Duration    time.Duration
	Results     []CommandResult
	FailedTests []string
}

type CommandResult struct {
	Command   string
	Success   bool
	Output    string
	Error     string
	Duration  time.Duration
	ExitCode  int
}