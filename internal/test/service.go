package test

import (
	"context"
	"fmt"
	"os/exec"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/BlindspotSoftware/rebAIser/internal/interfaces"
)

type Service struct {
	log      *logrus.Entry
	commands []interfaces.TestCommand
}

func NewService(commands []interfaces.TestCommand) interfaces.TestService {
	return &Service{
		log:      logrus.WithField("component", "test"),
		commands: commands,
	}
}

func (s *Service) RunTests(ctx context.Context, workingDir string) (*interfaces.TestResult, error) {
	s.log.WithField("workingDir", workingDir).Info("Running tests")

	// If no test commands configured, skip testing
	if len(s.commands) == 0 {
		s.log.Info("No test commands configured, skipping tests")
		return &interfaces.TestResult{
			Success:     true,
			Duration:    0,
			Results:     []interfaces.CommandResult{},
			FailedTests: []string{},
		}, nil
	}

	var results []interfaces.CommandResult
	var failedTests []string
	allSuccess := true
	startTime := time.Now()

	for _, testCmd := range s.commands {
		result, err := s.RunCommand(ctx, testCmd)
		if err != nil {
			s.log.WithError(err).WithField("command", testCmd.Name).Error("Failed to run test command")
			allSuccess = false
			failedTests = append(failedTests, testCmd.Name)
			continue
		}

		results = append(results, *result)
		if !result.Success {
			allSuccess = false
			failedTests = append(failedTests, testCmd.Name)
		}
	}

	return &interfaces.TestResult{
		Success:     allSuccess,
		Duration:    time.Since(startTime),
		Results:     results,
		FailedTests: failedTests,
	}, nil
}

func (s *Service) RunCommand(ctx context.Context, testCmd interfaces.TestCommand) (*interfaces.CommandResult, error) {
	s.log.WithField("command", testCmd.Name).Info("Running test command")

	startTime := time.Now()
	
	// Create context with timeout
	cmdCtx := ctx
	if testCmd.Timeout > 0 {
		var cancel context.CancelFunc
		cmdCtx, cancel = context.WithTimeout(ctx, testCmd.Timeout)
		defer cancel()
	}

	cmd := exec.CommandContext(cmdCtx, testCmd.Command, testCmd.Args...)
	cmd.Dir = testCmd.WorkingDir

	// Set environment variables
	if testCmd.Environment != nil {
		for key, value := range testCmd.Environment {
			cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", key, value))
		}
	}

	output, err := cmd.CombinedOutput()
	duration := time.Since(startTime)

	result := &interfaces.CommandResult{
		Command:  fmt.Sprintf("%s %s", testCmd.Command, testCmd.Args),
		Success:  err == nil,
		Output:   string(output),
		Duration: duration,
	}

	if err != nil {
		result.Error = err.Error()
		if exitError, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitError.ExitCode()
		}
	}

	s.log.WithFields(logrus.Fields{
		"command":  testCmd.Name,
		"success":  result.Success,
		"duration": duration,
	}).Info("Test command completed")

	return result, nil
}