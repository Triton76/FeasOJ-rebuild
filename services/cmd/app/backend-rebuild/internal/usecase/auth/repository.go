// Layer: Usecase (业务逻辑层)
// Responsibility: 定义认证领域的领域模型(User)和存储接口(UserRepository)
// Dependency: 不依赖具体数据库技术，只定义接口契约
package auth

import (
	"context"
	"time"
)

type User struct {
	ID           string
	Username     string
	Email        string
	PasswordHash string
	Role         string
	Avatar       string
	Synopsis     string
	Score        int
	Status       string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type UserRepository interface {
	Create(ctx context.Context, user User) (User, error)
	GetByUsername(ctx context.Context, username string) (User, error)
	GetByID(ctx context.Context, id string) (User, error)
	UpdatePasswordByEmail(ctx context.Context, email, passwordHash string, updatedAt time.Time) (bool, error)
}
