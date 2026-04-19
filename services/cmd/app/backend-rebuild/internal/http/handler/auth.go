// Layer: Handler (HTTP接入层/传输层)
// Responsibility: HTTP请求处理、参数解析、调用Usecase、返回HTTP响应
// Dependency: 依赖 Ports 层的服务接口，不依赖具体Usecase实现
package handler

import (
	"FeasOJ/app/backend-rebuild/internal/ports"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h Handlers) Register(c *gin.Context) {
	var req ports.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.Auth.Register(c.Request.Context(), req)
	if err != nil {
		fail(c, err)
		return
	}
	ok(c, resp)
}

func (h Handlers) Login(c *gin.Context) {
	var req ports.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.Auth.Login(c.Request.Context(), req)
	if err != nil {
		fail(c, err)
		return
	}
	ok(c, resp)
}

func (h Handlers) Verify(c *gin.Context) {
	resp, err := h.Auth.Verify(c.Request.Context())
	if err != nil {
		fail(c, err)
		return
	}
	ok(c, resp)
}
