// Layer: Router (路由层)
// Responsibility: 定义HTTP路由映射、中间件挂载、API分组
// Dependency: 依赖 Handler 层和 Middleware
package router

import (
	"FeasOJ/app/backend-rebuild/internal/config"
	"FeasOJ/app/backend-rebuild/internal/http/handler"
	"FeasOJ/app/backend-rebuild/internal/http/middleware"
	"net/http"

	"github.com/gin-gonic/gin"
)

func Register(r *gin.Engine, h handler.Handlers, cfg config.Config) {
	api := r.Group("/api/v1")
	api.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	api.POST("/auth/register", h.Register)
	api.POST("/auth/login", h.Login)
	api.POST("/auth/password/reset", h.ResetPassword)

	authed := api.Group("")
	authed.Use(middleware.HeaderVerify(cfg.JWTSecret))
	authed.GET("/auth/verify", h.Verify)

	authed.GET("/users/:user_id", h.GetProfile)
	authed.PATCH("/users/:user_id", h.UpdateProfile)
	authed.GET("/ranking", h.ListRanking)

	authed.POST("/classes", h.CreateClass)
	authed.POST("/classes/join", h.ApplyJoinClass)
	authed.POST("/classes/memberships/review", h.ReviewMembership)

	authed.GET("/problems", h.ListProblems)
	authed.GET("/problems/:problem_id", h.GetProblem)

	authed.GET("/contests", h.ListContests)
	authed.GET("/contests/:contest_id", h.GetContest)
	authed.POST("/contests/join", h.JoinContest)

	authed.GET("/discussions", h.ListDiscussions)
	authed.GET("/discussions/:discussion_id", h.GetDiscussion)
	authed.POST("/discussions", h.CreateDiscussion)
	authed.POST("/comments", h.CreateComment)

	authed.POST("/submit-records", h.CreateSubmission)
	authed.GET("/submit-records", h.ListSubmissions)

	admin := authed.Group("/admin")
	admin.Use(middleware.AdminOnly())
	admin.GET("/users", h.AdminListUsers)
	admin.PATCH("/users/status", h.AdminUpdateUserStatus)
}
