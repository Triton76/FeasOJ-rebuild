// Layer: Handler (HTTP接入层) - SubmitRecords模块
// Responsibility: 提交记录相关HTTP端点处理(创建提交/查询列表)
// Note: 当前依赖stub实现，返回not implemented
package handler

import (
	"FeasOJ/app/backend-rebuild/internal/ports"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h Handlers) CreateSubmission(c *gin.Context) {
	var req ports.CreateSubmissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.SubmitRecords.CreateSubmission(c.Request.Context(), req)
	if err != nil {
		notImplemented(c)
		return
	}
	ok(c, resp)
}

func (h Handlers) ListSubmissions(c *gin.Context) {
	var req ports.SubmissionsQuery
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.SubmitRecords.ListSubmissions(c.Request.Context(), req)
	if err != nil {
		notImplemented(c)
		return
	}
	ok(c, resp)
}
