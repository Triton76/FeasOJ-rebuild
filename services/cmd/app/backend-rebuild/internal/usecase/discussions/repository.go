package discussions

import (
	"context"
	"time"
)

type Discussion struct {
	ID        string
	Title     string
	Content   string
	UserID    string
	Username  string
	Avatar    string
	CreatedAt time.Time
}

type Comment struct {
	ID           string
	DiscussionID string
	Content      string
	UserID       string
	Username     string
	Avatar       string
	Profanity    bool
	CreatedAt    time.Time
}

type Repository interface {
	List(ctx context.Context, offset, limit int) ([]Discussion, error)
	GetByID(ctx context.Context, discussionID string) (Discussion, error)
	CreateDiscussion(ctx context.Context, d Discussion) (Discussion, error)
	CreateComment(ctx context.Context, c Comment) (Comment, error)
}
