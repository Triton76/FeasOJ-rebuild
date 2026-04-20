package admin

import "context"

type User struct {
	ID       string
	Username string
	Email    string
	Avatar   string
	Synopsis string
	Score    int
	Role     string
	Status   string
}

type UserRepository interface {
	ListUsers(ctx context.Context, offset, limit int) ([]User, error)
	UpdateUserStatus(ctx context.Context, userID, status string) (bool, error)
	UpdateUserRole(ctx context.Context, userID, role string) (bool, error)
	GetByID(ctx context.Context, userID string) (User, error)
}
