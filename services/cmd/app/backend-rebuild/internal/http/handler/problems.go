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
	req.ActorUserID = c.GetString("auth_user_id")
	req.ActorRole = c.GetString("auth_role")

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

func (h Handlers) CreateProblem(c *gin.Context) {
	var req ports.CreateProblemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	req.ActorUserID = c.GetString("auth_user_id")
	req.ActorRole = c.GetString("auth_role")

	resp, err := h.Problems.CreateProblem(c.Request.Context(), req)
	if err != nil {
		fail(c, err)
		return
	}
	ok(c, resp)
}

func (h Handlers) UpdateProblem(c *gin.Context) {
	problemID, err := strconv.ParseInt(c.Param("problem_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid problem_id"})
		return
	}

	var req ports.UpdateProblemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	req.ProblemID = problemID
	req.ActorUserID = c.GetString("auth_user_id")
	req.ActorRole = c.GetString("auth_role")

	resp, svcErr := h.Problems.UpdateProblem(c.Request.Context(), req)
	if svcErr != nil {
		fail(c, svcErr)
		return
	}
	ok(c, resp)
}

func (h Handlers) DeleteProblem(c *gin.Context) {
	problemID, err := strconv.ParseInt(c.Param("problem_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid problem_id"})
		return
	}

	req := ports.DeleteProblemRequest{
		ProblemID:   problemID,
		ActorUserID: c.GetString("auth_user_id"),
		ActorRole:   c.GetString("auth_role"),
	}
	if err := h.Problems.DeleteProblem(c.Request.Context(), req); err != nil {
		fail(c, err)
		return
	}
	noContent(c)
}
