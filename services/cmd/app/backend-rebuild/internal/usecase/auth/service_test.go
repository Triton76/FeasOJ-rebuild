package auth

import (
	"FeasOJ/app/backend-rebuild/internal/ports"
	"FeasOJ/app/backend-rebuild/internal/security"
	"context"
	"testing"
	"time"
)

type fakeAuthRepo struct {
	usersByID       map[string]User
	usersByUsername map[string]User
}

func newFakeAuthRepo() *fakeAuthRepo {
	return &fakeAuthRepo{
		usersByID:       make(map[string]User),
		usersByUsername: make(map[string]User),
	}
}

func (r *fakeAuthRepo) Create(ctx context.Context, user User) (User, error) {
	if _, ok := r.usersByUsername[user.Username]; ok {
		return User{}, ports.ErrConflict
	}
	r.usersByID[user.ID] = user
	r.usersByUsername[user.Username] = user
	return user, nil
}

func (r *fakeAuthRepo) GetByUsername(ctx context.Context, username string) (User, error) {
	u, ok := r.usersByUsername[username]
	if !ok {
		return User{}, ports.ErrNotFound
	}
	return u, nil
}

func (r *fakeAuthRepo) GetByID(ctx context.Context, id string) (User, error) {
	u, ok := r.usersByID[id]
	if !ok {
		return User{}, ports.ErrNotFound
	}
	return u, nil
}

func (r *fakeAuthRepo) UpdatePasswordByEmail(ctx context.Context, email, passwordHash string, updatedAt time.Time) (bool, error) {
	return false, ports.ErrNotImplemented
}

func TestRegisterAndLoginAndVerify(t *testing.T) {
	repo := newFakeAuthRepo()
	svc := NewService(repo, "test-secret", "test-issuer", 2*time.Hour)

	registered, err := svc.Register(context.Background(), ports.RegisterRequest{
		Username: "alice",
		Email:    "alice@example.com",
		Password: "password123",
	})
	if err != nil {
		t.Fatalf("register failed: %v", err)
	}
	if registered.Username != "alice" {
		t.Fatalf("unexpected username: %s", registered.Username)
	}

	loginResp, err := svc.Login(context.Background(), ports.LoginRequest{
		Username: "alice",
		Password: "password123",
	})
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}
	if loginResp.Token == "" {
		t.Fatalf("expected non-empty token")
	}

	claims, err := security.ParseToken("test-secret", loginResp.Token)
	if err != nil {
		t.Fatalf("parse token failed: %v", err)
	}
	ctx := security.WithClaims(context.Background(), claims)

	verifyResp, err := svc.Verify(ctx)
	if err != nil {
		t.Fatalf("verify failed: %v", err)
	}
	if verifyResp.User.ID != registered.ID {
		t.Fatalf("verify user mismatch: got %s want %s", verifyResp.User.ID, registered.ID)
	}
}

func TestRegisterInvalidEmail(t *testing.T) {
	repo := newFakeAuthRepo()
	svc := NewService(repo, "test-secret", "test-issuer", 2*time.Hour)

	_, err := svc.Register(context.Background(), ports.RegisterRequest{
		Username: "alice",
		Email:    "not-an-email",
		Password: "password123",
	})
	if err != ports.ErrInvalidArgument {
		t.Fatalf("expected ErrInvalidArgument, got %v", err)
	}
}
