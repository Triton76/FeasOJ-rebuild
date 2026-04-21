package server

import (
	"FeasOJ/app/judgecore/internal/judge"
	"FeasOJ/app/judgecore/server/handler"
	"FeasOJ/app/judgecore/server/middlewares"

	"github.com/gin-gonic/gin"
)

func LoadRouter(r *gin.Engine, pool *judge.JudgePool, codeDir string) {
	r.Use(middlewares.Logger())

	// Create a handler instance with its dependencies
	h := handler.NewHandler(codeDir)

	apiV1 := r.Group("/api/v1/judgecore")
	{
		apiV1.GET("/health", h.Health)
		apiV1.POST("/judge", h.Judge)
	}
}
