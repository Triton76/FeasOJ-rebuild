// Layer: Handler (HTTP接入层) - Users模块
// Responsibility: 用户相关HTTP端点处理(获取资料/更新资料/排行榜)
package handler

import (
	"FeasOJ/app/backend-rebuild/internal/ports"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h Handlers) GetProfile(c *gin.Context) {
	resp, err := h.Users.GetProfile(c.Request.Context(), c.Param("user_id"))
	if err != nil {
		fail(c, err)
		return
	}
	ok(c, resp)
}

func (h Handlers) UpdateProfile(c *gin.Context) {
	var req ports.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.Users.UpdateProfile(c.Request.Context(), c.Param("user_id"), req)
	if err != nil {
		fail(c, err)
		return
	}
	ok(c, resp)
}

func (h Handlers) ListRanking(c *gin.Context) {
	var req ports.RankingQuery
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.Users.ListRanking(c.Request.Context(), req)
	if err != nil {
		fail(c, err)
		return
	}
	ok(c, resp)
}
