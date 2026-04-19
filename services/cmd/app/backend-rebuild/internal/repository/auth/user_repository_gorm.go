// Layer: Repository (数据访问层/适配器层)
// Responsibility: 实现 Usecase 定义的 Repository 接口，提供具体的数据库操作(Gorm)
// Dependency: 依赖 Usecase 层的领域模型和 Repository 接口
package authrepo

import (
	"FeasOJ/app/backend-rebuild/internal/ports"
	authusecase "FeasOJ/app/backend-rebuild/internal/usecase/auth"
	"context"
	"errors"
	"time"

	"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
)

type userRow struct {
	ID           string    `gorm:"column:id"`
	Username     string    `gorm:"column:username"`
	Email        string    `gorm:"column:email"`
	PasswordHash string    `gorm:"column:password_hash"`
	Role         string    `gorm:"column:role"`
	Avatar       string    `gorm:"column:avatar"`
	Synopsis     string    `gorm:"column:synopsis"`
	Score        int       `gorm:"column:score"`
	Status       string    `gorm:"column:status"`
	CreatedAt    time.Time `gorm:"column:created_at"`
	UpdatedAt    time.Time `gorm:"column:updated_at"`
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

func (r *UserRepository) Create(ctx context.Context, user authusecase.User) (authusecase.User, error) {
	row := toRow(user)
	if err := r.db.WithContext(ctx).Create(&row).Error; err != nil {
		if isDuplicate(err) {
			return authusecase.User{}, ports.ErrConflict
		}
		return authusecase.User{}, err
	}
	return fromRow(row), nil
}

func (r *UserRepository) GetByUsername(ctx context.Context, username string) (authusecase.User, error) {
	var row userRow
	err := r.db.WithContext(ctx).Where("username = ?", username).Take(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return authusecase.User{}, ports.ErrNotFound
	}
	if err != nil {
		return authusecase.User{}, err
	}
	return fromRow(row), nil
}

func (r *UserRepository) GetByID(ctx context.Context, id string) (authusecase.User, error) {
	var row userRow
	err := r.db.WithContext(ctx).Where("id = ?", id).Take(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return authusecase.User{}, ports.ErrNotFound
	}
	if err != nil {
		return authusecase.User{}, err
	}
	return fromRow(row), nil
}

func (r *UserRepository) UpdatePasswordByEmail(ctx context.Context, email, passwordHash string, updatedAt time.Time) (bool, error) {
	result := r.db.WithContext(ctx).Model(&userRow{}).
		Where("email = ?", email).
		Updates(map[string]any{"password_hash": passwordHash, "updated_at": updatedAt})
	if result.Error != nil {
		return false, result.Error
	}
	return result.RowsAffected > 0, nil
}

func toRow(u authusecase.User) userRow {
	return userRow{
		ID:           u.ID,
		Username:     u.Username,
		Email:        u.Email,
		PasswordHash: u.PasswordHash,
		Role:         u.Role,
		Avatar:       u.Avatar,
		Synopsis:     u.Synopsis,
		Score:        u.Score,
		Status:       u.Status,
		CreatedAt:    u.CreatedAt,
		UpdatedAt:    u.UpdatedAt,
	}
}

func fromRow(r userRow) authusecase.User {
	return authusecase.User{
		ID:           r.ID,
		Username:     r.Username,
		Email:        r.Email,
		PasswordHash: r.PasswordHash,
		Role:         r.Role,
		Avatar:       r.Avatar,
		Synopsis:     r.Synopsis,
		Score:        r.Score,
		Status:       r.Status,
		CreatedAt:    r.CreatedAt,
		UpdatedAt:    r.UpdatedAt,
	}
}

func isDuplicate(err error) bool {
	var mysqlErr *mysql.MySQLError
	if errors.As(err, &mysqlErr) {
		return mysqlErr.Number == 1062
	}
	return false
}
