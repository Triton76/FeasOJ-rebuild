// Layer: Middleware (HTTP中间件层)
// Responsibility: JWT认证校验、权限控制(RBAC)、请求上下文注入
// Dependency: 依赖 Security 层进行Token解析，被 Router 层调用
package middleware

import (
	"FeasOJ/app/backend-rebuild/internal/ports"
	"FeasOJ/app/backend-rebuild/internal/security"
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func HeaderVerify(jwtSecret string, authService ports.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := strings.TrimSpace(c.GetHeader("Authorization"))
		if !strings.HasPrefix(authHeader, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing bearer token"})
			return
		}

		token := strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))
		claims, err := security.ParseToken(jwtSecret, token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}

		c.Set("auth_user_id", claims.UserID)
		c.Set("auth_role", claims.Role)
		c.Set("auth_status", claims.Status)

		ctx := security.WithClaims(c.Request.Context(), claims)
		c.Request = c.Request.WithContext(ctx)

		if authService != nil {
			if _, err := authService.Verify(c.Request.Context()); err != nil {
				status := http.StatusUnauthorized
				if errors.Is(err, ports.ErrForbidden) {
					status = http.StatusForbidden
				}
				c.AbortWithStatusJSON(status, gin.H{"error": "invalid credential state"})
				return
			}
		}

		c.Next()
	}
}

func AdminOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		role := c.GetString("auth_role")
		if role != "admin" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "admin required"})
			return
		}
		c.Next()
	}
}
