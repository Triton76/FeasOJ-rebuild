package admin

import (
	"FeasOJ/app/backend-rebuild/internal/ports"
	"context"
	"testing"
)

type fakeAdminRepo struct {
	users map[string]User
}

func newFakeAdminRepo() *fakeAdminRepo {
	return &fakeAdminRepo{users: map[string]User{}}
}

func (r *fakeAdminRepo) ListUsers(ctx context.Context, offset, limit int) ([]User, error) {
	resp := make([]User, 0, len(r.users))
	for _, u := range r.users {
		resp = append(resp, u)
	}
	return resp, nil
}

func (r *fakeAdminRepo) UpdateUserStatus(ctx context.Context, userID, status string) (bool, error) {
	u, ok := r.users[userID]
	if !ok {
		return false, nil
	}
	u.Status = status
	r.users[userID] = u
	return true, nil
}

func (r *fakeAdminRepo) UpdateUserRole(ctx context.Context, userID, role string) (bool, error) {
	u, ok := r.users[userID]
	if !ok {
		return false, nil
	}
	u.Role = role
	r.users[userID] = u
	return true, nil
}

func (r *fakeAdminRepo) GetByID(ctx context.Context, userID string) (User, error) {
	u, ok := r.users[userID]
	if !ok {
		return User{}, ports.ErrNotFound
	}
	return u, nil
}

func TestStatusAndRoleUpdatesRemainSeparated(t *testing.T) {
	repo := newFakeAdminRepo()
	repo.users["u-1"] = User{ID: "u-1", Username: "alice", Role: "student", Status: "active"}
	svc := NewService(repo)

	afterStatus, err := svc.UpdateUserStatus(context.Background(), ports.UpdateUserStatusRequest{UserID: "u-1", Status: "banned", ActorUserID: "admin-1"})
	if err != nil {
		t.Fatalf("UpdateUserStatus failed: %v", err)
	}
	if afterStatus.Status != "banned" {
		t.Fatalf("expected status=banned, got %s", afterStatus.Status)
	}
	if afterStatus.Role != "student" {
		t.Fatalf("expected role unchanged=student, got %s", afterStatus.Role)
	}

	afterRole, err := svc.UpdateUserRole(context.Background(), ports.UpdateUserRoleRequest{UserID: "u-1", Role: "teacher", ActorUserID: "admin-1"})
	if err != nil {
		t.Fatalf("UpdateUserRole failed: %v", err)
	}
	if afterRole.Role != "teacher" {
		t.Fatalf("expected role=teacher, got %s", afterRole.Role)
	}
	if afterRole.Status != "banned" {
		t.Fatalf("expected status unchanged=banned, got %s", afterRole.Status)
	}
}
