package problems

import (
	"FeasOJ/app/backend-rebuild/internal/ports"
	"FeasOJ/app/backend-rebuild/internal/security"
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
	items, err := s.repo.ListVisible(ctx, offset, limit, visibility, status, strings.TrimSpace(req.ActorUserID), strings.TrimSpace(req.ActorRole))
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

	claims, ok := security.ClaimsFromContext(ctx)
	actorUserID := ""
	actorRole := ""
	if ok {
		actorUserID = claims.UserID
		actorRole = claims.Role
	}

	p, err := s.repo.GetVisibleByID(ctx, problemID, actorUserID, actorRole)
	if errors.Is(err, ports.ErrNotFound) {
		return ports.ProblemDTO{}, ports.ErrNotFound
	}
	if err != nil {
		return ports.ProblemDTO{}, err
	}

	return toProblemDTO(p), nil
}

func (s *Service) GetProblemForJudge(ctx context.Context, problemID int64) (ports.ProblemDTO, error) {
	if s.repo == nil {
		return ports.ProblemDTO{}, ports.ErrNotImplemented
	}
	if problemID <= 0 {
		return ports.ProblemDTO{}, ports.ErrInvalidArgument
	}

	p, err := s.repo.GetByID(ctx, problemID)
	if err != nil {
		return ports.ProblemDTO{}, err
	}
	return toProblemDTO(p), nil
}

func (s *Service) CreateProblem(ctx context.Context, req ports.CreateProblemRequest) (ports.ProblemDTO, error) {
	if s.repo == nil {
		return ports.ProblemDTO{}, ports.ErrNotImplemented
	}
	if !canManageProblem(req.ActorRole) {
		return ports.ProblemDTO{}, ports.ErrForbidden
	}

	title := strings.TrimSpace(req.Title)
	content := strings.TrimSpace(req.Content)
	if title == "" || content == "" || req.TimeLimitMS <= 0 || req.MemoryLimitMB <= 0 {
		return ports.ProblemDTO{}, ports.ErrInvalidArgument
	}
	visibility := normalizeProblemVisibility(req.Visibility)
	status := normalizeProblemStatus(req.Status)
	if visibility == "" || status == "" {
		return ports.ProblemDTO{}, ports.ErrInvalidArgument
	}
	now := time.Now().UTC()

	created, err := s.repo.Create(ctx, Problem{
		Title:         title,
		Content:       content,
		Input:         req.Input,
		Output:        req.Output,
		Difficulty:    req.Difficulty,
		TimeLimitMS:   req.TimeLimitMS,
		MemoryLimitMB: req.MemoryLimitMB,
		OwnerUserID:   strings.TrimSpace(req.ActorUserID),
		ClassID:       strings.TrimSpace(req.ClassID),
		Visibility:    visibility,
		Status:        status,
		CreatedAt:     now,
		UpdatedAt:     now,
	})
	if err != nil {
		return ports.ProblemDTO{}, err
	}

	return toProblemDTO(created), nil
}

func (s *Service) UpdateProblem(ctx context.Context, req ports.UpdateProblemRequest) (ports.ProblemDTO, error) {
	if s.repo == nil {
		return ports.ProblemDTO{}, ports.ErrNotImplemented
	}
	if req.ProblemID <= 0 {
		return ports.ProblemDTO{}, ports.ErrInvalidArgument
	}
	if !canManageProblem(req.ActorRole) {
		return ports.ProblemDTO{}, ports.ErrForbidden
	}

	existing, err := s.repo.GetByID(ctx, req.ProblemID)
	if errors.Is(err, ports.ErrNotFound) {
		return ports.ProblemDTO{}, ports.ErrNotFound
	}
	if err != nil {
		return ports.ProblemDTO{}, err
	}
	if req.ActorRole != "admin" && existing.OwnerUserID != strings.TrimSpace(req.ActorUserID) {
		return ports.ProblemDTO{}, ports.ErrForbidden
	}

	title := strings.TrimSpace(req.Title)
	content := strings.TrimSpace(req.Content)
	if title == "" || content == "" || req.TimeLimitMS <= 0 || req.MemoryLimitMB <= 0 {
		return ports.ProblemDTO{}, ports.ErrInvalidArgument
	}
	visibility := normalizeProblemVisibility(req.Visibility)
	status := normalizeProblemStatus(req.Status)
	if visibility == "" || status == "" {
		return ports.ProblemDTO{}, ports.ErrInvalidArgument
	}

	existing.Title = title
	existing.Content = content
	existing.Input = req.Input
	existing.Output = req.Output
	existing.Difficulty = req.Difficulty
	existing.TimeLimitMS = req.TimeLimitMS
	existing.MemoryLimitMB = req.MemoryLimitMB
	existing.ClassID = strings.TrimSpace(req.ClassID)
	existing.Visibility = visibility
	existing.Status = status
	existing.UpdatedAt = time.Now().UTC()

	updated, err := s.repo.Update(ctx, existing)
	if err != nil {
		return ports.ProblemDTO{}, err
	}
	return toProblemDTO(updated), nil
}

func (s *Service) DeleteProblem(ctx context.Context, req ports.DeleteProblemRequest) error {
	if s.repo == nil {
		return ports.ErrNotImplemented
	}
	if req.ProblemID <= 0 {
		return ports.ErrInvalidArgument
	}
	if !canManageProblem(req.ActorRole) {
		return ports.ErrForbidden
	}

	p, err := s.repo.GetByID(ctx, req.ProblemID)
	if errors.Is(err, ports.ErrNotFound) {
		return ports.ErrNotFound
	}
	if err != nil {
		return err
	}
	if req.ActorRole != "admin" && p.OwnerUserID != strings.TrimSpace(req.ActorUserID) {
		return ports.ErrForbidden
	}

	if err := s.repo.Delete(ctx, req.ProblemID); err != nil {
		return err
	}
	return nil
}

func canManageProblem(role string) bool {
	role = strings.TrimSpace(role)
	return role == "teacher" || role == "admin"
}

func normalizeProblemVisibility(value string) string {
	value = strings.TrimSpace(value)
	switch value {
	case "public", "class", "private":
		return value
	default:
		return ""
	}
}

func normalizeProblemStatus(value string) string {
	value = strings.TrimSpace(value)
	switch value {
	case "draft", "published", "archived":
		return value
	default:
		return ""
	}
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
