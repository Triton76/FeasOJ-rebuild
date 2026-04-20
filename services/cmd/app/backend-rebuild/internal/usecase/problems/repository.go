package problems

import (
	"context"
	"time"
)

type Problem struct {
	ID            int64
	Title         string
	Content       string
	Input         string
	Output        string
	Difficulty    int
	TimeLimitMS   int
	MemoryLimitMB int
	OwnerUserID   string
	ClassID       string
	Visibility    string
	Status        string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type ProblemRepository interface {
	ListVisible(ctx context.Context, offset, limit int, visibility, status, actorUserID, actorRole string) ([]Problem, error)
	GetVisibleByID(ctx context.Context, problemID int64, actorUserID, actorRole string) (Problem, error)
	GetByID(ctx context.Context, problemID int64) (Problem, error)
	Create(ctx context.Context, p Problem) (Problem, error)
	Update(ctx context.Context, p Problem) (Problem, error)
	Delete(ctx context.Context, problemID int64) error
}
