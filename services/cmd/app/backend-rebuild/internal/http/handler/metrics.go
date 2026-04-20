package handler

import (
	"FeasOJ/app/backend-rebuild/internal/observability"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h Handlers) RuntimeMetrics(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"data": observability.Snapshot()})
}
