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
	FindClassByID(ctx context.Context, classID string) (Class, error)
	CreateMembership(ctx context.Context, m Membership) (Membership, error)
	FindClassByCode(ctx context.Context, code string) (Class, error)
	FindMembershipByID(ctx context.Context, membershipID string) (Membership, error)
	FindMembershipByClassAndUser(ctx context.Context, classID, userID string) (Membership, error)
	ListMembershipsByClassID(ctx context.Context, classID string) ([]Membership, error)
	ListMembershipsByUserID(ctx context.Context, userID string) ([]Membership, error)
	UpdateClass(ctx context.Context, classID, ownerUserID, name, description string, updatedAt time.Time) (bool, error)
	ArchiveClass(ctx context.Context, classID, ownerUserID string, updatedAt time.Time) (bool, error)
	UpdateMembershipStatus(ctx context.Context, membershipID, status string, joinedAt *time.Time, updatedAt time.Time) (bool, error)
}
