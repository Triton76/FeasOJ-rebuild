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

type Service struct {
	repo  Repository
	queue ports.SubmissionQueue
}

func NewService(repo Repository, queue ports.SubmissionQueue) *Service {
	if queue == nil {
		// 如果没有提供队列，创建一个空的 noop 队列
		queue = &NoOpQueue{}
	}
	return &Service{repo: repo, queue: queue}
}

// NoOpQueue 是一个不进行任何操作的队列实现，用于回退
type NoOpQueue struct{}

func (q *NoOpQueue) Enqueue(ctx context.Context, job ports.SubmissionJob) error {
	return nil
}

func (q *NoOpQueue) Dequeue(ctx context.Context) (*ports.SubmissionJob, error) {
	return nil, ports.ErrQueueEmpty
}

func (q *NoOpQueue) Acknowledge(ctx context.Context, submissionID int64) error {
	return nil
}

func (q *NoOpQueue) Size(ctx context.Context) (int, error) {
	return 0, nil
}

func (s *Service) CreateSubmission(ctx context.Context, req ports.CreateSubmissionRequest) (ports.SubmissionDTO, error) {
	if s.repo == nil {
		return ports.SubmissionDTO{}, ports.ErrNotImplemented
	}
	if req.ProblemID <= 0 || strings.TrimSpace(req.Language) == "" || strings.TrimSpace(req.SourceCode) == "" || strings.TrimSpace(req.UserID) == "" {
		return ports.SubmissionDTO{}, ports.ErrInvalidArgument
	}
	now := time.Now().UTC()
	// 第一步：先将提交持久化到数据库，状态为 pending
	item, err := s.repo.Create(ctx, Submission{
		UserID:      strings.TrimSpace(req.UserID),
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

	// 第二步：将提交入队，准备异步判题
	// 幂等键为 submission_id，重复入队时直接返回成功
	job := ports.SubmissionJob{
		SubmissionID: item.ID,
		UserID:       item.UserID,
		ProblemID:    item.ProblemID,
		ContestID:    item.ContestID,
		Language:     item.Language,
		SourceCode:   item.SourceCode,
	}
	if err := s.queue.Enqueue(ctx, job); err != nil {
		// 入队失败不影响提交本身的返回，但会记录日志（TODO 增加日志)
		// 至少一次语义：异步队列消费者会持续重试
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
