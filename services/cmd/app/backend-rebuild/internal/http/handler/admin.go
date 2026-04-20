// Layer: Handler (HTTP接入层) - Admin模块
// Responsibility: 管理后台HTTP端点处理(用户列表/状态更新)
// Note: 当前依赖stub实现，返回not implemented
package handler

import (
	"FeasOJ/app/backend-rebuild/internal/ports"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h Handlers) AdminListUsers(c *gin.Context) {
	var req ports.AdminUsersQuery
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.Admin.ListUsers(c.Request.Context(), req)
	if err != nil {
		fail(c, err)
		return
	}
	ok(c, resp)
}

func (h Handlers) AdminUpdateUserStatus(c *gin.Context) {
	var req ports.UpdateUserStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	req.ActorUserID = c.GetString("auth_user_id")

	resp, err := h.Admin.UpdateUserStatus(c.Request.Context(), req)
	if err != nil {
		fail(c, err)
		return
	}
	ok(c, resp)
}

func (h Handlers) AdminUpdateUserRole(c *gin.Context) {
	var req ports.UpdateUserRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	req.ActorUserID = c.GetString("auth_user_id")

	resp, err := h.Admin.UpdateUserRole(c.Request.Context(), req)
	if err != nil {
		fail(c, err)
		return
	}
	ok(c, resp)
}
