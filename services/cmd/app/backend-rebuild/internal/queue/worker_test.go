package queue

import (
	"FeasOJ/app/backend-rebuild/internal/ports"
	"context"
	"testing"
	"time"
)

type fakeSubmitServiceForWorker struct {
	called chan int64
}

func (s *fakeSubmitServiceForWorker) CreateSubmission(context.Context, ports.CreateSubmissionRequest) (ports.SubmissionDTO, error) {
	return ports.SubmissionDTO{}, ports.ErrNotImplemented
}

func (s *fakeSubmitServiceForWorker) ListSubmissions(context.Context, ports.SubmissionsQuery) ([]ports.SubmissionDTO, error) {
	return nil, ports.ErrNotImplemented
}

func (s *fakeSubmitServiceForWorker) MarkSubmissionJudging(_ context.Context, submissionID int64, _ string) (ports.SubmissionDTO, error) {
	select {
	case s.called <- submissionID:
	default:
	}
	return ports.SubmissionDTO{ID: submissionID, Result: "judging"}, nil
}

func (s *fakeSubmitServiceForWorker) WritebackSubmission(context.Context, ports.JudgeWritebackRequest) (ports.SubmissionDTO, error) {
	return ports.SubmissionDTO{}, ports.ErrNotImplemented
}

func TestJudgeWorkerConsumesAndMarksJudging(t *testing.T) {
	q := NewMemorySubmissionQueue()
	svc := &fakeSubmitServiceForWorker{called: make(chan int64, 1)}

	if err := q.Enqueue(context.Background(), ports.SubmissionJob{SubmissionID: 101}); err != nil {
		t.Fatalf("enqueue failed: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	worker := NewJudgeWorker(q, svc, 5*time.Millisecond)
	worker.Start(ctx)

	select {
	case id := <-svc.called:
		if id != 101 {
			t.Fatalf("expected submission id 101, got %d", id)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("worker did not consume and mark submission in time")
	}
}
