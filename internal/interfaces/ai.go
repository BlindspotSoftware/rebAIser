package interfaces

import "context"

type AIService interface {
	ResolveConflict(ctx context.Context, conflict GitConflict) (string, error)
	GenerateCommitMessage(ctx context.Context, changes []string) (string, error)
	GeneratePRDescription(ctx context.Context, commits []string, conflicts []GitConflict) (string, error)
}