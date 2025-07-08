package ai

import (
	"context"
	"fmt"

	"github.com/sirupsen/logrus"

	"github.com/9elements/rebaiser/internal/interfaces"
)

type Service struct {
	apiKey   string
	model    string
	maxTokens int
	log      *logrus.Entry
}

func NewService(apiKey, model string, maxTokens int) interfaces.AIService {
	return &Service{
		apiKey:    apiKey,
		model:     model,
		maxTokens: maxTokens,
		log:       logrus.WithField("component", "ai"),
	}
}

func (s *Service) ResolveConflict(ctx context.Context, conflict interfaces.GitConflict) (string, error) {
	s.log.WithField("file", conflict.File).Info("Resolving conflict with AI")

	// TODO: Implement OpenAI API call
	// For now, return a placeholder resolution
	resolution := fmt.Sprintf("// AI-resolved conflict for %s\n%s", conflict.File, conflict.Ours)
	
	return resolution, nil
}

func (s *Service) GenerateCommitMessage(ctx context.Context, changes []string) (string, error) {
	s.log.Info("Generating commit message")

	// TODO: Implement OpenAI API call
	// For now, return a placeholder message
	return "AI-assisted rebase with conflict resolution", nil
}

func (s *Service) GeneratePRDescription(ctx context.Context, commits []string, conflicts []interfaces.GitConflict) (string, error) {
	s.log.Info("Generating PR description")

	// TODO: Implement OpenAI API call
	// For now, return a placeholder description
	description := fmt.Sprintf("## AI-Assisted Rebase\n\nThis PR contains %d resolved conflicts and %d commits.", len(conflicts), len(commits))
	
	return description, nil
}