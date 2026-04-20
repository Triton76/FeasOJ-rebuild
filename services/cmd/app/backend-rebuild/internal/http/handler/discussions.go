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

	userID := c.GetString("auth_user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing auth_user_id"})
		return
	}

	req.UserID = userID
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

	userID := c.GetString("auth_user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing auth_user_id"})
		return
	}

	req.UserID = userID
	resp, err := h.Discussions.CreateComment(c.Request.Context(), req)
	if err != nil {
		fail(c, err)
		return
	}
	ok(c, resp)
}

func (h Handlers) ListComments(c *gin.Context) {
	var req ports.CommentsQuery
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	req.DiscussionID = c.Param("discussion_id")
	resp, err := h.Discussions.ListComments(c.Request.Context(), req)
	if err != nil {
		fail(c, err)
		return
	}
	ok(c, resp)
}

func (h Handlers) DeleteDiscussion(c *gin.Context) {
	req := ports.DeleteDiscussionRequest{
		DiscussionID: c.Param("discussion_id"),
		ActorUserID:  c.GetString("auth_user_id"),
		ActorRole:    c.GetString("auth_role"),
	}
	if err := h.Discussions.DeleteDiscussion(c.Request.Context(), req); err != nil {
		fail(c, err)
		return
	}
	noContent(c)
}

func (h Handlers) DeleteComment(c *gin.Context) {
	req := ports.DeleteCommentRequest{
		CommentID:   c.Param("comment_id"),
		ActorUserID: c.GetString("auth_user_id"),
		ActorRole:   c.GetString("auth_role"),
	}
	if err := h.Discussions.DeleteComment(c.Request.Context(), req); err != nil {
		fail(c, err)
		return
	}
	noContent(c)
}
