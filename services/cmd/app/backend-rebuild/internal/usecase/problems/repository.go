package problems

import "context"

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
}

type ProblemRepository interface {
	List(ctx context.Context, offset, limit int, visibility, status string) ([]Problem, error)
	GetByID(ctx context.Context, problemID int64) (Problem, error)
}
