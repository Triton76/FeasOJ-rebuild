// Layer: Ports (接口契约层)
// Responsibility: 队列端口定义，支持提交流程的异步处理
// Dependency: 不依赖任何其他层
// Semantics: 至少一次投递（At-Least-Once）语义，以 submission_id 作为幂等键
package ports

import "context"

const SubmissionJobContractV1 = "v1"

// SubmissionJob 表示一个待判题的提交任务
type SubmissionJob struct {
	ContractVersion string `json:"contract_version"`
	SubmissionID int64  `json:"submission_id"`
	UserID       string `json:"user_id"`
	ProblemID    int64  `json:"problem_id"`
	ContestID    int64  `json:"contest_id"`
	Language     string `json:"language"`
	SourceCode   string `json:"source_code"`
}

// SubmissionQueue 定义提交队列的接口
// 语义：至少一次投递 (At-Least-Once)
// - Enqueue: 如果 job.SubmissionID 已存在，幂等地返回成功（不重复入队）
// - Dequeue: 返回队列中的下一个任务，调用者需自行提交结果后再调用 Acknowledge
// - Acknowledge: 标记任务已处理完成
type SubmissionQueue interface {
	// Enqueue 将任务入队，以 submission_id 作为幂等键
	// 如果同一 submission_id 已入队，返回 ErrDuplicateSubmission
	Enqueue(ctx context.Context, job SubmissionJob) error

	// Dequeue 取出队列中的下一个任务（FIFO）
	// 如果队列为空，返回 ErrQueueEmpty
	Dequeue(ctx context.Context) (*SubmissionJob, error)

	// Acknowledge 标记任务已处理完成，可删除
	Acknowledge(ctx context.Context, submissionID int64) error

	// Size 返回当前队列中的任务数量
	Size(ctx context.Context) (int, error)
}
