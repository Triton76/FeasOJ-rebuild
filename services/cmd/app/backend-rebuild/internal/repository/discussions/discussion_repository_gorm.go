package discussionsrepo

import (
	"FeasOJ/app/backend-rebuild/internal/ports"
	discussionsusecase "FeasOJ/app/backend-rebuild/internal/usecase/discussions"
	"context"
	"errors"
	"time"

	"gorm.io/gorm"
)

type discussionRow struct {
	ID        string    `gorm:"column:id"`
	Title     string    `gorm:"column:title"`
	Content   string    `gorm:"column:content"`
	UserID    string    `gorm:"column:user_id"`
	CreatedAt time.Time `gorm:"column:created_at"`
}

func (discussionRow) TableName() string { return "discussions" }

type commentRow struct {
	ID           string    `gorm:"column:id"`
	DiscussionID string    `gorm:"column:discussion_id"`
	Content      string    `gorm:"column:content"`
	UserID       string    `gorm:"column:user_id"`
	Profanity    bool      `gorm:"column:profanity"`
	CreatedAt    time.Time `gorm:"column:created_at"`
}

func (commentRow) TableName() string { return "comments" }

type userRow struct {
	ID       string `gorm:"column:id"`
	Username string `gorm:"column:username"`
	Avatar   string `gorm:"column:avatar"`
}

func (userRow) TableName() string { return "users" }

type Repository struct{ db *gorm.DB }

func NewRepository(db *gorm.DB) *Repository { return &Repository{db: db} }

func (r *Repository) List(ctx context.Context, offset, limit int) ([]discussionsusecase.Discussion, error) {
	var rows []discussionRow
	err := r.db.WithContext(ctx).Order("created_at DESC").Offset(offset).Limit(limit).Find(&rows).Error
	if err != nil {
		return nil, err
	}
	resp := make([]discussionsusecase.Discussion, 0, len(rows))
	for _, row := range rows {
		resp = append(resp, r.toDiscussion(ctx, row))
	}
	return resp, nil
}

func (r *Repository) GetByID(ctx context.Context, discussionID string) (discussionsusecase.Discussion, error) {
	var row discussionRow
	err := r.db.WithContext(ctx).Where("id = ?", discussionID).Take(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return discussionsusecase.Discussion{}, ports.ErrNotFound
	}
	if err != nil {
		return discussionsusecase.Discussion{}, err
	}
	return r.toDiscussion(ctx, row), nil
}

func (r *Repository) CreateDiscussion(ctx context.Context, d discussionsusecase.Discussion) (discussionsusecase.Discussion, error) {
	row := discussionRow{ID: d.ID, Title: d.Title, Content: d.Content, UserID: d.UserID, CreatedAt: d.CreatedAt}
	if err := r.db.WithContext(ctx).Create(&row).Error; err != nil {
		return discussionsusecase.Discussion{}, err
	}
	return r.toDiscussion(ctx, row), nil
}

func (r *Repository) CreateComment(ctx context.Context, c discussionsusecase.Comment) (discussionsusecase.Comment, error) {
	var exists int64
	if err := r.db.WithContext(ctx).Model(&discussionRow{}).Where("id = ?", c.DiscussionID).Count(&exists).Error; err != nil {
		return discussionsusecase.Comment{}, err
	}
	if exists == 0 {
		return discussionsusecase.Comment{}, ports.ErrNotFound
	}

	row := commentRow{ID: c.ID, DiscussionID: c.DiscussionID, Content: c.Content, UserID: c.UserID, Profanity: c.Profanity, CreatedAt: c.CreatedAt}
	if err := r.db.WithContext(ctx).Create(&row).Error; err != nil {
		return discussionsusecase.Comment{}, err
	}
	return r.toComment(ctx, row), nil
}

func (r *Repository) ListCommentsByDiscussionID(ctx context.Context, discussionID string, offset, limit int) ([]discussionsusecase.Comment, error) {
	var rows []commentRow
	err := r.db.WithContext(ctx).
		Where("discussion_id = ?", discussionID).
		Order("created_at ASC").
		Offset(offset).
		Limit(limit).
		Find(&rows).Error
	if err != nil {
		return nil, err
	}

	resp := make([]discussionsusecase.Comment, 0, len(rows))
	for _, row := range rows {
		resp = append(resp, r.toComment(ctx, row))
	}
	return resp, nil
}

func (r *Repository) GetCommentByID(ctx context.Context, commentID string) (discussionsusecase.Comment, error) {
	var row commentRow
	err := r.db.WithContext(ctx).Where("id = ?", commentID).Take(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return discussionsusecase.Comment{}, ports.ErrNotFound
	}
	if err != nil {
		return discussionsusecase.Comment{}, err
	}
	return r.toComment(ctx, row), nil
}

func (r *Repository) DeleteDiscussion(ctx context.Context, discussionID string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("discussion_id = ?", discussionID).Delete(&commentRow{}).Error; err != nil {
			return err
		}
		res := tx.Where("id = ?", discussionID).Delete(&discussionRow{})
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return ports.ErrNotFound
		}
		return nil
	})
}

func (r *Repository) DeleteComment(ctx context.Context, commentID string) error {
	res := r.db.WithContext(ctx).Where("id = ?", commentID).Delete(&commentRow{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ports.ErrNotFound
	}
	return nil
}

func (r *Repository) toDiscussion(ctx context.Context, row discussionRow) discussionsusecase.Discussion {
	u := r.getUser(ctx, row.UserID)
	return discussionsusecase.Discussion{ID: row.ID, Title: row.Title, Content: row.Content, UserID: row.UserID, Username: u.Username, Avatar: u.Avatar, CreatedAt: row.CreatedAt}
}

func (r *Repository) toComment(ctx context.Context, row commentRow) discussionsusecase.Comment {
	u := r.getUser(ctx, row.UserID)
	return discussionsusecase.Comment{ID: row.ID, DiscussionID: row.DiscussionID, Content: row.Content, UserID: row.UserID, Username: u.Username, Avatar: u.Avatar, Profanity: row.Profanity, CreatedAt: row.CreatedAt}
}

func (r *Repository) getUser(ctx context.Context, userID string) userRow {
	if userID == "" {
		return userRow{}
	}
	var row userRow
	_ = r.db.WithContext(ctx).Where("id = ?", userID).Take(&row).Error
	return row
}
