// Layer: Usecase (业务逻辑层)
// Responsibility: 实现用户业务逻辑(获取资料/更新资料/排行榜)，纯业务无外部依赖
// Dependency: 依赖 Ports(接口契约) 和 Repository接口，不依赖具体存储实现
package users

import (
	"FeasOJ/app/backend-rebuild/internal/ports"
	"context"
	"errors"
	"strings"
	"time"
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

func (s *Service) GetProfile(ctx context.Context, userID string) (ports.UserDTO, error) {
	if s.repo == nil {
		return ports.UserDTO{}, ports.ErrNotImplemented
	}
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return ports.UserDTO{}, ports.ErrInvalidArgument
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

func (s *Service) UpdateProfile(ctx context.Context, userID string, req ports.UpdateProfileRequest) (ports.UserDTO, error) {
	if s.repo == nil {
		return ports.UserDTO{}, ports.ErrNotImplemented
	}
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return ports.UserDTO{}, ports.ErrInvalidArgument
	}

	avatar := strings.TrimSpace(req.Avatar)
	synopsis := strings.TrimSpace(req.Synopsis)
	affected, err := s.repo.UpdateProfile(ctx, userID, avatar, synopsis, time.Now().UTC())
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

func (s *Service) ListRanking(ctx context.Context, req ports.RankingQuery) ([]ports.RankingItem, error) {
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
	users, err := s.repo.ListRanking(ctx, offset, limit)
	if err != nil {
		return nil, err
	}

	resp := make([]ports.RankingItem, 0, len(users))
	for i, u := range users {
		resp = append(resp, ports.RankingItem{
			Rank:    offset + i + 1,
			UserDTO: toUserDTO(u),
		})
	}

	return resp, nil
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
