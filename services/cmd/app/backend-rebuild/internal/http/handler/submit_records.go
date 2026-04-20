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

	userID := c.GetString("auth_user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing auth_user_id"})
		return
	}

	req.UserID = userID
	resp, err := h.SubmitRecords.CreateSubmission(c.Request.Context(), req)
	if err != nil {
		fail(c, err)
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

	actorUserID := c.GetString("auth_user_id")
	if actorUserID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing auth_user_id"})
		return
	}
	if c.GetString("auth_role") != "admin" {
		req.UserID = actorUserID
	}

	resp, err := h.SubmitRecords.ListSubmissions(c.Request.Context(), req)
	if err != nil {
		fail(c, err)
		return
	}
	ok(c, resp)
}
