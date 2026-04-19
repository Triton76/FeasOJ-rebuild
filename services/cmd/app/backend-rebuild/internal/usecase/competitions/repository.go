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
	StartAt      *time.Time
	EndAt        *time.Time
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
	List(ctx context.Context, offset, limit int, visibility, ruleType, status string) ([]Contest, error)
	GetByID(ctx context.Context, contestID int64) (Contest, error)
	CreateParticipant(ctx context.Context, p Participant) (Participant, error)
}
