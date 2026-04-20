package problemsrepo

import (
	"FeasOJ/app/backend-rebuild/internal/ports"
	problemsusecase "FeasOJ/app/backend-rebuild/internal/usecase/problems"
	"context"
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"
)

type problemRow struct {
	ID            int64     `gorm:"column:id"`
	Title         string    `gorm:"column:title"`
	Content       string    `gorm:"column:content"`
	Input         string    `gorm:"column:input"`
	Output        string    `gorm:"column:output"`
	Difficulty    int       `gorm:"column:difficulty"`
	TimeLimitMS   int       `gorm:"column:time_limit_ms"`
	MemoryLimitMB int       `gorm:"column:memory_limit_mb"`
	OwnerUserID   string    `gorm:"column:owner_user_id"`
	ClassID       string    `gorm:"column:class_id"`
	Visibility    string    `gorm:"column:visibility"`
	Status        string    `gorm:"column:status"`
	CreatedAt     time.Time `gorm:"column:created_at"`
	UpdatedAt     time.Time `gorm:"column:updated_at"`
}

func (problemRow) TableName() string {
	return "problems"
}

type ProblemRepository struct {
	db *gorm.DB
}

func NewProblemRepository(db *gorm.DB) *ProblemRepository {
	return &ProblemRepository{db: db}
}

func (r *ProblemRepository) ListVisible(ctx context.Context, offset, limit int, visibility, status, actorUserID, actorRole string) ([]problemsusecase.Problem, error) {
	q := r.db.WithContext(ctx).Model(&problemRow{})
	q = applyProblemVisibilityScope(q, actorUserID, actorRole)
	if visibility != "" {
		q = q.Where("visibility = ?", visibility)
	}
	if status != "" {
		q = q.Where("status = ?", status)
	}

	var rows []problemRow
	err := q.Order("id DESC").Offset(offset).Limit(limit).Find(&rows).Error
	if err != nil {
		return nil, err
	}

	resp := make([]problemsusecase.Problem, 0, len(rows))
	for _, row := range rows {
		resp = append(resp, fromRow(row))
	}
	return resp, nil
}

func (r *ProblemRepository) GetVisibleByID(ctx context.Context, problemID int64, actorUserID, actorRole string) (problemsusecase.Problem, error) {
	q := r.db.WithContext(ctx).Model(&problemRow{}).Where("id = ?", problemID)
	q = applyProblemVisibilityScope(q, actorUserID, actorRole)

	var row problemRow
	err := q.Take(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return problemsusecase.Problem{}, ports.ErrNotFound
	}
	if err != nil {
		return problemsusecase.Problem{}, err
	}
	return fromRow(row), nil
}

func (r *ProblemRepository) GetByID(ctx context.Context, problemID int64) (problemsusecase.Problem, error) {
	var row problemRow
	err := r.db.WithContext(ctx).Where("id = ?", problemID).Take(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return problemsusecase.Problem{}, ports.ErrNotFound
	}
	if err != nil {
		return problemsusecase.Problem{}, err
	}
	return fromRow(row), nil
}

func (r *ProblemRepository) Create(ctx context.Context, p problemsusecase.Problem) (problemsusecase.Problem, error) {
	row := problemRow{
		Title:         p.Title,
		Content:       p.Content,
		Input:         p.Input,
		Output:        p.Output,
		Difficulty:    p.Difficulty,
		TimeLimitMS:   p.TimeLimitMS,
		MemoryLimitMB: p.MemoryLimitMB,
		OwnerUserID:   p.OwnerUserID,
		ClassID:       p.ClassID,
		Visibility:    p.Visibility,
		Status:        p.Status,
		CreatedAt:     p.CreatedAt,
		UpdatedAt:     p.UpdatedAt,
	}
	if err := r.db.WithContext(ctx).Create(&row).Error; err != nil {
		return problemsusecase.Problem{}, err
	}
	return fromRow(row), nil
}

func (r *ProblemRepository) Update(ctx context.Context, p problemsusecase.Problem) (problemsusecase.Problem, error) {
	updates := map[string]any{
		"title":           p.Title,
		"content":         p.Content,
		"input":           p.Input,
		"output":          p.Output,
		"difficulty":      p.Difficulty,
		"time_limit_ms":   p.TimeLimitMS,
		"memory_limit_mb": p.MemoryLimitMB,
		"class_id":        p.ClassID,
		"visibility":      p.Visibility,
		"status":          p.Status,
		"updated_at":      p.UpdatedAt,
	}
	res := r.db.WithContext(ctx).Model(&problemRow{}).Where("id = ?", p.ID).Updates(updates)
	if res.Error != nil {
		return problemsusecase.Problem{}, res.Error
	}
	if res.RowsAffected == 0 {
		return problemsusecase.Problem{}, ports.ErrNotFound
	}
	return r.GetByID(ctx, p.ID)
}

func (r *ProblemRepository) Delete(ctx context.Context, problemID int64) error {
	res := r.db.WithContext(ctx).Where("id = ?", problemID).Delete(&problemRow{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ports.ErrNotFound
	}
	return nil
}

func applyProblemVisibilityScope(q *gorm.DB, actorUserID, actorRole string) *gorm.DB {
	if strings.TrimSpace(actorRole) == "admin" {
		return q
	}
	actorUserID = strings.TrimSpace(actorUserID)
	if actorUserID == "" {
		return q.Where("visibility = ?", "public")
	}
	return q.Where(`
		visibility = 'public'
		OR owner_user_id = ?
		OR (
			visibility = 'class'
			AND class_id IN (
				SELECT class_id
				FROM class_memberships
				WHERE user_id = ? AND status = 'active'
			)
		)
	`, actorUserID, actorUserID)
}

func fromRow(r problemRow) problemsusecase.Problem {
	return problemsusecase.Problem{
		ID:            r.ID,
		Title:         r.Title,
		Content:       r.Content,
		Input:         r.Input,
		Output:        r.Output,
		Difficulty:    r.Difficulty,
		TimeLimitMS:   r.TimeLimitMS,
		MemoryLimitMB: r.MemoryLimitMB,
		OwnerUserID:   r.OwnerUserID,
		ClassID:       r.ClassID,
		Visibility:    r.Visibility,
		Status:        r.Status,
		CreatedAt:     r.CreatedAt,
		UpdatedAt:     r.UpdatedAt,
	}
}
