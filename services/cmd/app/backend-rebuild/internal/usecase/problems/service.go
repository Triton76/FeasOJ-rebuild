package problems

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
	repo ProblemRepository
}

func NewService(repo ProblemRepository) *Service {
	return &Service{repo: repo}
}

func (s *Service) ListProblems(ctx context.Context, req ports.ProblemsQuery) ([]ports.ProblemDTO, error) {
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
	visibility := strings.TrimSpace(req.Visibility)
	status := strings.TrimSpace(req.Status)
	items, err := s.repo.List(ctx, offset, limit, visibility, status)
	if err != nil {
		return nil, err
	}

	resp := make([]ports.ProblemDTO, 0, len(items))
	for _, p := range items {
		resp = append(resp, toProblemDTO(p))
	}
	return resp, nil
}

func (s *Service) GetProblem(ctx context.Context, problemID int64) (ports.ProblemDTO, error) {
	if s.repo == nil {
		return ports.ProblemDTO{}, ports.ErrNotImplemented
	}
	if problemID <= 0 {
		return ports.ProblemDTO{}, ports.ErrInvalidArgument
	}

	p, err := s.repo.GetByID(ctx, problemID)
	if errors.Is(err, ports.ErrNotFound) {
		return ports.ProblemDTO{}, ports.ErrNotFound
	}
	if err != nil {
		return ports.ProblemDTO{}, err
	}

	return toProblemDTO(p), nil
}

func toProblemDTO(p Problem) ports.ProblemDTO {
	return ports.ProblemDTO{
		ID:            p.ID,
		Title:         p.Title,
		Content:       p.Content,
		Input:         p.Input,
		Output:        p.Output,
		Difficulty:    p.Difficulty,
		TimeLimitMS:   p.TimeLimitMS,
		MemoryLimitMB: p.MemoryLimitMB,
		OwnerUserID:   p.OwnerUserID,
		ClassID:       p.ClassID,
		Visibility:    p.Visibility,
		Status:        p.Status,
	}
}
