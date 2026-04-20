// Layer: Handler (HTTP接入层) - Classes模块
// Responsibility: 班级相关HTTP端点处理(创建班级/申请加入/审核成员)
// Note: 当前依赖stub实现，返回not implemented
package handler

import (
	"FeasOJ/app/backend-rebuild/internal/ports"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h Handlers) CreateClass(c *gin.Context) {
	var req ports.CreateClassRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ownerUserID := c.GetString("auth_user_id")
	if ownerUserID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing auth_user_id"})
		return
	}

	req.OwnerUserID = ownerUserID
	resp, err := h.Classes.CreateClass(c.Request.Context(), req)
	if err != nil {
		fail(c, err)
		return
	}
	ok(c, resp)
}

func (h Handlers) UpdateClass(c *gin.Context) {
	var req ports.UpdateClassRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	actorUserID := c.GetString("auth_user_id")
	if actorUserID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing auth_user_id"})
		return
	}

	req.ClassID = c.Param("class_id")
	req.ActorUserID = actorUserID

	resp, err := h.Classes.UpdateClass(c.Request.Context(), req)
	if err != nil {
		fail(c, err)
		return
	}
	ok(c, resp)
}

func (h Handlers) ArchiveClass(c *gin.Context) {
	classID := c.Param("class_id")
	actorUserID := c.GetString("auth_user_id")
	if actorUserID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing auth_user_id"})
		return
	}

	if err := h.Classes.ArchiveClass(c.Request.Context(), classID, actorUserID); err != nil {
		fail(c, err)
		return
	}
	ok(c, gin.H{"archived": true})
}

func (h Handlers) ApplyJoinClass(c *gin.Context) {
	var req ports.ApplyJoinClassRequest
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
	resp, err := h.Classes.ApplyJoinClass(c.Request.Context(), req)
	if err != nil {
		fail(c, err)
		return
	}
	ok(c, resp)
}

func (h Handlers) ReviewMembership(c *gin.Context) {
	var req ports.ReviewMembershipRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	actorUserID := c.GetString("auth_user_id")
	if actorUserID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing auth_user_id"})
		return
	}
	req.ActorUserID = actorUserID

	resp, err := h.Classes.ReviewMembership(c.Request.Context(), req)
	if err != nil {
		fail(c, err)
		return
	}
	ok(c, resp)
}

func (h Handlers) ListClassMemberships(c *gin.Context) {
	actorUserID := c.GetString("auth_user_id")
	if actorUserID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing auth_user_id"})
		return
	}

	resp, err := h.Classes.ListClassMemberships(c.Request.Context(), c.Param("class_id"), actorUserID)
	if err != nil {
		fail(c, err)
		return
	}
	ok(c, resp)
}

func (h Handlers) ListMyMemberships(c *gin.Context) {
	userID := c.GetString("auth_user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing auth_user_id"})
		return
	}

	resp, err := h.Classes.ListMyMemberships(c.Request.Context(), userID)
	if err != nil {
		fail(c, err)
		return
	}
	ok(c, resp)
}
