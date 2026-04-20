package classesrepo

import (
	"FeasOJ/app/backend-rebuild/internal/ports"
	classesusecase "FeasOJ/app/backend-rebuild/internal/usecase/classes"
	"context"
	"errors"
	"time"

	"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
)

type classRow struct {
	ID          string    `gorm:"column:id"`
	Name        string    `gorm:"column:name"`
	Code        string    `gorm:"column:code"`
	Description string    `gorm:"column:description"`
	OwnerUserID string    `gorm:"column:owner_user_id"`
	Status      string    `gorm:"column:status"`
	CreatedAt   time.Time `gorm:"column:created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at"`
}

func (classRow) TableName() string { return "classes" }

type membershipRow struct {
	ID          string     `gorm:"column:id"`
	ClassID     string     `gorm:"column:class_id"`
	UserID      string     `gorm:"column:user_id"`
	RoleInClass string     `gorm:"column:role_in_class"`
	Status      string     `gorm:"column:status"`
	JoinedAt    *time.Time `gorm:"column:joined_at"`
	CreatedAt   time.Time  `gorm:"column:created_at"`
	UpdatedAt   time.Time  `gorm:"column:updated_at"`
}

func (membershipRow) TableName() string { return "class_memberships" }

type Repository struct{ db *gorm.DB }

func NewRepository(db *gorm.DB) *Repository { return &Repository{db: db} }

func (r *Repository) CreateClass(ctx context.Context, c classesusecase.Class) (classesusecase.Class, error) {
	row := classRow{ID: c.ID, Name: c.Name, Code: c.Code, Description: c.Description, OwnerUserID: c.OwnerUserID, Status: c.Status, CreatedAt: c.CreatedAt, UpdatedAt: c.UpdatedAt}
	if err := r.db.WithContext(ctx).Create(&row).Error; err != nil {
		if isDuplicate(err) {
			return classesusecase.Class{}, ports.ErrConflict
		}
		return classesusecase.Class{}, err
	}
	return toClass(row), nil
}

func (r *Repository) FindClassByID(ctx context.Context, classID string) (classesusecase.Class, error) {
	var row classRow
	err := r.db.WithContext(ctx).Where("id = ?", classID).Take(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return classesusecase.Class{}, ports.ErrNotFound
	}
	if err != nil {
		return classesusecase.Class{}, err
	}
	return toClass(row), nil
}

func (r *Repository) CreateMembership(ctx context.Context, m classesusecase.Membership) (classesusecase.Membership, error) {
	row := membershipRow{ID: m.ID, ClassID: m.ClassID, UserID: m.UserID, RoleInClass: m.RoleInClass, Status: m.Status, JoinedAt: m.JoinedAt, CreatedAt: m.CreatedAt, UpdatedAt: m.UpdatedAt}
	if err := r.db.WithContext(ctx).Create(&row).Error; err != nil {
		if isDuplicate(err) {
			return classesusecase.Membership{}, ports.ErrConflict
		}
		return classesusecase.Membership{}, err
	}
	return toMembership(row), nil
}

func (r *Repository) FindClassByCode(ctx context.Context, code string) (classesusecase.Class, error) {
	var row classRow
	err := r.db.WithContext(ctx).Where("code = ?", code).Take(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return classesusecase.Class{}, ports.ErrNotFound
	}
	if err != nil {
		return classesusecase.Class{}, err
	}
	return toClass(row), nil
}

func (r *Repository) UpdateClass(ctx context.Context, classID, ownerUserID, name, description string, updatedAt time.Time) (bool, error) {
	updates := map[string]any{
		"name":        name,
		"description": description,
		"updated_at":  updatedAt,
	}
	result := r.db.WithContext(ctx).
		Model(&classRow{}).
		Where("id = ? AND owner_user_id = ? AND status <> ?", classID, ownerUserID, "archived").
		Updates(updates)
	if result.Error != nil {
		if isDuplicate(result.Error) {
			return false, ports.ErrConflict
		}
		return false, result.Error
	}
	return result.RowsAffected > 0, nil
}

func (r *Repository) ArchiveClass(ctx context.Context, classID, ownerUserID string, updatedAt time.Time) (bool, error) {
	updates := map[string]any{
		"status":     "archived",
		"updated_at": updatedAt,
	}
	result := r.db.WithContext(ctx).
		Model(&classRow{}).
		Where("id = ? AND owner_user_id = ? AND status <> ?", classID, ownerUserID, "archived").
		Updates(updates)
	if result.Error != nil {
		return false, result.Error
	}
	return result.RowsAffected > 0, nil
}

func (r *Repository) FindMembershipByID(ctx context.Context, membershipID string) (classesusecase.Membership, error) {
	var row membershipRow
	err := r.db.WithContext(ctx).Where("id = ?", membershipID).Take(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return classesusecase.Membership{}, ports.ErrNotFound
	}
	if err != nil {
		return classesusecase.Membership{}, err
	}
	return toMembership(row), nil
}

func (r *Repository) FindMembershipByClassAndUser(ctx context.Context, classID, userID string) (classesusecase.Membership, error) {
	var row membershipRow
	err := r.db.WithContext(ctx).Where("class_id = ? AND user_id = ?", classID, userID).Take(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return classesusecase.Membership{}, ports.ErrNotFound
	}
	if err != nil {
		return classesusecase.Membership{}, err
	}
	return toMembership(row), nil
}

func (r *Repository) ListMembershipsByClassID(ctx context.Context, classID string) ([]classesusecase.Membership, error) {
	var rows []membershipRow
	if err := r.db.WithContext(ctx).Where("class_id = ?", classID).Order("created_at ASC").Find(&rows).Error; err != nil {
		return nil, err
	}
	items := make([]classesusecase.Membership, 0, len(rows))
	for _, row := range rows {
		items = append(items, toMembership(row))
	}
	return items, nil
}

func (r *Repository) ListMembershipsByUserID(ctx context.Context, userID string) ([]classesusecase.Membership, error) {
	var rows []membershipRow
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).Order("created_at ASC").Find(&rows).Error; err != nil {
		return nil, err
	}
	items := make([]classesusecase.Membership, 0, len(rows))
	for _, row := range rows {
		items = append(items, toMembership(row))
	}
	return items, nil
}

func (r *Repository) UpdateMembershipStatus(ctx context.Context, membershipID, status string, joinedAt *time.Time, updatedAt time.Time) (bool, error) {
	updates := map[string]any{"status": status, "updated_at": updatedAt, "joined_at": joinedAt}
	result := r.db.WithContext(ctx).Model(&membershipRow{}).Where("id = ? AND status = ?", membershipID, "pending").Updates(updates)
	if result.Error != nil {
		return false, result.Error
	}
	return result.RowsAffected > 0, nil
}

func toClass(r classRow) classesusecase.Class {
	return classesusecase.Class{ID: r.ID, Name: r.Name, Code: r.Code, Description: r.Description, OwnerUserID: r.OwnerUserID, Status: r.Status, CreatedAt: r.CreatedAt, UpdatedAt: r.UpdatedAt}
}

func toMembership(r membershipRow) classesusecase.Membership {
	return classesusecase.Membership{ID: r.ID, ClassID: r.ClassID, UserID: r.UserID, RoleInClass: r.RoleInClass, Status: r.Status, JoinedAt: r.JoinedAt, CreatedAt: r.CreatedAt, UpdatedAt: r.UpdatedAt}
}

func isDuplicate(err error) bool {
	var mysqlErr *mysql.MySQLError
	return errors.As(err, &mysqlErr) && mysqlErr.Number == 1062
}
