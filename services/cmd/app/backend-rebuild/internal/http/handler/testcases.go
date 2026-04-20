package handler

import (
	"FeasOJ/app/backend-rebuild/internal/ports"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

func (h Handlers) CreateTestcase(c *gin.Context) {
	problemID, err := strconv.ParseInt(c.Param("problem_id"), 10, 64)
	if err != nil || problemID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid problem_id"})
		return
	}

	actorUserID := c.GetString("auth_user_id")
	if actorUserID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing auth_user_id"})
		return
	}

	var req ports.CreateTestcaseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	req.ProblemID = problemID
	req.ActorUserID = actorUserID
	req.ActorRole = c.GetString("auth_role")

	resp, err := h.Testcases.CreateTestcase(c.Request.Context(), req)
	if err != nil {
		fail(c, err)
		return
	}
	ok(c, resp)
}

func (h Handlers) ListTestcases(c *gin.Context) {
	problemID, err := strconv.ParseInt(c.Param("problem_id"), 10, 64)
	if err != nil || problemID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid problem_id"})
		return
	}

	actorUserID := c.GetString("auth_user_id")
	if actorUserID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing auth_user_id"})
		return
	}

	resp, err := h.Testcases.ListTestcases(c.Request.Context(), ports.ListTestcasesRequest{
		ProblemID:   problemID,
		ActorUserID: actorUserID,
		ActorRole:   c.GetString("auth_role"),
	})
	if err != nil {
		fail(c, err)
		return
	}
	ok(c, resp)
}

func (h Handlers) UpdateTestcase(c *gin.Context) {
	problemID, err := strconv.ParseInt(c.Param("problem_id"), 10, 64)
	if err != nil || problemID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid problem_id"})
		return
	}

	actorUserID := c.GetString("auth_user_id")
	if actorUserID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing auth_user_id"})
		return
	}

	var req ports.UpdateTestcaseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	req.ProblemID = problemID
	req.TestcaseID = c.Param("testcase_id")
	req.ActorUserID = actorUserID
	req.ActorRole = c.GetString("auth_role")

	resp, err := h.Testcases.UpdateTestcase(c.Request.Context(), req)
	if err != nil {
		fail(c, err)
		return
	}
	ok(c, resp)
}

func (h Handlers) DeleteTestcase(c *gin.Context) {
	problemID, err := strconv.ParseInt(c.Param("problem_id"), 10, 64)
	if err != nil || problemID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid problem_id"})
		return
	}

	actorUserID := c.GetString("auth_user_id")
	if actorUserID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing auth_user_id"})
		return
	}

	err = h.Testcases.DeleteTestcase(c.Request.Context(), ports.DeleteTestcaseRequest{
		ProblemID:   problemID,
		TestcaseID:  c.Param("testcase_id"),
		ActorUserID: actorUserID,
		ActorRole:   c.GetString("auth_role"),
	})
	if err != nil {
		fail(c, err)
		return
	}
	ok(c, gin.H{"deleted": true})
}

func (h Handlers) ReorderTestcases(c *gin.Context) {
	problemID, err := strconv.ParseInt(c.Param("problem_id"), 10, 64)
	if err != nil || problemID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid problem_id"})
		return
	}

	actorUserID := c.GetString("auth_user_id")
	if actorUserID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing auth_user_id"})
		return
	}

	var req ports.ReorderTestcasesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	req.ProblemID = problemID
	req.ActorUserID = actorUserID
	req.ActorRole = c.GetString("auth_role")

	resp, err := h.Testcases.ReorderTestcases(c.Request.Context(), req)
	if err != nil {
		fail(c, err)
		return
	}
	ok(c, resp)
}

func (h Handlers) JudgeListTestcases(c *gin.Context) {
	if h.JudgeWritebackToken != "" {
		if strings.TrimSpace(c.GetHeader("X-Judge-Token")) != h.JudgeWritebackToken {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid judge token"})
			return
		}
	}

	problemID, err := strconv.ParseInt(c.Param("problem_id"), 10, 64)
	if err != nil || problemID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid problem_id"})
		return
	}

	resp, err := h.Testcases.ListTestcasesForJudge(c.Request.Context(), ports.JudgeListTestcasesRequest{ProblemID: problemID})
	if err != nil {
		fail(c, err)
		return
	}
	ok(c, resp)
}
