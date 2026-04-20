package submitrecords

import (
	"context"
	"time"
)

type Submission struct {
	ID          int64
	UserID      string
	ProblemID   int64
	ContestID   int64
	Language    string
	SourceCode  string
	Result      string
	Score       int
	SubmittedAt time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type Repository interface {
	Create(ctx context.Context, s Submission) (Submission, error)
	List(ctx context.Context, req Query) ([]Submission, error)
	GetByID(ctx context.Context, submissionID int64) (Submission, error)
	UpdateJudgeResult(ctx context.Context, submissionID int64, result string, score *int) (Submission, error)
}

type Query struct {
	UserID    string
	ProblemID int64
	ContestID int64
	Offset    int
	Limit     int
}
