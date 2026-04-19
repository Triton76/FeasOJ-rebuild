// Layer: Usecase (业务逻辑层)
// Responsibility: 定义用户领域的领域模型(User)和存储接口(UserRepository)
// Dependency: 不依赖具体数据库技术，只定义接口契约
package users

import (
	"context"
	"time"
)

type User struct {
	ID        string
	Username  string
	Email     string
	Avatar    string
	Synopsis  string
	Score     int
	Role      string
	Status    string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type UserRepository interface {
	GetByID(ctx context.Context, userID string) (User, error)
	UpdateProfile(ctx context.Context, userID, avatar, synopsis string, updatedAt time.Time) (bool, error)
	ListRanking(ctx context.Context, offset, limit int) ([]User, error)
}
