// Layer: Usecase (业务逻辑层)
// Responsibility: 实现用户业务逻辑(获取资料/更新资料/排行榜)，纯业务无外部依赖
// Dependency: 依赖 Ports(接口契约) 和 Repository接口，不依赖具体存储实现
package users

import (
	"FeasOJ/app/backend-rebuild/internal/ports"
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
)

const (
	defaultPage                 = 1
	defaultLimit                = 20
	maxLimit                    = 100
	defaultAvatarUploadMaxBytes = 2 * 1024 * 1024
)

type Service struct {
	repo                 UserRepository
	avatarStorage        AvatarStorage
	avatarUploadEnabled  bool
	avatarUploadMaxBytes int64
}

type Options struct {
	AvatarStorage        AvatarStorage
	AvatarUploadEnabled  bool
	AvatarUploadMaxBytes int64
}

var allowedAvatarContentTypes = map[string]string{
	"image/png":  ".png",
	"image/jpeg": ".jpg",
	"image/gif":  ".gif",
	"image/webp": ".webp",
}

func NewService(repo UserRepository, opts Options) *Service {
	maxBytes := opts.AvatarUploadMaxBytes
	if maxBytes <= 0 {
		maxBytes = defaultAvatarUploadMaxBytes
	}
	return &Service{
		repo:                 repo,
		avatarStorage:        opts.AvatarStorage,
		avatarUploadEnabled:  opts.AvatarUploadEnabled,
		avatarUploadMaxBytes: maxBytes,
	}
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

	var avatar *string
	if req.Avatar != nil {
		value := strings.TrimSpace(*req.Avatar)
		avatar = &value
	}

	var synopsis *string
	if req.Synopsis != nil {
		value := strings.TrimSpace(*req.Synopsis)
		synopsis = &value
	}

	if avatar == nil && synopsis == nil {
		return s.GetProfile(ctx, userID)
	}

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

func (s *Service) UploadAvatar(ctx context.Context, userID string, req ports.UploadAvatarRequest) (ports.AvatarUploadResponse, error) {
	if s.repo == nil {
		return ports.AvatarUploadResponse{}, ports.ErrNotImplemented
	}
	if !s.avatarUploadEnabled {
		return ports.AvatarUploadResponse{}, fmt.Errorf("avatar upload is disabled: %w", ports.ErrCapabilityDisabled)
	}
	if s.avatarStorage == nil {
		return ports.AvatarUploadResponse{}, ports.ErrNotImplemented
	}

	userID = strings.TrimSpace(userID)
	if userID == "" {
		return ports.AvatarUploadResponse{}, ports.ErrInvalidArgument
	}
	if _, err := s.repo.GetByID(ctx, userID); err != nil {
		return ports.AvatarUploadResponse{}, err
	}

	size := int64(len(req.Content))
	if size == 0 {
		return ports.AvatarUploadResponse{}, fmt.Errorf("avatar file is empty: %w", ports.ErrInvalidArgument)
	}
	if size > s.avatarUploadMaxBytes {
		return ports.AvatarUploadResponse{}, fmt.Errorf("avatar file exceeds max size %d bytes: %w", s.avatarUploadMaxBytes, ports.ErrInvalidArgument)
	}

	contentType := http.DetectContentType(req.Content)
	if idx := strings.Index(contentType, ";"); idx >= 0 {
		contentType = strings.TrimSpace(contentType[:idx])
	}
	ext, ok := allowedAvatarContentTypes[contentType]
	if !ok {
		return ports.AvatarUploadResponse{}, fmt.Errorf("avatar file type %q is not allowed: %w", contentType, ports.ErrInvalidArgument)
	}

	fileName, err := buildAvatarFilename(userID, ext)
	if err != nil {
		return ports.AvatarUploadResponse{}, err
	}

	storedName, err := s.avatarStorage.Save(ctx, fileName, req.Content)
	if err != nil {
		return ports.AvatarUploadResponse{}, err
	}

	return ports.AvatarUploadResponse{
		Avatar:      storedName,
		ContentType: contentType,
		SizeBytes:   size,
	}, nil
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

func buildAvatarFilename(userID, ext string) (string, error) {
	safeUserID := strings.Map(func(r rune) rune {
		switch {
		case r >= 'a' && r <= 'z':
			return r
		case r >= 'A' && r <= 'Z':
			return r
		case r >= '0' && r <= '9':
			return r
		case r == '-' || r == '_':
			return r
		default:
			return '-'
		}
	}, userID)
	if safeUserID == "" {
		safeUserID = "user"
	}

	buf := make([]byte, 6)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return fmt.Sprintf("%s-%d-%s%s", safeUserID, time.Now().UTC().UnixNano(), hex.EncodeToString(buf), ext), nil
}
