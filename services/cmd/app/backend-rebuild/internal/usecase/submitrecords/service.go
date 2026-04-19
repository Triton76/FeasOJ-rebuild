package submitrecords

import (
	"FeasOJ/app/backend-rebuild/internal/ports"
	"context"
	"strings"
	"time"
)

const (
	defaultPage  = 1
	defaultLimit = 20
	maxLimit     = 100
)

type Service struct{ repo Repository }

func NewService(repo Repository) *Service { return &Service{repo: repo} }

func (s *Service) CreateSubmission(ctx context.Context, req ports.CreateSubmissionRequest) (ports.SubmissionDTO, error) {
	if s.repo == nil {
		return ports.SubmissionDTO{}, ports.ErrNotImplemented
	}
	if req.ProblemID <= 0 || strings.TrimSpace(req.Language) == "" || strings.TrimSpace(req.SourceCode) == "" {
		return ports.SubmissionDTO{}, ports.ErrInvalidArgument
	}
	now := time.Now().UTC()
	item, err := s.repo.Create(ctx, Submission{
		UserID:      "",
		ProblemID:   req.ProblemID,
		ContestID:   req.ContestID,
		Language:    strings.TrimSpace(req.Language),
		SourceCode:  req.SourceCode,
		Result:      "pending",
		SubmittedAt: now,
		CreatedAt:   now,
		UpdatedAt:   now,
	})
	if err != nil {
		return ports.SubmissionDTO{}, err
	}
	return toDTO(item), nil
}

func (s *Service) ListSubmissions(ctx context.Context, req ports.SubmissionsQuery) ([]ports.SubmissionDTO, error) {
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

	items, err := s.repo.List(ctx, Query{UserID: strings.TrimSpace(req.UserID), ProblemID: req.ProblemID, ContestID: req.ContestID, Offset: (page - 1) * limit, Limit: limit})
	if err != nil {
		return nil, err
	}
	resp := make([]ports.SubmissionDTO, 0, len(items))
	for _, item := range items {
		resp = append(resp, toDTO(item))
	}
	return resp, nil
}

func toDTO(s Submission) ports.SubmissionDTO {
	return ports.SubmissionDTO{ID: s.ID, UserID: s.UserID, ProblemID: s.ProblemID, ContestID: s.ContestID, Language: s.Language, Result: s.Result, Score: s.Score, SubmittedAt: s.SubmittedAt.UTC().Format(time.RFC3339)}
}
