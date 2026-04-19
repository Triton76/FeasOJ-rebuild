// Layer: Handler (HTTP接入层) - Competitions模块
// Responsibility: 竞赛相关HTTP端点处理(列表/详情/加入)
// Note: 当前依赖stub实现，返回not implemented
package handler

import (
	"FeasOJ/app/backend-rebuild/internal/ports"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func (h Handlers) ListContests(c *gin.Context) {
	var req ports.ContestsQuery
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.Competitions.ListContests(c.Request.Context(), req)
	if err != nil {
		notImplemented(c)
		return
	}
	ok(c, resp)
}

func (h Handlers) GetContest(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("contest_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid contest_id"})
		return
	}

	resp, svcErr := h.Competitions.GetContest(c.Request.Context(), id)
	if svcErr != nil {
		notImplemented(c)
		return
	}
	ok(c, resp)
}

func (h Handlers) JoinContest(c *gin.Context) {
	var req ports.JoinContestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.Competitions.JoinContest(c.Request.Context(), req)
	if err != nil {
		notImplemented(c)
		return
	}
	ok(c, resp)
}
