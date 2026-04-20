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
	if cfg.AvatarUploadDir != "" {
		r.StaticFS("/api/v1/avatar", http.Dir(cfg.AvatarUploadDir))
	}

	api := r.Group("/api/v1")
	api.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	api.POST("/auth/register", h.Register)
	api.POST("/auth/login", h.Login)
	api.POST("/auth/password/reset/code", h.SendPasswordResetCode)
	api.POST("/auth/password/reset", h.ResetPassword)

	authed := api.Group("")
	authed.Use(middleware.HeaderVerify(cfg.JWTSecret, h.Auth))
	authed.GET("/auth/verify", h.Verify)

	authed.GET("/users/:user_id", h.GetProfile)
	authed.POST("/profile/avatar/upload", h.UploadAvatar)
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
	if cfg.EnableTestcaseAPIs {
		authed.POST("/problems/:problem_id/testcases", h.CreateTestcase)
		authed.GET("/problems/:problem_id/testcases", h.ListTestcases)
		authed.PATCH("/problems/:problem_id/testcases/:testcase_id", h.UpdateTestcase)
		authed.DELETE("/problems/:problem_id/testcases/:testcase_id", h.DeleteTestcase)
		authed.POST("/problems/:problem_id/testcases/reorder", h.ReorderTestcases)
	}

	authed.GET("/contests", h.ListContests)
	authed.POST("/contests", h.CreateContest)
	authed.GET("/contests/:contest_id", h.GetContest)
	authed.GET("/contests/:contest_id/problems", h.ListContestProblems)
	authed.PUT("/contests/:contest_id/problems", h.ReplaceContestProblems)
	authed.GET("/contests/:contest_id/participant/self", h.GetContestMembership)
	authed.GET("/contests/:contest_id/participants", h.ListContestParticipants)
	authed.DELETE("/contests/:contest_id/participant/self", h.QuitContest)
	if cfg.EnableScoreboard {
		authed.GET("/contests/:contest_id/scoreboard", h.GetScoreboard)
	}
	authed.PATCH("/contests/:contest_id", h.UpdateContest)
	authed.DELETE("/contests/:contest_id", h.DeleteContest)
	authed.POST("/contests/join", h.JoinContest)

	authed.GET("/discussions", h.ListDiscussions)
	authed.GET("/discussions/:discussion_id", h.GetDiscussion)
	authed.GET("/discussions/:discussion_id/comments", h.ListComments)
	authed.POST("/discussions", h.CreateDiscussion)
	authed.DELETE("/discussions/:discussion_id", h.DeleteDiscussion)
	authed.POST("/comments", h.CreateComment)
	authed.DELETE("/comments/:comment_id", h.DeleteComment)

	authed.POST("/submit-records", h.CreateSubmission)
	authed.GET("/submit-records", h.ListSubmissions)
	authed.GET("/metrics/runtime", h.RuntimeMetrics)
	if cfg.EnableJudgeWriteback {
		api.POST("/judge/writeback", h.JudgeWriteback)
	}
	if cfg.EnableTestcaseAPIs {
		api.GET("/judge/problems/:problem_id/testcases", h.JudgeListTestcases)
	}

	admin := authed.Group("/admin")
	admin.Use(middleware.AdminOnly())
	admin.GET("/users", h.AdminListUsers)
	admin.PATCH("/users/status", h.AdminUpdateUserStatus)
	admin.PATCH("/users/role", h.AdminUpdateUserRole)
}
