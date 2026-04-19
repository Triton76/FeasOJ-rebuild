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

	resp, err := h.Classes.CreateClass(c.Request.Context(), req)
	if err != nil {
		notImplemented(c)
		return
	}
	ok(c, resp)
}

func (h Handlers) ApplyJoinClass(c *gin.Context) {
	var req ports.ApplyJoinClassRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.Classes.ApplyJoinClass(c.Request.Context(), req)
	if err != nil {
		notImplemented(c)
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

	resp, err := h.Classes.ReviewMembership(c.Request.Context(), req)
	if err != nil {
		notImplemented(c)
		return
	}
	ok(c, resp)
}
