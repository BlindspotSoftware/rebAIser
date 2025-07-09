package ai

import (
	"context"
	"fmt"
	"strings"

	"github.com/sashabaranov/go-openai"
	"github.com/sirupsen/logrus"

	"github.com/BlindspotSoftware/rebAIser/internal/interfaces"
)

type Service struct {
	client    *openai.Client
	model     string
	maxTokens int
	log       *logrus.Entry
}

func NewService(apiKey, model string, maxTokens int) interfaces.AIService {
	return &Service{
		client:    openai.NewClient(apiKey),
		model:     model,
		maxTokens: maxTokens,
		log:       logrus.WithField("component", "ai"),
	}
}

func (s *Service) ResolveConflict(ctx context.Context, conflict interfaces.GitConflict) (string, error) {
	s.log.WithField("file", conflict.File).Info("Resolving conflict with AI")

	// Create a detailed prompt for conflict resolution
	prompt := s.buildConflictResolutionPrompt(conflict)

	resp, err := s.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model:     s.model,
		MaxTokens: s.maxTokens,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: "You are an expert software engineer helping resolve Git merge conflicts. Your task is to intelligently merge conflicting code changes, preserving the intent of both sides where possible. Always return only the resolved code without any markdown formatting or explanations.",
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: prompt,
			},
		},
		Temperature: 0.1, // Low temperature for more deterministic output
	})

	if err != nil {
		return "", fmt.Errorf("OpenAI API call failed: %w", err)
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no response from OpenAI API")
	}

	resolution := strings.TrimSpace(resp.Choices[0].Message.Content)
	s.log.WithFields(logrus.Fields{
		"file":       conflict.File,
		"tokens_used": resp.Usage.TotalTokens,
	}).Info("AI conflict resolution completed")

	return resolution, nil
}

func (s *Service) GenerateCommitMessage(ctx context.Context, changes []string) (string, error) {
	s.log.Info("Generating commit message")

	prompt := s.buildCommitMessagePrompt(changes)

	resp, err := s.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model:     s.model,
		MaxTokens: 100, // Commit messages should be short
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: "You are an expert at writing clear, concise Git commit messages following conventional commit format. Generate a single commit message that summarizes the changes. Use format: 'type: description' where type is one of: feat, fix, docs, style, refactor, test, chore. Keep it under 50 characters for the summary.",
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: prompt,
			},
		},
		Temperature: 0.3,
	})

	if err != nil {
		return "", fmt.Errorf("OpenAI API call failed: %w", err)
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no response from OpenAI API")
	}

	commitMessage := strings.TrimSpace(resp.Choices[0].Message.Content)
	s.log.WithFields(logrus.Fields{
		"message":     commitMessage,
		"tokens_used": resp.Usage.TotalTokens,
	}).Info("AI commit message generated")

	return commitMessage, nil
}

// GenerateCommitMessageWithConflicts generates a commit message that describes the nature of conflicts resolved
func (s *Service) GenerateCommitMessageWithConflicts(ctx context.Context, changes []string, conflicts []interfaces.GitConflict) (string, error) {
	s.log.Info("Generating commit message with conflict analysis")

	prompt := s.buildCommitMessageWithConflictsPrompt(changes, conflicts)

	resp, err := s.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model:     s.model,
		MaxTokens: 150, // Slightly more tokens for conflict analysis
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: "You are an expert at writing clear, concise Git commit messages following conventional commit format. Analyze the conflicts and generate a commit message that describes the nature of the conflicts resolved (e.g., 'config: reconcile compiler toolchain defaults', 'gpio: align drive strength configurations', 'devicetree: merge panel timing settings'). Use format: 'type: description' where type is one of: feat, fix, docs, style, refactor, test, chore, config. Keep it under 50 characters for the summary.",
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: prompt,
			},
		},
		Temperature: 0.3,
	})

	if err != nil {
		return "", fmt.Errorf("OpenAI API call failed: %w", err)
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no response from OpenAI API")
	}

	commitMessage := strings.TrimSpace(resp.Choices[0].Message.Content)
	s.log.WithFields(logrus.Fields{
		"message":     commitMessage,
		"tokens_used": resp.Usage.TotalTokens,
		"conflicts":   len(conflicts),
	}).Info("AI commit message with conflicts generated")

	return commitMessage, nil
}

func (s *Service) GeneratePRDescription(ctx context.Context, commits []string, conflicts []interfaces.GitConflict) (string, error) {
	s.log.Info("Generating PR description")

	prompt := s.buildPRDescriptionPrompt(commits, conflicts)

	resp, err := s.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model:     s.model,
		MaxTokens: s.maxTokens,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: "You are an expert at writing clear, professional GitHub pull request descriptions. Generate a well-structured PR description in markdown format that summarizes the changes, conflicts resolved, and any important notes for reviewers. Include sections for Summary, Changes, Conflicts Resolved (if any), and Testing.",
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: prompt,
			},
		},
		Temperature: 0.4,
	})

	if err != nil {
		return "", fmt.Errorf("OpenAI API call failed: %w", err)
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no response from OpenAI API")
	}

	description := strings.TrimSpace(resp.Choices[0].Message.Content)
	s.log.WithFields(logrus.Fields{
		"conflicts":   len(conflicts),
		"commits":     len(commits),
		"tokens_used": resp.Usage.TotalTokens,
	}).Info("AI PR description generated")

	return description, nil
}

// buildConflictResolutionPrompt creates a detailed prompt for AI conflict resolution
func (s *Service) buildConflictResolutionPrompt(conflict interfaces.GitConflict) string {
	return fmt.Sprintf(`I have a Git merge conflict in file: %s

Here's the conflict:

%s

The conflict markers show:
- HEAD (our changes):
%s

- Incoming changes (theirs):
%s

Please resolve this conflict by:
1. Analyzing both versions
2. Merging the changes intelligently
3. Preserving the intent of both sides where possible
4. Ensuring the code remains functional
5. Following the existing code style and patterns

Return only the resolved code without any markdown formatting, explanations, or conflict markers.`,
		conflict.File,
		conflict.Content,
		conflict.Ours,
		conflict.Theirs,
	)
}

// buildCommitMessagePrompt creates a prompt for generating commit messages
func (s *Service) buildCommitMessagePrompt(changes []string) string {
	if len(changes) == 0 {
		return "Generate a commit message for an AI-assisted rebase operation with conflict resolution."
	}

	filesChanged := strings.Join(changes, ", ")
	return fmt.Sprintf(`Generate a conventional commit message for the following changes:

Files modified: %s

This was an AI-assisted rebase operation that resolved merge conflicts.

Use conventional commit format: type: description
Where type is one of: feat, fix, docs, style, refactor, test, chore

Keep the summary under 50 characters.`,
		filesChanged,
	)
}

// buildCommitMessageWithConflictsPrompt creates a prompt for generating commit messages with conflict analysis
func (s *Service) buildCommitMessageWithConflictsPrompt(changes []string, conflicts []interfaces.GitConflict) string {
	var prompt strings.Builder

	prompt.WriteString("Generate a conventional commit message for an AI-assisted rebase operation that resolved merge conflicts.\n\n")

	if len(changes) > 0 {
		prompt.WriteString("Files modified:\n")
		for _, change := range changes {
			prompt.WriteString(fmt.Sprintf("- %s\n", change))
		}
		prompt.WriteString("\n")
	}

	if len(conflicts) > 0 {
		prompt.WriteString("Conflicts resolved:\n")
		for _, conflict := range conflicts {
			prompt.WriteString(fmt.Sprintf("- %s\n", conflict.File))
			if conflict.Ours != "" && conflict.Theirs != "" {
				prompt.WriteString(fmt.Sprintf("  Conflict type: %s\n", s.analyzeConflictType(conflict)))
			}
		}
		prompt.WriteString("\n")
	}

	prompt.WriteString("Analyze the conflicts and generate a commit message that describes the nature of the conflicts resolved.\n")
	prompt.WriteString("Examples:\n")
	prompt.WriteString("- 'config: reconcile compiler toolchain defaults'\n")
	prompt.WriteString("- 'gpio: align drive strength configurations'\n")
	prompt.WriteString("- 'devicetree: merge panel timing settings'\n")
	prompt.WriteString("- 'soc: update register definitions'\n")
	prompt.WriteString("- 'kconfig: resolve build configuration conflicts'\n")
	prompt.WriteString("\nUse conventional commit format: type: description")

	return prompt.String()
}

// analyzeConflictType provides a basic analysis of conflict type based on file path and content
func (s *Service) analyzeConflictType(conflict interfaces.GitConflict) string {
	file := strings.ToLower(conflict.File)
	content := strings.ToLower(conflict.Content)
	
	// Analyze based on file extension/path
	if strings.Contains(file, "kconfig") || strings.HasSuffix(file, ".kconfig") {
		return "Kconfig option definition"
	}
	if strings.Contains(file, "devicetree") || strings.HasSuffix(file, ".cb") || strings.HasSuffix(file, ".dts") {
		return "Device tree configuration"
	}
	if strings.Contains(file, "gpio") && strings.Contains(content, "gpio_") {
		return "GPIO pin configuration"
	}
	if strings.Contains(content, "register") || strings.Contains(content, "#define") {
		return "Register definition"
	}
	if strings.Contains(content, "config") || strings.Contains(content, "cfg") {
		return "Configuration setting"
	}
	if strings.Contains(content, "delay") || strings.Contains(content, "timing") {
		return "Timing parameter"
	}
	
	return "Code change"
}

// buildPRDescriptionPrompt creates a prompt for generating PR descriptions
func (s *Service) buildPRDescriptionPrompt(commits []string, conflicts []interfaces.GitConflict) string {
	var prompt strings.Builder

	prompt.WriteString("Generate a GitHub pull request description for an AI-assisted rebase operation.\n\n")

	if len(commits) > 0 {
		prompt.WriteString("Recent commits:\n")
		for _, commit := range commits {
			prompt.WriteString(fmt.Sprintf("- %s\n", commit))
		}
		prompt.WriteString("\n")
	}

	if len(conflicts) > 0 {
		prompt.WriteString(fmt.Sprintf("Conflicts resolved in %d files:\n", len(conflicts)))
		for _, conflict := range conflicts {
			prompt.WriteString(fmt.Sprintf("- %s\n", conflict.File))
		}
		prompt.WriteString("\n")
	}

	prompt.WriteString("This is an automated rebase operation that:\n")
	prompt.WriteString("- Rebased the internal repository against the latest upstream changes\n")
	if len(conflicts) > 0 {
		prompt.WriteString("- Used AI to resolve merge conflicts intelligently\n")
	} else {
		prompt.WriteString("- Completed successfully with no merge conflicts\n")
	}
	prompt.WriteString("- Ran all configured tests to ensure functionality\n")
	if len(conflicts) > 0 {
		prompt.WriteString("\nGenerate a professional, well-structured PR description in markdown format with sections for Summary, Changes, Conflicts Resolved, and Testing.")
	} else {
		prompt.WriteString("\nGenerate a professional, well-structured PR description in markdown format with sections for Summary, Changes, and Testing. Do not include a conflicts section since no conflicts occurred.")
	}

	return prompt.String()
}