package competitions

import (
	"FeasOJ/app/backend-rebuild/internal/ports"
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	defaultPage  = 1
	defaultLimit = 20
	maxLimit     = 100
)

type Service struct{ repo Repository }

func NewService(repo Repository) *Service { return &Service{repo: repo} }

func (s *Service) ListContests(ctx context.Context, req ports.ContestsQuery) ([]ports.ContestDTO, error) {
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
	items, err := s.repo.List(ctx, offset, limit, strings.TrimSpace(req.Visibility), strings.TrimSpace(req.RuleType), strings.TrimSpace(req.Status))
	if err != nil {
		return nil, err
	}

	resp := make([]ports.ContestDTO, 0, len(items))
	for _, item := range items {
		resp = append(resp, toContestDTO(item))
	}
	return resp, nil
}

func (s *Service) GetContest(ctx context.Context, contestID int64) (ports.ContestDTO, error) {
	if s.repo == nil {
		return ports.ContestDTO{}, ports.ErrNotImplemented
	}
	if contestID <= 0 {
		return ports.ContestDTO{}, ports.ErrInvalidArgument
	}
	c, err := s.repo.GetByID(ctx, contestID)
	if errors.Is(err, ports.ErrNotFound) {
		return ports.ContestDTO{}, ports.ErrNotFound
	}
	if err != nil {
		return ports.ContestDTO{}, err
	}
	return toContestDTO(c), nil
}

func (s *Service) JoinContest(ctx context.Context, req ports.JoinContestRequest) (ports.ContestParticipantDTO, error) {
	if s.repo == nil {
		return ports.ContestParticipantDTO{}, ports.ErrNotImplemented
	}
	if req.ContestID <= 0 || strings.TrimSpace(req.UserID) == "" {
		return ports.ContestParticipantDTO{}, ports.ErrInvalidArgument
	}
	now := time.Now().UTC()
	p, err := s.repo.CreateParticipant(ctx, Participant{
		ID:        uuid.NewString(),
		ContestID: req.ContestID,
		UserID:    strings.TrimSpace(req.UserID),
		Status:    "registered",
		JoinedAt:  &now,
		CreatedAt: now,
		UpdatedAt: now,
	})
	if err != nil {
		return ports.ContestParticipantDTO{}, err
	}
	return ports.ContestParticipantDTO{ID: p.ID, ContestID: p.ContestID, UserID: p.UserID, Status: p.Status}, nil
}

func toContestDTO(c Contest) ports.ContestDTO {
	startAt := ""
	if c.StartAt != nil {
		startAt = c.StartAt.UTC().Format(time.RFC3339)
	}
	endAt := ""
	if c.EndAt != nil {
		endAt = c.EndAt.UTC().Format(time.RFC3339)
	}
	return ports.ContestDTO{
		ID:           c.ID,
		Title:        c.Title,
		Subtitle:     c.Subtitle,
		Description:  c.Description,
		Announcement: c.Announcement,
		OwnerUserID:  c.OwnerUserID,
		ClassID:      c.ClassID,
		Visibility:   c.Visibility,
		RuleType:     c.RuleType,
		Status:       c.Status,
		IsEncrypted:  c.IsEncrypted,
		StartAt:      startAt,
		EndAt:        endAt,
	}
}
