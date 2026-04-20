package queue

import (
	"FeasOJ/app/backend-rebuild/internal/observability"
	"FeasOJ/app/backend-rebuild/internal/ports"
	"context"
	"errors"
	"time"
)

const defaultWorkerPollInterval = 300 * time.Millisecond

type JudgeWorker struct {
	queue        ports.SubmissionQueue
	submitSvc    ports.SubmitRecordsService
	pollInterval time.Duration
}

func NewJudgeWorker(queue ports.SubmissionQueue, submitSvc ports.SubmitRecordsService, pollInterval time.Duration) *JudgeWorker {
	if pollInterval <= 0 {
		pollInterval = defaultWorkerPollInterval
	}
	return &JudgeWorker{queue: queue, submitSvc: submitSvc, pollInterval: pollInterval}
}

func (w *JudgeWorker) Start(ctx context.Context) {
	if w == nil || w.queue == nil || w.submitSvc == nil {
		return
	}
	go w.loop(ctx)
}

func (w *JudgeWorker) loop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		job, err := w.queue.Dequeue(ctx)
		if err != nil {
			if errors.Is(err, ports.ErrQueueEmpty) {
				time.Sleep(w.pollInterval)
				continue
			}
			observability.LogJSON("submission.worker.dequeue_failed", map[string]any{"error": err.Error()})
			time.Sleep(w.pollInterval)
			continue
		}
		observability.IncQueueConsume()

		if _, err := w.submitSvc.MarkSubmissionJudging(ctx, job.SubmissionID, "worker"); err != nil {
			observability.LogJSON("submission.worker.mark_judging_failed", map[string]any{"submission_id": job.SubmissionID, "error": err.Error()})
			continue
		}
		observability.LogJSON("submission.worker.mark_judging", map[string]any{"submission_id": job.SubmissionID})
		if err := w.queue.Acknowledge(ctx, job.SubmissionID); err != nil {
			observability.LogJSON("submission.worker.ack_failed", map[string]any{"submission_id": job.SubmissionID, "error": err.Error()})
		}
	}
}
