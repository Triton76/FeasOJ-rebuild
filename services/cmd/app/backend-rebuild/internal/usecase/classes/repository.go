package classes

import (
	"context"
	"time"
)

type Class struct {
	ID          string
	Name        string
	Code        string
	Description string
	OwnerUserID string
	Status      string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type Membership struct {
	ID          string
	ClassID     string
	UserID      string
	RoleInClass string
	Status      string
	JoinedAt    *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type Repository interface {
	CreateClass(ctx context.Context, c Class) (Class, error)
	CreateMembership(ctx context.Context, m Membership) (Membership, error)
	FindClassByCode(ctx context.Context, code string) (Class, error)
	FindMembershipByID(ctx context.Context, membershipID string) (Membership, error)
	UpdateMembershipStatus(ctx context.Context, membershipID, status string, joinedAt *time.Time, updatedAt time.Time) (bool, error)
}
