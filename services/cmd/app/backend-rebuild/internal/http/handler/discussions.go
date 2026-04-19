// Layer: Handler (HTTP接入层) - Discussions模块
// Responsibility: 讨论区相关HTTP端点处理(列表/详情/创建讨论/创建评论)
// Note: 当前依赖stub实现，返回not implemented
package handler

import (
	"FeasOJ/app/backend-rebuild/internal/ports"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h Handlers) ListDiscussions(c *gin.Context) {
	var req ports.DiscussionsQuery
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.Discussions.ListDiscussions(c.Request.Context(), req)
	if err != nil {
		fail(c, err)
		return
	}
	ok(c, resp)
}

func (h Handlers) GetDiscussion(c *gin.Context) {
	resp, err := h.Discussions.GetDiscussion(c.Request.Context(), c.Param("discussion_id"))
	if err != nil {
		fail(c, err)
		return
	}
	ok(c, resp)
}

func (h Handlers) CreateDiscussion(c *gin.Context) {
	var req ports.CreateDiscussionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.Discussions.CreateDiscussion(c.Request.Context(), req)
	if err != nil {
		fail(c, err)
		return
	}
	ok(c, resp)
}

func (h Handlers) CreateComment(c *gin.Context) {
	var req ports.CreateCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.Discussions.CreateComment(c.Request.Context(), req)
	if err != nil {
		fail(c, err)
		return
	}
	ok(c, resp)
}
