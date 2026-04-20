package admin

import (
	"FeasOJ/app/backend-rebuild/internal/ports"
	"context"
	"errors"
	"strings"
)

const (
	defaultPage  = 1
	defaultLimit = 20
	maxLimit     = 100
)

type Service struct {
	repo UserRepository
}

func NewService(repo UserRepository) *Service {
	return &Service{repo: repo}
}

func (s *Service) ListUsers(ctx context.Context, req ports.AdminUsersQuery) ([]ports.UserDTO, error) {
	if s.repo == nil {
		return nil, ports.ErrNotImplemented
	}

	page := req.Page
	if page <= 0 {
		page = defaultPage
	}
	limit := req.Limit
	if limit <= 0 {
		limit = defaultLimit
	}
	if limit > maxLimit {
		limit = maxLimit
	}

	offset := (page - 1) * limit
	items, err := s.repo.ListUsers(ctx, offset, limit)
	if err != nil {
		return nil, err
	}

	resp := make([]ports.UserDTO, 0, len(items))
	for _, item := range items {
		resp = append(resp, toUserDTO(item))
	}
	return resp, nil
}

func (s *Service) UpdateUserStatus(ctx context.Context, req ports.UpdateUserStatusRequest) (ports.UserDTO, error) {
	if s.repo == nil {
		return ports.UserDTO{}, ports.ErrNotImplemented
	}

	userID := strings.TrimSpace(req.UserID)
	status := strings.TrimSpace(req.Status)
	if userID == "" {
		return ports.UserDTO{}, ports.ErrInvalidArgument
	}
	if status != "active" && status != "banned" {
		return ports.UserDTO{}, ports.ErrInvalidArgument
	}

	affected, err := s.repo.UpdateUserStatus(ctx, userID, status)
	if err != nil {
		return ports.UserDTO{}, err
	}
	if !affected {
		return ports.UserDTO{}, ports.ErrNotFound
	}

	u, err := s.repo.GetByID(ctx, userID)
	if errors.Is(err, ports.ErrNotFound) {
		return ports.UserDTO{}, ports.ErrNotFound
	}
	if err != nil {
		return ports.UserDTO{}, err
	}

	return toUserDTO(u), nil
}

func (s *Service) UpdateUserRole(ctx context.Context, req ports.UpdateUserRoleRequest) (ports.UserDTO, error) {
	if s.repo == nil {
		return ports.UserDTO{}, ports.ErrNotImplemented
	}

	userID := strings.TrimSpace(req.UserID)
	role := strings.TrimSpace(req.Role)
	if userID == "" {
		return ports.UserDTO{}, ports.ErrInvalidArgument
	}
	if role != "student" && role != "teacher" && role != "admin" {
		return ports.UserDTO{}, ports.ErrInvalidArgument
	}

	affected, err := s.repo.UpdateUserRole(ctx, userID, role)
	if err != nil {
		return ports.UserDTO{}, err
	}
	if !affected {
		return ports.UserDTO{}, ports.ErrNotFound
	}

	u, err := s.repo.GetByID(ctx, userID)
	if errors.Is(err, ports.ErrNotFound) {
		return ports.UserDTO{}, ports.ErrNotFound
	}
	if err != nil {
		return ports.UserDTO{}, err
	}

	return toUserDTO(u), nil
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
