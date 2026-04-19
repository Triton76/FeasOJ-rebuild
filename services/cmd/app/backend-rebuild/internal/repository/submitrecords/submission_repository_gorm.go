package submitrecordsrepo

import (
	submitrecordsusecase "FeasOJ/app/backend-rebuild/internal/usecase/submitrecords"
	"context"
	"database/sql"
	"time"

	"gorm.io/gorm"
)

type submissionRow struct {
	ID          int64         `gorm:"column:id"`
	UserID      string        `gorm:"column:user_id"`
	ProblemID   int64         `gorm:"column:problem_id"`
	ContestID   sql.NullInt64 `gorm:"column:contest_id"`
	Language    string        `gorm:"column:language"`
	SourceCode  string        `gorm:"column:source_code"`
	Result      string        `gorm:"column:result"`
	Score       sql.NullInt64 `gorm:"column:score"`
	SubmittedAt time.Time     `gorm:"column:submitted_at"`
	CreatedAt   time.Time     `gorm:"column:created_at"`
	UpdatedAt   time.Time     `gorm:"column:updated_at"`
}

func (submissionRow) TableName() string { return "submissions" }

type Repository struct{ db *gorm.DB }

func NewRepository(db *gorm.DB) *Repository { return &Repository{db: db} }

func (r *Repository) Create(ctx context.Context, s submitrecordsusecase.Submission) (submitrecordsusecase.Submission, error) {
	row := toRow(s)
	if err := r.db.WithContext(ctx).Create(&row).Error; err != nil {
		return submitrecordsusecase.Submission{}, err
	}
	return fromRow(row), nil
}

func (r *Repository) List(ctx context.Context, req submitrecordsusecase.Query) ([]submitrecordsusecase.Submission, error) {
	q := r.db.WithContext(ctx).Model(&submissionRow{})
	if req.UserID != "" {
		q = q.Where("user_id = ?", req.UserID)
	}
	if req.ProblemID > 0 {
		q = q.Where("problem_id = ?", req.ProblemID)
	}
	if req.ContestID > 0 {
		q = q.Where("contest_id = ?", req.ContestID)
	}

	var rows []submissionRow
	err := q.Order("submitted_at DESC").Offset(req.Offset).Limit(req.Limit).Find(&rows).Error
	if err != nil {
		return nil, err
	}
	resp := make([]submitrecordsusecase.Submission, 0, len(rows))
	for _, row := range rows {
		resp = append(resp, fromRow(row))
	}
	return resp, nil
}

func toRow(s submitrecordsusecase.Submission) submissionRow {
	row := submissionRow{ID: s.ID, UserID: s.UserID, ProblemID: s.ProblemID, Language: s.Language, SourceCode: s.SourceCode, Result: s.Result, SubmittedAt: s.SubmittedAt, CreatedAt: s.CreatedAt, UpdatedAt: s.UpdatedAt}
	if s.ContestID > 0 {
		row.ContestID = sql.NullInt64{Int64: s.ContestID, Valid: true}
	}
	if s.Score != 0 {
		row.Score = sql.NullInt64{Int64: int64(s.Score), Valid: true}
	}
	return row
}

func fromRow(r submissionRow) submitrecordsusecase.Submission {
	contestID := int64(0)
	if r.ContestID.Valid {
		contestID = r.ContestID.Int64
	}
	score := 0
	if r.Score.Valid {
		score = int(r.Score.Int64)
	}
	return submitrecordsusecase.Submission{ID: r.ID, UserID: r.UserID, ProblemID: r.ProblemID, ContestID: contestID, Language: r.Language, SourceCode: r.SourceCode, Result: r.Result, Score: score, SubmittedAt: r.SubmittedAt, CreatedAt: r.CreatedAt, UpdatedAt: r.UpdatedAt}
}
