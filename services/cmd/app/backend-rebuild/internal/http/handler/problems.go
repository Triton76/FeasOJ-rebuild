// Layer: Handler (HTTP接入层) - Problems模块
// Responsibility: 题目相关HTTP端点处理(列表/详情)
// Note: 当前依赖stub实现，返回not implemented
package handler

import (
	"FeasOJ/app/backend-rebuild/internal/ports"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func (h Handlers) ListProblems(c *gin.Context) {
	var req ports.ProblemsQuery
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.Problems.ListProblems(c.Request.Context(), req)
	if err != nil {
		fail(c, err)
		return
	}
	ok(c, resp)
}

func (h Handlers) GetProblem(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("problem_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid problem_id"})
		return
	}

	resp, svcErr := h.Problems.GetProblem(c.Request.Context(), id)
	if svcErr != nil {
		fail(c, svcErr)
		return
	}
	ok(c, resp)
}
