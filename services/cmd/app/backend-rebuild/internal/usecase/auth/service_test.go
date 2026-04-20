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
	usersByEmail    map[string]User
}

func newFakeAuthRepo() *fakeAuthRepo {
	return &fakeAuthRepo{
		usersByID:       make(map[string]User),
		usersByUsername: make(map[string]User),
		usersByEmail:    make(map[string]User),
	}
}

func (r *fakeAuthRepo) Create(ctx context.Context, user User) (User, error) {
	if _, ok := r.usersByUsername[user.Username]; ok {
		return User{}, ports.ErrConflict
	}
	r.usersByID[user.ID] = user
	r.usersByUsername[user.Username] = user
	r.usersByEmail[user.Email] = user
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
	u, ok := r.usersByEmail[email]
	if !ok {
		return false, nil
	}
	u.PasswordHash = passwordHash
	u.PasswordUpdatedAt = updatedAt
	u.UpdatedAt = updatedAt
	r.usersByID[u.ID] = u
	r.usersByUsername[u.Username] = u
	r.usersByEmail[email] = u
	return true, nil
}

func TestRegisterAndLoginAndVerify(t *testing.T) {
	repo := newFakeAuthRepo()
	svc := NewService(repo, "test-secret", "test-issuer", 2*time.Hour, true)

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
	svc := NewService(repo, "test-secret", "test-issuer", 2*time.Hour, true)

	_, err := svc.Register(context.Background(), ports.RegisterRequest{
		Username: "alice",
		Email:    "not-an-email",
		Password: "password123",
	})
	if err != ports.ErrInvalidArgument {
		t.Fatalf("expected ErrInvalidArgument, got %v", err)
	}
}

func TestPasswordResetDisabledPolicy(t *testing.T) {
	repo := newFakeAuthRepo()
	svc := NewService(repo, "test-secret", "test-issuer", 2*time.Hour, false)

	_, err := svc.SendPasswordResetCode(context.Background(), ports.PasswordResetCodeRequest{Email: "alice@example.com"})
	if err != ports.ErrCapabilityDisabled {
		t.Fatalf("expected ErrCapabilityDisabled, got %v", err)
	}
}

func TestPasswordResetRateLimitedAndInvalidOrExpiredCode(t *testing.T) {
	repo := newFakeAuthRepo()
	svc := NewService(repo, "test-secret", "test-issuer", 2*time.Hour, true)

	_, err := svc.Register(context.Background(), ports.RegisterRequest{Username: "alice", Email: "alice@example.com", Password: "password123"})
	if err != nil {
		t.Fatalf("register failed: %v", err)
	}

	if _, err := svc.SendPasswordResetCode(context.Background(), ports.PasswordResetCodeRequest{Email: "alice@example.com"}); err != nil {
		t.Fatalf("send code failed: %v", err)
	}
	if _, err := svc.SendPasswordResetCode(context.Background(), ports.PasswordResetCodeRequest{Email: "alice@example.com"}); err != ports.ErrRateLimited {
		t.Fatalf("expected ErrRateLimited, got %v", err)
	}

	if err := svc.ResetPassword(context.Background(), ports.PasswordResetRequest{Email: "alice@example.com", Code: "000000", NewPassword: "new-password"}); err != ports.ErrInvalidArgument {
		t.Fatalf("expected ErrInvalidArgument for invalid code, got %v", err)
	}

	rec := svc.resetCodes["alice@example.com"]
	rec.ExpiresAt = time.Now().UTC().Add(-time.Second)
	svc.resetCodes["alice@example.com"] = rec
	if err := svc.ResetPassword(context.Background(), ports.PasswordResetRequest{Email: "alice@example.com", Code: rec.Code, NewPassword: "new-password"}); err != ports.ErrInvalidArgument {
		t.Fatalf("expected ErrInvalidArgument for expired code, got %v", err)
	}
}

func TestPasswordResetInvalidatesIssuedToken(t *testing.T) {
	repo := newFakeAuthRepo()
	svc := NewService(repo, "test-secret", "test-issuer", 2*time.Hour, true)

	if _, err := svc.Register(context.Background(), ports.RegisterRequest{Username: "alice", Email: "alice@example.com", Password: "password123"}); err != nil {
		t.Fatalf("register failed: %v", err)
	}

	loginResp, err := svc.Login(context.Background(), ports.LoginRequest{Username: "alice", Password: "password123"})
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}
	claims, err := security.ParseToken("test-secret", loginResp.Token)
	if err != nil {
		t.Fatalf("parse token failed: %v", err)
	}

	if _, err := svc.SendPasswordResetCode(context.Background(), ports.PasswordResetCodeRequest{Email: "alice@example.com"}); err != nil {
		t.Fatalf("send code failed: %v", err)
	}
	rec := svc.resetCodes["alice@example.com"]
	if err := svc.ResetPassword(context.Background(), ports.PasswordResetRequest{Email: "alice@example.com", Code: rec.Code, NewPassword: "new-password"}); err != nil {
		t.Fatalf("reset password failed: %v", err)
	}

	if _, err := svc.Verify(security.WithClaims(context.Background(), claims)); err != ports.ErrUnauthorized {
		t.Fatalf("expected old token rejected after reset, got %v", err)
	}

	if _, err := svc.Login(context.Background(), ports.LoginRequest{Username: "alice", Password: "password123"}); err != ports.ErrUnauthorized {
		t.Fatalf("expected old password to be rejected after reset, got %v", err)
	}

	newLogin, err := svc.Login(context.Background(), ports.LoginRequest{Username: "alice", Password: "new-password"})
	if err != nil {
		t.Fatalf("login with new password failed: %v", err)
	}
	newClaims, err := security.ParseToken("test-secret", newLogin.Token)
	if err != nil {
		t.Fatalf("parse new token failed: %v", err)
	}
	if _, err := svc.Verify(security.WithClaims(context.Background(), newClaims)); err != nil {
		t.Fatalf("new token verify failed: %v", err)
	}
}
