package competitionsrepo

import (
	"FeasOJ/app/backend-rebuild/internal/ports"
	competitionsusecase "FeasOJ/app/backend-rebuild/internal/usecase/competitions"
	"context"
	"errors"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
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
	PasswordHash string     `gorm:"column:password_hash"`
	StartAt      *time.Time `gorm:"column:start_at"`
	EndAt        *time.Time `gorm:"column:end_at"`
	CreatedAt    time.Time  `gorm:"column:created_at"`
	UpdatedAt    time.Time  `gorm:"column:updated_at"`
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

type contestProblemRow struct {
	ID           string    `gorm:"column:id"`
	ContestID    int64     `gorm:"column:contest_id"`
	ProblemID    int64     `gorm:"column:problem_id"`
	DisplayOrder int       `gorm:"column:display_order"`
	Alias        string    `gorm:"column:alias"`
	CreatedAt    time.Time `gorm:"column:created_at"`
	UpdatedAt    time.Time `gorm:"column:updated_at"`
}

type scoreboardSubmissionRow struct {
	UserID      string    `gorm:"column:user_id"`
	Username    string    `gorm:"column:username"`
	ProblemID   int64     `gorm:"column:problem_id"`
	Result      string    `gorm:"column:result"`
	Score       int       `gorm:"column:score"`
	SubmittedAt time.Time `gorm:"column:submitted_at"`
}

func (participantRow) TableName() string { return "contest_participants" }
func (contestProblemRow) TableName() string { return "contest_problems" }

type Repository struct{ db *gorm.DB }

func NewRepository(db *gorm.DB) *Repository { return &Repository{db: db} }

func (r *Repository) ListVisible(ctx context.Context, offset, limit int, visibility, ruleType, status, actorUserID, actorRole string) ([]competitionsusecase.Contest, error) {
	q := r.db.WithContext(ctx).Model(&contestRow{})
	q = applyContestVisibilityScope(q, actorUserID, actorRole)
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

func (r *Repository) GetVisibleByID(ctx context.Context, contestID int64, actorUserID, actorRole string) (competitionsusecase.Contest, error) {
	q := r.db.WithContext(ctx).Model(&contestRow{}).Where("id = ?", contestID)
	q = applyContestVisibilityScope(q, actorUserID, actorRole)

	var row contestRow
	err := q.Take(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return competitionsusecase.Contest{}, ports.ErrNotFound
	}
	if err != nil {
		return competitionsusecase.Contest{}, err
	}
	return toContest(row), nil
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

func (r *Repository) Create(ctx context.Context, c competitionsusecase.Contest) (competitionsusecase.Contest, error) {
	row := contestRow{
		Title:        c.Title,
		Subtitle:     c.Subtitle,
		Description:  c.Description,
		Announcement: c.Announcement,
		OwnerUserID:  c.OwnerUserID,
		ClassID:      c.ClassID,
		Visibility:   c.Visibility,
		RuleType:     c.RuleType,
		Status:       c.Status,
		IsEncrypted:  c.IsEncrypted,
		PasswordHash: c.PasswordHash,
		StartAt:      c.StartAt,
		EndAt:        c.EndAt,
		CreatedAt:    c.CreatedAt,
		UpdatedAt:    c.UpdatedAt,
	}
	if err := r.db.WithContext(ctx).Create(&row).Error; err != nil {
		return competitionsusecase.Contest{}, err
	}
	return toContest(row), nil
}

func (r *Repository) Update(ctx context.Context, c competitionsusecase.Contest) (competitionsusecase.Contest, error) {
	updates := map[string]any{
		"title":         c.Title,
		"subtitle":      c.Subtitle,
		"description":   c.Description,
		"announcement":  c.Announcement,
		"class_id":      c.ClassID,
		"visibility":    c.Visibility,
		"rule_type":     c.RuleType,
		"status":        c.Status,
		"is_encrypted":  c.IsEncrypted,
		"password_hash": c.PasswordHash,
		"start_at":      c.StartAt,
		"end_at":        c.EndAt,
		"updated_at":    c.UpdatedAt,
	}
	res := r.db.WithContext(ctx).Model(&contestRow{}).Where("id = ?", c.ID).Updates(updates)
	if res.Error != nil {
		return competitionsusecase.Contest{}, res.Error
	}
	if res.RowsAffected == 0 {
		return competitionsusecase.Contest{}, ports.ErrNotFound
	}
	return r.GetByID(ctx, c.ID)
}

func (r *Repository) Delete(ctx context.Context, contestID int64) error {
	res := r.db.WithContext(ctx).Where("id = ?", contestID).Delete(&contestRow{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ports.ErrNotFound
	}
	return nil
}

func (r *Repository) ListProblemBindings(ctx context.Context, contestID int64) ([]competitionsusecase.ContestProblemBinding, error) {
	var rows []contestProblemRow
	err := r.db.WithContext(ctx).
		Where("contest_id = ?", contestID).
		Order("display_order ASC").
		Find(&rows).Error
	if err != nil {
		return nil, err
	}

	resp := make([]competitionsusecase.ContestProblemBinding, 0, len(rows))
	for _, row := range rows {
		resp = append(resp, competitionsusecase.ContestProblemBinding{
			ContestID:    row.ContestID,
			ProblemID:    row.ProblemID,
			DisplayOrder: row.DisplayOrder,
			Alias:        row.Alias,
			CreatedAt:    row.CreatedAt,
			UpdatedAt:    row.UpdatedAt,
		})
	}
	return resp, nil
}

func (r *Repository) ReplaceProblemBindings(ctx context.Context, contestID int64, items []competitionsusecase.ContestProblemBinding) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("contest_id = ?", contestID).Delete(&contestProblemRow{}).Error; err != nil {
			return err
		}
		if len(items) == 0 {
			return nil
		}

		rows := make([]contestProblemRow, 0, len(items))
		for _, item := range items {
			rows = append(rows, contestProblemRow{
				ID:           uuid.NewString(),
				ContestID:    contestID,
				ProblemID:    item.ProblemID,
				DisplayOrder: item.DisplayOrder,
				Alias:        strings.TrimSpace(item.Alias),
				CreatedAt:    item.CreatedAt,
				UpdatedAt:    item.UpdatedAt,
			})
		}

		if err := tx.Create(&rows).Error; err != nil {
			if isDuplicate(err) {
				return ports.ErrConflict
			}
			return err
		}
		return nil
	})
}

func (r *Repository) CountProblemBindings(ctx context.Context, contestID int64) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&contestProblemRow{}).Where("contest_id = ?", contestID).Count(&count).Error
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (r *Repository) CountExistingProblems(ctx context.Context, problemIDs []int64) (int64, error) {
	if len(problemIDs) == 0 {
		return 0, nil
	}
	var count int64
	err := r.db.WithContext(ctx).Table("problems").Where("id IN ?", problemIDs).Count(&count).Error
	if err != nil {
		return 0, err
	}
	return count, nil
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

func (r *Repository) ListScoreboardSubmissions(ctx context.Context, contestID int64, before time.Time) ([]competitionsusecase.ScoreboardSubmission, error) {
	var rows []scoreboardSubmissionRow
	err := r.db.WithContext(ctx).
		Table("submissions AS s").
		Select("s.user_id, u.username, s.problem_id, s.result, COALESCE(s.score, 0) AS score, s.submitted_at").
		Joins("JOIN users AS u ON u.id = s.user_id").
		Where("s.contest_id = ? AND s.submitted_at < ?", contestID, before).
		Order("s.submitted_at ASC").
		Find(&rows).Error
	if err != nil {
		return nil, err
	}

	resp := make([]competitionsusecase.ScoreboardSubmission, 0, len(rows))
	for _, row := range rows {
		resp = append(resp, competitionsusecase.ScoreboardSubmission{
			UserID:      row.UserID,
			Username:    row.Username,
			ProblemID:   row.ProblemID,
			Result:      row.Result,
			Score:       row.Score,
			SubmittedAt: row.SubmittedAt,
		})
	}
	return resp, nil
}

func toContest(r contestRow) competitionsusecase.Contest {
	return competitionsusecase.Contest{ID: r.ID, Title: r.Title, Subtitle: r.Subtitle, Description: r.Description, Announcement: r.Announcement, OwnerUserID: r.OwnerUserID, ClassID: r.ClassID, Visibility: r.Visibility, RuleType: r.RuleType, Status: r.Status, IsEncrypted: r.IsEncrypted, PasswordHash: r.PasswordHash, StartAt: r.StartAt, EndAt: r.EndAt, CreatedAt: r.CreatedAt, UpdatedAt: r.UpdatedAt}
}

func applyContestVisibilityScope(q *gorm.DB, actorUserID, actorRole string) *gorm.DB {
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

func isDuplicate(err error) bool {
	var mysqlErr *mysql.MySQLError
	return errors.As(err, &mysqlErr) && mysqlErr.Number == 1062
}
