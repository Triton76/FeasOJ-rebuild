// Layer: Ports (接口契约层)
// Responsibility: 定义领域错误码，统一系统错误语义
// Dependency: 不依赖任何其他层
package ports

import "errors"

var (
	ErrNotImplemented      = errors.New("not implemented")
	ErrInvalidArgument     = errors.New("invalid argument")
	ErrUnauthorized        = errors.New("unauthorized")
	ErrForbidden           = errors.New("forbidden")
	ErrNotFound            = errors.New("not found")
	ErrConflict            = errors.New("conflict")
	ErrRateLimited         = errors.New("rate limited")
	ErrDuplicateSubmission = errors.New("submission_id already in queue")
	ErrQueueEmpty          = errors.New("queue is empty")
)
