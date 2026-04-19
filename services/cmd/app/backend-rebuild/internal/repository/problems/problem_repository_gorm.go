package problemsrepo

import (
	"FeasOJ/app/backend-rebuild/internal/ports"
	problemsusecase "FeasOJ/app/backend-rebuild/internal/usecase/problems"
	"context"
	"errors"

	"gorm.io/gorm"
)

type problemRow struct {
	ID            int64  `gorm:"column:id"`
	Title         string `gorm:"column:title"`
	Content       string `gorm:"column:content"`
	Input         string `gorm:"column:input"`
	Output        string `gorm:"column:output"`
	Difficulty    int    `gorm:"column:difficulty"`
	TimeLimitMS   int    `gorm:"column:time_limit_ms"`
	MemoryLimitMB int    `gorm:"column:memory_limit_mb"`
	OwnerUserID   string `gorm:"column:owner_user_id"`
	ClassID       string `gorm:"column:class_id"`
	Visibility    string `gorm:"column:visibility"`
	Status        string `gorm:"column:status"`
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

func (r *ProblemRepository) List(ctx context.Context, offset, limit int, visibility, status string) ([]problemsusecase.Problem, error) {
	q := r.db.WithContext(ctx).Model(&problemRow{})
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
	}
}
