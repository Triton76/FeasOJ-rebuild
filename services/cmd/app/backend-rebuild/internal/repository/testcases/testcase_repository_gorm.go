package testcasesrepo

import (
	"FeasOJ/app/backend-rebuild/internal/ports"
	testcasesusecase "FeasOJ/app/backend-rebuild/internal/usecase/testcases"
	"context"
	"errors"
	"sort"
	"time"

	"gorm.io/gorm"
)

type testcaseRow struct {
	ID         string    `gorm:"column:id"`
	ProblemID  int64     `gorm:"column:problem_id"`
	InputData  string    `gorm:"column:input_data"`
	OutputData string    `gorm:"column:output_data"`
	IsSample   bool      `gorm:"column:is_sample"`
	SortOrder  int       `gorm:"column:sort_order"`
	CreatedAt  time.Time `gorm:"column:created_at"`
	UpdatedAt  time.Time `gorm:"column:updated_at"`
}

func (testcaseRow) TableName() string { return "test_cases" }

type problemOwnerRow struct {
	ID          int64  `gorm:"column:id"`
	OwnerUserID string `gorm:"column:owner_user_id"`
}

func (problemOwnerRow) TableName() string { return "problems" }

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository { return &Repository{db: db} }

func (r *Repository) GetProblemOwner(ctx context.Context, problemID int64) (string, error) {
	var row problemOwnerRow
	err := r.db.WithContext(ctx).Where("id = ?", problemID).Take(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return "", ports.ErrNotFound
	}
	if err != nil {
		return "", err
	}
	return row.OwnerUserID, nil
}

func (r *Repository) ListByProblem(ctx context.Context, problemID int64) ([]testcasesusecase.Testcase, error) {
	var rows []testcaseRow
	err := r.db.WithContext(ctx).
		Where("problem_id = ?", problemID).
		Order("sort_order ASC").
		Order("id ASC").
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	items := make([]testcasesusecase.Testcase, 0, len(rows))
	for _, row := range rows {
		items = append(items, toEntity(row))
	}
	return items, nil
}

func (r *Repository) Create(ctx context.Context, testcase testcasesusecase.Testcase) (testcasesusecase.Testcase, error) {
	now := time.Now().UTC()
	row := testcaseRow{
		ID:         testcase.ID,
		ProblemID:  testcase.ProblemID,
		InputData:  testcase.InputData,
		OutputData: testcase.OutputData,
		IsSample:   testcase.IsSample,
		SortOrder:  testcase.SortOrder,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	if err := r.db.WithContext(ctx).Create(&row).Error; err != nil {
		return testcasesusecase.Testcase{}, err
	}
	return toEntity(row), nil
}

func (r *Repository) Update(ctx context.Context, testcase testcasesusecase.Testcase) (testcasesusecase.Testcase, error) {
	updates := map[string]any{
		"input_data":  testcase.InputData,
		"output_data": testcase.OutputData,
		"is_sample":   testcase.IsSample,
		"updated_at":  time.Now().UTC(),
	}
	res := r.db.WithContext(ctx).
		Model(&testcaseRow{}).
		Where("id = ? AND problem_id = ?", testcase.ID, testcase.ProblemID).
		Updates(updates)
	if res.Error != nil {
		return testcasesusecase.Testcase{}, res.Error
	}
	if res.RowsAffected == 0 {
		return testcasesusecase.Testcase{}, ports.ErrNotFound
	}
	return r.GetByID(ctx, testcase.ProblemID, testcase.ID)
}

func (r *Repository) Delete(ctx context.Context, problemID int64, testcaseID string) error {
	res := r.db.WithContext(ctx).
		Where("id = ? AND problem_id = ?", testcaseID, problemID).
		Delete(&testcaseRow{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ports.ErrNotFound
	}
	return nil
}

func (r *Repository) GetByID(ctx context.Context, problemID int64, testcaseID string) (testcasesusecase.Testcase, error) {
	var row testcaseRow
	err := r.db.WithContext(ctx).
		Where("id = ? AND problem_id = ?", testcaseID, problemID).
		Take(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return testcasesusecase.Testcase{}, ports.ErrNotFound
	}
	if err != nil {
		return testcasesusecase.Testcase{}, err
	}
	return toEntity(row), nil
}

func (r *Repository) Reorder(ctx context.Context, problemID int64, orderedIDs []string) ([]testcasesusecase.Testcase, error) {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for index, id := range orderedIDs {
			res := tx.Model(&testcaseRow{}).
				Where("id = ? AND problem_id = ?", id, problemID).
				Updates(map[string]any{"sort_order": index + 1, "updated_at": time.Now().UTC()})
			if res.Error != nil {
				return res.Error
			}
			if res.RowsAffected == 0 {
				return ports.ErrNotFound
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	items, err := r.ListByProblem(ctx, problemID)
	if err != nil {
		return nil, err
	}
	sort.SliceStable(items, func(i, j int) bool {
		if items[i].SortOrder != items[j].SortOrder {
			return items[i].SortOrder < items[j].SortOrder
		}
		return items[i].ID < items[j].ID
	})
	return items, nil
}

func toEntity(row testcaseRow) testcasesusecase.Testcase {
	return testcasesusecase.Testcase{
		ID:         row.ID,
		ProblemID:  row.ProblemID,
		InputData:  row.InputData,
		OutputData: row.OutputData,
		IsSample:   row.IsSample,
		SortOrder:  row.SortOrder,
	}
}
