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

	authed := api.Group("")
	authed.Use(middleware.HeaderVerify(cfg.JWTSecret))
	authed.GET("/auth/verify", h.Verify)

	authed.GET("/users/:user_id", h.GetProfile)
	authed.PATCH("/profile", h.UpdateProfile)
	authed.GET("/ranking", h.ListRanking)

	authed.POST("/classes", h.CreateClass)
	authed.PATCH("/classes/:class_id", h.UpdateClass)
	authed.POST("/classes/:class_id/archive", h.ArchiveClass)
	authed.POST("/classes/join", h.ApplyJoinClass)
	authed.POST("/classes/memberships/review", h.ReviewMembership)
	authed.GET("/classes/:class_id/memberships", h.ListClassMemberships)
	authed.GET("/classes/memberships/self", h.ListMyMemberships)

	authed.GET("/problems", h.ListProblems)
	authed.POST("/problems", h.CreateProblem)
	authed.GET("/problems/:problem_id", h.GetProblem)
	authed.PATCH("/problems/:problem_id", h.UpdateProblem)
	authed.DELETE("/problems/:problem_id", h.DeleteProblem)

	authed.GET("/contests", h.ListContests)
	authed.POST("/contests", h.CreateContest)
	authed.GET("/contests/:contest_id", h.GetContest)
	if cfg.EnableScoreboard {
		authed.GET("/contests/:contest_id/scoreboard", h.GetScoreboard)
	}
	authed.PATCH("/contests/:contest_id", h.UpdateContest)
	authed.DELETE("/contests/:contest_id", h.DeleteContest)
	authed.POST("/contests/join", h.JoinContest)

	authed.GET("/discussions", h.ListDiscussions)
	authed.GET("/discussions/:discussion_id", h.GetDiscussion)
	authed.POST("/discussions", h.CreateDiscussion)
	authed.POST("/comments", h.CreateComment)

	authed.POST("/submit-records", h.CreateSubmission)
	authed.GET("/submit-records", h.ListSubmissions)
	authed.GET("/metrics/runtime", h.RuntimeMetrics)
	if cfg.EnableJudgeWriteback {
		api.POST("/judge/writeback", h.JudgeWriteback)
	}

	admin := authed.Group("/admin")
	admin.Use(middleware.AdminOnly())
	admin.GET("/users", h.AdminListUsers)
	admin.PATCH("/users/status", h.AdminUpdateUserStatus)
}
