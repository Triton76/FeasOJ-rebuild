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
	req.ActorUserID = c.GetString("auth_user_id")
	req.ActorRole = c.GetString("auth_role")

	resp, err := h.Competitions.ListContests(c.Request.Context(), req)
	if err != nil {
		fail(c, err)
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
		fail(c, svcErr)
		return
	}
	ok(c, resp)
}

func (h Handlers) CreateContest(c *gin.Context) {
	var req ports.CreateContestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	req.ActorUserID = c.GetString("auth_user_id")
	req.ActorRole = c.GetString("auth_role")

	resp, err := h.Competitions.CreateContest(c.Request.Context(), req)
	if err != nil {
		fail(c, err)
		return
	}
	ok(c, resp)
}

func (h Handlers) UpdateContest(c *gin.Context) {
	contestID, err := strconv.ParseInt(c.Param("contest_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid contest_id"})
		return
	}

	var req ports.UpdateContestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	req.ContestID = contestID
	req.ActorUserID = c.GetString("auth_user_id")
	req.ActorRole = c.GetString("auth_role")

	resp, svcErr := h.Competitions.UpdateContest(c.Request.Context(), req)
	if svcErr != nil {
		fail(c, svcErr)
		return
	}
	ok(c, resp)
}

func (h Handlers) DeleteContest(c *gin.Context) {
	contestID, err := strconv.ParseInt(c.Param("contest_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid contest_id"})
		return
	}

	req := ports.DeleteContestRequest{
		ContestID:   contestID,
		ActorUserID: c.GetString("auth_user_id"),
		ActorRole:   c.GetString("auth_role"),
	}
	if err := h.Competitions.DeleteContest(c.Request.Context(), req); err != nil {
		fail(c, err)
		return
	}
	noContent(c)
}

func (h Handlers) JoinContest(c *gin.Context) {
	var req ports.JoinContestRequest
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
	req.ActorRole = c.GetString("auth_role")
	resp, err := h.Competitions.JoinContest(c.Request.Context(), req)
	if err != nil {
		fail(c, err)
		return
	}
	ok(c, resp)
}

func (h Handlers) GetScoreboard(c *gin.Context) {
	contestID, err := strconv.ParseInt(c.Param("contest_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid contest_id"})
		return
	}

	resp, err := h.Competitions.GetScoreboard(c.Request.Context(), ports.ContestScoreboardQuery{ContestID: contestID})
	if err != nil {
		fail(c, err)
		return
	}
	ok(c, resp)
}

func (h Handlers) ListContestProblems(c *gin.Context) {
	contestID, err := strconv.ParseInt(c.Param("contest_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid contest_id"})
		return
	}

	req := ports.ContestProblemsQuery{
		ContestID:   contestID,
		ActorUserID: c.GetString("auth_user_id"),
		ActorRole:   c.GetString("auth_role"),
	}
	resp, err := h.Competitions.ListContestProblems(c.Request.Context(), req)
	if err != nil {
		fail(c, err)
		return
	}
	ok(c, resp)
}

func (h Handlers) GetContestProblem(c *gin.Context) {
	contestID, err := strconv.ParseInt(c.Param("contest_id"), 10, 64)
	if err != nil || contestID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid contest_id"})
		return
	}
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
	actorRole := c.GetString("auth_role")

	contest, err := h.Competitions.GetContest(c.Request.Context(), contestID)
	if err != nil {
		fail(c, err)
		return
	}

	if actorRole != "admin" && contest.OwnerUserID != actorUserID {
		membership, membershipErr := h.Competitions.GetContestMembership(c.Request.Context(), ports.ContestMembershipQuery{
			ContestID:   contestID,
			ActorUserID: actorUserID,
			ActorRole:   actorRole,
		})
		if membershipErr != nil {
			fail(c, membershipErr)
			return
		}
		if !membership.Joined || membership.Status == "quit" || membership.Status == "finished" {
			fail(c, ports.ErrForbidden)
			return
		}
	}

	bindings, err := h.Competitions.ListContestProblems(c.Request.Context(), ports.ContestProblemsQuery{
		ContestID:   contestID,
		ActorUserID: actorUserID,
		ActorRole:   actorRole,
	})
	if err != nil {
		fail(c, err)
		return
	}

	bound := false
	for _, item := range bindings {
		if item.ProblemID == problemID {
			bound = true
			break
		}
	}
	if !bound {
		fail(c, ports.ErrNotFound)
		return
	}

	resp, err := h.Problems.GetProblemForJudge(c.Request.Context(), problemID)
	if err != nil {
		fail(c, err)
		return
	}
	ok(c, resp)
}

func (h Handlers) ReplaceContestProblems(c *gin.Context) {
	contestID, err := strconv.ParseInt(c.Param("contest_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid contest_id"})
		return
	}

	var req ports.ReplaceContestProblemsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	req.ContestID = contestID
	req.ActorUserID = c.GetString("auth_user_id")
	req.ActorRole = c.GetString("auth_role")

	resp, svcErr := h.Competitions.ReplaceContestProblems(c.Request.Context(), req)
	if svcErr != nil {
		fail(c, svcErr)
		return
	}
	ok(c, resp)
}

func (h Handlers) GetContestMembership(c *gin.Context) {
	contestID, err := strconv.ParseInt(c.Param("contest_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid contest_id"})
		return
	}

	req := ports.ContestMembershipQuery{
		ContestID:   contestID,
		ActorUserID: c.GetString("auth_user_id"),
		ActorRole:   c.GetString("auth_role"),
	}
	resp, svcErr := h.Competitions.GetContestMembership(c.Request.Context(), req)
	if svcErr != nil {
		fail(c, svcErr)
		return
	}
	ok(c, resp)
}

func (h Handlers) ListContestParticipants(c *gin.Context) {
	contestID, err := strconv.ParseInt(c.Param("contest_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid contest_id"})
		return
	}

	req := ports.ContestParticipantsQuery{
		ContestID:   contestID,
		ActorUserID: c.GetString("auth_user_id"),
		ActorRole:   c.GetString("auth_role"),
	}
	resp, svcErr := h.Competitions.ListContestParticipants(c.Request.Context(), req)
	if svcErr != nil {
		fail(c, svcErr)
		return
	}
	ok(c, resp)
}

func (h Handlers) QuitContest(c *gin.Context) {
	contestID, err := strconv.ParseInt(c.Param("contest_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid contest_id"})
		return
	}

	req := ports.QuitContestRequest{
		ContestID:   contestID,
		ActorUserID: c.GetString("auth_user_id"),
		ActorRole:   c.GetString("auth_role"),
	}
	resp, svcErr := h.Competitions.QuitContest(c.Request.Context(), req)
	if svcErr != nil {
		fail(c, svcErr)
		return
	}
	ok(c, resp)
}
