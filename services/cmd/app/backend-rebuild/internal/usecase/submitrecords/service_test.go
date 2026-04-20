package submitrecords

import (
	"FeasOJ/app/backend-rebuild/internal/ports"
	"FeasOJ/app/backend-rebuild/internal/queue"
	"context"
	"errors"
	"testing"
)

type fakeSubmitRepo struct {
	nextID int64
	items  []Submission
}

func newFakeSubmitRepo() *fakeSubmitRepo {
	return &fakeSubmitRepo{nextID: 1, items: make([]Submission, 0)}
}

func (r *fakeSubmitRepo) Create(ctx context.Context, s Submission) (Submission, error) {
	s.ID = r.nextID
	r.nextID++
	r.items = append(r.items, s)
	return s, nil
}

func (r *fakeSubmitRepo) List(ctx context.Context, req Query) ([]Submission, error) {
	return r.items, nil
}

func (r *fakeSubmitRepo) GetByID(ctx context.Context, submissionID int64) (Submission, error) {
	for _, item := range r.items {
		if item.ID == submissionID {
			return item, nil
		}
	}
	return Submission{}, ports.ErrNotFound
}

func (r *fakeSubmitRepo) UpdateJudgeResult(ctx context.Context, submissionID int64, result string, score *int) (Submission, error) {
	for i := range r.items {
		if r.items[i].ID == submissionID {
			r.items[i].Result = result
			if score != nil {
				r.items[i].Score = *score
			}
			return r.items[i], nil
		}
	}
	return Submission{}, ports.ErrNotFound
}

func TestCreateSubmissionPersistsAndEnqueues(t *testing.T) {
	repo := newFakeSubmitRepo()
	q := queue.NewMemorySubmissionQueue()
	svc := NewService(repo, q)

	resp, err := svc.CreateSubmission(context.Background(), ports.CreateSubmissionRequest{
		UserID:     "u-1",
		ProblemID:  1001,
		ContestID:  2002,
		Language:   "cpp",
		SourceCode: "int main() { return 0; }",
	})
	if err != nil {
		t.Fatalf("CreateSubmission failed: %v", err)
	}
	if resp.ID == 0 {
		t.Fatalf("expected non-zero submission id")
	}
	if resp.UserID != "u-1" {
		t.Fatalf("unexpected user id: %s", resp.UserID)
	}

	size, err := q.Size(context.Background())
	if err != nil {
		t.Fatalf("queue size failed: %v", err)
	}
	if size != 1 {
		t.Fatalf("expected queue size 1, got %d", size)
	}
}

func TestMemoryQueueIdempotentBySubmissionID(t *testing.T) {
	q := queue.NewMemorySubmissionQueue()
	ctx := context.Background()
	job := ports.SubmissionJob{
		SubmissionID: 42,
		UserID:       "u-1",
		ProblemID:    1001,
		ContestID:    0,
		Language:     "cpp",
		SourceCode:   "int main(){}",
	}

	if err := q.Enqueue(ctx, job); err != nil {
		t.Fatalf("first enqueue failed: %v", err)
	}
	if err := q.Enqueue(ctx, job); err != nil {
		t.Fatalf("second enqueue should be idempotent, got: %v", err)
	}

	size, err := q.Size(ctx)
	if err != nil {
		t.Fatalf("queue size failed: %v", err)
	}
	if size != 1 {
		t.Fatalf("expected queue size 1 after duplicate enqueue, got %d", size)
	}

	deq, err := q.Dequeue(ctx)
	if err != nil {
		t.Fatalf("dequeue failed: %v", err)
	}
	if deq.SubmissionID != 42 {
		t.Fatalf("unexpected submission id: %d", deq.SubmissionID)
	}

	if err := q.Acknowledge(ctx, 42); err != nil {
		t.Fatalf("acknowledge failed: %v", err)
	}

	// 已 ack 后可再次入队
	if err := q.Enqueue(ctx, job); err != nil {
		t.Fatalf("enqueue after ack failed: %v", err)
	}
	size2, err := q.Size(ctx)
	if err != nil {
		t.Fatalf("queue size after re-enqueue failed: %v", err)
	}
	if size2 != 1 {
		t.Fatalf("expected queue size 1 after re-enqueue, got %d", size2)
	}

	_, err = q.Dequeue(ctx)
	if err != nil {
		t.Fatalf("final dequeue failed: %v", err)
	}
	_, err = q.Dequeue(ctx)
	if err != ports.ErrQueueEmpty {
		t.Fatalf("expected ErrQueueEmpty, got %v", err)
	}

}

func TestJudgeWritebackIdempotentBySubmissionID(t *testing.T) {
	repo := newFakeSubmitRepo()
	svc := NewService(repo, queue.NewMemorySubmissionQueue())

	created, err := svc.CreateSubmission(context.Background(), ports.CreateSubmissionRequest{
		UserID:     "u-1",
		ProblemID:  1001,
		Language:   "cpp",
		SourceCode: "int main(){return 0;}",
	})
	if err != nil {
		t.Fatalf("create submission failed: %v", err)
	}

	if _, err := svc.MarkSubmissionJudging(context.Background(), created.ID, "test"); err != nil {
		t.Fatalf("mark judging failed: %v", err)
	}

	score := 100
	first, err := svc.WritebackSubmission(context.Background(), ports.JudgeWritebackRequest{
		SubmissionID: created.ID,
		Result:       SubmissionResultAccepted,
		Score:        &score,
		Source:       "judgecore",
	})
	if err != nil {
		t.Fatalf("first writeback failed: %v", err)
	}
	if first.Result != SubmissionResultAccepted {
		t.Fatalf("expected accepted, got %s", first.Result)
	}

	second, err := svc.WritebackSubmission(context.Background(), ports.JudgeWritebackRequest{
		SubmissionID: created.ID,
		Result:       SubmissionResultAccepted,
		Score:        &score,
		Source:       "judgecore",
	})
	if err != nil {
		t.Fatalf("second writeback should be idempotent, got: %v", err)
	}
	if second.Result != SubmissionResultAccepted {
		t.Fatalf("expected accepted after duplicate writeback, got %s", second.Result)
	}
}

func TestJudgeWritebackRejectsTerminalRollback(t *testing.T) {
	repo := newFakeSubmitRepo()
	svc := NewService(repo, queue.NewMemorySubmissionQueue())

	created, err := svc.CreateSubmission(context.Background(), ports.CreateSubmissionRequest{
		UserID:     "u-1",
		ProblemID:  1001,
		Language:   "cpp",
		SourceCode: "int main(){return 0;}",
	})
	if err != nil {
		t.Fatalf("create submission failed: %v", err)
	}

	if _, err := svc.MarkSubmissionJudging(context.Background(), created.ID, "test"); err != nil {
		t.Fatalf("mark judging failed: %v", err)
	}

	if _, err := svc.WritebackSubmission(context.Background(), ports.JudgeWritebackRequest{
		SubmissionID: created.ID,
		Result:       SubmissionResultAccepted,
		Source:       "judgecore",
	}); err != nil {
		t.Fatalf("initial writeback failed: %v", err)
	}

	_, err = svc.WritebackSubmission(context.Background(), ports.JudgeWritebackRequest{
		SubmissionID: created.ID,
		Result:       SubmissionResultWrongAnswer,
		Source:       "judgecore",
	})
	if !errors.Is(err, ports.ErrConflict) {
		t.Fatalf("expected ErrConflict for terminal rollback, got %v", err)
	}
}
