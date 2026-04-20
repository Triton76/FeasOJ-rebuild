package testcases

import (
	"FeasOJ/app/backend-rebuild/internal/ports"
	"context"
	"strings"

	"github.com/google/uuid"
)

const (
	maxTestcaseContentSize = 10 * 1024 * 1024
	maxTestcasesPerProblem = 1000
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) CreateTestcase(ctx context.Context, req ports.CreateTestcaseRequest) (ports.TestcaseDTO, error) {
	if s.repo == nil {
		return ports.TestcaseDTO{}, ports.ErrNotImplemented
	}
	if req.ProblemID <= 0 {
		return ports.TestcaseDTO{}, ports.ErrInvalidArgument
	}
	if err := validateManageAuth(req.ActorRole, req.ActorUserID); err != nil {
		return ports.TestcaseDTO{}, err
	}
	if err := validatePayload(req.InputData, req.OutputData); err != nil {
		return ports.TestcaseDTO{}, err
	}
	if err := s.ensureManageScope(ctx, req.ProblemID, req.ActorRole, req.ActorUserID); err != nil {
		return ports.TestcaseDTO{}, err
	}

	items, err := s.repo.ListByProblem(ctx, req.ProblemID)
	if err != nil {
		return ports.TestcaseDTO{}, err
	}
	if len(items) >= maxTestcasesPerProblem {
		return ports.TestcaseDTO{}, ports.ErrInvalidArgument
	}

	item, err := s.repo.Create(ctx, Testcase{
		ID:         uuid.NewString(),
		ProblemID:  req.ProblemID,
		InputData:  strings.TrimSpace(req.InputData),
		OutputData: strings.TrimSpace(req.OutputData),
		IsSample:   req.IsSample,
		SortOrder:  len(items) + 1,
	})
	if err != nil {
		return ports.TestcaseDTO{}, err
	}
	return toDTO(item), nil
}

func (s *Service) ListTestcases(ctx context.Context, req ports.ListTestcasesRequest) ([]ports.TestcaseDTO, error) {
	if s.repo == nil {
		return nil, ports.ErrNotImplemented
	}
	if req.ProblemID <= 0 {
		return nil, ports.ErrInvalidArgument
	}
	if err := validateManageAuth(req.ActorRole, req.ActorUserID); err != nil {
		return nil, err
	}
	if err := s.ensureManageScope(ctx, req.ProblemID, req.ActorRole, req.ActorUserID); err != nil {
		return nil, err
	}

	items, err := s.repo.ListByProblem(ctx, req.ProblemID)
	if err != nil {
		return nil, err
	}
	resp := make([]ports.TestcaseDTO, 0, len(items))
	for _, item := range items {
		resp = append(resp, toDTO(item))
	}
	return resp, nil
}

func (s *Service) UpdateTestcase(ctx context.Context, req ports.UpdateTestcaseRequest) (ports.TestcaseDTO, error) {
	if s.repo == nil {
		return ports.TestcaseDTO{}, ports.ErrNotImplemented
	}
	if req.ProblemID <= 0 || strings.TrimSpace(req.TestcaseID) == "" {
		return ports.TestcaseDTO{}, ports.ErrInvalidArgument
	}
	if err := validateManageAuth(req.ActorRole, req.ActorUserID); err != nil {
		return ports.TestcaseDTO{}, err
	}
	if err := validatePayload(req.InputData, req.OutputData); err != nil {
		return ports.TestcaseDTO{}, err
	}
	if err := s.ensureManageScope(ctx, req.ProblemID, req.ActorRole, req.ActorUserID); err != nil {
		return ports.TestcaseDTO{}, err
	}

	item, err := s.repo.GetByID(ctx, req.ProblemID, strings.TrimSpace(req.TestcaseID))
	if err != nil {
		return ports.TestcaseDTO{}, err
	}
	item.InputData = strings.TrimSpace(req.InputData)
	item.OutputData = strings.TrimSpace(req.OutputData)
	item.IsSample = req.IsSample

	updated, err := s.repo.Update(ctx, item)
	if err != nil {
		return ports.TestcaseDTO{}, err
	}
	return toDTO(updated), nil
}

func (s *Service) DeleteTestcase(ctx context.Context, req ports.DeleteTestcaseRequest) error {
	if s.repo == nil {
		return ports.ErrNotImplemented
	}
	if req.ProblemID <= 0 || strings.TrimSpace(req.TestcaseID) == "" {
		return ports.ErrInvalidArgument
	}
	if err := validateManageAuth(req.ActorRole, req.ActorUserID); err != nil {
		return err
	}
	if err := s.ensureManageScope(ctx, req.ProblemID, req.ActorRole, req.ActorUserID); err != nil {
		return err
	}

	return s.repo.Delete(ctx, req.ProblemID, strings.TrimSpace(req.TestcaseID))
}

func (s *Service) ReorderTestcases(ctx context.Context, req ports.ReorderTestcasesRequest) ([]ports.TestcaseDTO, error) {
	if s.repo == nil {
		return nil, ports.ErrNotImplemented
	}
	if req.ProblemID <= 0 || len(req.TestcaseIDs) == 0 {
		return nil, ports.ErrInvalidArgument
	}
	if err := validateManageAuth(req.ActorRole, req.ActorUserID); err != nil {
		return nil, err
	}
	if err := s.ensureManageScope(ctx, req.ProblemID, req.ActorRole, req.ActorUserID); err != nil {
		return nil, err
	}

	seen := make(map[string]struct{}, len(req.TestcaseIDs))
	ordered := make([]string, 0, len(req.TestcaseIDs))
	for _, rawID := range req.TestcaseIDs {
		id := strings.TrimSpace(rawID)
		if id == "" {
			return nil, ports.ErrInvalidArgument
		}
		if _, ok := seen[id]; ok {
			return nil, ports.ErrInvalidArgument
		}
		seen[id] = struct{}{}
		ordered = append(ordered, id)
	}

	items, err := s.repo.ListByProblem(ctx, req.ProblemID)
	if err != nil {
		return nil, err
	}
	if len(items) != len(ordered) {
		return nil, ports.ErrInvalidArgument
	}
	for _, item := range items {
		if _, ok := seen[item.ID]; !ok {
			return nil, ports.ErrInvalidArgument
		}
	}

	reordered, err := s.repo.Reorder(ctx, req.ProblemID, ordered)
	if err != nil {
		return nil, err
	}
	resp := make([]ports.TestcaseDTO, 0, len(reordered))
	for _, item := range reordered {
		resp = append(resp, toDTO(item))
	}
	return resp, nil
}

func (s *Service) ListTestcasesForJudge(ctx context.Context, req ports.JudgeListTestcasesRequest) ([]ports.TestcaseDTO, error) {
	if s.repo == nil {
		return nil, ports.ErrNotImplemented
	}
	if req.ProblemID <= 0 {
		return nil, ports.ErrInvalidArgument
	}
	items, err := s.repo.ListByProblem(ctx, req.ProblemID)
	if err != nil {
		return nil, err
	}
	resp := make([]ports.TestcaseDTO, 0, len(items))
	for _, item := range items {
		resp = append(resp, toDTO(item))
	}
	return resp, nil
}

func validateManageAuth(role, userID string) error {
	role = strings.TrimSpace(role)
	if role != "teacher" && role != "admin" {
		return ports.ErrForbidden
	}
	if strings.TrimSpace(userID) == "" {
		return ports.ErrUnauthorized
	}
	return nil
}

func validatePayload(inputData, outputData string) error {
	inputData = strings.TrimSpace(inputData)
	outputData = strings.TrimSpace(outputData)
	if inputData == "" || outputData == "" {
		return ports.ErrInvalidArgument
	}
	if len(inputData) > maxTestcaseContentSize || len(outputData) > maxTestcaseContentSize {
		return ports.ErrInvalidArgument
	}
	return nil
}

func (s *Service) ensureManageScope(ctx context.Context, problemID int64, actorRole, actorUserID string) error {
	if strings.TrimSpace(actorRole) == "admin" {
		return nil
	}
	ownerUserID, err := s.repo.GetProblemOwner(ctx, problemID)
	if err != nil {
		return err
	}
	if strings.TrimSpace(ownerUserID) != strings.TrimSpace(actorUserID) {
		return ports.ErrForbidden
	}
	return nil
}

func toDTO(item Testcase) ports.TestcaseDTO {
	return ports.TestcaseDTO{
		ID:         item.ID,
		ProblemID:  item.ProblemID,
		InputData:  item.InputData,
		OutputData: item.OutputData,
		IsSample:   item.IsSample,
		SortOrder:  item.SortOrder,
	}
}
