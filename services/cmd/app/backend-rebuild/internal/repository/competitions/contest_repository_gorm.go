package competitionsrepo

import (
	"FeasOJ/app/backend-rebuild/internal/ports"
	competitionsusecase "FeasOJ/app/backend-rebuild/internal/usecase/competitions"
	"context"
	"errors"
	"time"

	"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
)

type contestRow struct {
	ID           int64      `gorm:"column:id"`
	Title        string     `gorm:"column:title"`
	Subtitle     string     `gorm:"column:subtitle"`
	Description  string     `gorm:"column:description"`
	Announcement string     `gorm:"column:announcement"`
	OwnerUserID  string     `gorm:"column:owner_user_id"`
	ClassID      string     `gorm:"column:class_id"`
	Visibility   string     `gorm:"column:visibility"`
	RuleType     string     `gorm:"column:rule_type"`
	Status       string     `gorm:"column:status"`
	IsEncrypted  bool       `gorm:"column:is_encrypted"`
	StartAt      *time.Time `gorm:"column:start_at"`
	EndAt        *time.Time `gorm:"column:end_at"`
}

func (contestRow) TableName() string { return "contests" }

type participantRow struct {
	ID        string     `gorm:"column:id"`
	ContestID int64      `gorm:"column:contest_id"`
	UserID    string     `gorm:"column:user_id"`
	Status    string     `gorm:"column:status"`
	JoinedAt  *time.Time `gorm:"column:joined_at"`
	CreatedAt time.Time  `gorm:"column:created_at"`
	UpdatedAt time.Time  `gorm:"column:updated_at"`
}

func (participantRow) TableName() string { return "contest_participants" }

type Repository struct{ db *gorm.DB }

func NewRepository(db *gorm.DB) *Repository { return &Repository{db: db} }

func (r *Repository) List(ctx context.Context, offset, limit int, visibility, ruleType, status string) ([]competitionsusecase.Contest, error) {
	q := r.db.WithContext(ctx).Model(&contestRow{})
	if visibility != "" {
		q = q.Where("visibility = ?", visibility)
	}
	if ruleType != "" {
		q = q.Where("rule_type = ?", ruleType)
	}
	if status != "" {
		q = q.Where("status = ?", status)
	}

	var rows []contestRow
	err := q.Order("start_at DESC").Offset(offset).Limit(limit).Find(&rows).Error
	if err != nil {
		return nil, err
	}

	resp := make([]competitionsusecase.Contest, 0, len(rows))
	for _, row := range rows {
		resp = append(resp, toContest(row))
	}
	return resp, nil
}

func (r *Repository) GetByID(ctx context.Context, contestID int64) (competitionsusecase.Contest, error) {
	var row contestRow
	err := r.db.WithContext(ctx).Where("id = ?", contestID).Take(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return competitionsusecase.Contest{}, ports.ErrNotFound
	}
	if err != nil {
		return competitionsusecase.Contest{}, err
	}
	return toContest(row), nil
}

func (r *Repository) CreateParticipant(ctx context.Context, p competitionsusecase.Participant) (competitionsusecase.Participant, error) {
	row := participantRow{ID: p.ID, ContestID: p.ContestID, UserID: p.UserID, Status: p.Status, JoinedAt: p.JoinedAt, CreatedAt: p.CreatedAt, UpdatedAt: p.UpdatedAt}
	if err := r.db.WithContext(ctx).Create(&row).Error; err != nil {
		if isDuplicate(err) {
			return competitionsusecase.Participant{}, ports.ErrConflict
		}
		return competitionsusecase.Participant{}, err
	}
	return competitionsusecase.Participant{ID: row.ID, ContestID: row.ContestID, UserID: row.UserID, Status: row.Status, JoinedAt: row.JoinedAt, CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt}, nil
}

func toContest(r contestRow) competitionsusecase.Contest {
	return competitionsusecase.Contest{ID: r.ID, Title: r.Title, Subtitle: r.Subtitle, Description: r.Description, Announcement: r.Announcement, OwnerUserID: r.OwnerUserID, ClassID: r.ClassID, Visibility: r.Visibility, RuleType: r.RuleType, Status: r.Status, IsEncrypted: r.IsEncrypted, StartAt: r.StartAt, EndAt: r.EndAt}
}

func isDuplicate(err error) bool {
	var mysqlErr *mysql.MySQLError
	return errors.As(err, &mysqlErr) && mysqlErr.Number == 1062
}
