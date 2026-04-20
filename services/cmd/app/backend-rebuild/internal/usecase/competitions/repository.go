package competitions

import (
	"context"
	"time"
)

type Contest struct {
	ID           int64
	Title        string
	Subtitle     string
	Description  string
	Announcement string
	OwnerUserID  string
	ClassID      string
	Visibility   string
	RuleType     string
	Status       string
	IsEncrypted  bool
	PasswordHash string
	StartAt      *time.Time
	EndAt        *time.Time
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type Participant struct {
	ID        string
	ContestID int64
	UserID    string
	Status    string
	CreatedAt time.Time
	UpdatedAt time.Time
	JoinedAt  *time.Time
}

type ScoreboardSubmission struct {
	UserID      string
	Username    string
	ProblemID   int64
	Result      string
	Score       int
	SubmittedAt time.Time
}

type ContestProblemBinding struct {
	ContestID    int64
	ProblemID    int64
	DisplayOrder int
	Alias        string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type Repository interface {
	ListVisible(ctx context.Context, offset, limit int, visibility, ruleType, status, actorUserID, actorRole string) ([]Contest, error)
	GetVisibleByID(ctx context.Context, contestID int64, actorUserID, actorRole string) (Contest, error)
	GetByID(ctx context.Context, contestID int64) (Contest, error)
	Create(ctx context.Context, c Contest) (Contest, error)
	Update(ctx context.Context, c Contest) (Contest, error)
	Delete(ctx context.Context, contestID int64) error
	ListProblemBindings(ctx context.Context, contestID int64) ([]ContestProblemBinding, error)
	ReplaceProblemBindings(ctx context.Context, contestID int64, items []ContestProblemBinding) error
	CountProblemBindings(ctx context.Context, contestID int64) (int64, error)
	CountExistingProblems(ctx context.Context, problemIDs []int64) (int64, error)
	CreateParticipant(ctx context.Context, p Participant) (Participant, error)
	ListScoreboardSubmissions(ctx context.Context, contestID int64, before time.Time) ([]ScoreboardSubmission, error)
}
