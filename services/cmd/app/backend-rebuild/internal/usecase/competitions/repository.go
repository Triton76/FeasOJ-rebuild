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

type Repository interface {
	ListVisible(ctx context.Context, offset, limit int, visibility, ruleType, status, actorUserID, actorRole string) ([]Contest, error)
	GetVisibleByID(ctx context.Context, contestID int64, actorUserID, actorRole string) (Contest, error)
	GetByID(ctx context.Context, contestID int64) (Contest, error)
	Create(ctx context.Context, c Contest) (Contest, error)
	Update(ctx context.Context, c Contest) (Contest, error)
	Delete(ctx context.Context, contestID int64) error
	CreateParticipant(ctx context.Context, p Participant) (Participant, error)
}
