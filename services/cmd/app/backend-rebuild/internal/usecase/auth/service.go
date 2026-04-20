// Layer: Usecase (业务逻辑层)
// Responsibility: 实现认证业务逻辑(注册/登录/校验/改密)，纯业务无外部依赖
// Dependency: 依赖 Ports(接口契约) 和 Repository接口，不依赖具体存储实现
package auth

import (
	"FeasOJ/app/backend-rebuild/internal/ports"
	"FeasOJ/app/backend-rebuild/internal/security"
	passwordutil "FeasOJ/pkg/auth"
	"context"
	"errors"
	"net/mail"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

type Service struct {
	repo                  UserRepository
	jwtSecret             string
	jwtIssuer             string
	jwtTTL                time.Duration
	passwordResetEnabled  bool
	passwordResetCodeTTL  time.Duration
	passwordResetInterval time.Duration
	mu                    sync.Mutex
	resetCodes            map[string]resetCodeRecord
	lastCodeSentAt        map[string]time.Time
}

type resetCodeRecord struct {
	Code      string
	ExpiresAt time.Time
}

func NewService(repo UserRepository, jwtSecret, jwtIssuer string, jwtTTL time.Duration, passwordResetEnabled bool) *Service {
	return &Service{
		repo:                  repo,
		jwtSecret:             jwtSecret,
		jwtIssuer:             jwtIssuer,
		jwtTTL:                jwtTTL,
		passwordResetEnabled:  passwordResetEnabled,
		passwordResetCodeTTL:  5 * time.Minute,
		passwordResetInterval: time.Minute,
		resetCodes:            map[string]resetCodeRecord{},
		lastCodeSentAt:        map[string]time.Time{},
	}
}

func (s *Service) Register(ctx context.Context, req ports.RegisterRequest) (ports.UserDTO, error) {
	if s.repo == nil {
		return ports.UserDTO{}, ports.ErrNotImplemented
	}

	username := strings.TrimSpace(req.Username)
	email := strings.TrimSpace(req.Email)
	password := strings.TrimSpace(req.Password)
	if username == "" || email == "" || password == "" {
		return ports.UserDTO{}, ports.ErrInvalidArgument
	}
	if _, err := mail.ParseAddress(email); err != nil {
		return ports.UserDTO{}, ports.ErrInvalidArgument
	}
	if len(password) < 6 {
		return ports.UserDTO{}, ports.ErrInvalidArgument
	}

	hash := passwordutil.EncryptPassword(password)
	if hash == "" {
		return ports.UserDTO{}, errors.New("failed to hash password")
	}

	now := time.Now().UTC()
	user := User{
		ID:                uuid.NewString(),
		Username:          username,
		Email:             email,
		PasswordHash:      hash,
		PasswordUpdatedAt: now,
		Role:              "student",
		Score:             0,
		Status:            "active",
		CreatedAt:         now,
		UpdatedAt:         now,
	}

	createdUser, err := s.repo.Create(ctx, user)
	if err != nil {
		return ports.UserDTO{}, err
	}

	return toUserDTO(createdUser), nil
}

func (s *Service) Login(ctx context.Context, req ports.LoginRequest) (ports.LoginResponse, error) {
	if s.repo == nil {
		return ports.LoginResponse{}, ports.ErrNotImplemented
	}

	username := strings.TrimSpace(req.Username)
	password := strings.TrimSpace(req.Password)
	if username == "" || password == "" {
		return ports.LoginResponse{}, ports.ErrInvalidArgument
	}

	user, err := s.repo.GetByUsername(ctx, username)
	if errors.Is(err, ports.ErrNotFound) {
		return ports.LoginResponse{}, ports.ErrUnauthorized
	}
	if err != nil {
		return ports.LoginResponse{}, err
	}

	if !passwordutil.VerifyPassword(password, user.PasswordHash) {
		return ports.LoginResponse{}, ports.ErrUnauthorized
	}
	if user.Status != "active" {
		return ports.LoginResponse{}, ports.ErrForbidden
	}

	token, err := security.GenerateToken(s.jwtSecret, s.jwtIssuer, s.jwtTTL, user.ID, user.Role, user.Status)
	if err != nil {
		return ports.LoginResponse{}, err
	}

	return ports.LoginResponse{
		Token: token,
		User:  toUserDTO(user),
	}, nil
}

func (s *Service) Verify(ctx context.Context) (ports.VerifyResponse, error) {
	if s.repo == nil {
		return ports.VerifyResponse{}, ports.ErrNotImplemented
	}

	claims, ok := security.ClaimsFromContext(ctx)
	if !ok || claims.UserID == "" {
		return ports.VerifyResponse{}, ports.ErrUnauthorized
	}

	user, err := s.repo.GetByID(ctx, claims.UserID)
	if errors.Is(err, ports.ErrNotFound) {
		return ports.VerifyResponse{}, ports.ErrUnauthorized
	}
	if err != nil {
		return ports.VerifyResponse{}, err
	}

	if user.Status != "active" {
		return ports.VerifyResponse{}, ports.ErrForbidden
	}
	if claims.IssuedAt == nil {
		return ports.VerifyResponse{}, ports.ErrUnauthorized
	}
	if !user.PasswordUpdatedAt.IsZero() && claims.IssuedAt.Time.Before(user.PasswordUpdatedAt) {
		return ports.VerifyResponse{}, ports.ErrUnauthorized
	}

	return ports.VerifyResponse{
		User: toUserDTO(user),
		Capabilities: ports.AuthCapabilities{
			PasswordResetEnabled: s.passwordResetEnabled,
		},
	}, nil
}

func (s *Service) SendPasswordResetCode(ctx context.Context, req ports.PasswordResetCodeRequest) (ports.PasswordResetCodeResponse, error) {
	if !s.passwordResetEnabled {
		return ports.PasswordResetCodeResponse{}, ports.ErrCapabilityDisabled
	}
	if s.repo == nil {
		return ports.PasswordResetCodeResponse{}, ports.ErrNotImplemented
	}

	email := strings.TrimSpace(req.Email)
	if _, err := mail.ParseAddress(email); err != nil {
		return ports.PasswordResetCodeResponse{}, ports.ErrInvalidArgument
	}

	now := time.Now().UTC()
	s.mu.Lock()
	defer s.mu.Unlock()
	if last, ok := s.lastCodeSentAt[email]; ok && now.Sub(last) < s.passwordResetInterval {
		return ports.PasswordResetCodeResponse{}, ports.ErrRateLimited
	}
	code := buildResetCode(now)
	s.resetCodes[email] = resetCodeRecord{Code: code, ExpiresAt: now.Add(s.passwordResetCodeTTL)}
	s.lastCodeSentAt[email] = now

	return ports.PasswordResetCodeResponse{ExpiresInSeconds: int(s.passwordResetCodeTTL.Seconds())}, nil
}

func (s *Service) ResetPassword(ctx context.Context, req ports.PasswordResetRequest) error {
	if !s.passwordResetEnabled {
		return ports.ErrCapabilityDisabled
	}
	if s.repo == nil {
		return ports.ErrNotImplemented
	}

	email := strings.TrimSpace(req.Email)
	code := strings.TrimSpace(req.Code)
	newPassword := strings.TrimSpace(req.NewPassword)
	if _, err := mail.ParseAddress(email); err != nil {
		return ports.ErrInvalidArgument
	}
	if code == "" || len(newPassword) < 6 {
		return ports.ErrInvalidArgument
	}

	now := time.Now().UTC()
	s.mu.Lock()
	rec, ok := s.resetCodes[email]
	if !ok || now.After(rec.ExpiresAt) || rec.Code != code {
		s.mu.Unlock()
		return ports.ErrInvalidArgument
	}
	delete(s.resetCodes, email)
	s.mu.Unlock()

	hash := passwordutil.EncryptPassword(newPassword)
	if hash == "" {
		return errors.New("failed to hash password")
	}
	affected, err := s.repo.UpdatePasswordByEmail(ctx, email, hash, now)
	if err != nil {
		return err
	}
	if !affected {
		return ports.ErrNotFound
	}
	return nil
}

func buildResetCode(now time.Time) string {
	v := now.UnixNano() % 1000000
	if v < 0 {
		v = -v
	}
	code := strconv.FormatInt(v, 10)
	for len(code) < 6 {
		code = "0" + code
	}
	return code
}

func toUserDTO(u User) ports.UserDTO {
	return ports.UserDTO{
		ID:       u.ID,
		Username: u.Username,
		Email:    u.Email,
		Avatar:   u.Avatar,
		Synopsis: u.Synopsis,
		Score:    u.Score,
		Role:     u.Role,
		Status:   u.Status,
	}
}
