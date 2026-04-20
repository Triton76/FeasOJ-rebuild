// Layer: Repository (数据访问层/适配器层)
// Responsibility: 实现 Usecase 定义的 Repository 接口，提供具体的数据库操作(Gorm)
// Dependency: 依赖 Usecase 层的领域模型和 Repository 接口
package usersrepo

import (
	"FeasOJ/app/backend-rebuild/internal/ports"
	usersusecase "FeasOJ/app/backend-rebuild/internal/usecase/users"
	"context"
	"errors"
	"time"

	"gorm.io/gorm"
)

type userRow struct {
	ID        string    `gorm:"column:id"`
	Username  string    `gorm:"column:username"`
	Email     string    `gorm:"column:email"`
	Avatar    string    `gorm:"column:avatar"`
	Synopsis  string    `gorm:"column:synopsis"`
	Score     int       `gorm:"column:score"`
	Role      string    `gorm:"column:role"`
	Status    string    `gorm:"column:status"`
	CreatedAt time.Time `gorm:"column:created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at"`
}

func (userRow) TableName() string {
	return "users"
}

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) GetByID(ctx context.Context, userID string) (usersusecase.User, error) {
	var row userRow
	err := r.db.WithContext(ctx).Where("id = ?", userID).Take(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return usersusecase.User{}, ports.ErrNotFound
	}
	if err != nil {
		return usersusecase.User{}, err
	}
	return fromRow(row), nil
}

func (r *UserRepository) UpdateProfile(ctx context.Context, userID string, avatar, synopsis *string, updatedAt time.Time) (bool, error) {
	updates := map[string]any{
		"updated_at": updatedAt,
	}
	if avatar != nil {
		updates["avatar"] = *avatar
	}
	if synopsis != nil {
		updates["synopsis"] = *synopsis
	}
	result := r.db.WithContext(ctx).Model(&userRow{}).Where("id = ?", userID).Updates(updates)
	if result.Error != nil {
		return false, result.Error
	}
	if result.RowsAffected > 0 {
		return true, nil
	}

	var count int64
	if err := r.db.WithContext(ctx).Model(&userRow{}).Where("id = ?", userID).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *UserRepository) ListRanking(ctx context.Context, offset, limit int) ([]usersusecase.User, error) {
	var rows []userRow
	err := r.db.WithContext(ctx).
		Order("score DESC").
		Order("created_at ASC").
		Offset(offset).
		Limit(limit).
		Find(&rows).Error
	if err != nil {
		return nil, err
	}

	result := make([]usersusecase.User, 0, len(rows))
	for _, row := range rows {
		result = append(result, fromRow(row))
	}
	return result, nil
}

func fromRow(r userRow) usersusecase.User {
	return usersusecase.User{
		ID:        r.ID,
		Username:  r.Username,
		Email:     r.Email,
		Avatar:    r.Avatar,
		Synopsis:  r.Synopsis,
		Score:     r.Score,
		Role:      r.Role,
		Status:    r.Status,
		CreatedAt: r.CreatedAt,
		UpdatedAt: r.UpdatedAt,
	}
}
