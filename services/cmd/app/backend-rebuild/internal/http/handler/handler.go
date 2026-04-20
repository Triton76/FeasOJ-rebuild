// Layer: Handler (HTTP接入层/传输层)
// Responsibility: Handler容器定义、统一响应封装、错误码映射
// Dependency: 依赖 Ports 层的服务接口和错误定义
package handler

import (
	"FeasOJ/app/backend-rebuild/internal/ports"
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type Handlers struct {
	Auth          ports.AuthService
	Users         ports.UsersService
	Classes       ports.ClassesService
	Problems      ports.ProblemsService
	Competitions  ports.CompetitionsService
	Discussions   ports.DiscussionsService
	SubmitRecords ports.SubmitRecordsService
	Admin         ports.AdminService
	JudgeWritebackToken string
}

func New(h Handlers) Handlers {
	h.JudgeWritebackToken = strings.TrimSpace(h.JudgeWritebackToken)
	return h
}

func ok(c *gin.Context, data any) {
	c.JSON(http.StatusOK, gin.H{"data": data})
}

func noContent(c *gin.Context) {
	c.Status(http.StatusNoContent)
}

func fail(c *gin.Context, err error) {
	status := http.StatusInternalServerError
	switch {
	case errors.Is(err, ports.ErrNotImplemented):
		status = http.StatusNotImplemented
	case errors.Is(err, ports.ErrInvalidArgument):
		status = http.StatusBadRequest
	case errors.Is(err, ports.ErrUnauthorized):
		status = http.StatusUnauthorized
	case errors.Is(err, ports.ErrForbidden):
		status = http.StatusForbidden
	case errors.Is(err, ports.ErrNotFound):
		status = http.StatusNotFound
	case errors.Is(err, ports.ErrConflict):
		status = http.StatusConflict
	case errors.Is(err, ports.ErrRateLimited):
		status = http.StatusTooManyRequests
	}
	c.JSON(status, gin.H{"error": err.Error()})
}

func notImplemented(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented in phase-1 skeleton"})
}
