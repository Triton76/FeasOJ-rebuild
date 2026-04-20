// Layer: Infrastructure (基础设施层)
// Responsibility: 内存队列实现，用于本地联调与测试
// Note: 仅用于第一阶段验证语义，后续可替换为真实 MQ 实现（如 RabbitMQ/Redis）
package queue

import (
	"FeasOJ/app/backend-rebuild/internal/observability"
	"FeasOJ/app/backend-rebuild/internal/ports"
	"context"
	"sync"
)

// MemorySubmissionQueue 基于内存的提交队列实现
// 线程安全，支持至少一次投递语义
type MemorySubmissionQueue struct {
	mu    sync.RWMutex
	jobs  []*ports.SubmissionJob // FIFO 队列
	seen  map[int64]bool         // 已入队的 submission_id 集合（幂等去重）
	acked map[int64]bool         // 已确认处理的 submission_id 集合
}

// NewMemorySubmissionQueue 创建新的内存队列
func NewMemorySubmissionQueue() ports.SubmissionQueue {
	return &MemorySubmissionQueue{
		jobs:  make([]*ports.SubmissionJob, 0),
		seen:  make(map[int64]bool),
		acked: make(map[int64]bool),
	}
}

// Enqueue 入队，以 submission_id 作为幂等键
func (q *MemorySubmissionQueue) Enqueue(ctx context.Context, job ports.SubmissionJob) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.seen[job.SubmissionID] {
		if q.acked[job.SubmissionID] {
			q.seen[job.SubmissionID] = false
			q.acked[job.SubmissionID] = false
		} else {
			// 尚未处理完成时重复入队，幂等返回成功
			return nil
		}
	}

	jobCopy := job
	q.jobs = append(q.jobs, &jobCopy)
	q.seen[job.SubmissionID] = true
	observability.IncQueueEnqueue()
	observability.LogJSON("submission.enqueued", map[string]any{"submission_id": job.SubmissionID, "queue_name": "memory", "attempt": 1})
	return nil
}

// Dequeue 取出队列中的下一个任务（FIFO）
func (q *MemorySubmissionQueue) Dequeue(ctx context.Context) (*ports.SubmissionJob, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	if len(q.jobs) == 0 {
		return nil, ports.ErrQueueEmpty
	}

	job := q.jobs[0]
	q.jobs = q.jobs[1:]
	observability.IncQueueConsume()
	observability.LogJSON("submission.dequeue", map[string]any{"submission_id": job.SubmissionID, "queue_name": "memory", "attempt": 1})
	return job, nil
}

// Acknowledge 标记任务已处理完成
func (q *MemorySubmissionQueue) Acknowledge(ctx context.Context, submissionID int64) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	if !q.seen[submissionID] {
		return ports.ErrNotFound
	}

	q.acked[submissionID] = true
	observability.LogJSON("submission.ack", map[string]any{"submission_id": submissionID, "queue_name": "memory"})
	return nil
}

// Size 返回当前队列中的任务数量
func (q *MemorySubmissionQueue) Size(ctx context.Context) (int, error) {
	q.mu.RLock()
	defer q.mu.RUnlock()

	return len(q.jobs), nil
}
