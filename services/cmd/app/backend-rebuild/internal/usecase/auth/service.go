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
	"strings"
	"time"

	"github.com/google/uuid"
)

type Service struct {
	repo      UserRepository
	jwtSecret string
	jwtIssuer string
	jwtTTL    time.Duration
}

func NewService(repo UserRepository, jwtSecret, jwtIssuer string, jwtTTL time.Duration) *Service {
	return &Service{
		repo:      repo,
		jwtSecret: jwtSecret,
		jwtIssuer: jwtIssuer,
		jwtTTL:    jwtTTL,
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
		ID:           uuid.NewString(),
		Username:     username,
		Email:        email,
		PasswordHash: hash,
		Role:         "student",
		Score:        0,
		Status:       "active",
		CreatedAt:    now,
		UpdatedAt:    now,
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

	return ports.VerifyResponse{User: toUserDTO(user)}, nil
}

func (s *Service) ResetPassword(ctx context.Context, req ports.ResetPasswordRequest) error {
	if s.repo == nil {
		return ports.ErrNotImplemented
	}

	email := strings.TrimSpace(req.Email)
	password := strings.TrimSpace(req.Password)
	rePassword := strings.TrimSpace(req.RePassword)
	if email == "" || req.Captcha == "" || password == "" || rePassword == "" {
		return ports.ErrInvalidArgument
	}
	if password != rePassword || len(password) < 6 {
		return ports.ErrInvalidArgument
	}

	hash := passwordutil.EncryptPassword(password)
	if hash == "" {
		return errors.New("failed to hash password")
	}

	affected, err := s.repo.UpdatePasswordByEmail(ctx, email, hash, time.Now().UTC())
	if err != nil {
		return err
	}
	if !affected {
		return ports.ErrNotFound
	}

	return nil
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
