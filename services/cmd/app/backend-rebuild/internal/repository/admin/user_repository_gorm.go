package adminrepo

import (
	"FeasOJ/app/backend-rebuild/internal/ports"
	adminusecase "FeasOJ/app/backend-rebuild/internal/usecase/admin"
	"context"
	"errors"

	"gorm.io/gorm"
)

type userRow struct {
	ID       string `gorm:"column:id"`
	Username string `gorm:"column:username"`
	Email    string `gorm:"column:email"`
	Avatar   string `gorm:"column:avatar"`
	Synopsis string `gorm:"column:synopsis"`
	Score    int    `gorm:"column:score"`
	Role     string `gorm:"column:role"`
	Status   string `gorm:"column:status"`
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

func (r *UserRepository) ListUsers(ctx context.Context, offset, limit int) ([]adminusecase.User, error) {
	var rows []userRow
	err := r.db.WithContext(ctx).
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&rows).Error
	if err != nil {
		return nil, err
	}

	resp := make([]adminusecase.User, 0, len(rows))
	for _, row := range rows {
		resp = append(resp, fromRow(row))
	}
	return resp, nil
}

func (r *UserRepository) UpdateUserStatus(ctx context.Context, userID, status string) (bool, error) {
	result := r.db.WithContext(ctx).Model(&userRow{}).Where("id = ?", userID).Update("status", status)
	if result.Error != nil {
		return false, result.Error
	}
	return result.RowsAffected > 0, nil
}

func (r *UserRepository) GetByID(ctx context.Context, userID string) (adminusecase.User, error) {
	var row userRow
	err := r.db.WithContext(ctx).Where("id = ?", userID).Take(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return adminusecase.User{}, ports.ErrNotFound
	}
	if err != nil {
		return adminusecase.User{}, err
	}
	return fromRow(row), nil
}

func fromRow(r userRow) adminusecase.User {
	return adminusecase.User{
		ID:       r.ID,
		Username: r.Username,
		Email:    r.Email,
		Avatar:   r.Avatar,
		Synopsis: r.Synopsis,
		Score:    r.Score,
		Role:     r.Role,
		Status:   r.Status,
	}
}
